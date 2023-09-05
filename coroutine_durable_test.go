//go:build durable

package coroutine_test

import (
	"testing"

	"github.com/stealthrocket/coroutine"
)

// TestCoroutine tests manually constructed coroutines.
func TestCoroutine(t *testing.T) {
	for _, test := range []struct {
		name   string
		coro   func(*coroutine.Context[int, any], int)
		arg    int
		yields []int
	}{
		{
			name:   "identity",
			coro:   identity,
			arg:    11,
			yields: []int{11},
		},

		{
			name:   "square generator",
			coro:   squareGenerator,
			arg:    4,
			yields: []int{1, 4, 9, 16},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := coroutine.New(func(c *coroutine.Context[int, any]) {
				test.coro(c, test.arg)
			})

			var yield int
			for g.Next() {
				if yield == len(test.yields) {
					t.Errorf("unexpected yield from coroutine")
					break
				}
				actual := g.Recv()
				expect := test.yields[yield]

				if actual != expect {
					t.Fatalf("coroutine yielded incorrect value at index %d: got %#v, expect %#v", yield, actual, expect)
				}

				yield++

				// Serialize => deserialize the context before resuming.
				c := g.Context()
				b, err := c.MarshalAppend(nil)
				if err != nil {
					t.Fatal(err)
				}
				var reconstructed coroutine.Context[int, any]
				if n, err := reconstructed.Unmarshal(b); err != nil {
					t.Fatal(err)
				} else if n != len(b) {
					t.Fatal("invalid number of bytes read when reconstructing context")
				}
				f := c.Func
				*c = reconstructed
				// TODO: the context reconstruction needs to capture the
				// coroutine entry point.
				//
				// https://www.notion.so/stealthrocket/Durable-Coroutines-1487e78403804b5f871cf37275a55cc8?pvs=4#395d316dc79e432ca58dd59df9f561f0
				c.Func = f
			}
			if yield < len(test.yields) {
				t.Errorf("coroutine did not yield the correct number of times: got %d, expect %d", yield, len(test.yields))
			}
		})
	}
}

func identity(c *coroutine.Context[int, any], n int) {
	// func identity(n int) {
	//   yield(n)
	// }
	c.Push()
	c.Yield(n)
	c.Pop()
}

func squareGenerator(c *coroutine.Context[int, any], n int) {
	// func squareGenerator(n int) {
	//   for i := 1; i <= n; i++ {
	//     yield(i * i)
	//   }
	// }

	// new stack frame, or reuse current frame on resume
	frame := c.Push()

	// variable declaration
	var (
		i int
	)

	// state restoration
	switch frame.IP {
	case 1:
		n = int(frame.Get(0).(coroutine.Int))
		i = int(frame.Get(1).(coroutine.Int))
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
