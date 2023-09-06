package testdata

import "github.com/stealthrocket/coroutine"

//go:generate coroc

func Square(n int) {
	for i := 1; i <= n; i++ {
		coroutine.Yield[int, any](i * i)
	}
}
