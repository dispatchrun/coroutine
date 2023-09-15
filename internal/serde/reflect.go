package serde

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

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
		return serializeVarint(int(Types.idOf(t)), b)
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
		return Types.typeOf(sID(n)), b
	}

	if n != -1 {
		panic(fmt.Errorf("unknown type first int: %d", n))
	}

	l, b := deserializeVarint(b)
	et, b := deserializeType(b)
	return reflect.ArrayOf(l, et), b
}

var (
	byteT = reflect.TypeOf(byte(0))
)

func serializeAny(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	if serde, ok := Types.serdeOf(t); ok {
		return serde.ser(p, b)
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't serialize reflect.Invalid"))
	case reflect.Bool:
		return SerializeBool(s, *(*bool)(p), b)
	case reflect.Int:
		return SerializeInt(s, *(*int)(p), b)
	case reflect.Int64:
		return SerializeInt64(s, *(*int64)(p), b)
	case reflect.Int32:
		return SerializeInt32(s, *(*int32)(p), b)
	case reflect.Int16:
		return SerializeInt16(s, *(*int16)(p), b)
	case reflect.Int8:
		return SerializeInt8(s, *(*int8)(p), b)
	case reflect.Uint:
		return SerializeUint(s, *(*uint)(p), b)
	case reflect.Uint64:
		return SerializeUint64(s, *(*uint64)(p), b)
	case reflect.Uint32:
		return SerializeUint32(s, *(*uint32)(p), b)
	case reflect.Uint16:
		return SerializeUint16(s, *(*uint16)(p), b)
	case reflect.Uint8:
		return SerializeUint8(s, *(*uint8)(p), b)
	case reflect.Float64:
		return SerializeFloat64(s, *(*float64)(p), b)
	case reflect.Float32:
		return SerializeFloat32(s, *(*float32)(p), b)
	case reflect.Complex64:
		return SerializeComplex64(s, *(*complex64)(p), b)
	case reflect.Complex128:
		return SerializeComplex128(s, *(*complex128)(p), b)
	case reflect.String:
		return SerializeString(s, (*string)(p), b)
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
	if serde, ok := Types.serdeOf(t); ok {
		return serde.des(p, b)
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't deserialize reflect.Invalid"))
	case reflect.Bool:
		return DeserializeBool(d, (*bool)(p), b)
	case reflect.Int:
		return DeserializeInt(d, (*int)(p), b)
	case reflect.Int64:
		return DeserializeInt64(d, (*int64)(p), b)
	case reflect.Int32:
		return DeserializeInt32(d, (*int32)(p), b)
	case reflect.Int16:
		return DeserializeInt16(d, (*int16)(p), b)
	case reflect.Int8:
		return DeserializeInt8(d, (*int8)(p), b)
	case reflect.Uint:
		return DeserializeUint(d, (*uint)(p), b)
	case reflect.Uint64:
		return DeserializeUint64(d, (*uint64)(p), b)
	case reflect.Uint32:
		return DeserializeUint32(d, (*uint32)(p), b)
	case reflect.Uint16:
		return DeserializeUint16(d, (*uint16)(p), b)
	case reflect.Uint8:
		return DeserializeUint8(d, (*uint8)(p), b)
	case reflect.Float64:
		return DeserializeFloat64(d, (*float64)(p), b)
	case reflect.Float32:
		return DeserializeFloat32(d, (*float32)(p), b)
	case reflect.Complex64:
		return DeserializeComplex64(d, (*complex64)(p), b)
	case reflect.Complex128:
		return DeserializeComplex128(d, (*complex128)(p), b)
	case reflect.String:
		return DeserializeString(d, (*string)(p), b)
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

func serializePointedAt(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func deserializePointedAt(d *Deserializer, t reflect.Type, b []byte) (reflect.Value, []byte) {
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

func serializeSlice(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p).Elem()

	b = serializeVarint(r.Len(), b)
	b = serializeVarint(r.Cap(), b)

	at := reflect.ArrayOf(r.Cap(), t.Elem())
	ap := r.UnsafePointer()

	b = serializePointedAt(s, at, ap, b)

	return b
}

func deserializeSlice(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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
	return serializePointedAt(s, t.Elem(), x, b)
}

func deserializePointer(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	ep, b := deserializePointedAt(d, t.Elem(), b)
	r := reflect.NewAt(t, p)
	r.Elem().Set(ep)
	return b
}

func serializeStruct(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		fp := unsafe.Add(p, ft.Offset)
		b = serializeAny(s, ft.Type, fp, b)
	}
	return b
}

func deserializeStruct(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		fp := unsafe.Add(p, ft.Offset)
		b = deserializeAny(d, ft.Type, fp, b)
	}
	return b
}

func serializeInterface(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func deserializeInterface(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
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

func SerializeString(s *Serializer, x *string, b []byte) []byte {
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

func DeserializeString(d *Deserializer, x *string, b []byte) []byte {
	l, b := deserializeVarint(b)

	if l == 0 {
		return b
	}

	at := reflect.ArrayOf(l, byteT)
	ar, b := deserializePointedAt(d, at, b)

	*x = unsafe.String((*byte)(ar.UnsafePointer()), l)
	return b
}

func SerializeBool(s *Serializer, x bool, b []byte) []byte {
	c := byte(0)
	if x {
		c = 1
	}
	return append(b, c)
}

func DeserializeBool(d *Deserializer, x *bool, b []byte) []byte {
	*x = b[0] == 1
	return b[1:]
}

func SerializeInt(s *Serializer, x int, b []byte) []byte {
	return SerializeInt64(s, int64(x), b)
}

func DeserializeInt(d *Deserializer, x *int, b []byte) []byte {
	*x = int(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func SerializeInt64(s *Serializer, x int64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, uint64(x))
}

func DeserializeInt64(d *Deserializer, x *int64, b []byte) []byte {
	*x = int64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func SerializeInt32(s *Serializer, x int32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, uint32(x))
}

func DeserializeInt32(d *Deserializer, x *int32, b []byte) []byte {
	*x = int32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func SerializeInt16(s *Serializer, x int16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, uint16(x))
}

func DeserializeInt16(d *Deserializer, x *int16, b []byte) []byte {
	*x = int16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func SerializeInt8(s *Serializer, x int8, b []byte) []byte {
	return append(b, byte(x))
}

func DeserializeInt8(d *Deserializer, x *int8, b []byte) []byte {
	*x = int8(b[0])
	return b[1:]
}

func SerializeUint(s *Serializer, x uint, b []byte) []byte {
	return SerializeUint64(s, uint64(x), b)
}

func DeserializeUint(d *Deserializer, x *uint, b []byte) []byte {
	*x = uint(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func SerializeUint64(s *Serializer, x uint64, b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, x)
}

func DeserializeUint64(d *Deserializer, x *uint64, b []byte) []byte {
	*x = uint64(binary.LittleEndian.Uint64(b[:8]))
	return b[8:]
}

func SerializeUint32(s *Serializer, x uint32, b []byte) []byte {
	return binary.LittleEndian.AppendUint32(b, x)
}

func DeserializeUint32(d *Deserializer, x *uint32, b []byte) []byte {
	*x = uint32(binary.LittleEndian.Uint32(b[:4]))
	return b[4:]
}

func SerializeUint16(s *Serializer, x uint16, b []byte) []byte {
	return binary.LittleEndian.AppendUint16(b, x)
}

func DeserializeUint16(d *Deserializer, x *uint16, b []byte) []byte {
	*x = uint16(binary.LittleEndian.Uint16(b[:2]))
	return b[2:]
}

func SerializeUint8(s *Serializer, x uint8, b []byte) []byte {
	return append(b, byte(x))
}

func DeserializeUint8(d *Deserializer, x *uint8, b []byte) []byte {
	*x = uint8(b[0])
	return b[1:]
}

func SerializeFloat32(s *Serializer, x float32, b []byte) []byte {
	return SerializeUint32(s, math.Float32bits(x), b)
}

func DeserializeFloat32(d *Deserializer, x *float32, b []byte) []byte {
	return DeserializeUint32(d, (*uint32)(unsafe.Pointer(x)), b)
}

func SerializeFloat64(s *Serializer, x float64, b []byte) []byte {
	return SerializeUint64(s, math.Float64bits(x), b)
}

func DeserializeFloat64(d *Deserializer, x *float64, b []byte) []byte {
	return DeserializeUint64(d, (*uint64)(unsafe.Pointer(x)), b)
}

func SerializeComplex64(s *Serializer, x complex64, b []byte) []byte {
	b = SerializeFloat32(s, real(x), b)
	b = SerializeFloat32(s, imag(x), b)
	return b
}

func DeserializeComplex64(d *Deserializer, x *complex64, b []byte) []byte {
	type complex64 struct {
		real float32
		img  float32
	}
	p := (*complex64)(unsafe.Pointer(x))
	b = DeserializeFloat32(d, &p.real, b)
	b = DeserializeFloat32(d, &p.img, b)
	return b
}

func SerializeComplex128(s *Serializer, x complex128, b []byte) []byte {
	b = SerializeFloat64(s, real(x), b)
	b = SerializeFloat64(s, imag(x), b)
	return b
}

func DeserializeComplex128(d *Deserializer, x *complex128, b []byte) []byte {
	type complex128 struct {
		real float64
		img  float64
	}
	p := (*complex128)(unsafe.Pointer(x))
	b = DeserializeFloat64(d, &p.real, b)
	b = DeserializeFloat64(d, &p.img, b)
	return b
}
