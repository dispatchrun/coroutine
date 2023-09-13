package main

import (
	"slices"
	"testing"

	"github.com/stealthrocket/coroutine"
	. "github.com/stealthrocket/coroutine/coroc/testdata"
)

func TestCoroutineYield(t *testing.T) {
	for _, test := range []struct {
		name   string
		coro   func(int)
		arg    int
		yields []int
	}{
		{
			name:   "identity",
			coro:   Identity,
			arg:    11,
			yields: []int{11},
		},
		{
			name:   "square generator",
			coro:   SquareGenerator,
			arg:    4,
			yields: []int{1, 4, 9, 16},
		},
		{
			name:   "square generator twice",
			coro:   SquareGeneratorTwice,
			arg:    4,
			yields: []int{1, 4, 9, 16, 1, 4, 9, 16},
		},
		{
			name:   "even square generator",
			coro:   EvenSquareGenerator,
			arg:    6,
			yields: []int{4, 16, 36},
		},
		{
			name:   "nested loops",
			coro:   NestedLoops,
			arg:    3,
			yields: []int{1, 2, 3, 2, 4, 6, 3, 6, 9, 2, 4, 6, 4, 8, 12, 6, 12, 18, 3, 6, 9, 6, 12, 18, 9, 18, 27},
		},
		{
			name:   "fizz buzz (1)",
			coro:   FizzBuzzIfGenerator,
			arg:    20,
			yields: []int{1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz, 16, 17, Fizz, 19, Buzz},
		},
		{
			name:   "fizz buzz (2)",
			coro:   FizzBuzzSwitchGenerator,
			arg:    20,
			yields: []int{1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz, 16, 17, Fizz, 19, Buzz},
		},
		{
			name:   "shadowing",
			coro:   Shadowing,
			yields: []int{0, 1, 0, 1, 2, 0, 2, 1, 0, 2, 1, 0, 1, 0, 13, 12, 11, 4, 2, 1, 2, 1},
		},
		{
			name:   "range over slice indices",
			coro:   RangeSliceIndexGenerator,
			yields: []int{0, 1, 2},
		},
		{
			name:   "range over array indices and values",
			coro:   RangeArrayIndexValueGenerator,
			yields: []int{0, 10, 1, 20, 2, 30},
		},
		{
			name:   "type switching",
			coro:   TypeSwitchingGenerator,
			yields: []int{1, 10, 2, 20, 4, 30, 8, 40},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := coroutine.New[int, any](func() {
				test.coro(test.arg)
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

				// If supported, serialize => deserialize the context
				// before resuming.
				assertSerializable(t, g)
			}
			if yield < len(test.yields) {
				t.Errorf("coroutine did not yield the correct number of times: got %d, expect %d", yield, len(test.yields))
			}
		})
	}
}

func TestCoroutineStop(t *testing.T) {
	coro := coroutine.New[int, any](func() {
		SquareGenerator(4)
	})

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
