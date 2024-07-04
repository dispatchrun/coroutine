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
	Visit(VisitContext, reflect.Value) bool

	// VisitBool is called when a bool value is encountered.
	VisitBool(VisitContext, reflect.Value)

	// VisitInt is called when a integer value is encountered.
	//
	// The value has a kind of reflect.{Int,Int8,Int16,Int32,Int64}.
	VisitInt(VisitContext, reflect.Value)

	// VisitUint is called when a unsigned integer value is encountered.
	//
	// The value has a kind of reflect.{Uint,Uint8,Uint16,Uint32,Uint64,Uintptr}.
	VisitUint(VisitContext, reflect.Value)

	// VisitFloat is called when a float value is encountered.
	//
	// The value has a kind of reflect.{Float32,Float64}.
	VisitFloat(VisitContext, reflect.Value)

	// VisitComplex is called when a complex value is encountered.
	//
	// The value has a kind of reflect.{Complex64,Complex128}.
	//
	// If the function returns true, the visitor will visit
	// the nested real and imaginary components.
	VisitComplex(VisitContext, reflect.Value) bool

	// VisitString is called when a string value is encountered.
	//
	// Note that the visitor does not visit the nested *byte pointer.
	VisitString(VisitContext, reflect.Value)

	// VisitPointer is called when a typed pointer is encountered.
	//
	// If the function returns true, the visitor will visit the value
	// the pointer points to (if not null).
	VisitPointer(VisitContext, reflect.Value) bool

	// VisitUnsafePointer is called when an unsafe.Pointer value is encountered.
	VisitUnsafePointer(VisitContext, reflect.Value)

	// VisitArray is called when an array value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// item in the array.
	VisitArray(VisitContext, reflect.Value) bool

	// VisitSlice is called when a slice value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// item in the slice.
	VisitSlice(VisitContext, reflect.Value) bool

	// VisitMap is called when a map value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// key/value pair in the map in an unspecified order.
	VisitMap(VisitContext, reflect.Value) bool

	// VisitChan is called when a channel value is encountered.
	//
	// The visitor does not currently visit items in the channel,
	// since doing so would be a blocking operation.
	//
	// TODO: support visiting buffered channel items
	VisitChan(VisitContext, reflect.Value) bool

	// VisitStruct is called when a struct value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// field in the struct.
	VisitStruct(VisitContext, reflect.Value) bool

	// VisitFunc is called when a function value is encountered.
	//
	// If the function returns true, and the appropriate flag is set,
	// the visitor will visit each closure var (if applicable).
	VisitFunc(VisitContext, reflect.Value) bool

	// VisitInterface is called when an interface value is encountered.
	//
	// If the function returns true, the visitor will visit the
	// interface element.
	VisitInterface(VisitContext, reflect.Value) bool
}

// VisitContext is Visitor context.
type VisitContext struct {
	// TODO
}

// VisitFlags controls a Visit operation.
type VisitFlags int

const (
	// VisitUnexportedFields instructs Visit to visit unexported struct fields.
	VisitUnexportedFields VisitFlags = 1 << iota

	// VisitClosures instructs Visit to visit values captured by closures.
	VisitClosures

	// VisitReflectValues instructs Visit to visit values contained within
	// nested reflect.Value.
	VisitReflectValues

	// VisitAll instructs Visit to visit all values in the graph.
	VisitAll = VisitUnexportedFields | VisitClosures | VisitReflectValues
)

// Visit walks a reflect.Value graph.
//
// The operation will follow pointers. It's the Visitor's responsibility
// to keep track of values/pointers that have been visited to prevent
// an infinite loop when there are cycles in the graph.
func Visit(visitor Visitor, v reflect.Value, flags VisitFlags) {
	// TODO:
	ctx := VisitContext{}

	if !visitor.Visit(ctx, v) {
		return
	}

	// Special case for reflect.Value.
	if (flags&VisitReflectValues) != 0 && v.Type() == ReflectValueType {
		rv := v.Interface().(reflect.Value)
		Visit(visitor, rv, flags)
		return
	}

	switch v.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't visit reflect.Invalid"))

	case reflect.Bool:
		visitor.VisitBool(ctx, v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		visitor.VisitInt(ctx, v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		visitor.VisitUint(ctx, v)

	case reflect.Float32, reflect.Float64:
		visitor.VisitFloat(ctx, v)

	case reflect.Complex64, reflect.Complex128:
		if visitor.VisitComplex(ctx, v) {
			c := v.Complex()
			visitor.VisitFloat(ctx, reflect.ValueOf(real(c)))
			visitor.VisitFloat(ctx, reflect.ValueOf(imag(c)))
		}

	case reflect.String:
		visitor.VisitString(ctx, v)

	case reflect.Pointer:
		if visitor.VisitPointer(ctx, v) && !v.IsNil() {
			Visit(visitor, v.Elem(), flags)
		}

	case reflect.UnsafePointer:
		visitor.VisitUnsafePointer(ctx, v)

	case reflect.Array:
		if visitor.VisitArray(ctx, v) {
			for i := 0; i < v.Len(); i++ {
				Visit(visitor, v.Index(i), flags)
			}
		}

	case reflect.Slice:
		if visitor.VisitSlice(ctx, v) {
			// TODO: iterate up to v.Cap() if a flag is set
			for i := 0; i < v.Len(); i++ {
				Visit(visitor, v.Index(i), flags)
			}
		}

	case reflect.Map:
		if visitor.VisitMap(ctx, v) && !v.IsNil() {
			iter := v.MapRange()
			for iter.Next() {
				Visit(visitor, iter.Key(), flags)
				Visit(visitor, iter.Value(), flags)
			}
		}

	case reflect.Chan:
		// TODO: visit buffered channel items if possible
		visitor.VisitChan(ctx, v)

	case reflect.Struct:
		if visitor.VisitStruct(ctx, v) {
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
		if visitor.VisitFunc(ctx, v) && !v.IsNil() && (flags&VisitClosures) != 0 {
			// The wrapper makes closure vars available.
			fv := FuncValueOf(v)
			if closure, ok := fv.Closure(); ok {
				Visit(visitor, closure, flags)
			}
		}

	case reflect.Interface:
		if visitor.VisitInterface(ctx, v) && !v.IsNil() {
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

func (DefaultVisitor) Visit(VisitContext, reflect.Value) bool          { return true }
func (DefaultVisitor) VisitBool(VisitContext, reflect.Value)           {}
func (DefaultVisitor) VisitInt(VisitContext, reflect.Value)            {}
func (DefaultVisitor) VisitUint(VisitContext, reflect.Value)           {}
func (DefaultVisitor) VisitFloat(VisitContext, reflect.Value)          {}
func (DefaultVisitor) VisitComplex(VisitContext, reflect.Value) bool   { return true }
func (DefaultVisitor) VisitString(VisitContext, reflect.Value)         {}
func (DefaultVisitor) VisitPointer(VisitContext, reflect.Value) bool   { return true }
func (DefaultVisitor) VisitUnsafePointer(VisitContext, reflect.Value)  {}
func (DefaultVisitor) VisitArray(VisitContext, reflect.Value) bool     { return true }
func (DefaultVisitor) VisitSlice(VisitContext, reflect.Value) bool     { return true }
func (DefaultVisitor) VisitMap(VisitContext, reflect.Value) bool       { return true }
func (DefaultVisitor) VisitChan(VisitContext, reflect.Value) bool      { return true }
func (DefaultVisitor) VisitStruct(VisitContext, reflect.Value) bool    { return true }
func (DefaultVisitor) VisitFunc(VisitContext, reflect.Value) bool      { return true }
func (DefaultVisitor) VisitInterface(VisitContext, reflect.Value) bool { return true }
