package coroutine_test

import (
	"reflect"
	"testing"

	"github.com/stealthrocket/coroutine"
)

// TestCoroutine tests manually constructed coroutines.
func TestCoroutine(t *testing.T) {
	for _, test := range []struct {
		name    string
		coro    coroutine.Coroutine
		args    []coroutine.Serializable
		yields  [][]coroutine.Serializable
		returns []coroutine.Serializable
	}{
		{
			name:    "identity",
			coro:    identity,
			args:    []coroutine.Serializable{coroutine.Int(11)},
			returns: []coroutine.Serializable{coroutine.Int(11)},
		},
		{
			name: "square generator",
			coro: squareGenerator,
			args: []coroutine.Serializable{coroutine.Int(4)},
			yields: [][]coroutine.Serializable{
				{coroutine.Int(1)},
				{coroutine.Int(4)},
				{coroutine.Int(9)},
				{coroutine.Int(16)},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := coroutine.NewContext(test.args...)

			var yield int
			for {
				returned := c.Call(test.coro)
				if returned {
					break
				}
				if yield == len(test.yields) {
					t.Errorf("unexpected yield from coroutine")
					break
				}
				expects := test.yields[yield]

				topFrame := c.Stack.Top()
				for i, expect := range expects {
					if !topFrame.Has(i) {
						t.Fatalf("coroutine did not yield an object at index %d: expect %#v", i, expect)
					} else if actual := topFrame.Get(i); !reflect.DeepEqual(actual, expect) {
						t.Fatalf("coroutine yielded incorrect value at index %d: got %#v, expect %#v", i, actual, expect)
					}
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

			f := c.Stack.Frames[0]
			offset := len(test.args)
			for i, expect := range test.returns {
				if !f.Has(offset + i) {
					t.Errorf("coroutine did not return value at index %d: expect %#v", i, expect)
				} else if actual := f.Get(offset + i); !reflect.DeepEqual(actual, expect) {
					t.Errorf("coroutine returned incorrect value at index %d: got %#v, expect %#v", i, actual, expect)
				}
			}
		})
	}
}

func identity(c *coroutine.Context) {
	// func identity(n int) int {
	//   return n
	// }

	frame := &c.Stack.Frames[c.Stack.FP]

	frame.Set(1, frame.Get(0))
}

func squareGenerator(c *coroutine.Context) {
	// func squareGenerator(n int) {
	//   for i := 1; i <= n; i++ {
	//     yield(i * i)
	//   }
	// }

	stack := &c.Stack
	frame := &stack.Frames[stack.FP]

	n := int(frame.Get(0).(coroutine.Int))

	var i int
	switch frame.IP {
	case 0:
		i = 1
		frame.Set(1, coroutine.Int(i))
		frame.IP = 1
		fallthrough
	case 1:
		i = int(frame.Get(1).(coroutine.Int))

		for i <= n {
			stack.Push(func() coroutine.Frame {
				arg0 := i * i
				return coroutine.Frame{
					Storage: coroutine.NewStorage([]coroutine.Serializable{
						coroutine.Int(arg0),
					}),
				}
			})
			yieldInt(c)
			stack.Pop()

			i++
			frame.Set(1, coroutine.Int(i))
		}
	}
}

func yieldInt(c *coroutine.Context) {
	if frame := c.Top(); !frame.Resume {
		frame.Resume = true
		coroutine.Unwind()
	}
}
