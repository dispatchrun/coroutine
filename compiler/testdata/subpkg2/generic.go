//go:build !durable

package subpkg2

import "cmp"

func Less[T cmp.Ordered](a, b T) bool {
	return a < b
}
