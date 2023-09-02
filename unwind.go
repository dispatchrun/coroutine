package coroutine

// Unwind initiates stack unwinding in a coroutine.
func Unwind() {
	panic(unwind)
}

var unwind struct{}

// Unwinding reports whether stack unwinding is taking place.
// It should be called inside a defer and given the value
// returned by recover().
func Unwinding(v any) bool {
	return v == unwind
}
