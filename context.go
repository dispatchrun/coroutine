package coroutine

// Context is passed to a coroutine and flows through all
// functions that Yield (or could yield).
type Context struct {
	Stack
	Heap

	// Value passed to Yield when a coroutine yields control back to its caller,
	// and value returned to the coroutine when the caller resumes it.
	yield any
}

// MarshalAppend appends a serialized Context to the provided buffer.
func (c *Context) MarshalAppend(b []byte) ([]byte, error) {
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
func (c *Context) Unmarshal(b []byte) (int, error) {
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

func (c *Context) Recv() any {
	return c.yield
}

func (c *Context) Send(value any) {
	c.yield = value
}

// TODO: do we have use cases for yielding more than one value?
func (c *Context) Yield(value any) any {
	if frame := c.Top(); frame.Resume {
		frame.Resume = false
		return c.yield
	} else {
		frame.Resume = true
		c.yield = value
		panic(unwind{})
	}
}

// Unwinding returns true if the coroutine is currently unwinding its stack.
func (c *Context) Unwinding() bool {
	return len(c.Frames) > 0 && c.Top().Resume
}
