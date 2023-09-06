//go:build durable

package coroutine

type Coroutine[R, S any] struct{ ctx *Context[R, S] }

func (c Coroutine[R, S]) Context() *Context[R, S] { return c.ctx }

func (c Coroutine[R, S]) Recv() R { return c.ctx.recv }

func (c Coroutine[R, S]) Send(v S) { c.ctx.send = v }

func (c Coroutine[R, S]) Stop() { c.ctx.stop = true }

func (c Coroutine[R, S]) Done() bool { return c.ctx.done }

func (c Coroutine[R, S]) Next() (hasNext bool) {
	if c.ctx.done {
		return false
	}

	g := getg()
	storeContext(g, c.ctx)

	defer func() {
		clearContext(g)

		if c.ctx.Unwinding() {
			recover()
			stop := c.ctx.stop
			c.ctx.done, hasNext = stop, !stop
		} else {
			c.ctx.done = true
		}
	}()

	c.ctx.Stack.FP = -1
	c.ctx.Func()
	return false
}

func New[R, S any](f func()) Coroutine[R, S] {
	return Coroutine[R, S]{ctx: &Context[R, S]{Func: f}}
}
