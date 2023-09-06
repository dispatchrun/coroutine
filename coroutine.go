package coroutine

// Run executes a coroutine to completion, calling f for each value that the
// coroutine yields, and sending back each value that f returns.
func Run[R, S any](c Coroutine[R, S], f func(R) S) {
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
	if c := loadContext(getg()); c != nil {
		return c.(*Context[R, S])
	} else {
		panic("coroutine.Yield: not called from a coroutine stack")
	}
}
