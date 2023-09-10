package serde

import (
	"encoding/binary"
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
	case reflect.Array:
		return serializeArray(s, r, b)
	case reflect.Interface:
		return reflectWrapSer(SerializeInterface, s, r, b)
	case reflect.Map:
		return serializeMap(s, r, b)
	case reflect.Pointer:
		return serializePointer(s, r, b)
	case reflect.Slice:
		return serializeSlice(s, r, b)
	case reflect.Struct:
		return serializeStruct(s, r, b)
	// Chan
	// Func
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
	case reflect.Array:
		return deserializeArray(d, t, p, b)
	case reflect.Slice:
		return deserializeSlice(d, t, p, b)
	case reflect.Map:
		return deserializeMap(d, t, p, b)
	case reflect.Struct:
		return deserializeStruct(d, t, p, b)
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

func serializeMap(s *Serializer, r reflect.Value, b []byte) []byte {
	size := 0
	if r.IsNil() {
		size = -1
	} else {
		size = r.Len()
	}
	b = binary.AppendVarint(b, int64(size))

	iter := r.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		b = serializeAny(s, k, b)
		b = serializeAny(s, v, b)
	}
	return b
}

func deserializeMap(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n, b := DeserializeMapSize(b)
	if n < 0 { // nil map
		return b
	}
	nv := reflect.MakeMapWithSize(t, n)
	r := reflect.NewAt(t, p)
	r.Elem().Set(nv)
	for i := 0; i < n; i++ {
		k := reflect.New(t.Key())
		b = deserializeAny(d, t.Key(), k.UnsafePointer(), b)
		v := reflect.New(t.Elem())
		b = deserializeAny(d, t.Elem(), v.UnsafePointer(), b)
		r.Elem().SetMapIndex(k.Elem(), v.Elem())
	}
	return b
}

func serializeSlice(s *Serializer, r reflect.Value, b []byte) []byte {
	b = SerializeSize(r.Len(), b)
	for i := 0; i < r.Len(); i++ {
		x := r.Index(i)
		b = serializeAny(s, x, b)
	}
	return b
}

func deserializeSlice(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n, b := DeserializeSize(b)
	nv := reflect.MakeSlice(t, n, n)
	r := reflect.NewAt(t, p)
	r.Elem().Set(nv)
	size := int(t.Elem().Size())
	p = nv.UnsafePointer()
	for i := 0; i < n; i++ {
		e := unsafe.Add(p, size*i)
		b = deserializeAny(d, t.Elem(), e, b)
	}
	return b
}

func serializeArray(s *Serializer, r reflect.Value, b []byte) []byte {
	for i := 0; i < r.Len(); i++ {
		x := r.Index(i)
		b = serializeAny(s, x, b)
	}
	return b
}

func deserializeArray(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	size := int(t.Elem().Size())
	for i := 0; i < t.Len(); i++ {
		e := unsafe.Add(p, size*i)
		b = deserializeAny(d, t.Elem(), e, b)
	}
	return b
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

func reflectFieldSupported(ft reflect.StructField) bool {
	// TODO
	return !ft.Anonymous && ft.IsExported()
}

func serializeStruct(s *Serializer, r reflect.Value, b []byte) []byte {
	t := r.Type()
	n := t.NumField()
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		if !reflectFieldSupported(ft) {
			continue
		}
		f := r.Field(i)
		b = serializeAny(s, f, b)
	}
	return b
}

func deserializeStruct(d *Deserializer, t reflect.Type, p unsafe.Pointer, b []byte) []byte {
	n := t.NumField()
	r := reflect.NewAt(t, p)
	for i := 0; i < n; i++ {
		ft := t.Field(i)
		if !reflectFieldSupported(ft) {
			continue
		}

		fr := reflect.New(ft.Type)
		b = deserializeAny(d, ft.Type, fr.UnsafePointer(), b)

		r.Elem().Field(i).Set(fr.Elem())
	}
	return b
}
