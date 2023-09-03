package coroutine_test

import (
	"reflect"
	"testing"

	"github.com/stealthrocket/coroutine"
)

// TestCoroutine tests manually constructed coroutines.
func TestCoroutine(t *testing.T) {
	for _, test := range []struct {
		name   string
		coro   func(*coroutine.Context, coroutine.Int)
		arg    coroutine.Int
		yields []coroutine.Serializable
	}{
		{
			name: "identity",
			coro: identity,
			arg:  coroutine.Int(11),
			yields: []coroutine.Serializable{
				coroutine.Int(11),
			},
		},

		{
			name: "square generator",
			coro: squareGenerator,
			arg:  coroutine.Int(4),
			yields: []coroutine.Serializable{
				coroutine.Int(1),
				coroutine.Int(4),
				coroutine.Int(9),
				coroutine.Int(16),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := new(coroutine.Context)

			var yield int
			for {
				c.Send(nil)
				returned := (func() bool {
					defer func() {
						if c.Unwinding() {
							recover()
						}
					}()
					c.Stack.FP = -1
					test.coro(c, test.arg)
					return true
				})()
				if returned {
					break
				}
				if yield == len(test.yields) {
					t.Errorf("unexpected yield from coroutine")
					break
				}
				actual := c.Recv()
				expect := test.yields[yield]

				if !reflect.DeepEqual(actual, expect) {
					t.Fatalf("coroutine yielded incorrect value at index %d: got %#v, expect %#v", yield, actual, expect)
				}

				yield++

				// Serialize => deserialize the context before resuming.
				b, err := c.MarshalAppend(nil)
				if err != nil {
					t.Fatal(err)
				}
				var reconstructed coroutine.Context
				if n, err := reconstructed.Unmarshal(b); err != nil {
					t.Fatal(err)
				} else if n != len(b) {
					t.Fatal("invalid number of bytes read when reconstructing context")
				}
				*c = reconstructed
			}
			if yield < len(test.yields) {
				t.Errorf("coroutine did not yield the correct number of times: got %d, expect %d", yield, len(test.yields))
			}
		})
	}
}

func identity(c *coroutine.Context, n coroutine.Int) {
	// func identity(n int) {
	//   yield(n)
	// }
	c.Push()
	c.Yield(coroutine.Int(n))
	c.Pop()
}

func squareGenerator(c *coroutine.Context, n coroutine.Int) {
	// func squareGenerator(n int) {
	//   for i := 1; i <= n; i++ {
	//     yield(i * i)
	//   }
	// }

	// new stack frame, or reuse current frame on resume
	frame := c.Push()

	// variable declaration
	var (
		i coroutine.Int
	)

	// state restoration
	switch frame.IP {
	case 1:
		n = frame.Get(0).(coroutine.Int)
		i = frame.Get(1).(coroutine.Int)
	}

	// state capture
	defer func() {
		if c.Unwinding() {
			switch frame.IP {
			case 1:
				frame.Set(0, coroutine.Int(n))
				frame.Set(1, coroutine.Int(i))
			}
		}
	}()

	// coroutine state machine
	switch frame.IP {
	case 0:
		i = 1
		frame.IP = 1
		fallthrough
	case 1:
		for i <= n {
			c.Yield(i * i)
			i++
		}
	}

	// pop stack frame now that the function call completed
	c.Pop()
}
