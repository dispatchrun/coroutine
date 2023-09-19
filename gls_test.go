package coroutine

import "testing"

func TestGLS(t *testing.T) {
	c := make(chan int)

	f := func(n int) {
		defer close(c)
		storeContext(getg(), n)

		load := func() int {
			v, _ := loadContext(getg()).(int)
			return v
		}

		c <- load()
		clearContext(getg())
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
			g := getg()
			for pb.Next() {
				_ = loadContext(g)
			}
		})
	})

	b.Run("storeContext", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			g := getg()
			for pb.Next() {
				storeContext(g, 42)
			}
		})
	})

	b.Run("clearContext", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			g := getg()
			for pb.Next() {
				clearContext(g)
			}
		})
	})

	b.Run("store load clear", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			g := getg()
			for pb.Next() {
				storeContext(g, 42)
				loadContext(g)
				clearContext(g)
			}
		})
	})
}
