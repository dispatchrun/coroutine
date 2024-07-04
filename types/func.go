// Package types contains the infrastructure needed to support serialization
// of Go types.
package types

import "github.com/dispatchrun/coroutine/internal/reflectext"

// RegisterFunc is a helper function used to register function types. The type
// parameter must be a function type, but no compile nor runtime checks are used
// to enforce it; passing anything other than a function type will likely result
// in panics later on when the program attempts to serialize the function value.
//
// The name argument is a unique identifier of the Go symbol that represents the
// function, which has the package path as prefix, and the dot-separated sequence
// identifying the function in the package.
func RegisterFunc[Type any](name string) {
	reflectext.RegisterFunc[Type](name)
}

// RegisterClosure is like RegisterFunc but the caller can specify the closure
// type (see types.Func for details).
func RegisterClosure[Type, Closure any](name string) {
	reflectext.RegisterClosure[Type, Closure](name)
}
