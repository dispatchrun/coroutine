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

// Push prepares the stack for an impending function call.
//
// The stack's frame pointer is incremented, and the stack is resized
// to make room for a new frame if the caller is on the topmost frame.
//
// If the caller is not on the topmost frame it means that a coroutine
// is being resumed and the next frame is already present on the stack.
//
// The Frame is returned by value rather than by reference, since the
// stack's underlying frame backing array might change. Callers
// intending to serialize the stack should call Store(fp, frame) for each
// frame during stack unwinding.
func (s *Stack) Push() (frame Frame, fp int) {
	if s.isTop() {
		s.Frames = append(s.Frames, Frame{})
	}
	s.FP++
	return s.Frames[s.FP], s.FP
}

// Pop pops the topmost stack frame after a function call.
func (s *Stack) Pop() {
	if !s.isTop() {
		panic("pop when caller is not on topmost frame")
	}
	s.Frames = s.Frames[:len(s.Frames)-1]
	s.FP--
}

// Store stores a frame at the specified index.
func (s *Stack) Store(i int, f Frame) {
	if i < 0 || i >= len(s.Frames) {
		panic("invalid frame index")
	}
	s.Frames[i] = f
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
}
