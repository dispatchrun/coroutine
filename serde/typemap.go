package serde

import (
	"github.com/stealthrocket/coroutine/internal/serde"
)

type SerializerFn[T any] func(*T, []byte) ([]byte, error)
type DeserializerFn[T any] func(*T, []byte) ([]byte, error)

func RegisterType[T any]() {
	serde.RegisterType[T]()
}

func RegisterTypeWithSerde[T any](ser SerializerFn[T], des DeserializerFn[T]) {
	serde.RegisterTypeWithSerde[T](ser, des)
}
