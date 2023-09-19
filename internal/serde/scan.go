package serde

import (
	"fmt"
	"reflect"
	"sort"
	"unsafe"
)

type regions []region

func (r *regions) dump() {
	fmt.Println("========== MEMORY REGIONS ==========")
	fmt.Println("Found", len(*r), "regions.")
	for i, r := range *r {
		fmt.Printf("#%d: [%d-%d] %d %s\n", i, r.start, r.end, r.size(), r.typ)
	}
	fmt.Println("====================================")
}

// debug function to ensure the state hold its invariants. panic if they don't.
func (r *regions) validate() {
	s := *r
	if len(s) == 0 {
		return
	}

	for i := 0; i < len(s); i++ {
		if uintptr(s[i].start) > uintptr(s[i].end) {
			panic(fmt.Errorf("region #%d has invalid bounds: start=%d end=%d delta=%d", i, s[i].start, s[i].end, s[i].size()))
		}
		if s[i].typ == nil {
			panic(fmt.Errorf("region #%d has nil type", i))
		}
		if i == 0 {
			continue
		}
		if uintptr(s[i].start) <= uintptr(s[i-1].end) {
			r.dump()
			panic(fmt.Errorf("region #%d and #%d overlap", i-1, i))
		}
	}
}

// size computes the amount of bytes coverred by all known regions.
func (r *regions) size() int {
	n := 0
	for _, r := range *r {
		n += r.size()
	}
	return n
}

func (r *regions) regionOf(p unsafe.Pointer) region {
	//	fmt.Printf("Searching regions for %d\n", p)
	addr := uintptr(p)
	s := *r
	if len(s) == 0 {
		//		fmt.Printf("\t=> No regions\n")
		return region{}
	}

	i := sort.Search(len(s), func(i int) bool {
		return uintptr(s[i].start) >= addr
	})
	//	fmt.Printf("\t=> i = %d\n", i)

	if i < len(s) && uintptr(s[i].start) == addr {
		return s[i]
	}

	if i > 0 {
		i--
	}
	if uintptr(s[i].start) > addr || uintptr(s[i].end) < addr {
		return region{}
	}
	return s[i]

}

func (r *regions) add(t reflect.Type, start unsafe.Pointer) {
	size := t.Size()
	if size == 0 {
		return
	}

	end := unsafe.Add(start, size-1)

	//	fmt.Printf("Adding [%d-%d[ %d %s\n", startAddr, endAddr, endAddr-startAddr, t)
	startSize := r.size()
	defer func() {
		//r.Dump()
		r.validate()
		endSize := r.size()
		if endSize < startSize {
			panic(fmt.Errorf("regions shrunk (%d -> %d)", startSize, endSize))
		}
	}()

	s := *r

	if len(s) == 0 {
		*r = append(s, region{
			start: start,
			end:   end,
			typ:   t,
		})
		return
	}

	// Invariants:
	// (1) len(s) > 0
	// (2) s is sorted by start address
	// (3) s contains no overlapping range

	i := sort.Search(len(s), func(i int) bool {
		return uintptr(s[i].start) >= uintptr(start)
	})
	//fmt.Println("\ti =", i)

	if i < len(s) && uintptr(s[i].start) == uintptr(start) {
		// Pointer is present in the set. If it's contained in the
		// region that already exists, we are done.
		if uintptr(s[i].end) >= uintptr(end) {
			return
		}

		// Otherwise extend the region.
		s[i].end = end
		s[i].typ = t

		// To maintain invariant (3), keep extending the selected region
		// until it becomes the last one or the next range is disjoint.
		r.extend(i)
		return
	}
	// Pointer did not point to the beginning of a region.

	// Attempt to grow the previous region.
	if i > 0 {
		if uintptr(start) <= uintptr(s[i-1].end) {
			if uintptr(end) >= uintptr(s[i-1].end) {
				s[i-1].end = end
				r.extend(i - 1)
			}
			return
		}
	}

	// Attempt to grow the next region.
	if i+1 < len(s) {
		if uintptr(end) >= uintptr(s[i+1].start) {
			s[i+1].start = start
			if uintptr(end) >= uintptr(s[i+1].end) {
				s[i+1].end = end
			}
			s[i+1].typ = t
			r.extend(i + 1)
			return
		}
	}

	// Just insert it.
	s = append(s, region{})
	copy(s[i+1:], s[i:])
	s[i] = region{start: start, end: end, typ: t}
	*r = s
	r.extend(i)
}

// extend attempts to grow region i by swallowing any region after it, as long
// as it would make one continous region. It is used after a modification of
// region i to maintain the invariants.
func (r *regions) extend(i int) {
	s := *r
	grown := 0
	for next := i + 1; next < len(s) && uintptr(s[i].end) >= uintptr(s[next].start); next++ {
		s[i].end = s[next].end
		grown++
	}
	copy(s[i+1:], s[i+1+grown:])
	*r = s[:len(s)-grown]
}

type region struct {
	start unsafe.Pointer // inclusive
	end   unsafe.Pointer // exclusive
	typ   reflect.Type
}

func (r region) valid() bool {
	return r.typ != nil
}

func (r region) size() int {
	return int(uintptr(r.end) - uintptr(r.start))
}

func (r region) offset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(r.start))
}

// scan the value of type t at address p recursively to build up the serializer
// state with necessary information for encoding. At the moment it only creates
// the memory regions table.
//
// It uses s.scanptrs to track which pointers it has already visited to avoid
// infinite loops. It does not clean it up after. I'm sure there is something
// more useful we could do with that.
func scan(s *Serializer, t reflect.Type, p unsafe.Pointer) {
	if p == nil {
		return
	}

	r := reflect.NewAt(t, p)
	if _, ok := s.scanptrs[r]; ok {
		return
	}
	s.scanptrs[r] = struct{}{}

	switch t.Kind() {
	case reflect.Invalid:
		panic("handling invalid reflect.Type")
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		s.regions.add(t, p)
	case reflect.Array:
		s.regions.add(t, p)
		et := t.Elem()
		es := int(et.Size())
		for i := 0; i < t.Len(); i++ {
			ep := unsafe.Add(p, es*i)
			scan(s, et, ep)
		}
	case reflect.Slice:
		sr := r.Elem()
		ep := sr.UnsafePointer()
		if ep == nil {
			return
		}
		// Estimate size of backing array.
		et := t.Elem()
		es := int(et.Size())

		// Create a new type for the backing array.
		xt := reflect.ArrayOf(sr.Cap(), t.Elem())
		s.regions.add(xt, ep)
		for i := 0; i < sr.Len(); i++ {
			ep := unsafe.Add(ep, es*i)
			scan(s, et, ep)
		}
	case reflect.Interface:
		x := *(*interface{})(p)
		et := reflect.TypeOf(x)
		eptr := (*iface)(p).ptr
		if eptr == nil {
			return
		}
		if inlined(et) {
			xp := (*iface)(p).ptr
			eptr = unsafe.Pointer(&xp)
		}

		scan(s, et, eptr)
	case reflect.Struct:
		s.regions.add(t, p)
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			ft := f.Type
			fp := unsafe.Add(p, f.Offset)
			scan(s, ft, fp)
		}
	case reflect.Pointer:
		ep := r.Elem().UnsafePointer()
		scan(s, t.Elem(), ep)
	case reflect.String:
		str := *(*string)(p)
		sp := unsafe.StringData(str)
		xt := reflect.ArrayOf(len(str), byteT)
		s.regions.add(xt, unsafe.Pointer(sp))
	case reflect.Map:
		m := r.Elem()
		if m.IsNil() || m.Len() == 0 {
			return
		}
		kt := t.Key()
		vt := t.Elem()
		iter := m.MapRange()
		for iter.Next() {
			k := iter.Key()
			kp := (*iface)(unsafe.Pointer(&k)).ptr
			scan(s, kt, kp)

			v := iter.Value()
			vp := (*iface)(unsafe.Pointer(&v)).ptr
			scan(s, vt, vp)
		}

	default:
		// TODO:
		// Chan
		// Func
		// UnsafePointer
	}
}
