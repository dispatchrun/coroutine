package types

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Global type register.
var Types *TypeMap = NewTypeMap()

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
func RegisterSerde[T any](
	serializer SerializerFn[T],
	deserializer DeserializerFn[T]) {
	registerSerde[T](Types, serializer, deserializer)
}

func registerSerde[T any](tm *TypeMap,
	serializer func(*Serializer, *T) error,
	deserializer func(*Deserializer, *T) error) {

	t := reflect.TypeOf((*T)(nil)).Elem()

	s := func(s *Serializer, p unsafe.Pointer) {
		if err := serializer(s, (*T)(p)); err != nil {
			panic(fmt.Errorf("serializing %s: %w", t, err))
		}
	}

	d := func(d *Deserializer, p unsafe.Pointer) {
		if err := deserializer(d, (*T)(p)); err != nil {
			panic(fmt.Errorf("deserializing %s: %w", t, err))
		}
	}

	tm.Attach(t, s, d)
}

type serializerFn func(*Serializer, unsafe.Pointer)
type deserializerFn func(d *Deserializer, p unsafe.Pointer)

type serde struct {
	id  int
	ser serializerFn
	des deserializerFn
}

type TypeMap struct {
	custom []reflect.Type
	cache  doublemap[reflect.Type, *typeinfo]
	serdes map[reflect.Type]serde
}

func NewTypeMap() *TypeMap {
	m := &TypeMap{
		serdes: make(map[reflect.Type]serde),
	}
	return m
}

func (m *TypeMap) Attach(t reflect.Type, ser serializerFn, des deserializerFn) {
	if ser == nil || des == nil {
		panic("both serializer and deserializer need to be provided")
	}

	s, exists := m.serdes[t]
	if !exists {
		s.id = len(m.custom)
		m.custom = append(m.custom, t)
	}
	s.ser = ser
	s.des = des

	m.serdes[t] = s
}

func (m *TypeMap) serdeOf(x reflect.Type) (serde, bool) {
	s, ok := m.serdes[x]
	return s, ok
}

type doublemap[K, V comparable] struct {
	fromK map[K]V
	fromV map[V]K
}

func (m *doublemap[K, V]) GetK(k K) (V, bool) {
	v, ok := m.fromK[k]
	return v, ok
}

func (m *doublemap[K, V]) GetV(v V) (K, bool) {
	k, ok := m.fromV[v]
	return k, ok
}

func (m *doublemap[K, V]) Add(k K, v V) V {
	if m.fromK == nil {
		m.fromK = make(map[K]V)
		m.fromV = make(map[V]K)
	}
	m.fromK[k] = v
	m.fromV[v] = k
	return v
}
