package coroutine

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestGLS(t *testing.T) {
	ch := make(chan any)
	key := uintptr(0)
	val := any(42)

	go with(&key, val, func() {
		ch <- load(key)
	})

	if v := <-ch; !reflect.DeepEqual(v, val) {
		t.Errorf("wrong value for key=%v: %v", key, *(*[2]unsafe.Pointer)(unsafe.Pointer(&v)))
	}
}

func BenchmarkGLS(b *testing.B) {
	key := uintptr(0)
	val := any(42)
	with(&key, val, func() {
		for i := 0; i < b.N; i++ {
			load(key)
		}
	})
}
