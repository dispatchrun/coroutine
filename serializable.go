package coroutine

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

// Serializable objects can be serialized to bytes.
type Serializable interface {
	// MarshalAppend marshals the object and appends the resulting bytes to
	// the provided buffer.
	MarshalAppend(b []byte) ([]byte, error)
}

// Deserializable objects can be deserialized from bytes.
type Deserializable interface {
	// Unmarshal unmarshals an object from a buffer. It returns the number of
	// bytes that were read from the buffer in order to reconstruct the object.
	Unmarshal(b []byte) (n int, err error)
}

// MarshalAppend appends a Serializable object to a buffer, along with
// information about the type of Serializable object that's being
// serialized. The bytes can later be passed to Unmarshal to
// reconstruct the same Serializable object.
func MarshalAppend(b []byte, s Serializable) ([]byte, error) {
	t, ok := serializableByReflectType[reflect.TypeOf(s)]
	if !ok {
		return nil, fmt.Errorf("serializable type %T has not been registered", s)
	}
	b = binary.AppendVarint(b, int64(t.id))
	return s.MarshalAppend(b)
}

// Unmarshal unmarshals a Serializable object from a buffer. It returns
// the object, and the number of bytes that were read from the buffer in
// order to reconstruct the object.
func Unmarshal(b []byte) (Serializable, int, error) {
	id, n := binary.Varint(b)
	if n <= 0 || int64(int(id)) != id {
		return nil, 0, fmt.Errorf("invalid serializable type info")
	}
	t, ok := serializableByID[int(id)]
	if !ok {
		return nil, 0, fmt.Errorf("serializable type %d not registered", id)
	}
	value, vn, err := t.constructor(b[n:])
	return value, n + vn, err
}

// RegisterSerializable registers a Serializable type for use with
// the top-level MarshalAppend and Unmarshal functions.
//
// The specified Serializable must implement Deserializable, either directly
// or indirectly (i.e. either s or *s implements Deserializable).
//
// A constructor for the Serializable is generated using reflection and passed
// to RegisterSerializableConstructor. It may be more efficient to manually
// generate a constructor and call RegisterSerializableConstructor instead.
func RegisterSerializable(s Serializable) {
	reflectType := reflect.TypeOf(s)

	switch {
	case reflectType.Implements(deserializableType):
		RegisterSerializableConstructor(s, func(b []byte) (Serializable, int, error) {
			v := reflect.Zero(reflectType).Interface()
			n, err := v.(Deserializable).Unmarshal(b)
			return v.(Serializable), n, err
		})
	case reflect.PtrTo(reflectType).Implements(deserializableType):
		RegisterSerializableConstructor(s, func(b []byte) (Serializable, int, error) {
			p := reflect.New(reflectType)
			n, err := p.Interface().(Deserializable).Unmarshal(b)
			return p.Elem().Interface().(Serializable), n, err
		})
	default:
		panic(fmt.Sprintf("type %T is not Deserializable", s))
	}
}

// RegisterSerializableConstructor registers a Serializable type for use with
// the top-level MarshalAppend and Unmarshal functions.
//
// The caller is expected to provide a constructor that unmarshals bytes into
// an instance of the specified Serializable.
//
// If the Serializable implements Deserializable, a constructor can instead
// automatically be generated using reflection. See RegisterSerializable.
func RegisterSerializableConstructor(s Serializable, constructor UnmarshalSerializable) {
	reflectType := reflect.TypeOf(s)
	if _, ok := serializableByReflectType[reflectType]; ok {
		panic(fmt.Sprintf("serializable type %T already registered", s))
	}

	t := &serializableType{
		id:          serializableNextID,
		constructor: constructor,
	}
	serializableNextID++

	serializableByReflectType[reflectType] = t
	serializableByID[t.id] = t
}

// UnmarshalSerializable unmarshals a Serializable object from a buffer.
// It returns the object, and the number of bytes that were read from the
// buffer in order to reconstruct the object.
type UnmarshalSerializable func([]byte) (Serializable, int, error)

var serializableByReflectType = map[reflect.Type]*serializableType{}
var serializableByID = map[int]*serializableType{}
var serializableNextID int

type serializableType struct {
	id          int
	constructor UnmarshalSerializable
}

var deserializableType = reflect.TypeOf((*Deserializable)(nil)).Elem()
