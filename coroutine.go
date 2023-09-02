package coroutine

// Coroutine is a function that can be suspended and resumed.
type Coroutine func(*Context)
