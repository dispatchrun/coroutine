package reflectext

import (
	"fmt"
	"reflect"
)

// Visitor visits values in a reflect.Value graph.
type Visitor interface {
	// Visit is called first for values in the graph.
	//
	// If the function returns false, the visitor does not call any
	// other methods and does not recurse into nested values.
	Visit(reflect.Value) bool

	// VisitBool is called when a bool value is encountered.
	VisitBool(reflect.Value)

	// VisitInt is called when a integer value is encountered.
	//
	// The value has a kind of reflect.{Int,Int8,Int16,Int32,Int64}.
	VisitInt(reflect.Value)

	// VisitUint is called when a unsigned integer value is encountered.
	//
	// The value has a kind of reflect.{Uint,Uint8,Uint16,Uint32,Uint64,Uintptr}.
	VisitUint(reflect.Value)

	// VisitFloat is called when a float value is encountered.
	//
	// The value has a kind of reflect.{Float32,Float64}.
	VisitFloat(reflect.Value)

	// VisitComplex is called when a complex value is encountered.
	//
	// The value has a kind of reflect.{Complex64,Complex128}.
	//
	// If the function returns true, the visitor will visit
	// the nested real and imaginary components.
	VisitComplex(reflect.Value) bool

	// VisitString is called when a string value is encountered.
	//
	// Note that the visitor does not visit the nested *byte pointer.
	VisitString(reflect.Value)

	// VisitPointer is called when a typed pointer is encountered.
	//
	// If the function returns true, the visitor will visit the value
	// the pointer points to (if not null).
	VisitPointer(reflect.Value) bool

	// VisitUnsafePointer is called when an unsafe.Pointer value is encountered.
	VisitUnsafePointer(reflect.Value)

	// VisitArray is called when an array value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// item in the array.
	VisitArray(reflect.Value) bool

	// VisitSlice is called when a slice value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// item in the slice.
	VisitSlice(reflect.Value) bool

	// VisitMap is called when a map value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// key/value pair in the map in an unspecified order.
	VisitMap(reflect.Value) bool

	// VisitChan is called when a channel value is encountered.
	//
	// The visitor does not currently visit items in the channel,
	// since doing so would be a blocking operation.
	//
	// TODO: support visiting buffered channel items
	VisitChan(reflect.Value) bool

	// VisitStruct is called when a struct value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// field in the struct.
	VisitStruct(reflect.Value) bool

	// VisitFunc is called when a function value is encountered.
	//
	// If the function returns true, and the appropriate flag is set,
	// the visitor will visit each closure var (if applicable).
	VisitFunc(reflect.Value) bool

	// VisitInterface is called when an interface value is encountered.
	//
	// If the function returns true, the visitor will visit the
	// interface element.
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
		visitor.VisitBool(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		visitor.VisitInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		visitor.VisitUint(v)

	case reflect.Float32, reflect.Float64:
		visitor.VisitFloat(v)

	case reflect.Complex64, reflect.Complex128:
		if visitor.VisitComplex(v) {
			c := v.Complex()
			visitor.VisitFloat(reflect.ValueOf(real(c)))
			visitor.VisitFloat(reflect.ValueOf(imag(c)))
		}

	case reflect.String:
		visitor.VisitString(v)

	case reflect.Pointer:
		if visitor.VisitPointer(v) && !v.IsNil() {
			Visit(visitor, v.Elem(), flags)
		}

	case reflect.UnsafePointer:
		visitor.VisitUnsafePointer(v)

	case reflect.Array:
		if visitor.VisitArray(v) {
			for i := 0; i < v.Len(); i++ {
				Visit(visitor, v.Index(i), flags)
			}
		}

	case reflect.Slice:
		if visitor.VisitSlice(v) {
			// TODO: iterate up to v.Cap() if a flag is set
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
			if (flags & VisitUnexportedFields) != 0 {
				// The wrapper makes unexported fields available.
				unrestricted := StructValueOf(v)
				for i := 0; i < unrestricted.NumField(); i++ {
					Visit(visitor, unrestricted.Field(i), flags)
				}
			} else {
				t := v.Type()
				for i := 0; i < v.NumField(); i++ {
					if ft := t.Field(i); ft.IsExported() {
						Visit(visitor, v.Field(i), flags)
					}
				}
			}
		}

	case reflect.Func:
		if visitor.VisitFunc(v) && !v.IsNil() && (flags&VisitClosures) != 0 {
			// The wrapper makes closure vars available.
			fv := FuncValueOf(v)
			if closure, ok := fv.Closure(); ok {
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

// DefaultVisitor is a Visitor that visits all values in a
// reflect.Value graph.
type DefaultVisitor struct{}

var _ Visitor = DefaultVisitor{}

func (DefaultVisitor) Visit(reflect.Value) bool          { return true }
func (DefaultVisitor) VisitBool(reflect.Value)           {}
func (DefaultVisitor) VisitInt(reflect.Value)            {}
func (DefaultVisitor) VisitUint(reflect.Value)           {}
func (DefaultVisitor) VisitFloat(reflect.Value)          {}
func (DefaultVisitor) VisitComplex(reflect.Value) bool   { return true }
func (DefaultVisitor) VisitString(reflect.Value)         {}
func (DefaultVisitor) VisitPointer(reflect.Value) bool   { return true }
func (DefaultVisitor) VisitUnsafePointer(reflect.Value)  {}
func (DefaultVisitor) VisitArray(reflect.Value) bool     { return true }
func (DefaultVisitor) VisitSlice(reflect.Value) bool     { return true }
func (DefaultVisitor) VisitMap(reflect.Value) bool       { return true }
func (DefaultVisitor) VisitChan(reflect.Value) bool      { return true }
func (DefaultVisitor) VisitStruct(reflect.Value) bool    { return true }
func (DefaultVisitor) VisitFunc(reflect.Value) bool      { return true }
func (DefaultVisitor) VisitInterface(reflect.Value) bool { return true }
