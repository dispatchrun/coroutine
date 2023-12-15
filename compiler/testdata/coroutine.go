//go:build !durable

package testdata

import (
	"time"
	"unsafe"

	"github.com/stealthrocket/coroutine"
)

//go:generate coroc

func SomeFunctionThatShouldExistInTheCompiledFile() {
}

func Identity(n int) {
	coroutine.Yield[int, any](n)
}

func SquareGenerator(n int) {
	for i := 1; i <= n; i++ {
		coroutine.Yield[int, any](i * i)
	}
}

func SquareGeneratorTwice(n int) {
	SquareGenerator(n)
	SquareGenerator(n)
}

func SquareGeneratorTwiceLoop(n int) {
	for i := 0; i < 2; i++ {
		SquareGenerator(n)
	}
}

func EvenSquareGenerator(n int) {
	for i := 1; i <= n; i++ {
		if mod2 := i % 2; mod2 == 0 {
			coroutine.Yield[int, any](i * i)
		}
	}
}

func NestedLoops(n int) int {
	var count int
	for i := 1; i <= n; i++ {
		for j := 1; j <= n; j++ {
			for k := 1; k <= n; k++ {
				coroutine.Yield[int, any](i * j * k)
				count++
			}
		}
	}
	return count
}

func FizzBuzzIfGenerator(n int) {
	for i := 1; i <= n; i++ {
		if i%3 == 0 && i%5 == 0 {
			coroutine.Yield[int, any](FizzBuzz)
		} else if i%3 == 0 {
			coroutine.Yield[int, any](Fizz)
		} else if mod5 := i % 5; mod5 == 0 {
			coroutine.Yield[int, any](Buzz)
		} else {
			coroutine.Yield[int, any](i)
		}
	}
}

func FizzBuzzSwitchGenerator(n int) {
	for i := 1; i <= n; i++ {
		switch {
		case i%3 == 0 && i%5 == 0:
			coroutine.Yield[int, any](FizzBuzz)
		case i%3 == 0:
			coroutine.Yield[int, any](Fizz)
		case i%5 == 0:
			coroutine.Yield[int, any](Buzz)
		default:
			coroutine.Yield[int, any](i)
		}
	}
}

func Shadowing(_ int) {
	i := 0
	coroutine.Yield[int, any](i) // 0

	if i := 1; true {
		coroutine.Yield[int, any](i) // 1
	}
	coroutine.Yield[int, any](i) // 0

	for i := 1; i < 3; i++ {
		coroutine.Yield[int, any](i) // 1, 2
	}
	coroutine.Yield[int, any](i) // 0

	switch i := 1; i {
	case 1:
		switch i := 2; i {
		default:
			coroutine.Yield[int, any](i) // 2
		}
		coroutine.Yield[int, any](i) // 1
	}

	coroutine.Yield[int, any](i) // 0
	{
		i := 1
		{
			i := 2
			coroutine.Yield[int, any](i) // 2
		}
		coroutine.Yield[int, any](i) // 1
	}

	coroutine.Yield[int, any](i) // 0
	var j = i
	{
		j := 1
		coroutine.Yield[int, any](j) // 1
	}
	coroutine.Yield[int, any](j) // 0

	const k = 11
	{
		const k = 12
		{
			k := 13
			coroutine.Yield[int, any](k) // 13
		}
		coroutine.Yield[int, any](k) // 12
	}
	coroutine.Yield[int, any](k) // 11

	type foo uint16
	{
		type foo uint32
		coroutine.Yield[int, any](int(unsafe.Sizeof(foo(0)))) // 4
	}
	coroutine.Yield[int, any](int(unsafe.Sizeof(foo(0)))) // 2

	const siz = 1
	type baz [siz]uint8
	{
		type bar [siz]uint8
		coroutine.Yield[int, any](int(unsafe.Sizeof(bar{}))) // 1
		const siz = unsafe.Sizeof(bar{}) * 2
		type baz [siz]uint8
		coroutine.Yield[int, any](int(unsafe.Sizeof(baz{}))) // 2
	}
	coroutine.Yield[int, any](int(unsafe.Sizeof(baz{}))) // 1
}

func RangeSliceIndexGenerator(_ int) {
	for i := range []int{10, 20, 30} {
		coroutine.Yield[int, any](i)
	}
}

func RangeArrayIndexValueGenerator(_ int) {
	for i, v := range [...]int{10, 20, 30} {
		coroutine.Yield[int, any](i)
		coroutine.Yield[int, any](v)
	}
}

func TypeSwitchingGenerator(_ int) {
	for _, val := range []any{int8(10), int16(20), int32(30), int64(40)} {
		switch val.(type) {
		case int8:
			coroutine.Yield[int, any](1)
		case int16:
			coroutine.Yield[int, any](2)
		case int32:
			coroutine.Yield[int, any](4)
		case int64:
			coroutine.Yield[int, any](8)
		}
		switch v := val.(type) {
		case int8:
			coroutine.Yield[int, any](int(v))
		case int16:
			coroutine.Yield[int, any](int(v))
		case int32:
			coroutine.Yield[int, any](int(v))
		case int64:
			coroutine.Yield[int, any](int(v))
		}
	}
}

func LoopBreakAndContinue(_ int) {
	for i := 0; i < 10; i++ {
		if mod2 := i % 2; mod2 == 0 {
			continue
		}
		if i > 5 {
			break
		}
		coroutine.Yield[int, any](i)
	}

outer:
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			coroutine.Yield[int, any](j)
			switch j {
			case 0:
				continue
			case 1:
				switch i {
				case 0:
					continue outer
				case 1:
					break outer
				}
			}
		}
	}
}

func RangeOverMaps(n int) {
	m := map[int]int{}
	for range m {
		panic("unreachable")
	}
	for _ = range m {
		panic("unreachable")
	}
	for _, _ = range m {
		panic("unreachable")
	}
	m[n] = n * 10
	for range m {
		coroutine.Yield[int, any](0)
	}
	for k := range m {
		coroutine.Yield[int, any](k)
	}
	for k, v := range m {
		coroutine.Yield[int, any](k)
		coroutine.Yield[int, any](v)
	}

	// Map iteration order is not deterministic, so to
	// test iteration with a map with more than one element
	// we'll build a map and then successively delete keys
	// while yielding the length of the map.
	m2 := make(map[int]struct{}, n)
	for i := 0; i < n; i++ {
		m2[i] = struct{}{}
	}
	coroutine.Yield[int, any](len(m2))
	for k := range m2 {
		delete(m2, k)
		coroutine.Yield[int, any](len(m2))
	}
}

func Range(n int, do func(int)) {
	for i := 0; i < n; i++ {
		do(i)
	}
}

func Double(n int) {
	coroutine.Yield[int, any](2 * n)
}

func RangeTriple(n int) {
	Range(n, func(i int) {
		coroutine.Yield[int, any](3 * i)
	})
}

func RangeTripleFuncValue(n int) {
	f := func(i int) {
		coroutine.Yield[int, any](3 * i)
	}
	Range(n, f)
}

func RangeReverseClosureCaptureByValue(n int) {
	i := 0
	f := func() {
		coroutine.Yield[int, any](n - (i + 1))
	}

	for i < n {
		f()
		i++
	}
}

func Range10ClosureCapturingValues() {
	i := 0
	n := 10
	f := func() bool {
		if i < n {
			coroutine.Yield[int, any](i)
			i++
			return true
		}
		return false
	}

	for f() {
	}
}

func Range10ClosureCapturingPointers() {
	i, n := 0, 10
	p := &i
	q := &n
	f := func() bool {
		if *p < *q {
			coroutine.Yield[int, any](*p)
			(*p)++
			return true
		}
		return false
	}

	for f() {
	}
}

func Range10ClosureHeterogenousCapture() {
	var (
		a int8    = 0
		b int16   = 1
		c int32   = 2
		d int64   = 3
		e uint8   = 4
		f uint16  = 5
		g uint32  = 6
		h uint64  = 7
		i uintptr = 8
		j         = func() int { return int(i) + 1 }
	)

	n := 0
	x := func() bool {
		var v int
		switch n {
		case 0:
			v = int(a)
		case 1:
			v = int(b)
		case 2:
			v = int(c)
		case 3:
			v = int(d)
		case 4:
			v = int(e)
		case 5:
			v = int(f)
		case 6:
			v = int(g)
		case 7:
			v = int(h)
		case 8:
			v = int(i)
		case 9:
			v = j()
		}
		coroutine.Yield[int, any](v)
		n++
		return n < 10
	}

	for x() {
	}
}

func Range10Heterogenous() {
	var (
		a int8    = 0
		b int16   = 1
		c int32   = 2
		d int64   = 3
		e uint8   = 4
		f uint16  = 5
		g uint32  = 6
		h uint64  = 7
		i uintptr = 8
	)

	for n := 0; n < 10; n++ {
		var v int
		switch n {
		case 0:
			v = int(a)
		case 1:
			v = int(b)
		case 2:
			v = int(c)
		case 3:
			v = int(d)
		case 4:
			v = int(e)
		case 5:
			v = int(f)
		case 6:
			v = int(g)
		case 7:
			v = int(h)
		case 8:
			v = int(i)
		case 9:
			v = int(n)
		}
		coroutine.Yield[int, any](v)
	}
}

func Select(n int) {
	select {
	default:
		coroutine.Yield[int, any](-1)
	}

	for i := 0; i < n; i++ {
		select {
		case <-time.After(0):
			if i >= 5 {
				break
			}
			coroutine.Yield[int, any](i)
		case <-time.After(1 * time.Second):
			panic("unreachable")
		}

	foo:
		select {
		case <-time.After(0):
			if i >= 6 {
				break foo
			}
			coroutine.Yield[int, any](i * 10)
		}
	}

	select {
	case <-time.After(0):
		for j := 0; j < 3; j++ {
			coroutine.Yield[int, any](j)
		}
	}
}

func YieldingExpressionDesugaring() {
	if x := a(b(1)); x == a(b(2)) {
	} else if y := a(b(3)); y == a(b(4))-1 {
		coroutine.Yield[int, any](a(b(5)) * 10)
	} else if a(b(100)) == 100 {
		panic("unreachable")
	}

	// TODO: support yields in the post iteration statement
	for i := a(b(6)); i < a(b(8)); i++ {
		coroutine.Yield[int, any](70)
	}

	switch x := a(b(9)); x {
	default:
		panic("unreachable")
	case a(b(10)):
		panic("unreachable")
	case a(b(11)):
		panic("unreachable")
	case a(b(12)) - 3: // true!
		a(b(13))
	case a(b(14)):
		panic("unreachable")
	}

	switch x := any(a(b(15))).(type) {
	case bool:
		panic("unreachable")
	case int:
		coroutine.Yield[int, any](x * 10)
	default:
		panic("unreachable")
	}

	// TODO: test select desugaring here too
}

func a(v int) int {
	coroutine.Yield[int, any](v)
	return v
}

func b(v int) int {
	coroutine.Yield[int, any](-v)
	return v
}

func YieldingDurations() {
	t := new(time.Duration)
	*t = time.Duration(100)

	f := func() {
		i := int(t.Nanoseconds())
		*t = time.Duration(i + 1)
		coroutine.Yield[int, any](i)
	}
	for i := 0; i < 10; i++ {
		f()
	}
}

func YieldAndDeferAssign(assign *int, yield, value int) {
	defer func() {
		*assign = value
	}()
	coroutine.Yield[int, any](yield)
}

func RangeYieldAndDeferAssign(n int) {
	for i := 0; i < n; {
		YieldAndDeferAssign(&i, i, i+1)
	}
}

type MethodGeneratorState struct{ i int }

func (s *MethodGeneratorState) MethodGenerator(n int) {
	for s.i = 0; s.i <= n; s.i++ {
		coroutine.Yield[int, any](s.i)
	}
}

func VarArgs(n int) {
	args := make([]int, n)
	for i := range args {
		args[i] = i
	}
	varArgs(args...)
}

func varArgs(args ...int) {
	for _, arg := range args {
		coroutine.Yield[int, any](arg)
	}
}

func ReturnNamedValue() (out int) {
	out = 5
	coroutine.Yield[int, any](11)
	out = 42
	return
}

type Box struct {
	x int
}

func (b *Box) Closure(y int) func(int) {
	return func(z int) {
		coroutine.Yield[int, any](b.x)
		coroutine.Yield[int, any](y)
		coroutine.Yield[int, any](z)
		b.x++
		y++
		z++ // mutation is lost
	}
}

func StructClosure(n int) {
	box := Box{10}
	fn := box.Closure(100)
	for i := 0; i < n; i++ {
		fn(1000)
	}
}

func IdentityGeneric[T any](n T) {
	coroutine.Yield[T, any](n)
}

func IdentityGenericInt(n int) {
	IdentityGeneric[int](n)
}

func IdentityGenericClosure[T any](n T) {
	fn := buildClosure(n)
	fn()
	fn()
}

func buildClosure[T any](n T) func() {
	return func() {
		coroutine.Yield[T, any](n)
	}
}

func IdentityGenericClosureInt(n int) {
	IdentityGenericClosure[int](n)
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type IdentityGenericStruct[T integer] struct {
	n T
}

func (i *IdentityGenericStruct[T]) Run() {
	coroutine.Yield[T, any](i.n)
}

func (i *IdentityGenericStruct[T]) Closure(n T) func(T) {
	return func(x T) {
		coroutine.Yield[T, any](i.n)
		i.n++
		coroutine.Yield[T, any](n)
		n++
		coroutine.Yield[T, any](x)
	}
}

func IdentityGenericStructInt(n int) {
	(&IdentityGenericStruct[int]{n: n}).Run()
}

func IdentityGenericStructClosureInt(n int) {
	fn := (&IdentityGenericStruct[int]{n: n}).Closure(100)
	fn(23)
	fn(45)
}
