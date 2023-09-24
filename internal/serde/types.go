package serde

import (
	"fmt"
	"reflect"
)

type typekind int

const (
	typeNone typekind = iota
	typeCustom
	typeBasic
	typePointer
	typeMap
	typeArray
	typeSlice
	typeStruct
	typeFunc
)

// typeinfo represents a type in the serialization format. It is a
// one-size-fits-all struct that contains everything needed to reconstruct a
// reflect.Type. This is because an interface-based approach is more difficult
// to get right, and we will be revamping serde anyway.
type typeinfo struct {
	kind typekind
	// - typeCustom uses this field to store the index in the typemap of the
	//   custom type it represents.
	// - typeBasic uses it to store the reflect.Kind it represents.
	// - typeArray stores its length
	// - typeFunc uses it to store the number of input arguments and whether
	//   its variadic as the first bit.
	val int
	// typeArray, typeSlice, typePointer, and TypeMap use this field to
	// store the information about the type they contain.
	elem   *typeinfo
	key    *typeinfo   // typeMap only
	fields []Field     // typeStruct only
	args   []*typeinfo // typeFunc only
}

func (t *typeinfo) reflectType(tm *TypeMap) reflect.Type {
	switch t.kind {
	case typeNone:
		return nil
	case typeCustom:
		return tm.custom[t.val]
	case typeBasic:
		switch reflect.Kind(t.val) {
		case reflect.Bool:
			return reflect.TypeOf(false)
		case reflect.Int:
			return reflect.TypeOf(int(0))
		case reflect.Int64:
			return reflect.TypeOf(int64(0))
		case reflect.Int32:
			return reflect.TypeOf(int32(0))
		case reflect.Int16:
			return reflect.TypeOf(int16(0))
		case reflect.Int8:
			return reflect.TypeOf(int8(0))
		case reflect.Uint:
			return reflect.TypeOf(uint(0))
		case reflect.Uint64:
			return reflect.TypeOf(uint64(0))
		case reflect.Uint32:
			return reflect.TypeOf(uint32(0))
		case reflect.Uint16:
			return reflect.TypeOf(uint16(0))
		case reflect.Uint8:
			return reflect.TypeOf(uint8(0))
		case reflect.Uintptr:
			return reflect.TypeOf(uintptr(0))
		case reflect.Float64:
			return reflect.TypeOf(float64(0))
		case reflect.Float32:
			return reflect.TypeOf(float32(0))
		case reflect.Complex64:
			return reflect.TypeOf(complex64(0))
		case reflect.Complex128:
			return reflect.TypeOf(complex128(0))
		case reflect.String:
			return reflect.TypeOf("")
		case reflect.Interface:
			return typeof[interface{}]()
		default:
			panic("Basic type unknown")
		}
	case typePointer:
		return reflect.PointerTo(tm.ToReflect(t.elem))
	case typeMap:
		return reflect.MapOf(tm.ToReflect(t.key), tm.ToReflect(t.elem))
	case typeArray:
		return reflect.ArrayOf(t.val, tm.ToReflect(t.elem))
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
		variadic := (t.val & 1) > 0
		in := t.val >> 1
		insouts := make([]reflect.Type, len(t.args))
		for i, t := range t.args {
			insouts[i] = tm.ToReflect(t)
		}
		return reflect.FuncOf(insouts[:in], insouts[in:], variadic)
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

func (m *TypeMap) ToReflect(t *typeinfo) reflect.Type {
	if x, ok := m.cache.GetV(t); ok {
		return x
	}
	x := t.reflectType(m)
	m.cache.Add(x, t)
	return x
}

func (m *TypeMap) ToType(t reflect.Type) *typeinfo {
	if x, ok := m.cache.GetK(t); ok {
		return x
	}

	if t == nil {
		return &typeinfo{kind: typeNone}
	}

	if named(t) {
		panic(fmt.Errorf("named type should be registered (%s)", t))
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic("can't handle reflect.Invalid")
	case reflect.Bool,
		reflect.Int,
		reflect.Int64,
		reflect.Int32,
		reflect.Int16,
		reflect.Int8,
		reflect.Uint,
		reflect.Uint64,
		reflect.Uint32,
		reflect.Uint16,
		reflect.Uint8,
		reflect.Uintptr,
		reflect.Float64,
		reflect.Float32,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String,
		reflect.Interface:
		return &typeinfo{
			kind: typeBasic,
			val:  int(t.Kind()),
		}
	case reflect.Array:
		return &typeinfo{
			kind: typeArray,
			elem: m.ToType(t.Elem()),
			val:  t.Len(),
		}
	case reflect.Map:
		return &typeinfo{
			kind: typeMap,
			key:  m.ToType(t.Key()),
			elem: m.ToType(t.Elem()),
		}
	case reflect.Pointer:
		return &typeinfo{
			kind: typePointer,
			elem: m.ToType(t.Elem()),
		}
	case reflect.Slice:
		return &typeinfo{
			kind: typeSlice,
			elem: m.ToType(t.Elem()),
		}
	case reflect.Struct:
		n := t.NumField()
		fields := make([]Field, n)
		for i := 0; i < n; i++ {
			f := t.Field(i)
			// Unexported fields are not supported.
			if !f.IsExported() {
				panic(fmt.Errorf("struct with unexported fields should be registered (%s)", t))
			}
			fields[i].name = f.Name
			fields[i].anon = f.Anonymous
			fields[i].index = f.Index
			fields[i].offset = f.Offset
			fields[i].tag = string(f.Tag)
			fields[i].typ = m.ToType(f.Type)
		}
		return &typeinfo{
			kind:   typeStruct,
			fields: fields,
		}
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
		return &typeinfo{
			kind: typeFunc,
			val:  nin<<1 | boolint(t.IsVariadic()),
			args: types,
		}
	default:
		panic(fmt.Errorf("unsupported reflect.Kind (%s)", t.Kind()))
	}
}

func boolint(x bool) int {
	if x {
		return 1
	}
	return 0
}

func named(t reflect.Type) bool {
	return t.PkgPath() != ""
}
