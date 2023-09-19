// Code generated by coroc. DO NOT EDIT

//go:build durable

package testdata

import (
	"github.com/stealthrocket/coroutine"
	serde "github.com/stealthrocket/coroutine/serde"
	runtime "runtime"
	sync "sync"
	atomic "sync/atomic"
	syscall "syscall"
	time "time"
	unsafe "unsafe"
)

func Identity(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	if _f.IP > 0 {
		n = _f.Get(0).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	coroutine.Yield[int, any](n)
}

func SquareGenerator(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 1
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		for ; _o0 <= n; _o0, _f.IP = _o0+1, 2 {
			coroutine.Yield[int, any](_o0 * _o0)
		}
	}
}

func SquareGeneratorTwice(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	if _f.IP > 0 {
		n = _f.Get(0).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		SquareGenerator(n)
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		SquareGenerator(n)
	}
}

func SquareGeneratorTwiceLoop(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 0
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		for ; _o0 < 2; _o0, _f.IP = _o0+1, 2 {
			SquareGenerator(n)
		}
	}
}

func EvenSquareGenerator(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	var _o1 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)
		_o1 = _f.Get(2).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_f.Set(2, _o1)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 1
		_f.IP = 2
		fallthrough
	case _f.IP < 4:
		for ; _o0 <= n; _o0, _f.IP = _o0+1, 2 {
			switch {
			case _f.IP < 3:
				_o1 = _o0 % 2
				_f.IP = 3
				fallthrough
			case _f.IP < 4:
				if _o1 == 0 {
					coroutine.Yield[int, any](_o0 * _o0)
				}
			}
		}
	}
}

func NestedLoops(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	var _o1 int
	var _o2 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)
		_o1 = _f.Get(2).(int)
		_o2 = _f.Get(3).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_f.Set(2, _o1)
			_f.Set(3, _o2)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 1
		_f.IP = 2
		fallthrough
	case _f.IP < 5:
		for ; _o0 <= n; _o0, _f.IP = _o0+1, 2 {
			switch {
			case _f.IP < 3:
				_o1 = 1
				_f.IP = 3
				fallthrough
			case _f.IP < 5:
				for ; _o1 <= n; _o1, _f.IP = _o1+1, 3 {
					switch {
					case _f.IP < 4:
						_o2 = 1
						_f.IP = 4
						fallthrough
					case _f.IP < 5:
						for ; _o2 <= n; _o2, _f.IP = _o2+1, 4 {
							coroutine.Yield[int, any](_o0 * _o1 * _o2)
						}
					}
				}
			}
		}
	}
}

func FizzBuzzIfGenerator(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	var _o1 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)

		_o1 = _f.Get(2).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_f.Set(2, _o1)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 1
		_f.IP = 2
		fallthrough
	case _f.IP < 7:
		for ; _o0 <= n; _o0, _f.IP = _o0+1, 2 {
			if _o0%3 == 0 && _o0%5 == 0 {
				coroutine.Yield[int, any](FizzBuzz)
			} else if _o0%3 == 0 {
				coroutine.Yield[int, any](Fizz)
			} else {
				_o1 = _o0 % 5
				if _o1 == 0 {
					coroutine.Yield[int, any](Buzz)
				} else {

					coroutine.Yield[int, any](_o0)
				}
			}
		}
	}
}

func FizzBuzzSwitchGenerator(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 1
		_f.IP = 2
		fallthrough
	case _f.IP < 6:
		for ; _o0 <= n; _o0, _f.IP = _o0+1, 2 {
			switch {
			case _o0%3 == 0 && _o0%5 == 0:
				coroutine.Yield[int, any](FizzBuzz)
			case _o0%3 == 0:
				coroutine.Yield[int, any](Fizz)
			case _o0%5 == 0:
				coroutine.Yield[int, any](Buzz)
			default:

				coroutine.Yield[int, any](_o0)
			}
		}
	}
}

func Shadowing(_ int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	var _o1 int
	var _o2 int
	var _o3 int
	var _o4 int
	var _o5 int
	var _o6 int
	var _o7 int
	var _o8 int

	const _o9 = 11

	const _o10 = 12
	var _o11 int

	type _o12 uint16

	type _o13 uint32

	const _o14 = 1
	type _o15 [_o14]uint8

	type _o16 [_o14]uint8

	const _o17 = unsafe.Sizeof(_o16{}) * 2
	type _o18 [_o17]uint8
	if _f.IP > 0 {
		_o0 = _f.Get(0).(int)

		_o1 = _f.Get(1).(int)

		_o2 = _f.Get(2).(int)

		_o3 = _f.Get(3).(int)

		_o4 = _f.Get(4).(int)

		_o5 = _f.Get(5).(int)

		_o6 = _f.Get(6).(int)

		_o7 = _f.Get(7).(int)

		_o8 = _f.Get(8).(int)

		_o11 = _f.Get(9).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _o0)
			_f.Set(1, _o1)
			_f.Set(2, _o2)
			_f.Set(3, _o3)
			_f.Set(4, _o4)
			_f.Set(5, _o5)
			_f.Set(6, _o6)
			_f.Set(7, _o7)
			_f.Set(8, _o8)
			_f.Set(9, _o11)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = 0
		_f.IP = 2
		fallthrough
	case _f.IP < 3:
		coroutine.Yield[int, any](_o0)
		_f.IP = 3
		fallthrough
	case _f.IP < 5:
		switch {
		case _f.IP < 4:

			_o1 = 1
			_f.IP = 4
			fallthrough
		case _f.IP < 5:
			if true {
				coroutine.Yield[int, any](_o1)
			}
		}
		_f.IP = 5
		fallthrough
	case _f.IP < 6:

		coroutine.Yield[int, any](_o0)
		_f.IP = 6
		fallthrough
	case _f.IP < 8:
		switch {
		case _f.IP < 7:

			_o2 = 1
			_f.IP = 7
			fallthrough
		case _f.IP < 8:
			for ; _o2 < 3; _o2, _f.IP = _o2+1, 7 {
				coroutine.Yield[int, any](_o2)
			}
		}
		_f.IP = 8
		fallthrough
	case _f.IP < 9:

		coroutine.Yield[int, any](_o0)
		_f.IP = 9
		fallthrough
	case _f.IP < 13:
		switch {
		case _f.IP < 10:

			_o3 = 1
			_f.IP = 10
			fallthrough
		case _f.IP < 13:
			switch _o3 {
			case 1:
				switch {
				case _f.IP < 12:
					switch {
					case _f.IP < 11:
						_o4 = 2
						_f.IP = 11
						fallthrough
					case _f.IP < 12:
						switch _o4 {
						default:

							coroutine.Yield[int, any](_o4)
						}
					}
					_f.IP = 12
					fallthrough
				case _f.IP < 13:

					coroutine.Yield[int, any](_o3)
				}
			}
		}
		_f.IP = 13
		fallthrough
	case _f.IP < 14:

		coroutine.Yield[int, any](_o0)
		_f.IP = 14
		fallthrough
	case _f.IP < 18:
		switch {
		case _f.IP < 15:

			_o5 = 1
			_f.IP = 15
			fallthrough
		case _f.IP < 17:
			switch {
			case _f.IP < 16:

				_o6 = 2
				_f.IP = 16
				fallthrough
			case _f.IP < 17:
				coroutine.Yield[int, any](_o6)
			}
			_f.IP = 17
			fallthrough
		case _f.IP < 18:

			coroutine.Yield[int, any](_o5)
		}
		_f.IP = 18
		fallthrough
	case _f.IP < 19:

		coroutine.Yield[int, any](_o0)
		_f.IP = 19
		fallthrough
	case _f.IP < 20:
		_o7 = _o0
		_f.IP = 20
		fallthrough
	case _f.IP < 22:
		switch {
		case _f.IP < 21:

			_o8 = 1
			_f.IP = 21
			fallthrough
		case _f.IP < 22:
			coroutine.Yield[int, any](_o8)
		}
		_f.IP = 22
		fallthrough
	case _f.IP < 23:

		coroutine.Yield[int, any](_o7)
		_f.IP = 23
		fallthrough
	case _f.IP < 26:
		switch {
		case _f.IP < 25:
			switch {
			case _f.IP < 24:

				_o11 = 13
				_f.IP = 24
				fallthrough
			case _f.IP < 25:
				coroutine.Yield[int, any](_o11)
			}
			_f.IP = 25
			fallthrough
		case _f.IP < 26:

			coroutine.Yield[int, any](_o10)
		}
		_f.IP = 26
		fallthrough
	case _f.IP < 27:

		coroutine.Yield[int, any](_o9)
		_f.IP = 27
		fallthrough
	case _f.IP < 28:

		coroutine.Yield[int, any](int(unsafe.Sizeof(_o13(0))))
		_f.IP = 28
		fallthrough
	case _f.IP < 29:

		coroutine.Yield[int, any](int(unsafe.Sizeof(_o12(0))))
		_f.IP = 29
		fallthrough
	case _f.IP < 31:
		switch {
		case _f.IP < 30:

			coroutine.Yield[int, any](int(unsafe.Sizeof(_o16{})))
			_f.IP = 30
			fallthrough
		case _f.IP < 31:

			coroutine.Yield[int, any](int(unsafe.Sizeof(_o18{})))
		}
		_f.IP = 31
		fallthrough
	case _f.IP < 32:

		coroutine.Yield[int, any](int(unsafe.Sizeof(_o15{})))
	}
}

func RangeSliceIndexGenerator(_ int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 []int
	var _o1 int
	if _f.IP > 0 {
		_o0 = _f.Get(0).([]int)
		_o1 = _f.Get(1).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _o0)
			_f.Set(1, _o1)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = []int{10, 20, 30}
		_f.IP = 2
		fallthrough
	case _f.IP < 4:
		switch {
		case _f.IP < 3:
			_o1 = 0
			_f.IP = 3
			fallthrough
		case _f.IP < 4:
			for ; _o1 < len(_o0); _o1, _f.IP = _o1+1, 3 {
				coroutine.Yield[int, any](_o1)
			}
		}
	}
}

func RangeArrayIndexValueGenerator(_ int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 [3]int
	var _o1 int
	var _o2 int
	if _f.IP > 0 {
		_o0 = _f.Get(0).([3]int)
		_o1 = _f.Get(1).(int)
		_o2 = _f.Get(2).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _o0)
			_f.Set(1, _o1)
			_f.Set(2, _o2)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = [...]int{10, 20, 30}
		_f.IP = 2
		fallthrough
	case _f.IP < 6:
		switch {
		case _f.IP < 3:
			_o1 = 0
			_f.IP = 3
			fallthrough
		case _f.IP < 6:
			for ; _o1 < len(_o0); _o1, _f.IP = _o1+1, 3 {
				switch {
				case _f.IP < 4:
					_o2 = _o0[_o1]
					_f.IP = 4
					fallthrough
				case _f.IP < 5:
					coroutine.Yield[int, any](_o1)
					_f.IP = 5
					fallthrough
				case _f.IP < 6:
					coroutine.Yield[int, any](_o2)
				}
			}
		}
	}
}

func TypeSwitchingGenerator(_ int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 []any
	var _o1 int
	var _o2 any
	if _f.IP > 0 {
		_o0 = _f.Get(0).([]any)
		_o1 = _f.Get(1).(int)
		_o2 = _f.Get(2).(any)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _o0)
			_f.Set(1, _o1)
			_f.Set(2, _o2)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = []any{int8(10), int16(20), int32(30), int64(40)}
		_f.IP = 2
		fallthrough
	case _f.IP < 12:
		switch {
		case _f.IP < 3:
			_o1 = 0
			_f.IP = 3
			fallthrough
		case _f.IP < 12:
			for ; _o1 < len(_o0); _o1, _f.IP = _o1+1, 3 {
				switch {
				case _f.IP < 4:
					_o2 = _o0[_o1]
					_f.IP = 4
					fallthrough
				case _f.IP < 8:
					switch _o2.(type) {
					case int8:
						coroutine.Yield[int, any](1)
					case int16:
						coroutine.Yield[int, any](2)
					case int32:
						coroutine.Yield[int, any](4)
					case int64:
						coroutine.Yield[int, any](8)
					}
					_f.IP = 8
					fallthrough
				case _f.IP < 12:
					switch v := _o2.(type) {
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
		}
	}
}

func LoopBreakAndContinue(_ int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	var _o1 int
	var _o2 int
	var _o3 int
	if _f.IP > 0 {
		_o0 = _f.Get(0).(int)
		_o1 = _f.Get(1).(int)

		_o2 = _f.Get(2).(int)
		_o3 = _f.Get(3).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, _o0)
			_f.Set(1, _o1)
			_f.Set(2, _o2)
			_f.Set(3, _o3)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 6:
		switch {
		case _f.IP < 2:
			_o0 = 0
			_f.IP = 2
			fallthrough
		case _f.IP < 6:
		_l0:
			for ; _o0 < 10; _o0, _f.IP = _o0+1, 2 {
				switch {
				case _f.IP < 4:
					switch {
					case _f.IP < 3:
						_o1 = _o0 % 2
						_f.IP = 3
						fallthrough
					case _f.IP < 4:
						if _o1 == 0 {
							continue _l0
						}
					}
					_f.IP = 4
					fallthrough
				case _f.IP < 5:
					if _o0 > 5 {
						break _l0
					}
					_f.IP = 5
					fallthrough
				case _f.IP < 6:

					coroutine.Yield[int, any](_o0)
				}
			}
		}
		_f.IP = 6
		fallthrough
	case _f.IP < 12:
		switch {
		case _f.IP < 7:

			_o2 = 0
			_f.IP = 7
			fallthrough
		case _f.IP < 12:
		_l1:
			for ; _o2 < 2; _o2, _f.IP = _o2+1, 7 {
				switch {
				case _f.IP < 8:
					_o3 = 0
					_f.IP = 8
					fallthrough
				case _f.IP < 12:
				_l2:
					for ; _o3 < 3; _o3, _f.IP = _o3+1, 8 {
						switch {
						case _f.IP < 9:
							coroutine.Yield[int, any](_o3)
							_f.IP = 9
							fallthrough
						case _f.IP < 12:
							switch _o3 {
							case 0:
								continue _l2
							case 1:
								switch _o2 {
								case 0:
									continue _l1
								case 1:
									break _l1
								}
							}
						}
					}
				}
			}
		}
	}
}

func RangeOverMaps(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 map[int]int
	var _o1 map[int]int
	var _o2 int
	var _o3 map[int]int
	var _o4 []int
	var _o5 []int
	var _o6 int
	var _o7 int
	var _o8 bool
	var _o9 map[int]int
	var _o10 []int
	var _o11 []int
	var _o12 int
	var _o13 int
	var _o14 bool
	var _o15 map[int]int
	var _o16 int
	var _o17 map[int]int
	var _o18 []int
	var _o19 []int
	var _o20 int
	var _o21 int
	var _o22 bool
	var _o23 map[int]int
	var _o24 []int
	var _o25 []int
	var _o26 int
	var _o27 int
	var _o28 int
	var _o29 bool
	var _o30 map[int]struct {
	}
	var _o31 int
	var _o32 map[int]struct {
	}
	var _o33 []int
	var _o34 []int
	var _o35 int
	var _o36 int
	var _o37 bool
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(map[int]int)
		_o1 = _f.Get(2).(map[int]int)
		_o2 = _f.Get(3).(int)
		_o3 = _f.Get(4).(map[int]int)
		_o4 = _f.Get(5).([]int)
		_o5 = _f.Get(6).([]int)
		_o6 = _f.Get(7).(int)
		_o7 = _f.Get(8).(int)
		_o8 = _f.Get(9).(bool)
		_o9 = _f.Get(10).(map[int]int)
		_o10 = _f.Get(11).([]int)
		_o11 = _f.Get(12).([]int)
		_o12 = _f.Get(13).(int)
		_o13 = _f.Get(14).(int)
		_o14 = _f.Get(15).(bool)
		_o15 = _f.Get(16).(map[int]int)
		_o16 = _f.Get(17).(int)
		_o17 = _f.Get(18).(map[int]int)
		_o18 = _f.Get(19).([]int)
		_o19 = _f.Get(20).([]int)
		_o20 = _f.Get(21).(int)

		_o21 = _f.Get(22).(int)
		_o22 = _f.Get(23).(bool)
		_o23 = _f.Get(24).(map[int]int)
		_o24 = _f.Get(25).([]int)
		_o25 = _f.Get(26).([]int)
		_o26 = _f.Get(27).(int)

		_o27 = _f.Get(28).(int)
		_o28 = _f.Get(29).(int)
		_o29 = _f.Get(30).(bool)

		_o30 = _f.Get(31).(map[int]struct {
		})
		_o31 = _f.Get(32).(int)
		_o32 = _f.Get(33).(map[int]struct {
		})
		_o33 = _f.Get(34).([]int)
		_o34 = _f.Get(35).([]int)
		_o35 = _f.Get(36).(int)

		_o36 = _f.Get(37).(int)
		_o37 = _f.Get(38).(bool)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_f.Set(2, _o1)
			_f.Set(3, _o2)
			_f.Set(4, _o3)
			_f.Set(5, _o4)
			_f.Set(6, _o5)
			_f.Set(7, _o6)
			_f.Set(8, _o7)
			_f.Set(9, _o8)
			_f.Set(10, _o9)
			_f.Set(11, _o10)
			_f.Set(12, _o11)
			_f.Set(13, _o12)
			_f.Set(14, _o13)
			_f.Set(15, _o14)
			_f.Set(16, _o15)
			_f.Set(17, _o16)
			_f.Set(18, _o17)
			_f.Set(19, _o18)
			_f.Set(20, _o19)
			_f.Set(21, _o20)
			_f.Set(22, _o21)
			_f.Set(23, _o22)
			_f.Set(24, _o23)
			_f.Set(25, _o24)
			_f.Set(26, _o25)
			_f.Set(27, _o26)
			_f.Set(28, _o27)
			_f.Set(29, _o28)
			_f.Set(30, _o29)
			_f.Set(31, _o30)
			_f.Set(32, _o31)
			_f.Set(33, _o32)
			_f.Set(34, _o33)
			_f.Set(35, _o34)
			_f.Set(36, _o35)
			_f.Set(37, _o36)
			_f.Set(38, _o37)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 2:
		_o0 = map[int]int{}
		_f.IP = 2
		fallthrough
	case _f.IP < 5:
		switch {
		case _f.IP < 3:
			_o1 = _o0
			_f.IP = 3
			fallthrough
		case _f.IP < 5:
			switch {
			case _f.IP < 4:
				_o2 = 0
				_f.IP = 4
				fallthrough
			case _f.IP < 5:
				for ; _o2 < len(_o1); _o2, _f.IP = _o2+1, 4 {
					panic("unreachable")
				}
			}
		}
		_f.IP = 5
		fallthrough
	case _f.IP < 13:
		switch {
		case _f.IP < 6:
			_o3 = _o0
			_f.IP = 6
			fallthrough
		case _f.IP < 8:
			switch {
			case _f.IP < 7:
				_o4 = make([]int, 0, len(_o3))
				_f.IP = 7
				fallthrough
			case _f.IP < 8:
				for _v4 := range _o3 {
					_o4 = append(_o4, _v4)
				}
			}
			_f.IP = 8
			fallthrough
		case _f.IP < 13:
			switch {
			case _f.IP < 9:
				_o5 = _o4
				_f.IP = 9
				fallthrough
			case _f.IP < 13:
				switch {
				case _f.IP < 10:
					_o6 = 0
					_f.IP = 10
					fallthrough
				case _f.IP < 13:
					for ; _o6 < len(_o5); _o6, _f.IP = _o6+1, 10 {
						switch {
						case _f.IP < 11:
							_o7 = _o5[_o6]
							_f.IP = 11
							fallthrough
						case _f.IP < 13:
							switch {
							case _f.IP < 12:
								_, _o8 = _o3[_o7]
								_f.IP = 12
								fallthrough
							case _f.IP < 13:
								if _o8 {
									panic("unreachable")
								}
							}
						}
					}
				}
			}
		}
		_f.IP = 13
		fallthrough
	case _f.IP < 21:
		switch {
		case _f.IP < 14:
			_o9 = _o0
			_f.IP = 14
			fallthrough
		case _f.IP < 16:
			switch {
			case _f.IP < 15:
				_o10 = make([]int, 0, len(_o9))
				_f.IP = 15
				fallthrough
			case _f.IP < 16:
				for _v11 := range _o9 {
					_o10 = append(_o10, _v11)
				}
			}
			_f.IP = 16
			fallthrough
		case _f.IP < 21:
			switch {
			case _f.IP < 17:
				_o11 = _o10
				_f.IP = 17
				fallthrough
			case _f.IP < 21:
				switch {
				case _f.IP < 18:
					_o12 = 0
					_f.IP = 18
					fallthrough
				case _f.IP < 21:
					for ; _o12 < len(_o11); _o12, _f.IP = _o12+1, 18 {
						switch {
						case _f.IP < 19:
							_o13 = _o11[_o12]
							_f.IP = 19
							fallthrough
						case _f.IP < 21:
							switch {
							case _f.IP < 20:
								_, _o14 = _o9[_o13]
								_f.IP = 20
								fallthrough
							case _f.IP < 21:
								if _o14 {
									panic("unreachable")
								}
							}
						}
					}
				}
			}
		}
		_f.IP = 21
		fallthrough
	case _f.IP < 22:

		_o0[n] = n * 10
		_f.IP = 22
		fallthrough
	case _f.IP < 25:
		switch {
		case _f.IP < 23:
			_o15 = _o0
			_f.IP = 23
			fallthrough
		case _f.IP < 25:
			switch {
			case _f.IP < 24:
				_o16 = 0
				_f.IP = 24
				fallthrough
			case _f.IP < 25:
				for ; _o16 < len(_o15); _o16, _f.IP = _o16+1, 24 {
					coroutine.Yield[int, any](0)
				}
			}
		}
		_f.IP = 25
		fallthrough
	case _f.IP < 33:
		switch {
		case _f.IP < 26:
			_o17 = _o0
			_f.IP = 26
			fallthrough
		case _f.IP < 28:
			switch {
			case _f.IP < 27:
				_o18 = make([]int, 0, len(_o17))
				_f.IP = 27
				fallthrough
			case _f.IP < 28:
				for _v20 := range _o17 {
					_o18 = append(_o18, _v20)
				}
			}
			_f.IP = 28
			fallthrough
		case _f.IP < 33:
			switch {
			case _f.IP < 29:
				_o19 = _o18
				_f.IP = 29
				fallthrough
			case _f.IP < 33:
				switch {
				case _f.IP < 30:
					_o20 = 0
					_f.IP = 30
					fallthrough
				case _f.IP < 33:
					for ; _o20 < len(_o19); _o20, _f.IP = _o20+1, 30 {
						switch {
						case _f.IP < 31:
							_o21 = _o19[_o20]
							_f.IP = 31
							fallthrough
						case _f.IP < 33:
							switch {
							case _f.IP < 32:
								_, _o22 = _o17[_o21]
								_f.IP = 32
								fallthrough
							case _f.IP < 33:
								if _o22 {
									coroutine.Yield[int, any](_o21)
								}
							}
						}
					}
				}
			}
		}
		_f.IP = 33
		fallthrough
	case _f.IP < 42:
		switch {
		case _f.IP < 34:
			_o23 = _o0
			_f.IP = 34
			fallthrough
		case _f.IP < 36:
			switch {
			case _f.IP < 35:
				_o24 = make([]int, 0, len(_o23))
				_f.IP = 35
				fallthrough
			case _f.IP < 36:
				for _v26 := range _o23 {
					_o24 = append(_o24, _v26)
				}
			}
			_f.IP = 36
			fallthrough
		case _f.IP < 42:
			switch {
			case _f.IP < 37:
				_o25 = _o24
				_f.IP = 37
				fallthrough
			case _f.IP < 42:
				switch {
				case _f.IP < 38:
					_o26 = 0
					_f.IP = 38
					fallthrough
				case _f.IP < 42:
					for ; _o26 < len(_o25); _o26, _f.IP = _o26+1, 38 {
						switch {
						case _f.IP < 39:
							_o27 = _o25[_o26]
							_f.IP = 39
							fallthrough
						case _f.IP < 42:
							switch {
							case _f.IP < 40:
								_o28, _o29 = _o23[_o27]
								_f.IP = 40
								fallthrough
							case _f.IP < 42:
								if _o29 {
									switch {
									case _f.IP < 41:
										coroutine.Yield[int, any](_o27)
										_f.IP = 41
										fallthrough
									case _f.IP < 42:
										coroutine.Yield[int, any](_o28)
									}
								}
							}
						}
					}
				}
			}
		}
		_f.IP = 42
		fallthrough
	case _f.IP < 43:

		_o30 = make(map[int]struct{}, n)
		_f.IP = 43
		fallthrough
	case _f.IP < 45:
		switch {
		case _f.IP < 44:
			_o31 = 0
			_f.IP = 44
			fallthrough
		case _f.IP < 45:
			for ; _o31 < n; _o31, _f.IP = _o31+1, 44 {
				_o30[_o31] = struct{}{}
			}
		}
		_f.IP = 45
		fallthrough
	case _f.IP < 46:

		coroutine.Yield[int, any](len(_o30))
		_f.IP = 46
		fallthrough
	case _f.IP < 55:
		switch {
		case _f.IP < 47:
			_o32 = _o30
			_f.IP = 47
			fallthrough
		case _f.IP < 49:
			switch {
			case _f.IP < 48:
				_o33 = make([]int, 0, len(_o32))
				_f.IP = 48
				fallthrough
			case _f.IP < 49:
				for _v32 := range _o32 {
					_o33 = append(_o33, _v32)
				}
			}
			_f.IP = 49
			fallthrough
		case _f.IP < 55:
			switch {
			case _f.IP < 50:
				_o34 = _o33
				_f.IP = 50
				fallthrough
			case _f.IP < 55:
				switch {
				case _f.IP < 51:
					_o35 = 0
					_f.IP = 51
					fallthrough
				case _f.IP < 55:
					for ; _o35 < len(_o34); _o35, _f.IP = _o35+1, 51 {
						switch {
						case _f.IP < 52:
							_o36 = _o34[_o35]
							_f.IP = 52
							fallthrough
						case _f.IP < 55:
							switch {
							case _f.IP < 53:
								_, _o37 = _o32[_o36]
								_f.IP = 53
								fallthrough
							case _f.IP < 55:
								if _o37 {
									switch {
									case _f.IP < 54:
										delete(_o30, _o36)
										_f.IP = 54
										fallthrough
									case _f.IP < 55:
										coroutine.Yield[int, any](len(_o30))
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func Select(n int) {
	_c := coroutine.LoadContext[int, any]()
	_f, _fp := _c.Push()
	var _o0 int
	var _o1 int
	var _o2 int
	var _o3 int
	var _o4 int
	var _o5 int
	if _f.IP > 0 {
		n = _f.Get(0).(int)
		_o0 = _f.Get(1).(int)

		_o1 = _f.Get(2).(int)
		_o2 = _f.Get(3).(int)
		_o3 = _f.Get(4).(int)
		_o4 = _f.Get(5).(int)

		_o5 = _f.Get(6).(int)
	}
	defer func() {
		if _c.Unwinding() {
			_f.Set(0, n)
			_f.Set(1, _o0)
			_f.Set(2, _o1)
			_f.Set(3, _o2)
			_f.Set(4, _o3)
			_f.Set(5, _o4)
			_f.Set(6, _o5)
			_c.Store(_fp, _f)
		} else {
			_c.Pop()
		}
	}()
	switch {
	case _f.IP < 4:
		switch {
		case _f.IP < 2:
			_o0 = 0
			_f.IP = 2
			fallthrough
		case _f.IP < 3:
			select {
			default:
				_o0 = 1

			}
			_f.IP = 3
			fallthrough
		case _f.IP < 4:
			switch _o0 {
			case 1:
				coroutine.Yield[int, any](-1)
			}
		}
		_f.IP = 4
		fallthrough
	case _f.IP < 15:
		switch {
		case _f.IP < 5:

			_o1 = 0
			_f.IP = 5
			fallthrough
		case _f.IP < 15:
			for ; _o1 < n; _o1, _f.IP = _o1+1, 5 {
				switch {
				case _f.IP < 11:
					switch {
					case _f.IP < 6:
						_o2 = 0
						_f.IP = 6
						fallthrough
					case _f.IP < 8:
						select {
						case <-time.After(0):
							_o2 = 1

						case <-time.After(1 * time.Second):
							_o2 = 2

						}
						_f.IP = 8
						fallthrough
					case _f.IP < 11:
					_l2:
						switch _o2 {
						case 1:
							switch {
							case _f.IP < 9:
								if _o1 >= 5 {
									break _l2
								}
								_f.IP = 9
								fallthrough
							case _f.IP < 10:

								coroutine.Yield[int, any](_o1)
							}
						case 2:

							panic("unreachable")
						}
					}
					_f.IP = 11
					fallthrough
				case _f.IP < 15:
					switch {
					case _f.IP < 12:
						_o3 = 0
						_f.IP = 12
						fallthrough
					case _f.IP < 13:

						select {
						case <-time.After(0):
							_o3 = 1

						}
						_f.IP = 13
						fallthrough
					case _f.IP < 15:
					_l3:
						switch _o3 {
						case 1:
							switch {
							case _f.IP < 14:
								if _o1 >= 6 {
									break _l3
								}
								_f.IP = 14
								fallthrough
							case _f.IP < 15:

								coroutine.Yield[int, any](_o1 * 10)
							}
						}
					}
				}
			}
		}
		_f.IP = 15
		fallthrough
	case _f.IP < 19:
		switch {
		case _f.IP < 16:
			_o4 = 0
			_f.IP = 16
			fallthrough
		case _f.IP < 17:

			select {
			case <-time.After(0):
				_o4 = 1

			}
			_f.IP = 17
			fallthrough
		case _f.IP < 19:
			switch _o4 {
			case 1:
				switch {
				case _f.IP < 18:
					_o5 = 0
					_f.IP = 18
					fallthrough
				case _f.IP < 19:
					for ; _o5 < 3; _o5, _f.IP = _o5+1, 18 {
						coroutine.Yield[int, any](_o5)
					}
				}
			}
		}
	}
}
func init() {
	serde.RegisterType[**byte]()
	serde.RegisterType[*[100000]uintptr]()
	serde.RegisterType[*[1125899906842623]byte]()
	serde.RegisterType[*[131072]uint16]()
	serde.RegisterType[*[140737488355327]byte]()
	serde.RegisterType[*[16]byte]()
	serde.RegisterType[*[171]uint8]()
	serde.RegisterType[*[1]uintptr]()
	serde.RegisterType[*[268435456]uintptr]()
	serde.RegisterType[*[281474976710655]uint32]()
	serde.RegisterType[*[2]byte]()
	serde.RegisterType[*[2]float32]()
	serde.RegisterType[*[2]float64]()
	serde.RegisterType[*[2]int32]()
	serde.RegisterType[*[2]uint32]()
	serde.RegisterType[*[2]uintptr]()
	serde.RegisterType[*[32]rune]()
	serde.RegisterType[*[32]uintptr]()
	serde.RegisterType[*[4]byte]()
	serde.RegisterType[*[562949953421311]uint16]()
	serde.RegisterType[*[65536]uintptr]()
	serde.RegisterType[*[70368744177663]uint16]()
	serde.RegisterType[*[8]byte]()
	serde.RegisterType[*[8]uint8]()
	serde.RegisterType[*[]byte]()
	serde.RegisterType[*[]uint64]()
	serde.RegisterType[*bool]()
	serde.RegisterType[*byte]()
	serde.RegisterType[*int32]()
	serde.RegisterType[*int64]()
	serde.RegisterType[*string]()
	serde.RegisterType[*uint]()
	serde.RegisterType[*uint16]()
	serde.RegisterType[*uint32]()
	serde.RegisterType[*uint64]()
	serde.RegisterType[*uint8]()
	serde.RegisterType[*uintptr]()
	serde.RegisterType[[0]byte]()
	serde.RegisterType[[0]uint8]()
	serde.RegisterType[[0]uintptr]()
	serde.RegisterType[[1000]uintptr]()
	serde.RegisterType[[100]byte]()
	serde.RegisterType[[1024]bool]()
	serde.RegisterType[[1024]byte]()
	serde.RegisterType[[1024]uint8]()
	serde.RegisterType[[1048576]uint8]()
	serde.RegisterType[[104]byte]()
	serde.RegisterType[[108]byte]()
	serde.RegisterType[[108]int8]()
	serde.RegisterType[[10]byte]()
	serde.RegisterType[[10]string]()
	serde.RegisterType[[128]byte]()
	serde.RegisterType[[128]uint64]()
	serde.RegisterType[[128]uintptr]()
	serde.RegisterType[[129]uint8]()
	serde.RegisterType[[131072]uintptr]()
	serde.RegisterType[[133]string]()
	serde.RegisterType[[13]int32]()
	serde.RegisterType[[14]byte]()
	serde.RegisterType[[14]int8]()
	serde.RegisterType[[15]uint64]()
	serde.RegisterType[[16384]byte]()
	serde.RegisterType[[16384]uint8]()
	serde.RegisterType[[16]byte]()
	serde.RegisterType[[16]int64]()
	serde.RegisterType[[16]uint64]()
	serde.RegisterType[[17]string]()
	serde.RegisterType[[1]byte]()
	serde.RegisterType[[1]uint64]()
	serde.RegisterType[[1]uint8]()
	serde.RegisterType[[1]uintptr]()
	serde.RegisterType[[20]byte]()
	serde.RegisterType[[21]byte]()
	serde.RegisterType[[23]uint64]()
	serde.RegisterType[[249]uint8]()
	serde.RegisterType[[24]byte]()
	serde.RegisterType[[24]uint32]()
	serde.RegisterType[[252]uintptr]()
	serde.RegisterType[[253]uintptr]()
	serde.RegisterType[[256]int8]()
	serde.RegisterType[[256]uint64]()
	serde.RegisterType[[2]byte]()
	serde.RegisterType[[2]int]()
	serde.RegisterType[[2]int32]()
	serde.RegisterType[[2]uint64]()
	serde.RegisterType[[2]uintptr]()
	serde.RegisterType[[32]byte]()
	serde.RegisterType[[32]string]()
	serde.RegisterType[[32]uint8]()
	serde.RegisterType[[32]uintptr]()
	serde.RegisterType[[33]float64]()
	serde.RegisterType[[3]byte]()
	serde.RegisterType[[3]int]()
	serde.RegisterType[[3]int64]()
	serde.RegisterType[[3]uint16]()
	serde.RegisterType[[3]uint32]()
	serde.RegisterType[[3]uint64]()
	serde.RegisterType[[4096]byte]()
	serde.RegisterType[[40]byte]()
	serde.RegisterType[[44]byte]()
	serde.RegisterType[[4]byte]()
	serde.RegisterType[[4]float64]()
	serde.RegisterType[[4]int64]()
	serde.RegisterType[[4]string]()
	serde.RegisterType[[4]uint16]()
	serde.RegisterType[[4]uint32]()
	serde.RegisterType[[4]uint64]()
	serde.RegisterType[[4]uintptr]()
	serde.RegisterType[[50]uintptr]()
	serde.RegisterType[[512]byte]()
	serde.RegisterType[[512]uintptr]()
	serde.RegisterType[[5]byte]()
	serde.RegisterType[[5]uint]()
	serde.RegisterType[[61]struct {
		Size    uint32
		Mallocs uint64
		Frees   uint64
	}]()
	serde.RegisterType[[64488]byte]()
	serde.RegisterType[[64]byte]()
	serde.RegisterType[[64]uintptr]()
	serde.RegisterType[[65528]byte]()
	serde.RegisterType[[65]int8]()
	serde.RegisterType[[65]uint32]()
	serde.RegisterType[[65]uintptr]()
	serde.RegisterType[[68]struct {
		Size    uint32
		Mallocs uint64
		Frees   uint64
	}]()
	serde.RegisterType[[68]uint16]()
	serde.RegisterType[[68]uint32]()
	serde.RegisterType[[68]uint64]()
	serde.RegisterType[[68]uint8]()
	serde.RegisterType[[6]byte]()
	serde.RegisterType[[6]int]()
	serde.RegisterType[[6]int8]()
	serde.RegisterType[[6]uintptr]()
	serde.RegisterType[[8192]byte]()
	serde.RegisterType[[8]byte]()
	serde.RegisterType[[8]string]()
	serde.RegisterType[[8]uint32]()
	serde.RegisterType[[8]uint64]()
	serde.RegisterType[[8]uint8]()
	serde.RegisterType[[96]byte]()
	serde.RegisterType[[96]int8]()
	serde.RegisterType[[9]string]()
	serde.RegisterType[[9]uintptr]()
	serde.RegisterType[[]*byte]()
	serde.RegisterType[[][]int32]()
	serde.RegisterType[[]byte]()
	serde.RegisterType[[]float64]()
	serde.RegisterType[[]int]()
	serde.RegisterType[[]int16]()
	serde.RegisterType[[]int32]()
	serde.RegisterType[[]int64]()
	serde.RegisterType[[]int8]()
	serde.RegisterType[[]rune]()
	serde.RegisterType[[]string]()
	serde.RegisterType[[]uint16]()
	serde.RegisterType[[]uint32]()
	serde.RegisterType[[]uint64]()
	serde.RegisterType[[]uint8]()
	serde.RegisterType[[]uintptr]()
	serde.RegisterType[atomic.Bool]()
	serde.RegisterType[atomic.Int32]()
	serde.RegisterType[atomic.Int64]()
	serde.RegisterType[atomic.Uint32]()
	serde.RegisterType[atomic.Uint64]()
	serde.RegisterType[atomic.Uintptr]()
	serde.RegisterType[atomic.Value]()
	serde.RegisterType[bool]()
	serde.RegisterType[byte]()
	serde.RegisterType[complex128]()
	serde.RegisterType[float32]()
	serde.RegisterType[float64]()
	serde.RegisterType[int]()
	serde.RegisterType[int16]()
	serde.RegisterType[int32]()
	serde.RegisterType[int64]()
	serde.RegisterType[int8]()
	serde.RegisterType[map[*byte][]byte]()
	serde.RegisterType[map[int]int]()
	serde.RegisterType[map[int]struct{}]()
	serde.RegisterType[map[string]bool]()
	serde.RegisterType[map[string]int]()
	serde.RegisterType[map[string]uint64]()
	serde.RegisterType[rune]()
	serde.RegisterType[runtime.BlockProfileRecord]()
	serde.RegisterType[runtime.Frame]()
	serde.RegisterType[runtime.Frames]()
	serde.RegisterType[runtime.Func]()
	serde.RegisterType[runtime.MemProfileRecord]()
	serde.RegisterType[runtime.MemStats]()
	serde.RegisterType[runtime.PanicNilError]()
	serde.RegisterType[runtime.Pinner]()
	serde.RegisterType[runtime.StackRecord]()
	serde.RegisterType[runtime.TypeAssertionError]()
	serde.RegisterType[string]()
	serde.RegisterType[struct {
		b bool
		x any
	}]()
	serde.RegisterType[struct {
		base uintptr
		end  uintptr
	}]()
	serde.RegisterType[struct {
		enabled bool
		pad     [3]byte
		needed  bool
		alignme uint64
	}]()
	serde.RegisterType[struct {
		fill     uint64
		capacity uint64
	}]()
	serde.RegisterType[struct {
		tick uint64
		i    int
	}]()
	serde.RegisterType[struct{}]()
	serde.RegisterType[sync.Cond]()
	serde.RegisterType[sync.Map]()
	serde.RegisterType[sync.Mutex]()
	serde.RegisterType[sync.Once]()
	serde.RegisterType[sync.Pool]()
	serde.RegisterType[sync.RWMutex]()
	serde.RegisterType[sync.WaitGroup]()
	serde.RegisterType[syscall.Cmsghdr]()
	serde.RegisterType[syscall.Credential]()
	serde.RegisterType[syscall.Dirent]()
	serde.RegisterType[syscall.EpollEvent]()
	serde.RegisterType[syscall.Errno]()
	serde.RegisterType[syscall.FdSet]()
	serde.RegisterType[syscall.Flock_t]()
	serde.RegisterType[syscall.Fsid]()
	serde.RegisterType[syscall.ICMPv6Filter]()
	serde.RegisterType[syscall.IPMreq]()
	serde.RegisterType[syscall.IPMreqn]()
	serde.RegisterType[syscall.IPv6MTUInfo]()
	serde.RegisterType[syscall.IPv6Mreq]()
	serde.RegisterType[syscall.IfAddrmsg]()
	serde.RegisterType[syscall.IfInfomsg]()
	serde.RegisterType[syscall.Inet4Pktinfo]()
	serde.RegisterType[syscall.Inet6Pktinfo]()
	serde.RegisterType[syscall.InotifyEvent]()
	serde.RegisterType[syscall.Iovec]()
	serde.RegisterType[syscall.Linger]()
	serde.RegisterType[syscall.Msghdr]()
	serde.RegisterType[syscall.NetlinkMessage]()
	serde.RegisterType[syscall.NetlinkRouteAttr]()
	serde.RegisterType[syscall.NetlinkRouteRequest]()
	serde.RegisterType[syscall.NlAttr]()
	serde.RegisterType[syscall.NlMsgerr]()
	serde.RegisterType[syscall.NlMsghdr]()
	serde.RegisterType[syscall.ProcAttr]()
	serde.RegisterType[syscall.PtraceRegs]()
	serde.RegisterType[syscall.RawSockaddr]()
	serde.RegisterType[syscall.RawSockaddrAny]()
	serde.RegisterType[syscall.RawSockaddrInet4]()
	serde.RegisterType[syscall.RawSockaddrInet6]()
	serde.RegisterType[syscall.RawSockaddrLinklayer]()
	serde.RegisterType[syscall.RawSockaddrNetlink]()
	serde.RegisterType[syscall.RawSockaddrUnix]()
	serde.RegisterType[syscall.Rlimit]()
	serde.RegisterType[syscall.RtAttr]()
	serde.RegisterType[syscall.RtGenmsg]()
	serde.RegisterType[syscall.RtMsg]()
	serde.RegisterType[syscall.RtNexthop]()
	serde.RegisterType[syscall.Rusage]()
	serde.RegisterType[syscall.Signal]()
	serde.RegisterType[syscall.SockFilter]()
	serde.RegisterType[syscall.SockFprog]()
	serde.RegisterType[syscall.SockaddrInet4]()
	serde.RegisterType[syscall.SockaddrInet6]()
	serde.RegisterType[syscall.SockaddrLinklayer]()
	serde.RegisterType[syscall.SockaddrNetlink]()
	serde.RegisterType[syscall.SockaddrUnix]()
	serde.RegisterType[syscall.SocketControlMessage]()
	serde.RegisterType[syscall.Stat_t]()
	serde.RegisterType[syscall.Statfs_t]()
	serde.RegisterType[syscall.SysProcAttr]()
	serde.RegisterType[syscall.SysProcIDMap]()
	serde.RegisterType[syscall.Sysinfo_t]()
	serde.RegisterType[syscall.TCPInfo]()
	serde.RegisterType[syscall.Termios]()
	serde.RegisterType[syscall.Time_t]()
	serde.RegisterType[syscall.Timespec]()
	serde.RegisterType[syscall.Timeval]()
	serde.RegisterType[syscall.Timex]()
	serde.RegisterType[syscall.Tms]()
	serde.RegisterType[syscall.Ucred]()
	serde.RegisterType[syscall.Ustat_t]()
	serde.RegisterType[syscall.Utimbuf]()
	serde.RegisterType[syscall.Utsname]()
	serde.RegisterType[syscall.WaitStatus]()
	serde.RegisterType[time.Duration]()
	serde.RegisterType[time.Location]()
	serde.RegisterType[time.Month]()
	serde.RegisterType[time.ParseError]()
	serde.RegisterType[time.Ticker]()
	serde.RegisterType[time.Time]()
	serde.RegisterType[time.Timer]()
	serde.RegisterType[time.Weekday]()
	serde.RegisterType[uint]()
	serde.RegisterType[uint16]()
	serde.RegisterType[uint32]()
	serde.RegisterType[uint64]()
	serde.RegisterType[uint8]()
	serde.RegisterType[uintptr]()
	serde.RegisterType[unsafe.Pointer]()
}
