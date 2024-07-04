package reflectext

import (
	"reflect"
	"unsafe"
)

var (
	AnyType = reflect.TypeFor[any]()

	BoolType = reflect.TypeFor[bool]()

	IntType   = reflect.TypeFor[int]()
	Int8Type  = reflect.TypeFor[int8]()
	Int16Type = reflect.TypeFor[int16]()
	Int32Type = reflect.TypeFor[int32]()
	Int64Type = reflect.TypeFor[int64]()

	UintType   = reflect.TypeFor[uint]()
	Uint8Type  = reflect.TypeFor[uint8]()
	Uint16Type = reflect.TypeFor[uint16]()
	Uint32Type = reflect.TypeFor[uint32]()
	Uint64Type = reflect.TypeFor[uint64]()

	Float32Type = reflect.TypeFor[float32]()
	Float64Type = reflect.TypeFor[float64]()

	Complex64Type  = reflect.TypeFor[complex64]()
	Complex128Type = reflect.TypeFor[complex128]()

	ByteType   = reflect.TypeFor[byte]()
	StringType = reflect.TypeFor[string]()

	UintptrType       = reflect.TypeFor[uintptr]()
	UnsafePointerType = reflect.TypeFor[unsafe.Pointer]()

	ReflectValueType = reflect.TypeFor[reflect.Value]()
	ReflectTypeType  = reflect.TypeFor[reflect.Type]()
)
