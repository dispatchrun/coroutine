//go:build durable

package coroutine

import (
	"reflect"
	"testing"
)

func TestLocalStorageStack(t *testing.T) {
	assert := func(want any) {
		if got := load(); !reflect.DeepEqual(got, want) {
			t.Helper()
			t.Errorf("wrong coroutine value: want=%#v got=%#v", want, got)
		}
	}

	test := func(v any, f func()) {
		execute(v, func() {
			assert(v)
			f()
			assert(v)
		})
	}

	ok := false
	test(1, func() {
		test(2, func() {
			test(3, func() {
				ok = true
			})
		})
	})

	if !ok {
		t.Error("test did not run")
	}
}
