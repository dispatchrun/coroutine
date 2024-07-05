package reflectext

import (
	"reflect"
	"unsafe"
)

// SliceValue is a wrapper for a slice value that adds the
// ability to set the underlying data pointer, length and
// capacity, without allocating memory and regardless of
// the type.
//
// THis wrapper is not necessary when it's a byte slice, since
// reflect.Value has a SetBytes method, and it's possible to
// construct a []byte from data/len/cap using unsafe.Slice.
//
// TODO: submit proposal for reflect.SetSlice[S ~[]E, E any](S) ?
// (reflect.Value).SetSlice[S, E] isn't possible because Go
// doesn't support generic methods.
type SliceValue struct{ reflect.Value }

// SliceValueOf converts a slice value into a SliceValue.
func SliceValueOf(v reflect.Value) SliceValue {
	if v.Kind() != reflect.Slice {
		panic("not a slice")
	}
	return SliceValue{v}
}

// SetSlice sets the slice data pointer, length and capacity.
func (v SliceValue) SetSlice(data unsafe.Pointer, len, cap int) {
	if !v.CanAddr() {
		panic("slice is not addressable")
	}
	*(*sliceHeader)(v.Addr().UnsafePointer()) = sliceHeader{data: data, len: len, cap: cap}
}

// RawArrayValue constructs a "raw" array from an element type,
// a base pointer and a length (count).
//
// The return value can be used like an array reflect.Value,
// providing equivalent Index and Len method.
//
// RawArrayValue is useful in cases where you'd like to avoid
// reflect.ArrayOf, which creates a reflect.Type that isn't
// garbage collected.
type RawArrayValue struct {
	typ  reflect.Type
	base unsafe.Pointer
	len  int
}

// RawArrayValue constructs a RawArrayValue.
func RawArrayValueOf(t reflect.Type, base unsafe.Pointer, len int) RawArrayValue {
	return RawArrayValue{t, base, len}
}

// Index returns the slice's i'th element.
func (r RawArrayValue) Index(i int) reflect.Value {
	if i < 0 || i >= r.len {
		panic("index out of range")
	}
	p := unsafe.Add(r.base, i*int(r.typ.Size()))
	return reflect.NewAt(r.typ, p).Elem()
}

// Len returns the array length.
func (r RawArrayValue) Len() int {
	return r.len
}

// StructValue is a wrapper for a struct value that provides
// unrestricted access to unexported fields (e.g. no read-only flag).
type StructValue struct {
	reflect.Value

	base unsafe.Pointer
}

// StructValueOf converts a struct value into a StructValue.
func StructValueOf(v reflect.Value) StructValue {
	if v.Kind() != reflect.Struct {
		panic("not a struct")
	}
	base := unsafeAddr(v)
	return StructValue{v, base}
}

// Field returns the i'th field of the struct v.
// It panics if v's Kind is not Struct or i is out of range.
func (v *StructValue) Field(i int) reflect.Value {
	if i > v.NumField() {
		panic("field out of range")
	}
	f := v.Type().Field(i)
	return reflect.NewAt(f.Type, unsafe.Add(v.base, f.Offset)).Elem()
}

// FuncValue is a wrapper for a func value that provides a way
// to mutate the function address, and access and mutate closure vars.
type FuncValue struct{ reflect.Value }

// FuncValueOf converts a func value into a FuncValue.
func FuncValueOf(v reflect.Value) FuncValue {
	if v.Kind() != reflect.Func {
		panic("not a func")
	}
	return FuncValue{v}
}

// Closure retrieves the function's closure variables as a struct value.
//
// The first field in the struct is the address of the function, and the
// remaining fields hold closure vars.
//
// Closure vars are only available for functions that have type information
// registered at runtime. See RegisterClosure for more information.
func (v FuncValue) Closure() (reflect.Value, bool) {
	addr := uintptr(v.UnsafePointer())
	if f := FuncByAddr(addr); f == nil {
		// function not found at addr
	} else if f.Type == nil {
		// function type info not registered
	} else if f.Closure != nil {
		h := *(**functionHeader)(unsafeAddr(v.Value))
		if h.addr != addr {
			panic("invalid closure")
		}
		closure := reflect.NewAt(f.Closure, unsafe.Pointer(h)).Elem()
		return closure, true
	}
	return reflect.Value{}, false
}

// SetClosure sets the function address and closure vars.
func (v FuncValue) SetClosure(addr uintptr, c reflect.Value) {
	if c.Kind() != reflect.Struct || c.Type().Field(0).Type.Kind() != reflect.Uintptr {
		panic("invalid closure vars")
	}
	h := (*functionHeader)(unsafeAddr(c))
	h.addr = addr
	p := v.Addr().UnsafePointer()
	*(*unsafe.Pointer)(p) = unsafe.Pointer(h)
}

// SetAddr sets the function address, and clears any closure vars.
func (v FuncValue) SetAddr(addr uintptr) {
	h := &functionHeader{addr: addr}
	p := v.Addr().UnsafePointer()
	*(*unsafe.Pointer)(p) = unsafe.Pointer(h)
}

// InterfaceValue is a wrapper for an interface value that provides
// access to a pointer to the interface data.
type InterfaceValue struct{ reflect.Value }

// InterfaceValueOf converts an interface value into a InterfaceValue.
func InterfaceValueOf(v reflect.Value) InterfaceValue {
	if v.Kind() != reflect.Interface {
		panic("not an interface")
	}
	return InterfaceValue{v}
}

// DataPointer is a pointer to the interface data.
func (v InterfaceValue) DataPointer() unsafe.Pointer {
	return unsafeAddr(v.Value)
}

func unsafeAddr(v reflect.Value) unsafe.Pointer {
	if v.CanAddr() && v.Kind() != reflect.Interface {
		return v.Addr().UnsafePointer()
	}
	vi := v.Interface()
	i := (*interfaceHeader)(unsafe.Pointer(&vi))
	if ifaceInline(reflect.TypeOf(vi)) {
		return unsafe.Pointer(&i.ptr)
	}
	return i.ptr
}

func ifaceInline(t reflect.Type) bool {
	if t == nil {
		return false
	}
	switch t.Kind() {
	case reflect.Func:
		return true
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Struct:
		return t.NumField() == 1 && ifaceInline(t.Field(0).Type)
	case reflect.Array:
		return t.Len() == 1 && ifaceInline(t.Elem())
	default:
		return false
	}
}

// NamedTypeOffset is an opaque identifier for a named type.
//
// It is used to round-trip named types for a given version of the program.
type NamedTypeOffset uint64

// OffsetForNamedType gets the offset of a named type.
func OffsetForNamedType(t reflect.Type) NamedTypeOffset {
	// FIXME: investigate
	// if t.Name() == "" {
	//   panic("not a named type")
	// }
	tptr := (*interfaceHeader)(unsafe.Pointer(&t)).ptr
	bptr := (*interfaceHeader)(unsafe.Pointer(&ByteType)).ptr
	return NamedTypeOffset(uintptr(tptr) - uintptr(bptr))
}

// NamedTypeForOffset gets the named type for an offset.
func NamedTypeForOffset(offset NamedTypeOffset) reflect.Type {
	biface := (*interfaceHeader)(unsafe.Pointer(&ByteType))
	tiface := &interfaceHeader{
		typ: biface.typ,
		ptr: unsafe.Add(biface.ptr, offset),
	}
	return *(*reflect.Type)(unsafe.Pointer(tiface))
}

const internedCount = 256

var internedBase unsafe.Pointer

func init() {
	zero := 0
	var x interface{} = zero
	internedBase = (*interfaceHeader)(unsafe.Pointer(&x)).ptr
}

func InternedInt(p unsafe.Pointer) (int, bool) {
	if interned(p) {
		return internedOffset(p), true
	}
	return 0, false
}

func interned(p unsafe.Pointer) bool {
	return uintptr(p) >= uintptr(internedBase) && uintptr(p) < uintptr(internedBase)+internedCount
}

func internedOffset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(internedBase))
}

func InternedIntPointer(offset int) unsafe.Pointer {
	return unsafe.Add(internedBase, offset)
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

type functionHeader struct {
	addr uintptr
	// closure vars follow...
}

type interfaceHeader struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}
