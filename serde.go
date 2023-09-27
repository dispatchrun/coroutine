package coroutine

import (
	"github.com/stealthrocket/coroutine/internal/serde"
)

// Serializer holds the state during coroutine serialization. It's an opaque
// value that needs to be passed throughout [Serialize] calls. See
// [RegisterSerde] for details.
type Serializer = serde.Serializer

// Deserializer holds the state during coroutine deserialization. It's an opaque
// value that needst to be passed throughout [DeserializeTo] calls. See
// [RegisterSerde] for details.
type Deserializer = serde.Deserializer

// Serialize a value. See [RegisterSerde].
func Serialize[T any](s *Serializer, x T) {
	serde.SerializeT(s, x)
}

// Deserialize a value to the provided non-nil pointer. See [RegisterSerde].
func DeserializeTo[T any](d *Deserializer, x *T) {
	serde.DeserializeTo(d, x)
}

// SerializerFn is the signature of custom serializer functions. Use the
// [Serialize] function to drive the [Serializer]. Returning an error results in
// the program panicking.
type SerializerFn[T any] func(*Serializer, *T) error

// DeserializerFn is the signature of customer deserializer functions. Use the
// [Deserialize] function to drive the [Deserializer]. Returning an error
// results in the program panicking.
type DeserializerFn[T any] func(*Deserializer, *T) error

// RegisterSerde attaches custom serialization and deserialization functions to
// type T.
//
// Coroutine state is serialized and deserialized when calling [Context.Marshal]
// and [Context.Unmarshal] respectively.
//
// Go basic types, structs, interfaces, slices, arrays, or any combination of
// them have built-in serialization and deserialization mechanisms. Channels and
// sync values do not.
//
// Custom serializer and deserializer functions can be attached to types using
// [RegisterSerde] to control how they are serialized, and possibly perform
// additional initialization on deserialization. Those functions are drivers for
// [Serializer] and [Deserializer], that need to invoke [Serialize] and
// [DeserializeTo] in order to actually perform serialization and
// deserialization operations. Pointers to the same address are detected as such
// to be reconstructed as pointing to the same value. Slices are serialized by
// first serializing their backing array, and then length and capacity. As a
// result, slices sharing the same backing array are deserialized into one array
// with two shared slices, just like the original state was. Elements between
// length and capacity are also preserved.
func RegisterSerde[T any](ser SerializerFn[T], des DeserializerFn[T]) {
	serde.RegisterSerde[T](ser, des)
}
