package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

func serializeType(s *Serializer, t reflect.Type) {
	x := s.types.ToType(t)
	serializeVarint(s, int(x))
}

func deserializeType(d *Deserializer) reflect.Type {
	id := deserializeVarint(d)
	return d.types.ToReflect(typeid(id))
}

func serializeAny(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	if serde, ok := s.serdes.serdeOf(t); ok {
		serde.ser(s, t, p)
		return
	}

	switch t {
	case reflectValueType:
		v := *(*reflect.Value)(p)
		serializeType(s, v.Type())
		serializeReflectValue(s, v.Type(), v)
		return
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't serialize reflect.Invalid"))
	case reflect.Bool:
		serializeBool(s, *(*bool)(p))
	case reflect.Int:
		serializeInt(s, *(*int)(p))
	case reflect.Int64:
		serializeInt64(s, *(*int64)(p))
	case reflect.Int32:
		serializeInt32(s, *(*int32)(p))
	case reflect.Int16:
		serializeInt16(s, *(*int16)(p))
	case reflect.Int8:
		serializeInt8(s, *(*int8)(p))
	case reflect.Uint:
		serializeUint(s, *(*uint)(p))
	case reflect.Uint64:
		serializeUint64(s, *(*uint64)(p))
	case reflect.Uint32:
		serializeUint32(s, *(*uint32)(p))
	case reflect.Uint16:
		serializeUint16(s, *(*uint16)(p))
	case reflect.Uint8:
		serializeUint8(s, *(*uint8)(p))
	case reflect.Uintptr:
		serializeUintptr(s, *(*uintptr)(p))
	case reflect.Float64:
		serializeFloat64(s, *(*float64)(p))
	case reflect.Float32:
		serializeFloat32(s, *(*float32)(p))
	case reflect.Complex64:
		serializeComplex64(s, *(*complex64)(p))
	case reflect.Complex128:
		serializeComplex128(s, *(*complex128)(p))
	case reflect.String:
		serializeString(s, (*string)(p))
	case reflect.Array:
		serializeArray(s, t, p)
	case reflect.Interface:
		serializeInterface(s, t, p)
	case reflect.Map:
		serializeMap(s, t, p)
	case reflect.Pointer:
		serializePointer(s, t, p)
	case reflect.UnsafePointer:
		serializeUnsafePointer(s, p)
	case reflect.Slice:
		serializeSlice(s, t, p)
	case reflect.Struct:
		serializeStruct(s, t, p)
	case reflect.Func:
		serializeFunc(s, t, p)
	// Chan
	default:
		panic(fmt.Errorf("reflection cannot serialize type %s", t))
	}
}

func deserializeAny(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	if serde, ok := d.serdes.serdeOf(t); ok {
		serde.des(d, t, p)
		return
	}

	switch t {
	case reflectValueType:
		rt := deserializeType(d)
		v := deserializeReflectValue(d, rt)
		reflect.NewAt(reflectValueType, p).Elem().Set(reflect.ValueOf(v))
		return
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't deserialize reflect.Invalid"))
	case reflect.Bool:
		deserializeBool(d, (*bool)(p))
	case reflect.Int:
		deserializeInt(d, (*int)(p))
	case reflect.Int64:
		deserializeInt64(d, (*int64)(p))
	case reflect.Int32:
		deserializeInt32(d, (*int32)(p))
	case reflect.Int16:
		deserializeInt16(d, (*int16)(p))
	case reflect.Int8:
		deserializeInt8(d, (*int8)(p))
	case reflect.Uint:
		deserializeUint(d, (*uint)(p))
	case reflect.Uint64:
		deserializeUint64(d, (*uint64)(p))
	case reflect.Uint32:
		deserializeUint32(d, (*uint32)(p))
	case reflect.Uint16:
		deserializeUint16(d, (*uint16)(p))
	case reflect.Uint8:
		deserializeUint8(d, (*uint8)(p))
	case reflect.Uintptr:
		deserializeUintptr(d, (*uintptr)(p))
	case reflect.Float64:
		deserializeFloat64(d, (*float64)(p))
	case reflect.Float32:
		deserializeFloat32(d, (*float32)(p))
	case reflect.Complex64:
		deserializeComplex64(d, (*complex64)(p))
	case reflect.Complex128:
		deserializeComplex128(d, (*complex128)(p))
	case reflect.String:
		deserializeString(d, (*string)(p))
	case reflect.Interface:
		deserializeInterface(d, t, p)
	case reflect.Pointer:
		deserializePointer(d, t, p)
	case reflect.UnsafePointer:
		deserializeUnsafePointer(d, p)
	case reflect.Array:
		deserializeArray(d, t, p)
	case reflect.Slice:
		deserializeSlice(d, t, p)
	case reflect.Map:
		deserializeMap(d, t, p)
	case reflect.Struct:
		deserializeStruct(d, t, p)
	case reflect.Func:
		deserializeFunc(d, t, p)
	default:
		panic(fmt.Errorf("reflection cannot deserialize type %s", t))
	}
}

var reflectValueType = reflect.TypeOf(reflect.Value{})

func serializeReflectValue(s *Serializer, t reflect.Type, v reflect.Value) {
	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't serialize reflect.Invalid"))
	case reflect.Bool:
		serializeBool(s, v.Bool())
	case reflect.Int:
		serializeInt(s, int(v.Int()))
	case reflect.Int8:
		serializeInt8(s, int8(v.Int()))
	case reflect.Int16:
		serializeInt16(s, int16(v.Int()))
	case reflect.Int32:
		serializeInt32(s, int32(v.Int()))
	case reflect.Int64:
		serializeInt64(s, v.Int())
	case reflect.Uint:
		serializeUint(s, uint(v.Uint()))
	case reflect.Uint8:
		serializeUint8(s, uint8(v.Uint()))
	case reflect.Uint16:
		serializeUint16(s, uint16(v.Uint()))
	case reflect.Uint32:
		serializeUint32(s, uint32(v.Uint()))
	case reflect.Uint64:
		serializeUint64(s, v.Uint())
	case reflect.Float32:
		serializeFloat32(s, float32(v.Float()))
	case reflect.Float64:
		serializeFloat64(s, v.Float())
	case reflect.Complex64:
		serializeComplex64(s, complex64(v.Complex()))
	case reflect.Complex128:
		serializeComplex128(s, complex128(v.Complex()))
	case reflect.String:
		str := v.String()
		serializeString(s, &str)
	case reflect.Array:
		et := t.Elem()
		for i := 0; i < t.Len(); i++ {
			serializeReflectValue(s, et, v.Index(i))
		}
	case reflect.Slice:
		sl := slice{data: v.UnsafePointer(), len: v.Len(), cap: v.Cap()}
		serializeSlice(s, t, unsafe.Pointer(&sl))
	case reflect.Map:
		serializeMapReflect(s, t, v)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				panic("not implemented: serializing reflect.Value(struct) with unexported fields")
			}
			serializeReflectValue(s, f.Type, v.Field(i))
		}
	case reflect.Func:
		if addr := v.Pointer(); addr != 0 {
			if fn := FuncByAddr(addr); fn != nil && fn.Closure != nil {
				panic("not implemented: serializing reflect.Value(closure)")
			}
			indirect := unsafe.Pointer(&addr)
			serializeFunc(s, t, unsafe.Pointer(&indirect))
		} else {
			serializeFunc(s, t, unsafe.Pointer(&addr))
		}
	case reflect.Pointer:
		serializePointedAt(s, t.Elem(), v.UnsafePointer())
	default:
		panic(fmt.Sprintf("not implemented: serializing reflect.Value with type %s (%s)", t, t.Kind()))
	}
}

func deserializeReflectValue(d *Deserializer, t reflect.Type) (v reflect.Value) {
	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't deserialize reflect.Invalid"))
	case reflect.Bool:
		var value bool
		deserializeBool(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Int:
		var value int
		deserializeInt(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Int8:
		var value int8
		deserializeInt8(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Int16:
		var value int16
		deserializeInt16(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Int32:
		var value int32
		deserializeInt32(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Int64:
		var value int64
		deserializeInt64(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Uint:
		var value uint
		deserializeUint(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Uint8:
		var value uint8
		deserializeUint8(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Uint16:
		var value uint16
		deserializeUint16(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Uint32:
		var value uint32
		deserializeUint32(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Uint64:
		var value uint64
		deserializeUint64(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Float32:
		var value float32
		deserializeFloat32(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Float64:
		var value float64
		deserializeFloat64(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Complex64:
		var value complex64
		deserializeComplex64(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Complex128:
		var value complex128
		deserializeComplex128(d, &value)
		v = reflect.ValueOf(value)
	case reflect.String:
		var value string
		deserializeString(d, &value)
		v = reflect.ValueOf(value)
	case reflect.Array:
		v = reflect.New(t).Elem()
		deserializeArray(d, t, unsafe.Pointer(v.UnsafeAddr()))
	case reflect.Slice:
		var value slice
		deserializeSlice(d, t, unsafe.Pointer(&value))
		v = reflect.New(t).Elem()
		*(*slice)(unsafe.Pointer(v.UnsafeAddr())) = value
	case reflect.Map:
		v = reflect.New(t).Elem()
		var p uintptr // FIXME: what should this be?
		deserializeMapReflect(d, t, v, unsafe.Pointer(&p))
	case reflect.Struct:
		v = reflect.New(t).Elem()
		for i := 0; i < t.NumField(); i++ {
			fv := deserializeReflectValue(d, t.Field(i).Type)
			v.Field(i).Set(fv)
		}
	case reflect.Func:
		var fn *Func
		deserializeFunc(d, t, unsafe.Pointer(&fn))
		v = reflect.New(t).Elem()
		if fn != nil {
			p := unsafe.Pointer(v.UnsafeAddr())
			*(*unsafe.Pointer)(p) = unsafe.Pointer(&fn.Addr)
		}
	case reflect.Pointer:
		ep := deserializePointedAt(d, t.Elem())
		v = reflect.New(t).Elem()
		v.Set(ep)
	default:
		panic(fmt.Sprintf("not implemented: deserializing reflect.Value with type %s", t))
	}
	return
}

func serializePointedAt(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	// If this is a nil pointer, write it as such.
	if p == nil {
		serializeVarint(s, 0)
		return
	}

	if static(p) {
		serializeVarint(s, -1)
		off := staticOffset(p)
		serializeVarint(s, off)
		return
	}

	id, new := s.assignPointerID(p)
	serializeVarint(s, int(id))
	if !new {
		// This exact pointer has already been serialized. Write its ID
		// and move on.
		return
	}

	// Now, this is pointer that is seen for the first time.

	// Check the region of this pointer.
	r := s.containers.of(p)

	// If this pointer does not belong to any region, write a negative
	// offset to flag it is on its own, and write its data.
	if !r.valid() {
		if t == nil {
			panic("cannot serialize unsafe.Pointer pointing to region of unknown size")
		}
		serializeVarint(s, -1)
		serializeAny(s, t, p)
		return
	}

	// The pointer points into a memory region.
	offset := int(r.offset(p))
	serializeVarint(s, offset)

	// Write the type of the container.
	serializeType(s, r.typ)

	// Serialize the parent. If offset is zero, we reuse the id to store the
	// parent. We could have a more compact representation here, but right
	// now we need this since the pointers <> id map in the serializer does
	// not discriminate between the container and the first element of it.
	if offset == 0 {
		serializeVarint(s, int(id))
		serializeVarint(s, -1)
		serializeAny(s, r.typ, r.addr)
		return
	}
	serializePointedAt(s, r.typ, r.addr)
}

func deserializePointedAt(d *Deserializer, t reflect.Type) reflect.Value {
	// This function is a bit different than the other deserialize* ones
	// because it deserializes into an unknown location. As a result,
	// instead of taking an unsafe.Pointer as an input, it returns a
	// reflect.Value that contains a *T (where T is given by the argument
	// t).

	ptr, id := d.readPtr()
	if ptr != nil || id == 0 { // pointer already seen or nil
		return reflect.NewAt(t, ptr)
	}

	offset := deserializeVarint(d)

	// Negative offset means this is either a container or a standalone
	// value.
	if offset < 0 {
		e := reflect.New(t)
		ep := e.UnsafePointer()
		d.store(id, ep)
		deserializeAny(d, t, ep)
		return e
	}

	// This pointer points into a container. Deserialize that one first,
	// then return the pointer itself with an offset.
	ct := deserializeType(d)

	// cp is a pointer to the container
	cp := deserializePointedAt(d, ct)

	// Create the pointer with an offset into the container.
	ep := unsafe.Add(cp.UnsafePointer(), offset)
	r := reflect.NewAt(t, ep)
	return r
}

func serializeMap(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	r := reflect.NewAt(t, p).Elem()
	serializeMapReflect(s, t, r)
}

func serializeMapReflect(s *Serializer, t reflect.Type, r reflect.Value) {
	if r.IsNil() {
		serializeVarint(s, 0)
		return
	}

	mapptr := r.UnsafePointer()

	id, new := s.assignPointerID(mapptr)
	serializeVarint(s, int(id))
	if !new {
		return
	}

	size := r.Len()

	serializeVarint(s, size)

	// TODO: allocs
	iter := r.MapRange()
	k := reflect.New(t.Key()).Elem()
	v := reflect.New(t.Elem()).Elem()
	for iter.Next() {
		k.Set(iter.Key())
		v.Set(iter.Value())
		serializeAny(s, t.Key(), k.Addr().UnsafePointer())
		serializeAny(s, t.Elem(), v.Addr().UnsafePointer())
	}
}

func deserializeMap(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	r := reflect.NewAt(t, p).Elem()
	deserializeMapReflect(d, t, r, p)
}

func deserializeMapReflect(d *Deserializer, t reflect.Type, r reflect.Value, p unsafe.Pointer) {
	ptr, id := d.readPtr()
	if id == 0 {
		// nil map
		return
	}
	if ptr != nil {
		// already deserialized at ptr
		existing := reflect.NewAt(t, ptr).Elem()
		r.Set(existing)
		return
	}

	n := deserializeVarint(d)
	if n < 0 { // nil map
		return
	}
	nv := reflect.MakeMapWithSize(t, n)
	r.Set(nv)
	d.store(id, p)
	for i := 0; i < n; i++ {
		k := reflect.New(t.Key())
		deserializeAny(d, t.Key(), k.UnsafePointer())
		v := reflect.New(t.Elem())
		deserializeAny(d, t.Elem(), v.UnsafePointer())
		r.SetMapIndex(k.Elem(), v.Elem())
	}
}

func serializeSlice(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	r := reflect.NewAt(t, p).Elem()

	serializeVarint(s, r.Len())
	serializeVarint(s, r.Cap())

	at := reflect.ArrayOf(r.Cap(), t.Elem())
	ap := r.UnsafePointer()

	serializePointedAt(s, at, ap)
}

func deserializeSlice(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	l := deserializeVarint(d)
	c := deserializeVarint(d)

	at := reflect.ArrayOf(c, t.Elem())
	ar := deserializePointedAt(d, at)

	if ar.IsNil() {
		return
	}

	s := (*slice)(p)
	s.data = ar.UnsafePointer()
	s.cap = c
	s.len = l
}

func serializeArray(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	n := t.Len()
	te := t.Elem()
	ts := int(te.Size())
	for i := 0; i < n; i++ {
		pe := unsafe.Add(p, ts*i)
		serializeAny(s, te, pe)
	}
}

func deserializeArray(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	size := int(t.Elem().Size())
	te := t.Elem()
	for i := 0; i < t.Len(); i++ {
		pe := unsafe.Add(p, size*i)
		deserializeAny(d, te, pe)
	}
}

func serializePointer(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	r := reflect.NewAt(t, p).Elem()
	x := r.UnsafePointer()
	serializePointedAt(s, t.Elem(), x)
}

func deserializePointer(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	ep := deserializePointedAt(d, t.Elem())
	r := reflect.NewAt(t, p)
	r.Elem().Set(ep)
}

func serializeUnsafePointer(s *Serializer, p unsafe.Pointer) {
	if p == nil {
		serializePointedAt(s, nil, nil)
	} else {
		serializePointedAt(s, nil, *(*unsafe.Pointer)(p))
	}
}

var unsafePointerType = reflect.TypeOf(unsafe.Pointer(nil))

func deserializeUnsafePointer(d *Deserializer, p unsafe.Pointer) {
	r := reflect.NewAt(unsafePointerType, p)

	ep := deserializePointedAt(d, unsafePointerType)
	if !ep.IsNil() {
		up := ep.UnsafePointer()
		r.Elem().Set(reflect.ValueOf(up))
	}
}

func serializeStruct(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	serializeStructFields(s, p, t.NumField(), t.Field)
}

func deserializeStruct(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	deserializeStructFields(d, p, t.NumField(), t.Field)
}

func serializeStructFields(s *Serializer, p unsafe.Pointer, n int, field func(int) reflect.StructField) {
	for i := 0; i < n; i++ {
		ft := field(i)
		fp := unsafe.Add(p, ft.Offset)
		serializeAny(s, ft.Type, fp)
	}
}

func deserializeStructFields(d *Deserializer, p unsafe.Pointer, n int, field func(int) reflect.StructField) {
	for i := 0; i < n; i++ {
		ft := field(i)
		fp := unsafe.Add(p, ft.Offset)
		deserializeAny(d, ft.Type, fp)
	}
}

func serializeFunc(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	// p is a pointer to a function value, function values are pointers to a
	// memory location starting with the address of the function, hence the
	// double indirection here.
	p = *(*unsafe.Pointer)(p)
	if p == nil {
		// Function IDs start at 1; use 0 to represent nil ptr.
		serializeVarint(s, 0)
		return
	}

	addr := *(*uintptr)(p)
	id, closure := s.funcs.RegisterAddr(addr)
	serializeVarint(s, int(id))

	if closure != nil {
		// Skip the first field, which is the function ptr.
		serializeStructFields(s, p, closure.NumField()-1, func(i int) reflect.StructField {
			return closure.Field(i + 1)
		})
	}
}

func deserializeFunc(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	id := deserializeVarint(d)
	if id == 0 {
		*(**Func)(p) = nil
		return
	}

	fn := d.funcs.ToFunc(funcid(id))
	if fn.Type == nil {
		panic(fn.Name + ": function type is missing")
	}
	if !t.AssignableTo(fn.Type) {
		panic(fn.Name + ": function type mismatch: " + fn.Type.String() + " != " + t.String())
	}

	if fn.Closure != nil {
		t := fn.Closure
		v := reflect.New(t)

		closure := v.UnsafePointer()
		*(*uintptr)(closure) = fn.Addr

		deserializeStructFields(d, closure, t.NumField()-1, func(i int) reflect.StructField {
			return t.Field(i + 1)
		})

		*(*unsafe.Pointer)(p) = closure
	} else {
		*(**Func)(p) = fn
	}
}

func serializeInterface(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	i := (*iface)(p)

	if i.typ == nil {
		serializeType(s, nil)
		return
	}

	et := reflect.TypeOf(reflect.NewAt(t, p).Elem().Interface())
	serializeType(s, et)

	eptr := i.ptr
	if inlined(et) {
		xp := i.ptr
		eptr = unsafe.Pointer(&xp)
		// noescape?
	}

	serializePointedAt(s, et, eptr)
}

func deserializeInterface(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	// Deserialize the type
	et := deserializeType(d)
	if et == nil {
		return
	}

	// Deserialize the pointer
	ep := deserializePointedAt(d, et)

	// Store the result in the interface
	r := reflect.NewAt(t, p)
	if !ep.IsNil() {
		r.Elem().Set(ep.Elem())
	} else {
		r.Elem().Set(reflect.Zero(et))
	}
}

func serializeString(s *Serializer, x *string) {
	// Serialize string as a size and a pointer to an array of bytes.

	l := len(*x)
	serializeVarint(s, l)

	if l == 0 {
		return
	}

	at := reflect.ArrayOf(l, byteT)
	ap := unsafe.Pointer(unsafe.StringData(*x))

	serializePointedAt(s, at, ap)
}

func deserializeString(d *Deserializer, x *string) {
	l := deserializeVarint(d)

	if l == 0 {
		return
	}

	at := reflect.ArrayOf(l, byteT)
	ar := deserializePointedAt(d, at)

	*x = unsafe.String((*byte)(ar.UnsafePointer()), l)
}

func serializeBool(s *Serializer, x bool) {
	c := byte(0)
	if x {
		c = 1
	}
	s.b = append(s.b, c)
}

func deserializeBool(d *Deserializer, x *bool) {
	*x = d.b[0] == 1
	d.b = d.b[1:]
}

func serializeInt(s *Serializer, x int) {
	serializeInt64(s, int64(x))
}

func deserializeInt(d *Deserializer, x *int) {
	*x = int(binary.LittleEndian.Uint64(d.b[:8]))
	d.b = d.b[8:]
}

func serializeInt64(s *Serializer, x int64) {
	s.b = binary.LittleEndian.AppendUint64(s.b, uint64(x))
}

func deserializeInt64(d *Deserializer, x *int64) {
	*x = int64(binary.LittleEndian.Uint64(d.b[:8]))
	d.b = d.b[8:]
}

func serializeInt32(s *Serializer, x int32) {
	s.b = binary.LittleEndian.AppendUint32(s.b, uint32(x))
}

func deserializeInt32(d *Deserializer, x *int32) {
	*x = int32(binary.LittleEndian.Uint32(d.b[:4]))
	d.b = d.b[4:]
}

func serializeInt16(s *Serializer, x int16) {
	s.b = binary.LittleEndian.AppendUint16(s.b, uint16(x))
}

func deserializeInt16(d *Deserializer, x *int16) {
	*x = int16(binary.LittleEndian.Uint16(d.b[:2]))
	d.b = d.b[2:]
}

func serializeInt8(s *Serializer, x int8) {
	s.b = append(s.b, byte(x))
}

func deserializeInt8(d *Deserializer, x *int8) {
	*x = int8(d.b[0])
	d.b = d.b[1:]
}

func serializeUint(s *Serializer, x uint) {
	serializeUint64(s, uint64(x))
}

func deserializeUint(d *Deserializer, x *uint) {
	*x = uint(binary.LittleEndian.Uint64(d.b[:8]))
	d.b = d.b[8:]
}

func serializeUint64(s *Serializer, x uint64) {
	s.b = binary.LittleEndian.AppendUint64(s.b, x)
}

func deserializeUint64(d *Deserializer, x *uint64) {
	*x = uint64(binary.LittleEndian.Uint64(d.b[:8]))
	d.b = d.b[8:]
}

func serializeUint32(s *Serializer, x uint32) {
	s.b = binary.LittleEndian.AppendUint32(s.b, x)
}

func deserializeUint32(d *Deserializer, x *uint32) {
	*x = uint32(binary.LittleEndian.Uint32(d.b[:4]))
	d.b = d.b[4:]
}

func serializeUint16(s *Serializer, x uint16) {
	s.b = binary.LittleEndian.AppendUint16(s.b, x)
}

func deserializeUint16(d *Deserializer, x *uint16) {
	*x = uint16(binary.LittleEndian.Uint16(d.b[:2]))
	d.b = d.b[2:]
}

func serializeUint8(s *Serializer, x uint8) {
	s.b = append(s.b, byte(x))
}

func deserializeUint8(d *Deserializer, x *uint8) {
	*x = uint8(d.b[0])
	d.b = d.b[1:]
}

func serializeUintptr(s *Serializer, x uintptr) {
	serializeUint64(s, uint64(x))
}

func deserializeUintptr(d *Deserializer, x *uintptr) {
	u := uint64(0)
	deserializeUint64(d, &u)
	*x = uintptr(u)
}

func serializeFloat32(s *Serializer, x float32) {
	serializeUint32(s, math.Float32bits(x))
}

func deserializeFloat32(d *Deserializer, x *float32) {
	deserializeUint32(d, (*uint32)(unsafe.Pointer(x)))
}

func serializeFloat64(s *Serializer, x float64) {
	serializeUint64(s, math.Float64bits(x))
}

func deserializeFloat64(d *Deserializer, x *float64) {
	deserializeUint64(d, (*uint64)(unsafe.Pointer(x)))
}

func serializeComplex64(s *Serializer, x complex64) {
	serializeFloat32(s, real(x))
	serializeFloat32(s, imag(x))
}

func deserializeComplex64(d *Deserializer, x *complex64) {
	type complex64 struct {
		real float32
		img  float32
	}
	p := (*complex64)(unsafe.Pointer(x))
	deserializeFloat32(d, &p.real)
	deserializeFloat32(d, &p.img)
}

func serializeComplex128(s *Serializer, x complex128) {
	serializeFloat64(s, real(x))
	serializeFloat64(s, imag(x))
}

func deserializeComplex128(d *Deserializer, x *complex128) {
	type complex128 struct {
		real float64
		img  float64
	}
	p := (*complex128)(unsafe.Pointer(x))
	deserializeFloat64(d, &p.real)
	deserializeFloat64(d, &p.img)
}

func typeof[X any]() reflect.Type {
	return reflect.TypeOf((*X)(nil)).Elem()
}

var byteT = typeof[byte]()
