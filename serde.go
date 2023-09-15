package coroutine

// serde.go contains the reflection based serialization and deserialization
// procedures. It does not do any type memoization, as eventually codegen should
// be able to generate code for types. Almost nothing is optimized, as we are
// iterating on how it works to get it right first.

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"sort"
	"unsafe"
)

// Serialize x at the end of b, returning it.
//
// To serialize interfaces, the global type register needs to be fed with
// possible types they can contain. If using coroc, it automatically generates
// init functions to register types likely to be used in the program. If not,
// use [RegisterType] to manually add a type to the register. Because
// [Serialize] starts with an interface, at least the type of the provided value
// needs to be registered.
//
// The output of Serialize can be reconstructed back to a Go value using
// [Deserialize].
func Serialize(x any, b []byte) []byte {
	s := newSerializer()
	w := &x // w is *interface{}
	wr := reflect.ValueOf(w)
	p := wr.UnsafePointer() // *interface{}
	t := wr.Elem().Type()   // what x contains

	scan(s, t, p)
	// scan dirties s.scanptrs, so clean it up.
	clear(s.scanptrs)

	//	s.regions.Dump()

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

// Serializable values can be manually serialized to bytes. Types that implement
// this interface are serialized with the MarshalAppend method and deserialized
// with Unmarshal, instead of the built-in decoders.
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
// system.
//
// coroc usually generates calls to this function. It should be called in an
// init function so that types are always registered in the same order.
//
// Types are recursively added, as well as *T.
func RegisterType[T any]() {
	tm.add(reflect.TypeOf((*T)(nil)).Elem())
}

// RegisterTypeWithSerde is the same as [RegisterType] but assigns serialization
// and deserialization for this type.
func RegisterTypeWithSerde[T any](
	serializer func(*T, []byte) ([]byte, error),
	deserializer func(*T, []byte) ([]byte, error)) {

	RegisterType[T]()
	t := reflect.TypeOf((*T)(nil)).Elem()

	s := func(p unsafe.Pointer, b []byte) []byte {
		b, err := serializer((*T)(p), b)
		if err != nil {
			panic(fmt.Errorf("serializing %s: %w", t, err))
		}
		return b
	}

	d := func(p unsafe.Pointer, b []byte) []byte {
		b, err := deserializer((*T)(p), b)
		if err != nil {
			panic(fmt.Errorf("deserializing %s: %w", t, err))
		}
		return b
	}

	tm.attach(t, s, d)
}

type deserializer struct {
	// TODO: make it a slice since pointer ids is the sequence of integers
	// starting at 1.
	ptrs map[sID]unsafe.Pointer
}

func newDeserializer() *deserializer {
	return &deserializer{
		ptrs: make(map[sID]unsafe.Pointer),
	}
}

func (d *deserializer) readPtr(b []byte) (unsafe.Pointer, sID, []byte) {
	x, n := binary.Varint(b)
	i := sID(x)
	p := d.ptrs[i]

	slog.Debug("Deserializer ReadPtr", "i", i, "p", p, "n", n)
	return p, i, b[n:]
}

func (d *deserializer) store(i sID, p unsafe.Pointer) {
	if d.ptrs[i] != nil {
		panic(fmt.Errorf("trying to overwirte known ID %d with %p", i, p))
	}
	d.ptrs[i] = p
}

// serializer holds the state for serialization.
//
// The ptrs value maps from pointers to IDs. Each time the serialization process
// encounters a pointer, it assigns it a new unique ID for its given address.
// This mechanism allows writing shared data only once. The actual value is
// written the first time a given pointer ID is encountered.
//
// The regions value contains ranges of memory held by container types. They are
// the values that actually own memory: basic types (bool, numbers), structs,
// and arrays.
//
// Serialization starts with scanning the graph of values to find all the
// containers and add the range of memory they occupy into the map. Regions
// belong to the outermost container. For example:
//
//	struct X {
//	  struct Y {
//	    int
//	  }
//	}
//
// creates only one region: the struct X. Both struct Y and the int are
// containers, but they are included in the region of struct X.
//
// Those two mechanisms allow the deserialization of pointers that point to
// shared memory. Only outermost containers are serialized. All pointers either
// point to a container, or an offset into that container.
type serializer struct {
	ptrs    map[unsafe.Pointer]sID
	regions regions

	// TODO: move out. just used temporarily by scan
	scanptrs map[reflect.Value]struct{}
}

func newSerializer() *serializer {
	return &serializer{
		ptrs:     make(map[unsafe.Pointer]sID),
		scanptrs: make(map[reflect.Value]struct{}),
	}
}

func serializeVarint(size int, b []byte) []byte {
	return binary.AppendVarint(b, int64(size))
}

func deserializeVarint(b []byte) (int, []byte) {
	l, n := binary.Varint(b)
	return int(l), b[n:]
}

// A type is serialized as follow:
//
// - No type (t is nil) => varint(0)
// - Any type but array => varint(1-MaxInt)
// - Array type [X]T    => varint(-1) varint(X) [serialize T]
//
// This is so that we can represent slices as pointers to arrays, with a size
// not known at compile time (so precise array type hasn't been registered.
func serializeType(t reflect.Type, b []byte) []byte {
	if t == nil {
		return serializeVarint(0, b)
	}

	if t.Kind() != reflect.Array {
		return serializeVarint(int(tm.idOf(t)), b)
	}

	b = serializeVarint(-1, b)
	b = serializeVarint(t.Len(), b)
	return serializeType(t.Elem(), b)
}

func deserializeType(b []byte) (reflect.Type, []byte) {
	n, b := deserializeVarint(b)
	if n == 0 {
		return nil, b
	}

	if n > 0 {
		return tm.typeOf(sID(n)), b
	}

	if n != -1 {
		panic(fmt.Errorf("unknown type first int: %d", n))
	}

	l, b := deserializeVarint(b)
	et, b := deserializeType(b)
	return reflect.ArrayOf(l, et), b
}

// Used for unsafe access to internals of interface{} and reflect.Value.
type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// Used instead of reflect.SliceHeader to use an unsafe.Pointer instead of
// uintptr.
type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

var (
	serializableT = reflect.TypeOf((*Serializable)(nil)).Elem()
	byteT         = reflect.TypeOf(byte(0))
)

func serializeAny(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	if serde, ok := tm.serdeOf(t); ok {
		return serde.ser(p, b)
	}

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
		return s.SerializeBool(*(*bool)(p), b)
	case reflect.Int:
		return s.SerializeInt(*(*int)(p), b)
	case reflect.Int64:
		return s.SerializeInt64(*(*int64)(p), b)
	case reflect.Int32:
		return s.SerializeInt32(*(*int32)(p), b)
	case reflect.Int16:
		return s.SerializeInt16(*(*int16)(p), b)
	case reflect.Int8:
		return s.SerializeInt8(*(*int8)(p), b)
	case reflect.Uint:
		return s.SerializeUint(*(*uint)(p), b)
	case reflect.Uint64:
		return s.SerializeUint64(*(*uint64)(p), b)
	case reflect.Uint32:
		return s.SerializeUint32(*(*uint32)(p), b)
	case reflect.Uint16:
		return s.SerializeUint16(*(*uint16)(p), b)
	case reflect.Uint8:
		return s.SerializeUint8(*(*uint8)(p), b)
	case reflect.Float64:
		return s.SerializeFloat64(*(*float64)(p), b)
	case reflect.Float32:
		return s.SerializeFloat32(*(*float32)(p), b)
	case reflect.Complex64:
		return s.SerializeComplex64(*(*complex64)(p), b)
	case reflect.Complex128:
		return s.SerializeComplex128(*(*complex128)(p), b)
	case reflect.String:
		return s.SerializeString((*string)(p), b)
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
	if serde, ok := tm.serdeOf(t); ok {
		return serde.des(p, b)
	}

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
		return d.DeserializeBool((*bool)(p), b)
	case reflect.Int:
		return d.DeserializeInt((*int)(p), b)
	case reflect.Int64:
		return d.DeserializeInt64((*int64)(p), b)
	case reflect.Int32:
		return d.DeserializeInt32((*int32)(p), b)
	case reflect.Int16:
		return d.DeserializeInt16((*int16)(p), b)
	case reflect.Int8:
		return d.DeserializeInt8((*int8)(p), b)
	case reflect.Uint:
		return d.DeserializeUint((*uint)(p), b)
	case reflect.Uint64:
		return d.DeserializeUint64((*uint64)(p), b)
	case reflect.Uint32:
		return d.DeserializeUint32((*uint32)(p), b)
	case reflect.Uint16:
		return d.DeserializeUint16((*uint16)(p), b)
	case reflect.Uint8:
		return d.DeserializeUint8((*uint8)(p), b)
	case reflect.Float64:
		return d.DeserializeFloat64((*float64)(p), b)
	case reflect.Float32:
		return d.DeserializeFloat32((*float32)(p), b)
	case reflect.Complex64:
		return d.DeserializeComplex64((*complex64)(p), b)
	case reflect.Complex128:
		return d.DeserializeComplex128((*complex128)(p), b)
	case reflect.String:
		return d.DeserializeString((*string)(p), b)
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

// Returns true if it created a new ID (false if reused one).
func (s *serializer) assignPointerID(p unsafe.Pointer) (sID, bool) {
	id, ok := s.ptrs[p]
	if !ok {
		id = sID(len(s.ptrs) + 1)
		s.ptrs[p] = id
	}
	return id, !ok
}

func serializePointedAt(s *serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	//	fmt.Printf("Serialize pointed at: %d (%s)\n", p, t)
	// If this is a nil pointer, write it as such.
	if p == nil {
		//		fmt.Printf("\t=>NIL\n")
		return serializeVarint(0, b)
	}

	id, new := s.assignPointerID(p)
	b = serializeVarint(int(id), b)
	//	fmt.Printf("\t=>Assigned ID %d\n", id)
	if !new {
		//		fmt.Printf("\t=>Already seen\n")
		// This exact pointer has already been serialized. Write its ID
		// and move on.
		return b
	}
	//	fmt.Printf("\t=>New pointer\n")

	// Now, this is pointer that is seen for the first time.

	// Check the region of this pointer.
	r := s.regions.regionOf(p)

	// If this pointer does not belong to any region or is the container of
	// the region, write a negative offset to flag it is on its own, and
	// write its data.
	if !r.valid() || (r.offset(p) == 0 && t == r.typ) {
		//		fmt.Printf("\t=>Is container (region %t)\n", r.Valid())
		b = serializeVarint(-1, b)
		return serializeAny(s, t, p, b)
	}

	// The pointer points into a memory region.
	offset := r.offset(p)
	b = serializeVarint(offset, b)

	//	fmt.Printf("\t=>Offset in container: %d\n", offset)

	// Write the type of the container.
	b = serializeType(r.typ, b)
	//	fmt.Printf("\t=>Container at: %d (%s)\n", r.Pointer(), r.Type())

	// Serialize the parent.
	return serializePointedAt(s, r.typ, r.start, b)
}

func deserializePointedAt(d *deserializer, t reflect.Type, b []byte) (reflect.Value, []byte) {

	// This function is a bit different than the other deserialize* ones
	// because it deserializes into an unknown location. As a result,
	// instead of taking an unsafe.Pointer as an input, it returns a
	// reflect.Value that contains a *T (where T is given by the argument
	// t).

	//	fmt.Printf("Deserialize pointed at: %s\n", t)

	ptr, id, b := d.readPtr(b)
	//	fmt.Printf("\t=> ptr=%d, id=%d\n", ptr, id)
	if ptr != nil || id == 0 { // pointer already seen or nil
		//		fmt.Printf("\t=>Returning existing data\n")
		return reflect.NewAt(t, ptr), b
	}

	offset, b := deserializeVarint(b)
	//	fmt.Printf("\t=>Read offset %d\n", offset)

	// Negative offset means this is either a container or a standalone
	// value.
	if offset < 0 {
		e := reflect.New(t)
		ep := e.UnsafePointer()
		d.store(id, ep)
		//		fmt.Printf("\t=>Negative offset: container %d\n", ep)
		return e, deserializeAny(d, t, ep, b)
	}

	// This pointer points into a container. Deserialize that one first,
	// then return the pointer itself with an offset.
	ct, b := deserializeType(b)

	//	fmt.Printf("\t=>Container type: %s\n", ct)

	// cp is a pointer to the container
	cp, b := deserializePointedAt(d, ct, b)

	// Create the pointer with an offset into the container.
	ep := unsafe.Add(cp.UnsafePointer(), offset)
	r := reflect.NewAt(t, ep)
	d.store(id, ep)
	//	fmt.Printf("\t=>Returning id=%d ep=%d\n", id, ep)
	return r, b
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
	n, b := deserializeVarint(b)
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

	b = serializeVarint(r.Len(), b)
	b = serializeVarint(r.Cap(), b)

	at := reflect.ArrayOf(r.Cap(), t.Elem())
	ap := r.UnsafePointer()

	b = serializePointedAt(s, at, ap, b)

	return b
}

func deserializeSlice(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	l, b := deserializeVarint(b)
	c, b := deserializeVarint(b)

	at := reflect.ArrayOf(c, t.Elem())
	ar, b := deserializePointedAt(d, at, b)

	if ar.IsNil() {
		return b
	}

	s := (*slice)(p)
	s.data = ar.UnsafePointer()
	s.cap = c
	s.len = l
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

	if i.typ == nil {
		return serializeType(nil, b)
	}

	x := *(*interface{})(p)
	et := reflect.TypeOf(x)
	b = serializeType(et, b)

	eptr := i.ptr
	if inlined(et) {
		xp := i.ptr
		eptr = unsafe.Pointer(&xp)
		// noescape?
	}

	return serializePointedAt(s, et, eptr, b)
}

func deserializeInterface(d *deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	// Deserialize the type
	et, b := deserializeType(b)
	if et == nil {
		return b
	}

	// Deserialize the pointer
	ep, b := deserializePointedAt(d, et, b)

	// Store the result in the interface
	r := reflect.NewAt(t, p)
	r.Elem().Set(ep.Elem())

	return b
}

func (s *serializer) SerializeString(x *string, b []byte) []byte {
	// Serialize string as a size and a pointer to an array of bytes.

	l := len(*x)
	b = serializeVarint(l, b)

	if l == 0 {
		return b
	}

	at := reflect.ArrayOf(l, byteT)
	ap := unsafe.Pointer(unsafe.StringData(*x))

	return serializePointedAt(s, at, ap, b)
}

func (d *deserializer) DeserializeString(x *string, b []byte) []byte {
	l, b := deserializeVarint(b)

	if l == 0 {
		return b
	}

	at := reflect.ArrayOf(l, byteT)
	ar, b := deserializePointedAt(d, at, b)

	*x = unsafe.String((*byte)(ar.UnsafePointer()), l)
	return b
}

func (s *serializer) SerializeBool(x bool, b []byte) []byte {
	c := byte(0)
	if x {
		c = 1
	}
	return append(b, c)
}

func (d *deserializer) DeserializeBool(x *bool, b []byte) []byte {
	*x = b[0] == 1
	return b[1:]
}

func (s *serializer) SerializeInt(x int, b []byte) []byte {
	return s.SerializeInt64(int64(x), b)
}

func (d *deserializer) DeserializeInt(x *int, b []byte) []byte {
	*x = int(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func (s *serializer) SerializeInt64(x int64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, uint64(x))
}

func (d *deserializer) DeserializeInt64(x *int64, b []byte) []byte {
	*x = int64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func (s *serializer) SerializeInt32(x int32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, uint32(x))
}

func (d *deserializer) DeserializeInt32(x *int32, b []byte) []byte {
	*x = int32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func (s *serializer) SerializeInt16(x int16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, uint16(x))
}

func (d *deserializer) DeserializeInt16(x *int16, b []byte) []byte {
	*x = int16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func (s *serializer) SerializeInt8(x int8, b []byte) []byte {
	return append(b, byte(x))
}

func (d *deserializer) DeserializeInt8(x *int8, b []byte) []byte {
	*x = int8(b[0])
	return b[1:]
}

func (s *serializer) SerializeUint(x uint, b []byte) []byte {
	return s.SerializeUint64(uint64(x), b)
}

func (d *deserializer) DeserializeUint(x *uint, b []byte) []byte {
	*x = uint(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func (s *serializer) SerializeUint64(x uint64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, x)
}

func (s *deserializer) DeserializeUint64(x *uint64, b []byte) []byte {
	*x = uint64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func (s *serializer) SerializeUint32(x uint32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, x)
}

func (d *deserializer) DeserializeUint32(x *uint32, b []byte) []byte {
	*x = uint32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func (s *serializer) SerializeUint16(x uint16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, x)
}

func (d *deserializer) DeserializeUint16(x *uint16, b []byte) []byte {
	*x = uint16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func (s *serializer) SerializeUint8(x uint8, b []byte) []byte {
	return append(b, byte(x))
}

func (d *deserializer) DeserializeUint8(x *uint8, b []byte) []byte {
	*x = uint8(b[0])
	return b[1:]
}

func (s *serializer) SerializeFloat32(x float32, b []byte) []byte {
	return s.SerializeUint32(math.Float32bits(x), b)
}

func (d *deserializer) DeserializeFloat32(x *float32, b []byte) []byte {
	return d.DeserializeUint32((*uint32)(unsafe.Pointer(x)), b)
}

func (s *serializer) SerializeFloat64(x float64, b []byte) []byte {
	return s.SerializeUint64(math.Float64bits(x), b)
}

func (d *deserializer) DeserializeFloat64(x *float64, b []byte) []byte {
	return d.DeserializeUint64((*uint64)(unsafe.Pointer(x)), b)
}

func (s *serializer) SerializeComplex64(x complex64, b []byte) []byte {
	b = s.SerializeFloat32(real(x), b)
	b = s.SerializeFloat32(imag(x), b)
	return b
}

func (d *deserializer) DeserializeComplex64(x *complex64, b []byte) []byte {
	type complex64 struct {
		real float32
		img  float32
	}
	p := (*complex64)(unsafe.Pointer(x))
	b = d.DeserializeFloat32(&p.real, b)
	b = d.DeserializeFloat32(&p.img, b)
	return b
}

func (s *serializer) SerializeComplex128(x complex128, b []byte) []byte {
	b = s.SerializeFloat64(real(x), b)
	b = s.SerializeFloat64(imag(x), b)
	return b
}

func (d *deserializer) DeserializeComplex128(x *complex128, b []byte) []byte {
	type complex128 struct {
		real float64
		img  float64
	}
	p := (*complex128)(unsafe.Pointer(x))
	b = d.DeserializeFloat64(&p.real, b)
	b = d.DeserializeFloat64(&p.img, b)
	return b
}

// returns true iff type t would be inlined in an interface.
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

type serializerFn func(p unsafe.Pointer, b []byte) []byte
type deserializerFn func(p unsafe.Pointer, b []byte) []byte

type serde struct {
	ser serializerFn
	des deserializerFn
}

type typeMap struct {
	byID   map[sID]reflect.Type
	byType map[reflect.Type]sID
	serdes map[reflect.Type]serde
}

func newTypeMap() *typeMap {
	return &typeMap{
		byID:   make(map[sID]reflect.Type),
		byType: make(map[reflect.Type]sID),
		serdes: make(map[reflect.Type]serde),
	}
}

func (m *typeMap) addExact(t reflect.Type) {
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

func (m *typeMap) attach(t reflect.Type, ser serializerFn, des deserializerFn) {
	if ser == nil || des == nil {
		panic("both serializer and deserializer need to be provided")
	}

	_, ok := m.byType[t]
	if !ok {
		panic(fmt.Errorf("register type %s before attaching serde", t))
	}

	m.serdes[t] = serde{ser: ser, des: des}
}

func (m *typeMap) add(t reflect.Type) {
	if m.exists(t) {
		return
	}
	m.addExact(t)
	m.addExact(reflect.PointerTo(t))

	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		m.add(t.Elem())
	case reflect.Map:
		m.add(t.Key())
		m.add(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			m.add(t.Field(i).Type)
		}
	}
}

func (m *typeMap) idOf(x reflect.Type) sID {
	id, ok := m.byType[x]
	if !ok {
		panic(fmt.Errorf("type '%s' is not registered", x))
	}
	return id
}

func (m *typeMap) typeOf(x sID) reflect.Type {
	t, ok := m.byID[x]
	if !ok {
		panic(fmt.Errorf("type id '%d' not registered", x))
	}
	return t
}

func (m *typeMap) serdeOf(x reflect.Type) (serde, bool) {
	s, ok := m.serdes[x]
	return s, ok
}

// Global type register.
var tm *typeMap = newTypeMap()

type regions []region

func (r *regions) dump() {
	fmt.Println("========== MEMORY REGIONS ==========")
	fmt.Println("Found", len(*r), "regions.")
	for i, r := range *r {
		fmt.Printf("#%d: [%d-%d[ %d %s\n", i, r.start, r.end, r.size(), r.typ)
	}
	fmt.Println("====================================")
}

// debug function to ensure the state hold its invariants. panic if they don't.
func (r *regions) validate() {
	s := *r
	if len(s) == 0 {
		return
	}

	for i := 0; i < len(s); i++ {
		if uintptr(s[i].start) >= uintptr(s[i].end) {
			panic(fmt.Errorf("region #%d has invalid bounds: start=%d end=%d delta=%d", i, s[i].start, s[i].end, s[i].size()))
		}
		if s[i].typ == nil {
			panic(fmt.Errorf("region #%d has nil type", i))
		}
		if i == 0 {
			continue
		}
		if uintptr(s[i].start) < uintptr(s[i-1].end) {
			r.dump()
			panic(fmt.Errorf("region #%d and #%d overlap", i-1, i))
		}
	}
}

// size computes the amount of bytes coverred by all known regions.
func (r *regions) size() int {
	n := 0
	for _, r := range *r {
		n += r.size()
	}
	return n
}

func (r *regions) regionOf(p unsafe.Pointer) region {
	//	fmt.Printf("Searching regions for %d\n", p)
	addr := uintptr(p)
	s := *r
	if len(s) == 0 {
		//		fmt.Printf("\t=> No regions\n")
		return region{}
	}

	i := sort.Search(len(s), func(i int) bool {
		return uintptr(s[i].start) >= addr
	})
	//	fmt.Printf("\t=> i = %d\n", i)

	if i < len(s) && uintptr(s[i].start) == addr {
		return s[i]
	}

	if i > 0 {
		i--
	}
	if uintptr(s[i].start) > addr || uintptr(s[i].end) <= addr {
		return region{}
	}
	return s[i]

}

func (r *regions) add(t reflect.Type, start unsafe.Pointer) {
	size := t.Size()
	if size == 0 {
		return
	}

	end := unsafe.Add(start, size)

	//	fmt.Printf("Adding [%d-%d[ %d %s\n", startAddr, endAddr, endAddr-startAddr, t)
	startSize := r.size()
	defer func() {
		//r.Dump()
		r.validate()
		endSize := r.size()
		if endSize < startSize {
			panic(fmt.Errorf("regions shrunk (%d -> %d)", startSize, endSize))
		}
	}()

	s := *r

	if len(s) == 0 {
		*r = append(s, region{
			start: start,
			end:   end,
			typ:   t,
		})
		return
	}

	// Invariants:
	// (1) len(s) > 0
	// (2) s is sorted by start address
	// (3) s contains no overlapping range

	i := sort.Search(len(s), func(i int) bool {
		return uintptr(s[i].start) >= uintptr(start)
	})
	//fmt.Println("\ti =", i)

	if i < len(s) && uintptr(s[i].start) == uintptr(start) {
		// Pointer is present in the set. If it's contained in the
		// region that already exists, we are done.
		if uintptr(s[i].end) >= uintptr(end) {
			return
		}

		// Otherwise extend the region.
		s[i].end = end
		s[i].typ = t

		// To maintain invariant (3), keep extending the selected region
		// until it becomes the last one or the next range is disjoint.
		r.extend(i)
		return
	}
	// Pointer did not point to the beginning of a region.

	// Attempt to grow the previous region.
	if i > 0 {
		if uintptr(start) < uintptr(s[i-1].end) {
			if uintptr(end) > uintptr(s[i-1].end) {
				s[i-1].end = end
				r.extend(i - 1)
			}
			return
		}
	}

	// Attempt to grow the next region.
	if i+1 < len(s) {
		if uintptr(end) > uintptr(s[i+1].start) {
			s[i+1].start = start
			if uintptr(end) > uintptr(s[i+1].end) {
				s[i+1].end = end
			}
			s[i+1].typ = t
			r.extend(i + 1)
			return
		}
	}

	// Just insert it.
	s = append(s, region{})
	copy(s[i+1:], s[i:])
	s[i] = region{start: start, end: end, typ: t}
	*r = s
	r.extend(i)
}

// extend attempts to grow region i by swallowing any region after it, as long
// as it would make one continous region. It is used after a modification of
// region i to maintain the invariants.
func (r *regions) extend(i int) {
	s := *r
	grown := 0
	for next := i + 1; next < len(s) && uintptr(s[i].end) > uintptr(s[next].start); next++ {
		s[i].end = s[next].end
		grown++
	}
	copy(s[i+1:], s[i+1+grown:])
	*r = s[:len(s)-grown]
}

type region struct {
	start unsafe.Pointer // inclusive
	end   unsafe.Pointer // exclusive
	typ   reflect.Type
}

func (r region) valid() bool {
	return r.typ != nil
}

func (r region) size() int {
	return int(uintptr(r.end) - uintptr(r.start))
}

func (r region) offset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(r.start))
}

// scan the value of type t at address p recursively to build up the serializer
// state with necessary information for encoding. At the moment it only creates
// the memory regions table.
//
// It uses s.scanptrs to track which pointers it has already visited to avoid
// infinite loops. It does not clean it up after. I'm sure there is something
// more useful we could do with that.
func scan(s *serializer, t reflect.Type, p unsafe.Pointer) {
	if p == nil {
		return
	}

	r := reflect.NewAt(t, p)
	if _, ok := s.scanptrs[r]; ok {
		return
	}
	s.scanptrs[r] = struct{}{}

	switch t.Kind() {
	case reflect.Invalid:
		panic("handling invalid reflect.Type")
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		s.regions.add(t, p)
	case reflect.Array:
		s.regions.add(t, p)
		et := t.Elem()
		es := int(et.Size())
		for i := 0; i < t.Len(); i++ {
			ep := unsafe.Add(p, es*i)
			scan(s, et, ep)
		}
	case reflect.Slice:
		sr := r.Elem()
		ep := sr.UnsafePointer()
		if ep == nil {
			return
		}
		// Estimate size of backing array.
		et := t.Elem()
		es := int(et.Size())

		// Create a new type for the backing array.
		xt := reflect.ArrayOf(sr.Cap(), t.Elem())
		s.regions.add(xt, ep)
		for i := 0; i < sr.Len(); i++ {
			ep := unsafe.Add(ep, es*i)
			scan(s, et, ep)
		}
	case reflect.Interface:
		x := *(*interface{})(p)
		et := reflect.TypeOf(x)
		eptr := (*iface)(p).ptr
		if eptr == nil {
			return
		}
		if inlined(et) {
			xp := (*iface)(p).ptr
			eptr = unsafe.Pointer(&xp)
		}

		scan(s, et, eptr)
	case reflect.Struct:
		s.regions.add(t, p)
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			ft := f.Type
			fp := unsafe.Add(p, f.Offset)
			scan(s, ft, fp)
		}
	case reflect.Pointer:
		ep := r.Elem().UnsafePointer()
		scan(s, t.Elem(), ep)
	case reflect.String:
		str := *(*string)(p)
		sp := unsafe.StringData(str)
		xt := reflect.ArrayOf(len(str), byteT)
		s.regions.add(xt, unsafe.Pointer(sp))
	case reflect.Map:
		m := r.Elem()
		if m.IsNil() || m.Len() == 0 {
			return
		}
		kt := t.Key()
		vt := t.Elem()
		iter := m.MapRange()
		for iter.Next() {
			k := iter.Key()
			kp := (*iface)(unsafe.Pointer(&k)).ptr
			scan(s, kt, kp)

			v := iter.Value()
			vp := (*iface)(unsafe.Pointer(&v)).ptr
			scan(s, vt, vp)
		}

	default:
		// TODO:
		// Chan
		// Func
		// UnsafePointer
	}
}
