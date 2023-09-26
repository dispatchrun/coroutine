package gls

// getg is like the compiler intrisinc runtime.getg which retrieves the current
// goroutine object.
//
// https://github.com/golang/go/blob/a2647f08f0c4e540540a7ae1b9ba7e668e6fed80/src/runtime/HACKING.md?plain=1#L44-L54
func getg() uintptr
