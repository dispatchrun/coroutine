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
