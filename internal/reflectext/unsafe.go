package reflectext

import (
	"reflect"
	"unsafe"
)

// SetSlice sets the slice data pointer, length and capacity.
func SetSlice(v reflect.Value, data unsafe.Pointer, len, cap int) {
	if v.Kind() != reflect.Slice {
		panic("not a slice")
	} else if !v.CanAddr() {
		panic("slice is not addressable")
	}
	type sliceHeader struct { // see reflect.SliceHeader
		data unsafe.Pointer
		len  int
		cap  int
	}
	*(*sliceHeader)(v.Addr().UnsafePointer()) = sliceHeader{data: data, len: len, cap: cap}
}

// FunctionHeader is the container for function pointers
// and closure vars.
type FunctionHeader struct {
	Addr unsafe.Pointer
	// closure vars follow...
}

// Used for unsafe access to internals of interface{} and reflect.Value.
type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// FIXME: don't export this
func IfacePtr(p unsafe.Pointer, t reflect.Type) unsafe.Pointer {
	i := (*iface)(p)
	if ifaceInline(t) {
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
	tptr := (*iface)(unsafe.Pointer(&t)).ptr
	bptr := (*iface)(unsafe.Pointer(&ByteType)).ptr
	return NamedTypeOffset(uintptr(tptr) - uintptr(bptr))
}

// NamedTypeForOffset gets the named type for an offset.
func NamedTypeForOffset(offset NamedTypeOffset) reflect.Type {
	biface := (*iface)(unsafe.Pointer(&ByteType))
	tiface := &iface{
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
	internedBase = IfacePtr(unsafe.Pointer(&x), nil)
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
