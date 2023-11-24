package types

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"

	coroutinev1 "github.com/stealthrocket/coroutine/gen/proto/go/coroutine/v1"
)

// Inspect inspects serialized durable coroutine state.
//
// The input should be a buffer produced by (*coroutine.Context).Marshal
// or by types.Serialize.
func Inspect(b []byte) (*State, error) {
	var state coroutinev1.State
	if err := state.UnmarshalVT(b); err != nil {
		return nil, err
	}
	return &State{state: &state}, nil
}

// State wraps durable coroutine state.
type State struct {
	state *coroutinev1.State
}

// BuildID returns the build ID of the program that generated this state.
func (s *State) BuildID() string {
	return s.state.Build.Id
}

// OS returns the operating system the coroutine was compiled for.
func (s *State) OS() string {
	return s.state.Build.Os
}

// Arch returns the architecture the coroutine was compiled for.
func (s *State) Arch() string {
	return s.state.Build.Arch
}

// NumType returns the number of types referenced by the coroutine.
func (s *State) NumType() int {
	return len(s.state.Types)
}

// Type returns a type by index.
func (s *State) Type(i int) *Type {
	if i < 0 || i >= len(s.state.Types) {
		panic(fmt.Sprintf("type %d not found", i))
	}
	return &Type{
		state: s,
		typ:   s.state.Types[i],
		index: i,
	}
}

// NumFunction returns the number of functions/methods/closures
// referenced by the coroutine.
func (s *State) NumFunction() int {
	return len(s.state.Functions)
}

// Function returns a function by index.
func (s *State) Function(i int) *Function {
	if i < 0 || i >= len(s.state.Functions) {
		panic(fmt.Sprintf("function %d not found", i))
	}
	return &Function{
		state:    s,
		function: s.state.Functions[i],
		index:    i,
	}
}

// NumRegion returns the number of memory regions referenced by the
// coroutine.
func (s *State) NumRegion() int {
	return len(s.state.Regions)
}

// Region retrieves a region by index.
func (s *State) Region(i int) *Region {
	if i < 0 || i >= len(s.state.Regions) {
		panic(fmt.Sprintf("region %d not found", i))
	}
	return &Region{
		state:  s,
		region: s.state.Regions[i],
		index:  i,
	}
}

// NumString returns the number of strings referenced by types.
func (s *State) NumString() int {
	return len(s.state.Strings)
}

// String retrieves a string by index.
func (s *State) String(i int) string {
	if i < 0 || i >= len(s.state.Strings) {
		panic(fmt.Sprintf("string %d not found", i))
	}
	return s.state.Strings[i]
}

// Root is the root object that was serialized.
func (s *State) Root() *Region {
	return &Region{
		state:  s,
		region: s.state.Root,
		index:  -1,
	}
}

// Type is a type referenced by a durable coroutine.
type Type struct {
	state *State
	typ   *coroutinev1.Type
	index int
}

// Index is the index of the type in the serialized state, or -1
// if the type is derived from a serialized type.
func (t *Type) Index() int {
	return t.index
}

// Name is the name of the type within the package it was defined.
func (t *Type) Name() string {
	if t.typ.Name == 0 {
		return ""
	}
	return t.state.String(int(t.typ.Name - 1))
}

// Package is the name of the package that defines the type.
func (t *Type) Package() string {
	if t.typ.Package == 0 {
		return ""
	}
	return t.state.String(int(t.typ.Package - 1))
}

// Kind is the underlying kind for this type.
func (t *Type) Kind() reflect.Kind {
	switch t.typ.Kind {
	case coroutinev1.Kind_KIND_BOOL:
		return reflect.Bool
	case coroutinev1.Kind_KIND_INT:
		return reflect.Int
	case coroutinev1.Kind_KIND_INT8:
		return reflect.Int8
	case coroutinev1.Kind_KIND_INT16:
		return reflect.Int16
	case coroutinev1.Kind_KIND_INT32:
		return reflect.Int32
	case coroutinev1.Kind_KIND_INT64:
		return reflect.Int64
	case coroutinev1.Kind_KIND_UINT:
		return reflect.Uint
	case coroutinev1.Kind_KIND_UINT8:
		return reflect.Uint8
	case coroutinev1.Kind_KIND_UINT16:
		return reflect.Uint16
	case coroutinev1.Kind_KIND_UINT32:
		return reflect.Uint32
	case coroutinev1.Kind_KIND_UINT64:
		return reflect.Uint64
	case coroutinev1.Kind_KIND_UINTPTR:
		return reflect.Uintptr
	case coroutinev1.Kind_KIND_FLOAT32:
		return reflect.Float32
	case coroutinev1.Kind_KIND_FLOAT64:
		return reflect.Float64
	case coroutinev1.Kind_KIND_COMPLEX64:
		return reflect.Complex64
	case coroutinev1.Kind_KIND_COMPLEX128:
		return reflect.Complex128
	case coroutinev1.Kind_KIND_ARRAY:
		return reflect.Array
	case coroutinev1.Kind_KIND_CHAN:
		return reflect.Chan
	case coroutinev1.Kind_KIND_FUNC:
		return reflect.Func
	case coroutinev1.Kind_KIND_INTERFACE:
		return reflect.Interface
	case coroutinev1.Kind_KIND_MAP:
		return reflect.Map
	case coroutinev1.Kind_KIND_POINTER:
		return reflect.Pointer
	case coroutinev1.Kind_KIND_SLICE:
		return reflect.Slice
	case coroutinev1.Kind_KIND_STRING:
		return reflect.String
	case coroutinev1.Kind_KIND_STRUCT:
		return reflect.Struct
	case coroutinev1.Kind_KIND_UNSAFE_POINTER:
		return reflect.UnsafePointer
	default:
		panic(fmt.Sprintf("invalid type kind %s", t.typ.Kind))
	}
}

// Elem is the type of an array, slice, pointer, chan or map's element.
func (t *Type) Elem() *Type {
	if t.typ.Elem == 0 {
		return nil
	}
	return t.state.Type(int(t.typ.Elem - 1))
}

// Key is the key type for map types.
func (t *Type) Key() *Type {
	if t.typ.Key == 0 {
		return nil
	}
	return t.state.Type(int(t.typ.Key - 1))
}

// NumField is the number of struct fields for struct types.
func (t *Type) NumField() int {
	return len(t.typ.Fields)
}

// Field retrieves a struct field by index.
func (t *Type) Field(i int) *Field {
	if i < 0 || i >= len(t.typ.Fields) {
		return nil
	}
	return &Field{
		state: t.state,
		field: t.typ.Fields[i],
	}
}

// NumParam is the number of parameters for function types.
func (t *Type) NumParam() int {
	return len(t.typ.Params)
}

// Param is the type of a function parameter with the specified index.
func (t *Type) Param(i int) *Type {
	if i < 0 || i >= len(t.typ.Params) {
		return nil
	}
	return t.state.Type(int(t.typ.Params[i] - 1))
}

// NumResult is the number of results for function types.
func (t *Type) NumResult() int {
	return len(t.typ.Results)
}

// Result is the type of a function result with the specified index.
func (t *Type) Result(i int) *Type {
	if i < 0 || i >= len(t.typ.Results) {
		return nil
	}
	return t.state.Type(int(t.typ.Results[i] - 1))
}

// Len is the length of an array type.
func (t *Type) Len() int {
	return int(t.typ.Length)
}

// MemoryOffset is the location of this type in memory.
//
// The offset is only applicable to the build that generated the state.
func (t *Type) MemoryOffset() uint64 {
	return t.typ.MemoryOffset
}

// ChanDir is the direction of a channel type.
func (t *Type) ChanDir() reflect.ChanDir {
	switch t.typ.ChanDir {
	case coroutinev1.ChanDir_CHAN_DIR_RECV:
		return reflect.RecvDir
	case coroutinev1.ChanDir_CHAN_DIR_SEND:
		return reflect.SendDir
	case coroutinev1.ChanDir_CHAN_DIR_BOTH:
		return reflect.BothDir
	default:
		panic(fmt.Sprintf("invalid chan dir %s", t.typ.ChanDir))
	}
}

// Variadic is true for function types with a variadic final argument.
func (t *Type) Variadic() bool {
	return t.typ.Variadic
}

// Opaue is true for types that had a custom serializer registered
// in the program that generated the coroutine state. Custom types
// are opaque and cannot be inspected.
func (t *Type) Opaque() bool {
	return t.typ.CustomSerializer > 0
}

// Format implements fmt.Formatter.
func (t *Type) Format(s fmt.State, v rune) {
	name := t.Name()
	if pkg := t.Package(); pkg != "" {
		if name == "" {
			name = fmt.Sprintf("<anon %s>", t.Kind())
		}
		name = pkg + "." + name
	}

	if t.Opaque() {
		if name == "" {
			name = fmt.Sprintf("<anon %s>", t.Kind())
		}
		if t.typ.Kind == coroutinev1.Kind_KIND_POINTER {
			name = "*" + name
		}
		s.Write([]byte(name))
		return
	}

	verbose := s.Flag('+') || s.Flag('#')
	if name != "" && !verbose {
		s.Write([]byte(name))
		return
	}

	var primitiveKind string
	switch t.typ.Kind {
	case coroutinev1.Kind_KIND_BOOL:
		primitiveKind = "bool"
	case coroutinev1.Kind_KIND_INT:
		primitiveKind = "int"
	case coroutinev1.Kind_KIND_INT8:
		primitiveKind = "int8"
	case coroutinev1.Kind_KIND_INT16:
		primitiveKind = "int16"
	case coroutinev1.Kind_KIND_INT32:
		primitiveKind = "int32"
	case coroutinev1.Kind_KIND_INT64:
		primitiveKind = "int64"
	case coroutinev1.Kind_KIND_UINT:
		primitiveKind = "uint"
	case coroutinev1.Kind_KIND_UINT8:
		primitiveKind = "uint8"
	case coroutinev1.Kind_KIND_UINT16:
		primitiveKind = "uint16"
	case coroutinev1.Kind_KIND_UINT32:
		primitiveKind = "uint32"
	case coroutinev1.Kind_KIND_UINT64:
		primitiveKind = "uint64"
	case coroutinev1.Kind_KIND_UINTPTR:
		primitiveKind = "uintptr"
	case coroutinev1.Kind_KIND_FLOAT32:
		primitiveKind = "float32"
	case coroutinev1.Kind_KIND_FLOAT64:
		primitiveKind = "float64"
	case coroutinev1.Kind_KIND_COMPLEX64:
		primitiveKind = "complex64"
	case coroutinev1.Kind_KIND_COMPLEX128:
		primitiveKind = "complex128"
	case coroutinev1.Kind_KIND_STRING:
		primitiveKind = "string"
	case coroutinev1.Kind_KIND_INTERFACE:
		primitiveKind = "interface"
	case coroutinev1.Kind_KIND_UNSAFE_POINTER:
		primitiveKind = "unsafe.Pointer"
	}
	if primitiveKind != "" {
		if name == primitiveKind {
			name = ""
		}
		var result string
		switch {
		case (name == "error" && primitiveKind == "interface") ||
			(name == "any" && primitiveKind == "interface") ||
			(name == "byte" && primitiveKind == "uint8") ||
			(name == "rune" && primitiveKind == "int32"):
			result = name
		case name != "":
			result = fmt.Sprintf("(%s=%s)", name, primitiveKind)
		default:
			result = primitiveKind
		}
		s.Write([]byte(result))
		return
	}

	var elemPrefix string
	switch t.typ.Kind {
	case coroutinev1.Kind_KIND_ARRAY:
		elemPrefix = fmt.Sprintf("[%d]", t.Len())
	case coroutinev1.Kind_KIND_CHAN:
		switch t.typ.ChanDir {
		case coroutinev1.ChanDir_CHAN_DIR_RECV:
			elemPrefix = "<-chan "
		case coroutinev1.ChanDir_CHAN_DIR_SEND:
			elemPrefix = "chan<- "
		default:
			elemPrefix = "chan "
		}
	case coroutinev1.Kind_KIND_POINTER:
		elemPrefix = "*"
	case coroutinev1.Kind_KIND_SLICE:
		elemPrefix = "[]"
	}
	if elemPrefix != "" {
		if name != "" {
			elemPrefix = fmt.Sprintf("(%s=%s", name, elemPrefix)
		}
		s.Write([]byte(elemPrefix))
		t.Elem().Format(withoutFlags{s}, v)
		if name != "" {
			s.Write([]byte(")"))
		}
		return
	}

	if name != "" {
		s.Write([]byte(fmt.Sprintf("(%s=", name)))
	}
	switch t.typ.Kind {
	case coroutinev1.Kind_KIND_FUNC:
		s.Write([]byte("func("))
		paramCount := t.NumParam()
		for i := 0; i < paramCount; i++ {
			if i > 0 {
				s.Write([]byte(", "))
			}
			if i == paramCount-1 && t.Variadic() {
				s.Write([]byte("..."))
			}
			t.Param(i).Format(withoutFlags{s}, v)
		}
		s.Write([]byte(")"))
		n := t.NumResult()
		if n > 0 {
			s.Write([]byte(" "))
		}
		if n > 1 {
			s.Write([]byte("("))
		}
		for i := 0; i < n; i++ {
			if i > 0 {
				s.Write([]byte(", "))
			}
			t.Result(i).Format(withoutFlags{s}, v)
		}
		if n > 1 {
			s.Write([]byte(")"))
		}
		if name != "" {
			s.Write([]byte(")"))
		}

	case coroutinev1.Kind_KIND_MAP:
		s.Write([]byte("map["))
		t.Key().Format(withoutFlags{s}, v)
		s.Write([]byte("]"))
		t.Elem().Format(withoutFlags{s}, v)

	case coroutinev1.Kind_KIND_STRUCT:
		n := t.NumField()
		if n == 0 {
			s.Write([]byte("struct{}"))
		} else {
			s.Write([]byte("struct{ "))
			for i := 0; i < n; i++ {
				if i > 0 {
					s.Write([]byte("; "))
				}
				f := t.Field(i)
				if !f.Anonymous() {
					s.Write([]byte(f.Name()))
					s.Write([]byte(" "))
				}
				f.Type().Format(withoutFlags{State: s}, v)
			}
			s.Write([]byte(" }"))
		}

	default:
		s.Write([]byte("invalid"))
	}
	if name != "" {
		s.Write([]byte(")"))
	}
}

type withoutFlags struct{ fmt.State }

func (withoutFlags) Flag(c int) bool { return false }

// Field is a struct field.
type Field struct {
	state *State
	field *coroutinev1.Field
}

// Name is the name of the field.
func (f *Field) Name() string {
	if f.field.Name == 0 {
		return ""
	}
	return f.state.String(int(f.field.Name - 1))
}

// Package is the package path that qualifies a lwer case (unexported)
// field name. It is empty for upper case (exported) field names.
func (f *Field) Package() string {
	if f.field.Package == 0 {
		return ""
	}
	return f.state.String(int(f.field.Package - 1))
}

// Type is the type of the field.
func (f *Field) Type() *Type {
	return f.state.Type(int(f.field.Type - 1))
}

// Offset is the offset of the field within its struct, in bytes.
func (f *Field) Offset() uint64 {
	return f.field.Offset
}

// Anonymous is true of the field is an embedded field (with a name
// derived from its type).
func (f *Field) Anonymous() bool {
	return f.field.Anonymous
}

// Tag contains struct field metadata.
func (f *Field) Tag() reflect.StructTag {
	return reflect.StructTag(f.field.Tag)
}

// Function is a function, method or closure referenced by the coroutine.
type Function struct {
	state    *State
	function *coroutinev1.Function
	index    int
}

// Name is the name of the function.
func (f *Function) Name() string {
	if f.function.Name == 0 {
		return ""
	}
	return f.state.String(int(f.function.Name - 1))
}

// Index is the index of the function in the serialized state.
func (f *Function) Index() int {
	return f.index
}

// Type is the type of the function.
func (f *Function) Type() *Type {
	return f.state.Type(int(f.function.Type - 1))
}

// ClosureType returns the memory layout for closure functions.
//
// The returned type is a struct where the first field is a function
// pointer and the remaining fields are the variables from outer scopes
// that are referenced by the closure.
//
// Nil is returned for functions that are not closures.
func (f *Function) ClosureType() *Type {
	if f.function.Closure == 0 {
		return nil
	}
	return f.state.Type(int(f.function.Closure - 1))
}

// String is the name of the function.
func (f *Function) String() string {
	return f.Name()
}

// Region is a region of memory referenced by the coroutine.
type Region struct {
	state  *State
	region *coroutinev1.Region
	index  int
}

// Index is the index of the region in the serialized state,
// or -1 if this is the root region.
func (t *Region) Index() int {
	return t.index
}

// Type is the type of the region.
func (r *Region) Type() *Type {
	t := r.state.Type(int((r.region.Type >> 1) - 1))
	if r.region.Type&1 == 1 {
		t = newArrayType(r.state, int64(r.region.ArrayLength), t)
	}
	return t
}

func newArrayType(state *State, length int64, t *Type) *Type {
	idx := t.Index()
	if idx < 0 {
		panic("BUG")
	}
	return &Type{
		state: state,
		typ: &coroutinev1.Type{
			Kind:   coroutinev1.Kind_KIND_ARRAY,
			Length: int64(length),
			Elem:   uint32(idx + 1),
		},
		index: -1, // aka. a derived type
	}
}

// Size is the size of the region in bytes.
func (r *Region) Size() int64 {
	return int64(len(r.region.Data))
}

// String is a summary of the region in string form.
func (r *Region) String() string {
	return fmt.Sprintf("Region(%d byte(s), %#v)", len(r.region.Data), r.Type())
}

// Scan returns an region scanner.
func (r *Region) Scan() *Scanner {
	return &Scanner{
		state: r.state,
		src:   r,
		data:  r.region.Data,
	}
}

// Scanner scans a Region.
type Scanner struct {
	state *State

	src   *Region
	data  []byte
	pos   int
	stack []scanstep
	err   error
	done  bool

	// set during iteration
	kind     reflect.Kind
	region   *Region
	offset   int64
	typ      *Type
	field    *Field
	function *Function
	nil      bool
	len      int
	cap      int
	data1    uint64
	data2    uint64
	custom   bool
}

type scanstep struct {
	st        scantype
	idx       int
	len       int
	customtil uint64
	typ       *Type
}

type scantype int

const (
	scanprimitive scantype = iota
	scanarray
	scanstruct
	scanmap
	scanclosure
	scancustom
)

// Next is true if there is more to scan.
func (s *Scanner) Next() bool {
	if s.err != nil || s.done || s.pos >= len(s.data) {
		return false
	}

	s.kind = reflect.Invalid
	s.region = nil
	s.offset = 0
	s.typ = nil
	s.field = nil
	s.function = nil
	s.nil = false
	s.len = 0
	s.cap = 0
	s.data1 = 0
	s.data2 = 0

	if len(s.stack) == 0 { // init
		return s.readAny(s.src.Type(), 0)
	}

	for len(s.stack) > 0 {
		last := &s.stack[len(s.stack)-1]
		switch last.st {
		case scanprimitive:

		case scanarray:
			last.idx++
			if last.idx < last.len {
				return s.readAny(last.typ.Elem(), len(s.stack))
			}

		case scanstruct:
			last.idx++
			if last.idx < last.len {
				s.field = last.typ.Field(last.idx)
				return s.readAny(s.field.Type(), len(s.stack))
			}

		case scanmap:
			last.idx++
			if last.idx < last.len {
				var t *Type
				if last.idx%2 == 0 {
					t = last.typ.Key()
				} else {
					t = last.typ.Elem()
				}
				return s.readAny(t, len(s.stack))
			}

		case scanclosure:
			if last.typ != nil {
				ct := last.typ
				last.typ = nil // only read closure struct once
				s.kind = reflect.Struct
				s.typ = ct
				return s.readStruct(ct, 1)
			}

		case scancustom:
			if uint64(s.pos) < last.customtil {
				if !s.readType() {
					return false
				}
				return s.readAny(s.typ, len(s.stack))
			}
			if uint64(s.pos) > last.customtil {
				s.err = fmt.Errorf("invalid custom object size")
				return false
			}
			s.custom = false
		}
		s.stack = s.stack[:len(s.stack)-1] // pop
	}

	if s.pos != len(s.data) {
		s.err = fmt.Errorf("trailing bytes")
	} else {
		s.done = true // prevent re-init
	}
	return false
}

// Pos is the position of the scanner, in terms of number of bytes into
// the region.
func (s *Scanner) Pos() int {
	return s.pos
}

// Depth is the depth of the scan stack.
func (s *Scanner) Depth() int {
	return len(s.stack)
}

// Kind is the kind of entity the scanner is pointing to.
func (s *Scanner) Kind() reflect.Kind {
	return s.kind
}

// Region is the region and offset the scanner is pointing to.
func (s *Scanner) Region() (*Region, int64) {
	return s.region, s.offset
}

// Type is the type the scanner is pointing to.
func (s *Scanner) Type() *Type {
	return s.typ
}

// Field is the field the scanner is pointing to.
func (s *Scanner) Field() *Field {
	return s.field
}

// Function is the function the scanner is pointing to.
func (s *Scanner) Function() *Function {
	return s.function
}

// Custom is true if the scanner is scanning an object for
// which a custom serializer was registered.
func (s *Scanner) Custom() bool {
	return s.custom
}

// Nil is true if the scanner is pointing to nil.
func (s *Scanner) Nil() bool {
	return s.nil
}

// Len is the length of the string, slice, array or map
// the scanner is pointing to.
func (s *Scanner) Len() int {
	return s.len
}

// Cap is the capacity of the slice the scanner is pointing to.
func (s *Scanner) Cap() int {
	return s.cap
}

// Bool returns the bool the scanner points to.
func (s *Scanner) Bool() bool {
	return s.data1 == 1
}

// Int returns the int the scanner points to.
func (s *Scanner) Int() int {
	return int(s.data1)
}

// Int8 returns the int8 the scanner points to.
func (s *Scanner) Int8() int8 {
	return int8(s.data1)
}

// Int16 returns the int16 the scanner points to.
func (s *Scanner) Int16() int16 {
	return int16(s.data1)
}

// Int32 returns the int32 the scanner points to.
func (s *Scanner) Int32() int32 {
	return int32(s.data1)
}

// Int64 returns the int64 the scanner points to.
func (s *Scanner) Int64() int64 {
	return int64(s.data1)
}

// Uint returns the uint8 the scanner points to.
func (s *Scanner) Uint() uint {
	return uint(s.data1)
}

// Uint8 returns the uint8 the scanner points to.
func (s *Scanner) Uint8() uint8 {
	return uint8(s.data1)
}

// Uint16 returns the uint16 the scanner points to.
func (s *Scanner) Uint16() uint16 {
	return uint16(s.data1)
}

// Uint32 returns the uint32 the scanner points to.
func (s *Scanner) Uint32() uint32 {
	return uint32(s.data1)
}

// Uint64 returns the uint64 the scanner points to.
func (s *Scanner) Uint64() uint64 {
	return s.data1
}

// Uintptr returns the uintptr the scanner points to.
func (s *Scanner) Uintptr() uintptr {
	return uintptr(s.data1)
}

// Float32 returns the float32 the scanner points to.
func (s *Scanner) Float32() float32 {
	return math.Float32frombits(uint32(s.data1))
}

// Float64 returns the float64 the scanner points to.
func (s *Scanner) Float64() float64 {
	return math.Float64frombits(s.data1)
}

// Complex64 returns the complex64 the scanner points to.
func (s *Scanner) Complex64() complex64 {
	r := math.Float32frombits(uint32(s.data1))
	i := math.Float32frombits(uint32(s.data2))
	return complex(r, i)
}

// Complex128 returns the complex128 the scanner points to.
func (s *Scanner) Complex128() complex128 {
	r := math.Float64frombits(s.data1)
	i := math.Float64frombits(s.data2)
	return complex(r, i)
}

// Close closes the scanner and returns any errors that occurred during scanning.
func (s *Scanner) Close() error {
	return s.err
}

func (s *Scanner) readAny(t *Type, depth int) (ok bool) {
	s.typ = t
	s.kind = t.Kind()

	if depth == 0 && t.Kind() == reflect.Map {
		// Map regions encode the contents of a map.
		// When maps are seen within a region (nested in another
		// object), a reference to the map region is stored instead.
		// Handle the first case here, and the reference case below.
		return s.readMap()
	}

	if t.Opaque() {
		return s.readCustom()
	}

	switch t.Kind() {
	case reflect.Array:
		return s.readArray(t)
	case reflect.Struct:
		if t.Package() == "reflect" {
			panic("not implemented: reflection")
		}
		return s.readStruct(t, 0)
	case reflect.Func:
		return s.readFunc(t)
	case reflect.Chan:
		panic("not implemented: channels")
	}

	s.stack = append(s.stack, scanstep{st: scanprimitive})

	switch t.Kind() {
	case reflect.Uint8, reflect.Int8, reflect.Bool:
		return s.readUint8()
	case reflect.Uint16, reflect.Int16:
		return s.readUint16()
	case reflect.Uint32, reflect.Int32, reflect.Float32:
		return s.readUint32()
	case reflect.Uint64, reflect.Int64, reflect.Uint, reflect.Int, reflect.Float64:
		return s.readUint64()
	case reflect.Complex64:
		return s.readComplex64()
	case reflect.Complex128:
		return s.readComplex128()
	case reflect.String:
		return s.readString()
	case reflect.Slice:
		return s.readSlice()
	case reflect.Pointer, reflect.UnsafePointer, reflect.Map: // references
		return s.readRegionPointer()
	case reflect.Interface:
		return s.readInterface()
	default:
		panic("not implemented")
	}
}

func (s *Scanner) readType() (ok bool) {
	id, ok := s.getVarint()
	if !ok {
		return false
	}
	t := s.state.Type(int(id - 1))

	len, ok := s.getVarint()
	if !ok {
		return false
	}
	if len >= 0 {
		t = newArrayType(s.state, len, t)
	}
	s.typ = t
	return true
}

func (s *Scanner) readUint8() (ok bool) {
	s.data1 = uint64(s.data[s.pos])
	s.pos++
	return true
}

func (s *Scanner) readUint16() (ok bool) {
	if len(s.data)-s.pos < 2 {
		return false
	}
	s.data1 = uint64(binary.LittleEndian.Uint16(s.data[s.pos:]))
	s.pos += 2
	return true
}

func (s *Scanner) readUint32() (ok bool) {
	if len(s.data)-s.pos < 4 {
		return false
	}
	s.data1 = uint64(binary.LittleEndian.Uint32(s.data[s.pos:]))
	s.pos += 4
	return true
}

func (s *Scanner) readUint64() (ok bool) {
	if len(s.data)-s.pos < 8 {
		return false
	}
	s.data1 = uint64(binary.LittleEndian.Uint64(s.data[s.pos:]))
	s.pos += 8
	return true
}

func (s *Scanner) readComplex64() (ok bool) {
	if len(s.data)-s.pos < 8 {
		return false
	}
	s.data1 = uint64(binary.LittleEndian.Uint32(s.data[s.pos:]))
	s.data2 = uint64(binary.LittleEndian.Uint32(s.data[s.pos+4:]))
	s.pos += 8
	return true
}

func (s *Scanner) readComplex128() (ok bool) {
	if len(s.data)-s.pos < 16 {
		return false
	}
	s.data1 = binary.LittleEndian.Uint64(s.data[s.pos:])
	s.data2 = binary.LittleEndian.Uint64(s.data[s.pos+8:])
	s.pos += 16
	return true
}

func (s *Scanner) readString() (ok bool) {
	n, ok := s.getVarint()
	if !ok {
		return ok
	}
	s.len = int(n)
	if s.len == 0 {
		return true
	}
	return s.readRegionPointer()
}

func (s *Scanner) readSlice() (ok bool) {
	n, ok := s.getVarint()
	if !ok {
		return ok
	}
	s.len = int(n)

	n, ok = s.getVarint()
	if !ok {
		return ok
	}
	s.cap = int(n)

	return s.readRegionPointer()
}

func (s *Scanner) readArray(t *Type) (ok bool) {
	s.len = t.Len()
	s.stack = append(s.stack, scanstep{
		st:  scanarray,
		idx: -1,
		len: s.len,
		typ: t,
	})
	return true
}

func (s *Scanner) readStruct(t *Type, fromField int) (ok bool) {
	s.stack = append(s.stack, scanstep{
		st:  scanstruct,
		idx: fromField - 1,
		len: t.NumField(),
		typ: t,
	})
	return true
}

func (s *Scanner) readFunc(t *Type) (ok bool) {
	id, ok := s.getVarint()
	if !ok {
		return false
	}
	if id == 0 {
		s.nil = true
		return true
	}
	s.function = s.state.Function(int(id - 1))

	ct := s.function.ClosureType()
	if ct != nil {
		s.stack = append(s.stack, scanstep{
			st:  scanclosure,
			typ: ct,
		})
	}
	return true
}

func (s *Scanner) readMap() (ok bool) {
	n, ok := s.getVarint()
	if !ok {
		return false
	}
	s.len = int(n)

	t := s.src.Type()
	if len(s.stack) > 0 || t.Kind() != reflect.Map {
		panic("unexpected inline map")
	}

	s.stack = append(s.stack, scanstep{
		st:  scanmap,
		idx: -1,
		len: int(n * 2),
		typ: t,
	})
	return true
}

func (s *Scanner) readInterface() (ok bool) {
	nonNil := s.getBool()
	if !nonNil {
		s.nil = true
		return true
	}
	if !s.readType() {
		return false
	}
	return s.readRegionPointer()
}

func (s *Scanner) readRegionPointer() (ok bool) {
	tag, ok := s.getVarint()
	if !ok {
		return false
	}
	if tag == 0 {
		s.nil = true
		return true
	}
	if tag == -1 { // static
		offset, ok := s.getVarint()
		if !ok {
			return false
		}
		s.data1 = uint64(offset)
		return true
	}
	s.region = s.state.Region(int(tag - 1))

	offset, ok := s.getVarint()
	if !ok {
		return false
	}
	s.offset = offset
	return true
}

func (s *Scanner) readCustom() (ok bool) {
	s.custom = true
	if len(s.data)-s.pos < 8 {
		s.err = io.ErrShortBuffer
		return false
	}
	size := binary.LittleEndian.Uint64(s.data[s.pos:])
	if uint64(s.pos)+size > uint64(len(s.data)) {
		s.err = fmt.Errorf("invalid custom object size")
		return false
	}
	s.stack = append(s.stack, scanstep{
		st:        scancustom,
		customtil: uint64(s.pos) + size,
	})
	s.pos += 8
	return true
}

func (s *Scanner) getBool() bool {
	// loop invariant: s.pos < len(s.data)
	value := s.data[s.pos] == 1
	s.pos++
	return value
}

func (s *Scanner) getVarint() (value int64, ok bool) {
	// loop invariant: s.pos < len(s.data)
	var n int
	value, n = binary.Varint(s.data[s.pos:])
	if n <= 0 {
		s.err = io.ErrShortBuffer
		return
	}
	s.pos += n
	return value, true
}
