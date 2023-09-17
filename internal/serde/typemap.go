package serde

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Global type register.
var Types *TypeMap = NewTypeMap()

// RegisterType into the global register to make it known to the serialization
// system.
//
// coroc usually generates calls to this function. It should be called in an
// init function so that types are always registered in the same order.
//
// Types are recursively added, as well as *T.
func RegisterType[T any]() {
	Types.Add(reflect.TypeOf((*T)(nil)).Elem())
}

// RegisterTypeWithSerde is the same as [RegisterType] but assigns serialization
// and deserialization for this type.
func RegisterTypeWithSerde[T any](
	serializer func(*Serializer, *T) error,
	deserializer func(*Deserializer, *T) error) {

	RegisterType[T]()
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

	Types.Attach(t, s, d)
}

type SerializerFn func(*Serializer, unsafe.Pointer)
type DeserializerFn func(d *Deserializer, p unsafe.Pointer)

type serde struct {
	ser SerializerFn
	des DeserializerFn
}

type TypeMap struct {
	byID   map[sID]reflect.Type
	byType map[reflect.Type]sID
	serdes map[reflect.Type]serde
}

func NewTypeMap() *TypeMap {
	return &TypeMap{
		byID:   make(map[sID]reflect.Type),
		byType: make(map[reflect.Type]sID),
		serdes: make(map[reflect.Type]serde),
	}
}

func (m *TypeMap) Attach(t reflect.Type, ser SerializerFn, des DeserializerFn) {
	if ser == nil || des == nil {
		panic("both serializer and deserializer need to be provided")
	}

	_, ok := m.byType[t]
	if !ok {
		panic(fmt.Errorf("register type %s before attaching serde", t))
	}

	m.serdes[t] = serde{ser: ser, des: des}
}

func (m *TypeMap) Add(t reflect.Type) {
	if m.exists(t) {
		return
	}
	m.addExact(t)
	m.addExact(reflect.PointerTo(t))

	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		m.Add(t.Elem())
	case reflect.Map:
		m.Add(t.Key())
		m.Add(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			m.Add(t.Field(i).Type)
		}
	}
}

func (m *TypeMap) addExact(t reflect.Type) {
	if _, ok := m.byType[t]; ok {
		return
	}
	id := sID(len(m.byID)) + 1
	m.byType[t] = id
	m.byID[id] = t
}

func (m *TypeMap) exists(t reflect.Type) bool {
	_, ok := m.byType[t]
	return ok
}

func (m *TypeMap) idOf(x reflect.Type) sID {
	id, ok := m.byType[x]
	if !ok {
		panic(fmt.Errorf("type '%s' is not registered", x))
	}
	return id
}

func (m *TypeMap) typeOf(x sID) reflect.Type {
	t, ok := m.byID[x]
	if !ok {
		panic(fmt.Errorf("type id '%d' not registered", x))
	}
	return t
}

func (m *TypeMap) serdeOf(x reflect.Type) (serde, bool) {
	s, ok := m.serdes[x]
	return s, ok
}
