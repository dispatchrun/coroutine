package coroutine

import (
	"errors"
)

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

// Context returns the coroutine's associated Context.
func (c Coroutine[R, S]) Context() *Context[R, S] { return c.ctx }

// Context is passed to a coroutine and flows through all
// functions that Yield (or could yield).
type Context[R, S any] struct {
	// Value passed to Yield when a coroutine yields control back to its caller,
	// and value returned to the coroutine when the caller resumes it.
	//
	// Keep as first fields so they don't use any space if they are the empty
	// struct.
	recv R
	send S

	// Booleans managing the state of the coroutine.
	done   bool
	stop   bool
	resume bool //nolint

	context
}

// Run executes a coroutine to completion, calling f for each value that the
// coroutine yields, and sending back each value that f returns.
func Run[R, S any](c Coroutine[R, S], f func(R) S) {
	// The coroutine is run to completion, but f might panic in which case we
	// don't want to leave it in an uncompleted state and interrupt it instead.
	defer func() {
		if !c.Done() {
			c.Stop()
			c.Next()
		}
	}()

	for c.Next() {
		r := c.Recv()
		s := f(r)
		c.Send(s)
	}
}

// Yield sends v to the generator and pauses the execution of the coroutine
// until the Next method is called on the associated generator.
//
// The function panics when called on a stack where no active coroutine exists,
// or if the type parameters do not match those of the coroutine.
func Yield[R, S any](v R) S {
	return LoadContext[R, S]().Yield(v)
}

// LoadContext returns the context for the current coroutine.
//
// The function panics when called on a stack where no active coroutine exists,
// or if the type parameters do not match those of the coroutine.
func LoadContext[R, S any]() *Context[R, S] {
	switch c := load(gctx).(type) {
	case *Context[R, S]:
		return c
	case nil:
		panic("coroutine.Yield: not called from a coroutine stack")
	default:
		panic("coroutine.Yield: coroutine type mismatch")
	}
}

var gctx uintptr

func load(k uintptr) any

func with(k *uintptr, v any, f func())

// ErrNotDurable is an error that occurs when attempting to
// serialize a coroutine that is not durable.
var ErrNotDurable = errors.New("only durable coroutines can be serialized")
