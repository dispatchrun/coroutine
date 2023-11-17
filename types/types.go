package types

import (
	"fmt"
	"reflect"
	"sync"
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

type typemap struct {
	serdes *serdemap
	cache  doublemap[reflect.Type, *typeinfo]
}

func newTypeMap(serdes *serdemap) *typemap {
	return &typemap{serdes: serdes}
}

func (m *typemap) ToReflect(t *typeinfo) reflect.Type {
	if t.offset != 0 {
		return typeForOffset(t.offset)
	}
	if x, ok := m.cache.getV(t); ok {
		return x
	}

	var x reflect.Type
	switch t.kind {
	case coroutinev1.Kind_KIND_NIL:
	case coroutinev1.Kind_KIND_BOOL:
		x = reflect.TypeOf(false)
	case coroutinev1.Kind_KIND_INT:
		x = reflect.TypeOf(int(0))
	case coroutinev1.Kind_KIND_INT8:
		x = reflect.TypeOf(int8(0))
	case coroutinev1.Kind_KIND_INT16:
		x = reflect.TypeOf(int16(0))
	case coroutinev1.Kind_KIND_INT32:
		x = reflect.TypeOf(int32(0))
	case coroutinev1.Kind_KIND_INT64:
		x = reflect.TypeOf(int64(0))
	case coroutinev1.Kind_KIND_UINT:
		x = reflect.TypeOf(uint(0))
	case coroutinev1.Kind_KIND_UINT8:
		x = reflect.TypeOf(uint8(0))
	case coroutinev1.Kind_KIND_UINT16:
		x = reflect.TypeOf(uint16(0))
	case coroutinev1.Kind_KIND_UINT32:
		x = reflect.TypeOf(uint32(0))
	case coroutinev1.Kind_KIND_UINT64:
		x = reflect.TypeOf(uint64(0))
	case coroutinev1.Kind_KIND_UINTPTR:
		x = reflect.TypeOf(uintptr(0))
	case coroutinev1.Kind_KIND_FLOAT32:
		x = reflect.TypeOf(float32(0))
	case coroutinev1.Kind_KIND_FLOAT64:
		x = reflect.TypeOf(float64(0))
	case coroutinev1.Kind_KIND_COMPLEX64:
		x = reflect.TypeOf(complex64(0))
	case coroutinev1.Kind_KIND_COMPLEX128:
		x = reflect.TypeOf(complex128(0))
	case coroutinev1.Kind_KIND_STRING:
		x = reflect.TypeOf("")
	case coroutinev1.Kind_KIND_INTERFACE:
		x = typeof[interface{}]()
	case coroutinev1.Kind_KIND_POINTER:
		x = reflect.PointerTo(m.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_UNSAFE_POINTER:
		x = reflect.TypeOf(unsafe.Pointer(nil))
	case coroutinev1.Kind_KIND_MAP:
		x = reflect.MapOf(m.ToReflect(t.key), m.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_ARRAY:
		x = reflect.ArrayOf(t.len, m.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_SLICE:
		x = reflect.SliceOf(m.ToReflect(t.elem))
	case coroutinev1.Kind_KIND_STRUCT:
		fields := make([]reflect.StructField, len(t.fields))
		for i, f := range t.fields {
			fields[i].Name = f.name
			fields[i].PkgPath = f.pkgPath
			fields[i].Tag = reflect.StructTag(f.tag)
			fields[i].Index = f.index
			fields[i].Offset = f.offset
			fields[i].Anonymous = f.anon
			fields[i].Type = m.ToReflect(f.typ)
		}
		x = reflect.StructOf(fields)
	case coroutinev1.Kind_KIND_FUNC:
		params := make([]reflect.Type, len(t.params))
		for i, t := range t.params {
			params[i] = m.ToReflect(t)
		}
		results := make([]reflect.Type, len(t.results))
		for i, t := range t.results {
			results[i] = m.ToReflect(t)
		}
		x = reflect.FuncOf(params, results, t.variadic)
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
		x = reflect.ChanOf(dir, m.ToReflect(t.elem))
	default:
		panic("invalid type kind: " + t.kind.String())
	}

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

	ti := &typeinfo{
		name:    t.Name(),
		pkgPath: t.PkgPath(),
	}

	if t.Name() != "" {
		ti.offset = offsetForType(t)
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
			if !f.IsExported() && ti.offset == 0 {
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

type doublemap[K, V comparable] struct {
	fromK map[K]V
	fromV map[V]K

	mu sync.Mutex
}

func (m *doublemap[K, V]) getK(k K) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.fromK[k]
	return v, ok
}

func (m *doublemap[K, V]) getV(v V) (K, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	k, ok := m.fromV[v]
	return k, ok
}

func (m *doublemap[K, V]) add(k K, v V) V {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.fromK == nil {
		m.fromK = make(map[K]V)
		m.fromV = make(map[V]K)
	}
	m.fromK[k] = v
	m.fromV[v] = k
	return v
}
