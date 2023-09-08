package serde

import (
	"fmt"
	"reflect"
	"unsafe"
)

// reflect.go contains the reflection based serialization and deserialization
// procedures. It is used to handle values the codegen pass cannot generate
// methods for (interfaces). Eventually codegen should be able to generate code
// for types.

type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func ifacePointer(v interface{}) unsafe.Pointer {
	return (*iface)(unsafe.Pointer(&v)).ptr
}

func serializeAny(s *Serializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	switch t.Kind() {
	case reflect.Int:
		return SerializeInt(*(*int)(p), b)
	default:
		panic(fmt.Errorf("reflection cannot serialize type %s", t))
	}
}

func deserializeAny(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	switch t.Kind() {
	case reflect.Int:
		x, b := DeserializeInt(b)
		*(*int)(p) = x
		return b
	default:
		panic(fmt.Errorf("reflection cannot deserialize type %s", t))
	}
}
