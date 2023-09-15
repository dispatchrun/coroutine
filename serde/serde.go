package serde

import (
	"reflect"
	"unsafe"

	"github.com/stealthrocket/coroutine/internal/serde"
)

// Serialize is a blah.
type Serializer = serde.Serializer

// TODO
type Deserializer = serde.Deserializer

func Serialize[T any](s *Serializer, x T) {
	var p unsafe.Pointer
	r := reflect.ValueOf(x)
	t := r.Type()
	if r.CanAddr() {
		p = r.Addr().UnsafePointer()
	} else {
		n := reflect.New(t)
		n.Elem().Set(r)
		p = n.UnsafePointer()
	}
	serde.SerializeAny(s, t, p)
}

func DeserializeTo[T any](d *Deserializer, x *T) {
	r := reflect.ValueOf(x)
	t := r.Type().Elem()
	p := r.UnsafePointer()
	serde.DeserializeAny(d, t, p)
}

// func SerializeString(s *Serializer, x *string, b []byte) []byte {
// 	return serde.SerializeString(s, x, b)
// }

// func DeserializeString(d *Deserializer, x *string, b []byte) []byte {
// 	return serde.DeserializeString(d, x, b)
// }

// func SerializeBool(s *Serializer, x bool, b []byte) []byte {
// 	return serde.SerializeBool(s, x, b)
// }

// func DeserializeBool(d *Deserializer, x *bool, b []byte) []byte {
// 	return serde.DeserializeBool(d, x, b)
// }

// func SerializeInt(s *Serializer, x int, b []byte) []byte {
// 	return serde.SerializeInt(s, x, b)
// }

// func DeserializeInt(d *Deserializer, x *int, b []byte) []byte {
// 	return serde.DeserializeInt(d, x, b)
// }

// func SerializeInt64(s *Serializer, x int64, b []byte) []byte {
// 	return serde.SerializeInt64(s, x, b)
// }

// func DeserializeInt64(d *Deserializer, x *int64, b []byte) []byte {
// 	return serde.DeserializeInt64(d, x, b)
// }

// func SerializeInt32(s *Serializer, x int32, b []byte) []byte {
// 	return serde.SerializeInt32(s, x, b)
// }

// func DeserializeInt32(d *Deserializer, x *int32, b []byte) []byte {
// 	return serde.DeserializeInt32(d, x, b)
// }

// func SerializeInt16(s *Serializer, x int16, b []byte) []byte {
// 	return serde.SerializeInt16(s, x, b)
// }

// func DeserializeInt16(d *Deserializer, x *int16, b []byte) []byte {
// 	return serde.DeserializeInt16(d, x, b)
// }

// func SerializeInt8(s *Serializer, x int8, b []byte) []byte {
// 	return serde.SerializeInt8(s, x, b)
// }

// func DeserializeInt8(d *Deserializer, x *int8, b []byte) []byte {
// 	return serde.DeserializeInt8(d, x, b)
// }

// func SerializeUint(s *Serializer, x uint, b []byte) []byte {
// 	return serde.SerializeUint(s, x, b)
// }

// func DeserializeUint(d *Deserializer, x *uint, b []byte) []byte {
// 	return serde.DeserializeUint(d, x, b)
// }

// func SerializeUint64(s *Serializer, x uint64, b []byte) []byte {
// 	return serde.SerializeUint64(s, x, b)
// }

// func DeserializeUint64(d *Deserializer, x *uint64, b []byte) []byte {
// 	return serde.DeserializeUint64(d, x, b)
// }

// func SerializeUint32(s *Serializer, x uint32, b []byte) []byte {
// 	return serde.SerializeUint32(s, x, b)
// }

// func DeserializeUint32(d *Deserializer, x *uint32, b []byte) []byte {
// 	return serde.DeserializeUint32(d, x, b)
// }

// func SerializeUint16(s *Serializer, x uint16, b []byte) []byte {
// 	return serde.SerializeUint16(s, x, b)
// }

// func DeserializeUint16(d *Deserializer, x *uint16, b []byte) []byte {
// 	return serde.DeserializeUint16(d, x, b)
// }

// func SerializeUint8(s *Serializer, x uint8, b []byte) []byte {
// 	return serde.SerializeUint8(s, x, b)
// }

// func DeserializeUint8(d *Deserializer, x *uint8, b []byte) []byte {
// 	return serde.DeserializeUint8(d, x, b)
// }

// func SerializeFloat32(s *Serializer, x float32, b []byte) []byte {
// 	return serde.SerializeFloat32(s, x, b)
// }

// func DeserializeFloat32(d *Deserializer, x *float32, b []byte) []byte {
// 	return serde.DeserializeFloat32(d, x, b)
// }

// func SerializeFloat64(s *Serializer, x float64, b []byte) []byte {
// 	return serde.SerializeFloat64(s, x, b)
// }

// func DeserializeFloat64(d *Deserializer, x *float64, b []byte) []byte {
// 	return serde.DeserializeFloat64(d, x, b)
// }

// func SerializeComplex64(s *Serializer, x complex64, b []byte) []byte {
// 	return serde.SerializeComplex64(s, x, b)
// }

// func DeserializeComplex64(d *Deserializer, x *complex64, b []byte) []byte {
// 	return serde.DeserializeComplex64(d, x, b)
// }

// func SerializeComplex128(s *Serializer, x complex128, b []byte) []byte {
// 	return serde.SerializeComplex128(s, x, b)
// }

// func DeserializeComplex128(d *Deserializer, x *complex128, b []byte) []byte {
// 	return serde.DeserializeComplex128(d, x, b)
// }
