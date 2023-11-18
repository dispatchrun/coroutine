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
	"runtime"
	"unsafe"

	coroutinev1 "github.com/stealthrocket/coroutine/gen/proto/go/coroutine/v1"
)

// sID is the unique sID of a pointer or type in the serialized format.
type sID int64

// ErrBuildIDMismatch is an error that occurs when a program attempts
// to deserialize objects from another build.
var ErrBuildIDMismatch = errors.New("build ID mismatch")

// Information about the current build. This is attached to serialized
// items, and checked at deserialization time to ensure compatibility.
var buildInfo *coroutinev1.Build

func init() {
	buildInfo = &coroutinev1.Build{
		Id:   buildID,
		Os:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// Serialize x.
//
// The output of Serialize can be reconstructed back to a Go value using
// [Deserialize].
func Serialize(x any) ([]byte, error) {
	s := newSerializer()
	w := &x // w is *interface{}
	wr := reflect.ValueOf(w)
	p := wr.UnsafePointer() // *interface{}
	t := wr.Elem().Type()   // what x contains

	// Scan pointers to collect memory regions.
	s.scan(t, p)

	serializeAny(s, t, p)

	state := &coroutinev1.State{
		State:     s.b,
		Build:     buildInfo,
		Types:     s.types.types,
		Functions: s.funcs.funcs,
	}
	return state.MarshalVT()
}

// Deserialize value from b. Return left over bytes.
func Deserialize(b []byte) (interface{}, error) {
	var state coroutinev1.State
	if err := state.UnmarshalVT(b); err != nil {
		return nil, err
	}
	if state.Build.Id != buildInfo.Id {
		return nil, fmt.Errorf("%w: got %v, expect %v", ErrBuildIDMismatch, state.Build.Id, buildInfo.Id)
	}

	d := newDeserializer(state.State, state.Types, state.Functions)
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
	serdes *serdemap
	types  *typemap
	funcs  *funcmap

	// TODO: make it a slice since pointer ids is the sequence of integers
	// starting at 1.
	ptrs map[sID]unsafe.Pointer

	// input
	b []byte
}

func newDeserializer(b []byte, ctypes []*coroutinev1.Type, cfuncs []*coroutinev1.Function) *Deserializer {
	types := newTypeMap(serdes, ctypes)
	return &Deserializer{
		serdes: serdes,
		types:  types,
		funcs:  newFuncMap(types, cfuncs),
		ptrs:   make(map[sID]unsafe.Pointer),
		b:      b,
	}
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
	serdes     *serdemap
	types      *typemap
	funcs      *funcmap
	ptrs       map[unsafe.Pointer]sID
	containers containers

	// Output
	b []byte
}

func newSerializer() *Serializer {
	types := newTypeMap(serdes, nil)

	return &Serializer{
		serdes: serdes,
		types:  types,
		funcs:  newFuncMap(types, nil),
		ptrs:   make(map[unsafe.Pointer]sID),
		b:      make([]byte, 0, 128),
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
