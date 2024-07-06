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
	Visit(VisitorContext, reflect.Value) bool

	// VisitBool is called when a bool value is encountered.
	VisitBool(VisitorContext, reflect.Value)

	// VisitInt is called when a integer value is encountered.
	//
	// The value has a kind of reflect.{Int,Int8,Int16,Int32,Int64}.
	VisitInt(VisitorContext, reflect.Value)

	// VisitUint is called when a unsigned integer value is encountered.
	//
	// The value has a kind of reflect.{Uint,Uint8,Uint16,Uint32,Uint64,Uintptr}.
	VisitUint(VisitorContext, reflect.Value)

	// VisitFloat is called when a float value is encountered.
	//
	// The value has a kind of reflect.{Float32,Float64}.
	VisitFloat(VisitorContext, reflect.Value)

	// VisitComplex is called when a complex value is encountered.
	//
	// The value has a kind of reflect.{Complex64,Complex128}.
	//
	// If the function returns true, the visitor will visit
	// the nested real and imaginary components.
	VisitComplex(VisitorContext, reflect.Value) bool

	// VisitString is called when a string value is encountered.
	//
	// Note that the visitor does not visit the nested *byte pointer.
	VisitString(VisitorContext, reflect.Value)

	// VisitPointer is called when a typed pointer is encountered.
	//
	// If the function returns true, the visitor will visit the value
	// the pointer points to (if not null).
	VisitPointer(VisitorContext, reflect.Value) bool

	// VisitUnsafePointer is called when an unsafe.Pointer value is encountered.
	VisitUnsafePointer(VisitorContext, reflect.Value)

	// VisitArray is called when an array value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// item in the array.
	VisitArray(VisitorContext, reflect.Value) bool

	// VisitSlice is called when a slice value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// item in the slice.
	VisitSlice(VisitorContext, reflect.Value) bool

	// VisitMap is called when a map value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// key/value pair in the map in an unspecified order.
	VisitMap(VisitorContext, reflect.Value) bool

	// VisitChan is called when a channel value is encountered.
	//
	// The visitor does not currently visit items in the channel,
	// since doing so would be a blocking operation.
	//
	// TODO: support visiting buffered channel items
	VisitChan(VisitorContext, reflect.Value) bool

	// VisitStruct is called when a struct value is encountered.
	//
	// If the function returns true, the visitor will visit each
	// field in the struct.
	VisitStruct(VisitorContext, reflect.Value) bool

	// VisitFunc is called when a function value is encountered.
	//
	// If the function returns true, and the appropriate flag is set,
	// the visitor will visit each closure var (if applicable).
	VisitFunc(VisitorContext, reflect.Value) bool

	// VisitInterface is called when an interface value is encountered.
	//
	// If the function returns true, the visitor will visit the
	// interface element.
	VisitInterface(VisitorContext, reflect.Value) bool
}

// NewVisitor creates a visitor.
func NewVisitor(impl Visitor, flags VisitorFlags) VisitorContext {
	return VisitorContext{impl: impl, flags: flags}
}

// VisitorFlags controls a Visit operation.
type VisitorFlags int

const (
	// VisitUnexportedFields instructs Visit to visit unexported struct fields.
	VisitUnexportedFields VisitorFlags = 1 << iota

	// VisitClosures instructs Visit to visit values captured by closures.
	//
	// Closure types must be registered at runtime. See RegisterClosure for
	// more information.
	VisitClosures

	// VisitReflectValues instructs Visit to visit values contained within
	// nested reflect.Value.
	VisitReflectValues

	// VisitSliceLenToCap instructs Visit to visit slice elements between
	// the slice length and slice capacity.
	VisitSliceLenToCap

	// VisitAll instructs Visit to visit all values in the graph.
	VisitAll = VisitUnexportedFields | VisitClosures | VisitReflectValues | VisitSliceLenToCap
)

// Has is true if the flag is set.
func (flags VisitorFlags) Has(flag VisitorFlags) bool {
	return (flags & flag) != 0
}

// VisitorContext is Visitor context.
type VisitorContext struct {
	parent *VisitorContext
	impl   Visitor
	flags  VisitorFlags
}

// Visit walks a reflect.Value graph.
//
// The operation will follow all pointers. It's the Visitor's responsibility
// to keep track of values/pointers that have been visited to prevent
// an infinite loop when there are cycles in the graph.
func (ctx VisitorContext) Visit(v reflect.Value) {
	visitor := ctx.impl
	if !visitor.Visit(ctx, v) {
		return
	}

	// Special case for reflect.Value.
	if ctx.flags.Has(VisitReflectValues) && v.Type() == ReflectValueType {
		ctx.Visit(v.Interface().(reflect.Value))
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
			ctx.Visit(v.Elem())
		}

	case reflect.UnsafePointer:
		visitor.VisitUnsafePointer(ctx, v)

	case reflect.Array:
		if visitor.VisitArray(ctx, v) {
			for i := 0; i < v.Len(); i++ {
				ctx.Visit(v.Index(i))
			}
		}

	case reflect.Slice:
		if visitor.VisitSlice(ctx, v) {
			if ctx.flags.Has(VisitSliceLenToCap) {
				v = MakeSlice(v.Type(), v.UnsafePointer(), v.Cap(), v.Cap())
			}
			for i := 0; i < v.Len(); i++ {
				ctx.Visit(v.Index(i))
			}
		}

	case reflect.Map:
		if visitor.VisitMap(ctx, v) && !v.IsNil() {
			iter := v.MapRange()
			for iter.Next() {
				ctx.Visit(iter.Key())
				ctx.Visit(iter.Value())
			}
		}

	case reflect.Chan:
		// TODO: visit buffered channel items if possible
		visitor.VisitChan(ctx, v)

	case reflect.Struct:
		if visitor.VisitStruct(ctx, v) {
			if ctx.flags.Has(VisitUnexportedFields) {
				unrestricted := StructValueOf(v)
				for i := 0; i < unrestricted.NumField(); i++ {
					ctx.Visit(unrestricted.Field(i))
				}
			} else {
				t := v.Type()
				for i := 0; i < v.NumField(); i++ {
					if ft := t.Field(i); ft.IsExported() {
						ctx.Visit(v.Field(i))
					}
				}
			}
		}

	case reflect.Func:
		if visitor.VisitFunc(ctx, v) && !v.IsNil() && ctx.flags.Has(VisitClosures) {
			if closure, ok := FuncValueOf(v).Closure(); ok {
				ctx.Visit(closure)
			}
		}

	case reflect.Interface:
		if visitor.VisitInterface(ctx, v) && !v.IsNil() {
			ctx.Visit(v.Elem())
		}

	default:
		panic("unreachable")
	}
}

// Fork creates a new Visitor context, linked to the current
// context and its location.
func (ctx VisitorContext) Fork(impl Visitor) VisitorContext {
	return VisitorContext{
		parent: &ctx,
		impl:   impl,
		flags:  ctx.flags,
	}
}

// DefaultVisitor is a Visitor that visits all values in a
// reflect.Value graph.
type DefaultVisitor struct{}

var _ Visitor = DefaultVisitor{}

func (DefaultVisitor) Visit(VisitorContext, reflect.Value) bool          { return true }
func (DefaultVisitor) VisitBool(VisitorContext, reflect.Value)           {}
func (DefaultVisitor) VisitInt(VisitorContext, reflect.Value)            {}
func (DefaultVisitor) VisitUint(VisitorContext, reflect.Value)           {}
func (DefaultVisitor) VisitFloat(VisitorContext, reflect.Value)          {}
func (DefaultVisitor) VisitComplex(VisitorContext, reflect.Value) bool   { return true }
func (DefaultVisitor) VisitString(VisitorContext, reflect.Value)         {}
func (DefaultVisitor) VisitPointer(VisitorContext, reflect.Value) bool   { return true }
func (DefaultVisitor) VisitUnsafePointer(VisitorContext, reflect.Value)  {}
func (DefaultVisitor) VisitArray(VisitorContext, reflect.Value) bool     { return true }
func (DefaultVisitor) VisitSlice(VisitorContext, reflect.Value) bool     { return true }
func (DefaultVisitor) VisitMap(VisitorContext, reflect.Value) bool       { return true }
func (DefaultVisitor) VisitChan(VisitorContext, reflect.Value) bool      { return true }
func (DefaultVisitor) VisitStruct(VisitorContext, reflect.Value) bool    { return true }
func (DefaultVisitor) VisitFunc(VisitorContext, reflect.Value) bool      { return true }
func (DefaultVisitor) VisitInterface(VisitorContext, reflect.Value) bool { return true }
