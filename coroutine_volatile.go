//go:build !durable

package coroutine

import (
	"runtime"
	"sync"
	"unsafe"
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
		execute(c, func() {
			defer func() {
				c.done = true
				close(c.next)
			}()

			<-c.next

			if !c.stop {
				f()
			}
		})
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

// The offset from the high address of the stack pointer where the v argument
// of the execute function is stored.
//
// We use a once value to lazily initialize the value when executing coroutines
// because we must compute the exact distance from the high stack pointer on the
// coroutine entry point code path. After initialization, the global offset
// variable is only read from the same goroutine, so there is no race since the
// last write is always observed.
var (
	offset     uintptr
	offsetOnce sync.Once
)

// The load function returns the value passed as first argument to the call to
// execute that started the coroutine.
func load() any {
	g := getg()
	p := unsafe.Pointer(g.stack.hi - offset)
	return *(*any)(p)
}

// The execute function is the entry point of coroutines, it pushes the
// coroutine context in v to the stack, registering it to be retrieved by
// calling load, and invokes f as the entry point.
//
// The coroutine continues execution until a yield point is reached or until
// the function passed as entry point returns.
//
// The function has go:nosplit because the address of its local variables must
// remain stable
//
//go:nosplit
//go:noinline
func execute(v any, f func()) {
	p := unsafe.Pointer(&v)

	offsetOnce.Do(func() {
		g := getg()
		// In volatile mode a new goroutine is started to back each coroutine,
		// which means that we have control over the distance from the call to
		// with and the base pointer of the goroutine stack; we can store the
		// offset in a global. It does not matter if this write is performed
		// from concurrent threads, it always has the same value.
		offset = g.stack.hi - uintptr(p)
	})

	f()

	// Keep the variable alive so we know that it will remain on the stack and
	// won't be erased by the GC.
	runtime.KeepAlive(v)
}
