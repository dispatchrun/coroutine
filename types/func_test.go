package types

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestAddressOfNonFunctionValue(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("taking the address of a non-function value did not panic")
		}
	}()
	FuncAddr(0)
}

func TestAddressOfNilFunctionValue(t *testing.T) {
	if addr := FuncAddr((func())(nil)); addr != 0 {
		t.Errorf("wrong address for nil function value: want=0 got=%v", addr)
	}
}

func TestFunctionAddress(t *testing.T) {
	name, addr1 := sentinel()
	addr2 := FuncAddr(sentinel)

	if addr1 != addr2 {
		t.Errorf("%s: function address mismatch: want=%#v got=%#v", name, addr1, addr2)
	}
}

func TestFunctionLookup(t *testing.T) {
	name := "github.com/stealthrocket/coroutine/types.TestFunctionLookup"
	sym1 := FuncByName(name)
	sym2 := FuncByAddr(FuncAddr(TestFunctionLookup))

	if sym1 != sym2 {
		t.Errorf("%s: symbols returned by name and address lookups mismatch: %p != %p", name, sym1, sym2)
	}
}

func TestClosureAddress(t *testing.T) {
	f := op(42, 1)
	p := *(*unsafe.Pointer)(unsafe.Pointer(&f))
	c := (*closure)(p)

	name := "github.com/stealthrocket/coroutine/types.op.func1"
	addr1 := c.addr
	addr2 := FuncAddr(f)

	if addr1 != addr2 {
		t.Errorf("%s: closure address mismatch: want=%#v got=%#v", name, addr1, addr2)
	}
}

func TestClosureLookup(t *testing.T) {
	f := op(1, 2)

	name := "github.com/stealthrocket/coroutine/types.op.func1"
	sym1 := FuncByName(name)
	sym2 := FuncByAddr(FuncAddr(f))

	if sym1 != sym2 {
		t.Errorf("%s: symbols returned by name and address lookups mismatch: %p != %p", name, sym1, sym2)
	}
}

func TestRehydrateFunction(t *testing.T) {
	f := FuncByName("github.com/stealthrocket/coroutine/types.op")
	v := reflect.New(f.Type)
	p := v.UnsafePointer()

	// The reflect.New call constructs a (*func(int, int) func() int) value,
	// a pointer to a function value; the layout of a function value in memory
	// is a pointer to a location starting with the address of the referenced
	// function. This is the same layout as the Func type so we can assign the
	// function value the value of the *Func pointer since it starts with the
	// in-memory address of the corresponding Go function.
	(*(**Func)(p)) = f

	op := *(*func(int, int) func() int)(p)
	fn := op(1, 1)

	if res := fn(); res != 2 {
		t.Errorf("wrong value returned by rehydrated function: want=2 got=%d", res)
	}
}

func TestRehydrateClosure(t *testing.T) {
	f := FuncByName("github.com/stealthrocket/coroutine/types.op.func1")
	v := reflect.New(f.Closure)
	p := v.UnsafePointer()

	// When deserializing a closure its memory layout starts with the function
	// address followed by the free vars it captured.
	*(*uintptr)(p) = f.Addr
	c := (*opFunc1Closure)(p)
	c.a = 40
	c.b = 2

	fn := *(*func() int)(unsafe.Pointer(&p))

	if res := fn(); res != 42 {
		t.Errorf("wrong value returned by rehydrated closure: want=42 got=%d", res)
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

type opFunc1Closure struct {
	_ uintptr
	// closure free vars
	a int
	b int
}

// This init function contains the work that would normally be done by the
// compiler to generate the reflect data necessary to serialize functions.
func init() {
	op := FuncByName("github.com/stealthrocket/coroutine/types.op")
	op.Type = reflect.TypeOf(func(int, int) (_ func() int) { return })

	fn := FuncByName("github.com/stealthrocket/coroutine/types.op.func1")
	fn.Type = reflect.TypeOf(func() (_ int) { return })
	fn.Closure = reflect.TypeOf(opFunc1Closure{})
}
