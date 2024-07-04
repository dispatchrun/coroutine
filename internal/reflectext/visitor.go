package reflectext

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Visitor visits values in a reflect.Value graph.
type Visitor interface {
	// Visit is called first for values in the graph.
	//
	// If the function returns false, the visitor does not call any
	// other methods and does not recurse into nested values.
	Visit(reflect.Value) bool

	// VisitBool is called when a bool value is encountered.
	VisitBool(bool)

	// VisitInt is called when a int value is encountered.
	VisitInt(int)

	// VisitInt8 is called when a int8 value is encountered.
	VisitInt8(int8)

	// VisitInt16 is called when a int16 value is encountered.
	VisitInt16(int16)

	// VisitInt32 is called when a int32 value is encountered.
	VisitInt32(int32)

	// VisitInt64 is called when a int64 value is encountered.
	VisitInt64(int64)

	// VisitUint is called when a uint value is encountered.
	VisitUint(uint)

	// VisitUint8 is called when a uint8 value is encountered.
	VisitUint8(uint8)

	// VisitUint16 is called when a uint16 value is encountered.
	VisitUint16(uint16)

	// VisitUint32 is called when a uint32 value is encountered.
	VisitUint32(uint32)

	// VisitUint64 is called when a uint64 value is encountered.
	VisitUint64(uint64)

	// VisitUintptr is called when a uintptr value is encountered.
	VisitUintptr(uintptr)

	// VisitFloat32 is called when a float32 value is encountered.
	VisitFloat32(float32)

	// VisitFloat64 is called when a float64 value is encountered.
	VisitFloat64(float64)

	// VisitComplex64 is called when a complex64 value is encountered.
	//
	// If the function returns true, the visitor will visit
	// the nested float32 real and imaginary components.
	VisitComplex64(complex64) bool

	// VisitComplex128 is called when a complex128 value is encountered.
	//
	// If the function returns true, the visitor will visit
	// the nested float64 real and imaginary components.
	VisitComplex128(complex128) bool

	// VisitString is called when a string value is encountered.
	//
	// Note that the visitor does not visit the nested *byte pointer.
	VisitString(string)

	// VisitUnsafePointer is called when an unsafe.Pointer value is encountered.
	VisitUnsafePointer(unsafe.Pointer)

	// VisitPointer is called when a typed pointer is encountered.
	//
	// If the function returns true, the visitor will visit the value
	// the pointer points to (if not null).
	VisitPointer(reflect.Value) bool

	VisitArray(reflect.Value) bool
	VisitSlice(reflect.Value) bool
	VisitMap(reflect.Value) bool
	VisitChan(reflect.Value) bool
	VisitStruct(reflect.Value) bool
	VisitFunc(reflect.Value) bool
	VisitInterface(reflect.Value) bool
}

// VisitFlags controls a Visit operation.
type VisitFlags int

const (
	// VisitUnexportedFields instructs Visit to visit unexported struct fields.
	VisitUnexportedFields VisitFlags = 1 << iota

	// VisitClosures instructs Visit to visit values captured by closures.
	VisitClosures
)

// Visit walks a reflect.Value graph.
func Visit(visitor Visitor, v reflect.Value, flags VisitFlags) {
	if !visitor.Visit(v) {
		return
	}

	// Special case for reflect.Value.
	if v.Type() == ReflectValueType {
		rv := v.Interface().(reflect.Value)
		Visit(visitor, rv, flags)
		return
	}

	switch v.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't visit reflect.Invalid"))

	case reflect.Bool:
		visitor.VisitBool(v.Bool())

	case reflect.Int:
		visitor.VisitInt(int(v.Int()))

	case reflect.Int8:
		visitor.VisitInt8(int8(v.Int()))

	case reflect.Int16:
		visitor.VisitInt16(int16(v.Int()))

	case reflect.Int32:
		visitor.VisitInt32(int32(v.Int()))

	case reflect.Int64:
		visitor.VisitInt64(v.Int())

	case reflect.Uint:
		visitor.VisitUint(uint(v.Uint()))

	case reflect.Uint8:
		visitor.VisitUint8(uint8(v.Uint()))

	case reflect.Uint16:
		visitor.VisitUint16(uint16(v.Uint()))

	case reflect.Uint32:
		visitor.VisitUint32(uint32(v.Uint()))

	case reflect.Uint64:
		visitor.VisitUint64(v.Uint())

	case reflect.Uintptr:
		visitor.VisitUintptr(uintptr(v.Uint()))

	case reflect.Float32:
		visitor.VisitFloat32(float32(v.Float()))

	case reflect.Float64:
		visitor.VisitFloat64(float64(v.Float()))

	case reflect.String:
		visitor.VisitString(v.String())

	case reflect.Complex64:
		c := complex64(v.Complex())
		if visitor.VisitComplex64(c) {
			visitor.VisitFloat32(real(c))
			visitor.VisitFloat32(imag(c))
		}

	case reflect.Complex128:
		c := v.Complex()
		if visitor.VisitComplex128(c) {
			visitor.VisitFloat64(real(c))
			visitor.VisitFloat64(imag(c))
		}

	case reflect.UnsafePointer:
		visitor.VisitUnsafePointer(v.UnsafePointer())

	case reflect.Pointer:
		if visitor.VisitPointer(v) && !v.IsNil() {
			Visit(visitor, v.Elem(), flags)
		}
	case reflect.Array:
		if visitor.VisitArray(v) {
			for i := 0; i < v.Len(); i++ {
				Visit(visitor, v.Index(i), flags)
			}
		}

	case reflect.Slice:
		if visitor.VisitSlice(v) {
			for i := 0; i < v.Len(); i++ {
				Visit(visitor, v.Index(i), flags)
			}
		}

	case reflect.Map:
		if visitor.VisitMap(v) && !v.IsNil() {
			iter := v.MapRange()
			for iter.Next() {
				Visit(visitor, iter.Key(), flags)
				Visit(visitor, iter.Value(), flags)
			}
		}

	case reflect.Chan:
		// TODO: visit buffered channel items if possible
		visitor.VisitChan(v)

	case reflect.Struct:
		if visitor.VisitStruct(v) {
			p := unsafePtr(v)
			t := v.Type()
			for i := 0; i < v.NumField(); i++ {
				ft := t.Field(i)
				if ft.IsExported() {
					Visit(visitor, v.Field(i), flags)
				} else if (flags & VisitUnexportedFields) != 0 {
					field := reflect.NewAt(ft.Type, unsafe.Add(p, ft.Offset)).Elem()
					Visit(visitor, field, flags)
				}
			}
		}

	case reflect.Func:
		if visitor.VisitFunc(v) && !v.IsNil() {
			addr := v.UnsafePointer()
			if f := FuncByAddr(uintptr(addr)); f == nil {
				// function not found at addr
			} else if f.Type == nil {
				// function type info not registered
			} else if f.Closure != nil && (flags&VisitClosures) != 0 {
				fp := *(**FunctionHeader)(unsafePtr(v))
				if fp.Addr != addr {
					panic("invalid closure")
				}
				closure := reflect.NewAt(f.Closure, unsafe.Pointer(fp)).Elem()
				Visit(visitor, closure, flags)
			}
		}

	case reflect.Interface:
		if visitor.VisitInterface(v) && !v.IsNil() {
			Visit(visitor, v.Elem(), flags)
		}

	default:
		panic("unreachable")
	}
}

func unsafePtr(v reflect.Value) unsafe.Pointer {
	i := v.Interface()
	return IfacePtr(unsafe.Pointer(&i), reflect.TypeOf(i))
}

// DefaultVisitor is a Visitor that visits all values in a
// reflect.Value graph.
type DefaultVisitor struct{}

var _ Visitor = DefaultVisitor{}

func (DefaultVisitor) Visit(reflect.Value) bool          { return true }
func (DefaultVisitor) VisitBool(bool)                    {}
func (DefaultVisitor) VisitInt(int)                      {}
func (DefaultVisitor) VisitInt8(int8)                    {}
func (DefaultVisitor) VisitInt16(int16)                  {}
func (DefaultVisitor) VisitInt32(int32)                  {}
func (DefaultVisitor) VisitInt64(int64)                  {}
func (DefaultVisitor) VisitUint(uint)                    {}
func (DefaultVisitor) VisitUint8(uint8)                  {}
func (DefaultVisitor) VisitUint16(uint16)                {}
func (DefaultVisitor) VisitUint32(uint32)                {}
func (DefaultVisitor) VisitUint64(uint64)                {}
func (DefaultVisitor) VisitUintptr(uintptr)              {}
func (DefaultVisitor) VisitFloat32(float32)              {}
func (DefaultVisitor) VisitFloat64(float64)              {}
func (DefaultVisitor) VisitString(string)                {}
func (DefaultVisitor) VisitUnsafePointer(unsafe.Pointer) {}
func (DefaultVisitor) VisitComplex64(complex64) bool     { return true }
func (DefaultVisitor) VisitComplex128(complex128) bool   { return true }
func (DefaultVisitor) VisitPointer(reflect.Value) bool   { return true }
func (DefaultVisitor) VisitArray(reflect.Value) bool     { return true }
func (DefaultVisitor) VisitSlice(reflect.Value) bool     { return true }
func (DefaultVisitor) VisitMap(reflect.Value) bool       { return true }
func (DefaultVisitor) VisitChan(reflect.Value) bool      { return true }
func (DefaultVisitor) VisitStruct(reflect.Value) bool    { return true }
func (DefaultVisitor) VisitFunc(reflect.Value) bool      { return true }
func (DefaultVisitor) VisitInterface(reflect.Value) bool { return true }
