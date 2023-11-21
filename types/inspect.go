package types

import coroutinev1 "github.com/stealthrocket/coroutine/gen/proto/go/coroutine/v1"

// Inspect inspects serialized durable coroutine state.
//
// The input should be a buffer produced by (*coroutine.Context).Marshal
// or by types.Serialize.
func Inspect(b []byte) (*State, error) {
	var state coroutinev1.State
	if err := state.UnmarshalVT(b); err != nil {
		return nil, err
	}
	return &State{state: &state}, nil
}

// State wraps durable coroutine state.
type State struct {
	state *coroutinev1.State
}

// BuildID returns the build ID of the program that generated this state.
func (s *State) BuildID() string {
	return s.state.Build.Id
}

// OS returns the operating system the coroutine was compiled for.
func (s *State) OS() string {
	return s.state.Build.Os
}

// Arch returns the architecture the coroutine was compiled for.
func (s *State) Arch() string {
	return s.state.Build.Arch
}

// NumType returns the number of types referenced by the coroutine.
func (s *State) NumType() int {
	return len(s.state.Types)
}

// NumFunction returns the number of functions/methods/closures
// referenced by the coroutine.
func (s *State) NumFunction() int {
	return len(s.state.Functions)
}

// NumRegion returns the number of memory regions referenced by the
// coroutine.
func (s *State) NumRegion() int {
	return len(s.state.Regions)
}
