package serde

// serde.go contains the reflection based serialization and deserialization
// procedures. It does not do any type memoization, as eventually codegen should
// be able to generate code for types. Almost nothing is optimized, as we are
// iterating on how it works to get it right first.

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"
)

// sID is the unique sID of a pointer or type in the serialized format.
type sID int64

// Serialize x at the end of b, returning it.
//
// To serialize interfaces, the global type register needs to be fed with
// possible types they can contain. If using coroc, it automatically generates
// init functions to register types likely to be used in the program. If not,
// use [RegisterType] to manually add a type to the register. Because
// [Serialize] starts with an interface, at least the type of the provided value
// needs to be registered.
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

	SerializeAny(s, t, p)
	return s.b
}

// Deserialize value from b. Return left over bytes.
func Deserialize(b []byte) (interface{}, []byte) {
	d := newDeserializer(b)
	var x interface{}
	px := &x
	t := reflect.TypeOf(px).Elem()
	p := unsafe.Pointer(px)
	deserializeInterface(d, t, p)
	return x, d.b
}

type Deserializer struct {
	// TODO: make it a slice since pointer ids is the sequence of integers
	// starting at 1.
	ptrs map[sID]unsafe.Pointer

	// input
	b []byte
}

func newDeserializer(b []byte) *Deserializer {
	return &Deserializer{
		ptrs: make(map[sID]unsafe.Pointer),
		b:    b,
	}
}

func (d *Deserializer) readPtr() (unsafe.Pointer, sID) {
	x, n := binary.Varint(d.b)
	d.b = d.b[n:]
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
	return &Serializer{
		ptrs:     make(map[unsafe.Pointer]sID),
		scanptrs: make(map[reflect.Value]struct{}),
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

func serializeBool(s *Serializer, v bool) {
	SerializeBool(s, v)
}

func deserializeBool(d *Deserializer) (v bool) {
	DeserializeBool(d, &v)
	return
}
