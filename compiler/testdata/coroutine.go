//go:build !durable

package testdata

import (
	"time"
	"unsafe"

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

func SquareGeneratorTwice(n int) {
	SquareGenerator(n)
	SquareGenerator(n)
}

func SquareGeneratorTwiceLoop(n int) {
	for i := 0; i < 2; i++ {
		SquareGenerator(n)
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

func Shadowing(_ int) {
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
		switch i := 2; i {
		default:
			coroutine.Yield[int, any](i) // 2
		}
		coroutine.Yield[int, any](i) // 1
	}

	coroutine.Yield[int, any](i) // 0
	{
		i := 1
		{
			i := 2
			coroutine.Yield[int, any](i) // 2
		}
		coroutine.Yield[int, any](i) // 1
	}

	coroutine.Yield[int, any](i) // 0
	var j = i
	{
		j := 1
		coroutine.Yield[int, any](j) // 1
	}
	coroutine.Yield[int, any](j) // 0

	const k = 11
	{
		const k = 12
		{
			k := 13
			coroutine.Yield[int, any](k) // 13
		}
		coroutine.Yield[int, any](k) // 12
	}
	coroutine.Yield[int, any](k) // 11

	type foo uint16
	{
		type foo uint32
		coroutine.Yield[int, any](int(unsafe.Sizeof(foo(0)))) // 4
	}
	coroutine.Yield[int, any](int(unsafe.Sizeof(foo(0)))) // 2

	const siz = 1
	type baz [siz]uint8
	{
		type bar [siz]uint8
		coroutine.Yield[int, any](int(unsafe.Sizeof(bar{}))) // 1
		const siz = unsafe.Sizeof(bar{}) * 2
		type baz [siz]uint8
		coroutine.Yield[int, any](int(unsafe.Sizeof(baz{}))) // 2
	}
	coroutine.Yield[int, any](int(unsafe.Sizeof(baz{}))) // 1
}

func RangeSliceIndexGenerator(_ int) {
	for i := range []int{10, 20, 30} {
		coroutine.Yield[int, any](i)
	}
}

func RangeArrayIndexValueGenerator(_ int) {
	for i, v := range [...]int{10, 20, 30} {
		coroutine.Yield[int, any](i)
		coroutine.Yield[int, any](v)
	}
}

func TypeSwitchingGenerator(_ int) {
	for _, val := range []any{int8(10), int16(20), int32(30), int64(40)} {
		switch val.(type) {
		case int8:
			coroutine.Yield[int, any](1)
		case int16:
			coroutine.Yield[int, any](2)
		case int32:
			coroutine.Yield[int, any](4)
		case int64:
			coroutine.Yield[int, any](8)
		}
		switch v := val.(type) {
		case int8:
			coroutine.Yield[int, any](int(v))
		case int16:
			coroutine.Yield[int, any](int(v))
		case int32:
			coroutine.Yield[int, any](int(v))
		case int64:
			coroutine.Yield[int, any](int(v))
		}
	}
}

func LoopBreakAndContinue(_ int) {
	for i := 0; i < 10; i++ {
		if mod2 := i % 2; mod2 == 0 {
			continue
		}
		if i > 5 {
			break
		}
		coroutine.Yield[int, any](i)
	}

outer:
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			coroutine.Yield[int, any](j)
			switch j {
			case 0:
				continue
			case 1:
				switch i {
				case 0:
					continue outer
				case 1:
					break outer
				}
			}
		}
	}
}

func RangeOverMaps(n int) {
	m := map[int]int{}
	for range m {
		panic("unreachable")
	}
	for _ = range m {
		panic("unreachable")
	}
	for _, _ = range m {
		panic("unreachable")
	}
	m[n] = n * 10
	for range m {
		coroutine.Yield[int, any](0)
	}
	for k := range m {
		coroutine.Yield[int, any](k)
	}
	for k, v := range m {
		coroutine.Yield[int, any](k)
		coroutine.Yield[int, any](v)
	}

	// Map iteration order is not deterministic, so to
	// test iteration with a map with more than one element
	// we'll build a map and then successively delete keys
	// while yielding the length of the map.
	m2 := make(map[int]struct{}, n)
	for i := 0; i < n; i++ {
		m2[i] = struct{}{}
	}
	coroutine.Yield[int, any](len(m2))
	for k := range m2 {
		delete(m2, k)
		coroutine.Yield[int, any](len(m2))
	}
}

func Range(n int, do func(int)) {
	for i := 0; i < n; i++ {
		do(i)
	}
}

func Double(n int) {
	coroutine.Yield[int, any](2 * n)
}

// func RangeTriple(n int) {
// 	Range(n, func(i int) {
// 		coroutine.Yield[int, any](3 * i)
// 	})
// }

func RangeTripleFuncValue(n int) {
	f := func(i int) {
		coroutine.Yield[int, any](3 * i)
	}
	Range(n, f)
}

func Range10Closure() {
	i := 0
	n := 10
	f := func() bool {
		if i < n {
			coroutine.Yield[int, any](i)
			i++
			return true
		}
		return false
	}

	for f() {
	}
}

func Select(n int) {
	select {
	default:
		coroutine.Yield[int, any](-1)
	}

	for i := 0; i < n; i++ {
		select {
		case <-time.After(0):
			if i >= 5 {
				break
			}
			coroutine.Yield[int, any](i)
		case <-time.After(1 * time.Second):
			panic("unreachable")
		}

	foo:
		select {
		case <-time.After(0):
			if i >= 6 {
				break foo
			}
			coroutine.Yield[int, any](i * 10)
		}
	}

	select {
	case <-time.After(0):
		for j := 0; j < 3; j++ {
			coroutine.Yield[int, any](j)
		}
	}
}
