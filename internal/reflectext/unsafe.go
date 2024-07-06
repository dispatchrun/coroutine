package reflectext

import (
	"reflect"
	"unsafe"
)

// MakeSlice makes a slice from a raw data pointer, length
// and capacity.
//
// There's already a reflect.MakeSlice(typ, len, cap), but
// it doesn't let you set the underlying data pointer.
// This function is useful in cases where you want/need to
// avoid making a copy, either to avoid the extra allocations
// or because you need to create references or aliased sliced
// of the same underlying array.
//
// This function is not necessary when the type is known at
// compile time. If the type is known at compile time then
// it's possible to construct a slice from data/len/cap
// using unsafe.Slice. You can then create a reflect.Value
// using reflect.ValueOf(slice), or you can assign a slice
// to an existing reflect.Value as long as it's a byte slice
// using SetBytes:
//
//	ints := unsafe.Slice((*int)(data), cap)[:len:cap])
//	v := reflect.ValueOf(ints)
//
//	bytes := unsafe.Slice((*byte)(data), cap)[:len:cap]
//	v := reflect.MakeSlice(typ, 0, 0)
//	v.SetBytes(bytes) // (if you have a reflect.Value already)
//
// The function is also not necessary if a copy will suffice:
//
//	slice := reflect.MakeSlice(typ, len, cap)
//	size := cap * int(typ.Elem().Size())
//	copy(
//	  unsafe.Slice((*byte)(v.UnsafePointer()), size),
//	  unsafe.Slice((*byte)(data), size))
//
// Unfortunately, there doesn't seem to be a way to construct
// a slice from a data/len/cap when the type isn't known until
// runtime (and only a reflect.Type is available).
func MakeSlice(typ reflect.Type, data unsafe.Pointer, len, cap int) reflect.Value {
	s := reflect.New(typ).Elem()
	*(*sliceHeader)(s.Addr().UnsafePointer()) = sliceHeader{data: data, len: len, cap: cap}
	return s
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
	return StructValue{Value: v, base: unsafeAddr(v)}
}

// Field returns the i'th field of the struct.
func (v *StructValue) Field(i int) reflect.Value {
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
	// If the value isn't addressable, we can box the
	// value and then get the address of the data from
	// within the interface.
	vi := v.Interface()
	i := (*interfaceHeader)(unsafe.Pointer(&vi))
	if inlineIfaceData(reflect.TypeOf(vi)) {
		return unsafe.Pointer(&i.ptr)
	}
	return i.ptr
}

func inlineIfaceData(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Func:
		return true
	case reflect.Ptr:
		return true
	case reflect.Map:
		return true
	case reflect.Struct:
		return t.NumField() == 1 && inlineIfaceData(t.Field(0).Type)
	case reflect.Array:
		return t.Len() == 1 && inlineIfaceData(t.Elem())
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

// InternedValue checks whether the pointer points to the interned value
// region, and if so returns the offset and true. If the pointer does not
// point to the interned value region, the function returns false.
//
// Go interns boxed booleans and small integers in order to avoid
// allocations.
func InternedValue(p unsafe.Pointer) (InternedValueOffset, bool) {
	if uintptr(p) >= uintptr(internedValueBase) && uintptr(p) < uintptr(internedValueBase)+internedValueBytes {
		return InternedValueOffset(uintptr(p) - uintptr(internedValueBase)), true
	}
	return 0, false
}

// InternedValueOffset is an offset into the interned value region of
// memory.
type InternedValueOffset int

// UnsafePointer is the interned value offset as a pointer.
func (o InternedValueOffset) UnsafePointer() unsafe.Pointer {
	return unsafe.Add(internedValueBase, uintptr(o))
}

const internedValueBytes = 256 // determined experimentally
var internedValueBase unsafe.Pointer

func init() {
	zero := 0
	var x interface{} = zero
	internedValueBase = (*interfaceHeader)(unsafe.Pointer(&x)).ptr
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
