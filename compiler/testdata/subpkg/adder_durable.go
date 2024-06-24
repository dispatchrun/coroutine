//go:build durable

package subpkg

import _types "github.com/dispatchrun/coroutine/types"
//go:noinline
func Adder(n int) func(int) int {
	return func(x int) int {
		return x + n
	}
}
func init() {
	_types.RegisterFunc[func(n int) (_ func(int) int)]("github.com/dispatchrun/coroutine/compiler/testdata/subpkg.Adder")
	_types.RegisterClosure[func(x int) (_ int), struct {
		F  uintptr
		X0 int
	}]("github.com/dispatchrun/coroutine/compiler/testdata/subpkg.Adder.func1")
}
