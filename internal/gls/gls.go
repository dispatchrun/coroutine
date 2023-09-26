package gls

import "sync"

// goroutine local storage; the map contains one entry for each goroutine that
// is started to power a coroutine.
//
// TOOD: the global mutex is likely going to become a contention point in highly
// parallel programs, here's how we should fix:
//
//   - create a sharded map with 64 buckets, each bucket contains a map
//   - use a sync.Mutex in each bucket for synchronization; cheaper than RWMutex
//   - mask the value of g to determine in which bucket its GLS is stored
//
// An alternative to using a global map could be to analyze the memory layout of
// the runtime.g type and determine if there is spare room after the struct to
// store the Context pointer: the Go memory allocate uses size classes to park
// objects in buckets, there is often spare space after large values like the
// runtime.g type since they will be assigned to the size class greater or equal
// to their type size. We only need 4 or 8 bytes of spare space on 32 or 64 bit
// architectures. This approach would remove all potential contention accessing
// and synchronizing on global state, and would also turn the map lookups into
// simple memory loads.
var (
	gmutex sync.RWMutex
	gstate map[G]any
)

// G is a reference to a goroutine, and provides a way
// to load, store and clear a goroutine local context.
type G uintptr

// Context retrieves the goroutine local storage for contexts.
func Context() G {
	return G(getg())
}

// Load loads the goroutine local context.
func (g G) Load() any {
	gmutex.RLock()
	v := gstate[g]
	gmutex.RUnlock()
	return v
}

// Store stores the goroutine local context.
func (g G) Store(c any) {
	gmutex.Lock()
	if gstate == nil {
		gstate = make(map[G]any)
	}
	gstate[g] = c
	gmutex.Unlock()
}

// Clear clears the goroutine local context.
func (g G) Clear() {
	gmutex.Lock()
	delete(gstate, g)
	gmutex.Unlock()
}
