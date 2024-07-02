//go:build !durable

package subpkg3

import "github.com/dispatchrun/coroutine/compiler/testdata/subpkg2"

type CustomInt int

func IsLess(a, b int) bool {
	return subpkg2.Less[CustomInt](CustomInt(a), CustomInt(b))
}
