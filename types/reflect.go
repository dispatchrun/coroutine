package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"unsafe"

	coroutinev1 "github.com/dispatchrun/coroutine/gen/proto/go/coroutine/v1"
	"github.com/dispatchrun/coroutine/internal/reflectext"
)

func (s *Serializer) Visit(v reflect.Value) bool {
	t := v.Type()

	// Special case for values with a custom serializer registered.
	if serde, ok := s.serdes.serdeByType(t); ok {
		offset := len(s.b)
		s.b = append(s.b, 0, 0, 0, 0, 0, 0, 0, 0) // store a 64-bit size placeholder
		serde.ser(s, v)
		binary.LittleEndian.PutUint64(s.b[offset:], uint64(len(s.b)-offset))
		return false
	}

	// Special cases for reflect.Type and reflect.Value.
	switch t {
	case reflectext.ReflectTypeType:
		rt := v.Interface().(reflect.Type)
		serializeType(s, rt)
		return false

	case reflectext.ReflectValueType:
		rv := v.Interface().(reflect.Value)
		serializeType(s, rv.Type())
		reflectext.Visit(s, rv, reflectext.VisitUnexportedFields|reflectext.VisitClosures) // FIXME: propagate flags
		return false
	}

	return true
}

func (s *Serializer) VisitBool(v bool) { serializeBool(s, v) }

func (s *Serializer) VisitInt(v int)     { serializeInt(s, v) }
func (s *Serializer) VisitInt8(v int8)   { serializeInt8(s, v) }
func (s *Serializer) VisitInt16(v int16) { serializeInt16(s, v) }
func (s *Serializer) VisitInt32(v int32) { serializeInt32(s, v) }
func (s *Serializer) VisitInt64(v int64) { serializeInt64(s, v) }

func (s *Serializer) VisitUint(v uint)       { serializeUint(s, v) }
func (s *Serializer) VisitUint8(v uint8)     { serializeUint8(s, v) }
func (s *Serializer) VisitUint16(v uint16)   { serializeUint16(s, v) }
func (s *Serializer) VisitUint32(v uint32)   { serializeUint32(s, v) }
func (s *Serializer) VisitUint64(v uint64)   { serializeUint64(s, v) }
func (s *Serializer) VisitUintptr(v uintptr) { serializeUint64(s, uint64(v)) }

func (s *Serializer) VisitFloat32(v float32) { serializeFloat32(s, v) }
func (s *Serializer) VisitFloat64(v float64) { serializeFloat64(s, v) }

func (s *Serializer) VisitString(v string) {
	serializeVarint(s, len(v))
	if len(v) > 0 {
		p := unsafe.Pointer(unsafe.StringData(v))
		serializePointedAt(s, reflectext.ByteType, len(v), p)
	}
}

func (s *Serializer) VisitSlice(v reflect.Value) bool {
	serializeVarint(s, v.Len())
	serializeVarint(s, v.Cap())
	serializePointedAt(s, v.Type().Elem(), v.Cap(), v.UnsafePointer())
	return false
}

func (s *Serializer) VisitInterface(v reflect.Value) bool {
	if v.IsNil() {
		serializeBool(s, false)
		return false
	}
	serializeBool(s, true)

	et := v.Elem().Type()
	serializeType(s, et)

	p := reflectext.InterfacePointer(v)
	if et.Kind() == reflect.Array {
		serializePointedAt(s, et.Elem(), et.Len(), p)
	} else {
		serializePointedAt(s, et, -1, p)
	}
	return false
}

func (s *Serializer) VisitMap(v reflect.Value) bool {
	serializeMap(s, v)
	return false
}

func (s *Serializer) VisitFunc(v reflect.Value) bool {
	if v.IsNil() {
		// Function IDs start at 1; use 0 to represent nil ptr.
		serializeVarint(s, 0)
		return false
	}
	id, _ := s.funcs.RegisterAddr(v.UnsafePointer())
	serializeVarint(s, int(id))
	return true
}

func (s *Serializer) VisitPointer(v reflect.Value) bool {
	serializePointedAt(s, v.Type().Elem(), -1, v.UnsafePointer())
	return false
}

func (s *Serializer) VisitUnsafePointer(p unsafe.Pointer) {
	serializePointedAt(s, nil, -1, p)
}

func (s *Serializer) VisitChan(v reflect.Value) bool {
	panic("not implemented: reflect.Chan")
}

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

func deserializeValue(d *Deserializer, t reflect.Type, vp reflect.Value) {
	v := vp.Elem()

	if serde, ok := d.serdes.serdeByType(t); ok {
		d.b = d.b[8:] // skip size prefix
		serde.des(d, vp)
		return
	}

	switch t {
	case reflectext.ReflectTypeType:
		rt, _ := deserializeType(d)
		v.Set(reflect.ValueOf(rt))
		return
	case reflectext.ReflectValueType:
		rt, length := deserializeType(d)
		if length >= 0 {
			// We can't avoid the ArrayOf call here. We need to build a
			// reflect.Type in order to return a reflect.Value. The only
			// time this path is taken is if the user has explicitly serialized
			// a reflect.Value, or some other data type that contains or points
			// to a reflect.Value.
			rt = reflect.ArrayOf(length, rt)
		}
		rp := reflect.New(rt)
		deserializeValue(d, rt, rp)
		v.Set(reflect.ValueOf(rp.Elem()))
		return
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't deserialize reflect.Invalid"))
	case reflect.Bool:
		var value bool
		deserializeBool(d, &value)
		v.SetBool(value)
	case reflect.Int:
		var value int
		deserializeInt(d, &value)
		v.SetInt(int64(value))
	case reflect.Int8:
		var value int8
		deserializeInt8(d, &value)
		v.SetInt(int64(value))
	case reflect.Int16:
		var value int16
		deserializeInt16(d, &value)
		v.SetInt(int64(value))
	case reflect.Int32:
		var value int32
		deserializeInt32(d, &value)
		v.SetInt(int64(value))
	case reflect.Int64:
		var value int64
		deserializeInt64(d, &value)
		v.SetInt(value)
	case reflect.Uint:
		var value uint
		deserializeUint(d, &value)
		v.SetUint(uint64(value))
	case reflect.Uint8:
		var value uint8
		deserializeUint8(d, &value)
		v.SetUint(uint64(value))
	case reflect.Uint16:
		var value uint16
		deserializeUint16(d, &value)
		v.SetUint(uint64(value))
	case reflect.Uint32:
		var value uint32
		deserializeUint32(d, &value)
		v.SetUint(uint64(value))
	case reflect.Uint64:
		var value uint64
		deserializeUint64(d, &value)
		v.SetUint(value)
	case reflect.Uintptr:
		var value uint64
		deserializeUint64(d, &value)
		v.SetUint(value)
	case reflect.Float32:
		var value float32
		deserializeFloat32(d, &value)
		v.SetFloat(float64(value))
	case reflect.Float64:
		var value float64
		deserializeFloat64(d, &value)
		v.SetFloat(value)
	case reflect.Complex64:
		var value complex64
		deserializeComplex64(d, &value)
		v.SetComplex(complex128(value))
	case reflect.Complex128:
		var value complex128
		deserializeComplex128(d, &value)
		v.SetComplex(value)
	case reflect.String:
		var value string
		deserializeString(d, &value)
		v.SetString(value)
	case reflect.Array:
		deserializeArray(d, t, v.Addr().UnsafePointer())
	case reflect.Slice:
		len := deserializeVarint(d)
		cap := deserializeVarint(d)
		data := deserializePointedAt(d, t.Elem(), cap)
		if data == nil {
			return
		}
		reflectext.SetSlice(v, data, len, cap)
	case reflect.Map:
		deserializeMap(d, t, v, vp.UnsafePointer())
	case reflect.Struct:
		deserializeStructFields(d, vp.UnsafePointer(), t.NumField(), t.Field)
	case reflect.Func:
		deserializeFunc(d, v)
	case reflect.Interface:
		var ok bool
		deserializeBool(d, &ok)
		if ok {
			et, length := deserializeType(d)
			if et != nil {
				if ep := deserializePointedAt(d, et, length); ep != nil {
					// FIXME: is there a way to avoid ArrayOf+NewAt here? We can
					//  access the iface via p. We can set the ptr, but not the typ.
					if length >= 0 {
						et = reflect.ArrayOf(length, et)
					}
					v.Set(reflect.NewAt(et, ep).Elem())
				} else {
					v.Set(reflect.Zero(et))
				}

			}
		}
	case reflect.Pointer:
		ep := deserializePointedAt(d, t.Elem(), -1)
		v.Set(reflect.NewAt(t.Elem(), ep))
	case reflect.UnsafePointer:
		p := deserializePointedAt(d, reflectext.UnsafePointerType, -1)
		v.SetPointer(p)
	default:
		panic(fmt.Sprintf("not implemented: deserializing reflect.Value with type %s", t))
	}
}

func serializePointedAt(s *Serializer, et reflect.Type, length int, p unsafe.Pointer) {
	// If this is a nil pointer, write it as such.
	if p == nil {
		serializeVarint(s, 0)
		return
	}

	if offset, ok := reflectext.InternedInt(p); ok {
		serializeVarint(s, -1)
		serializeVarint(s, offset)
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
		v := reflect.NewAt(r.typ, r.addr).Elem()
		serializeMap(s, v)
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
			v := reflect.NewAt(r.typ, unsafe.Add(r.addr, i*es)).Elem()
			reflectext.Visit(regionSer, v, reflectext.VisitUnexportedFields|reflectext.VisitClosures) // FIXME: propagate flags
		}
	} else {
		v := reflect.NewAt(r.typ, r.addr).Elem()
		reflectext.Visit(regionSer, v, reflectext.VisitUnexportedFields|reflectext.VisitClosures) // FIXME: propagate flags
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
		deserializeMap(d, t, m.Elem(), m.UnsafePointer())
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
		return reflectext.InternedIntPointer(offset)
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
					vp := reflect.NewAt(regionType, unsafe.Add(p, elemSize*i))
					deserializeValue(regionDeser, regionType, vp)
				}
			}
		} else {
			container := reflect.New(regionType)
			p = container.UnsafePointer()
			d.store(sID(id), p)
			regionDeser := d.fork(region.Data)
			vp := reflect.NewAt(regionType, p)
			deserializeValue(regionDeser, regionType, vp)
		}

	}

	// Create the pointer with an offset into the container.
	return unsafe.Add(p, offset)
}

func serializeMap(s *Serializer, v reflect.Value) {
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

	iter := v.MapRange()
	for iter.Next() {
		reflectext.Visit(regionSer, iter.Key(), reflectext.VisitUnexportedFields|reflectext.VisitClosures)   // FIXME: propagate flags
		reflectext.Visit(regionSer, iter.Value(), reflectext.VisitUnexportedFields|reflectext.VisitClosures) // FIXME: propagate flags
	}

	region.Data = regionSer.b
}

func deserializeMap(d *Deserializer, t reflect.Type, r reflect.Value, p unsafe.Pointer) {
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
		kp := reflect.New(t.Key())
		deserializeValue(regionDeser, t.Key(), kp)
		vp := reflect.New(t.Elem())
		deserializeValue(regionDeser, t.Elem(), vp)
		r.SetMapIndex(kp.Elem(), vp.Elem())
	}
}

func deserializeArray(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	et := t.Elem()
	size := int(et.Size())
	for i := 0; i < t.Len(); i++ {
		vp := reflect.NewAt(et, unsafe.Add(p, size*i))
		deserializeValue(d, et, vp)
	}
}

func deserializeStructFields(d *Deserializer, p unsafe.Pointer, n int, field func(int) reflect.StructField) {
	for i := 0; i < n; i++ {
		ft := field(i)
		vp := reflect.NewAt(ft.Type, unsafe.Add(p, ft.Offset))
		deserializeValue(d, ft.Type, vp)
	}
}

func deserializeFunc(d *Deserializer, v reflect.Value) {
	id := deserializeVarint(d)
	if id == 0 {
		v.SetZero()
		return
	}

	t := v.Type()
	fn := d.funcs.ToFunc(funcid(id))
	if fn.Type == nil {
		panic(fn.Name + ": function type is missing")
	}
	if !t.AssignableTo(fn.Type) {
		panic(fn.Name + ": function type mismatch: " + fn.Type.String() + " != " + t.String())
	}

	var header *reflectext.FunctionHeader
	if fn.Closure != nil {
		t := fn.Closure
		cv := reflect.New(t)

		closure := (*reflectext.FunctionHeader)(cv.UnsafePointer())

		deserializeStructFields(d, unsafe.Pointer(closure), t.NumField(), func(i int) reflect.StructField {
			return t.Field(i)
		})

		closure.Addr = unsafe.Pointer(fn.Addr)
		header = closure
	} else {
		header = &reflectext.FunctionHeader{Addr: unsafe.Pointer(fn.Addr)}
	}

	p := v.Addr().UnsafePointer()
	*(*unsafe.Pointer)(p) = unsafe.Pointer(header)
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

func deserializeString(d *Deserializer, x *string) {
	l := deserializeVarint(d)

	if l == 0 {
		return
	}

	ar := deserializePointedAt(d, reflectext.ByteType, l)

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

func deserializeComplex64(d *Deserializer, x *complex64) {
	var real float32
	var imag float32
	deserializeFloat32(d, &real)
	deserializeFloat32(d, &imag)
	*x = complex(real, imag)
}

func deserializeComplex128(d *Deserializer, x *complex128) {
	var real float64
	var imag float64
	deserializeFloat64(d, &real)
	deserializeFloat64(d, &imag)
	*x = complex(real, imag)
}
