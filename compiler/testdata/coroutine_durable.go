//go:build durable

package testdata

import (
	coroutine "github.com/stealthrocket/coroutine"
	time "time"
	unsafe "unsafe"
)
import _types "github.com/stealthrocket/coroutine/types"

func SomeFunctionThatShouldExistInTheCompiledFile() {
}

//go:noinline
func Identity(n int) { coroutine.Yield[int, any](n) }

//go:noinline
func SquareGenerator(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 1
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		for ; _f0.X1 <= _f0.X0; _f0.X1, _f0.IP = _f0.X1+1, 2 {
			coroutine.Yield[int, any](_f0.X1 * _f0.X1)
		}
	}
}

//go:noinline
func SquareGeneratorTwice(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
	} = coroutine.Push[struct {
		IP int
		X0 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		SquareGenerator(_f0.X0)
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		SquareGenerator(_f0.X0)
	}
}

//go:noinline
func SquareGeneratorTwiceLoop(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 0
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		for ; _f0.X1 < 2; _f0.X1, _f0.IP = _f0.X1+1, 2 {
			SquareGenerator(_f0.X0)
		}
	}
}

//go:noinline
func EvenSquareGenerator(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
		X2 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
			X2 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 1
		_f0.IP = 2
		fallthrough
	case _f0.IP < 4:
		for ; _f0.X1 <= _f0.X0; _f0.X1, _f0.IP = _f0.X1+1, 2 {
			switch {
			case _f0.IP < 3:
				_f0.X2 = _f0.X1 % 2
				_f0.IP = 3
				fallthrough
			case _f0.IP < 4:
				if _f0.X2 == 0 {
					coroutine.Yield[int, any](_f0.X1 * _f0.X1)
				}
			}
		}
	}
}

//go:noinline
func NestedLoops(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
		X2 int
		X3 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 int
		X3 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
			X2 int
			X3 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 1
		_f0.IP = 2
		fallthrough
	case _f0.IP < 5:
		for ; _f0.X1 <= _f0.X0; _f0.X1, _f0.IP = _f0.X1+1, 2 {
			switch {
			case _f0.IP < 3:
				_f0.X2 = 1
				_f0.IP = 3
				fallthrough
			case _f0.IP < 5:
				for ; _f0.X2 <= _f0.X0; _f0.X2, _f0.IP = _f0.X2+1, 3 {
					switch {
					case _f0.IP < 4:
						_f0.X3 = 1
						_f0.IP = 4
						fallthrough
					case _f0.IP < 5:
						for ; _f0.X3 <= _f0.X0; _f0.X3, _f0.IP = _f0.X3+1, 4 {
							coroutine.Yield[int, any](_f0.X1 * _f0.X2 * _f0.X3)
						}
					}
				}
			}
		}
	}
}

//go:noinline
func FizzBuzzIfGenerator(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
		X2 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
			X2 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 1
		_f0.IP = 2
		fallthrough
	case _f0.IP < 7:
		for ; _f0.X1 <= _f0.X0; _f0.X1, _f0.IP = _f0.X1+1, 2 {
			if _f0.X1%
				3 == 0 && _f0.X1%5 == 0 {
				coroutine.Yield[int, any](FizzBuzz)
			} else {
				if _f0.X1%
					3 == 0 {
					coroutine.Yield[int, any](Fizz)
				} else {
					switch {
					case _f0.IP < 5:
						_f0.X2 = _f0.X1 % 5
						_f0.IP = 5
						fallthrough
					case _f0.IP < 7:
						if _f0.X2 == 0 {
							coroutine.Yield[int, any](Buzz)
						} else {

							coroutine.Yield[int, any](_f0.X1)
						}
					}
				}
			}
		}
	}
}

//go:noinline
func FizzBuzzSwitchGenerator(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
		X2 bool
		X3 bool
		X4 bool
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 bool
		X3 bool
		X4 bool
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
			X2 bool
			X3 bool
			X4 bool
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 1
		_f0.IP = 2
		fallthrough
	case _f0.IP < 9:
		for ; _f0.X1 <= _f0.X0; _f0.X1, _f0.IP = _f0.X1+1, 2 {
			switch {
			default:
				switch {
				case _f0.IP < 3:
					_f0.X2 = _f0.X1%
						3 == 0 && _f0.X1%5 == 0
					_f0.IP = 3
					fallthrough
				case _f0.IP < 9:
					if _f0.X2 {
						coroutine.Yield[int, any](FizzBuzz)
					} else {
						switch {
						case _f0.IP < 5:
							_f0.X3 = _f0.X1%
								3 == 0
							_f0.IP = 5
							fallthrough
						case _f0.IP < 9:
							if _f0.X3 {
								coroutine.Yield[int, any](Fizz)
							} else {
								switch {
								case _f0.IP < 7:
									_f0.X4 = _f0.X1%
										5 == 0
									_f0.IP = 7
									fallthrough
								case _f0.IP < 9:
									if _f0.X4 {
										coroutine.Yield[int, any](Buzz)
									} else {

										coroutine.Yield[int, any](_f0.X1)
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

//go:noinline
func Shadowing(_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP  int
		X0  int
		X1  int
		X2  int
		X3  int
		X4  int
		X5  bool
		X6  int
		X7  int
		X8  int
		X9  int
		X10 int
		X11 int
		X12 int
		X13 uintptr
		X14 int
		X15 uintptr
		X16 int
		X17 uintptr
		X18 int
		X19 uintptr
		X20 int
		X21 uintptr
		X22 int
	} = coroutine.Push[struct {
		IP  int
		X0  int
		X1  int
		X2  int
		X3  int
		X4  int
		X5  bool
		X6  int
		X7  int
		X8  int
		X9  int
		X10 int
		X11 int
		X12 int
		X13 uintptr
		X14 int
		X15 uintptr
		X16 int
		X17 uintptr
		X18 int
		X19 uintptr
		X20 int
		X21 uintptr
		X22 int
	}](&_c.Stack)

	const _o0 = 11

	const _o1 = 12

	type _o2 uint16

	type _o3 uint32

	const _o4 = 1
	type _o5 [_o4]uint8

	type _o6 [_o4]uint8

	const _o7 = unsafe.Sizeof(_o6{}) * 2
	type _o8 [_o7]uint8
	if _f0.IP == 0 {
		*_f0 = struct {
			IP  int
			X0  int
			X1  int
			X2  int
			X3  int
			X4  int
			X5  bool
			X6  int
			X7  int
			X8  int
			X9  int
			X10 int
			X11 int
			X12 int
			X13 uintptr
			X14 int
			X15 uintptr
			X16 int
			X17 uintptr
			X18 int
			X19 uintptr
			X20 int
			X21 uintptr
			X22 int
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X0 = 0
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		coroutine.Yield[int, any](_f0.X0)
		_f0.IP = 3
		fallthrough
	case _f0.IP < 5:
		switch {
		case _f0.IP < 4:
			_f0.X1 = 1
			_f0.IP = 4
			fallthrough
		case _f0.IP < 5:
			if true {
				coroutine.Yield[int, any](_f0.X1)
			}
		}
		_f0.IP = 5
		fallthrough
	case _f0.IP < 6:

		coroutine.Yield[int, any](_f0.X0)
		_f0.IP = 6
		fallthrough
	case _f0.IP < 8:
		switch {
		case _f0.IP < 7:
			_f0.X2 = 1
			_f0.IP = 7
			fallthrough
		case _f0.IP < 8:
			for ; _f0.X2 < 3; _f0.X2, _f0.IP = _f0.X2+1, 7 {
				coroutine.Yield[int, any](_f0.X2)
			}
		}
		_f0.IP = 8
		fallthrough
	case _f0.IP < 9:

		coroutine.Yield[int, any](_f0.X0)
		_f0.IP = 9
		fallthrough
	case _f0.IP < 16:
		switch {
		case _f0.IP < 10:
			_f0.X3 = 1
			_f0.IP = 10
			fallthrough
		case _f0.IP < 11:
			_f0.X4 = _f0.X3
			_f0.IP = 11
			fallthrough
		case _f0.IP < 16:
			switch {
			default:
				switch {
				case _f0.IP < 12:
					_f0.X5 = _f0.X4 ==
						1
					_f0.IP = 12
					fallthrough
				case _f0.IP < 16:
					if _f0.X5 {
						switch {
						case _f0.IP < 15:
							switch {
							case _f0.IP < 13:
								_f0.X6 = 2
								_f0.IP = 13
								fallthrough
							case _f0.IP < 14:
								_f0.X7 = _f0.X6
								_f0.IP = 14
								fallthrough
							case _f0.IP < 15:
								switch {
								default:

									coroutine.Yield[int, any](_f0.X6)
								}
							}
							_f0.IP = 15
							fallthrough
						case _f0.IP < 16:

							coroutine.Yield[int, any](_f0.X3)
						}
					}
				}
			}
		}
		_f0.IP = 16
		fallthrough
	case _f0.IP < 17:

		coroutine.Yield[int, any](_f0.X0)
		_f0.IP = 17
		fallthrough
	case _f0.IP < 21:
		switch {
		case _f0.IP < 18:
			_f0.X8 = 1
			_f0.IP = 18
			fallthrough
		case _f0.IP < 20:
			switch {
			case _f0.IP < 19:
				_f0.X9 = 2
				_f0.IP = 19
				fallthrough
			case _f0.IP < 20:
				coroutine.Yield[int, any](_f0.X9)
			}
			_f0.IP = 20
			fallthrough
		case _f0.IP < 21:

			coroutine.Yield[int, any](_f0.X8)
		}
		_f0.IP = 21
		fallthrough
	case _f0.IP < 22:

		coroutine.Yield[int, any](_f0.X0)
		_f0.IP = 22
		fallthrough
	case _f0.IP < 23:
		_f0.X10 = _f0.X0
		_f0.IP = 23
		fallthrough
	case _f0.IP < 25:
		switch {
		case _f0.IP < 24:
			_f0.X11 = 1
			_f0.IP = 24
			fallthrough
		case _f0.IP < 25:
			coroutine.Yield[int, any](_f0.X11)
		}
		_f0.IP = 25
		fallthrough
	case _f0.IP < 26:

		coroutine.Yield[int, any](_f0.X10)
		_f0.IP = 26
		fallthrough
	case _f0.IP < 29:
		switch {
		case _f0.IP < 28:
			switch {
			case _f0.IP < 27:
				_f0.X12 = 13
				_f0.IP = 27
				fallthrough
			case _f0.IP < 28:
				coroutine.Yield[int, any](_f0.X12)
			}
			_f0.IP = 28
			fallthrough
		case _f0.IP < 29:

			coroutine.Yield[int, any](_o1)
		}
		_f0.IP = 29
		fallthrough
	case _f0.IP < 30:

		coroutine.Yield[int, any](_o0)
		_f0.IP = 30
		fallthrough
	case _f0.IP < 33:
		switch {
		case _f0.IP < 31:
			_f0.X13 = unsafe.Sizeof(_o3(0))
			_f0.IP = 31
			fallthrough
		case _f0.IP < 32:
			_f0.X14 = int(_f0.X13)
			_f0.IP = 32
			fallthrough
		case _f0.IP < 33:
			coroutine.Yield[int, any](_f0.X14)
		}
		_f0.IP = 33
		fallthrough
	case _f0.IP < 34:
		_f0.X15 = unsafe.Sizeof(_o2(0))
		_f0.IP = 34
		fallthrough
	case _f0.IP < 35:
		_f0.X16 = int(_f0.X15)
		_f0.IP = 35
		fallthrough
	case _f0.IP < 36:
		coroutine.Yield[int, any](_f0.X16)
		_f0.IP = 36
		fallthrough
	case _f0.IP < 42:
		switch {
		case _f0.IP < 37:
			_f0.X17 = unsafe.Sizeof(_o6{})
			_f0.IP = 37
			fallthrough
		case _f0.IP < 38:
			_f0.X18 = int(_f0.X17)
			_f0.IP = 38
			fallthrough
		case _f0.IP < 39:
			coroutine.Yield[int, any](_f0.X18)
			_f0.IP = 39
			fallthrough
		case _f0.IP < 40:
			_f0.X19 = unsafe.Sizeof(_o8{})
			_f0.IP = 40
			fallthrough
		case _f0.IP < 41:
			_f0.X20 = int(_f0.X19)
			_f0.IP = 41
			fallthrough
		case _f0.IP < 42:
			coroutine.Yield[int, any](_f0.X20)
		}
		_f0.IP = 42
		fallthrough
	case _f0.IP < 43:
		_f0.X21 = unsafe.Sizeof(_o5{})
		_f0.IP = 43
		fallthrough
	case _f0.IP < 44:
		_f0.X22 = int(_f0.X21)
		_f0.IP = 44
		fallthrough
	case _f0.IP < 45:
		coroutine.Yield[int, any](_f0.X22)
	}
}

//go:noinline
func RangeSliceIndexGenerator(_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 []int
		X1 int
	} = coroutine.Push[struct {
		IP int
		X0 []int
		X1 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 []int
			X1 int
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X0 = []int{10, 20, 30}
		_f0.IP = 2
		fallthrough
	case _f0.IP < 4:
		switch {
		case _f0.IP < 3:
			_f0.X1 = 0
			_f0.IP = 3
			fallthrough
		case _f0.IP < 4:
			for ; _f0.X1 < len(_f0.X0); _f0.X1, _f0.IP = _f0.X1+1, 3 {
				coroutine.Yield[int, any](_f0.X1)
			}
		}
	}
}

//go:noinline
func RangeArrayIndexValueGenerator(_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 [3]int
		X1 int
		X2 int
	} = coroutine.Push[struct {
		IP int
		X0 [3]int
		X1 int
		X2 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 [3]int
			X1 int
			X2 int
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X0 = [...]int{10, 20, 30}
		_f0.IP = 2
		fallthrough
	case _f0.IP < 6:
		switch {
		case _f0.IP < 3:
			_f0.X1 = 0
			_f0.IP = 3
			fallthrough
		case _f0.IP < 6:
			for ; _f0.X1 < len(_f0.X0); _f0.X1, _f0.IP = _f0.X1+1, 3 {
				switch {
				case _f0.IP < 4:
					_f0.X2 = _f0.X0[_f0.X1]
					_f0.IP = 4
					fallthrough
				case _f0.IP < 5:
					coroutine.Yield[int, any](_f0.X1)
					_f0.IP = 5
					fallthrough
				case _f0.IP < 6:
					coroutine.Yield[int, any](_f0.X2)
				}
			}
		}
	}
}

//go:noinline
func TypeSwitchingGenerator(_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 []any
		X1 int
		X2 any
	} = coroutine.Push[struct {
		IP int
		X0 []any
		X1 int
		X2 any
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 []any
			X1 int
			X2 any
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X0 = []any{int8(10), int16(20), int32(30), int64(40)}
		_f0.IP = 2
		fallthrough
	case _f0.IP < 12:
		switch {
		case _f0.IP < 3:
			_f0.X1 = 0
			_f0.IP = 3
			fallthrough
		case _f0.IP < 12:
			for ; _f0.X1 < len(_f0.X0); _f0.X1, _f0.IP = _f0.X1+1, 3 {
				switch {
				case _f0.IP < 4:
					_f0.X2 = _f0.X0[_f0.X1]
					_f0.IP = 4
					fallthrough
				case _f0.IP < 8:
					switch _f0.X2.(type) {
					case int8:
						coroutine.Yield[int, any](1)
					case int16:
						coroutine.Yield[int, any](2)
					case int32:
						coroutine.Yield[int, any](4)
					case int64:
						coroutine.Yield[int, any](8)
					}
					_f0.IP = 8
					fallthrough
				case _f0.IP < 12:
					switch v := _f0.X2.(type) {
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

//go:noinline
func LoopBreakAndContinue(_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
		X2 int
		X3 int
		X4 int
		X5 bool
		X6 bool
		X7 int
		X8 bool
		X9 bool
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 int
		X3 int
		X4 int
		X5 bool
		X6 bool
		X7 int
		X8 bool
		X9 bool
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
			X2 int
			X3 int
			X4 int
			X5 bool
			X6 bool
			X7 int
			X8 bool
			X9 bool
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 6:
		switch {
		case _f0.IP < 2:
			_f0.X0 = 0
			_f0.IP = 2
			fallthrough
		case _f0.IP < 6:
		_l0:
			for ; _f0.X0 < 10; _f0.X0, _f0.IP = _f0.X0+1, 2 {
				switch {
				case _f0.IP < 4:
					{
						_f0.X1 = _f0.X0 % 2
						if _f0.X1 == 0 {
							continue _l0
						}
					}
					_f0.IP = 4
					fallthrough
				case _f0.IP < 5:
					if _f0.X0 >
						5 {
						break _l0
					}
					_f0.IP = 5
					fallthrough
				case _f0.IP < 6:

					coroutine.Yield[int, any](_f0.X0)
				}
			}
		}
		_f0.IP = 6
		fallthrough
	case _f0.IP < 18:
		switch {
		case _f0.IP < 7:
			_f0.X2 = 0
			_f0.IP = 7
			fallthrough
		case _f0.IP < 18:
		_l1:
			for ; _f0.X2 < 2; _f0.X2, _f0.IP = _f0.X2+1, 7 {
				switch {
				case _f0.IP < 8:
					_f0.X3 = 0
					_f0.IP = 8
					fallthrough
				case _f0.IP < 18:
				_l2:
					for ; _f0.X3 < 3; _f0.X3, _f0.IP = _f0.X3+1, 8 {
						switch {
						case _f0.IP < 9:
							coroutine.Yield[int, any](_f0.X3)
							_f0.IP = 9
							fallthrough
						case _f0.IP < 18:
							{
								_f0.X4 = _f0.X3
								switch {
								default:
									{
										_f0.X5 = _f0.X4 ==

											0
										if _f0.X5 {
											continue _l2
										} else {
											_f0.X6 = _f0.X4 ==

												1
											if _f0.X6 {
												{
													_f0.X7 = _f0.X2
													switch {
													default:
														{
															_f0.X8 = _f0.X7 ==

																0
															if _f0.X8 {
																continue _l1
															} else {
																_f0.X9 = _f0.X7 ==

																	1
																if _f0.X9 {
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
							}
						}
					}
				}
			}
		}
	}
}

//go:noinline
func RangeOverMaps(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP  int
		X0  int
		X1  map[int]int
		X2  map[int]int
		X3  int
		X4  map[int]int
		X5  []int
		X6  []int
		X7  int
		X8  int
		X9  bool
		X10 map[int]int
		X11 []int
		X12 []int
		X13 int
		X14 int
		X15 int
		X16 bool
		X17 map[int]struct {
		}
		X18 int
		X19 map[int]struct {
		}
		X20 []int
		X21 []int
		X22 int
		X23 int
		X24 bool
	} = coroutine.Push[struct {
		IP  int
		X0  int
		X1  map[int]int
		X2  map[int]int
		X3  int
		X4  map[int]int
		X5  []int
		X6  []int
		X7  int
		X8  int
		X9  bool
		X10 map[int]int
		X11 []int
		X12 []int
		X13 int
		X14 int
		X15 int
		X16 bool
		X17 map[int]struct {
		}
		X18 int
		X19 map[int]struct {
		}
		X20 []int
		X21 []int
		X22 int
		X23 int
		X24 bool
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP  int
			X0  int
			X1  map[int]int
			X2  map[int]int
			X3  int
			X4  map[int]int
			X5  []int
			X6  []int
			X7  int
			X8  int
			X9  bool
			X10 map[int]int
			X11 []int
			X12 []int
			X13 int
			X14 int
			X15 int
			X16 bool
			X17 map[int]struct {
			}
			X18 int
			X19 map[int]struct {
			}
			X20 []int
			X21 []int
			X22 int
			X23 int
			X24 bool
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = map[int]int{}
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		for range _f0.X1 {
			panic("unreachable")
		}
		_f0.IP = 3
		fallthrough
	case _f0.IP < 4:
		for _ = range _f0.X1 {
			panic("unreachable")
		}
		_f0.IP = 4
		fallthrough
	case _f0.IP < 5:
		for _, _ = range _f0.X1 {
			panic("unreachable")
		}
		_f0.IP = 5
		fallthrough
	case _f0.IP < 6:
		_f0.X1[_f0.X0] = _f0.X0 * 10
		_f0.IP = 6
		fallthrough
	case _f0.IP < 9:
		switch {
		case _f0.IP < 7:
			_f0.X2 = _f0.X1
			_f0.IP = 7
			fallthrough
		case _f0.IP < 9:
			switch {
			case _f0.IP < 8:
				_f0.X3 = 0
				_f0.IP = 8
				fallthrough
			case _f0.IP < 9:
				for ; _f0.X3 < len(_f0.X2); _f0.X3, _f0.IP = _f0.X3+1, 8 {

					coroutine.Yield[int, any](0)
				}
			}
		}
		_f0.IP = 9
		fallthrough
	case _f0.IP < 17:
		switch {
		case _f0.IP < 10:
			_f0.X4 = _f0.X1
			_f0.IP = 10
			fallthrough
		case _f0.IP < 12:
			{
				_f0.X5 = make([]int, 0, len(_f0.X4))
				for _v4 := range _f0.X4 {
					_f0.X5 = append(_f0.X5, _v4)
				}
			}
			_f0.IP = 12
			fallthrough
		case _f0.IP < 17:
			switch {
			case _f0.IP < 13:
				_f0.X6 = _f0.X5
				_f0.IP = 13
				fallthrough
			case _f0.IP < 17:
				switch {
				case _f0.IP < 14:
					_f0.X7 = 0
					_f0.IP = 14
					fallthrough
				case _f0.IP < 17:
					for ; _f0.X7 < len(_f0.X6); _f0.X7, _f0.IP = _f0.X7+1, 14 {
						switch {
						case _f0.IP < 15:
							_f0.X8 = _f0.X6[_f0.X7]
							_f0.IP = 15
							fallthrough
						case _f0.IP < 17:
							switch {
							case _f0.IP < 16:
								_, _f0.X9 = _f0.X4[_f0.X8]
								_f0.IP = 16
								fallthrough
							case _f0.IP < 17:
								if _f0.X9 {

									coroutine.Yield[int, any](_f0.X8)
								}
							}
						}
					}
				}
			}
		}
		_f0.IP = 17
		fallthrough
	case _f0.IP < 26:
		switch {
		case _f0.IP < 18:
			_f0.X10 = _f0.X1
			_f0.IP = 18
			fallthrough
		case _f0.IP < 20:
			{
				_f0.X11 = make([]int, 0, len(_f0.X10))
				for _v10 := range _f0.X10 {
					_f0.X11 = append(_f0.X11, _v10)
				}
			}
			_f0.IP = 20
			fallthrough
		case _f0.IP < 26:
			switch {
			case _f0.IP < 21:
				_f0.X12 = _f0.X11
				_f0.IP = 21
				fallthrough
			case _f0.IP < 26:
				switch {
				case _f0.IP < 22:
					_f0.X13 = 0
					_f0.IP = 22
					fallthrough
				case _f0.IP < 26:
					for ; _f0.X13 < len(_f0.X12); _f0.X13, _f0.IP = _f0.X13+1, 22 {
						switch {
						case _f0.IP < 23:
							_f0.X14 = _f0.X12[_f0.X13]
							_f0.IP = 23
							fallthrough
						case _f0.IP < 26:
							switch {
							case _f0.IP < 24:
								_f0.X15, _f0.X16 = _f0.X10[_f0.X14]
								_f0.IP = 24
								fallthrough
							case _f0.IP < 26:
								if _f0.X16 {
									switch {
									case _f0.IP < 25:

										coroutine.Yield[int, any](_f0.X14)
										_f0.IP = 25
										fallthrough
									case _f0.IP < 26:
										coroutine.Yield[int, any](_f0.X15)
									}
								}
							}
						}
					}
				}
			}
		}
		_f0.IP = 26
		fallthrough
	case _f0.IP < 27:
		_f0.X17 = make(map[int]struct{}, _f0.X0)
		_f0.IP = 27
		fallthrough
	case _f0.IP < 28:
		for _f0.X18 = 0; _f0.X18 < _f0.X0; _f0.X18++ {
			_f0.X17[_f0.X18] = struct{}{}
		}
		_f0.IP = 28
		fallthrough
	case _f0.IP < 29:
		coroutine.Yield[int, any](len(_f0.X17))
		_f0.IP = 29
		fallthrough
	case _f0.IP < 38:
		switch {
		case _f0.IP < 30:
			_f0.X19 = _f0.X17
			_f0.IP = 30
			fallthrough
		case _f0.IP < 32:
			{
				_f0.X20 = make([]int, 0, len(_f0.X19))
				for _v16 := range _f0.X19 {
					_f0.X20 = append(_f0.X20, _v16)
				}
			}
			_f0.IP = 32
			fallthrough
		case _f0.IP < 38:
			switch {
			case _f0.IP < 33:
				_f0.X21 = _f0.X20
				_f0.IP = 33
				fallthrough
			case _f0.IP < 38:
				switch {
				case _f0.IP < 34:
					_f0.X22 = 0
					_f0.IP = 34
					fallthrough
				case _f0.IP < 38:
					for ; _f0.X22 < len(_f0.X21); _f0.X22, _f0.IP = _f0.X22+1, 34 {
						switch {
						case _f0.IP < 35:
							_f0.X23 = _f0.X21[_f0.X22]
							_f0.IP = 35
							fallthrough
						case _f0.IP < 38:
							switch {
							case _f0.IP < 36:
								_, _f0.X24 = _f0.X19[_f0.X23]
								_f0.IP = 36
								fallthrough
							case _f0.IP < 38:
								if _f0.X24 {
									switch {
									case _f0.IP < 37:

										delete(_f0.X17, _f0.X23)
										_f0.IP = 37
										fallthrough
									case _f0.IP < 38:
										coroutine.Yield[int, any](len(_f0.X17))
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

//go:noinline
func Range(_fn0 int, _fn1 func(int)) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 func(int)
		X2 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 func(int)
		X2 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 func(int)
			X2 int
		}{X0: _fn0, X1: _fn1}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X2 = 0
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		for ; _f0.X2 < _f0.X0; _f0.X2, _f0.IP = _f0.X2+1, 2 {
			_f0.X1(_f0.X2)
		}
	}
}

//go:noinline
func Double(n int) { coroutine.Yield[int, any](2 * n) }

//go:noinline
func RangeTriple(n int) {
	Range(n, func(i int) { coroutine.Yield[int, any](3 * i) })
}

//go:noinline
func RangeTripleFuncValue(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 func(int)
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 func(int)
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 func(int)
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = func(i int) { coroutine.Yield[int, any](3 * i) }
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:

		Range(_f0.X0, _f0.X1)
	}
}

//go:noinline
func RangeReverseClosureCaptureByValue(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
		X2 func()
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 func()
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
			X2 func()
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 0
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		_f0.X2 = func() { coroutine.Yield[int, any](_f0.X0 - (_f0.X1 + 1)) }
		_f0.IP = 3
		fallthrough
	case _f0.IP < 5:
		for ; _f0.X1 < _f0.X0; _f0.IP = 3 {
			switch {
			case _f0.IP < 4:
				_f0.X2()
				_f0.IP = 4
				fallthrough
			case _f0.IP < 5:
				_f0.X1++
			}
		}
	}
}

//go:noinline
func Range10ClosureCapturingValues() {
	_c := coroutine.LoadContext[int, any]()
	var _f1 *struct {
		IP int
		X0 int
		X1 int
		X2 func() bool
		X3 bool
		X4 bool
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 func() bool
		X3 bool
		X4 bool
	}](&_c.Stack)
	if _f1.IP == 0 {
		*_f1 = struct {
			IP int
			X0 int
			X1 int
			X2 func() bool
			X3 bool
			X4 bool
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f1.IP < 2:
		_f1.X0 = 0
		_f1.IP = 2
		fallthrough
	case _f1.IP < 3:
		_f1.X1 = 10
		_f1.IP = 3
		fallthrough
	case _f1.IP < 4:
		_f1.X2 = func() (_ bool) {
			_c := coroutine.LoadContext[int, any]()
			var _f0 *struct {
				IP int
			} = coroutine.Push[struct {
				IP int
			}](&_c.Stack)
			if _f0.IP == 0 {
				*_f0 = struct {
					IP int
				}{}
			}
			defer func() {
				if !_c.Unwinding() {
					coroutine.Pop(&_c.Stack)
				}
			}()
			switch {
			case _f0.IP < 4:
				if _f1.X0 < _f1.X1 {
					switch {
					case _f0.IP < 2:
						coroutine.Yield[int, any](_f1.X0)
						_f0.IP = 2
						fallthrough
					case _f0.IP < 3:
						_f1.X0++
						_f0.IP = 3
						fallthrough
					case _f0.IP < 4:
						return true
					}
				}
				_f0.IP = 4
				fallthrough
			case _f0.IP < 5:

				return false
			}
			return
		}
		_f1.IP = 4
		fallthrough
	case _f1.IP < 7:
	_l0:
		for ; ; _f1.IP = 4 {
			switch {
			case _f1.IP < 5:
				_f1.X3 = _f1.X2()
				_f1.IP = 5
				fallthrough
			case _f1.IP < 6:
				_f1.X4 = !_f1.X3
				_f1.IP = 6
				fallthrough
			case _f1.IP < 7:
				if _f1.X4 {
					break _l0
				}
			}
		}
	}
}

//go:noinline
func Range10ClosureCapturingPointers() {
	_c := coroutine.LoadContext[int, any]()
	var _f1 *struct {
		IP int
		X0 int
		X1 int
		X2 *int
		X3 *int
		X4 func() bool
		X5 bool
		X6 bool
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
		X2 *int
		X3 *int
		X4 func() bool
		X5 bool
		X6 bool
	}](&_c.Stack)
	if _f1.IP == 0 {
		*_f1 = struct {
			IP int
			X0 int
			X1 int
			X2 *int
			X3 *int
			X4 func() bool
			X5 bool
			X6 bool
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f1.IP < 2:
		_f1.X0, _f1.X1 = 0, 10
		_f1.IP = 2
		fallthrough
	case _f1.IP < 3:
		_f1.X2 = &_f1.X0
		_f1.IP = 3
		fallthrough
	case _f1.IP < 4:
		_f1.X3 = &_f1.X1
		_f1.IP = 4
		fallthrough
	case _f1.IP < 5:
		_f1.X4 = func() (_ bool) {
			_c := coroutine.LoadContext[int, any]()
			var _f0 *struct {
				IP int
			} = coroutine.Push[struct {
				IP int
			}](&_c.Stack)
			if _f0.IP == 0 {
				*_f0 = struct {
					IP int
				}{}
			}
			defer func() {
				if !_c.Unwinding() {
					coroutine.Pop(&_c.Stack)
				}
			}()
			switch {
			case _f0.IP < 4:
				if *_f1.X2 < *_f1.X3 {
					switch {
					case _f0.IP < 2:
						coroutine.Yield[int, any](*_f1.X2)
						_f0.IP = 2
						fallthrough
					case _f0.IP < 3:
						(*_f1.X2)++
						_f0.IP = 3
						fallthrough
					case _f0.IP < 4:
						return true
					}
				}
				_f0.IP = 4
				fallthrough
			case _f0.IP < 5:

				return false
			}
			return
		}
		_f1.IP = 5
		fallthrough
	case _f1.IP < 8:
	_l0:
		for ; ; _f1.IP = 5 {
			switch {
			case _f1.IP < 6:
				_f1.X5 = _f1.X4()
				_f1.IP = 6
				fallthrough
			case _f1.IP < 7:
				_f1.X6 = !_f1.X5
				_f1.IP = 7
				fallthrough
			case _f1.IP < 8:
				if _f1.X6 {
					break _l0
				}
			}
		}
	}
}

//go:noinline
func Range10ClosureHeterogenousCapture() {
	_c := coroutine.LoadContext[int, any]()
	var _f1 *struct {
		IP  int
		X0  int8
		X1  int16
		X2  int32
		X3  int64
		X4  uint8
		X5  uint16
		X6  uint32
		X7  uint64
		X8  uintptr
		X9  func() int
		X10 int
		X11 func() bool
		X12 bool
		X13 bool
	} = coroutine.Push[struct {
		IP  int
		X0  int8
		X1  int16
		X2  int32
		X3  int64
		X4  uint8
		X5  uint16
		X6  uint32
		X7  uint64
		X8  uintptr
		X9  func() int
		X10 int
		X11 func() bool
		X12 bool
		X13 bool
	}](&_c.Stack)
	if _f1.IP == 0 {
		*_f1 = struct {
			IP  int
			X0  int8
			X1  int16
			X2  int32
			X3  int64
			X4  uint8
			X5  uint16
			X6  uint32
			X7  uint64
			X8  uintptr
			X9  func() int
			X10 int
			X11 func() bool
			X12 bool
			X13 bool
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f1.IP < 11:
		{
			_f1.X0 = 0
			_f1.X1 = 1
			_f1.X2 = 2
			_f1.X3 = 3
			_f1.X4 = 4
			_f1.X5 = 5
			_f1.X6 = 6
			_f1.X7 = 7
			_f1.X8 = 8
			_f1.X9 = func() int { return int(_f1.X8) + 1 }
		}
		_f1.IP = 11
		fallthrough
	case _f1.IP < 12:
		_f1.X10 = 0
		_f1.IP = 12
		fallthrough
	case _f1.IP < 13:
		_f1.X11 = func() (_ bool) {
			_c := coroutine.LoadContext[int, any]()
			var _f0 *struct {
				IP  int
				X0  int
				X1  int
				X2  bool
				X3  bool
				X4  bool
				X5  bool
				X6  bool
				X7  bool
				X8  bool
				X9  bool
				X10 bool
				X11 bool
			} = coroutine.Push[struct {
				IP  int
				X0  int
				X1  int
				X2  bool
				X3  bool
				X4  bool
				X5  bool
				X6  bool
				X7  bool
				X8  bool
				X9  bool
				X10 bool
				X11 bool
			}](&_c.Stack)
			if _f0.IP == 0 {
				*_f0 = struct {
					IP  int
					X0  int
					X1  int
					X2  bool
					X3  bool
					X4  bool
					X5  bool
					X6  bool
					X7  bool
					X8  bool
					X9  bool
					X10 bool
					X11 bool
				}{}
			}
			defer func() {
				if !_c.Unwinding() {
					coroutine.Pop(&_c.Stack)
				}
			}()
			switch {
			case _f0.IP < 2:
				_f0.IP = 2
				fallthrough
			case _f0.IP < 13:
				switch {
				case _f0.IP < 3:
					_f0.X1 = _f1.X10
					_f0.IP = 3
					fallthrough
				case _f0.IP < 13:
					switch {
					default:
						if _f0.X2 = _f0.X1 ==

							0; _f0.X2 {
							_f0.X0 = int(_f1.X0)
						} else if _f0.X3 = _f0.X1 ==
							1; _f0.X3 {
							_f0.X0 = int(_f1.X1)
						} else if _f0.X4 = _f0.X1 ==
							2; _f0.X4 {
							_f0.X0 = int(_f1.X2)
						} else if _f0.X5 = _f0.X1 ==
							3; _f0.X5 {
							_f0.X0 = int(_f1.X3)
						} else if _f0.X6 = _f0.X1 ==
							4; _f0.X6 {
							_f0.X0 = int(_f1.X4)
						} else if _f0.X7 = _f0.X1 ==
							5; _f0.X7 {
							_f0.X0 = int(_f1.X5)
						} else if _f0.X8 = _f0.X1 ==
							6; _f0.X8 {
							_f0.X0 = int(_f1.X6)
						} else if _f0.X9 = _f0.X1 ==
							7; _f0.X9 {
							_f0.X0 = int(_f1.X7)
						} else if _f0.X10 = _f0.X1 ==
							8; _f0.X10 {
							_f0.X0 = int(_f1.X8)
						} else if _f0.X11 = _f0.X1 ==
							9; _f0.X11 {
							_f0.X0 = _f1.X9()
						}
					}
				}
				_f0.IP = 13
				fallthrough
			case _f0.IP < 14:

				coroutine.Yield[int, any](_f0.X0)
				_f0.IP = 14
				fallthrough
			case _f0.IP < 15:
				_f1.X10++
				_f0.IP = 15
				fallthrough
			case _f0.IP < 16:
				return _f1.X10 < 10
			}
			return
		}
		_f1.IP = 13
		fallthrough
	case _f1.IP < 16:
	_l0:
		for ; ; _f1.IP = 13 {
			switch {
			case _f1.IP < 14:
				_f1.X12 = _f1.X11()
				_f1.IP = 14
				fallthrough
			case _f1.IP < 15:
				_f1.X13 = !_f1.X12
				_f1.IP = 15
				fallthrough
			case _f1.IP < 16:
				if _f1.X13 {
					break _l0
				}
			}
		}
	}
}

//go:noinline
func Range10Heterogenous() {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP  int
		X0  int8
		X1  int16
		X2  int32
		X3  int64
		X4  uint8
		X5  uint16
		X6  uint32
		X7  uint64
		X8  uintptr
		X9  int
		X10 int
	} = coroutine.Push[struct {
		IP  int
		X0  int8
		X1  int16
		X2  int32
		X3  int64
		X4  uint8
		X5  uint16
		X6  uint32
		X7  uint64
		X8  uintptr
		X9  int
		X10 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP  int
			X0  int8
			X1  int16
			X2  int32
			X3  int64
			X4  uint8
			X5  uint16
			X6  uint32
			X7  uint64
			X8  uintptr
			X9  int
			X10 int
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 10:
		{
			_f0.X0 = 0
			_f0.X1 = 1
			_f0.X2 = 2
			_f0.X3 = 3
			_f0.X4 = 4
			_f0.X5 = 5
			_f0.X6 = 6
			_f0.X7 = 7
			_f0.X8 = 8
		}
		_f0.IP = 10
		fallthrough
	case _f0.IP < 23:
		switch {
		case _f0.IP < 11:
			_f0.X9 = 0
			_f0.IP = 11
			fallthrough
		case _f0.IP < 23:
			for ; _f0.X9 < 10; _f0.X9, _f0.IP = _f0.X9+1, 11 {
				switch {
				case _f0.IP < 12:
					_f0.IP = 12
					fallthrough
				case _f0.IP < 22:

					switch _f0.X9 {
					case 0:
						_f0.X10 = int(_f0.X0)
					case 1:
						_f0.X10 = int(_f0.X1)
					case 2:
						_f0.X10 = int(_f0.X2)
					case 3:
						_f0.X10 = int(_f0.X3)
					case 4:
						_f0.X10 = int(_f0.X4)
					case 5:
						_f0.X10 = int(_f0.X5)
					case 6:
						_f0.X10 = int(_f0.X6)
					case 7:
						_f0.X10 = int(_f0.X7)
					case 8:
						_f0.X10 = int(_f0.X8)
					case 9:
						_f0.X10 = int(_f0.X9)
					}
					_f0.IP = 22
					fallthrough
				case _f0.IP < 23:
					coroutine.Yield[int, any](_f0.X10)
				}
			}
		}
	}
}

//go:noinline
func Select(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP  int
		X0  int
		X1  int
		X2  int
		X3  bool
		X4  int
		X5  int
		X6  <-chan time.Time
		X7  <-chan time.Time
		X8  int
		X9  bool
		X10 bool
		X11 int
		X12 <-chan time.Time
		X13 int
		X14 bool
		X15 int
		X16 <-chan time.Time
		X17 int
		X18 bool
		X19 int
	} = coroutine.Push[struct {
		IP  int
		X0  int
		X1  int
		X2  int
		X3  bool
		X4  int
		X5  int
		X6  <-chan time.Time
		X7  <-chan time.Time
		X8  int
		X9  bool
		X10 bool
		X11 int
		X12 <-chan time.Time
		X13 int
		X14 bool
		X15 int
		X16 <-chan time.Time
		X17 int
		X18 bool
		X19 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP  int
			X0  int
			X1  int
			X2  int
			X3  bool
			X4  int
			X5  int
			X6  <-chan time.Time
			X7  <-chan time.Time
			X8  int
			X9  bool
			X10 bool
			X11 int
			X12 <-chan time.Time
			X13 int
			X14 bool
			X15 int
			X16 <-chan time.Time
			X17 int
			X18 bool
			X19 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 6:
		switch {
		case _f0.IP < 2:
			_f0.X1 = 0
			_f0.IP = 2
			fallthrough
		case _f0.IP < 3:
			select {
			default:
				_f0.X1 = 1
			}
			_f0.IP = 3
			fallthrough
		case _f0.IP < 6:
			switch {
			case _f0.IP < 4:
				_f0.X2 = _f0.X1
				_f0.IP = 4
				fallthrough
			case _f0.IP < 6:
				switch {
				default:
					switch {
					case _f0.IP < 5:
						_f0.X3 = _f0.X2 == 1
						_f0.IP = 5
						fallthrough
					case _f0.IP < 6:
						if _f0.X3 {

							coroutine.Yield[int, any](-1)
						}
					}
				}
			}
		}
		_f0.IP = 6
		fallthrough
	case _f0.IP < 24:
		switch {
		case _f0.IP < 7:
			_f0.X4 = 0
			_f0.IP = 7
			fallthrough
		case _f0.IP < 24:
			for ; _f0.X4 < _f0.X0; _f0.X4, _f0.IP = _f0.X4+1, 7 {
				switch {
				case _f0.IP < 17:
					switch {
					case _f0.IP < 8:
						_f0.X5 = 0
						_f0.IP = 8
						fallthrough
					case _f0.IP < 9:
						_f0.X6 = time.After(0)
						_f0.IP = 9
						fallthrough
					case _f0.IP < 10:
						_f0.X7 = time.After(1 * time.Second)
						_f0.IP = 10
						fallthrough
					case _f0.IP < 12:
						select {
						case <-_f0.X6:
							_f0.X5 = 1
						case <-_f0.X7:
							_f0.X5 = 2
						}
						_f0.IP = 12
						fallthrough
					case _f0.IP < 17:
						switch {
						case _f0.IP < 13:
							_f0.X8 = _f0.X5
							_f0.IP = 13
							fallthrough
						case _f0.IP < 17:
						_l2:
							switch {
							default:
								switch {
								case _f0.IP < 14:
									_f0.X9 = _f0.X8 == 1
									_f0.IP = 14
									fallthrough
								case _f0.IP < 17:
									if _f0.X9 {
										switch {
										case _f0.IP < 15:
											if _f0.X4 >=
												5 {
												break _l2
											}
											_f0.IP = 15
											fallthrough
										case _f0.IP < 16:

											coroutine.Yield[int, any](_f0.X4)
										}
									} else if _f0.X10 = _f0.X8 == 2; _f0.X10 {

										panic("unreachable")
									}
								}
							}
						}
					}
					_f0.IP = 17
					fallthrough
				case _f0.IP < 24:
					switch {
					case _f0.IP < 18:
						_f0.X11 = 0
						_f0.IP = 18
						fallthrough
					case _f0.IP < 19:
						_f0.X12 = time.After(0)
						_f0.IP = 19
						fallthrough
					case _f0.IP < 20:
						select {
						case <-_f0.X12:
							_f0.X11 = 1
						}
						_f0.IP = 20
						fallthrough
					case _f0.IP < 24:
						switch {
						case _f0.IP < 21:
							_f0.X13 = _f0.X11
							_f0.IP = 21
							fallthrough
						case _f0.IP < 24:
						_l3:
							switch {
							default:
								switch {
								case _f0.IP < 22:
									_f0.X14 = _f0.X13 == 1
									_f0.IP = 22
									fallthrough
								case _f0.IP < 24:
									if _f0.X14 {
										switch {
										case _f0.IP < 23:
											if _f0.X4 >=
												6 {
												break _l3
											}
											_f0.IP = 23
											fallthrough
										case _f0.IP < 24:

											coroutine.Yield[int, any](_f0.X4 * 10)
										}
									}
								}
							}
						}
					}
				}
			}
		}
		_f0.IP = 24
		fallthrough
	case _f0.IP < 31:
		switch {
		case _f0.IP < 25:
			_f0.X15 = 0
			_f0.IP = 25
			fallthrough
		case _f0.IP < 26:
			_f0.X16 = time.After(0)
			_f0.IP = 26
			fallthrough
		case _f0.IP < 27:
			select {
			case <-_f0.X16:
				_f0.X15 = 1
			}
			_f0.IP = 27
			fallthrough
		case _f0.IP < 31:
			switch {
			case _f0.IP < 28:
				_f0.X17 = _f0.X15
				_f0.IP = 28
				fallthrough
			case _f0.IP < 31:
				switch {
				default:
					switch {
					case _f0.IP < 29:
						_f0.X18 = _f0.X17 == 1
						_f0.IP = 29
						fallthrough
					case _f0.IP < 31:
						if _f0.X18 {
							switch {
							case _f0.IP < 30:
								_f0.X19 = 0
								_f0.IP = 30
								fallthrough
							case _f0.IP < 31:
								for ; _f0.X19 < 3; _f0.X19, _f0.IP = _f0.X19+1, 30 {
									coroutine.Yield[int, any](_f0.X19)
								}
							}
						}
					}
				}
			}
		}
	}
}

//go:noinline
func YieldingExpressionDesugaring() {
	_c := coroutine.LoadContext[int, any]()
	var _f1 *struct {
		IP  int
		X0  int
		X1  int
		X2  int
		X3  int
		X4  int
		X5  int
		X6  int
		X7  int
		X8  int
		X9  bool
		X10 int
		X11 int
		X12 int
		X13 int
		X14 int
		X15 bool
		X16 int
		X17 int
		X18 int
		X19 int
		X20 int
		X21 bool
		X22 int
		X23 int
		X24 int
		X25 int
		X26 bool
		X27 bool
		X28 int
		X29 int
		X30 int
		X31 int
		X32 int
		X33 bool
		X34 int
		X35 int
		X36 bool
		X37 int
		X38 int
		X39 int
		X40 bool
		X41 int
		X42 int
		X43 int
		X44 bool
		X45 int
		X46 int
		X47 any
		X48 []func()
	} = coroutine.Push[struct {
		IP  int
		X0  int
		X1  int
		X2  int
		X3  int
		X4  int
		X5  int
		X6  int
		X7  int
		X8  int
		X9  bool
		X10 int
		X11 int
		X12 int
		X13 int
		X14 int
		X15 bool
		X16 int
		X17 int
		X18 int
		X19 int
		X20 int
		X21 bool
		X22 int
		X23 int
		X24 int
		X25 int
		X26 bool
		X27 bool
		X28 int
		X29 int
		X30 int
		X31 int
		X32 int
		X33 bool
		X34 int
		X35 int
		X36 bool
		X37 int
		X38 int
		X39 int
		X40 bool
		X41 int
		X42 int
		X43 int
		X44 bool
		X45 int
		X46 int
		X47 any
		X48 []func()
	}](&_c.Stack)
	if _f1.IP == 0 {
		*_f1 = struct {
			IP  int
			X0  int
			X1  int
			X2  int
			X3  int
			X4  int
			X5  int
			X6  int
			X7  int
			X8  int
			X9  bool
			X10 int
			X11 int
			X12 int
			X13 int
			X14 int
			X15 bool
			X16 int
			X17 int
			X18 int
			X19 int
			X20 int
			X21 bool
			X22 int
			X23 int
			X24 int
			X25 int
			X26 bool
			X27 bool
			X28 int
			X29 int
			X30 int
			X31 int
			X32 int
			X33 bool
			X34 int
			X35 int
			X36 bool
			X37 int
			X38 int
			X39 int
			X40 bool
			X41 int
			X42 int
			X43 int
			X44 bool
			X45 int
			X46 int
			X47 any
			X48 []func()
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			defer coroutine.Pop(&_c.Stack)
			for _, f := range _f1.X48 {
				defer f()
			}
		}
	}()
	switch {
	case _f1.IP < 2:
		_f1.X0 = 111
		_f1.IP = 2
		fallthrough
	case _f1.IP < 4:
		switch {
		case _f1.IP < 3:
			_f1.X1 = _f1.X0
			_f1.IP = 3
			fallthrough
		case _f1.IP < 4:
			_f1.X48 = append(_f1.X48, func() {
				func(_fn0 int) {
					_c := coroutine.LoadContext[int, any]()
					var _f0 *struct {
						IP int
						X0 int
						X1 int
						X2 int
					} = coroutine.Push[struct {
						IP int
						X0 int
						X1 int
						X2 int
					}](&_c.Stack)
					if _f0.IP == 0 {
						*_f0 = struct {
							IP int
							X0 int
							X1 int
							X2 int
						}{X0: _fn0}
					}
					defer func() {
						if !_c.Unwinding() {
							coroutine.Pop(&_c.Stack)
						}
					}()
					switch {
					case _f0.IP < 2:
						_f0.X1 = b(_f0.X0)
						_f0.IP = 2
						fallthrough
					case _f0.IP < 3:
						_f0.X2 = a(_f0.X1)
						_f0.IP = 3
						fallthrough
					case _f0.IP < 4:
						coroutine.Yield[int, any](_f0.X2)
					}
				}(_f1.X1)
			})
		}
		_f1.IP = 4
		fallthrough
	case _f1.IP < 5:
		_f1.X0 = 222
		_f1.IP = 5
		fallthrough
	case _f1.IP < 8:
		switch {
		case _f1.IP < 6:
			_f1.X2 = b(_f1.X0)
			_f1.IP = 6
			fallthrough
		case _f1.IP < 7:
			_f1.X3 = a(_f1.X2)
			_f1.IP = 7
			fallthrough
		case _f1.IP < 8:
			_f1.X48 = append(_f1.X48, func() {
				coroutine.Yield[int, any](_f1.X3)
			})
		}
		_f1.IP = 8
		fallthrough
	case _f1.IP < 9:
		_f1.X0 = 333
		_f1.IP = 9
		fallthrough
	case _f1.IP < 11:
		switch {
		case _f1.IP < 10:
			_f1.X4 = _f1.X0
			_f1.IP = 10
			fallthrough
		case _f1.IP < 11:
			_f1.X48 = append(_f1.X48, func() {
				coroutine.Yield[int, any](_f1.X4)
			})
		}
		_f1.IP = 11
		fallthrough
	case _f1.IP < 12:
		_f1.X0 = 444
		_f1.IP = 12
		fallthrough
	case _f1.IP < 32:
		switch {
		case _f1.IP < 13:
			_f1.X5 = b(1)
			_f1.IP = 13
			fallthrough
		case _f1.IP < 14:
			_f1.X6 = a(_f1.X5)
			_f1.IP = 14
			fallthrough
		case _f1.IP < 15:
			_f1.X7 = b(2)
			_f1.IP = 15
			fallthrough
		case _f1.IP < 16:
			_f1.X8 = a(_f1.X7)
			_f1.IP = 16
			fallthrough
		case _f1.IP < 17:
			_f1.X9 = _f1.X6 == _f1.X8
			_f1.IP = 17
			fallthrough
		case _f1.IP < 32:
			if _f1.X9 {
			} else {
				switch {
				case _f1.IP < 19:
					_f1.X10 = b(3)
					_f1.IP = 19
					fallthrough
				case _f1.IP < 20:
					_f1.X11 = a(_f1.X10)
					_f1.IP = 20
					fallthrough
				case _f1.IP < 21:
					_f1.X12 = b(4)
					_f1.IP = 21
					fallthrough
				case _f1.IP < 22:
					_f1.X13 = a(_f1.X12)
					_f1.IP = 22
					fallthrough
				case _f1.IP < 23:
					_f1.X14 = _f1.X13 - 1
					_f1.IP = 23
					fallthrough
				case _f1.IP < 24:
					_f1.X15 = _f1.X11 == _f1.X14
					_f1.IP = 24
					fallthrough
				case _f1.IP < 32:
					if _f1.X15 {
						switch {
						case _f1.IP < 25:
							_f1.X16 = b(5)
							_f1.IP = 25
							fallthrough
						case _f1.IP < 26:
							_f1.X17 = a(_f1.X16)
							_f1.IP = 26
							fallthrough
						case _f1.IP < 27:
							_f1.X18 = _f1.X17 * 10
							_f1.IP = 27
							fallthrough
						case _f1.IP < 28:
							coroutine.Yield[int, any](_f1.X18)
						}
					} else {
						switch {
						case _f1.IP < 29:
							_f1.X19 = b(100)
							_f1.IP = 29
							fallthrough
						case _f1.IP < 30:
							_f1.X20 = a(_f1.X19)
							_f1.IP = 30
							fallthrough
						case _f1.IP < 31:
							_f1.X21 = _f1.X20 == 100
							_f1.IP = 31
							fallthrough
						case _f1.IP < 32:
							if _f1.X21 {
								panic("unreachable")
							}
						}
					}
				}
			}
		}
		_f1.IP = 32
		fallthrough
	case _f1.IP < 40:
		switch {
		case _f1.IP < 33:
			_f1.X22 = b(6)
			_f1.IP = 33
			fallthrough
		case _f1.IP < 34:
			_f1.X23 = a(_f1.X22)
			_f1.IP = 34
			fallthrough
		case _f1.IP < 40:
		_l0:
			for ; ; _f1.X23, _f1.IP = _f1.X23+1, 34 {
				switch {
				case _f1.IP < 39:
					switch {
					case _f1.IP < 35:
						_f1.X24 = b(8)
						_f1.IP = 35
						fallthrough
					case _f1.IP < 36:
						_f1.X25 = a(_f1.X24)
						_f1.IP = 36
						fallthrough
					case _f1.IP < 37:
						_f1.X26 = _f1.X23 < _f1.X25
						_f1.IP = 37
						fallthrough
					case _f1.IP < 38:
						_f1.X27 = !_f1.X26
						_f1.IP = 38
						fallthrough
					case _f1.IP < 39:
						if _f1.X27 {
							break _l0
						}
					}
					_f1.IP = 39
					fallthrough
				case _f1.IP < 40:
					coroutine.Yield[int, any](70)
				}
			}
		}
		_f1.IP = 40
		fallthrough
	case _f1.IP < 62:
		switch {
		case _f1.IP < 41:
			_f1.X28 = b(9)
			_f1.IP = 41
			fallthrough
		case _f1.IP < 42:
			_f1.X29 = a(_f1.X28)
			_f1.IP = 42
			fallthrough
		case _f1.IP < 43:
			_f1.X30 = _f1.X29
			_f1.IP = 43
			fallthrough
		case _f1.IP < 62:
			switch {
			default:
				switch {
				case _f1.IP < 44:
					_f1.X31 = b(10)
					_f1.IP = 44
					fallthrough
				case _f1.IP < 45:
					_f1.X32 = a(_f1.X31)
					_f1.IP = 45
					fallthrough
				case _f1.IP < 46:
					_f1.X33 = _f1.X30 == _f1.X32
					_f1.IP = 46
					fallthrough
				case _f1.IP < 62:
					if _f1.X33 {
						panic("unreachable")
					} else {
						switch {
						case _f1.IP < 48:
							_f1.X34 = b(11)
							_f1.IP = 48
							fallthrough
						case _f1.IP < 49:
							_f1.X35 = a(_f1.X34)
							_f1.IP = 49
							fallthrough
						case _f1.IP < 50:
							_f1.X36 = _f1.X30 == _f1.X35
							_f1.IP = 50
							fallthrough
						case _f1.IP < 62:
							if _f1.X36 {
								panic("unreachable")
							} else {
								switch {
								case _f1.IP < 52:
									_f1.X37 = b(12)
									_f1.IP = 52
									fallthrough
								case _f1.IP < 53:
									_f1.X38 = a(_f1.X37)
									_f1.IP = 53
									fallthrough
								case _f1.IP < 54:
									_f1.X39 = _f1.X38 - 3
									_f1.IP = 54
									fallthrough
								case _f1.IP < 55:
									_f1.X40 = _f1.X30 == _f1.X39
									_f1.IP = 55
									fallthrough
								case _f1.IP < 62:
									if _f1.X40 {
										switch {
										case _f1.IP < 56:
											_f1.X41 = b(13)
											_f1.IP = 56
											fallthrough
										case _f1.IP < 57:
											a(_f1.X41)
										}
									} else {
										switch {
										case _f1.IP < 58:
											_f1.X42 = b(14)
											_f1.IP = 58
											fallthrough
										case _f1.IP < 59:
											_f1.X43 = a(_f1.X42)
											_f1.IP = 59
											fallthrough
										case _f1.IP < 60:
											_f1.X44 = _f1.X30 == _f1.X43
											_f1.IP = 60
											fallthrough
										case _f1.IP < 62:
											if _f1.X44 {
												panic("unreachable")
											} else {
												panic("unreachable")
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
		_f1.IP = 62
		fallthrough
	case _f1.IP < 68:
		switch {
		case _f1.IP < 63:
			_f1.X45 = b(15)
			_f1.IP = 63
			fallthrough
		case _f1.IP < 64:
			_f1.X46 = a(_f1.X45)
			_f1.IP = 64
			fallthrough
		case _f1.IP < 65:
			_f1.X47 = any(_f1.X46)
			_f1.IP = 65
			fallthrough
		case _f1.IP < 68:
			switch x := _f1.X47.(type) {
			case bool:
				panic("unreachable")
			case int:
				coroutine.Yield[int, any](x * 10)
			default:
				panic("unreachable")
			}
		}
	}
}

//go:noinline
func a(_fn0 int) (_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
	} = coroutine.Push[struct {
		IP int
		X0 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		coroutine.Yield[int, any](_f0.X0)
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		return _f0.X0
	}
	return
}

//go:noinline
func b(_fn0 int) (_ int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
	} = coroutine.Push[struct {
		IP int
		X0 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		coroutine.Yield[int, any](-_f0.X0)
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		return _f0.X0
	}
	return
}

//go:noinline
func YieldingDurations() {
	_c := coroutine.LoadContext[int, any]()
	var _f1 *struct {
		IP int
		X0 *time.Duration
		X1 time.Duration
		X2 func()
		X3 int
	} = coroutine.Push[struct {
		IP int
		X0 *time.Duration
		X1 time.Duration
		X2 func()
		X3 int
	}](&_c.Stack)
	if _f1.IP == 0 {
		*_f1 = struct {
			IP int
			X0 *time.Duration
			X1 time.Duration
			X2 func()
			X3 int
		}{}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f1.IP < 2:
		_f1.X0 = new(time.Duration)
		_f1.IP = 2
		fallthrough
	case _f1.IP < 3:
		_f1.X1 = time.Duration(100)
		_f1.IP = 3
		fallthrough
	case _f1.IP < 4:
		*_f1.X0 = _f1.X1
		_f1.IP = 4
		fallthrough
	case _f1.IP < 5:
		_f1.X2 = func() {
			_c := coroutine.LoadContext[int, any]()
			var _f0 *struct {
				IP int
				X0 int64
				X1 int
				X2 time.Duration
			} = coroutine.Push[struct {
				IP int
				X0 int64
				X1 int
				X2 time.Duration
			}](&_c.Stack)
			if _f0.IP == 0 {
				*_f0 = struct {
					IP int
					X0 int64
					X1 int
					X2 time.Duration
				}{}
			}
			defer func() {
				if !_c.Unwinding() {
					coroutine.Pop(&_c.Stack)
				}
			}()
			switch {
			case _f0.IP < 2:
				_f0.X0 = _f1.X0.
					Nanoseconds()
				_f0.IP = 2
				fallthrough
			case _f0.IP < 3:
				_f0.X1 = int(_f0.X0)
				_f0.IP = 3
				fallthrough
			case _f0.IP < 4:
				_f0.X2 = time.Duration(_f0.X1 + 1)
				_f0.IP = 4
				fallthrough
			case _f0.IP < 5:
				*_f1.X0 = _f0.X2
				_f0.IP = 5
				fallthrough
			case _f0.IP < 6:
				coroutine.Yield[int, any](_f0.X1)
			}
		}
		_f1.IP = 5
		fallthrough
	case _f1.IP < 7:
		switch {
		case _f1.IP < 6:
			_f1.X3 = 0
			_f1.IP = 6
			fallthrough
		case _f1.IP < 7:
			for ; _f1.X3 < 10; _f1.X3, _f1.IP = _f1.X3+1, 6 {
				_f1.X2()
			}
		}
	}
}

//go:noinline
func YieldAndDeferAssign(_fn0 *int, _fn1, _fn2 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 *int
		X1 int
		X2 int
		X3 []func()
	} = coroutine.Push[struct {
		IP int
		X0 *int
		X1 int
		X2 int
		X3 []func()
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 *int
			X1 int
			X2 int
			X3 []func()
		}{X0: _fn0, X1: _fn1, X2: _fn2}
	}
	defer func() {
		if !_c.Unwinding() {
			defer coroutine.Pop(&_c.Stack)
			for _, f := range _f0.X3 {
				defer f()
			}
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X3 = append(_f0.X3, func() {
			*_f0.X0 = _f0.X2
		})
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		coroutine.Yield[int, any](_f0.X1)
	}
}

//go:noinline
func RangeYieldAndDeferAssign(_fn0 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 int
		X1 int
	} = coroutine.Push[struct {
		IP int
		X0 int
		X1 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 int
			X1 int
		}{X0: _fn0}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X1 = 0
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		for ; _f0.X1 < _f0.X0; _f0.IP = 2 {
			YieldAndDeferAssign(&_f0.X1, _f0.X1, _f0.X1+1)
		}
	}
}

type MethodGeneratorState struct{ i int }

//go:noinline
func (_fn0 *MethodGeneratorState) MethodGenerator(_fn1 int) {
	_c := coroutine.LoadContext[int, any]()
	var _f0 *struct {
		IP int
		X0 *MethodGeneratorState
		X1 int
	} = coroutine.Push[struct {
		IP int
		X0 *MethodGeneratorState
		X1 int
	}](&_c.Stack)
	if _f0.IP == 0 {
		*_f0 = struct {
			IP int
			X0 *MethodGeneratorState
			X1 int
		}{X0: _fn0, X1: _fn1}
	}
	defer func() {
		if !_c.Unwinding() {
			coroutine.Pop(&_c.Stack)
		}
	}()
	switch {
	case _f0.IP < 2:
		_f0.X0.
			i = 0
		_f0.IP = 2
		fallthrough
	case _f0.IP < 3:
		for ; _f0.X0.i <= _f0.X1; _f0.X0.i, _f0.IP = _f0.X0.i+1, 2 {
			coroutine.Yield[int, any](_f0.X0.i)
		}
	}
}
func init() {
	_types.RegisterFunc[func(n int)]("github.com/stealthrocket/coroutine/compiler/testdata.Double")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.EvenSquareGenerator")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.FizzBuzzIfGenerator")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.FizzBuzzSwitchGenerator")
	_types.RegisterFunc[func(n int)]("github.com/stealthrocket/coroutine/compiler/testdata.Identity")
	_types.RegisterFunc[func(_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.LoopBreakAndContinue")
	_types.RegisterFunc[func(_fn1 int)]("github.com/stealthrocket/coroutine/compiler/testdata.MethodGenerator")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.NestedLoops")
	_types.RegisterFunc[func(_fn0 int, _fn1 func(int))]("github.com/stealthrocket/coroutine/compiler/testdata.Range")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureCapturingPointers")
	_types.RegisterClosure[func() (_ bool), struct {
		F  uintptr
		X0 *struct {
			IP int
			X0 int
			X1 int
			X2 *int
			X3 *int
			X4 func() bool
			X5 bool
			X6 bool
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureCapturingPointers.func2")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureCapturingValues")
	_types.RegisterClosure[func() (_ bool), struct {
		F  uintptr
		X0 *struct {
			IP int
			X0 int
			X1 int
			X2 func() bool
			X3 bool
			X4 bool
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureCapturingValues.func2")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureHeterogenousCapture")
	_types.RegisterClosure[func() int, struct {
		F  uintptr
		X0 *struct {
			IP  int
			X0  int8
			X1  int16
			X2  int32
			X3  int64
			X4  uint8
			X5  uint16
			X6  uint32
			X7  uint64
			X8  uintptr
			X9  func() int
			X10 int
			X11 func() bool
			X12 bool
			X13 bool
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureHeterogenousCapture.func2")
	_types.RegisterClosure[func() (_ bool), struct {
		F  uintptr
		X0 *struct {
			IP  int
			X0  int8
			X1  int16
			X2  int32
			X3  int64
			X4  uint8
			X5  uint16
			X6  uint32
			X7  uint64
			X8  uintptr
			X9  func() int
			X10 int
			X11 func() bool
			X12 bool
			X13 bool
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.Range10ClosureHeterogenousCapture.func3")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.Range10Heterogenous")
	_types.RegisterFunc[func(_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeArrayIndexValueGenerator")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeOverMaps")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeReverseClosureCaptureByValue")
	_types.RegisterClosure[func(), struct {
		F  uintptr
		X0 *struct {
			IP int
			X0 int
			X1 int
			X2 func()
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.RangeReverseClosureCaptureByValue.func2")
	_types.RegisterFunc[func(_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeSliceIndexGenerator")
	_types.RegisterFunc[func(n int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeTriple")
	_types.RegisterFunc[func(i int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeTriple.func1")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeTripleFuncValue")
	_types.RegisterFunc[func(i int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeTripleFuncValue.func2")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.RangeYieldAndDeferAssign")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.Select")
	_types.RegisterFunc[func(_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.Shadowing")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.SomeFunctionThatShouldExistInTheCompiledFile")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.SquareGenerator")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.SquareGeneratorTwice")
	_types.RegisterFunc[func(_fn0 int)]("github.com/stealthrocket/coroutine/compiler/testdata.SquareGeneratorTwiceLoop")
	_types.RegisterFunc[func(_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.TypeSwitchingGenerator")
	_types.RegisterFunc[func(_fn0 *int, _fn1, _fn2 int)]("github.com/stealthrocket/coroutine/compiler/testdata.YieldAndDeferAssign")
	_types.RegisterClosure[func(), struct {
		F  uintptr
		X0 *struct {
			IP int
			X0 *int
			X1 int
			X2 int
			X3 []func()
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.YieldAndDeferAssign.func2")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.YieldingDurations")
	_types.RegisterClosure[func(), struct {
		F  uintptr
		X0 *struct {
			IP int
			X0 *time.Duration
			X1 time.Duration
			X2 func()
			X3 int
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.YieldingDurations.func2")
	_types.RegisterFunc[func()]("github.com/stealthrocket/coroutine/compiler/testdata.YieldingExpressionDesugaring")
	_types.RegisterClosure[func(), struct {
		F  uintptr
		X0 *struct {
			IP  int
			X0  int
			X1  int
			X2  int
			X3  int
			X4  int
			X5  int
			X6  int
			X7  int
			X8  int
			X9  bool
			X10 int
			X11 int
			X12 int
			X13 int
			X14 int
			X15 bool
			X16 int
			X17 int
			X18 int
			X19 int
			X20 int
			X21 bool
			X22 int
			X23 int
			X24 int
			X25 int
			X26 bool
			X27 bool
			X28 int
			X29 int
			X30 int
			X31 int
			X32 int
			X33 bool
			X34 int
			X35 int
			X36 bool
			X37 int
			X38 int
			X39 int
			X40 bool
			X41 int
			X42 int
			X43 int
			X44 bool
			X45 int
			X46 int
			X47 any
			X48 []func()
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.YieldingExpressionDesugaring.func2")
	_types.RegisterClosure[func(), struct {
		F  uintptr
		X0 *struct {
			IP  int
			X0  int
			X1  int
			X2  int
			X3  int
			X4  int
			X5  int
			X6  int
			X7  int
			X8  int
			X9  bool
			X10 int
			X11 int
			X12 int
			X13 int
			X14 int
			X15 bool
			X16 int
			X17 int
			X18 int
			X19 int
			X20 int
			X21 bool
			X22 int
			X23 int
			X24 int
			X25 int
			X26 bool
			X27 bool
			X28 int
			X29 int
			X30 int
			X31 int
			X32 int
			X33 bool
			X34 int
			X35 int
			X36 bool
			X37 int
			X38 int
			X39 int
			X40 bool
			X41 int
			X42 int
			X43 int
			X44 bool
			X45 int
			X46 int
			X47 any
			X48 []func()
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.YieldingExpressionDesugaring.func3")
	_types.RegisterClosure[func(), struct {
		F  uintptr
		X0 *struct {
			IP  int
			X0  int
			X1  int
			X2  int
			X3  int
			X4  int
			X5  int
			X6  int
			X7  int
			X8  int
			X9  bool
			X10 int
			X11 int
			X12 int
			X13 int
			X14 int
			X15 bool
			X16 int
			X17 int
			X18 int
			X19 int
			X20 int
			X21 bool
			X22 int
			X23 int
			X24 int
			X25 int
			X26 bool
			X27 bool
			X28 int
			X29 int
			X30 int
			X31 int
			X32 int
			X33 bool
			X34 int
			X35 int
			X36 bool
			X37 int
			X38 int
			X39 int
			X40 bool
			X41 int
			X42 int
			X43 int
			X44 bool
			X45 int
			X46 int
			X47 any
			X48 []func()
		}
	}]("github.com/stealthrocket/coroutine/compiler/testdata.YieldingExpressionDesugaring.func4")
	_types.RegisterFunc[func(_fn0 int) (_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.a")
	_types.RegisterFunc[func(_fn0 int) (_ int)]("github.com/stealthrocket/coroutine/compiler/testdata.b")
}
