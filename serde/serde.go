package serde

import (
	"github.com/stealthrocket/coroutine/internal/serde"
)

// Serializer holds the state during coroutine serialization.
type Serializer = serde.Serializer

// Deserializer holds the state during coroutine deserialization.
type Deserializer = serde.Deserializer

// Serialize a value.
func Serialize[T any](s *Serializer, x T) {
	serde.SerializeT(s, x)
}

// Deserialize a value to the provided non-nil pointer.
func DeserializeTo[T any](d *Deserializer, x *T) {
	serde.DeserializeTo(d, x)
}
