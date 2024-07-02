//go:build durable

package subpkg3

import subpkg2 "github.com/dispatchrun/coroutine/compiler/testdata/subpkg2"
import _types "github.com/dispatchrun/coroutine/types"

type CustomInt int

func IsLess(a, b int) bool {
	return subpkg2.Less[CustomInt](CustomInt(a), CustomInt(b))
}
func init() {
	_types.RegisterFunc[func(a, b int) (_ bool)]("github.com/dispatchrun/coroutine/compiler/testdata/subpkg3.IsLess")
}
