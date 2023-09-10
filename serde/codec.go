package serde

import (
	"fmt"
	"reflect"
)

// ID is the unique ID of a pointer or type in the serialized format.
type ID int64

type SerializeFn func(s *Serializer, x any, b []byte) []byte
type DeserializeFn func(s *Deserializer, b []byte) (any, []byte)

type typeCodec struct {
	id           ID
	rtype        reflect.Type
	serializer   SerializeFn
	deserializer DeserializeFn
}

type typeMap struct {
	byID   map[ID]typeCodec
	byType map[reflect.Type]ID
}

func newTypeMap() *typeMap {
	return &typeMap{
		byID:   make(map[ID]typeCodec),
		byType: make(map[reflect.Type]ID),
	}
}

func (t *typeMap) Add(x reflect.Type) ID {
	return t.AddWithCodec(x, nil, nil)
}

func (t *typeMap) AddWithCodec(x reflect.Type, ser SerializeFn, des DeserializeFn) ID {
	if id, ok := t.byType[x]; ok {
		return id
	}
	i := ID(len(t.byID))
	t.byID[i] = typeCodec{
		id:           i,
		rtype:        x,
		serializer:   ser,
		deserializer: des,
	}
	t.byType[x] = i
	return i
}

func (t *typeMap) IDof(x reflect.Type) ID {
	id, ok := t.byType[x]
	if !ok {
		panic(fmt.Errorf("type '%s' is not registered", x))
	}
	return id
}

func (t *typeMap) TypeOf(x ID) reflect.Type {
	codec, ok := t.byID[x]
	if !ok {
		panic(fmt.Errorf("type id '%d' not registered", x))
	}
	return codec.rtype
}

func (t *typeMap) CodecOf(x reflect.Type) (typeCodec, bool) {
	if t == nil {
		return typeCodec{}, false
	}
	id, ok := t.byType[x]
	if !ok {
		return typeCodec{}, false
	}
	return t.byID[id], true
}

var tm *typeMap = newTypeMap()

func RegisterType(x reflect.Type) {
	tm.Add(x)
}

func RegisterTypeWithCodec(x reflect.Type, ser SerializeFn, des DeserializeFn) {
	tm.AddWithCodec(x, ser, des)
}
