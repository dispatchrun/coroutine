package serde

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Global type register.
var Types *TypeMap = NewTypeMap()

// RegisterType into the global register to make it known to the serialization
// system. It is only required to register named types.
//
// coroc usually generates calls to this function. It should be called in an
// init function so that types are always registered in the same order.
//
// Named types are recursively added.
func RegisterType[T any]() {
	registerType[T](Types)
}

// Scan T and add all named types to the type map.
func registerType[T any](tm *TypeMap) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	tm.Add(t)
	addNamedTypes(tm, make(set[reflect.Type]), t)
}

func addNamedTypes(tm *TypeMap, seen set[reflect.Type], t reflect.Type) {
	if seen.has(t) {
		return
	}
	seen.add(t)
	if named(t) {
		tm.Add(t)
	}
	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			addNamedTypes(tm, seen, f.Type)
		}
	case reflect.Func:
		for i := 0; i < t.NumIn(); i++ {
			addNamedTypes(tm, seen, t.In(i))
		}
		for i := 0; i < t.NumOut(); i++ {
			addNamedTypes(tm, seen, t.Out(i))
		}
	case reflect.Map:
		addNamedTypes(tm, seen, t.Key())
		fallthrough
	case reflect.Slice, reflect.Array, reflect.Pointer:
		addNamedTypes(tm, seen, t.Elem())
	}
}

// RegisterTypeWithSerde is the same as [RegisterType] but assigns serialization
// and deserialization for this type.
func RegisterTypeWithSerde[T any](
	serializer func(*Serializer, *T) error,
	deserializer func(*Deserializer, *T) error) {
	registerTypeWithSerde[T](Types, serializer, deserializer)
}

func registerTypeWithSerde[T any](tm *TypeMap,
	serializer func(*Serializer, *T) error,
	deserializer func(*Deserializer, *T) error) {

	registerType[T](tm)
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

	_, ok := m.cache.GetK(t)
	if !ok {
		panic(fmt.Errorf("register type %s before attaching serde", t))
	}

	m.serdes[t] = serde{ser: ser, des: des}
}

func (m *TypeMap) Add(t reflect.Type) {
	if _, ok := m.cache.GetK(t); ok {
		return
	}

	x := &typeinfo{kind: typeCustom, val: len(m.custom)}
	m.custom = append(m.custom, t)
	m.cache.Add(t, x)
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

func (m *doublemap[K, V]) Add(k K, v V) {
	if m.fromK == nil {
		m.fromK = make(map[K]V)
		m.fromV = make(map[V]K)
	}
	m.fromK[k] = v
	m.fromV[v] = k
}

type set[T comparable] map[T]struct{}

func (s set[T]) has(x T) bool {
	_, ok := s[x]
	return ok
}

func (s set[T]) add(x T) {
	s[x] = struct{}{}
}
