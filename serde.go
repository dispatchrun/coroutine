package coroutine

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"unsafe"
)

// serde.go contains the reflection based serialization and deserialization
// procedures. It does not do any type memoization, as eventually codegen should
// be able to generate code for types.
//
// It depends on the global type register being fed with possible types
// contained in interfaces. coroc automatically generates init() functions to
// register the types likely to be used in the program. Use RegisterType[T] to
// manually add a type to the register.

// Serialize x at the end of b, returning it.
func Serialize(x any, b []byte) []byte {
	s := newSerializer()
	w := &x // w is *interface{}
	wr := reflect.ValueOf(w)
	p := wr.UnsafePointer() // *interface{}
	t := wr.Elem().Type()   // what x contains
	return serializeAny(s, t, p, b)
}

// Deserialize value from b. Return left over bytes.
func Deserialize(b []byte) (interface{}, []byte) {
	d := newDeserializer()
	var x interface{}
	px := &x
	t := reflect.TypeOf(px).Elem()
	p := unsafe.Pointer(px)
	b = deserializeInterface(d, t, p, b)
	return x, b
}

// Serializable objects can be manually serialized to bytes. Types that
// implement this interface are serialized with the MarshalAppend method and
// deserialized with Unmarshal, instead of the built-in decoders.
type Serializable interface {
	// MarshalAppend marshals the object and appends the resulting bytes to
	// the provided buffer.
	MarshalAppend(b []byte) ([]byte, error)

	// Unmarshal unmarshals an object from a buffer. It returns the number
	// of bytes that were read from the buffer in order to reconstruct the
	// object.
	Unmarshal(b []byte) (n int, err error)
}

// RegisterType into the global register to make it known to the serialization
// system. coroc usually generates calls to this function.
//
// Types are recursively added, as well as *T.
func RegisterType[T any]() {
	tm.Add(reflect.TypeOf((*T)(nil)).Elem())
}

type deserializer struct {
	// TODO: make it a slice
	ptrs map[sID]unsafe.Pointer
}

func newDeserializer() *deserializer {
	return &deserializer{
		ptrs: make(map[sID]unsafe.Pointer),
	}
}

func (d *deserializer) ReadPtr(b []byte) (unsafe.Pointer, sID, []byte) {
	x, n := binary.Varint(b)
	i := sID(x)
	p := d.ptrs[i]

	slog.Debug("Deserializer ReadPtr", "i", i, "p", p, "n", n)
	return p, i, b[n:]
}

func (d *deserializer) Store(i sID, p unsafe.Pointer) {
	if d.ptrs[i] != nil {
		panic(fmt.Errorf("trying to overwirte known ID %d with %p", i, p))
	}
	d.ptrs[i] = p
}

type serializer struct {
	ptrs map[unsafe.Pointer]sID
}

func newSerializer() *serializer {
	return &serializer{
		ptrs: make(map[unsafe.Pointer]sID),
	}
}

func (s *serializer) WritePtr(p unsafe.Pointer, b []byte) (bool, []byte) {
	off := len(b)
	if p == nil {
		slog.Debug("Serializer WritePtr wrote <nil> pointer", "offset", off)
		return true, binary.AppendVarint(b, 0)
	}
	i, ok := s.ptrs[p]
	if !ok {
		i = sID(len(s.ptrs) + 1)
		s.ptrs[p] = i
	}
	slog.Debug("Serializer WritePtr", "i", i, "p", p, "offset", off)
	return ok, binary.AppendVarint(b, int64(i))
}

func serializeSize(size int, b []byte) []byte {
	return binary.AppendVarint(b, int64(size))
}

func deserializeSize(b []byte) (int, []byte) {
	l, n := binary.Varint(b)
	return int(l), b[n:]
}

type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

var (
	serializableT = reflect.TypeOf((*Serializable)(nil)).Elem()
)

func serializeAny(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	if t.Implements(serializableT) {
		b, err := reflect.NewAt(t, p).Elem().Interface().(Serializable).MarshalAppend(b)
		if err != nil {
			panic(fmt.Errorf("unhandled marshalappend err: %w", err))
		}
		return b
	}

	if reflect.PointerTo(t).Implements(serializableT) {
		b, err := reflect.NewAt(t, p).Interface().(Serializable).MarshalAppend(b)
		if err != nil {
			panic(fmt.Errorf("unhandled marshalappend err: %w", err))
		}
		return b
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't serialize reflect.Invalid"))
	case reflect.Bool:
		return serializeBool(s, *(*bool)(p), b)
	case reflect.Int:
		return serializeInt(s, *(*int)(p), b)
	case reflect.Int64:
		return serializeInt64(s, *(*int64)(p), b)
	case reflect.Int32:
		return serializeInt32(s, *(*int32)(p), b)
	case reflect.Int16:
		return serializeInt16(s, *(*int16)(p), b)
	case reflect.Int8:
		return serializeInt8(s, *(*int8)(p), b)
	case reflect.Uint:
		return serializeUint(s, *(*uint)(p), b)
	case reflect.Uint64:
		return serializeUint64(s, *(*uint64)(p), b)
	case reflect.Uint32:
		return serializeUint32(s, *(*uint32)(p), b)
	case reflect.Uint16:
		return serializeUint16(s, *(*uint16)(p), b)
	case reflect.Uint8:
		return serializeUint8(s, *(*uint8)(p), b)
	case reflect.Float64:
		return serializeFloat64(s, *(*float64)(p), b)
	case reflect.Float32:
		return serializeFloat32(s, *(*float32)(p), b)
	case reflect.Complex64:
		return serializeComplex64(s, *(*complex64)(p), b)
	case reflect.Complex128:
		return serializeComplex128(s, *(*complex128)(p), b)
	case reflect.String:
		return serializeString(s, (*string)(p), b)
	case reflect.Array:
		return serializeArray(s, t, p, b)
	case reflect.Interface:
		return serializeInterface(s, t, p, b)
	case reflect.Map:
		return serializeMap(s, t, p, b)
	case reflect.Pointer:
		return serializePointer(s, t, p, b)
	case reflect.Slice:
		return serializeSlice(s, t, p, b)
	case reflect.Struct:
		return serializeStruct(s, t, p, b)
	// Chan
	// Func
	// UnsafePointer
	default:
		panic(fmt.Errorf("reflection cannot serialize type %s", t))
	}
}

func deserializeAny(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	if t.Implements(serializableT) {
		i, err := reflect.NewAt(t, p).Elem().Interface().(Serializable).Unmarshal(b)
		if err != nil {
			panic(fmt.Errorf("unhandled unmarshal err: %w", err))
		}
		return b[i:]
	}
	if reflect.PointerTo(t).Implements(serializableT) {
		i, err := reflect.NewAt(t, p).Interface().(Serializable).Unmarshal(b)
		if err != nil {
			panic(fmt.Errorf("unhandled unmarshal err: %w", err))
		}
		return b[i:]
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't deserialize reflect.Invalid"))
	case reflect.Bool:
		return deserializeBool(d, (*bool)(p), b)
	case reflect.Int:
		return deserializeInt(d, (*int)(p), b)
	case reflect.Int64:
		return deserializeInt64(d, (*int64)(p), b)
	case reflect.Int32:
		return deserializeInt32(d, (*int32)(p), b)
	case reflect.Int16:
		return deserializeInt16(d, (*int16)(p), b)
	case reflect.Int8:
		return deserializeInt8(d, (*int8)(p), b)
	case reflect.Uint:
		return deserializeUint(d, (*uint)(p), b)
	case reflect.Uint64:
		return deserializeUint64(d, (*uint64)(p), b)
	case reflect.Uint32:
		return deserializeUint32(d, (*uint32)(p), b)
	case reflect.Uint16:
		return deserializeUint16(d, (*uint16)(p), b)
	case reflect.Uint8:
		return deserializeUint8(d, (*uint8)(p), b)
	case reflect.Float64:
		return deserializeFloat64(d, (*float64)(p), b)
	case reflect.Float32:
		return deserializeFloat32(d, (*float32)(p), b)
	case reflect.Complex64:
		return deserializeComplex64(d, (*complex64)(p), b)
	case reflect.Complex128:
		return deserializeComplex128(d, (*complex128)(p), b)
	case reflect.String:
		return deserializeString(d, (*string)(p), b)
	case reflect.Interface:
		return deserializeInterface(d, t, p, b)
	case reflect.Pointer:
		return deserializePointer(d, t, p, b)
	case reflect.Array:
		return deserializeArray(d, t, p, b)
	case reflect.Slice:
		return deserializeSlice(d, t, p, b)
	case reflect.Map:
		return deserializeMap(d, t, p, b)
	case reflect.Struct:
		return deserializeStruct(d, t, p, b)
	default:
		panic(fmt.Errorf("reflection cannot deserialize type %s", t))
	}
}

func serializePointedAt(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	exists, b := s.WritePtr(p, b)
	if exists {
		return b
	}

	// serialize the actual data if needed
	return serializeAny(s, t, p, b)
}

func serializeMap(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	size := 0
	r := reflect.NewAt(t, p).Elem()
	if r.IsNil() {
		size = -1
	} else {
		size = r.Len()
	}
	b = binary.AppendVarint(b, int64(size))

	// TODO: allocs
	iter := r.MapRange()
	k := reflect.New(t.Key()).Elem()
	v := reflect.New(t.Elem()).Elem()
	for iter.Next() {
		k.Set(iter.Key())
		v.Set(iter.Value())
		b = serializeAny(s, t.Key(), k.Addr().UnsafePointer(), b)
		b = serializeAny(s, t.Elem(), v.Addr().UnsafePointer(), b)
	}
	return b
}

func deserializeMap(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n, b := deserializeSize(b)
	if n < 0 { // nil map
		return b
	}
	nv := reflect.MakeMapWithSize(t, n)
	r := reflect.NewAt(t, p)
	r.Elem().Set(nv)
	for i := 0; i < n; i++ {
		k := reflect.New(t.Key())
		b = deserializeAny(d, t.Key(), k.UnsafePointer(), b)
		v := reflect.New(t.Elem())
		b = deserializeAny(d, t.Elem(), v.UnsafePointer(), b)
		r.Elem().SetMapIndex(k.Elem(), v.Elem())
	}
	return b
}

func serializeSlice(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p).Elem()

	b = serializeSize(r.Len(), b)
	b = serializeSize(r.Cap(), b)

	at := reflect.ArrayOf(r.Cap(), t.Elem())
	ap := r.UnsafePointer()

	b = serializePointedAt(s, at, ap, b)

	return b
}

func deserializeSlice(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {

	l, b := deserializeSize(b)
	c, b := deserializeSize(b)

	at := reflect.ArrayOf(c, t.Elem())
	ar, b := deserializePointedAt(d, at, b)

	if ar.IsNil() {
		return b
	}

	s := (*reflect.SliceHeader)(p)
	s.Data = uintptr(ar.UnsafePointer())
	s.Cap = c
	s.Len = l
	return b
}

func serializeArray(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.Len()
	te := t.Elem()
	ts := int(te.Size())
	for i := 0; i < n; i++ {
		pe := unsafe.Add(p, ts*i)
		b = serializeAny(s, te, pe, b)
	}
	return b
}

func deserializeArray(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	size := int(t.Elem().Size())
	te := t.Elem()
	for i := 0; i < t.Len(); i++ {
		pe := unsafe.Add(p, size*i)
		b = deserializeAny(d, te, pe, b)
	}
	return b
}

func serializePointer(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p).Elem()
	x := r.UnsafePointer()
	return serializePointedAt(s, t.Elem(), x, b)
}

func deserializePointedAt(d *deserializer, t reflect.Type, b []byte) (reflect.Value, []byte) {

	// This function is a bit different than the other deserialize* ones
	// because it deserializes into an unknown location. As a result,
	// instead of taking an unsafe.Pointer as an input, it returns a
	// reflect.Value that contains a *T (where T is given by the argument
	// t).

	x, i, b := d.ReadPtr(b)
	if x != nil || i == 0 { // pointer already seen or nil
		return reflect.NewAt(t, x), b
	}

	e := reflect.New(t)
	ep := e.UnsafePointer()
	d.Store(i, ep)

	return e, deserializeAny(d, t, ep, b)
}

func deserializePointer(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	ep, b := deserializePointedAt(d, t.Elem(), b)

	r := reflect.NewAt(t, p)
	r.Elem().Set(ep)

	return b
}

func serializeStruct(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		fp := unsafe.Add(p, ft.Offset)
		b = serializeAny(s, ft.Type, fp, b)
	}
	return b
}

func deserializeStruct(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		fp := unsafe.Add(p, ft.Offset)
		b = deserializeAny(d, ft.Type, fp, b)
	}
	return b
}

func serializeInterface(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	i := (*iface)(p)

	// Serialize empty interface as just -1.
	//
	// TODO: there's probably a bug here for an interface with a type
	// pointer but a nil data pointer.
	if i.typ == nil || i.ptr == nil {
		return binary.AppendVarint(b, -1)
	}

	x := *(*interface{})(p)
	et := reflect.TypeOf(x)

	id := tm.IDof(et)
	b = binary.AppendVarint(b, int64(id))

	eptr := i.ptr
	if inlined(et) {
		xp := i.ptr
		eptr = unsafe.Pointer(&xp)
		// noescape?
	}

	return serializePointedAt(s, et, eptr, b)
}

func deserializeInterface(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	// Deserialize the type id
	tid, n := binary.Varint(b)
	b = b[n:]
	if tid == -1 {
		// nothing to do?
		return b
	}
	et := tm.TypeOf(sID(tid))

	// Deserialize the pointer
	ep, b := deserializePointedAt(d, et, b)

	// Store the result in the interface
	r := reflect.NewAt(t, p)
	r.Elem().Set(ep.Elem())

	return b
}

func serializeString(s *serializer, x *string, b []byte) []byte {
	b = binary.AppendVarint(b, int64(len(*x)))
	return append(b, *x...)
}

func deserializeString(d *deserializer, x *string, b []byte) []byte {
	l, n := binary.Varint(b)
	b = b[n:]
	*x = string(b[:l])
	return b[l:]
}

func serializeBool(s *serializer, x bool, b []byte) []byte {
	c := byte(0)
	if x {
		c = 1
	}
	return append(b, c)
}

func deserializeBool(d *deserializer, x *bool, b []byte) []byte {
	*x = b[0] == 1
	return b[1:]
}

func serializeInt(s *serializer, x int, b []byte) []byte {
	return serializeInt64(s, int64(x), b)
}

func deserializeInt(d *deserializer, x *int, b []byte) []byte {
	*x = int(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeInt64(s *serializer, x int64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, uint64(x))
}

func deserializeInt64(d *deserializer, x *int64, b []byte) []byte {
	*x = int64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeInt32(s *serializer, x int32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, uint32(x))
}

func deserializeInt32(d *deserializer, x *int32, b []byte) []byte {
	*x = int32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func serializeInt16(s *serializer, x int16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, uint16(x))
}

func deserializeInt16(d *deserializer, x *int16, b []byte) []byte {
	*x = int16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func serializeInt8(s *serializer, x int8, b []byte) []byte {
	return append(b, byte(x))
}

func deserializeInt8(d *deserializer, x *int8, b []byte) []byte {
	*x = int8(b[0])
	return b[1:]
}

func serializeUint(s *serializer, x uint, b []byte) []byte {
	return serializeUint64(s, uint64(x), b)
}

func deserializeUint(d *deserializer, x *uint, b []byte) []byte {
	*x = uint(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeUint64(s *serializer, x uint64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, x)
}

func deserializeUint64(d *deserializer, x *uint64, b []byte) []byte {
	*x = uint64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeUint32(s *serializer, x uint32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, x)
}

func deserializeUint32(d *deserializer, x *uint32, b []byte) []byte {
	*x = uint32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func serializeUint16(s *serializer, x uint16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, x)
}

func deserializeUint16(d *deserializer, x *uint16, b []byte) []byte {
	*x = uint16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func serializeUint8(s *serializer, x uint8, b []byte) []byte {
	return append(b, byte(x))
}

func deserializeUint8(d *deserializer, x *uint8, b []byte) []byte {
	*x = uint8(b[0])
	return b[1:]
}

func serializeFloat32(s *serializer, x float32, b []byte) []byte {
	return serializeUint32(s, math.Float32bits(x), b)
}

func deserializeFloat32(d *deserializer, x *float32, b []byte) []byte {
	return deserializeUint32(d, (*uint32)(unsafe.Pointer(x)), b)
}

func serializeFloat64(s *serializer, x float64, b []byte) []byte {
	return serializeUint64(s, math.Float64bits(x), b)
}

func deserializeFloat64(d *deserializer, x *float64, b []byte) []byte {
	return deserializeUint64(d, (*uint64)(unsafe.Pointer(x)), b)
}

func serializeComplex64(s *serializer, x complex64, b []byte) []byte {
	b = serializeFloat32(s, real(x), b)
	b = serializeFloat32(s, imag(x), b)
	return b
}

func deserializeComplex64(d *deserializer, x *complex64, b []byte) []byte {
	// TODO: remove allocs
	var r float32
	b = deserializeFloat32(d, &r, b)
	var i float32
	b = deserializeFloat32(d, &i, b)
	*x = complex(r, i)
	return b
}

func serializeComplex128(s *serializer, x complex128, b []byte) []byte {
	b = serializeFloat64(s, real(x), b)
	b = serializeFloat64(s, imag(x), b)
	return b
}

func deserializeComplex128(d *deserializer, x *complex128, b []byte) []byte {
	// TODO: remove allocs
	var r float64
	b = deserializeFloat64(d, &r, b)
	var i float64
	b = deserializeFloat64(d, &i, b)
	*x = complex(r, i)
	return b
}

func inlined(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Struct:
		return t.NumField() == 1 && inlined(t.Field(0).Type)
	default:
		return false
	}
}

// sID is the unique sID of a pointer or type in the serialized format.
type sID int64

type typeMap struct {
	byID   map[sID]reflect.Type
	byType map[reflect.Type]sID
}

func newTypeMap() *typeMap {
	return &typeMap{
		byID:   make(map[sID]reflect.Type),
		byType: make(map[reflect.Type]sID),
	}
}

func (m *typeMap) add(t reflect.Type) {
	if _, ok := m.byType[t]; ok {
		return
	}
	id := sID(len(m.byID)) + 1
	m.byType[t] = id
	m.byID[id] = t
}

func (m *typeMap) exists(t reflect.Type) bool {
	_, ok := m.byType[t]
	return ok
}

func (m *typeMap) Add(t reflect.Type) {
	if m.exists(t) {
		return
	}
	m.add(t)
	m.add(reflect.PointerTo(t))

	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		m.Add(t.Elem())
	case reflect.Map:
		m.Add(t.Key())
		m.Add(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			m.Add(t.Field(i).Type)
		}
	}
}

func (m *typeMap) IDof(x reflect.Type) sID {
	id, ok := m.byType[x]
	if !ok {
		panic(fmt.Errorf("type '%s' is not registered", x))
	}
	return id
}

func (m *typeMap) TypeOf(x sID) reflect.Type {
	t, ok := m.byID[x]
	if !ok {
		panic(fmt.Errorf("type id '%d' not registered", x))
	}
	return t
}

// Global type register.
var tm *typeMap = newTypeMap()
