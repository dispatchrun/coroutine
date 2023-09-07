//go:build !durable

package testdata

import (
	"github.com/stealthrocket/coroutine"
)

//go:generate coroc --output coroutine_durable.go --tags durable

func Identity(n int) {
	coroutine.Yield[int, any](n)
}

func SquareGenerator(n int) {
	for i := 1; i <= n; i++ {
		coroutine.Yield[int, any](i * i)
	}
}
