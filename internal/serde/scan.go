package serde

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

// Returns true iff c starts before x.
func (c container) before(x container) bool {
	return uintptr(c.addr) <= uintptr(x.addr)
}

func (c container) after(x container) bool {
	return uintptr(c.addr) > uintptr(x.addr)
}

// Size in bytes of c.
func (c container) size() uintptr {
	return c.typ.Size()
}

func (c container) isStruct() bool {
	return c.typ.Kind() == reflect.Struct
}

func (c container) isArray() bool {
	return c.typ.Kind() == reflect.Array
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
	return fmt.Sprintf("[%d-%d[ %d %s", c.addr, uintptr(c.addr)+c.size(), c.size(), c.typ)
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

func (c *containers) add(t reflect.Type, p unsafe.Pointer) {
	if t.Size() == 0 {
		return
	}

	if p == nil {
		panic("tried to add nil pointer")
	}
	switch t.Kind() {
	case reflect.Struct, reflect.Array:
	default:
		panic(fmt.Errorf("tried to add non struct or array container: %s (%s)", t, t.Kind()))
	}

	defer func() {
		r := recover()
		if r != nil {
			c.dump()
			panic(r)
		}
	}()

	x := container{addr: p, typ: t}
	i := c.insert(x)
	c.fixup(i)
	if i > 0 {
		c.fixup(i - 1)
	}

	c.dump()
}

func (c *containers) fixup(i int) {
	s := *c

	log.Println("fixup:", i)

	if i == len(s)-1 {
		return
	}

	x := s[i]
	next := s[i+1]

	if !x.overlaps(next) {
		// Not at least an overlap, nothing to do.
		log.Println("=> no overlap")
		return
	}

	if x.contains(next) {
		log.Println("=>contains")
		if x.isStruct() {
			// Struct fully contains next element. Remove the next
			// element and nothing else to do.
			c.remove(i + 1)
			return
		}
		// Array fully contains next container. Nothing to do
		return
	}

	// There is some overlap. The only thing we accept to merge are arrays
	// of the same type.
	if !x.isArray() || !next.isArray() || x.typ.Elem() != next.typ.Elem() {
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

	elemSize := a.typ.Elem().Size()

	// sanity check alignment
	if (uintptr(b.addr)-uintptr(a.addr))%uintptr(elemSize) != 0 {
		panic("overlapping arrays aren't aligned")
	}

	// new element count of the array
	newlen := int((uintptr(b.addr)-uintptr(a.addr))/elemSize) + b.typ.Len()
	s[i].typ = reflect.ArrayOf(newlen, a.typ.Elem())

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
	log.Print("inserting ", x)
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
	fmt.Println("i=", i)
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
	case reflect.Array:
		s.containers.add(t, p)
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
		s.containers.add(xt, ep)
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
		s.containers.add(t, p)
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
		s.containers.add(xt, unsafe.Pointer(sp))
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
