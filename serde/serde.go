package serde

import (
	"reflect"
	"unsafe"

	"github.com/stealthrocket/coroutine/internal/serde"
)

// Serialize is a blah.
type Serializer = serde.Serializer

// TODO
type Deserializer = serde.Deserializer

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

func DeserializeTo[T any](d *Deserializer, x *T) {
	r := reflect.ValueOf(x)
	t := r.Type().Elem()
	p := r.UnsafePointer()
	serde.DeserializeAny(d, t, p)
}
