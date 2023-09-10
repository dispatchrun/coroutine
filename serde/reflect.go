package serde

import (
	"fmt"
	"log/slog"
	"reflect"
	"unsafe"
)

// reflect.go contains the reflection based serialization and deserialization
// procedures. It is used to handle values the codegen pass cannot generate
// methods for (interfaces). Eventually codegen should be able to generate code
// for types.

type iface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func ifacePointer(v interface{}) unsafe.Pointer {
	return (*iface)(unsafe.Pointer(&v)).ptr
}

func serializeAny(s *Serializer, r reflect.Value, b []byte) []byte {
	t := r.Type()
	codec, exists := tm.CodecOf(t)
	if exists && codec.serializer != nil {
		slog.Debug("using codec", "codec", codec, "type", t)
		return codec.serializer(s, r.Interface(), b)
	} else {
		slog.Debug("type has not codec", "type", t)
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't serialize reflect.Invalid"))
	case reflect.Bool:
		return reflectWrapSer(SerializeBool, s, r, b)
	case reflect.Int:
		return reflectWrapSer(SerializeInt, s, r, b)
	case reflect.Int64:
		return reflectWrapSer(SerializeInt64, s, r, b)
	case reflect.Int32:
		return reflectWrapSer(SerializeInt32, s, r, b)
	case reflect.Int16:
		return reflectWrapSer(SerializeInt16, s, r, b)
	case reflect.Int8:
		return reflectWrapSer(SerializeInt8, s, r, b)
	case reflect.Uint:
		return reflectWrapSer(SerializeUint, s, r, b)
	case reflect.Uint64:
		return reflectWrapSer(SerializeUint64, s, r, b)
	case reflect.Uint32:
		return reflectWrapSer(SerializeUint32, s, r, b)
	case reflect.Uint16:
		return reflectWrapSer(SerializeUint16, s, r, b)
	case reflect.Uint8:
		return reflectWrapSer(SerializeUint8, s, r, b)
	case reflect.Float64:
		return reflectWrapSer(SerializeFloat64, s, r, b)
	case reflect.Float32:
		return reflectWrapSer(SerializeFloat32, s, r, b)
	case reflect.Complex64:
		return reflectWrapSer(SerializeComplex64, s, r, b)
	case reflect.Complex128:
		return reflectWrapSer(SerializeComplex128, s, r, b)
	case reflect.String:
		return reflectWrapSer(SerializeString, s, r, b)
	// Array
	// Chan
	// Func
	case reflect.Interface:
		return reflectWrapSer(SerializeInterface, s, r, b)
	// Map
	case reflect.Pointer:
		return serializePointer(s, r, b)
	// Slice
	// String
	// Struct
	// UnsafePointer
	default:
		panic(fmt.Errorf("reflection cannot serialize type %s", t))
	}
}

func deserializeAny(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	codec, exists := tm.CodecOf(t)
	if exists && codec.deserializer != nil {
		v, b := codec.deserializer(d, b)
		rv := reflect.ValueOf(v)
		target := reflect.NewAt(t, p)
		target.Elem().Set(rv)
		return b
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic(fmt.Errorf("can't deserialize reflect.Invalid"))
	case reflect.Bool:
		return reflectWrapDes(DeserializeBool, d, p, b)
	case reflect.Int:
		return reflectWrapDes(DeserializeInt, d, p, b)
	case reflect.Int64:
		return reflectWrapDes(DeserializeInt64, d, p, b)
	case reflect.Int32:
		return reflectWrapDes(DeserializeInt32, d, p, b)
	case reflect.Int16:
		return reflectWrapDes(DeserializeInt16, d, p, b)
	case reflect.Int8:
		return reflectWrapDes(DeserializeInt8, d, p, b)
	case reflect.Uint:
		return reflectWrapDes(DeserializeUint, d, p, b)
	case reflect.Uint64:
		return reflectWrapDes(DeserializeUint64, d, p, b)
	case reflect.Uint32:
		return reflectWrapDes(DeserializeUint32, d, p, b)
	case reflect.Uint16:
		return reflectWrapDes(DeserializeUint16, d, p, b)
	case reflect.Uint8:
		return reflectWrapDes(DeserializeUint8, d, p, b)
	case reflect.Float64:
		return reflectWrapDes(DeserializeFloat64, d, p, b)
	case reflect.Float32:
		return reflectWrapDes(DeserializeFloat32, d, p, b)
	case reflect.Complex64:
		return reflectWrapDes(DeserializeComplex64, d, p, b)
	case reflect.Complex128:
		return reflectWrapDes(DeserializeComplex128, d, p, b)
	case reflect.String:
		return reflectWrapDes(DeserializeString, d, p, b)
	case reflect.Interface:
		return reflectWrapDes(DeserializeInterface, d, p, b)
	case reflect.Pointer:
		return deserializePointer(d, t, p, b)
	default:
		panic(fmt.Errorf("reflection cannot deserialize type %s", t))
	}
}

func reflectWrapSer[T any](f func(*Serializer, T, []byte) []byte, s *Serializer, r reflect.Value, b []byte) []byte {
	return f(s, r.Interface().(T), b)
}

func reflectWrapDes[T any](f func(*Deserializer, []byte) (T, []byte), d *Deserializer, p unsafe.Pointer, b []byte) []byte {
	x, b := f(d, b)
	*(*T)(p) = x
	return b
}

func dumptr(r reflect.Value) {
	fmt.Printf("=>%s, %s\n", r, r.Type())
}

func serializePointer(s *Serializer, r reflect.Value, b []byte) []byte {
	x := r.UnsafePointer()
	ok, b := s.WritePtr(x, b)
	if !ok {
		b = serializeAny(s, r.Elem(), b)
	}
	return b
}

func deserializePointer(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	r := reflect.NewAt(t, p)
	x, i, b := d.ReadPtr(b)
	if x != nil || i == 0 {
		r.Elem().Set(reflect.NewAt(t.Elem(), x))
	} else {
		newthing := reflect.New(t.Elem())
		r.Elem().Set(newthing)
		d.Store(i, newthing.UnsafePointer())
		b = deserializeAny(d, t.Elem(), newthing.UnsafePointer(), b)
	}
	return b
}
