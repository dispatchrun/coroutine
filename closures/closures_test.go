package closures

import (
	"testing"
	"unsafe"
)

func TestAddressOfNonFunctionValue(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("taking the address of a non-function value did not panic")
		}
	}()
	Address(0)
}

func TestAddressOfNilFunctionValue(t *testing.T) {
	if addr := Address((func())(nil)); addr != 0 {
		t.Errorf("wrong address for nil function value: want=0 got=%v", addr)
	}
}

func TestFunctionAddress(t *testing.T) {
	name, addr1 := sentinel()
	addr2 := Address(sentinel)

	if addr1 != addr2 {
		t.Errorf("%s: function address mismatch: want=%#v got=%#v", name, addr1, addr2)
	}
}

func TestFunctionLookup(t *testing.T) {
	name := "github.com/stealthrocket/coroutine/closures.TestFunctionLookup"
	sym1 := LookupByName(name)
	sym2 := LookupByAddr(Address(TestFunctionLookup))

	if sym1 != sym2 {
		t.Errorf("%s: symbols returned by name and address lookups mismatch: %p != %p", name, sym1, sym2)
	}
}

func TestClosureAddress(t *testing.T) {
	f := op(42, 1)
	p := *(*unsafe.Pointer)(unsafe.Pointer(&f))
	c := (*closure)(p)

	name := "github.com/stealthrocket/coroutine/closures.op.func1"
	addr1 := c.addr
	addr2 := Address(f)

	if addr1 != addr2 {
		t.Errorf("%s: closure address mismatch: want=%#v got=%#v", name, addr1, addr2)
	}
}

func TestClosureLookup(t *testing.T) {
	f := op(1, 2)

	name := "github.com/stealthrocket/coroutine/closures.op.func1"
	sym1 := LookupByName(name)
	sym2 := LookupByAddr(Address(f))

	if sym1 != sym2 {
		t.Errorf("%s: symbols returned by name and address lookups mismatch: %p != %p", name, sym1, sym2)
	}
}

// We have to set go:noinline on this function to make sure that the symbol
// remains in the compiled program otherwise the function tables will not
// contain an entry for it.
//
//go:noinline
func op(a, b int) func() int {
	return func() int { return a + b }
}
