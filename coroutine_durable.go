//go:build durable

package coroutine

import "github.com/stealthrocket/coroutine/internal/serde"

// New creates a new coroutine which executes f as entry point.
func New[R, S any](f func()) Coroutine[R, S] {
	return Coroutine[R, S]{
		ctx: &Context[R, S]{
			context: context{entry: f},
		},
	}
}

// Next executes the coroutine until its next yield point, or until completion.
// The method returns true if the coroutine entered a yield point, after which
// the program should call Recv to obtain the value that the coroutine yielded,
// and Send to set the value that will be returned from the yield point.
func (c Coroutine[R, S]) Next() (hasNext bool) {
	if c.ctx.done {
		return false
	}

	g := getg()
	storeContext(g, c.ctx)

	defer func() {
		clearContext(g)

		switch err := recover(); err {
		case nil:
		case unwind{}:
		default:
			// TODO: can we figure out a way to know when we are unwinding the
			// stack and only recover then so we don't alter the panic stack?
			panic(err)
		}

		if c.ctx.Unwinding() {
			stop := c.ctx.stop
			c.ctx.done, hasNext = stop, !stop
		} else {
			c.ctx.done = true
		}
	}()

	c.ctx.Stack.FP = -1
	c.ctx.entry()
	return false
}

type context struct {
	// Entry point of the coroutine, this is captured so the associated
	// generator can call into the coroutine to start or resume it at the
	// last yield point.
	entry func()

	Stack
}

// TODO: do we have use cases for yielding more than one value?
func (c *Context[R, S]) Yield(value R) S {
	if c.resume {
		c.resume = false
		if c.stop {
			panic(unwind{})
		}
		return c.send
	} else {
		if c.stop {
			panic("cannot yield from a coroutine that has been stopped")
		}
		var zero S
		c.resume = true
		c.send = zero
		c.recv = value
		panic(unwind{})
	}
}

type unwind struct{}

// Unwinding returns true if the coroutine is currently unwinding its stack.
func (c *Context[R, S]) Unwinding() bool {
	return c.resume
}

// Marshal returns a serialized Context.
func (c *Context[R, S]) Marshal() ([]byte, error) {
	return c.MarshalAppend(nil)
}

// MarshalAppend appends a serialized Context to the provided buffer.
func (c *Context[R, S]) MarshalAppend(b []byte) ([]byte, error) {
	s := serde.Serialize(&serializedCoroutine{
		entry:  c.entry,
		stack:  c.Stack,
		resume: c.resume,
	})
	return append(b, s...), nil
}

type serializedCoroutine struct {
	entry  func()
	stack  Stack
	resume bool
}

// Unmarshal deserializes a Context from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// context.
func (c *Context[R, S]) Unmarshal(b []byte) (int, error) {
	start := len(b)
	v, b := serde.Deserialize(b)
	s := v.(*serializedCoroutine)
	c.entry = s.entry
	c.Stack = s.stack
	c.resume = s.resume
	sn := start - len(b)
	return sn, nil
}
