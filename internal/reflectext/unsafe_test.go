package reflectext_test

import (
	"reflect"
	"testing"

	"github.com/dispatchrun/coroutine/internal/reflectext"
)

const size = 256 // 256 bytes total
const width = 8  // 8 bytes per slot (enough for a (u)int64)
const count = size / width

func TestInternedValue(t *testing.T) {
	testInternedInteger[int](t)
	testInternedInteger[int8](t)
	testInternedInteger[int16](t)
	testInternedInteger[int32](t)
	testInternedInteger[int64](t)

	testInternedInteger[uint](t)
	testInternedInteger[uint8](t)
	testInternedInteger[uint16](t)
	testInternedInteger[uint32](t)
	testInternedInteger[uint64](t)
	testInternedInteger[uintptr](t)

	// Structs and arrays with one element occupy
	// the same space as the element. As such, certain
	// elements are interned in the same way.
	assertInterned(t, struct{ v int }{0}, 0)
	assertInterned(t, [1]int{0}, 0)

	assertInterned(t, false, 0)
	assertInterned(t, true, width)
	assertInterned(t, struct{ v bool }{false}, 0)
	assertInterned(t, [1]bool{true}, width)

	assertNotInterned(t, float64(0))
	assertNotInterned(t, "")
}

func testInternedInteger[T integer](t *testing.T) {
	for i := T(0); i < count; i++ {
		var boxed any = i
		assertInterned(t, boxed, reflectext.InternedValueOffset(i)*width)
	}
	assertNotInterned(t, T(count))
	assertNotInterned(t, T(0)-1)
}

func assertInterned(t *testing.T, value any, want reflectext.InternedValueOffset) {
	v := reflect.ValueOf(&value).Elem()
	p := reflectext.InterfaceValueOf(v).DataPointer()
	offset, ok := reflectext.InternedValue(p)
	if !ok {
		t.Errorf("%v was not interned as expected", value)
	} else if offset != want {
		t.Errorf("unexpected offset for interned value %v: got %v, want %v", value, offset, want)
	}
	v2 := reflect.NewAt(reflect.TypeOf(value), offset.UnsafePointer()).Elem()
	if !reflect.DeepEqual(v.Interface(), v2.Interface()) {
		t.Errorf("unexpected round-trip: got %v, want %v", v2, v)
	}
}

func assertNotInterned(t *testing.T, value any) {
	v := reflect.ValueOf(&value).Elem()
	p := reflectext.InterfaceValueOf(v).DataPointer()
	_, ok := reflectext.InternedValue(p)
	if ok {
		t.Errorf("value %v was interned unexpectedly", value)
	}
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
