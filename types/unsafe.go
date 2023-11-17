package types

import (
	"reflect"
	"unsafe"
)

// Used for unsafe access to internals of interface{} and reflect.Value.
type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// Used instead of reflect.SliceHeader to use an unsafe.Pointer instead of
// uintptr.
type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// Used for unsafe access to a function pointer and closure vars.
type function struct {
	addr unsafe.Pointer
	// closure vars follow...
}

// returns true iff type t would be inlined in an interface.
func inlined(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Func:
		return true
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Struct:
		return t.NumField() == 1 && inlined(t.Field(0).Type)
	case reflect.Array:
		return t.Len() == 1 && inlined(t.Elem())
	default:
		return false
	}
}

var staticuint64s unsafe.Pointer

func init() {
	zero := 0
	var x interface{} = zero
	staticuint64s = (*iface)(unsafe.Pointer(&x)).ptr
}

func static(p unsafe.Pointer) bool {
	return uintptr(p) >= uintptr(staticuint64s) && uintptr(p) < uintptr(staticuint64s)+256
}

func staticOffset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(staticuint64s))
}

func staticPointer(offset int) unsafe.Pointer {
	return unsafe.Add(staticuint64s, offset)
}

// namedType offset is the number of bytes from the address of the 'byte' type
// value to the ptr field of a reflect.Type. It is used to roundtrip named types
// for a given version of the program.
type namedTypeOffset uint64

func offsetForType(t reflect.Type) namedTypeOffset {
	tptr := (*iface)(unsafe.Pointer(&t)).ptr
	bptr := (*iface)(unsafe.Pointer(&byteT)).ptr
	return namedTypeOffset(uintptr(tptr) - uintptr(bptr))
}

func typeForOffset(offset namedTypeOffset) reflect.Type {
	biface := (*iface)(unsafe.Pointer(&byteT))
	tiface := &iface{
		typ: biface.typ,
		ptr: unsafe.Add(biface.ptr, offset),
	}
	return *(*reflect.Type)(unsafe.Pointer(tiface))
}
