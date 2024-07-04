package types

import (
	"fmt"
	"reflect"
	"unsafe"

	coroutinev1 "github.com/dispatchrun/coroutine/gen/proto/go/coroutine/v1"
	"github.com/dispatchrun/coroutine/internal/reflectext"
)

type typeid = uint32

type typemap struct {
	serdes  *serdemap
	strings *stringmap

	types []*coroutinev1.Type
	cache doublemap[typeid, reflect.Type]
}

func newTypeMap(serdes *serdemap, strings *stringmap, types []*coroutinev1.Type) *typemap {
	return &typemap{
		serdes:  serdes,
		strings: strings,
		types:   types,
	}
}

func (m *typemap) register(t *coroutinev1.Type) typeid {
	m.types = append(m.types, t)
	id := typeid(len(m.types)) // note that IDs start at 1
	return id
}

func (m *typemap) lookup(id typeid) *coroutinev1.Type {
	if id == 0 || id > uint32(len(m.types)) {
		return nil
	}
	return m.types[id-1]
}

func (m *typemap) ToReflect(id typeid) reflect.Type {
	if x, ok := m.cache.getK(id); ok {
		return x
	}

	t := m.lookup(id)
	if t == nil {
		panic(fmt.Sprintf("type %d not found", id))
	}

	if t.CustomSerializer > 0 {
		if t.MemoryOffset != 0 {
			et := typeForOffset(namedTypeOffset(t.MemoryOffset))
			if t.Kind == coroutinev1.Kind_KIND_POINTER {
				et = reflect.PointerTo(et)
			}
			return et
		}
		id := serdeid(t.CustomSerializer)
		return m.serdes.serdeByID(id).typ
	}

	if t.MemoryOffset != 0 {
		return typeForOffset(namedTypeOffset(t.MemoryOffset))
	}

	var x reflect.Type
	switch t.Kind {
	case coroutinev1.Kind_KIND_BOOL:
		x = boolT

	case coroutinev1.Kind_KIND_INT:
		x = intT

	case coroutinev1.Kind_KIND_INT8:
		x = int8T

	case coroutinev1.Kind_KIND_INT16:
		x = int16T

	case coroutinev1.Kind_KIND_INT32:
		x = int32T

	case coroutinev1.Kind_KIND_INT64:
		x = int64T

	case coroutinev1.Kind_KIND_UINT:
		x = uintT

	case coroutinev1.Kind_KIND_UINT8:
		x = uint8T

	case coroutinev1.Kind_KIND_UINT16:
		x = uint16T

	case coroutinev1.Kind_KIND_UINT32:
		x = uint32T

	case coroutinev1.Kind_KIND_UINT64:
		x = uint64T

	case coroutinev1.Kind_KIND_UINTPTR:
		x = uintptrT

	case coroutinev1.Kind_KIND_FLOAT32:
		x = float32T

	case coroutinev1.Kind_KIND_FLOAT64:
		x = float64T

	case coroutinev1.Kind_KIND_COMPLEX64:
		x = complex64T

	case coroutinev1.Kind_KIND_COMPLEX128:
		x = complex128T

	case coroutinev1.Kind_KIND_STRING:
		x = stringT

	case coroutinev1.Kind_KIND_INTERFACE:
		x = anyT

	case coroutinev1.Kind_KIND_POINTER:
		x = reflect.PointerTo(m.ToReflect(typeid(t.Elem)))

	case coroutinev1.Kind_KIND_UNSAFE_POINTER:
		x = unsafePointerT

	case coroutinev1.Kind_KIND_MAP:
		x = reflect.MapOf(m.ToReflect(typeid(t.Key)), m.ToReflect(typeid(t.Elem)))

	case coroutinev1.Kind_KIND_ARRAY:
		x = reflect.ArrayOf(int(t.Length), m.ToReflect(typeid(t.Elem)))

	case coroutinev1.Kind_KIND_SLICE:
		x = reflect.SliceOf(m.ToReflect(typeid(t.Elem)))

	case coroutinev1.Kind_KIND_STRUCT:
		fields := make([]reflect.StructField, len(t.Fields))
		for i, f := range t.Fields {
			fields[i].Name = m.strings.Lookup(stringid(f.Name))
			fields[i].PkgPath = m.strings.Lookup(stringid(f.Package))
			fields[i].Tag = reflect.StructTag(f.Tag)

			index := make([]int, len(f.Index))
			for i, idx := range f.Index {
				index[i] = int(idx)
			}
			fields[i].Index = index
			fields[i].Offset = uintptr(f.Offset)
			fields[i].Anonymous = f.Anonymous
			fields[i].Type = m.ToReflect(typeid(f.Type))
		}
		x = reflect.StructOf(fields)

	case coroutinev1.Kind_KIND_FUNC:
		params := make([]reflect.Type, len(t.Params))
		for i, t := range t.Params {
			params[i] = m.ToReflect(typeid(t))
		}
		results := make([]reflect.Type, len(t.Results))
		for i, t := range t.Results {
			results[i] = m.ToReflect(typeid(t))
		}
		x = reflect.FuncOf(params, results, t.Variadic)

	case coroutinev1.Kind_KIND_CHAN:
		var dir reflect.ChanDir
		switch t.ChanDir {
		case coroutinev1.ChanDir_CHAN_DIR_RECV:
			dir = reflect.RecvDir
		case coroutinev1.ChanDir_CHAN_DIR_SEND:
			dir = reflect.SendDir
		case coroutinev1.ChanDir_CHAN_DIR_BOTH:
			dir = reflect.BothDir
		default:
			panic("invalid chan dir: " + t.ChanDir.String())
		}
		x = reflect.ChanOf(dir, m.ToReflect(typeid(t.Elem)))

	default:
		panic("invalid type kind: " + t.Kind.String())
	}

	m.cache.add(id, x)
	return x
}

func (m *typemap) ToType(t reflect.Type) typeid {
	if t == nil {
		panic("nil reflect.Type")
	}

	if x, ok := m.cache.getV(t); ok {
		return x
	}

	ti := &coroutinev1.Type{
		Name:    m.strings.Intern(t.Name()),
		Package: m.strings.Intern(t.PkgPath()),
		Kind:    kindOf(t.Kind()),
	}

	if t.Name() != "" || t.Kind() == reflect.Interface {
		ti.MemoryOffset = uint64(offsetForType(t))
	}

	// Register the incomplete type now before recursing,
	// in case the type references itself.
	id := m.register(ti)
	m.cache.add(id, t)

	// Types with custom serializers are opaque.
	if s, ok := m.serdes.serdeByType(t); ok {
		if t.Name() == "" && t.Kind() == reflect.Pointer {
			if et := t.Elem(); et.Name() != "" {
				ti.MemoryOffset = uint64(offsetForType(et))
				ti.Name = m.strings.Intern(et.Name())
				ti.Package = m.strings.Intern(et.PkgPath())
			}
		}
		ti.CustomSerializer = s.id
		return id
	}

	switch t.Kind() { // note: already checked above in kindOf() call
	case reflect.Array:
		ti.Length = int64(t.Len())
		ti.Elem = m.ToType(t.Elem())

	case reflect.Map:
		ti.Key = m.ToType(t.Key())
		ti.Elem = m.ToType(t.Elem())

	case reflect.Pointer:
		ti.Elem = m.ToType(t.Elem())

	case reflect.Slice:
		ti.Elem = m.ToType(t.Elem())

	case reflect.Struct:
		ti.Fields = make([]*coroutinev1.Field, t.NumField())
		for i := range ti.Fields {
			f := t.Field(i)
			if !f.IsExported() && ti.MemoryOffset == 0 {
				ti.MemoryOffset = uint64(offsetForType(t))
			}
			index := make([]int32, len(f.Index))
			for j := range index {
				index[j] = int32(f.Index[j])
			}
			ti.Fields[i] = &coroutinev1.Field{
				Name:      m.strings.Intern(f.Name),
				Package:   m.strings.Intern(f.PkgPath),
				Offset:    uint64(f.Offset),
				Anonymous: f.Anonymous,
				Tag:       string(f.Tag),
				Type:      m.ToType(f.Type),
				Index:     index,
			}
		}

	case reflect.Func:
		ti.Params = make([]uint32, t.NumIn())
		for i := range ti.Params {
			ti.Params[i] = m.ToType(t.In(i))
		}
		ti.Results = make([]uint32, t.NumOut())
		for i := range ti.Results {
			ti.Results[i] = m.ToType(t.Out(i))
		}
		ti.Variadic = t.IsVariadic()

	case reflect.Chan:
		ti.Elem = m.ToType(t.Elem())
		switch t.ChanDir() {
		case reflect.RecvDir:
			ti.ChanDir = coroutinev1.ChanDir_CHAN_DIR_RECV
		case reflect.SendDir:
			ti.ChanDir = coroutinev1.ChanDir_CHAN_DIR_SEND
		case reflect.BothDir:
			ti.ChanDir = coroutinev1.ChanDir_CHAN_DIR_BOTH
		}
	}
	return id
}

func kindOf(kind reflect.Kind) coroutinev1.Kind {
	switch kind {
	case reflect.Invalid:
		panic("can't handle reflect.Invalid")
	case reflect.Bool:
		return coroutinev1.Kind_KIND_BOOL
	case reflect.Int:
		return coroutinev1.Kind_KIND_INT
	case reflect.Int8:
		return coroutinev1.Kind_KIND_INT8
	case reflect.Int16:
		return coroutinev1.Kind_KIND_INT16
	case reflect.Int32:
		return coroutinev1.Kind_KIND_INT32
	case reflect.Int64:
		return coroutinev1.Kind_KIND_INT64
	case reflect.Uint:
		return coroutinev1.Kind_KIND_UINT
	case reflect.Uint8:
		return coroutinev1.Kind_KIND_UINT8
	case reflect.Uint16:
		return coroutinev1.Kind_KIND_UINT16
	case reflect.Uint32:
		return coroutinev1.Kind_KIND_UINT32
	case reflect.Uint64:
		return coroutinev1.Kind_KIND_UINT64
	case reflect.Uintptr:
		return coroutinev1.Kind_KIND_UINTPTR
	case reflect.Float32:
		return coroutinev1.Kind_KIND_FLOAT32
	case reflect.Float64:
		return coroutinev1.Kind_KIND_FLOAT64
	case reflect.Complex64:
		return coroutinev1.Kind_KIND_COMPLEX64
	case reflect.Complex128:
		return coroutinev1.Kind_KIND_COMPLEX128
	case reflect.String:
		return coroutinev1.Kind_KIND_STRING
	case reflect.Interface:
		return coroutinev1.Kind_KIND_INTERFACE
	case reflect.Array:
		return coroutinev1.Kind_KIND_ARRAY
	case reflect.Map:
		return coroutinev1.Kind_KIND_MAP
	case reflect.Pointer:
		return coroutinev1.Kind_KIND_POINTER
	case reflect.UnsafePointer:
		return coroutinev1.Kind_KIND_UNSAFE_POINTER
	case reflect.Slice:
		return coroutinev1.Kind_KIND_SLICE
	case reflect.Struct:
		return coroutinev1.Kind_KIND_STRUCT
	case reflect.Func:
		return coroutinev1.Kind_KIND_FUNC
	case reflect.Chan:
		return coroutinev1.Kind_KIND_CHAN
	default:
		panic(fmt.Errorf("unsupported reflect.Kind (%s)", kind))
	}
}

type funcid = uint32

type funcmap struct {
	types   *typemap
	strings *stringmap

	funcs []*coroutinev1.Function
	cache doublemap[typeid, *reflectext.Func]
}

func newFuncMap(types *typemap, strings *stringmap, funcs []*coroutinev1.Function) *funcmap {
	return &funcmap{
		types:   types,
		strings: strings,
		funcs:   funcs,
	}
}

func (m *funcmap) register(f *coroutinev1.Function) typeid {
	m.funcs = append(m.funcs, f)
	id := funcid(len(m.funcs)) // note that IDs start at 1
	return id
}

func (m *funcmap) lookup(id funcid) *coroutinev1.Function {
	if id == 0 || id > uint32(len(m.funcs)) {
		return nil
	}
	return m.funcs[id-1]
}

func (m *funcmap) ToFunc(id funcid) *reflectext.Func {
	if x, ok := m.cache.getK(id); ok {
		return x
	}
	cf := m.lookup(id)
	if cf == nil {
		panic(fmt.Sprintf("function ID %d not found", id))
	}
	name := m.strings.Lookup(cf.Name)
	f := reflectext.FuncByName(name)
	if f == nil {
		panic(fmt.Sprintf("function %s not found", name))
	}
	return f
}

func (m *funcmap) RegisterAddr(addr unsafe.Pointer) (id funcid, closureType reflect.Type) {
	f := reflectext.FuncByAddr(uintptr(addr))
	if f == nil {
		panic(fmt.Sprintf("function not found at address %v", addr))
	}
	if f.Type == nil {
		panic(fmt.Sprintf("type information not registered for function %s (%p)", f.Name, addr))
	}

	var closureTypeID typeid
	if f.Closure != nil {
		closureTypeID = m.types.ToType(f.Closure)
	}

	id = m.register(&coroutinev1.Function{
		Name:    m.strings.Intern(f.Name),
		Type:    m.types.ToType(f.Type),
		Closure: closureTypeID,
	})

	return id, f.Closure
}

type doublemap[K, V comparable] struct {
	fromK map[K]V
	fromV map[V]K
}

func (m *doublemap[K, V]) getK(k K) (V, bool) {
	v, ok := m.fromK[k]
	return v, ok
}

func (m *doublemap[K, V]) getV(v V) (K, bool) {
	k, ok := m.fromV[v]
	return k, ok
}

func (m *doublemap[K, V]) add(k K, v V) V {
	if m.fromK == nil {
		m.fromK = make(map[K]V)
		m.fromV = make(map[V]K)
	}
	m.fromK[k] = v
	m.fromV[v] = k
	return v
}

var (
	anyT = reflect.TypeFor[any]()

	boolT = reflect.TypeFor[bool]()

	intT   = reflect.TypeFor[int]()
	int8T  = reflect.TypeFor[int8]()
	int16T = reflect.TypeFor[int16]()
	int32T = reflect.TypeFor[int32]()
	int64T = reflect.TypeFor[int64]()

	uintT   = reflect.TypeFor[uint]()
	uint8T  = reflect.TypeFor[uint8]()
	uint16T = reflect.TypeFor[uint16]()
	uint32T = reflect.TypeFor[uint32]()
	uint64T = reflect.TypeFor[uint64]()

	float32T = reflect.TypeFor[float32]()
	float64T = reflect.TypeFor[float64]()

	complex64T  = reflect.TypeFor[complex64]()
	complex128T = reflect.TypeFor[complex128]()

	byteT   = reflect.TypeFor[byte]()
	stringT = reflect.TypeFor[string]()

	uintptrT       = reflect.TypeFor[uintptr]()
	unsafePointerT = reflect.TypeFor[unsafe.Pointer]()

	reflectValueT = reflect.TypeFor[reflect.Value]()
	reflectTypeT  = reflect.TypeFor[reflect.Type]()
)
