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
	reflectext.NewVisitor(s, reflectext.VisitAll).Visit(v)
}

func (d *Deserializer) Deserialize(v reflect.Value) {
	reflectext.NewVisitor(d, reflectext.VisitAll).Visit(v)
}

func (s *Serializer) Visit(ctx reflectext.VisitorContext, v reflect.Value) bool {
	t := v.Type()

	// Special case for values with a custom serializer registered.
	if serde, ok := s.serdes.serdeByType(t); ok {
		offset := len(s.buffer)
		// Store an 8 byte (64-bit) size placeholder. It's used when scanning
		// the region (see inspect.go) to skip over the opaque parts that the
		// custom serialization routine has created. This allows systems
		// that don't have access to the routine to make sense of the region's
		// serialized representation.
		s.buffer = append(s.buffer, 0, 0, 0, 0, 0, 0, 0, 0)
		serde.ser(s, v)
		// Fill in the size of the opaque region now that it's known.
		binary.LittleEndian.PutUint64(s.buffer[offset:], uint64(len(s.buffer)-offset))
		return false
	}

	// Special cases for reflect.Type and reflect.Value.
	switch t {
	case reflectext.ReflectTypeType:
		rt := v.Interface().(reflect.Type)
		s.appendReflectType(rt)
		return false

	case reflectext.ReflectValueType:
		rv := v.Interface().(reflect.Value)
		s.appendReflectType(rv.Type())
		ctx.Visit(rv)
		return false
	}

	return true
}

func (d *Deserializer) Visit(ctx reflectext.VisitorContext, v reflect.Value) bool {
	t := v.Type()

	// Special case for values with a custom serializer registered.
	if serde, ok := d.serdes.serdeByType(t); ok {
		// Skip the size prefix, since we will defer to the custom
		// deserialization routine to make sense of the opaque bytes.
		d.buffer = d.buffer[8:]
		serde.des(d, v)
		return false
	}

	// Special cases for reflect.Type and reflect.Value.
	switch t {
	case reflectext.ReflectTypeType:
		rt, _ := d.reflectType()
		v.Set(reflect.ValueOf(rt))
		return false

	case reflectext.ReflectValueType:
		rt, length := d.reflectType()
		if length >= 0 {
			// ArrayOf is generally something that should be avoided, since
			// it creates a new reflect.Type that isn't garbage collected.
			// Unfortunately we can't avoid the ArrayOf call here. We need to build
			// a reflect.Type in order to construct a reflect.Value to deserialize
			// into. Note that the only time this path is taken is if the user has
			// explicitly serialized a reflect.Value, or some other data type that
			// contains or points to a reflect.Value.
			rt = reflect.ArrayOf(length, rt)
		}
		rv := reflect.New(rt).Elem()
		ctx.Visit(rv)
		v.Set(reflect.ValueOf(rv))
		return false
	}

	return true
}

func (s *Serializer) VisitBool(ctx reflectext.VisitorContext, v reflect.Value) {
	s.appendBool(v.Bool())
}

func (d *Deserializer) VisitBool(ctx reflectext.VisitorContext, v reflect.Value) {
	v.SetBool(d.bool())
}

func (s *Serializer) VisitInt(ctx reflectext.VisitorContext, v reflect.Value) {
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
		s.appendInt64(i)
	}
}

func (d *Deserializer) VisitInt(ctx reflectext.VisitorContext, v reflect.Value) {
	var i int64
	switch v.Kind() {
	case reflect.Int:
		i = int64(d.int())
	case reflect.Int8:
		i = int64(d.int8())
	case reflect.Int16:
		i = int64(d.int16())
	case reflect.Int32:
		i = int64(d.int32())
	case reflect.Int64:
		i = d.int64()
	}
	v.SetInt(i)
}

func (s *Serializer) VisitUint(ctx reflectext.VisitorContext, v reflect.Value) {
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

func (d *Deserializer) VisitUint(ctx reflectext.VisitorContext, v reflect.Value) {
	var u uint64
	switch v.Kind() {
	case reflect.Uint:
		u = uint64(d.uint())
	case reflect.Uint8:
		u = uint64(d.uint8())
	case reflect.Uint16:
		u = uint64(d.uint16())
	case reflect.Uint32:
		u = uint64(d.uint32())
	case reflect.Uint64:
		u = d.uint64()
	case reflect.Uintptr:
		u = uint64(d.uintptr())
	}
	v.SetUint(u)
}

func (s *Serializer) VisitFloat(ctx reflectext.VisitorContext, v reflect.Value) {
	f := v.Float()
	switch v.Kind() {
	case reflect.Float32:
		s.appendFloat32(float32(f))
	case reflect.Float64:
		s.appendFloat64(float64(f))
	}
}

func (d *Deserializer) VisitFloat(ctx reflectext.VisitorContext, v reflect.Value) {
	var f float64
	switch v.Kind() {
	case reflect.Float32:
		f = float64(d.float32())
	case reflect.Float64:
		f = d.float64()
	}
	v.SetFloat(f)
}

func (s *Serializer) VisitString(ctx reflectext.VisitorContext, v reflect.Value) {
	str := v.String()
	siz := len(str)
	s.appendVarint(siz)
	if siz > 0 {
		p := unsafe.Pointer(unsafe.StringData(str))
		s.serializeRegion(ctx, reflectext.ByteType, siz, p)
	}
}

func (d *Deserializer) VisitString(ctx reflectext.VisitorContext, v reflect.Value) {
	len := d.varint()
	if len > 0 {
		p := d.deserializeRegion(ctx, reflectext.ByteType, len)
		s := unsafe.String((*byte)(p), len)
		v.SetString(s)
	}
}

func (s *Serializer) VisitSlice(ctx reflectext.VisitorContext, v reflect.Value) bool {
	s.appendVarint(v.Len())
	s.appendVarint(v.Cap())
	s.serializeRegion(ctx, v.Type().Elem(), v.Cap(), v.UnsafePointer())
	return false
}

func (d *Deserializer) VisitSlice(ctx reflectext.VisitorContext, v reflect.Value) bool {
	len := d.varint()
	cap := d.varint()
	t := v.Type()
	data := d.deserializeRegion(ctx, t.Elem(), cap)
	if data != nil {
		v.Set(reflectext.MakeSlice(t, data, len, cap))
	}
	return false
}

func (s *Serializer) VisitInterface(ctx reflectext.VisitorContext, v reflect.Value) bool {
	if v.IsNil() {
		s.appendBool(false)
		return false
	}
	s.appendBool(true)

	et := v.Elem().Type()
	s.appendReflectType(et)

	p := reflectext.InterfaceValueOf(v).DataPointer()
	if et.Kind() == reflect.Array {
		s.serializeRegion(ctx, et.Elem(), et.Len(), p)
	} else {
		s.serializeRegion(ctx, et, -1, p)
	}
	return false
}

func (d *Deserializer) VisitInterface(ctx reflectext.VisitorContext, v reflect.Value) bool {
	if notNil := d.bool(); !notNil {
		return false // nil
	}
	et, length := d.reflectType()
	if et == nil {
		return false
	}
	if ep := d.deserializeRegion(ctx, et, length); ep != nil {
		// FIXME: is there a way to avoid ArrayOf+NewAt here? We can
		//  access the iface via p. We can set the ptr, but not the typ.
		if length >= 0 {
			et = reflect.ArrayOf(length, et)
		}
		v.Set(reflect.NewAt(et, ep).Elem())
	} else {
		v.Set(reflect.Zero(et))
	}
	return false
}

func (s *Serializer) VisitMap(ctx reflectext.VisitorContext, v reflect.Value) bool {
	if v.IsNil() {
		s.appendVarint(0) // id=0 means nil
		return false
	}

	p := v.UnsafePointer()

	id, isNew := s.assignPointerID(p)
	s.appendVarint(int(id))
	s.appendVarint(0) // offset, for compat with other region references

	if !isNew {
		return false
	}

	region := &coroutinev1.Region{
		Type: s.types.ToType(v.Type()) << 1,
	}
	s.regions = append(s.regions, region)

	regionSerializer := s.fork()
	regionSerializer.appendVarint(v.Len())

	mapVisitor := ctx.Fork(regionSerializer)

	iter := v.MapRange()
	for iter.Next() {
		mapVisitor.Visit(iter.Key())
		mapVisitor.Visit(iter.Value())
	}

	region.Data = regionSerializer.buffer

	return false
}

func (d *Deserializer) VisitMap(ctx reflectext.VisitorContext, v reflect.Value) bool {
	// Nil map.
	id := d.varint()
	if id == 0 {
		v.SetZero()
		return false
	}

	_ = d.varint() // offset, for compatibility with other region references

	// Map we've already deserialized, grab the pointer
	// and create a reference.
	t := v.Type()
	p := d.ptrs[sID(id)]
	if p != nil {
		existing := reflect.NewAt(t, p).Elem()
		v.Set(existing)
		return false
	}

	if id > len(d.regions) {
		panic(fmt.Sprintf("region %d not found", id))
	}
	region := d.regions[id-1]
	regionDeserializer := d.fork(region.Data)

	n := regionDeserializer.varint()
	if n < 0 { // nil map
		panic("invalid map size")
	}

	keyType := t.Key()
	valType := t.Elem()

	mapVisitor := ctx.Fork(regionDeserializer)

	nv := reflect.MakeMapWithSize(t, n)
	v.Set(nv)
	d.store(sID(id), v.Addr().UnsafePointer())
	for i := 0; i < n; i++ {
		kv := reflect.New(keyType).Elem()
		mapVisitor.Visit(kv)
		vv := reflect.New(valType).Elem()
		mapVisitor.Visit(vv)
		v.SetMapIndex(kv, vv)
	}
	return false
}

func (s *Serializer) VisitFunc(ctx reflectext.VisitorContext, v reflect.Value) bool {
	if v.IsNil() {
		// Function IDs start at 1; use 0 to represent nil ptr.
		s.appendVarint(0)
		return false
	}
	id, _ := s.funcs.RegisterAddr(v.UnsafePointer())
	s.appendVarint(int(id))
	return true
}

func (d *Deserializer) VisitFunc(ctx reflectext.VisitorContext, v reflect.Value) bool {
	id := d.varint()
	if id == 0 {
		v.SetZero()
		return false
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
		closure := reflect.New(fn.Closure).Elem()
		fv.SetClosure(fn.Addr, closure)
	} else {

		fv.SetAddr(fn.Addr)
	}
	return true
}

func (s *Serializer) VisitPointer(ctx reflectext.VisitorContext, v reflect.Value) bool {
	s.serializeRegion(ctx, v.Type().Elem(), -1, v.UnsafePointer())
	return false
}

func (d *Deserializer) VisitPointer(ctx reflectext.VisitorContext, v reflect.Value) bool {
	t := v.Type()
	p := d.deserializeRegion(ctx, t.Elem(), -1)
	v.Set(reflect.NewAt(t.Elem(), p))
	return false
}

func (s *Serializer) VisitUnsafePointer(ctx reflectext.VisitorContext, v reflect.Value) {
	s.serializeRegion(ctx, nil, -1, v.UnsafePointer())
}

func (d *Deserializer) VisitUnsafePointer(ctx reflectext.VisitorContext, v reflect.Value) {
	p := d.deserializeRegion(ctx, reflectext.UnsafePointerType, -1)
	v.SetPointer(p)
}

func (s *Serializer) VisitChan(ctx reflectext.VisitorContext, v reflect.Value) bool {
	panic("not implemented: reflect.Chan")
}

func (s *Deserializer) VisitChan(ctx reflectext.VisitorContext, v reflect.Value) bool {
	panic("not implemented: reflect.Chan")
}

func (s *Serializer) appendReflectType(t reflect.Type) {
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

func (d *Deserializer) reflectType() (reflect.Type, int) {
	id := d.varint()
	length := d.varint()
	t := d.types.ToReflect(typeid(id))
	return t, length
}

func (s *Serializer) serializeRegion(ctx reflectext.VisitorContext, et reflect.Type, length int, p unsafe.Pointer) {
	// If this is a nil pointer, write it as such.
	if p == nil {
		s.appendVarint(0)
		return
	}

	if offset, ok := reflectext.InternedValue(p); ok {
		s.appendVarint(-1)
		s.appendVarint(int(offset))
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
		ctx.Visit(v)
		return
	}

	id, isNew := s.assignPointerID(r.addr)
	s.appendVarint(int(id))

	offset := int(r.offset(p))
	s.appendVarint(offset)

	if !isNew {
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
	// TODO: fast paths for other basic types
	if r.len >= 0 && r.typ.Kind() == reflect.Uint8 {
		if r.len > 0 {
			region.Data = unsafe.Slice((*byte)(r.addr), r.len)
		}
		return
	}

	regionSerializer := s.fork()
	regionVisitor := ctx.Fork(regionSerializer)
	if r.len >= 0 { // array
		data := reflectext.MakeSlice(reflect.SliceOf(r.typ), r.addr, r.len, r.len)
		for i := 0; i < data.Len(); i++ {
			regionVisitor.Visit(data.Index(i))
		}
	} else {
		v := reflect.NewAt(r.typ, r.addr).Elem()
		regionVisitor.Visit(v)
	}
	region.Data = regionSerializer.buffer
}

func (d *Deserializer) deserializeRegion(ctx reflectext.VisitorContext, t reflect.Type, length int) unsafe.Pointer {
	if length < 0 && t.Kind() == reflect.Map {
		m := reflect.New(t)
		p := m.UnsafePointer()
		ctx.Visit(m.Elem())
		return p
	}

	id := d.varint()
	if id == 0 {
		return unsafe.Pointer(nil)
	}

	offset := d.varint()
	if id == -1 {
		return reflectext.InternedValueOffset(offset).UnsafePointer()
	}

	p := d.ptrs[sID(id)]
	if p == nil {
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
			//
			// TODO: fast paths for other basic types
			if regionType.Kind() == reflect.Uint8 {
				if length > 0 {
					copy(unsafe.Slice((*byte)(p), length), region.Data)
				}
			} else {
				regionDeserializer := d.fork(region.Data)
				regionVisitor := ctx.Fork(regionDeserializer)
				data := reflectext.MakeSlice(reflect.SliceOf(regionType), p, length, length)
				for i := 0; i < length; i++ {
					regionVisitor.Visit(data.Index(i))
				}
			}
		} else {
			container := reflect.New(regionType)
			p = container.UnsafePointer()
			d.store(sID(id), p)
			regionDeserializer := d.fork(region.Data)
			regionVisitor := ctx.Fork(regionDeserializer)
			v := reflect.NewAt(regionType, p).Elem()
			regionVisitor.Visit(v)
		}
	}

	return unsafe.Add(p, offset)
}
