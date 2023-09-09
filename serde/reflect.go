package serde

import (
	"fmt"
	"log/slog"
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

	codec, exists := tm.CodecOf(t)
	if exists {
		x := reflect.NewAt(t, p).Elem().Interface()
		slog.Debug("using codec", "codec", codec, "type", t)
		return codec.serializer(s, x, b)
	} else {
		slog.Debug("type has not codec", "type", t)
	}

	switch t.Kind() {
	case reflect.Int:
		return SerializeInt(s, *(*int)(p), b)
	default:
		panic(fmt.Errorf("reflection cannot serialize type %s", t))
	}
}

func deserializeAny(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	codec, exists := tm.CodecOf(t)
	if exists {
		v, b := codec.deserializer(d, b)
		rv := reflect.ValueOf(v)
		target := reflect.NewAt(t, p)
		target.Elem().Set(rv)
		return b
	}

	switch t.Kind() {
	case reflect.Int:
		x, b := DeserializeInt(d, b)
		*(*int)(p) = x
		return b
	default:
		panic(fmt.Errorf("reflection cannot deserialize type %s", t))
	}
}
