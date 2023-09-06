//go:build !durable

package coroutine

type Context[R, S any] struct {
	recv R
	send S
	next chan struct{}
}

func (c *Context[R, S]) Yield(v R) S {
	var zero S
	c.send = zero
	c.recv = v
	c.next <- struct{}{}
	<-c.next
	return c.send
}
