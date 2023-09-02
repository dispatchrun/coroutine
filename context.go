package coroutine

// Context is passed to a coroutine and flows through all
// functions that Yield (or could yield).
type Context struct {
	Stack
	Heap

	YieldValue Serializable

	Unwinding bool
}

// NewContext constructs a Context and prepares its stack so that
// a coroutine can be invoked with the specified arguments.
func NewContext(args ...Serializable) *Context {
	return &Context{
		Stack: Stack{
			Frames: []Frame{
				{
					Storage: NewStorage(args),
				},
			},
		},
	}
}

// Call invokes or resumes a coroutine.
//
// The boolean return value indicates whether the coroutine returned
// normally (true) or whether it yielded (false).
func (c *Context) Call(fn Coroutine) (returned bool) {
	defer func() {
		if err := recover(); Unwinding(err) {
			returned = false
		} else if err != nil {
			panic(err)
		}
	}()

	c.Stack.FP = 0

	fn(c)
	return true
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

// TODO: do we have use cases for yielding more than one value?
// TODO: the program should be able to send a value that is returned by Yield on resume
func (c *Context) Yield(value Serializable, capture func()) {
	if frame := c.Top(); frame.Resume {
		frame.Resume = false
	} else {
		frame.Resume = true
		c.Unwinding = true
		c.YieldValue = value
		capture()
		Unwind()
	}
}
