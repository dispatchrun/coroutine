//go:build durable

package coroutine

// Stack is the call stack for a coroutine.
type Stack struct {
	// FP is the frame pointer. Functions always use the Frame
	// located at Frames[FP].
	FP int

	// Frames is the set of stack frames.
	Frames []Frame
}

// Top returns the top of the call stack.
func (s *Stack) Top() *Frame {
	if len(s.Frames) == 0 {
		panic("no stack frames")
	}
	return &s.Frames[len(s.Frames)-1]
}

// Push prepares the stack for an impending function call.
//
// The stack's frame pointer is incremented, and a Frame is pushed to the
// stack if the caller is on the topmost frame.
//
// If the caller is not on the topmost frame it means that a coroutine
// is being resumed and the next frame is already present on the stack.
func (s *Stack) Push() *Frame {
	if s.isTop() {
		s.Frames = append(s.Frames, Frame{})
	}
	s.FP++
	return &s.Frames[s.FP]
}

// Pop pops the topmost stack frame after a function call.
func (s *Stack) Pop() {
	if !s.isTop() {
		panic("pop when caller is not on topmost frame")
	}
	s.Frames = s.Frames[:len(s.Frames)-1]
	s.FP--
}

func (s *Stack) isTop() bool {
	return s.FP == len(s.Frames)-1
}

// Frame is a stack frame.
//
// A frame is created when a function is called and torn down after it
// returns. A Frame holds the position of execution within that function,
// and any Serializable objects that it uses or returns.
type Frame struct {
	// IP is the instruction pointer.
	IP int

	// Storage holds the Serializable objects on the frame.
	Storage

	// Resume is true if the function associated with the frame
	// previously yielded.
	Resume bool
}
