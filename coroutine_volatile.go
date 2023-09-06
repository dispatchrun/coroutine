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

// Next executes the coroutine until its next yield point, or until completion.
// The method returns true if the coroutine entered a yield point, after which
// the program should call Recv to obtain the value that the coroutine yielded,
// and Send to set the value that will be returned from the yield point.
func (c Coroutine[R, S]) Next() bool {
	c.ctx.next <- struct{}{}
	_, ok := <-c.ctx.next
	return ok
}

// New creates a new coroutine which executes f.
func New[R, S any](f func()) Coroutine[R, S] {
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

	return Coroutine[R, S]{ctx: c}
}
