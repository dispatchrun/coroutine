//go:build !durable

package subpkg

func Adder(n int) func(int) int {
	return func(x int) int {
		return x + n
	}
}
