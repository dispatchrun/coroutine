//go:build durable

package testdata

import "github.com/stealthrocket/coroutine"

func Identity(n int) {
	c := coroutine.LoadContext[int, any]()
	c.Push()
	c.Yield(n)
	c.Pop()
}

func SquareGenerator(n int) {
	c := coroutine.LoadContext[int, any]()

	// new stack frame, or reuse current frame on resume
	frame := c.Push()

	// variable declaration
	var (
		i int
	)

	// state restoration
	switch frame.IP {
	case 1:
		n = int(frame.Get(0).(coroutine.Int))
		i = int(frame.Get(1).(coroutine.Int))
	}

	// state capture
	defer func() {
		if c.Unwinding() {
			switch frame.IP {
			case 1:
				frame.Set(0, coroutine.Int(n))
				frame.Set(1, coroutine.Int(i))
			}
		}
	}()

	// coroutine state machine
	switch frame.IP {
	case 0:
		i = 1
		frame.IP = 1
		fallthrough
	case 1:
		for i <= n {
			c.Yield(i * i)
			i++
		}
	}

	// pop stack frame now that the function call completed
	c.Pop()
}
