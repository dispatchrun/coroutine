package serde

import (
	"github.com/stealthrocket/coroutine/internal/serde"
)

// SerializerFn is the signature of custom serializer functions.
type SerializerFn[T any] func(*Serializer, *T) error

// DeserializerFn is the signature of customer deserializer functions.
type DeserializerFn[T any] func(*Deserializer, *T) error

// RegisterSerde attaches custom serialization and deserialization functions to
// type T.
func RegisterSerde[T any](ser SerializerFn[T], des DeserializerFn[T]) {
	serde.RegisterSerde[T](ser, des)
}
