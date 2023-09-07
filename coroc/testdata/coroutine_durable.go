// Code generated by coroc. DO NOT EDIT

//go:build durable

package testdata

import (
	"github.com/stealthrocket/coroutine"
)

func Identity(n int) {
	_c := coroutine.LoadContext[int, any]()

	_f := _c.Push()
	defer _c.Pop()

	if _c.Rewinding() {
		n = int(_f.Get(0).(coroutine.Int))
	}

	defer func() {
		if _c.Unwinding() {
			_f.Set(0, coroutine.Int(n))
		}
	}()

	coroutine.Yield[int, any](n)
}

func SquareGenerator(n int) {
	_c := coroutine.LoadContext[int, any]()

	_f := _c.Push()
	defer _c.Pop()

	var (
		_v0 int
	)

	if _c.Rewinding() {
		n = int(_f.Get(0).(coroutine.Int))
		_v0 = int(_f.Get(1).(coroutine.Int))
	}

	defer func() {
		if _c.Unwinding() {
			_f.Set(0, coroutine.Int(n))
			_f.Set(1, coroutine.Int(_v0))
		}
	}()

	switch {
	case _f.IP < 1:
		_v0 = 1
		_f.IP = 1
		fallthrough
	case _f.IP < 2:
		for ; _v0 <= n; _v0++ {
			coroutine.Yield[int, any](_v0 * _v0)
		}
	}
}

