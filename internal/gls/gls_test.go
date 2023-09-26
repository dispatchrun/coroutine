package gls

import "testing"

func TestGLS(t *testing.T) {
	c := make(chan int)

	f := func(n int) {
		defer close(c)
		Context().Store(n)

		load := func() int {
			v, _ := Context().Load().(int)
			return v
		}

		c <- load()
		Context().Clear()
		c <- load()
	}

	go f(42)

	if v, ok := <-c; !ok || v != 42 {
		t.Errorf("unexpected first value: want=(42,true) got=(%v,%v)", v, ok)
	}
	if v, ok := <-c; !ok || v != 0 {
		t.Errorf("unexpected second value: want=(0,true) got=(%v,%v)", v, ok)
	}
	if v, ok := <-c; ok {
		t.Errorf("too many values received: want=(0,false) got=(%v,%v)", v, ok)
	}
}

func BenchmarkGLS(b *testing.B) {
	b.Run("getg", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = getg()
			}
		})
	})

	b.Run("loadContext", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Context().Load()
			}
		})
	})

	b.Run("storeContext", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				Context().Store(42)
			}
		})
	})

	b.Run("clearContext", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				Context().Clear()
			}
		})
	})
}
