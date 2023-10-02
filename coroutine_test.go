package coroutine

import (
	"reflect"
	"testing"
)

func TestLocalStorage(t *testing.T) {
	with(42, func() {
		if v := load(); !reflect.DeepEqual(v, 42) {
			t.Errorf("wrong value: %v", v)
		}
	})
}

//go:noinline
func weirdLoop(n int, f func()) int {
	if n == 0 {
		f()
		return 0
	} else {
		return weirdLoop(n-1, f) + 1 // just in case Go ever implements tail recursion
	}
}

func TestLocalStorageGrowStack(t *testing.T) {
	with("hello", func() {
		weirdLoop(100e3, func() {
			if v := load(); v != "hello" {
				t.Errorf("wrong value: %v", v)
			}
		})
	})
}

func BenchmarkLocalStorage(b *testing.B) {
	with("hello", func() {
		for i := 0; i < b.N; i++ {
			load()
		}
	})
}
