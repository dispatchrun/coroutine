package reflectext

import (
	"reflect"
	"unsafe"
)

// SliceValue is a wrapper for a slice value that adds the
// ability to set the underlying data pointer, length and
// capacity, without allocating memory and regardless of
// the type.
type SliceValue struct{ reflect.Value }

// SetSlice sets the slice data pointer, length and capacity.
func (v SliceValue) SetSlice(data unsafe.Pointer, len, cap int) {
	if v.Kind() != reflect.Slice {
		panic("not a slice")
	} else if !v.CanAddr() {
		panic("slice is not addressable")
	}
	*(*sliceHeader)(v.Addr().UnsafePointer()) = sliceHeader{data: data, len: len, cap: cap}
}

// StructValue is a wrapper for a struct value that provides
// access to unexported fields without the read-only flag.
type StructValue struct {
	reflect.Value

	base unsafe.Pointer
}

// Field returns the i'th field of the struct v.
// It panics if v's Kind is not Struct or i is out of range.
func (v *StructValue) Field(i int) reflect.Value {
	if v.Kind() != reflect.Struct {
		panic("not a struct")
	} else if i > v.NumField() {
		panic("field out of range")
	}
	if v.base == nil {
		// This requires at least one allocation. Cache the base
		// pointer so that we only pay the cost once per struct
		// during iteration rather than once per field.
		v.base = unsafeInterfacePointer(v.Value)
	}
	f := v.Type().Field(i)
	return reflect.NewAt(f.Type, unsafe.Add(v.base, f.Offset)).Elem()
}

// FunctionValue is a wrapper for a function value that
// provides access to closure vars.
type FunctionValue struct{ reflect.Value }

// Closure retrieves the function's closure variables as a struct value.
//
// The first field in the struct is the address of the function, and the
// remaining fields hold closure vars.
//
// Closure vars are only available for functions that have type information
// registered at runtime. See RegisterClosure for more information.
func (v FunctionValue) Closure() (reflect.Value, bool) {
	if v.Kind() != reflect.Func {
		panic("not a func")
	}
	addr := v.UnsafePointer()
	if f := FuncByAddr(uintptr(addr)); f == nil {
		// function not found at addr
	} else if f.Type == nil {
		// function type info not registered
	} else if f.Closure != nil {
		fh := *(**FunctionHeader)(unsafeInterfacePointer(v.Value))
		if fh.Addr != addr {
			panic("invalid closure")
		}
		closure := reflect.NewAt(f.Closure, unsafe.Pointer(fh)).Elem()
		return closure, true
	}
	return reflect.Value{}, false
}

// InterfacePointer extracts the data pointer from an interface.
func InterfacePointer(v reflect.Value) unsafe.Pointer {
	if v.Kind() != reflect.Interface {
		panic("not an interface")
	}
	return unsafeInterfacePointer(v)
}

func unsafeInterfacePointer(v reflect.Value) unsafe.Pointer {
	vi := v.Interface()
	i := (*interfaceHeader)(unsafe.Pointer(&vi))
	if ifaceInline(reflect.TypeOf(vi)) {
		return unsafe.Pointer(&i.ptr)
	}
	return i.ptr
}

func ifaceInline(t reflect.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind() {
	case reflect.Func:
		return true
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Struct:
		return t.NumField() == 1 && ifaceInline(t.Field(0).Type)
	case reflect.Array:
		return t.Len() == 1 && ifaceInline(t.Elem())
	default:
		return false
	}
}

// NamedTypeOffset is an opaque identifier for a named type.
//
// It is used to round-trip named types for a given version of the program.
type NamedTypeOffset uint64

// OffsetForNamedType gets the offset of a named type.
func OffsetForNamedType(t reflect.Type) NamedTypeOffset {
	// FIXME: investigate
	// if t.Name() == "" {
	//   panic("not a named type")
	// }
	tptr := (*interfaceHeader)(unsafe.Pointer(&t)).ptr
	bptr := (*interfaceHeader)(unsafe.Pointer(&ByteType)).ptr
	return NamedTypeOffset(uintptr(tptr) - uintptr(bptr))
}

// NamedTypeForOffset gets the named type for an offset.
func NamedTypeForOffset(offset NamedTypeOffset) reflect.Type {
	biface := (*interfaceHeader)(unsafe.Pointer(&ByteType))
	tiface := &interfaceHeader{
		typ: biface.typ,
		ptr: unsafe.Add(biface.ptr, offset),
	}
	return *(*reflect.Type)(unsafe.Pointer(tiface))
}

const internedCount = 256

var internedBase unsafe.Pointer

func init() {
	zero := 0
	var x interface{} = zero
	internedBase = (*interfaceHeader)(unsafe.Pointer(&x)).ptr
}

func InternedInt(p unsafe.Pointer) (int, bool) {
	if interned(p) {
		return internedOffset(p), true
	}
	return 0, false
}

func interned(p unsafe.Pointer) bool {
	return uintptr(p) >= uintptr(internedBase) && uintptr(p) < uintptr(internedBase)+internedCount
}

func internedOffset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(internedBase))
}

func InternedIntPointer(offset int) unsafe.Pointer {
	return unsafe.Add(internedBase, offset)
}

// FunctionHeader is the container for function pointers
// and closure vars.
type FunctionHeader struct {
	Addr unsafe.Pointer
	// closure vars follow...
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

type interfaceHeader struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}
