package compiler

import (
	"slices"
	"testing"

	"github.com/stealthrocket/coroutine"
	. "github.com/stealthrocket/coroutine/compiler/testdata"
	"github.com/stealthrocket/coroutine/types"
)

func init() {
	// Breaks if the compiler did not retain simple top-level functions in the
	// output file.
	SomeFunctionThatShouldExistInTheCompiledFile()
}

func TestCoroutineYield(t *testing.T) {
	tests := []struct {
		name   string
		coro   func()
		yields []int
		skip   bool
	}{
		{
			name:   "identity",
			coro:   func() { Identity(11) },
			yields: []int{11},
		},

		{
			name:   "square generator",
			coro:   func() { SquareGenerator(4) },
			yields: []int{1, 4, 9, 16},
		},

		{
			name:   "square generator twice",
			coro:   func() { SquareGeneratorTwice(4) },
			yields: []int{1, 4, 9, 16, 1, 4, 9, 16},
		},

		{
			name:   "square generator twice loop",
			coro:   func() { SquareGeneratorTwiceLoop(4) },
			yields: []int{1, 4, 9, 16, 1, 4, 9, 16},
		},

		{
			name:   "even square generator",
			coro:   func() { EvenSquareGenerator(6) },
			yields: []int{4, 16, 36},
		},

		{
			name:   "nested loops",
			coro:   func() { NestedLoops(3) },
			yields: []int{1, 2, 3, 2, 4, 6, 3, 6, 9, 2, 4, 6, 4, 8, 12, 6, 12, 18, 3, 6, 9, 6, 12, 18, 9, 18, 27},
		},

		{
			name:   "fizz buzz (1)",
			coro:   func() { FizzBuzzIfGenerator(20) },
			yields: []int{1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz, 16, 17, Fizz, 19, Buzz},
		},

		{
			name:   "fizz buzz (2)",
			coro:   func() { FizzBuzzSwitchGenerator(20) },
			yields: []int{1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz, 16, 17, Fizz, 19, Buzz},
		},

		{
			name:   "shadowing",
			coro:   func() { Shadowing(0) },
			yields: []int{0, 1, 0, 1, 2, 0, 2, 1, 0, 2, 1, 0, 1, 0, 13, 12, 11, 4, 2, 1, 2, 1},
		},

		{
			name:   "range over slice indices",
			coro:   func() { RangeSliceIndexGenerator(0) },
			yields: []int{0, 1, 2},
		},

		{
			name:   "range over array indices and values",
			coro:   func() { RangeArrayIndexValueGenerator(0) },
			yields: []int{0, 10, 1, 20, 2, 30},
		},

		{
			name:   "range over deferred function",
			coro:   func() { RangeYieldAndDeferAssign(5) },
			yields: []int{0, 1, 2, 3, 4},
		},

		{
			name:   "type switching",
			coro:   func() { TypeSwitchingGenerator(0) },
			yields: []int{1, 10, 2, 20, 4, 30, 8, 40},
		},

		{
			name:   "loop break and continue",
			coro:   func() { LoopBreakAndContinue(0) },
			yields: []int{1, 3, 5, 0, 1, 0, 1},
		},

		{
			name:   "range over maps",
			coro:   func() { RangeOverMaps(5) },
			yields: []int{0, 5, 5, 50, 5, 4, 3, 2, 1, 0},
		},

		{
			name:   "range over function",
			coro:   func() { Range(10, Double) },
			yields: []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18},
		},

		{
			name:   "reverse range over closure capturing by value",
			coro:   func() { RangeReverseClosureCaptureByValue(10) },
			yields: []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		},

		{
			name:   "range over anonymous function",
			coro:   func() { RangeTriple(4) },
			yields: []int{0, 3, 6, 9},
		},

		{
			name:   "range over anonymous function value",
			coro:   func() { RangeTripleFuncValue(4) },
			yields: []int{0, 3, 6, 9},
		},

		{
			name:   "range over closure capturing values",
			coro:   Range10ClosureCapturingValues,
			yields: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},

		{
			name:   "range over closure capturing pointers",
			coro:   Range10ClosureCapturingPointers,
			yields: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},

		{
			name:   "range over closure capturing heterogenous values",
			coro:   Range10ClosureHeterogenousCapture,
			yields: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},

		{
			name:   "range with heterogenous values",
			coro:   Range10Heterogenous,
			yields: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},

		{
			name:   "select",
			coro:   func() { Select(8) },
			yields: []int{-1, 0, 0, 1, 10, 2, 20, 3, 30, 4, 40, 50, 0, 1, 2},
			// TODO: re-enable test once either chan serialization is supported,
			//  or the desugaring pass skips statements that cannot yield (which
			//  will reduce temporary vars and avoid the need to deser type chan)
			skip: true,
		},

		{
			name: "yielding expression desugaring",
			coro: func() { YieldingExpressionDesugaring() },
			yields: []int{
				-1, 1, -2, 2, -3, 3, -4, 4, -5, 5, 50, // if
				-6, 6, -8, 8, 70, -8, 8, 70, -8, 8, // for
				-9, 9, -10, 10, -11, 11, -12, 12, -13, 13, // switch
				-15, 15, 150, // type switch
			},
		},

		{
			name:   "yield imported type time.Duration",
			coro:   YieldingDurations,
			yields: []int{100, 101, 102, 103, 104, 105, 106, 107, 108, 109},
		},

		{
			name:   "methods",
			coro:   func() { var s MethodGeneratorState; s.MethodGenerator(5) },
			yields: []int{0, 1, 2, 3, 4, 5},
		},
	}

	// This emulates the installation of function type information by the
	// compiler because we are not doing codegen for the test files in this
	// package.
	for _, test := range tests {
		a := types.FuncAddr(test.coro)
		f := types.FuncByAddr(a)
		types.RegisterFunc[func()](f.Name)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skip("test is disabled")
			}

			g := coroutine.New[int, any](test.coro)

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

				// If supported, serialize => deserialize the context
				// before resuming.
				b, err := g.Context().Marshal()
				if err != nil {
					if err == coroutine.ErrNotDurable {
						continue
					}
					t.Fatal(err)
				}

				reconstructed := coroutine.New[int, any](test.coro)
				if n, err := reconstructed.Context().Unmarshal(b); err != nil {
					t.Fatal(err)
				} else if n != len(b) {
					t.Fatal("invalid number of bytes read when reconstructing context")
				}
				g = reconstructed
			}
			if yield < len(test.yields) {
				t.Errorf("coroutine did not yield the correct number of times: got %d, expect %d", yield, len(test.yields))
			}
		})
	}
}

func TestCoroutineStop(t *testing.T) {
	coro := coroutine.New[int, any](func() { SquareGenerator(4) })

	values := []int{}
	coroutine.Run(coro, func(v int) any {
		if v > 10 {
			coro.Stop()
		} else {
			values = append(values, v)
		}
		return nil
	})

	if !slices.Equal(values, []int{1, 4, 9}) {
		t.Errorf("wrong values yield by coroutine: %#v", values)
	}
}
