package coroutine

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strconv"
)

// Storage is a sparse collection of Serializable objects.
type Storage struct {
	// This is private so that the data structure is allowed to switch
	// the in-memory representation dynamically (e.g. a map[int]Serializable
	// may be more efficient for very sparse maps).
	objects []Serializable
}

// NewStorage creates a Storage.
func NewStorage(objects []Serializable) Storage {
	return Storage{objects: objects}
}

// Has is true if an object is defined for a specific index.
func (v *Storage) Has(i int) bool {
	return i >= 0 && i < len(v.objects) && v.objects[i] != nil
}

// Get gets the object for a specific index.
func (v *Storage) Get(i int) Serializable {
	if !v.Has(i) {
		panic("missing object " + strconv.Itoa(i))
	}
	return v.objects[i]
}

// Delete gets the object for a specific index.
func (v *Storage) Delete(i int) {
	if !v.Has(i) {
		panic("missing object " + strconv.Itoa(i))
	}
	v.objects[i] = nil
}

// Set sets the object for a specific index.
func (v *Storage) Set(i int, value Serializable) {
	if n := i + 1; n > len(v.objects) {
		v.objects = slices.Grow(v.objects, n-len(v.objects))
		v.objects = v.objects[:n]
	}
	v.objects[i] = value
}

func (v *Storage) shrink() {
	i := len(v.objects) - 1
	for i >= 0 && v.objects[i] == nil {
		i--
	}
	v.objects = v.objects[:i+1]
}

// MarshalAppend appends objects to the provided buffer.
func (v *Storage) MarshalAppend(b []byte) ([]byte, error) {
	v.shrink()

	// This is a sparse map. For each object we encode the object as well
	// as its index into the map. We also encode the number of objects as
	// well as the length of the data structure. The latter isn't strictly
	// required, but is used as a hint for the deserializer so that it can
	// preallocate the necessary space.
	var objectCount int
	for _, s := range v.objects {
		if s == nil {
			continue
		}
		objectCount++
	}

	b = binary.AppendVarint(b, int64(len(v.objects)))
	b = binary.AppendVarint(b, int64(objectCount))

	for i, s := range v.objects {
		if s == nil {
			continue
		}
		b = binary.AppendVarint(b, int64(i))

		var err error
		b, err = MarshalAppend(b, s)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

// Unmarshal deserializes a Frame from the provided buffer, returning
// the number of bytes that were read in order to reconstruct the
// frame.
func (v *Storage) Unmarshal(b []byte) (int, error) {
	size, n := binary.Varint(b)
	if n <= 0 || int64(int(size)) != size {
		return 0, fmt.Errorf("invalid storage size: %v", b)
	}

	count, vn := binary.Varint(b[n:])
	if vn <= 0 || int64(int(count)) != count {
		return 0, fmt.Errorf("invalid storage count: %v", b)
	}
	n += vn

	objects := make([]Serializable, size)
	for i := 0; i < int(count); i++ {
		id, vn := binary.Varint(b[n:])
		if vn <= 0 || int64(int(id)) != id {
			return 0, fmt.Errorf("invalid storage id: %v", b)
		}
		n += vn

		s, sn, err := Unmarshal(b[n:])
		if err != nil {
			return 0, fmt.Errorf("invalid storage object: %w", err)
		}
		n += sn

		objects[id] = s
	}

	v.objects = objects
	return n, nil
}
