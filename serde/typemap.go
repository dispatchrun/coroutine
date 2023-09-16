package serde

import (
	"github.com/stealthrocket/coroutine/internal/serde"
)

// SerializerFn is the signature of custom serializer functions.
type SerializerFn[T any] func(*Serializer, *T) error

// DeserializerFn is the signature of customer deserializer functions.
type DeserializerFn[T any] func(*Deserializer, *T) error

// RegisterType adds T to the global type register, as well as *T and all
// types contained within.
//
// Types can be registered multiple times, but care should be taken to always
// register them in a deterministic order; init functions are a good place for
// that. Most of the time, coroc takes care of registering types.
func RegisterType[T any]() {
	serde.RegisterType[T]()
}

// RegisterTypeWithSerde adds T to the global type register in the same way
// [RegisterType] does, but also attaches custom serialization and
// deserialization functions to T.
//
// If T already has custom serialization and deserialization functions,
// [RegisterTypeWithSerde] panics.
func RegisterTypeWithSerde[T any](ser SerializerFn[T], des DeserializerFn[T]) {
	serde.RegisterTypeWithSerde[T](ser, des)
}
