package coroutine

import (
	"encoding/binary"
	"fmt"
)

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

// MarshalAppend appends a serialized Stack to the provided buffer.
func (s *Stack) MarshalAppend(b []byte) ([]byte, error) {
	// The frame pointer is always set to zero when invoking or
	// resuming a coroutine, so no need to serialize it here.
	b = binary.AppendVarint(b, int64(len(s.Frames)))
	for i := range s.Frames {
		var err error
		b, err = s.Frames[i].MarshalAppend(b)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

// Unmarshal deserializes a Stack from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// stack.
func (s *Stack) Unmarshal(b []byte) (int, error) {
	frameCount, n := binary.Varint(b)
	if n <= 0 {
		return 0, fmt.Errorf("invalid stack frame count: %v", b)
	}
	s.Frames = make([]Frame, frameCount)
	for i := range s.Frames {
		fn, err := s.Frames[i].Unmarshal(b[n:])
		if err != nil {
			return 0, err
		}
		n += fn
	}
	return n, nil
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

// MarshalAppend appends a serialized Frame to the provided buffer.
func (f *Frame) MarshalAppend(b []byte) ([]byte, error) {
	b = binary.AppendVarint(b, int64(f.IP))
	if f.Resume {
		b = append(b, 1)
	} else {
		b = append(b, 0)
	}
	return f.Storage.MarshalAppend(b)
}

// Unmarshal deserializes a Frame from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// frame.
func (f *Frame) Unmarshal(b []byte) (int, error) {
	ip, n := binary.Varint(b)
	if n <= 0 || int64(int(ip)) != ip {
		return 0, fmt.Errorf("invalid frame instruction pointer: %v", b)
	}
	if n >= len(b) || (b[n] != 0 && b[n] != 1) {
		return 0, fmt.Errorf("invalid frame resume flag: %v", b)
	}
	f.Resume = b[n] == 1
	n++

	var storage Storage
	sn, err := storage.Unmarshal(b[n:])
	if err != nil {
		return 0, err
	}
	n += sn

	f.IP = int(ip)
	f.Storage = storage
	return n, nil
}
