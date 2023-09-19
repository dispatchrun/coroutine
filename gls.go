package coroutine

import (
	"sync"
	"unsafe"
)

// goroutine local storage; the map contains one entry for each goroutine that
// is started to power a coroutine.
//
// An alternative to using a global map could be to analyze the memory layout of
// the runtime.g type and determine if there is spare room after the struct to
// store the Context pointer: the Go memory allocate uses size classes to park
// objects in buckets, there is often spare space after large values like the
// runtime.g type since they will be assigned to the size class greater or equal
// to their type size. We only need 8 or 16 bytes of spare space on 32 or 64 bit
// architectures to store the context type and value. This approach would remove
// all potential contention accessing and synchronizing on global state, and
// would also turn the map lookups into simple memory loads.
var gstate glsTable

const glsTableBuckets = 64

type glsTable [glsTableBuckets]glsBucket

func (t *glsTable) bucket(k unsafe.Pointer) *glsBucket {
	h := uintptr(k)
	// murmur3 hash finalizer; hashing pointers is necessary to ensure a good
	// distribution of keys across buckets, otherwise the alignment and
	// collocation done by the memory allocator group all keys in a few
	// buckets.
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	// bucket selection
	h &= glsTableBuckets - 1
	return &t[h]
}

func (t *glsTable) load(k unsafe.Pointer) any {
	return t.bucket(k).load(k)
}

func (t *glsTable) store(k unsafe.Pointer, v any) {
	t.bucket(k).store(k, v)
}

func (t *glsTable) clear(k unsafe.Pointer) {
	t.bucket(k).clear(k)
}

type glsBucket struct {
	values sync.Map
}

func (b *glsBucket) load(k unsafe.Pointer) any {
	v, _ := b.values.Load(k)
	return v
}

func (b *glsBucket) store(k unsafe.Pointer, v any) {
	b.values.Store(k, v)
}

func (b *glsBucket) clear(k unsafe.Pointer) {
	b.values.Delete(k)
}

func loadContext(g unsafe.Pointer) any {
	return gstate.load(g)
}

func storeContext(g unsafe.Pointer, c any) {
	gstate.store(g, c)
}

func clearContext(g unsafe.Pointer) {
	gstate.clear(g)
}
