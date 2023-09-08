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

func EvenSquareGenerator(n int) {
	for i := 1; i <= n; i++ {
		if mod2 := i % 2; mod2 == 0 {
			coroutine.Yield[int, any](i * i)
		}
	}
}

func NestedLoops(n int) {
	for i := 1; i <= n; i++ {
		for j := 1; j <= n; j++ {
			for k := 1; k <= n; k++ {
				coroutine.Yield[int, any](i * j * k)
			}
		}
	}
}

func FizzBuzzIfGenerator(n int) {
	for i := 1; i <= n; i++ {
		if i%3 == 0 && i%5 == 0 {
			coroutine.Yield[int, any](FizzBuzz)
		} else if i%3 == 0 {
			coroutine.Yield[int, any](Fizz)
		} else if mod5 := i % 5; mod5 == 0 {
			coroutine.Yield[int, any](Buzz)
		} else {
			coroutine.Yield[int, any](i)
		}
	}
}

func FizzBuzzSwitchGenerator(n int) {
	for i := 1; i <= n; i++ {
		switch {
		case i%3 == 0 && i%5 == 0:
			coroutine.Yield[int, any](FizzBuzz)
		case i%3 == 0:
			coroutine.Yield[int, any](Fizz)
		case i%5 == 0:
			coroutine.Yield[int, any](Buzz)
		default:
			coroutine.Yield[int, any](i)
		}
	}
}

func Shadowing(n int) {
	i := 0
	coroutine.Yield[int, any](i) // 0
	if i := 1; true {
		coroutine.Yield[int, any](i) // 1
	}
	coroutine.Yield[int, any](i) // 0
	for i := 1; i < 3; i++ {
		coroutine.Yield[int, any](i) // 1, 2
	}
	coroutine.Yield[int, any](i) // 0
	switch i := 1; i {
	case 1:
		coroutine.Yield[int, any](i) // 1
	}
	coroutine.Yield[int, any](i) // 0
}
