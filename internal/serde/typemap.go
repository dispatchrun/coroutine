package serde

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Global type register.
var Types *TypeMap = NewTypeMap()

// RegisterSerde assigns custom functions to serialize and deserialize a
// specific type.
func RegisterSerde[T any](
	serializer func(*Serializer, *T) error,
	deserializer func(*Deserializer, *T) error) {
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

type SerializerFn func(*Serializer, unsafe.Pointer)
type DeserializerFn func(d *Deserializer, p unsafe.Pointer)

type serde struct {
	id  int
	ser SerializerFn
	des DeserializerFn
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

func (m *TypeMap) Attach(t reflect.Type, ser SerializerFn, des DeserializerFn) {
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

type set[T comparable] map[T]struct{}

func (s set[T]) has(x T) bool {
	_, ok := s[x]
	return ok
}

func (s set[T]) add(x T) {
	s[x] = struct{}{}
}
