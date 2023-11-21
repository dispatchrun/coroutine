package types

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Global serde register.
var serdes *serdemap = newSerdeMap()

// SerializerFunc is the signature of custom serializer functions. Use the
// [Serialize] function to drive the [Serializer]. Returning an error results in
// the program panicking.
type SerializerFunc[T any] func(*Serializer, *T) error

// DeserializerFunc is the signature of customer deserializer functions. Use the
// [Deserialize] function to drive the [Deserializer]. Returning an error
// results in the program panicking.
type DeserializerFunc[T any] func(*Deserializer, *T) error

// Register attaches custom serialization and deserialization functions to
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
// [Register] to control how they are serialized, and possibly perform
// additional initialization on deserialization. Those functions are drivers for
// [Serializer] and [Deserializer], that need to invoke [Serialize] and
// [DeserializeTo] in order to actually perform serialization and
// deserialization operations. Pointers to the same address are detected as such
// to be reconstructed as pointing to the same value. Slices are serialized by
// first serializing their backing array, and then length and capacity. As a
// result, slices sharing the same backing array are deserialized into one array
// with two shared slices, just like the original state was. Elements between
// length and capacity are also preserved.
func Register[T any](
	serializer SerializerFunc[T],
	deserializer DeserializerFunc[T]) {
	registerSerde[T](serdes, serializer, deserializer)
}

func registerSerde[T any](serdes *serdemap,
	serializer func(*Serializer, *T) error,
	deserializer func(*Deserializer, *T) error) {

	t := reflect.TypeOf((*T)(nil)).Elem()

	s := func(s *Serializer, actualType reflect.Type, p unsafe.Pointer) {
		if t != actualType {
			v := reflect.NewAt(actualType, p).Elem()
			box := reflect.New(t)
			box.Elem().Set(v.Convert(t))
			p = box.UnsafePointer()
		}
		if err := serializer(s, (*T)(p)); err != nil {
			panic(fmt.Errorf("serializing %s: %w", t, err))
		}
	}

	d := func(d *Deserializer, actualType reflect.Type, p unsafe.Pointer) {
		if t != actualType {
			box := reflect.New(t)
			boxp := box.UnsafePointer()
			if err := deserializer(d, (*T)(boxp)); err != nil {
				panic(fmt.Errorf("deserializing %s: %w", t, err))
			}
			v := reflect.NewAt(actualType, p)
			reinterpreted := reflect.ValueOf(box.Elem().Interface())
			v.Elem().Set(reinterpreted)
		} else {
			if err := deserializer(d, (*T)(p)); err != nil {
				panic(fmt.Errorf("deserializing %s: %w", t, err))
			}
		}
	}

	serdes.attach(t, s, d)
}

type serializerFunc func(*Serializer, reflect.Type, unsafe.Pointer)
type deserializerFunc func(*Deserializer, reflect.Type, unsafe.Pointer)

type serdeid = uint32

type serde struct {
	id  serdeid
	typ reflect.Type
	ser serializerFunc
	des deserializerFunc
}

type serdemap struct {
	serdes     []serde
	serdesByT  map[reflect.Type]serde
	interfaces []serde
}

func newSerdeMap() *serdemap {
	return &serdemap{
		serdesByT: make(map[reflect.Type]serde),
	}
}

func (m *serdemap) attach(t reflect.Type, ser serializerFunc, des deserializerFunc) {
	if ser == nil || des == nil {
		panic("both serializer and deserializer need to be provided")
	}

	s := m.serdesByT[t]
	s.id = serdeid(len(m.serdes))
	s.typ = t
	s.ser = ser
	s.des = des
	m.serdes = append(m.serdes, s)
	m.serdesByT[t] = s

	if t.Kind() == reflect.Interface {
		m.interfaces = append(m.interfaces, s)
	}
}

func (m *serdemap) serdeByType(x reflect.Type) (serde, bool) {
	s, ok := m.serdesByT[x]
	if ok {
		return s, true
	}
	for i := range m.interfaces {
		s := m.interfaces[i]
		if x.Implements(s.typ) {
			return s, true
		}
	}
	return serde{}, false
}

func (m *serdemap) serdeByID(id serdeid) serde {
	if int(id) >= len(m.serdes) {
		panic(fmt.Sprintf("serde %d not found", id))
	}
	return m.serdes[id]
}
