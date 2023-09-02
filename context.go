package coroutine

// Context is passed to a coroutine and flows through all
// functions that Yield (or could yield).
type Context struct {
	Stack
	Heap
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
