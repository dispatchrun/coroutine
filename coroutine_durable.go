//go:build durable

package coroutine

import (
	"sync"

	"github.com/stealthrocket/coroutine/types"
)

// Durable is a constant which takes the values true or false depending on
// whether the program is built with the "durable" tag.
const Durable = true

// New creates a new coroutine which executes f as entry point.
func New[R, S any](f func()) Coroutine[R, S] {
	c := Coroutine[R, S]{
		ctx: &Context[R, S]{
			context: context{entry: f},
		},
	}
	c.ctx.cond.L = &c.ctx.lock
	return c
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

// Marshal returns a serialized Context.
func (c *Context[R, S]) Marshal() ([]byte, error) {
	return types.Serialize(&serializedCoroutine{
		entry:  c.entry,
		stack:  c.Stack,
		resume: c.resume,
	}), nil
}

// Unmarshal deserializes a Context from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// context.
func (c *Context[R, S]) Unmarshal(b []byte) (int, error) {
	start := len(b)
	v, b := types.Deserialize(b)
	s := v.(*serializedCoroutine)
	c.entry = s.entry
	c.Stack = s.stack
	c.resume = s.resume
	sn := start - len(b)
	return sn, nil
}

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

	c.ctx.lock.Lock()

	go with(&gctx, c.ctx, func() {
		defer func() {
			switch v := recover().(type) {
			case nil:
			case unwind:
			default:
				// TODO: can we figure out a way to know when we are unwinding the
				// stack and only recover then so we don't alter the panic stack?
				panic(v)
			}

			c.ctx.cond.Signal()
		}()

		c.ctx.Stack.FP = -1
		c.ctx.entry()
	})

	c.ctx.cond.Wait()
	c.ctx.lock.Unlock()

	if c.ctx.Unwinding() {
		stop := c.ctx.stop
		c.ctx.done, hasNext = stop, !stop
	} else {
		c.ctx.done = true
	}

	return hasNext
}

type context struct {
	// Entry point of the coroutine, this is captured so the associated
	// generator can call into the coroutine to start or resume it at the
	// last yield point.
	entry func()

	lock sync.Mutex
	cond sync.Cond

	Stack
}

type unwind struct{}

// Unwinding returns true if the coroutine is currently unwinding its stack.
func (c *Context[R, S]) Unwinding() bool {
	return c.resume
}
