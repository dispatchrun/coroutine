//go:build !durable

package coroutine

import (
	"sync"
	"unsafe"
)

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
		g := getg()
		gls.Store(g, c)
		defer gls.Delete(g)
		defer close(c.next)

		<-c.next
		f(c)
	}()

	return Generator[R, S]{ctx: c}
}

// goroutine local storage; the map contains one entry for each goroutine that
// is started to power a coroutine.
var gls sync.Map

// getg is like the compiler intrisinc runtime.getg which retrieves the current
// goroutine object.
//
// https://github.com/golang/go/blob/a2647f08f0c4e540540a7ae1b9ba7e668e6fed80/src/runtime/HACKING.md?plain=1#L44-L54
func getg() unsafe.Pointer
