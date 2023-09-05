//go:build !durable

package coroutine

import "github.com/jtolds/gls"

type Generator[R, S any] struct {
	ctx *Context[R, S]
}

func (g Generator[R, S]) Recv() R {
	return g.ctx.recv
}

func (g Generator[R, S]) Send(v S) {
	g.ctx.send = v
}

func (g Generator[R, S]) Next() bool {
	g.ctx.next <- struct{}{}
	_, ok := <-g.ctx.next
	return ok
}

func New[R, S any](f func(*Context[R, S])) Generator[R, S] {
	c := &Context[R, S]{
		next: make(chan struct{}),
	}

	go func() {
		defer close(c.next)

		goroutine.SetValues(gls.Values{contextKey{}: c}, func() {
			<-c.next
			f(c)
		})
	}()

	return Generator[R, S]{ctx: c}
}
