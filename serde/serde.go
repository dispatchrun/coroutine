package serde

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"reflect"
	"unsafe"
)

// Deserializer contains the state of the deserializer.
type Deserializer struct {
	// TODO: make it a slice
	ptrs map[ID]unsafe.Pointer
}

func (d *Deserializer) ReadPtr(b []byte) (unsafe.Pointer, ID, []byte) {
	x, n := binary.Varint(b)
	i := ID(x)
	p := d.ptrs[i]

	slog.Debug("Deserializer ReadPtr", "i", i, "p", p, "n", n)
	return p, i, b[n:]
}

func (d *Deserializer) Store(i ID, p unsafe.Pointer) {
	if d.ptrs[i] != nil {
		panic(fmt.Errorf("trying to overwirte known ID %d with %p", i, p))
	}
	d.ptrs[i] = p
}

func newDeserializer(d *Deserializer) *Deserializer {
	if d == nil {
		d = &Deserializer{
			ptrs: make(map[ID]unsafe.Pointer),
		}
	}
	return d
}

// Serializer contains the state of the serializer.
type Serializer struct {
	ptrs map[unsafe.Pointer]ID
}

func (s *Serializer) WritePtr(p unsafe.Pointer, b []byte) (bool, []byte) {
	off := len(b)
	if p == nil {
		slog.Debug("Serializer WritePtr wrote <nil> pointer", "offset", off)
		return true, binary.AppendVarint(b, 0)
	}
	i, ok := s.ptrs[p]
	if !ok {
		i = ID(len(s.ptrs) + 1)
		s.ptrs[p] = i
	}
	slog.Debug("Serializer WritePtr", "i", i, "p", p, "offset", off)
	return ok, binary.AppendVarint(b, int64(i))
}

func EnsureSerializer(s *Serializer) *Serializer {
	if s == nil {
		s = &Serializer{
			ptrs: make(map[unsafe.Pointer]ID),
		}
	}
	return s
}

// Serializable objects can be serialized to bytes.
type Serializable interface {
	// MarshalAppend marshals the object and appends the resulting bytes to
	// the provided buffer.
	MarshalAppend(b []byte) ([]byte, error)

	// Unmarshal unmarshals an object from a buffer. It returns the number
	// of bytes that were read from the buffer in order to reconstruct the
	// object.
	Unmarshal(b []byte) (n int, err error)
}

// Write size of slice or map because we want to distinguish an initialized
// empty value from a a nil value.
func serializeSize(v reflect.Value, b []byte) []byte {
	size := -1
	if !v.IsNil() {
		size = v.Len()
	}
	return binary.AppendVarint(b, int64(size))
}

// returns -1 if the original value was nil
func deserializeSize(b []byte) (int, []byte) {
	l, n := binary.Varint(b)
	return int(l), b[n:]
}
