package main

import (
	"testing"

	"github.com/stealthrocket/coroutine"
	"github.com/stealthrocket/coroutine/coroc/testdata"
)

func TestCoroutine(t *testing.T) {
	for _, test := range []struct {
		name   string
		coro   func(int)
		arg    int
		yields []int
	}{
		{
			name:   "identity",
			coro:   testdata.Identity,
			arg:    11,
			yields: []int{11},
		},
		{
			name:   "square generator",
			coro:   testdata.SquareGenerator,
			arg:    4,
			yields: []int{1, 4, 9, 16},
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
