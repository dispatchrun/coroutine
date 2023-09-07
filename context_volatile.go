//go:build !durable

package coroutine

import (
	"runtime"
)

type Context[R, S any] struct {
	recv R
	send S
	next chan struct{}
	stop bool
	done bool
}

func (c *Context[R, S]) Yield(v R) S {
	if c.stop {
		panic("cannot yield from a coroutine that has been stopped")
	}
	var zero S
	c.send = zero
	c.recv = v
	c.next <- struct{}{}
	<-c.next
	if c.stop {
		runtime.Goexit()
	}
	return c.send
}
