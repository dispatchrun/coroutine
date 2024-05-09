//go:build durable

package coroutine

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/dispatchrun/coroutine/types"
)

// Durable is a constant which takes the values true or false depending on
// whether the program is built with the "durable" tag.
const Durable = true

// New creates a new coroutine which executes f as entry point.
//
//go:noinline
func New[R, S any](f func()) Coroutine[R, S] {
	// The function has the go:noinline tag because we want to ensure that the
	// context will be allocated on the heap. If the context remains allocated
	// on the stack it might escape when returned by a call to LoadContext that
	// the compiler cannot track.
	return Coroutine[R, S]{
		ctx: &Context[R, S]{
			context: context[R]{entry: f},
		},
	}
}

// New creates a new coroutine which executes f as entry point.
//
//go:noinline
func NewWithReturn[R, S any](f func() R) Coroutine[R, S] {
	// The function has the go:noinline tag because we want to ensure that the
	// context will be allocated on the heap. If the context remains allocated
	// on the stack it might escape when returned by a call to LoadContext that
	// the compiler cannot track.
	return Coroutine[R, S]{
		ctx: &Context[R, S]{
			context: context[R]{entryR: f},
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

type serializedCoroutine[R any] struct {
	entry  func()
	entryR func() R
	stack  Stack
	resume bool
}

// Marshal returns a serialized Context.
func (c *Context[R, S]) Marshal() ([]byte, error) {
	return types.Serialize(&serializedCoroutine[R]{
		entry:  c.entry,
		entryR: c.entryR,
		stack:  c.Stack,
		resume: c.resume,
	})
}

// Unmarshal deserializes a Context from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// context.
func (c *Context[R, S]) Unmarshal(b []byte) error {
	v, err := types.Deserialize(b)
	if err != nil {
		if errors.Is(err, types.ErrBuildIDMismatch) {
			err = ErrInvalidState
		}
		return err
	}
	s := v.(*serializedCoroutine[R])
	c.entry = s.entry
	c.entryR = s.entryR
	c.Stack = s.stack
	c.resume = s.resume
	return nil
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

	execute(c.ctx, func() {
		defer func() {
			switch v := recover().(type) {
			case nil:
			case unwind:
			default:
				// TODO: can we figure out a way to know when we are unwinding the
				// stack and only recover then so we don't alter the panic stack?
				panic(v)
			}

			if c.ctx.Unwinding() {
				stop := c.ctx.stop
				c.ctx.done, hasNext = stop, !stop
			} else {
				c.ctx.done = true
			}
		}()

		c.ctx.Stack.FP = -1
		if c.ctx.entry != nil {
			c.ctx.entry()
		} else {
			c.ctx.result = c.ctx.entryR()
		}
	})

	return hasNext
}

type context[R any] struct {
	// Entry point of the coroutine, this is captured so the associated
	// generator can call into the coroutine to start or resume it at the
	// last yield point.
	//
	// The raw func (via New) and func returning R (via NewWithReturn)
	// are stored separately to work around a limitation with the compiler.
	// In volatile mode we only store the latter, and support the former
	// by creating a closure that calls the func() and returns the zero
	// value R. The compiler does not yet support compiling generic
	// functions so the strategy doesn't work in durable mode.
	entry  func()
	entryR func() R
	Stack
}

type unwind struct{}

// Unwinding returns true if the coroutine is currently unwinding its stack.
func (c *Context[R, S]) Unwinding() bool {
	return c.resume
}

// The load function returns the value passed as first argument to the call to
// execute that started the coroutine.
func load() any {
	g := getg()
	endOfG := (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(g)) + sizeOfG))
	return *(*any)(unsafe.Pointer(g.stack.hi - *endOfG))
}

// The execute function is the entry point of coroutines, it pushes the
// coroutine context in v to the stack, registering it to be retrieved by
// calling load, and invokes f as the entry point.
//
// The coroutine continues execution until a yield point is reached or until
// the function passed as entry point returns.
//
//go:nosplit
//go:noinline
func execute(v any, f func()) {
	g := getg()
	p := unsafe.Pointer(&v)

	// On 64 bits architectures, the g struct is 408 bytes, but allocated on the
	// heap it uses a class size of 416 bytes, which means that we have 8 bytes
	// unused at the end of the struct where we can store the offset.
	//
	// TODO: since we are recompiling the code, we should inject a field in the
	// runtime's g struct instead of relying on finding spare space.
	endOfG := (*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(g)) + sizeOfG))

	// We reload the g from thread local storage so we don't have to store it on
	// this stack frame, it will be the same value as the one we've seen at the
	// beginning of the function.
	//
	// TODO: this code path does not get executed if a panic occurs, we either
	// need to disallow the coroutine entry point to panic (i.e., volatile does
	// this implicitly since we use a goroutine), or we need to figure out how
	// to declare defers in assembly code.
	prevOffset := *endOfG
	defer func() { *endOfG = prevOffset }()

	*endOfG = g.stack.hi - uintptr(p)
	f()

	runtime.KeepAlive(v)
}
