package serde

import (
	"fmt"
	"reflect"
)

// ID is the unique ID of a pointer or type in the serialized format.
type ID int64

type typeMap struct {
	byID   map[ID]reflect.Type
	byType map[reflect.Type]ID
}

func newTypeMap() *typeMap {
	return &typeMap{
		byID:   make(map[ID]reflect.Type),
		byType: make(map[reflect.Type]ID),
	}
}

func (m *typeMap) add(t reflect.Type) {
	if _, ok := m.byType[t]; ok {
		return
	}
	id := ID(len(m.byID)) + 1
	m.byType[t] = id
	m.byID[id] = t
}

func (m *typeMap) exists(t reflect.Type) bool {
	_, ok := m.byType[t]
	return ok
}

func (m *typeMap) Add(t reflect.Type) {
	if m.exists(t) {
		return
	}
	m.add(t)
	m.add(reflect.PointerTo(t))

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

func (m *typeMap) IDof(x reflect.Type) ID {
	id, ok := m.byType[x]
	if !ok {
		panic(fmt.Errorf("type '%s' is not registered", x))
	}
	return id
}

func (m *typeMap) TypeOf(x ID) reflect.Type {
	t, ok := m.byID[x]
	if !ok {
		panic(fmt.Errorf("type id '%d' not registered", x))
	}
	return t
}

var tm *typeMap = newTypeMap()

func RegisterType[T any]() {
	tm.Add(reflect.TypeOf((*T)(nil)).Elem())
}
