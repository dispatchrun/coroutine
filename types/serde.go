package types

// serde.go contains the reflection based serialization and deserialization
// procedures. It does not do any type memoization, as eventually codegen should
// be able to generate code for types. Almost nothing is optimized, as we are
// iterating on how it works to get it right first.

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	coroutinev1 "github.com/dispatchrun/coroutine/gen/proto/go/coroutine/v1"
	"github.com/dispatchrun/coroutine/internal/reflectext"
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
		Id:   reflectext.BuildID(),
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

	i := &x // w is *interface{}
	vp := reflect.ValueOf(i)
	t := vp.Elem().Type() // what x contains

	// Scan pointers to collect memory regions.
	s.scan(t, vp.UnsafePointer())

	s.Serialize(vp.Elem())

	state := &coroutinev1.State{
		Build:     buildInfo,
		Types:     s.types.types,
		Functions: s.funcs.funcs,
		Strings:   s.strings.strings,
		Regions:   s.regions,
		Root: &coroutinev1.Region{
			Type: s.types.ToType(t) << 1,
			Data: s.buffer,
		},
	}
	return state.MarshalVT()
}

// Deserialize value from b. Return left over bytes.
func Deserialize(b []byte) (x interface{}, err error) {
	defer func() {
		// FIXME: the deserialize*() functions panic on invalid input
		if e := recover(); e != nil {
			err = fmt.Errorf("cannot deserialize state: %v", e)
		}
	}()

	var state coroutinev1.State
	if err := state.UnmarshalVT(b); err != nil {
		return nil, err
	}
	if state.Build.Id != buildInfo.Id {
		return nil, fmt.Errorf("%w: got %v, expect %v", ErrBuildIDMismatch, state.Build.Id, buildInfo.Id)
	}

	d := newDeserializer(state.Root.Data, state.Types, state.Functions, state.Regions, state.Strings)

	ip := &x // w is *interface{}
	vp := reflect.ValueOf(ip)
	t := vp.Elem().Type() // what x contains

	deserializeValue(d, t, vp)

	if len(d.buffer) != 0 {
		err = errors.New("trailing bytes")
	}
	return
}

type Deserializer struct {
	*deserializerContext

	// input
	buffer []byte
}

type deserializerContext struct {
	serdes  *serdemap
	types   *typemap
	funcs   *funcmap
	regions []*coroutinev1.Region
	ptrs    map[sID]unsafe.Pointer
}

func newDeserializer(b []byte, ctypes []*coroutinev1.Type, cfuncs []*coroutinev1.Function, regions []*coroutinev1.Region, cstrings []string) *Deserializer {
	strings := newStringMap(cstrings)
	types := newTypeMap(serdes, strings, ctypes)
	return &Deserializer{
		&deserializerContext{
			serdes:  serdes,
			types:   types,
			funcs:   newFuncMap(types, strings, cfuncs),
			regions: regions,
			ptrs:    make(map[sID]unsafe.Pointer),
		},
		b,
	}
}

func (d *Deserializer) fork(b []byte) *Deserializer {
	return &Deserializer{
		d.deserializerContext,
		b,
	}
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
	reflectext.DefaultVisitor

	*serializerContext

	// Output
	buffer []byte
}

type serializerContext struct {
	serdes     *serdemap
	types      *typemap
	funcs      *funcmap
	strings    *stringmap
	ptrs       map[unsafe.Pointer]sID
	regions    []*coroutinev1.Region
	containers containers
}

func newSerializer() *Serializer {
	strings := newStringMap(nil)
	types := newTypeMap(serdes, strings, nil)

	return &Serializer{
		serializerContext: &serializerContext{
			serdes:  serdes,
			types:   types,
			strings: strings,
			funcs:   newFuncMap(types, strings, nil),
			ptrs:    make(map[unsafe.Pointer]sID),
		},
		buffer: make([]byte, 0, 128),
	}
}

func (s *Serializer) fork() *Serializer {
	return &Serializer{
		serializerContext: s.serializerContext,
		buffer:            make([]byte, 0, 128),
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

// Serialize a value. See [RegisterSerde].
func SerializeT[T any](s *Serializer, x T) {
	v := reflect.ValueOf(x)
	s.appendReflectType(v.Type())
	s.Serialize(v)
}

// Deserialize a value to the provided non-nil pointer. See [RegisterSerde].
func DeserializeTo[T any](d *Deserializer, x *T) {
	v := reflect.ValueOf(x)
	t := v.Type().Elem()
	actualType, length := d.reflectType()
	if length < 0 {
		if t != actualType {
			panic(fmt.Sprintf("cannot deserialize %s as %s", actualType, t))
		}
	} else if t.Kind() != reflect.Array || t.Len() != length || t != actualType.Elem() {
		panic(fmt.Sprintf("cannot deserialize [%d]%s as %s", length, actualType, t))
	}
	deserializeValue(d, t, v)
}
