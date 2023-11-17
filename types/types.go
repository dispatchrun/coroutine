package types

import (
	"fmt"
	"reflect"
	"unsafe"

	coroutinev1 "github.com/stealthrocket/coroutine/gen/proto/go/coroutine/v1"
)

// typeinfo represents a type in the serialization format. It is a
// one-size-fits-all struct that contains everything needed to reconstruct a
// reflect.Type. This is because an interface-based approach is more difficult
// to get right, and we will be revamping serde anyway.
type typeinfo struct {
	kind coroutinev1.Kind

	name string

	pkgPath string

	// Only present for named types. See documentation of [namedTypeOffset].
	offset namedTypeOffset

	// elem is the type of element for array, slice, pointer, chan and map types
	elem *typeinfo

	// key is key type for a map type
	key *typeinfo

	// fields is the set of fields for a struct type
	fields []Field

	// params/results are the inputs and outputs for a function type
	params  []*typeinfo
	results []*typeinfo

	// len is the length of an array type
	len int

	// variadic is true if the type represents a function with a variadic argument
	variadic bool

	// custom is true if a custom serializer has been registered for this type
	custom bool

	// dir is the direction of a channel type
	dir coroutinev1.ChanDir
}

type Field struct {
	name    string
	pkgPath string
	typ     *typeinfo
	index   []int
	offset  uintptr
	anon    bool
	tag     string
}

func (t *typeinfo) reflectType(tm *typemap) reflect.Type {
	if t.offset != 0 {
		return typeForOffset(t.offset)
	}

	switch t.kind {
	case coroutinev1.Kind_KIND_NIL:
		return nil
	case coroutinev1.Kind_KIND_BOOL:
		return reflect.TypeOf(false)
	case coroutinev1.Kind_KIND_INT:
		return reflect.TypeOf(int(0))
	case coroutinev1.Kind_KIND_INT8:
		return reflect.TypeOf(int8(0))
	case coroutinev1.Kind_KIND_INT16:
		return reflect.TypeOf(int16(0))
	case coroutinev1.Kind_KIND_INT32:
		return reflect.TypeOf(int32(0))
	case coroutinev1.Kind_KIND_INT64:
		return reflect.TypeOf(int64(0))
	case coroutinev1.Kind_KIND_UINT:
		return reflect.TypeOf(uint(0))
	case coroutinev1.Kind_KIND_UINT8:
		return reflect.TypeOf(uint8(0))
	case coroutinev1.Kind_KIND_UINT16:
		return reflect.TypeOf(uint16(0))
	case coroutinev1.Kind_KIND_UINT32:
		return reflect.TypeOf(uint32(0))
	case coroutinev1.Kind_KIND_UINT64:
		return reflect.TypeOf(uint64(0))
	case coroutinev1.Kind_KIND_UINTPTR:
		return reflect.TypeOf(uintptr(0))
	case coroutinev1.Kind_KIND_FLOAT32:
		return reflect.TypeOf(float32(0))
	case coroutinev1.Kind_KIND_FLOAT64:
		return reflect.TypeOf(float64(0))
	case coroutinev1.Kind_KIND_COMPLEX64:
		return reflect.TypeOf(complex64(0))
	case coroutinev1.Kind_KIND_COMPLEX128:
		return reflect.TypeOf(complex128(0))
	case coroutinev1.Kind_KIND_STRING:
		return reflect.TypeOf("")
	case coroutinev1.Kind_KIND_INTERFACE:
		return typeof[interface{}]()
	case coroutinev1.Kind_KIND_POINTER:
		return reflect.PointerTo(tm.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_UNSAFE_POINTER:
		return reflect.TypeOf(unsafe.Pointer(nil))
	case coroutinev1.Kind_KIND_MAP:
		return reflect.MapOf(tm.ToReflect(t.key), tm.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_ARRAY:
		return reflect.ArrayOf(t.len, tm.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_SLICE:
		return reflect.SliceOf(tm.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_STRUCT:
		fields := make([]reflect.StructField, len(t.fields))
		for i, f := range t.fields {
			fields[i].Name = f.name
			fields[i].PkgPath = f.pkgPath
			fields[i].Tag = reflect.StructTag(f.tag)
			fields[i].Index = f.index
			fields[i].Offset = f.offset
			fields[i].Anonymous = f.anon
			fields[i].Type = tm.ToReflect(f.typ)
		}
		return reflect.StructOf(fields)
	case coroutinev1.Kind_KIND_FUNC:
		params := make([]reflect.Type, len(t.params))
		for i, t := range t.params {
			params[i] = tm.ToReflect(t)
		}
		results := make([]reflect.Type, len(t.results))
		for i, t := range t.results {
			results[i] = tm.ToReflect(t)
		}
		return reflect.FuncOf(params, results, t.variadic)
	case coroutinev1.Kind_KIND_CHAN:
		var dir reflect.ChanDir
		switch t.dir {
		case coroutinev1.ChanDir_CHAN_DIR_RECV:
			dir = reflect.RecvDir
		case coroutinev1.ChanDir_CHAN_DIR_SEND:
			dir = reflect.SendDir
		case coroutinev1.ChanDir_CHAN_DIR_BOTH:
			dir = reflect.BothDir
		default:
			panic("invalid chan dir: " + t.dir.String())
		}
		return reflect.ChanOf(dir, tm.ToReflect(t.elem))
	default:
		panic("invalid type kind: " + t.kind.String())
	}
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
		return m.cache.add(t, &typeinfo{kind: coroutinev1.Kind_KIND_NIL})
	}

	var offset namedTypeOffset
	if t.Name() != "" {
		offset = offsetForType(t)
		// Technically types with an offset do not need more information
		// than that. However for debugging purposes also generate the
		// rest of the type information.
	}

	ti := &typeinfo{
		name:    t.Name(),
		pkgPath: t.PkgPath(),
		offset:  offset,
	}

	if _, ok := m.serdes.serdeOf(t); ok {
		ti.custom = true
	}

	m.cache.add(t, ti) // add now for recursion
	switch t.Kind() {
	case reflect.Invalid:
		panic("can't handle reflect.Invalid")
	case reflect.Bool:
		ti.kind = coroutinev1.Kind_KIND_BOOL
	case reflect.Int:
		ti.kind = coroutinev1.Kind_KIND_INT
	case reflect.Int8:
		ti.kind = coroutinev1.Kind_KIND_INT8
	case reflect.Int16:
		ti.kind = coroutinev1.Kind_KIND_INT16
	case reflect.Int32:
		ti.kind = coroutinev1.Kind_KIND_INT32
	case reflect.Int64:
		ti.kind = coroutinev1.Kind_KIND_INT64
	case reflect.Uint:
		ti.kind = coroutinev1.Kind_KIND_UINT
	case reflect.Uint8:
		ti.kind = coroutinev1.Kind_KIND_UINT8
	case reflect.Uint16:
		ti.kind = coroutinev1.Kind_KIND_UINT16
	case reflect.Uint32:
		ti.kind = coroutinev1.Kind_KIND_UINT32
	case reflect.Uint64:
		ti.kind = coroutinev1.Kind_KIND_UINT64
	case reflect.Uintptr:
		ti.kind = coroutinev1.Kind_KIND_UINTPTR
	case reflect.Float32:
		ti.kind = coroutinev1.Kind_KIND_FLOAT32
	case reflect.Float64:
		ti.kind = coroutinev1.Kind_KIND_FLOAT64
	case reflect.Complex64:
		ti.kind = coroutinev1.Kind_KIND_COMPLEX64
	case reflect.Complex128:
		ti.kind = coroutinev1.Kind_KIND_COMPLEX128
	case reflect.String:
		ti.kind = coroutinev1.Kind_KIND_STRING
	case reflect.Interface:
		ti.kind = coroutinev1.Kind_KIND_INTERFACE
	case reflect.Array:
		ti.kind = coroutinev1.Kind_KIND_ARRAY
		ti.len = t.Len()
		ti.elem = m.ToType(t.Elem())
	case reflect.Map:
		ti.kind = coroutinev1.Kind_KIND_MAP
		ti.key = m.ToType(t.Key())
		ti.elem = m.ToType(t.Elem())
	case reflect.Pointer:
		ti.kind = coroutinev1.Kind_KIND_POINTER
		ti.elem = m.ToType(t.Elem())
	case reflect.UnsafePointer:
		ti.kind = coroutinev1.Kind_KIND_UNSAFE_POINTER
	case reflect.Slice:
		ti.kind = coroutinev1.Kind_KIND_SLICE
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
		ti.kind = coroutinev1.Kind_KIND_STRUCT
		ti.fields = fields
	case reflect.Func:
		params := make([]*typeinfo, t.NumIn())
		for i := range params {
			params[i] = m.ToType(t.In(i))
		}
		results := make([]*typeinfo, t.NumOut())
		for i := range results {
			results[i] = m.ToType(t.Out(i))
		}
		ti.kind = coroutinev1.Kind_KIND_FUNC
		ti.params = params
		ti.results = results
		ti.variadic = t.IsVariadic()
	case reflect.Chan:
		ti.kind = coroutinev1.Kind_KIND_CHAN
		ti.elem = m.ToType(t.Elem())
		switch t.ChanDir() {
		case reflect.RecvDir:
			ti.dir = coroutinev1.ChanDir_CHAN_DIR_RECV
		case reflect.SendDir:
			ti.dir = coroutinev1.ChanDir_CHAN_DIR_SEND
		case reflect.BothDir:
			ti.dir = coroutinev1.ChanDir_CHAN_DIR_BOTH
		}
	default:
		panic(fmt.Errorf("unsupported reflect.Kind (%s)", t.Kind()))
	}
	return ti
}
