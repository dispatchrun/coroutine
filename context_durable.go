//go:build durable

package coroutine

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

	// Booleans managing the completion state of the coroutine.
	done bool
	stop bool

	// Entry point of the coroutine, this is captured so the associated
	// generator can call into the coroutine to start or resume it at the
	// last yield point.
	Func func()

	Stack
	Heap
}

// MarshalAppend appends a serialized Context to the provided buffer.
func (c *Context[R, S]) MarshalAppend(b []byte) ([]byte, error) {
	var err error
	b, err = c.Stack.MarshalAppend(b)
	if err != nil {
		return b, err
	}
	return c.Heap.MarshalAppend(b)
}

// Unmarshal deserializes a Context from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// context.
func (c *Context[R, S]) Unmarshal(b []byte) (int, error) {
	sn, err := c.Stack.Unmarshal(b)
	if err != nil {
		return 0, err
	}
	hn, err := c.Heap.Unmarshal(b[sn:])
	if err != nil {
		return 0, err
	}
	return sn + hn, err
}

// TODO: do we have use cases for yielding more than one value?
func (c *Context[R, S]) Yield(value R) S {
	if c.stop {
		panic("cannot yield from a coroutine that has been stopped")
	}
	if frame := c.Top(); frame.Resume {
		frame.Resume = false
		if c.stop {
			panic(unwind{})
		}
		return c.send
	} else {
		var zero S
		frame.Resume = true
		c.send = zero
		c.recv = value
		panic(unwind{})
	}
}

// Unwinding returns true if the coroutine is currently unwinding its stack.
func (c *Context[R, S]) Unwinding() bool {
	return len(c.Frames) > 0 && c.Top().Resume
}

// Rewinding returns true if the coroutine is currently rewinding its stack.
func (c *Context[R, S]) Rewinding() bool {
	return len(c.Frames) > 0 && c.Top().Resume
}

type unwind struct{}
