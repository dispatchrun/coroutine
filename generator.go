//go:build !durable

package coroutine

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

func New[R, S any](f func()) Generator[R, S] {
	c := &Context[R, S]{
		next: make(chan struct{}),
	}

	go func() {
		g := getg()
		storeContext(g, c)
		defer clearContext(g)
		defer close(c.next)

		<-c.next
		f()
	}()

	return Generator[R, S]{ctx: c}
}
