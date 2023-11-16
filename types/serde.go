package types

// serde.go contains the reflection based serialization and deserialization
// procedures. It does not do any type memoization, as eventually codegen should
// be able to generate code for types. Almost nothing is optimized, as we are
// iterating on how it works to get it right first.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

// sID is the unique sID of a pointer or type in the serialized format.
type sID int64

// ErrBuildIDMismatch is an error that occurs when a program attempts
// to deserialize objects from another build.
var ErrBuildIDMismatch = errors.New("build ID mismatch")

// Serialize x.
//
// The output of Serialize can be reconstructed back to a Go value using
// [Deserialize].
func Serialize(x any) []byte {
	s := newSerializer()
	w := &x // w is *interface{}
	wr := reflect.ValueOf(w)
	p := wr.UnsafePointer() // *interface{}
	t := wr.Elem().Type()   // what x contains

	scan(s, t, p)
	// scan dirties s.scanptrs, so clean it up.
	clear(s.scanptrs)

	serializeAny(s, t, p)
	return s.b
}

// Deserialize value from b. Return left over bytes.
func Deserialize(b []byte) (interface{}, error) {
	d, err := newDeserializer(b)
	if err != nil {
		return nil, err
	}
	var x interface{}
	px := &x
	t := reflect.TypeOf(px).Elem()
	p := unsafe.Pointer(px)
	deserializeInterface(d, t, p)
	if len(d.b) != 0 {
		return nil, errors.New("trailing bytes")
	}
	return x, nil
}

type Deserializer struct {
	// TODO: make it a slice since pointer ids is the sequence of integers
	// starting at 1.
	ptrs map[sID]unsafe.Pointer

	// input
	b []byte
}

func newDeserializer(b []byte) (*Deserializer, error) {
	buildIDLength, n := binary.Varint(b)
	if n <= 0 || buildIDLength <= 0 || buildIDLength > int64(len(buildID)) || int64(len(b)-n) < buildIDLength {
		return nil, fmt.Errorf("missing or invalid build ID")
	}
	b = b[n:]
	serializedBuildID := string(b[:buildIDLength])
	b = b[buildIDLength:]
	if serializedBuildID != buildID {
		return nil, fmt.Errorf("%w: got %v, expect %v", ErrBuildIDMismatch, serializedBuildID, buildID)
	}

	return &Deserializer{
		ptrs: make(map[sID]unsafe.Pointer),
		b:    b,
	}, nil
}

func (d *Deserializer) readPtr() (unsafe.Pointer, sID) {
	x, n := binary.Varint(d.b)
	d.b = d.b[n:]

	// pointer into static uint64 table
	if x == -1 {
		x, n = binary.Varint(d.b)
		d.b = d.b[n:]
		p := staticPointer(int(x))
		return p, 0
	}

	i := sID(x)
	p := d.ptrs[i]
	return p, i
}

func (d *Deserializer) store(i sID, p unsafe.Pointer) {
	if d.ptrs[i] != nil {
		panic(fmt.Errorf("trying to overwrite known ID %d with %p", i, p))
	}
	d.ptrs[i] = p
}

// Serializer holds the state for serialization.
//
// The ptrs value maps from pointers to IDs. Each time the serialization process
// encounters a pointer, it assigns it a new unique ID for its given address.
// This mechanism allows writing shared data only once. The actual value is
// written the first time a given pointer ID is encountered.
//
// The containers value has ranges of memory held by container types. They are
// the values that actually own memory: structs and arrays.
//
// Serialization starts with scanning the graph of values to find all the
// containers and add the range of memory they occupy into the map. Regions
// belong to the outermost container. For example:
//
//	struct X {
//	  struct Y {
//	    int
//	  }
//	}
//
// creates only one container: the struct X. Both struct Y and the int are
// containers, but they are included in the region of struct X.
//
// Those two mechanisms allow the deserialization of pointers that point to
// shared memory. Only outermost containers are serialized. All pointers either
// point to a container, or an offset into that container.
type Serializer struct {
	ptrs       map[unsafe.Pointer]sID
	containers containers

	// TODO: move out. just used temporarily by scan
	scanptrs map[reflect.Value]struct{}

	// Output
	b []byte
}

func newSerializer() *Serializer {
	b := make([]byte, 0, 128)
	b = binary.AppendVarint(b, int64(len(buildID)))
	b = append(b, buildID...)

	return &Serializer{
		ptrs:     make(map[unsafe.Pointer]sID),
		scanptrs: make(map[reflect.Value]struct{}),
		b:        b,
	}
}

// Returns true if it created a new ID (false if reused one).
func (s *Serializer) assignPointerID(p unsafe.Pointer) (sID, bool) {
	id, ok := s.ptrs[p]
	if !ok {
		id = sID(len(s.ptrs) + 1)
		s.ptrs[p] = id
	}
	return id, !ok
}

func serializeVarint(s *Serializer, size int) {
	s.b = binary.AppendVarint(s.b, int64(size))
}

func deserializeVarint(d *Deserializer) int {
	l, n := binary.Varint(d.b)
	d.b = d.b[n:]
	return int(l)
}

// Serialize a value. See [RegisterSerde].
func SerializeT[T any](s *Serializer, x T) {
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
	serializeAny(s, t, p)
}

// Deserialize a value to the provided non-nil pointer. See [RegisterSerde].
func DeserializeTo[T any](d *Deserializer, x *T) {
	r := reflect.ValueOf(x)
	t := r.Type().Elem()
	p := r.UnsafePointer()
	deserializeAny(d, t, p)
}
