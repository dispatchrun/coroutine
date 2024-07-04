package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"

	coroutinev1 "github.com/dispatchrun/coroutine/gen/proto/go/coroutine/v1"
	"github.com/dispatchrun/coroutine/internal/reflectext"
)

func (s *Serializer) Serialize(v reflect.Value) {
	reflectext.Visit(s, v, reflectext.VisitAll)
}

func (s *Serializer) Visit(ctx reflectext.VisitContext, v reflect.Value) bool {
	t := v.Type()

	// Special case for values with a custom serializer registered.
	if serde, ok := s.serdes.serdeByType(t); ok {
		offset := len(s.buffer)
		s.buffer = append(s.buffer, 0, 0, 0, 0, 0, 0, 0, 0) // store a 64-bit size placeholder
		serde.ser(s, v)
		binary.LittleEndian.PutUint64(s.buffer[offset:], uint64(len(s.buffer)-offset))
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
		s.Serialize(rv)
		return false
	}

	return true
}

func (s *Serializer) VisitBool(ctx reflectext.VisitContext, v reflect.Value) {
	s.appendBool(v.Bool())
}

func (s *Serializer) VisitInt(ctx reflectext.VisitContext, v reflect.Value) {
	i := v.Int()
	switch v.Kind() {
	case reflect.Int:
		s.appendInt(int(i))
	case reflect.Int8:
		s.appendInt8(int8(i))
	case reflect.Int16:
		s.appendInt16(int16(i))
	case reflect.Int32:
		s.appendInt32(int32(i))
	case reflect.Int64:
		s.appendInt64(int64(i))
	}
}

func (s *Serializer) VisitUint(ctx reflectext.VisitContext, v reflect.Value) {
	u := v.Uint()
	switch v.Kind() {
	case reflect.Uint:
		s.appendUint(uint(u))
	case reflect.Uint8:
		s.appendUint8(uint8(u))
	case reflect.Uint16:
		s.appendUint16(uint16(u))
	case reflect.Uint32:
		s.appendUint32(uint32(u))
	case reflect.Uint64:
		s.appendUint64(uint64(u))
	case reflect.Uintptr:
		s.appendUintptr(uintptr(u))
	}
}

func (s *Serializer) VisitFloat(ctx reflectext.VisitContext, v reflect.Value) {
	f := v.Float()
	switch v.Kind() {
	case reflect.Float32:
		s.appendFloat32(float32(f))
	case reflect.Float64:
		s.appendFloat64(float64(f))
	}
}

func (s *Serializer) VisitString(ctx reflectext.VisitContext, v reflect.Value) {
	str := v.String()
	siz := len(str)
	s.appendVarint(siz)
	if siz > 0 {
		p := unsafe.Pointer(unsafe.StringData(str))
		serializeRegion(s, reflectext.ByteType, siz, p)
	}
}

func (s *Serializer) VisitSlice(ctx reflectext.VisitContext, v reflect.Value) bool {
	s.appendVarint(v.Len())
	s.appendVarint(v.Cap())
	serializeRegion(s, v.Type().Elem(), v.Cap(), v.UnsafePointer())
	return false
}

func (s *Serializer) VisitInterface(ctx reflectext.VisitContext, v reflect.Value) bool {
	if v.IsNil() {
		s.appendBool(false)
		return false
	}
	s.appendBool(true)

	et := v.Elem().Type()
	serializeType(s, et)

	p := reflectext.InterfaceValueOf(v).DataPointer()
	if et.Kind() == reflect.Array {
		serializeRegion(s, et.Elem(), et.Len(), p)
	} else {
		serializeRegion(s, et, -1, p)
	}
	return false
}

func (s *Serializer) VisitMap(ctx reflectext.VisitContext, v reflect.Value) bool {
	serializeMap(s, v)
	return false
}

func (s *Serializer) VisitFunc(ctx reflectext.VisitContext, v reflect.Value) bool {
	if v.IsNil() {
		// Function IDs start at 1; use 0 to represent nil ptr.
		s.appendVarint(0)
		return false
	}
	id, _ := s.funcs.RegisterAddr(v.UnsafePointer())
	s.appendVarint(int(id))
	return true
}

func (s *Serializer) VisitPointer(ctx reflectext.VisitContext, v reflect.Value) bool {
	serializeRegion(s, v.Type().Elem(), -1, v.UnsafePointer())
	return false
}

func (s *Serializer) VisitUnsafePointer(ctx reflectext.VisitContext, v reflect.Value) {
	serializeRegion(s, nil, -1, v.UnsafePointer())
}

func (s *Serializer) VisitChan(ctx reflectext.VisitContext, v reflect.Value) bool {
	panic("not implemented: reflect.Chan")
}

func serializeType(s *Serializer, t reflect.Type) {
	if t != nil && t.Kind() == reflect.Array {
		id := s.types.ToType(t.Elem())
		s.appendVarint(int(id))
		s.appendVarint(t.Len())
	} else {
		id := s.types.ToType(t)
		s.appendVarint(int(id))
		s.appendVarint(-1)
	}
}

func deserializeType(d *Deserializer) (reflect.Type, int) {
	id := d.varint()
	length := d.varint()
	t := d.types.ToReflect(typeid(id))
	return t, length
}

func deserializeValue(d *Deserializer, t reflect.Type, vp reflect.Value) {
	v := vp.Elem()

	if serde, ok := d.serdes.serdeByType(t); ok {
		d.buffer = d.buffer[8:] // skip size prefix
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
		v.SetBool(d.bool())
	case reflect.Int:
		v.SetInt(int64(d.int()))
	case reflect.Int8:
		v.SetInt(int64(d.int8()))
	case reflect.Int16:
		v.SetInt(int64(d.int16()))
	case reflect.Int32:
		v.SetInt(int64(d.int32()))
	case reflect.Int64:
		v.SetInt(d.int64())
	case reflect.Uint:
		v.SetUint(uint64(d.uint()))
	case reflect.Uint8:
		v.SetUint(uint64(d.uint8()))
	case reflect.Uint16:
		v.SetUint(uint64(d.uint16()))
	case reflect.Uint32:
		v.SetUint(uint64(d.uint32()))
	case reflect.Uint64:
		v.SetUint(d.uint64())
	case reflect.Uintptr:
		v.SetUint(uint64(d.uintptr()))
	case reflect.Float32:
		v.SetFloat(float64(d.float32()))
	case reflect.Float64:
		v.SetFloat(d.float64())
	case reflect.Complex64:
		real := d.float32()
		imag := d.float32()
		v.SetComplex(complex128(complex(real, imag)))
	case reflect.Complex128:
		real := d.float64()
		imag := d.float64()
		v.SetComplex(complex(real, imag))
	case reflect.String:
		var value string
		deserializeString(d, &value)
		v.SetString(value)
	case reflect.Array:
		deserializeArray(d, t, v.Addr().UnsafePointer())
	case reflect.Slice:
		len := d.varint()
		cap := d.varint()
		data := deserializeRegion(d, t.Elem(), cap)
		if data == nil {
			return
		}
		reflectext.SliceValueOf(v).SetSlice(data, len, cap)
	case reflect.Map:
		deserializeMap(d, t, v, vp.UnsafePointer())
	case reflect.Struct:
		deserializeStructFields(d, vp.UnsafePointer(), t.NumField(), t.Field)
	case reflect.Func:
		deserializeFunc(d, v)
	case reflect.Interface:
		if notNil := d.bool(); notNil {
			et, length := deserializeType(d)
			if et != nil {
				if ep := deserializeRegion(d, et, length); ep != nil {
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
		ep := deserializeRegion(d, t.Elem(), -1)
		v.Set(reflect.NewAt(t.Elem(), ep))
	case reflect.UnsafePointer:
		p := deserializeRegion(d, reflectext.UnsafePointerType, -1)
		v.SetPointer(p)
	default:
		panic(fmt.Sprintf("not implemented: deserializing reflect.Value with type %s", t))
	}
}

func serializeRegion(s *Serializer, et reflect.Type, length int, p unsafe.Pointer) {
	// If this is a nil pointer, write it as such.
	if p == nil {
		s.appendVarint(0)
		return
	}

	if offset, ok := reflectext.InternedInt(p); ok {
		s.appendVarint(-1)
		s.appendVarint(offset)
		return
	}

	// Find the region of this pointer.
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
	s.appendVarint(int(id))

	offset := int(r.offset(p))
	s.appendVarint(offset)

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

	regionSerializer := s.fork()
	if r.len >= 0 { // array
		es := int(r.typ.Size())
		for i := 0; i < r.len; i++ {
			v := reflect.NewAt(r.typ, unsafe.Add(r.addr, i*es)).Elem()
			regionSerializer.Serialize(v)
		}
	} else {
		v := reflect.NewAt(r.typ, r.addr).Elem()
		regionSerializer.Serialize(v)
	}
	region.Data = regionSerializer.buffer
}

func deserializeRegion(d *Deserializer, t reflect.Type, length int) unsafe.Pointer {
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

	id := d.varint()
	if id == 0 {
		// Nil pointer.
		return unsafe.Pointer(nil)
	}

	offset := d.varint()
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
		s.appendVarint(0)
		return
	}

	mapptr := v.UnsafePointer()

	id, new := s.assignPointerID(mapptr)
	s.appendVarint(int(id))
	s.appendVarint(0) // offset, for compat with other region references

	if !new {
		return
	}

	size := v.Len()

	t := v.Type()
	region := &coroutinev1.Region{
		Type: s.types.ToType(t) << 1,
	}
	s.regions = append(s.regions, region)

	regionSerializer := s.fork()
	regionSerializer.appendVarint(size)

	iter := v.MapRange()
	for iter.Next() {
		regionSerializer.Serialize(iter.Key())
		regionSerializer.Serialize(iter.Value())
	}

	region.Data = regionSerializer.buffer
}

func deserializeMap(d *Deserializer, t reflect.Type, r reflect.Value, p unsafe.Pointer) {
	id := d.varint()
	if id == 0 {
		r.SetZero()
		return
	}

	_ = d.varint() // offset

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

	n := regionDeser.varint()
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
	id := d.varint()
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

	fv := reflectext.FuncValueOf(v)
	if fn.Closure != nil {
		t := fn.Closure
		closure := reflect.New(t)
		p := closure.UnsafePointer()
		deserializeStructFields(d, p, t.NumField(), t.Field)
		fv.SetClosure(fn.Addr, closure)
	} else {
		fv.SetAddr(fn.Addr)
	}
}

func deserializeInterface(d *Deserializer, t reflect.Type, p unsafe.Pointer) {
	if notNil := d.bool(); !notNil /* i.e. nil */ {
		return
	}

	// Deserialize the type
	et, length := deserializeType(d)
	if et == nil {
		return
	}

	// Deserialize the pointer
	ep := deserializeRegion(d, et, length)

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
	l := d.varint()
	if l == 0 {
		return
	}
	ar := deserializeRegion(d, reflectext.ByteType, l)
	*x = unsafe.String((*byte)(ar), l)
}
