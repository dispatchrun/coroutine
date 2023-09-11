package serde

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

// reflect.go contains the reflection based serialization and deserialization
// procedures. It does not do any type memoization, as eventually codegen should
// be able to generate code for types.

// Serialize x at the end of b, returning it.
func Serialize(x any, b []byte) []byte {
	s := EnsureSerializer(nil)
	w := &x // w is *interface{}
	wr := reflect.ValueOf(w)
	p := wr.UnsafePointer() // *interface{}
	t := wr.Elem().Type()   // what x contains
	return serializeAny(s, t, p, b)
}

// Deserialize value from b.
func Deserialize(b []byte) (interface{}, []byte) {
	d := newDeserializer(nil)
	var x interface{}
	px := &x
	t := reflect.TypeOf(px).Elem()
	p := unsafe.Pointer(px)
	b = deserializeInterface(d, t, p, b)
	return x, b
}

var (
	serializableT = reflect.TypeOf((*Serializable)(nil)).Elem()
)

type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func serializeAny(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func deserializeAny(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func serializeMap(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func deserializeMap(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func serializeSlice(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p).Elem()
	b = serializeSize(r, b)
	te := t.Elem()
	start := r.UnsafePointer()
	size := int(te.Size())
	for i := 0; i < r.Len(); i++ {
		pe := unsafe.Add(start, i*size)
		b = serializeAny(s, te, pe, b)
	}
	return b
}

func deserializeSlice(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n, b := deserializeSize(b)
	if n < 0 {
		return b
	}
	nv := reflect.MakeSlice(t, n, n)
	r := reflect.NewAt(t, p)
	r.Elem().Set(nv)
	size := int(t.Elem().Size())
	p = nv.UnsafePointer()
	for i := 0; i < n; i++ {
		e := unsafe.Add(p, size*i)
		b = deserializeAny(d, t.Elem(), e, b)
	}
	return b
}

func serializeArray(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.Len()
	te := t.Elem()
	ts := int(te.Size())
	for i := 0; i < n; i++ {
		pe := unsafe.Add(p, ts*i)
		b = serializeAny(s, te, pe, b)
	}
	return b
}

func deserializeArray(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	size := int(t.Elem().Size())
	te := t.Elem()
	for i := 0; i < t.Len(); i++ {
		pe := unsafe.Add(p, size*i)
		b = deserializeAny(d, te, pe, b)
	}
	return b
}

func serializePointer(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p).Elem()
	x := r.UnsafePointer()
	ok, b := s.WritePtr(x, b)
	if !ok {
		b = serializeAny(s, t.Elem(), x, b)
	}
	return b
}

func deserializePointer(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p)
	x, i, b := d.ReadPtr(b)
	if x != nil || i == 0 { // pointer already seen or nil
		r.Elem().Set(reflect.NewAt(t.Elem(), x))
	} else {
		newthing := reflect.New(t.Elem())
		r.Elem().Set(newthing)
		d.Store(i, newthing.UnsafePointer())
		b = deserializeAny(d, t.Elem(), newthing.UnsafePointer(), b)
	}
	return b
}

func reflectFieldSupported(ft reflect.StructField) bool {
	// TODO
	return true
	//	return !ft.Anonymous //&& ft.IsExported()
}

func serializeStruct(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		if !reflectFieldSupported(ft) {
			continue
		}
		fp := unsafe.Add(p, ft.Offset)
		b = serializeAny(s, ft.Type, fp, b)
	}
	return b
}

func deserializeStruct(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		if !reflectFieldSupported(ft) {
			continue
		}

		fp := unsafe.Add(p, ft.Offset)
		b = deserializeAny(d, ft.Type, fp, b)
	}
	return b
}

func serializeInterface(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

	exists, b := s.WritePtr(eptr, b)
	if exists {
		return b
	}

	// serialize the actual data if needed
	return serializeAny(s, et, eptr, b)
}

func deserializeInterface(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	tid, n := binary.Varint(b)
	b = b[n:]
	if tid == -1 {
		// nothing to do?
		return b
	}

	te := tm.TypeOf(ID(tid))

	pe, id, b := d.ReadPtr(b)
	if pe != nil { // already been deserialized
		r := reflect.NewAt(t, p)
		val := reflect.NewAt(t, pe).Elem()
		r.Elem().Set(val)
		return b
	}
	if id == 0 { // nil pointer
		r := reflect.NewAt(t, p)
		val := reflect.Zero(reflect.PointerTo(te)).Elem()
		r.Elem().Set(val)
		return b
	}

	pre := reflect.New(te)
	pe = pre.UnsafePointer()
	d.Store(id, p)

	b = deserializeAny(d, te, pe, b)
	reflect.NewAt(t, p).Elem().Set(pre.Elem())
	return b
}

func serializeString(s *Serializer, x *string, b []byte) []byte {
	b = binary.AppendVarint(b, int64(len(*x)))
	return append(b, *x...)
}

func deserializeString(d *Deserializer, x *string, b []byte) []byte {
	l, n := binary.Varint(b)
	b = b[n:]
	*x = string(b[:l])
	return b[l:]
}

func serializeBool(s *Serializer, x bool, b []byte) []byte {
	c := byte(0)
	if x {
		c = 1
	}
	return append(b, c)
}

func deserializeBool(d *Deserializer, x *bool, b []byte) []byte {
	*x = b[0] == 1
	return b[1:]
}

func serializeInt(s *Serializer, x int, b []byte) []byte {
	return serializeInt64(s, int64(x), b)
}

func deserializeInt(d *Deserializer, x *int, b []byte) []byte {
	*x = int(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeInt64(s *Serializer, x int64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, uint64(x))
}

func deserializeInt64(d *Deserializer, x *int64, b []byte) []byte {
	*x = int64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeInt32(s *Serializer, x int32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, uint32(x))
}

func deserializeInt32(d *Deserializer, x *int32, b []byte) []byte {
	*x = int32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func serializeInt16(s *Serializer, x int16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, uint16(x))
}

func deserializeInt16(d *Deserializer, x *int16, b []byte) []byte {
	*x = int16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func serializeInt8(s *Serializer, x int8, b []byte) []byte {
	return append(b, byte(x))
}

func deserializeInt8(d *Deserializer, x *int8, b []byte) []byte {
	*x = int8(b[0])
	return b[1:]
}

func serializeUint(s *Serializer, x uint, b []byte) []byte {
	return serializeUint64(s, uint64(x), b)
}

func deserializeUint(d *Deserializer, x *uint, b []byte) []byte {
	*x = uint(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeUint64(s *Serializer, x uint64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, x)
}

func deserializeUint64(d *Deserializer, x *uint64, b []byte) []byte {
	*x = uint64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func serializeUint32(s *Serializer, x uint32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, x)
}

func deserializeUint32(d *Deserializer, x *uint32, b []byte) []byte {
	*x = uint32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func serializeUint16(s *Serializer, x uint16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, x)
}

func deserializeUint16(d *Deserializer, x *uint16, b []byte) []byte {
	*x = uint16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func serializeUint8(s *Serializer, x uint8, b []byte) []byte {
	return append(b, byte(x))
}

func deserializeUint8(d *Deserializer, x *uint8, b []byte) []byte {
	*x = uint8(b[0])
	return b[1:]
}

func serializeFloat32(s *Serializer, x float32, b []byte) []byte {
	return serializeUint32(s, math.Float32bits(x), b)
}

func deserializeFloat32(d *Deserializer, x *float32, b []byte) []byte {
	return deserializeUint32(d, (*uint32)(unsafe.Pointer(x)), b)
}

func serializeFloat64(s *Serializer, x float64, b []byte) []byte {
	return serializeUint64(s, math.Float64bits(x), b)
}

func deserializeFloat64(d *Deserializer, x *float64, b []byte) []byte {
	return deserializeUint64(d, (*uint64)(unsafe.Pointer(x)), b)
}

func serializeComplex64(s *Serializer, x complex64, b []byte) []byte {
	b = serializeFloat32(s, real(x), b)
	b = serializeFloat32(s, imag(x), b)
	return b
}

func deserializeComplex64(d *Deserializer, x *complex64, b []byte) []byte {
	// TODO: remove allocs
	var r float32
	b = deserializeFloat32(d, &r, b)
	var i float32
	b = deserializeFloat32(d, &i, b)
	*x = complex(r, i)
	return b
}

func serializeComplex128(s *Serializer, x complex128, b []byte) []byte {
	b = serializeFloat64(s, real(x), b)
	b = serializeFloat64(s, imag(x), b)
	return b
}

func deserializeComplex128(d *Deserializer, x *complex128, b []byte) []byte {
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
