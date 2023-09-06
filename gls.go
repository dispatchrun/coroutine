package coroutine

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
var (
	gmutex sync.RWMutex
	gstate map[uintptr]any
)

func loadContext(g uintptr) any {
	gmutex.RLock()
	v := gstate[g]
	gmutex.RUnlock()
	return v
}

func storeContext(g uintptr, c any) {
	gmutex.Lock()
	if gstate == nil {
		gstate = make(map[uintptr]any)
	}
	gstate[g] = c
	gmutex.Unlock()
}

func clearContext(g uintptr) {
	gmutex.Lock()
	delete(gstate, g)
	gmutex.Unlock()
}
