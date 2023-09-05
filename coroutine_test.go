//go:build !durable

package coroutine_test

import (
	"slices"
	"testing"

	"github.com/stealthrocket/coroutine"
)

// TestCoroutine tests manually constructed coroutines.
func TestCoroutine(t *testing.T) {
	for _, test := range []struct {
		name   string
		coro   func(int)
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
			g := coroutine.New(func(*coroutine.Context[int, any]) {
				test.coro(test.arg)
			})

			var yields []int
			coroutine.Run(g, func(v int) any {
				yields = append(yields, v)
				return nil
			})

			if !slices.Equal(yields, test.yields) {
				t.Fatalf("coroutine yielded incorrect yield values: got %#v, expect %#v", yields, test.yields)
			}
		})
	}
}

func identity(n int) {
	coroutine.Yield[int, any](n)
}

func squareGenerator(n int) {
	for i := 1; i <= n; i++ {
		coroutine.Yield[int, any](i * i)
	}
}
