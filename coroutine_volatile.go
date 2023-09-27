//go:build !durable

package coroutine

import (
	"runtime"

	"github.com/stealthrocket/coroutine/internal/gls"
)

// Durable is a constant which takes the values true or false depending on
// whether the program is built with the "durable" tag.
const Durable = false

// New creates a new coroutine which executes f as entry point.
func New[R, S any](f func()) Coroutine[R, S] {
	c := &Context[R, S]{
		context: context{
			next: make(chan struct{}),
		},
	}

	go func() {
		g := gls.Context()
		g.Store(c)

		defer func() {
			c.done = true
			close(c.next)
			g.Clear()
		}()

		<-c.next

		if !c.stop {
			f()
		}
	}()

	return Coroutine[R, S]{ctx: c}
}

// Next executes the coroutine until its next yield point, or until completion.
// The method returns true if the coroutine entered a yield point, after which
// the program should call Recv to obtain the value that the coroutine yielded,
// and Send to set the value that will be returned from the yield point.
func (c Coroutine[R, S]) Next() bool {
	if c.ctx.done {
		return false
	}
	c.ctx.next <- struct{}{}
	_, ok := <-c.ctx.next
	return ok
}

type context struct {
	next chan struct{}
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

func (c *Context[R, S]) Marshal() ([]byte, error) {
	return nil, ErrNotDurable
}

func (c *Context[R, S]) Unmarshal(b []byte) (int, error) {
	return 0, ErrNotDurable
}
