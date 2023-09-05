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

func Yield[R, S any](v R) S {
	if c, _ := gls.Load(getg()); c != nil {
		return c.(*Context[R, S]).Yield(v)
	} else {
		panic("coroutine.Yield: not called from a coroutine stack")
	}
}
