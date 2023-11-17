package types

import (
	"fmt"
	"reflect"
	"unsafe"
)

type typekind int

const (
	typeNone typekind = iota
	typeBool
	typeInt
	typeInt8
	typeInt16
	typeInt32
	typeInt64
	typeUint
	typeUint8
	typeUint16
	typeUint32
	typeUint64
	typeUintptr
	typeFloat32
	typeFloat64
	typeComplex64
	typeComplex128
	typeArray
	typeChan
	typeFunc
	typeInterface
	typeMap
	typePointer
	typeSlice
	typeString
	typeStruct
	typeUnsafePointer
)

// typeinfo represents a type in the serialization format. It is a
// one-size-fits-all struct that contains everything needed to reconstruct a
// reflect.Type. This is because an interface-based approach is more difficult
// to get right, and we will be revamping serde anyway.
type typeinfo struct {
	kind typekind

	// Only present for named types. See documentation of [namedTypeOffset].
	offset namedTypeOffset

	// - typeFunc uses it to store the number of input arguments
	val int

	// typeArray, typeSlice, typePointer, typeChan and typeMap use this field to
	// store the information about the type they contain.
	elem *typeinfo

	key    *typeinfo   // typeMap only
	fields []Field     // typeStruct only
	args   []*typeinfo // typeFunc only
	dir    chanDir     // typeChan only

	// len is the length of an array type
	len int

	// variadic is true if the type represents a function with a variadic argument
	variadic bool

	// custom is true if a custom serializer has been registered for this type
	custom bool
}

type chanDir int

const (
	recvDir chanDir             = 1 << iota // <-chan
	sendDir                                 // chan<-
	bothDir = recvDir | sendDir             // chan
)

func (t *typeinfo) reflectType(tm *typemap) reflect.Type {
	if t.offset != 0 {
		return typeForOffset(t.offset)
	}

	switch t.kind {
	case typeNone:
		return nil
	case typeBool:
		return reflect.TypeOf(false)
	case typeInt:
		return reflect.TypeOf(int(0))
	case typeInt8:
		return reflect.TypeOf(int8(0))
	case typeInt16:
		return reflect.TypeOf(int16(0))
	case typeInt32:
		return reflect.TypeOf(int32(0))
	case typeInt64:
		return reflect.TypeOf(int64(0))
	case typeUint:
		return reflect.TypeOf(uint(0))
	case typeUint8:
		return reflect.TypeOf(uint8(0))
	case typeUint16:
		return reflect.TypeOf(uint16(0))
	case typeUint32:
		return reflect.TypeOf(uint32(0))
	case typeUint64:
		return reflect.TypeOf(uint64(0))
	case typeUintptr:
		return reflect.TypeOf(uintptr(0))
	case typeFloat32:
		return reflect.TypeOf(float32(0))
	case typeFloat64:
		return reflect.TypeOf(float64(0))
	case typeComplex64:
		return reflect.TypeOf(complex64(0))
	case typeComplex128:
		return reflect.TypeOf(complex128(0))
	case typeString:
		return reflect.TypeOf("")
	case typeInterface:
		return typeof[interface{}]()
	case typePointer:
		return reflect.PointerTo(tm.ToReflect(t.elem))
	case typeUnsafePointer:
		return reflect.TypeOf(unsafe.Pointer(nil))
	case typeMap:
		return reflect.MapOf(tm.ToReflect(t.key), tm.ToReflect(t.elem))
	case typeArray:
		return reflect.ArrayOf(t.len, tm.ToReflect(t.elem))
	case typeSlice:
		return reflect.SliceOf(tm.ToReflect(t.elem))
	case typeStruct:
		fields := make([]reflect.StructField, len(t.fields))
		for i, f := range t.fields {
			fields[i].Name = f.name
			fields[i].Tag = reflect.StructTag(f.tag)
			fields[i].Index = f.index
			fields[i].Offset = f.offset
			fields[i].Anonymous = f.anon
			fields[i].Type = tm.ToReflect(f.typ)
		}
		return reflect.StructOf(fields)
	case typeFunc:
		in := t.val
		insouts := make([]reflect.Type, len(t.args))
		for i, t := range t.args {
			insouts[i] = tm.ToReflect(t)
		}
		return reflect.FuncOf(insouts[:in], insouts[in:], t.variadic)
	case typeChan:
		var dir reflect.ChanDir
		switch t.dir {
		case recvDir:
			dir = reflect.RecvDir
		case sendDir:
			dir = reflect.SendDir
		case bothDir:
			dir = reflect.BothDir
		}
		return reflect.ChanOf(dir, tm.ToReflect(t.elem))
	}
	panic(fmt.Errorf("unknown typekind: %d", t.kind))
}

type Field struct {
	name   string
	typ    *typeinfo
	index  []int
	offset uintptr
	anon   bool
	tag    string
}

func (m *typemap) ToReflect(t *typeinfo) reflect.Type {
	if x, ok := m.cache.getV(t); ok {
		return x
	}
	x := t.reflectType(m)
	m.cache.add(x, t)
	return x
}

func (m *typemap) ToType(t reflect.Type) *typeinfo {
	if x, ok := m.cache.getK(t); ok {
		return x
	}

	if t == nil {
		return m.cache.add(t, &typeinfo{kind: typeNone})
	}

	var offset namedTypeOffset
	if named(t) {
		offset = offsetForType(t)
		// Technically types with an offset do not need more information
		// than that. However for debugging purposes also generate the
		// rest of the type information.
	}

	ti := &typeinfo{offset: offset}

	if _, ok := m.serdes.serdeOf(t); ok {
		ti.custom = true
	}

	m.cache.add(t, ti) // add now for recursion
	switch t.Kind() {
	case reflect.Invalid:
		panic("can't handle reflect.Invalid")
	case reflect.Bool:
		ti.kind = typeBool
	case reflect.Int:
		ti.kind = typeInt
	case reflect.Int8:
		ti.kind = typeInt8
	case reflect.Int16:
		ti.kind = typeInt16
	case reflect.Int32:
		ti.kind = typeInt32
	case reflect.Int64:
		ti.kind = typeInt64
	case reflect.Uint:
		ti.kind = typeUint
	case reflect.Uint8:
		ti.kind = typeUint8
	case reflect.Uint16:
		ti.kind = typeUint16
	case reflect.Uint32:
		ti.kind = typeUint32
	case reflect.Uint64:
		ti.kind = typeUint64
	case reflect.Uintptr:
		ti.kind = typeUintptr
	case reflect.Float32:
		ti.kind = typeFloat32
	case reflect.Float64:
		ti.kind = typeFloat64
	case reflect.Complex64:
		ti.kind = typeComplex64
	case reflect.Complex128:
		ti.kind = typeComplex128
	case reflect.String:
		ti.kind = typeString
	case reflect.Interface:
		ti.kind = typeInterface
	case reflect.Array:
		ti.kind = typeArray
		ti.len = t.Len()
		ti.elem = m.ToType(t.Elem())
	case reflect.Map:
		ti.kind = typeMap
		ti.key = m.ToType(t.Key())
		ti.elem = m.ToType(t.Elem())
	case reflect.Pointer:
		ti.kind = typePointer
		ti.elem = m.ToType(t.Elem())
	case reflect.UnsafePointer:
		ti.kind = typeUnsafePointer
	case reflect.Slice:
		ti.kind = typeSlice
		ti.elem = m.ToType(t.Elem())
	case reflect.Struct:
		n := t.NumField()
		fields := make([]Field, n)
		for i := 0; i < n; i++ {
			f := t.Field(i)
			if !f.IsExported() && offset == 0 {
				ti.offset = offsetForType(t)
			}
			fields[i].name = f.Name
			fields[i].anon = f.Anonymous
			fields[i].index = f.Index
			fields[i].offset = f.Offset
			fields[i].tag = string(f.Tag)
			fields[i].typ = m.ToType(f.Type)
		}
		ti.kind = typeStruct
		ti.fields = fields
	case reflect.Func:
		nin := t.NumIn()
		nout := t.NumOut()
		types := make([]*typeinfo, nin+nout)
		for i := 0; i < nin; i++ {
			types[i] = m.ToType(t.In(i))
		}
		for i := 0; i < nout; i++ {
			types[nin+i] = m.ToType(t.Out(i))
		}
		ti.kind = typeFunc
		ti.val = nin
		ti.variadic = t.IsVariadic()
		ti.args = types
	case reflect.Chan:
		ti.kind = typeChan
		ti.elem = m.ToType(t.Elem())
		switch t.ChanDir() {
		case reflect.RecvDir:
			ti.dir = recvDir
		case reflect.SendDir:
			ti.dir = sendDir
		case reflect.BothDir:
			ti.dir = bothDir
		}
	default:
		panic(fmt.Errorf("unsupported reflect.Kind (%s)", t.Kind()))
	}
	return ti
}

func boolint(x bool) int {
	if x {
		return 1
	}
	return 0
}

func named(t reflect.Type) bool {
	return t.Name() != ""
}
