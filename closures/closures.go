// Package closures contains the infrastructure needed to support serialization
// of Go function values.
package closures

import (
	"debug/gosym"
	"io"
	"reflect"
	"runtime"
	"unsafe"
)

// Func represents a function in the program.
type Func struct {
	// The address where the function exists in the program memory.
	Addr uintptr

	// The name that uniquely represents the function.
	//
	// For regular functions, this values has the form <package>.<function>.
	//
	// For closures, this value has the form <package>.<function>.func<N>, where
	// N starts at 1 and increments for each closure defined in the function.
	Name string

	// A type representing the signature of the function value.
	//
	// This field is nil if the type is unknown; by default the field is nil and
	// the program is expected to initialize it to a non-nil value for functions
	// that may be serialized.
	//
	// If non-nil, the type must be of kind reflect.Func.
	Type reflect.Type

	// A struct type representing the memory layout of the closure.
	//
	// This field is nil if the type is unknown; by default the field is nil and
	// the program is expected to initialize it to a non-nil value for closures
	// that may be serialized. For regular functions, this field can remain nil
	// since regular functions do not capture any values.
	//
	// If non-nil, the first field of the struct type must be a uintptr intended
	// to hold the address to the function value.
	Closure reflect.Type
}

// Go function values are pointers to an object starting with the function
// address, whether they are referencing top-level functions or closures.
//
// The Address function uses this type to dereference the function value and
// access the address of the function in memory.
type closure struct{ addr uintptr }

// Address returns the address in memory of the closure passed as argument.
//
// This function can only resolve addresses of closures in the compilation unit
// that it is part of; for example, if compiled in a Go plugin, it can only
// resolve the address of functions within that plugin, and the main program
// cannot resolve addresses of functions in the plugins it loaded.
//
// If the argument is a nil function value, the return address is zero.
//
// The function panics if called with a value which is not a function.
func Address(fn any) uintptr {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panic("value must be a function")
	}
	v := (*[2]unsafe.Pointer)(unsafe.Pointer(&fn))
	c := (*closure)(v[1])
	if c == nil {
		return 0
	}
	return c.addr
}

// LookupByName returns the function associated with the given name.
//
// Addresses in the returned Func value hold the value of the symbol location in
// the program memory.
//
// If the name passed as argument does not represent any function, the function
// returns nil.
func LookupByName(name string) *Func { return functionsByName[name] }

// LookupByAddr returns the function associated with the given address.
//
// Addresses in the returned Func value hold the value of the symbol location in
// the program memory.
//
// If the address passed as argument is not the address of a function in the
// program, the function returns nil.
func LookupByAddr(addr uintptr) *Func { return functionsByAddr[addr] }

var (
	functionsByName map[string]*Func
	functionsByAddr map[uintptr]*Func
)

func initFunctionTables(pclntab, symtab []byte) {
	table, err := gosym.NewTable(symtab, gosym.NewLineTable(pclntab, 0))
	if err != nil {
		panic("cannot read symtab: " + err.Error())
	}

	sentinelName, sentinelAddr := sentinel()

	tableFunc := table.LookupFunc(sentinelName)
	offset := uint64(sentinelAddr) - tableFunc.Entry

	functions := make([]Func, len(table.Funcs))
	for i, fn := range table.Funcs {
		functions[i] = Func{
			Addr: uintptr(fn.Entry + offset),
			Name: fn.Name,
		}
	}

	functionsByName = make(map[string]*Func, len(functions))
	functionsByAddr = make(map[uintptr]*Func, len(functions))

	for i := range functions {
		f := &functions[i]
		functionsByName[f.Name] = f
		functionsByAddr[f.Addr] = f
	}
}

func readAll(r io.ReaderAt, size uint64) ([]byte, error) {
	b := make([]byte, size)
	n, err := r.ReadAt(b, 0)
	if err != nil && n < len(b) {
		return nil, err
	}
	return b, nil
}

//go:noinline
func sentinel() (name string, addr uintptr) {
	pc := [1]uintptr{}
	runtime.Callers(0, pc[:])

	fn := runtime.FuncForPC(pc[0])
	return fn.Name(), fn.Entry()
}
