//go:build durable

package subpkg2

import (
	cmp "cmp"
	subpkg3 "github.com/dispatchrun/coroutine/compiler/testdata/subpkg3"
)
import _types "github.com/dispatchrun/coroutine/types"

func Less[T cmp.Ordered](a, b T) bool {
	return a < b
}
func init() {
	_types.RegisterFunc[func(a, b subpkg3.CustomInt) (_ bool)]("github.com/dispatchrun/coroutine/compiler/testdata/subpkg2.Less[go.shape.int]")
}
