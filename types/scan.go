package types

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"unsafe"
)

type container struct {
	addr unsafe.Pointer
	typ  reflect.Type
	len  int // >=0 for arrays, -1 for other types
}

// Returns true iff at least one byte of the address space is shared between c
// and x (opposite of [disjoints]).
func (c container) overlaps(x container) bool {
	return !c.disjoints(x)
}

// Returns true iff there is not a single byte of the address space is shared
// between c and x (opposite of [overlaps]).
func (c container) disjoints(x container) bool {
	return (uintptr(c.addr)+c.size()) <= uintptr(x.addr) ||
		(uintptr(x.addr)+x.size()) <= uintptr(c.addr)
}

// Returns true iff x is fully included in c.
func (c container) contains(x container) bool {
	return uintptr(x.addr) >= uintptr(c.addr) && uintptr(x.addr)+x.size() <= uintptr(c.addr)+c.size()
}

func (c container) after(x container) bool {
	return uintptr(c.addr) > uintptr(x.addr)
}

// Size in bytes of c.
func (c container) size() uintptr {
	if c.len >= 0 {
		return uintptr(c.len) * c.typ.Size()
	}
	return c.typ.Size()
}

func (c container) isStruct() bool {
	return !c.isArray() && c.typ.Kind() == reflect.Struct
}

func (c container) isArray() bool {
	return c.len >= 0
}

func (c container) valid() bool {
	return c.typ != nil
}

func (c container) has(p unsafe.Pointer) bool {
	return uintptr(p) >= uintptr(c.addr) && uintptr(p) < (uintptr(c.addr)+c.size())
}

func (c container) offset(p unsafe.Pointer) uintptr {
	return uintptr(p) - uintptr(c.addr)
}

func (c container) compare(p unsafe.Pointer) int {
	if c.has(p) {
		return 0
	}
	if uintptr(p) < uintptr(c.addr) {
		return -1
	}
	return 1
}

func (c container) String() string {
	return fmt.Sprintf("[%d-%d] %d %s(%d)", c.addr, uintptr(c.addr)+c.size(), c.size(), c.typ, c.len)
}

type containers []container

func (c *containers) dump() {
	s := *c
	log.Printf("====================== CONTAINERS ======================")
	log.Printf("Count: %d", len(s))
	for i, x := range s {
		log.Printf("#%d: %s", i, x)
	}
	log.Printf("========================================================")
}

func (c *containers) of(p unsafe.Pointer) container {
	s := *c
	i, found := sort.Find(len(s), func(i int) int {
		return s[i].compare(p)
	})
	if !found {
		return container{}
	}
	return s[i]
}

func (c *containers) add(t reflect.Type, length int, p unsafe.Pointer) {
	if length == 0 {
		return
	}
	if t.Size() == 0 {
		return
	}

	if p == nil {
		panic("tried to add nil pointer")
	}

	defer func() {
		r := recover()
		if r != nil {
			c.dump()
			panic(r)
		}
	}()

	x := container{addr: p, typ: t, len: length}
	i := c.insert(x)
	c.fixup(i)
	if i > 0 {
		c.fixup(i - 1)
	}
}

func (c *containers) fixup(i int) {
	s := *c

	if i == len(s)-1 {
		return
	}

	x := s[i]
	next := s[i+1]

	if !x.overlaps(next) {
		// Not at least an overlap, nothing to do.
		return
	}

	if x.contains(next) {
		if x.isStruct() {
			// Struct fully contains next element. Remove the next
			// element and nothing else to do.
			c.remove(i + 1)
			return
		}
		c.remove(i + 1)
		// Array fully contains next container. Nothing to do
		return
	}

	// There is some overlap. The only thing we accept to merge are arrays
	// of the same type.
	if !x.isArray() || !next.isArray() || x.typ != next.typ {
		panic(fmt.Errorf("only support merging arrays of same type (%s, %s)", x.typ, next.typ))
	}

	c.merge(i)

	// Do it again in case the merge connected new areas.
	c.fixup(i)
}

func (c *containers) merge(i int) {
	s := *c
	a := s[i]
	b := s[i+1]

	elemSize := a.typ.Size()

	// sanity check alignment
	if (uintptr(b.addr)-uintptr(a.addr))%uintptr(elemSize) != 0 {
		panic("overlapping arrays aren't aligned")
	}

	s[i].len = int((uintptr(b.addr)-uintptr(a.addr))/elemSize) + b.len

	c.remove(i + 1)
}

func (c *containers) remove(i int) {
	before := len(*c)
	s := *c
	copy(s[i:], s[i+1:])
	*c = s[:len(s)-1]
	after := len(*c)
	if after >= before {
		panic("did not remove anything")
	}
}

func (c *containers) insert(x container) int {
	*c = append(*c, container{})
	s := *c
	// Find where to insert the new container. By start address first, then
	// by decreasing size (so that the bigger container comes before).
	i := sort.Search(len(s)-1, func(i int) bool {
		if s[i].after(x) {
			return true
		}
		if s[i].addr == x.addr {
			return x.size() > s[i].size()
		}
		return false
	})
	copy(s[i+1:], s[i:])
	s[i] = x

	// Debug assertion.
	for i, x := range s {
		if i == 0 {
			continue
		}
		if uintptr(x.addr) < uintptr(s[i-1].addr) {
			panic("bad address order after insert")
		}
		if uintptr(x.addr) == uintptr(s[i-1].addr) {
			if x.size() > s[i-1].size() {
				panic("invalid size order after insert")
			}
		}
	}

	return i
}

// scan the value of type t at address p recursively to build up the serializer
// state with necessary information for encoding. At the moment it only creates
// the memory regions table.
//
// It uses s.scanptrs to track which pointers it has already visited to avoid
// infinite loops. It does not clean it up after. I'm sure there is something
// more useful we could do with that.
func (s *Serializer) scan(t reflect.Type, p unsafe.Pointer) {
	s.scan1(t, p, map[reflect.Value]struct{}{})
}

func (s *Serializer) scan1(t reflect.Type, p unsafe.Pointer, seen map[reflect.Value]struct{}) {
	if p == nil {
		return
	}

	// Don't scan types where custom serialization routines
	// have been registered.
	if _, ok := s.serdes.serdeByType(t); ok {
		return
	}

	r := reflect.NewAt(t, p)
	if _, ok := seen[r]; ok {
		return
	}
	seen[r] = struct{}{}

	if r.IsNil() {
		return
	}

	switch t.Kind() {
	case reflect.Invalid:
		panic("handling invalid reflect.Type")
	case reflect.Array:
		s.containers.add(t.Elem(), t.Len(), p)
		et := t.Elem()
		es := int(et.Size())
		for i := 0; i < t.Len(); i++ {
			ep := unsafe.Add(p, es*i)
			s.scan1(et, ep, seen)
		}
	case reflect.Slice:
		sr := r.Elem()
		if sr.IsNil() {
			return
		}
		ep := sr.UnsafePointer()
		if ep == nil {
			return
		}
		// Estimate size of backing array.
		et := t.Elem()
		es := int(et.Size())

		s.containers.add(et, sr.Cap(), ep)
		for i := 0; i < sr.Len(); i++ {
			ep := unsafe.Add(ep, es*i)
			s.scan1(et, ep, seen)
		}
	case reflect.Interface:
		if r.Elem().IsNil() {
			return
		}
		it := r.Elem().Interface()
		if it == nil {
			return
		}
		et := reflect.TypeOf(it)

		eptr := ifacePtr(p, et)
		s.scan1(et, eptr, seen)
	case reflect.Struct:
		s.containers.add(t, -1, p)
		n := t.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			ft := f.Type
			fp := unsafe.Add(p, f.Offset)
			s.scan1(ft, fp, seen)
		}
	case reflect.Pointer:
		if r.Elem().IsNil() {
			return
		}
		ep := r.Elem().UnsafePointer()
		s.scan1(t.Elem(), ep, seen)
	case reflect.String:
		str := *(*string)(p)
		sp := unsafe.StringData(str)
		if sp == nil {
			// empty strings are represented as nil pointers.
			return
		}
		s.containers.add(byteT, len(str), unsafe.Pointer(sp))
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
			ki := k.Interface()
			kp := ifacePtr(unsafe.Pointer(&ki), kt)
			s.scan1(kt, kp, seen)

			v := iter.Value()
			vi := v.Interface()
			vp := ifacePtr(unsafe.Pointer(&vi), vt)
			s.scan1(vt, vp, seen)
		}
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
		// nothing to do
	default:
		// TODO:
		// Chan
		// Func
		// UnsafePointer
	}
}
