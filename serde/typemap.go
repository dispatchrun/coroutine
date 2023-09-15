package serde

import (
	"github.com/stealthrocket/coroutine/internal/serde"
)

type SerializerFn[T any] func(*Serializer, *T) error
type DeserializerFn[T any] func(*Deserializer, *T) error

func RegisterType[T any]() {
	serde.RegisterType[T]()
}

func RegisterTypeWithSerde[T any](ser SerializerFn[T], des DeserializerFn[T]) {
	serde.RegisterTypeWithSerde[T](ser, des)
}
