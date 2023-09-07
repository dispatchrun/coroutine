package coroutine

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
	switch c := loadContext(getg()).(type) {
	case *Context[R, S]:
		return c
	case nil:
		panic("coroutine.Yield: not called from a coroutine stack")
	default:
		panic("coroutine.Yield: coroutine type mismatch")
	}
}
