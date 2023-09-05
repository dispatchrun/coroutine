//go:build durable

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

func (g Generator[R, S]) Next() (hasNext bool) {
	defer func() {
		if g.ctx.Unwinding() {
			recover()
			hasNext = true
		}
	}()
	g.ctx.Stack.FP = -1
	g.ctx.Func(g.ctx)
	return false
}

func (g Generator[R, S]) Context() *Context[R, S] {
	return g.ctx
}

func New[R, S any](f func(*Context[R, S])) Generator[R, S] {
	return Generator[R, S]{ctx: &Context[R, S]{Func: f}}
}
