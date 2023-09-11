//go:build durable

package coroutine

import (
	"slices"
	"strconv"
)

// Storage is a sparse collection of Serializable objects.
type Storage struct {
	// This is private so that the data structure is allowed to switch
	// the in-memory representation dynamically (e.g. a map[int]Serializable
	// may be more efficient for very sparse maps).
	objects []any
}

// NewStorage creates a Storage.
func NewStorage(objects []any) Storage {
	return Storage{objects: objects}
}

// Has is true if an object is defined for a specific index.
func (v *Storage) Has(i int) bool {
	return i >= 0 && i < len(v.objects) && v.objects[i] != nil
}

// Get gets the object for a specific index.
func (v *Storage) Get(i int) any {
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
func (v *Storage) Set(i int, value any) {
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
