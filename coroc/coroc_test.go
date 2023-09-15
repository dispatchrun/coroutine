package main

import (
	"slices"
	"testing"

	"github.com/stealthrocket/coroutine"
	. "github.com/stealthrocket/coroutine/coroc/testdata"
)

type identity struct{ arg int }

func (c identity) Call() { Identity(c.arg) }

type squareGenerator struct{ arg int }

func (c squareGenerator) Call() { SquareGenerator(c.arg) }

type squareGeneratorTwice struct{ arg int }

func (c squareGeneratorTwice) Call() { SquareGeneratorTwice(c.arg) }

type evenSquareGenerator struct{ arg int }

func (c evenSquareGenerator) Call() { EvenSquareGenerator(c.arg) }

type nestedLoops struct{ arg int }

func (c nestedLoops) Call() { NestedLoops(c.arg) }

type fizzBuzzIfGenerator struct{ arg int }

func (c fizzBuzzIfGenerator) Call() { FizzBuzzIfGenerator(c.arg) }

type fizzBuzzSwitchGenerator struct{ arg int }

func (c fizzBuzzSwitchGenerator) Call() { FizzBuzzSwitchGenerator(c.arg) }

type shadowing struct{ arg int }

func (c shadowing) Call() { Shadowing(c.arg) }

type rangeSliceIndexGenerator struct{ arg int }

func (c rangeSliceIndexGenerator) Call() { RangeSliceIndexGenerator(c.arg) }

type rangeArrayIndexValueGenerator struct{ arg int }

func (c rangeArrayIndexValueGenerator) Call() { RangeArrayIndexValueGenerator(c.arg) }

type typeSwitchingGenerator struct{ arg int }

func (c typeSwitchingGenerator) Call() { TypeSwitchingGenerator(c.arg) }

type loopBreakAndContinue struct{ arg int }

func (c loopBreakAndContinue) Call() { LoopBreakAndContinue(c.arg) }

func TestCoroutineYield(t *testing.T) {
	for _, test := range []struct {
		name   string
		coro   coroutine.Closure
		yields []int
	}{
		{
			name:   "identity",
			coro:   identity{arg: 11},
			yields: []int{11},
		},

		{
			name:   "square generator",
			coro:   squareGenerator{arg: 4},
			yields: []int{1, 4, 9, 16},
		},

		{
			name:   "square generator twice",
			coro:   squareGeneratorTwice{arg: 4},
			yields: []int{1, 4, 9, 16, 1, 4, 9, 16},
		},

		{
			name:   "even square generator",
			coro:   evenSquareGenerator{arg: 6},
			yields: []int{4, 16, 36},
		},

		{
			name:   "nested loops",
			coro:   nestedLoops{arg: 3},
			yields: []int{1, 2, 3, 2, 4, 6, 3, 6, 9, 2, 4, 6, 4, 8, 12, 6, 12, 18, 3, 6, 9, 6, 12, 18, 9, 18, 27},
		},

		{
			name:   "fizz buzz (1)",
			coro:   fizzBuzzIfGenerator{arg: 20},
			yields: []int{1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz, 16, 17, Fizz, 19, Buzz},
		},

		{
			name:   "fizz buzz (2)",
			coro:   fizzBuzzSwitchGenerator{arg: 20},
			yields: []int{1, 2, Fizz, 4, Buzz, Fizz, 7, 8, Fizz, Buzz, 11, Fizz, 13, 14, FizzBuzz, 16, 17, Fizz, 19, Buzz},
		},

		{
			name:   "shadowing",
			coro:   shadowing{arg: 0},
			yields: []int{0, 1, 0, 1, 2, 0, 2, 1, 0, 2, 1, 0, 1, 0, 13, 12, 11, 4, 2, 1, 2, 1},
		},

		{
			name:   "range over slice indices",
			coro:   rangeSliceIndexGenerator{arg: 0},
			yields: []int{0, 1, 2},
		},

		{
			name:   "range over array indices and values",
			coro:   rangeArrayIndexValueGenerator{arg: 0},
			yields: []int{0, 10, 1, 20, 2, 30},
		},

		{
			name:   "type switching",
			coro:   typeSwitchingGenerator{},
			yields: []int{1, 10, 2, 20, 4, 30, 8, 40},
		},

		{
			name:   "loop break and continue",
			coro:   loopBreakAndContinue{},
			yields: []int{1, 3, 5},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
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
				assertSerializable(t, g)
			}
			if yield < len(test.yields) {
				t.Errorf("coroutine did not yield the correct number of times: got %d, expect %d", yield, len(test.yields))
			}
		})
	}
}

func TestCoroutineStop(t *testing.T) {
	coro := coroutine.New[int, any](squareGenerator{arg: 4})

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
