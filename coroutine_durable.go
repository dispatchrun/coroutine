//go:build durable

package coroutine

import (
	"slices"
	"strconv"

	"github.com/stealthrocket/coroutine/internal/gls"
	"github.com/stealthrocket/coroutine/internal/serde"
)

// New creates a new coroutine which executes f as entry point.
func New[R, S any](f func()) Coroutine[R, S] {
	return Coroutine[R, S]{
		ctx: &Context[R, S]{
			context: context{entry: f},
		},
	}
}

// Stack is the call stack for a coroutine.
type Stack struct {
	// FP is the frame pointer. Functions always use the Frame
	// located at Frames[FP].
	FP int

	// Frames is the set of stack frames.
	Frames []any
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
func Push[Frame any](s *Stack) *Frame {
	if s.isTop() {
		s.Frames = append(s.Frames, new(Frame))
	}
	s.FP++
	return s.Frames[s.FP].(*Frame)
}

// Pop pops the topmost stack frame after a function call.
func Pop(s *Stack) {
	if !s.isTop() {
		panic("pop when caller is not on topmost frame")
	}
	i := len(s.Frames) - 1
	s.Frames[i] = nil
	s.Frames = s.Frames[:i]
	s.FP--
}

func (s *Stack) isTop() bool {
	return s.FP == len(s.Frames)-1
}

type serializedCoroutine struct {
	entry  func()
	stack  Stack
	resume bool
}

// Context is passed to a coroutine and flows through all
// functions that Yield (or could yield).
type Context[R, S any] struct {
	// Value passed to Yield when a coroutine yields control back to its caller,
	// and value returned to the coroutine when the caller resumes it.
	//
	// Keep as first fields so they don't use any space if they are the empty
	// struct.
	recv R
	send S

	// Booleans managing the state of the coroutine.
	done   bool
	stop   bool
	resume bool

	// Entry point of the coroutine, this is captured so the associated
	// generator can call into the coroutine to start or resume it at the
	// last yield point.
	entry func()

	Stack
}

// MarshalAppend appends a serialized Context to the provided buffer.
func (c *Context[R, S]) MarshalAppend(b []byte) ([]byte, error) {
	s := serde.Serialize(&serializedCoroutine{
		entry:  c.entry,
		stack:  c.Stack,
		resume: c.resume,
	})
	return append(b, s...), nil
}

// Unmarshal deserializes a Context from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// context.
func (c *Context[R, S]) Unmarshal(b []byte) (int, error) {
	start := len(b)
	v, b := serde.Deserialize(b)
	s := v.(*serializedCoroutine)
	c.entry = s.entry
	c.Stack = s.stack
	c.resume = s.resume
	sn := start - len(b)
	return sn, nil
}

// TODO: do we have use cases for yielding more than one value?
func (c *Context[R, S]) Yield(value R) S {
	if c.resume {
		c.resume = false
		if c.stop {
			panic(unwind{})
		}
		return c.send
	} else {
		if c.stop {
			panic("cannot yield from a coroutine that has been stopped")
		}
		var zero S
		c.resume = true
		c.send = zero
		c.recv = value
		panic(unwind{})
	}
}

// Next executes the coroutine until its next yield point, or until completion.
// The method returns true if the coroutine entered a yield point, after which
// the program should call Recv to obtain the value that the coroutine yielded,
// and Send to set the value that will be returned from the yield point.
func (c Coroutine[R, S]) Next() (hasNext bool) {
	if c.ctx.done {
		return false
	}

	g := gls.Context()

	g.Store(c.ctx)

	defer func() {
		g.Clear()

		switch err := recover(); err {
		case nil:
		case unwind{}:
		default:
			// TODO: can we figure out a way to know when we are unwinding the
			// stack and only recover then so we don't alter the panic stack?
			panic(err)
		}

		if c.ctx.Unwinding() {
			stop := c.ctx.stop
			c.ctx.done, hasNext = stop, !stop
		} else {
			c.ctx.done = true
		}
	}()

	c.ctx.Stack.FP = -1
	c.ctx.entry()
	return false
}

type context struct {
	// Entry point of the coroutine, this is captured so the associated
	// generator can call into the coroutine to start or resume it at the
	// last yield point.
	entry func()

	Stack
}

// TODO: do we have use cases for yielding more than one value?
func (c *Context[R, S]) Yield(value R) S {
	if c.resume {
		c.resume = false
		if c.stop {
			panic(unwind{})
		}
		return c.send
	} else {
		if c.stop {
			panic("cannot yield from a coroutine that has been stopped")
		}
		var zero S
		c.resume = true
		c.send = zero
		c.recv = value
		panic(unwind{})
	}
}

type unwind struct{}

// Unwinding returns true if the coroutine is currently unwinding its stack.
func (c *Context[R, S]) Unwinding() bool {
	return c.resume
}

// Marshal returns a serialized Context.
func (c *Context[R, S]) Marshal() ([]byte, error) {
	return c.MarshalAppend(nil)
}

// MarshalAppend appends a serialized Context to the provided buffer.
func (c *Context[R, S]) MarshalAppend(b []byte) ([]byte, error) {
	s := serde.Serialize(&serializedCoroutine{
		entry:  c.entry,
		stack:  c.Stack,
		resume: c.resume,
	})
	return append(b, s...), nil
}

type serializedCoroutine struct {
	entry  func()
	stack  Stack
	resume bool
}

// Unmarshal deserializes a Context from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// context.
func (c *Context[R, S]) Unmarshal(b []byte) (int, error) {
	start := len(b)
	v, b := serde.Deserialize(b)
	s := v.(*serializedCoroutine)
	c.entry = s.entry
	c.Stack = s.stack
	c.resume = s.resume
	sn := start - len(b)
	return sn, nil
}

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

// Storage is a sparse collection of Serializable objects.
type Storage struct {
	// This is private so that the data structure is allowed to switch
	// the in-memory representation dynamically (e.g. a map[int]Serializable
	// may be more efficient for very sparse maps).
	objects []any
}

// NewStorage creates a Storage.
func NewStorage(objects []any) Storage {
	return Storage{objects: objects}
}

// Has is true if an object is defined for a specific index.
func (v *Storage) Has(i int) bool {
	return i >= 0 && i < len(v.objects)
}

// Get gets the object for a specific index.
func (v *Storage) Get(i int) any {
	if !v.Has(i) {
		panic("missing object " + strconv.Itoa(i))
	}
	return v.objects[i]
}

// Delete gets the object for a specific index.
func (v *Storage) Delete(i int) {
	if !v.Has(i) {
		panic("missing object " + strconv.Itoa(i))
	}
	v.objects[i] = nil
}

// Set sets the object for a specific index.
func (v *Storage) Set(i int, value any) {
	if n := i + 1; n > len(v.objects) {
		v.objects = slices.Grow(v.objects, n-len(v.objects))
		v.objects = v.objects[:n]
	}
	v.objects[i] = value
}

func (v *Storage) shrink() {
	i := len(v.objects) - 1
	for i >= 0 && v.objects[i] == nil {
		i--
	}
	v.objects = v.objects[:i+1]
}
