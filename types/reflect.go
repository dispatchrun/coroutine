package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"

	coroutinev1 "github.com/dispatchrun/coroutine/gen/proto/go/coroutine/v1"
)

func serializeType(s *Serializer, t reflect.Type) {
	if t != nil && t.Kind() == reflect.Array {
		id := s.types.ToType(t.Elem())
		serializeVarint(s, int(id))
		serializeVarint(s, t.Len())
	} else {
		id := s.types.ToType(t)
		serializeVarint(s, int(id))
		serializeVarint(s, -1)
	}

}

func deserializeType(d *Deserializer) (reflect.Type, int) {
	id := deserializeVarint(d)
	length := deserializeVarint(d)
	t := d.types.ToReflect(typeid(id))
	return t, length
}

func serializeAny(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	if serde, ok := s.serdes.serdeByType(t); ok {
		offset := len(s.b)
		s.b = append(s.b, 0, 0, 0, 0, 0, 0, 0, 0) // store a 64-bit size placeholder
		serde.ser(s, t, p)
		binary.LittleEndian.PutUint64(s.b[offset:], uint64(len(s.b)-offset))
		return
	}

	switch t {
	case reflectValueT:
		v := *(*reflect.Value)(p)
		serializeType(s, v.Type())
		serializeReflectValue(s, v.Type(), v)
		return
	}

	v := reflect.NewAt(t, p).Elem()
	serializeReflectValue(s, t, v)
	return
}

func deserializeAny(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	if serde, ok := d.serdes.serdeByType(t); ok {
		d.b = d.b[8:] // skip size prefix
		serde.des(d, t, p)
		return
	}

	v := reflect.NewAt(t, p).Elem()

	switch t {
	case reflectValueT:
		rt, length := deserializeType(d)
		if length >= 0 {
			// We can't avoid the ArrayOf call here. We need to build a
			// reflect.Type in order to return a reflect.Value. The only
			// time this path is taken is if the user has explicitly serialized
			// a reflect.Value, or some other data type that contains or points
			// to a reflect.Value.
			rt = reflect.ArrayOf(length, rt)
		}
		rv := deserializeReflectValue(d, rt)
		v.Set(reflect.ValueOf(rv))
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
		if n := t.Len(); n > 0 {
			vi := v.Interface()
			p := ifacePtr(unsafe.Pointer(&vi), t)
			et := t.Elem()
			es := int(et.Size())
			for i := 0; i < n; i++ {
				e := unsafe.Add(p, es*i)
				serializeAny(s, et, e)
			}
		}
	case reflect.Slice:
		sl := slice{data: v.UnsafePointer(), len: v.Len(), cap: v.Cap()}
		serializeSlice(s, t, unsafe.Pointer(&sl))
	case reflect.Map:
		serializeMapReflect(s, v)
	case reflect.Struct:
		vi := v.Interface()
		p := ifacePtr(unsafe.Pointer(&vi), t)

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			// This is necessary to avoid the RO flag that's added when
			// accessing unexported fields via (*reflect.Value).Field.
			fp := unsafe.Add(p, f.Offset)
			serializeAny(s, f.Type, fp)
		}
	case reflect.Func:
		if addr := v.Pointer(); addr != 0 {
			vi := v.Interface()
			p := ifacePtr(unsafe.Pointer(&vi), t)
			serializeFunc(s, t, p)
		} else {
			serializeFunc(s, t, unsafe.Pointer(&addr))
		}
	case reflect.Pointer:
		serializePointedAt(s, t.Elem(), -1, v.UnsafePointer())
	case reflect.Interface:
		serializeInterface(s, v)
	case reflect.UnsafePointer:
		serializeUnsafePointer(s, v.UnsafePointer())
	default:
		// TODO: reflect.Chan
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
		deserializeArray(d, t, v.Addr().UnsafePointer())
	case reflect.Slice:
		var value slice
		deserializeSlice(d, t, unsafe.Pointer(&value))
		v = reflect.New(t).Elem()
		*(*slice)(unsafe.Pointer(v.UnsafeAddr())) = value
	case reflect.Map:
		vp := reflect.New(t)
		deserializeMapReflect(d, t, vp.Elem(), vp.UnsafePointer())
		v = vp.Elem()
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
		ep := deserializePointedAt(d, t.Elem(), -1)
		v = reflect.New(t).Elem()
		v.Set(reflect.NewAt(t.Elem(), ep))
	case reflect.UnsafePointer:
		p := deserializePointedAt(d, unsafePointerT, -1)
		v = reflect.ValueOf(p)
	default:
		panic(fmt.Sprintf("not implemented: deserializing reflect.Value with type %s", t))
	}
	return
}

func serializePointedAt(s *Serializer, et reflect.Type, length int, p unsafe.Pointer) {
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

	// Check the region of this pointer.
	r := s.containers.of(p)

	// If the pointer does not point to a known region encountered via
	// scanning, create a new temporary region. This can occur when a
	// custom serializer emits memory regions during serialization (and
	// after the root object has been scanned). Note that we do not scan
	// the memory region! This means it's not possible to alias this
	// memory region (or other regions it points to that aren't known
	// to the serializer). Scanning here might cause known regions to
	// expand, invalidating those that have already been encoded.
	if !r.valid() {
		if et == nil {
			panic("cannot serialize unsafe.Pointer pointing to region of unknown size")
		}
		r.addr = p
		r.typ = et
		r.len = length
	}

	if r.len < 0 && r.typ.Kind() == reflect.Map {
		serializeMap(s, r.typ, r.addr)
		return
	}

	id, new := s.assignPointerID(r.addr)
	serializeVarint(s, int(id))

	offset := int(r.offset(p))
	serializeVarint(s, offset)

	if !new {
		return
	}

	region := &coroutinev1.Region{
		Type: s.types.ToType(r.typ) << 1,
	}
	if r.len >= 0 {
		region.Type |= 1
		region.ArrayLength = uint32(r.len)
	}
	s.regions = append(s.regions, region)

	// Fast path for byte arrays.
	if r.len >= 0 && r.typ.Kind() == reflect.Uint8 {
		if r.len > 0 {
			region.Data = unsafe.Slice((*byte)(r.addr), r.len)
		}
		return
	}

	regionSer := s.fork()
	if r.len >= 0 { // array
		es := int(r.typ.Size())
		for i := 0; i < r.len; i++ {
			serializeAny(regionSer, r.typ, unsafe.Add(r.addr, i*es))
		}
	} else {
		serializeAny(regionSer, r.typ, r.addr)
	}
	region.Data = regionSer.b
}

func deserializePointedAt(d *Deserializer, t reflect.Type, length int) unsafe.Pointer {
	// This function is a bit different than the other deserialize* ones
	// because it deserializes into an unknown location. As a result,
	// instead of taking an unsafe.Pointer as an input, it returns an
	// unsafe.Pointer to a deserialized object.

	if length < 0 && t.Kind() == reflect.Map {
		m := reflect.New(t)
		p := m.UnsafePointer()
		deserializeMapReflect(d, t, m.Elem(), m.UnsafePointer())
		return p
	}

	id := deserializeVarint(d)
	if id == 0 {
		// Nil pointer.
		return unsafe.Pointer(nil)
	}

	offset := deserializeVarint(d)
	if id == -1 {
		// Pointer into static uint64 table.
		return staticPointer(offset)
	}

	p := d.ptrs[sID(id)]
	if p == nil {
		// Deserialize the region.
		if int(id) > len(d.regions) {
			panic(fmt.Sprintf("region %d not found", id))
		}
		region := d.regions[id-1]

		regionType := d.types.ToReflect(typeid(region.Type >> 1))

		if region.Type&1 == 1 {
			elemSize := int(regionType.Size())
			length := int(region.ArrayLength)
			data := make([]byte, elemSize*length)
			p = unsafe.Pointer(unsafe.SliceData(data))
			d.store(sID(id), p)

			// Fast path for byte arrays.
			if regionType.Kind() == reflect.Uint8 {
				if length > 0 {
					copy(unsafe.Slice((*byte)(p), length), region.Data)
				}
			} else {
				regionDeser := d.fork(region.Data)
				for i := 0; i < length; i++ {
					deserializeAny(regionDeser, regionType, unsafe.Add(p, elemSize*i))
				}
			}
		} else {
			container := reflect.New(regionType)
			p = container.UnsafePointer()
			d.store(sID(id), p)
			regionDeser := d.fork(region.Data)
			deserializeAny(regionDeser, regionType, p)
		}

	}

	// Create the pointer with an offset into the container.
	return unsafe.Add(p, offset)
}

func serializeMap(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	v := reflect.NewAt(t, p).Elem()
	serializeMapReflect(s, v)
}

func serializeMapReflect(s *Serializer, v reflect.Value) {
	if v.IsNil() {
		serializeVarint(s, 0)
		return
	}

	mapptr := v.UnsafePointer()

	id, new := s.assignPointerID(mapptr)
	serializeVarint(s, int(id))
	serializeVarint(s, 0) // offset, for compat with other region references

	if !new {
		return
	}

	size := v.Len()

	t := v.Type()
	region := &coroutinev1.Region{
		Type: s.types.ToType(t) << 1,
	}
	s.regions = append(s.regions, region)

	regionSer := s.fork()
	serializeVarint(regionSer, size)

	// TODO: allocs
	iter := v.MapRange()
	mk := reflect.New(t.Key()).Elem()
	mv := reflect.New(t.Elem()).Elem()
	for iter.Next() {
		mk.Set(iter.Key())
		mv.Set(iter.Value())
		serializeAny(regionSer, t.Key(), mk.Addr().UnsafePointer())
		serializeAny(regionSer, t.Elem(), mv.Addr().UnsafePointer())
	}

	region.Data = regionSer.b
}

func deserializeMap(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	r := reflect.NewAt(t, p).Elem()
	deserializeMapReflect(d, t, r, p)
}

func deserializeMapReflect(d *Deserializer, t reflect.Type, r reflect.Value, p unsafe.Pointer) {
	id := deserializeVarint(d)
	if id == 0 {
		r.SetZero()
		return
	}

	_ = deserializeVarint(d) // offset

	ptr := d.ptrs[sID(id)]
	if ptr != nil {
		existing := reflect.NewAt(t, ptr).Elem()
		r.Set(existing)
		return
	}

	if id > len(d.regions) {
		panic(fmt.Sprintf("region %d not found", id))
	}
	region := d.regions[id-1]

	regionDeser := d.fork(region.Data)

	n := deserializeVarint(regionDeser)
	if n < 0 { // nil map
		panic("invalid map size")
	}

	nv := reflect.MakeMapWithSize(t, n)
	r.Set(nv)
	d.store(sID(id), p)
	for i := 0; i < n; i++ {
		k := reflect.New(t.Key())
		deserializeAny(regionDeser, t.Key(), k.UnsafePointer())
		v := reflect.New(t.Elem())
		deserializeAny(regionDeser, t.Elem(), v.UnsafePointer())
		r.SetMapIndex(k.Elem(), v.Elem())
	}
}

func serializeSlice(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	r := reflect.NewAt(t, p).Elem()

	serializeVarint(s, r.Len())
	serializeVarint(s, r.Cap())
	serializePointedAt(s, t.Elem(), r.Cap(), r.UnsafePointer())
}

func deserializeSlice(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	l := deserializeVarint(d)
	c := deserializeVarint(d)

	ar := deserializePointedAt(d, t.Elem(), c)
	if ar == nil {
		return
	}

	s := (*slice)(p)
	s.data = ar
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
	serializePointedAt(s, t.Elem(), -1, x)
}

func deserializePointer(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	ep := deserializePointedAt(d, t.Elem(), -1)
	r := reflect.NewAt(t, p)
	r.Elem().Set(reflect.NewAt(t.Elem(), ep))
}

func serializeUnsafePointer(s *Serializer, p unsafe.Pointer) {
	if p == nil {
		serializePointedAt(s, nil, -1, nil)
	} else {
		serializePointedAt(s, nil, -1, *(*unsafe.Pointer)(p))
	}
}

func deserializeUnsafePointer(d *Deserializer, p unsafe.Pointer) {
	r := reflect.NewAt(unsafePointerT, p)

	ep := deserializePointedAt(d, unsafePointerT, -1)
	if ep != nil {
		r.Elem().Set(reflect.ValueOf(ep))
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
	fn := *(**function)(p)
	if fn == nil {
		// Function IDs start at 1; use 0 to represent nil ptr.
		serializeVarint(s, 0)
		return
	}

	id, closure := s.funcs.RegisterAddr(fn.addr)
	serializeVarint(s, int(id))

	if closure != nil {
		p = unsafe.Pointer(fn)
		// Skip the first field, which is the function ptr.
		serializeStructFields(s, p, closure.NumField()-1, func(i int) reflect.StructField {
			return closure.Field(i + 1)
		})
	}
}

func deserializeFunc(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	id := deserializeVarint(d)
	if id == 0 {
		*(**function)(p) = nil
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
		// Avoid an allocation by storing a pointer to the immutable Func.
		// This works because the addr is the first element in both the Func
		// and function structs.
		*(**function)(p) = (*function)(unsafe.Pointer(fn))
	}
}

func serializeInterface(s *Serializer, v reflect.Value) {
	if v.IsNil() {
		serializeBool(s, false)
		return
	}
	serializeBool(s, true)

	et := reflect.TypeOf(v.Interface())
	serializeType(s, et)

	vi := v.Interface()

	eptr := ifacePtr(unsafe.Pointer(&vi), et)

	if et.Kind() == reflect.Array {
		serializePointedAt(s, et.Elem(), et.Len(), eptr)
	} else {
		serializePointedAt(s, et, -1, eptr)
	}
}

func deserializeInterface(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	var ok bool
	deserializeBool(d, &ok)
	if !ok {
		return
	}

	// Deserialize the type
	et, length := deserializeType(d)
	if et == nil {
		return
	}

	// Deserialize the pointer
	ep := deserializePointedAt(d, et, length)

	// Store the result in the interface
	r := reflect.NewAt(t, p)
	if ep != nil {
		// FIXME: is there a way to avoid ArrayOf+NewAt here? We can
		//  access the iface via p. We can set the ptr, but not the typ.
		if length >= 0 {
			et = reflect.ArrayOf(length, et)
		}
		x := reflect.NewAt(et, ep)
		r.Elem().Set(x.Elem())
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

	p := unsafe.Pointer(unsafe.StringData(*x))

	serializePointedAt(s, byteT, l, p)
}

func deserializeString(d *Deserializer, x *string) {
	l := deserializeVarint(d)

	if l == 0 {
		return
	}

	ar := deserializePointedAt(d, byteT, l)

	*x = unsafe.String((*byte)(ar), l)
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
	var real float32
	var imag float32
	deserializeFloat32(d, &real)
	deserializeFloat32(d, &imag)
	*x = complex(real, imag)
}

func serializeComplex128(s *Serializer, x complex128) {
	serializeFloat64(s, real(x))
	serializeFloat64(s, imag(x))
}

func deserializeComplex128(d *Deserializer, x *complex128) {
	var real float64
	var imag float64
	deserializeFloat64(d, &real)
	deserializeFloat64(d, &imag)
	*x = complex(real, imag)
}
