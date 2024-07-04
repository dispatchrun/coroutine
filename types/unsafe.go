package types

import (
	"unsafe"

	"github.com/dispatchrun/coroutine/internal/reflectext"
)

// Used for unsafe access to a function pointer and closure vars.
type function struct {
	addr unsafe.Pointer
	// closure vars follow...
}

var staticuint64s unsafe.Pointer

func init() {
	zero := 0
	var x interface{} = zero
	staticuint64s = reflectext.IfacePtr(unsafe.Pointer(&x), nil)
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
