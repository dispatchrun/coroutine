package serde

import (
	"reflect"
	"unsafe"

	"github.com/stealthrocket/coroutine/internal/serde"
)

// Serializer holds the state during coroutine serialization.
type Serializer = serde.Serializer

// Deserializer holds the state during coroutine deserialization.
type Deserializer = serde.Deserializer

// Serialize a value.
func Serialize[T any](s *Serializer, x T) {
	var p unsafe.Pointer
	r := reflect.ValueOf(x)
	t := r.Type()
	if r.CanAddr() {
		p = r.Addr().UnsafePointer()
	} else {
		n := reflect.New(t)
		n.Elem().Set(r)
		p = n.UnsafePointer()
	}
	serde.SerializeAny(s, t, p)
}

// Deserialize a value to the provided non-nil pointer.
func DeserializeTo[T any](d *Deserializer, x *T) {
	r := reflect.ValueOf(x)
	t := r.Type().Elem()
	p := r.UnsafePointer()
	serde.DeserializeAny(d, t, p)
}
