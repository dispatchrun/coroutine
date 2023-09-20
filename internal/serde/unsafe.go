package serde

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
