//go:build !durable

package coroutine

// Coroutine instances expose APIs allowing the program to drive the execution
// of coroutines.
//
// The type parameter R represents the type of values that the program can
// receive from the coroutine (what it yields), and the type parameter S is
// what the program can send back to a coroutine yield point.
type Coroutine[R, S any] struct{ ctx *Context[R, S] }

// Recv returns the last value that the coroutine has yielded. The method must
// be called only after a call to Next has returned true, or the return value is
// undefined. Calling the method multiple times after a call to Next returns the
// same value each time.
func (c Coroutine[R, S]) Recv() R { return c.ctx.recv }

// Send sets the value that will be seen by the coroutine after it resumes from
// a yield point. Calling the method multiple times before a call to Next does
// not result in sending multiple values, only the last value sent will be seen
// by the coroutine.
func (c Coroutine[R, S]) Send(v S) { c.ctx.send = v }

// Stop interrupts the coroutine. On the next call to Next, the coroutine will
// not return from its yield point; instead, it unwinds its call stack, calling
// each defer statement in the inverse order that they were declared.
//
// Stop is idempotent, calling it multiple times or after completion of the
// coroutine has no effect.
//
// This method is just an interrupt mechanism, the program does not have to call
// it to release the coroutine resources after completion.
func (c Coroutine[R, S]) Stop() { c.ctx.stop = true }

// Done returns true if the coroutine completed, either because it was stopped
// or because its function returned.
func (c Coroutine[R, S]) Done() bool { return c.ctx.done }

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

// New creates a new coroutine which executes f as entry point.
func New[R, S any](f func()) Coroutine[R, S] {
	c := &Context[R, S]{
		next: make(chan struct{}),
	}

	go func() {
		g := getg()
		storeContext(g, c)

		defer func() {
			c.done = true
			close(c.next)
			clearContext(g)
		}()

		<-c.next

		if !c.stop {
			f()
		}
	}()

	return Coroutine[R, S]{ctx: c}
}
