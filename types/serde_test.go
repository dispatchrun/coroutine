package types

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
	"unsafe"
)

func TestSerdeTime(t *testing.T) {
	t.Run("time zero", func(t *testing.T) {
		testSerdeTime(t, time.Time{})
	})

	t.Run("time.Now", func(t *testing.T) {
		testSerdeTime(t, time.Now())
	})

	t.Run("fixed zone", func(t *testing.T) {
		parsed, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
		if err != nil {
			t.Error(err)
		}
		loc, err := time.LoadLocation("US/Eastern")
		if err != nil {
			t.Error("failed to load location", err)
		}
		t2 := parsed.In(loc)

		testSerdeTime(t, t2)
	})
}

func testSerdeTime(t *testing.T, x time.Time) {
	b, err := Serialize(x)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Deserialize(b)
	if err != nil {
		t.Fatal(err)
	}

	assertCanInspect(t, b)

	if !x.Equal(out.(time.Time)) {
		t.Errorf("expected %v, got %v", x, out)
	}
}

func assertCanInspect(t *testing.T, b []byte) {
	c, err := Inspect(b)
	if err != nil {
		t.Fatal(err)
	}
	regions := []*Region{c.Root()}
	for i := 0; i < c.NumRegion(); i++ {
		regions = append(regions, c.Region(i))
	}
	for _, region := range regions {
		if typ := region.Type(); typ.Package() == "reflect" {
			// FIXME
			continue
		}
		s := region.Scan()
		for s.Next() {
			//
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

type EasyStruct struct {
	A int
	B string
}

type funcType func(int) error

func identity(v int) int { return v }

func TestReflect(t *testing.T) {
	intv := int(100)
	intp := &intv
	intpp := &intp
	type ctxKey1 struct{}

	RegisterFunc[func(int) int]("github.com/stealthrocket/coroutine/types.identity")

	var emptyMap map[string]struct{}

	cases := []any{
		map[string]*EasyStruct{"a": {A: 30}},
		map[string]map[string]*EasyStruct{"a": {"b": {A: 30}}},
		map[string][1]*EasyStruct{"a": {{A: 30}}},
		"foo",
		true,
		int(42),
		int64(11),
		int32(10),
		int16(9),
		int8(8),
		uint(42),
		uint64(11),
		uint32(10),
		uint16(9),
		uint8(8),
		intp,
		intpp,
		[2]int{1, 2},
		[]int{1, 2, 3},
		map[string]int{"one": 1, "two": 2},
		emptyMap,
		EasyStruct{
			A: 52,
			B: "test",
		},
		[]any{},
		*new([]any),
		[]any{nil, nil, nil},
		[]any{nil, 42, nil},
		struct{ a *int }{intp},
		struct{ a, b int }{a: 1, b: 2},
		[1][2]int{{1, 2}},

		[]any{(*int)(nil)},

		funcType(nil),

		func() {},
		func(int) int { return 42 },

		[1]*int{intp},

		context.Background(),
		context.TODO(),
		context.WithValue(context.Background(), ctxKey1{}, "hello"),

		"",
		struct{}{},
		errors.New("test"),
		unsafe.Pointer(nil),

		// Primitives
		reflect.ValueOf("foo"),
		reflect.ValueOf(true),
		reflect.ValueOf(int(1)),
		reflect.ValueOf(int8(math.MaxInt8)),
		reflect.ValueOf(int16(-math.MaxInt16)),
		reflect.ValueOf(int32(math.MaxInt32)),
		reflect.ValueOf(int64(-math.MaxInt64)),
		reflect.ValueOf(uint(1)),
		reflect.ValueOf(uint8(math.MaxUint8)),
		reflect.ValueOf(uint16(math.MaxUint16)),
		reflect.ValueOf(uint32(math.MaxUint8)),
		reflect.ValueOf(uint64(math.MaxUint64)),
		reflect.ValueOf(float32(3.14)),
		reflect.ValueOf(float64(math.MaxFloat64)),

		// Arrays
		reflect.ValueOf([32]byte{0: 1, 15: 2, 31: 3}),

		// Slices
		reflect.ValueOf([]byte("foo")),
		reflect.ValueOf([][]byte{[]byte("foo"), []byte("bar")}),
		reflect.ValueOf([]string{"foo", "bar"}),
		reflect.ValueOf([]int{}),
		reflect.ValueOf([]int(nil)),

		// Maps
		reflect.ValueOf(map[string]string{"foo": "bar", "abc": "xyz"}),
		reflect.ValueOf(http.Header{"Content-Length": []string{"11"}, "X-Forwarded-For": []string{"1.1.1.1", "2.2.2.2"}}),
		reflect.ValueOf(emptyMap),

		// Structs
		reflect.ValueOf(struct{ A, B int }{1, 2}),

		// Pointers
		reflect.ValueOf(errors.New("fail")),

		// Funcs
		reflect.ValueOf(identity),
		reflect.ValueOf(funcType(nil)),
	}

	for _, x := range cases {
		t := reflect.TypeOf(x)

		if t.Kind() == reflect.Func {
			a := FuncAddr(x)
			if f := FuncByAddr(a); f != nil {
				f.Type = t
			}
		}
	}

	for i, x := range cases {
		x := x
		typ := reflect.TypeOf(x)
		t.Run(fmt.Sprintf("%d-%s", i, typ), func(t *testing.T) {
			b, err := Serialize(x)
			if err != nil {
				t.Fatal(err)
			}
			out, err := Deserialize(b)
			if err != nil {
				t.Fatal(err)
			}

			assertEqual(t, x, out)

			assertCanInspect(t, b)
		})
	}
}

func TestReflectUnsafePointer(t *testing.T) {
	type unsafePointerStruct struct{ p unsafe.Pointer }
	var selfRef unsafePointerStruct
	selfRef.p = unsafe.Pointer(&selfRef)

	b, err := Serialize(&selfRef)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Deserialize(b)
	if err != nil {
		t.Fatal(err)
	}

	res := out.(*unsafePointerStruct)
	if unsafe.Pointer(res) != unsafe.Pointer(res.p) {
		t.Errorf("unsafe.Pointer was not restored correctly")
	}
}

func TestReflectFunc(t *testing.T) {
	RegisterFunc[func(int) int]("github.com/stealthrocket/coroutine/types.identity")

	b, err := Serialize(reflect.ValueOf(identity))
	if err != nil {
		t.Fatal(err)
	}

	out, err := Deserialize(b)
	if err != nil {
		t.Fatal(err)
	}

	fn := out.(reflect.Value)
	res := fn.Call([]reflect.Value{reflect.ValueOf(3)})
	if len(res) != 1 || !res[0].CanInt() || res[0].Int() != 3 {
		t.Errorf("unexpected identity(3) results: %#v", res)
	}
}

func TestReflectClosure(t *testing.T) {
	v := 3
	fn := func() int {
		return v
	}

	RegisterClosure[func() int, struct {
		F  uintptr
		X0 int
	}]("github.com/stealthrocket/coroutine/types.TestReflectClosure.func1")

	t.Run("raw", func(t *testing.T) {
		b, err := Serialize(fn)
		if err != nil {
			t.Fatal(err)
		}

		out, err := Deserialize(b)
		if err != nil {
			t.Fatal(err)
		}

		rfn := out.(func() int)
		if res := rfn(); res != v {
			t.Errorf("unexpected closure call results: %#v", res)
		}
	})

	t.Run("reflect-value", func(t *testing.T) {
		// FIXME: get reflect.Value(closure) working
		t.Skipf("reflect.Value(closure) is not working correctly")

		b, err := Serialize(reflect.ValueOf(fn))
		if err != nil {
			t.Fatal(err)
		}

		out, err := Deserialize(b)
		if err != nil {
			t.Fatal(err)
		}

		rfn := out.(reflect.Value)
		if res := rfn.Call(nil); len(res) != 1 || !res[0].CanInt() || res[0].Int() != int64(v) {
			t.Errorf("unexpected closure reflect call results: %#v", res)
		}
	})
}

func TestErrors(t *testing.T) {
	s := struct {
		X5 error
	}{}

	assertRoundTrip(t, s)
}

func TestEmptyStructs(t *testing.T) {
	assertRoundTrip(t, struct{}{})

	type X struct {
		first struct{}
		last  int
	}

	assertRoundTrip(t, X{})
	assertRoundTrip(t, X{first: struct{}{}, last: 42})

	type Y struct {
		first int
		last  struct{}
	}

	assertRoundTrip(t, Y{})
	assertRoundTrip(t, Y{first: 42, last: struct{}{}})
}

func TestInt257(t *testing.T) {
	one := 1
	x := []any{
		true,
		one,
	}
	assertRoundTrip(t, x)
}

func TestReflectCustom(t *testing.T) {
	ser := func(s *Serializer, x *int) error {
		str := strconv.Itoa(*x)
		b := binary.BigEndian.AppendUint64(nil, uint64(len(str)))
		b = append(b, str...)
		SerializeT(s, b)
		return nil
	}

	des := func(d *Deserializer, x *int) error {
		var b []byte
		DeserializeTo(d, &b)

		n := binary.BigEndian.Uint64(b[:8])
		b = b[8:]
		s := string(b[:n])
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*x = i
		return nil
	}

	// bytes created by ser(42):
	int42 := []byte{
		// big endian size
		0, 0, 0, 0, 0, 0, 0, 2,
		// the int as a string
		52, // 4
		50, // 2
	}

	testReflect(t, "int wrapper", func(t *testing.T) {
		Register[int](ser, des)

		x := 42
		p := &x

		assertRoundTrip(t, p)

		b, err := Serialize(p)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type in field", func(t *testing.T) {
		type Y struct {
			custom int
		}
		type X struct {
			foo string
			y   Y
		}

		Register[int](ser, des)

		x := X{
			foo: "test",
			y:   Y{custom: 42},
		}

		assertRoundTrip(t, x)

		b, err := Serialize(x)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type in field of pointed at struct", func(t *testing.T) {
		type Y struct {
			foo    string
			custom int
		}
		type X struct {
			int *int
			y   *Y
		}

		Register[int](ser, des)

		x := &X{y: &Y{}}
		x.y.foo = "test"
		x.y.custom = 42
		x.int = &x.y.custom

		assertRoundTrip(t, x)
		b, err := Serialize(x)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type in slice", func(t *testing.T) {
		Register[int](ser, des)
		x := []int{1, 2, 3, 42, 5, 6}
		assertRoundTrip(t, x)
		b, err := Serialize(x)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type of struct", func(t *testing.T) {
		ser := func(s *Serializer, x *http.Client) error {
			i := uint64(x.Timeout)
			SerializeT(s, i)
			return nil
		}

		des := func(d *Deserializer, x *http.Client) error {
			var i uint64
			DeserializeTo(d, &i)
			x.Timeout = time.Duration(i)
			return nil
		}

		Register[http.Client](ser, des)

		x := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
			Timeout: 42000,
		}

		// Without custom serializer, it would panic because of the
		// unserializable function in CheckRedirect.

		b, err := Serialize(x)
		if err != nil {
			t.Fatal(err)
		}

		out, err := Deserialize(b)
		if err != nil {
			t.Fatal(err)
		}

		assertEqual(t, x.Timeout, out.(http.Client).Timeout)
	})
}

func TestReflectSharing(t *testing.T) {
	testReflect(t, "maps of ints", func(t *testing.T) {
		m := map[int]int{1: 2, 3: 4}

		type X struct {
			a map[int]int
			b map[int]int
		}

		x := X{
			a: m,
			b: m,
		}

		// make sure map is shared beforehand
		x.a[5] = 6
		assertEqual(t, 6, x.b[5])

		out := assertRoundTrip(t, x)

		// check map is shared after
		out.a[7] = 8
		assertEqual(t, 8, out.b[7])
	})

	testReflect(t, "slice backing array", func(t *testing.T) {
		data := make([]int, 10)
		for i := range data {
			data[i] = i
		}

		type X struct {
			s1 []int
			s2 []int
			s3 []int
		}

		orig := X{
			s1: data[0:3],
			s2: data[2:8],
			s3: data[7:10],
		}
		assertEqual(t, []int{0, 1, 2}, orig.s1)
		assertEqual(t, []int{2, 3, 4, 5, 6, 7}, orig.s2)
		assertEqual(t, []int{7, 8, 9}, orig.s3)

		assertEqual(t, 10, cap(orig.s1))
		assertEqual(t, 3, len(orig.s1))
		assertEqual(t, 8, cap(orig.s2))
		assertEqual(t, 6, len(orig.s2))
		assertEqual(t, 3, cap(orig.s3))
		assertEqual(t, 3, len(orig.s3))

		out := assertRoundTrip(t, orig)

		// verify that the initial arrays were shared
		orig.s1[2] = 42
		assertEqual(t, 42, orig.s2[0])
		orig.s2[5] = 11
		assertEqual(t, 11, orig.s3[0])

		// verify the result's underlying array is shared
		out.s1[2] = 42
		assertEqual(t, 42, out.s2[0])
		out.s2[5] = 11
		assertEqual(t, 11, out.s3[0])
	})

	testReflect(t, "slice backing array with set capacities", func(t *testing.T) {
		data := make([]int, 10)
		for i := range data {
			data[i] = i
		}

		type X struct {
			s1 []int
			s2 []int
			s3 []int
		}

		orig := X{
			s1: data[0:3:3],
			s2: data[2:8:8],
			s3: data[7:10:10],
		}
		assertEqual(t, []int{0, 1, 2}, orig.s1)
		assertEqual(t, []int{2, 3, 4, 5, 6, 7}, orig.s2)
		assertEqual(t, []int{7, 8, 9}, orig.s3)

		assertEqual(t, 3, cap(orig.s1))
		assertEqual(t, 3, len(orig.s1))
		assertEqual(t, 6, cap(orig.s2))
		assertEqual(t, 6, len(orig.s2))
		assertEqual(t, 3, cap(orig.s3))
		assertEqual(t, 3, len(orig.s3))

		out := assertRoundTrip(t, orig)

		// verify that the initial arrays were shared
		orig.s1[2] = 42
		assertEqual(t, 42, orig.s2[0])
		orig.s2[5] = 11
		assertEqual(t, 11, orig.s3[0])

		// verify the result's underlying array is shared
		out.s1[2] = 42
		assertEqual(t, 42, out.s2[0])
		out.s2[5] = 11
		assertEqual(t, 11, out.s3[0])
	})

	testReflect(t, "struct fields extra pointers", func(t *testing.T) {
		type A struct {
			X, Y int
		}

		type B struct {
			P *int
		}

		type X struct {
			B *B
			A *A
			// putting A after B to make sure A gets serialized
			// first because of dependencies, not just because it's
			// earlier than B in the fields list.
		}

		x := X{
			A: new(A),
			B: new(B),
		}
		x.B.P = &x.A.Y

		// verify the original pointer is correct
		x.A.Y = 42
		assertEqual(t, 42, *x.B.P)

		out := assertRoundTrip(t, x)

		// verify the resulting pointer is correct
		out.A.Y = 11
		assertEqual(t, 11, *out.B.P)
	})

	testReflect(t, "struct with pointer to itself", func(t *testing.T) {
		type X struct {
			z *X
		}

		x := &X{}
		x.z = x
		assertEqual(t, x, x.z)

		out := assertRoundTrip(t, x)

		assertEqual(t, out, out.z)
	})

	testReflect(t, "nested struct fields", func(t *testing.T) {
		type Z struct {
			v int64
		}
		type Y struct {
			v Z
		}
		type X struct {
			v Y
		}

		x := X{Y{Z{42}}}

		assertRoundTrip(t, x)
	})

	testReflect(t, "nested struct fields not first", func(t *testing.T) {
		type Z struct {
			v int64
		}
		type Y struct {
			b int
			v Z
		}
		type X struct {
			a int
			v Y
		}

		x := X{a: 1,
			v: Y{
				b: 2,
				v: Z{42},
			},
		}

		assertRoundTrip(t, x)
	})

	testReflect(t, "pointer intra struct field", func(t *testing.T) {
		type Z struct {
			v string
		}
		type Y struct {
			z *Z
		}
		type X struct {
			z Z
			y Y
		}

		x := &X{}
		x.z.v = "hello"
		x.y.z = &x.z

		assertEqual(t, unsafe.Pointer(x), unsafe.Pointer(x.y.z))

		out := assertRoundTrip(t, x)

		out.z.v = "test"

		assertEqual(t, "test", out.y.z.v)
	})

	testReflect(t, "slices with same backing array but no joined cap", func(t *testing.T) {
		data := make([]int, 10)
		for i := range data {
			data[i] = i
		}

		assertEqual(t, 10, cap(data))

		type X struct {
			s1 []int
			s2 []int
		}

		x := X{
			s1: data[0:3:3],
			s2: data[8:10:10],
		}

		assertEqual(t, 3, cap(x.s1))
		assertEqual(t, 2, cap(x.s2))

		out := assertRoundTrip(t, x)

		// check underlying arrays are not shared
		out.s1 = append(out.s1, 1, 1, 1, 1, 1, 1)
		assertEqual(t, 8, out.s2[0])
	})

	testReflect(t, "pointers to shared data in maps", func(t *testing.T) {
		data := make([]int, 3)
		for i := range data {
			data[i] = i
		}

		x := map[string][]int{
			"un":    data[0:1],
			"deux":  data[0:2],
			"trois": data[0:3],
		}

		out := assertRoundTrip(t, x)

		out["un"][0] = 100
		out["deux"][1] = 200
		out["trois"][2] = 300

		assertEqual(t, []int{100, 200, 300}, out["trois"])
	})
}

func assertEqual(t *testing.T, expected, actual any) {
	t.Helper()

	if !deepEqual(expected, actual) {
		t.Error("unexpected value:")
		t.Logf("   got: %#v", actual)
		t.Logf("expect: %#v", expected)
	}
}

func deepEqual(v1, v2 any) bool {
	t1 := reflect.TypeOf(v1)
	t2 := reflect.TypeOf(v2)

	if t1 != t2 {
		return false
	}

	if t1 == reflect.TypeOf(reflect.Value{}) {
		return equalReflectValue(v1.(reflect.Value), v2.(reflect.Value))
	}

	if t1.Kind() == reflect.Func {
		return FuncAddr(v1) == FuncAddr(v2)
	}

	return reflect.DeepEqual(v1, v2)
}

func equalReflectValue(v1, v2 reflect.Value) bool {
	if v1.Type() != v2.Type() {
		return false
	}
	switch v1.Kind() {
	case reflect.Bool:
		return v1.Bool() == v2.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v1.Int() == v2.Int()

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v1.Uint() == v2.Uint()

	case reflect.Float32, reflect.Float64:
		return v1.Float() == v2.Float()

	case reflect.Complex64, reflect.Complex128:
		return v1.Complex() == v2.Complex()

	case reflect.String:
		return v1.String() == v2.String()

	case reflect.Array:
		if v1.Len() != v2.Len() {
			return false
		}
		for i := 0; i < v1.Len(); i++ {
			if !equalReflectValue(v1.Index(i), v2.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Slice:
		if v1.Len() != v2.Len() {
			return false
		} else if v1.Cap() != v2.Cap() {
			return false
		}
		for i := 0; i < v1.Len(); i++ {
			if !equalReflectValue(v1.Index(i), v2.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Map:
		if v1.Len() != v2.Len() {
			return false
		}
		it := v1.MapRange()
		for it.Next() {
			k := it.Key()
			mv1 := it.Value()
			mv2 := v2.MapIndex(k)
			if !equalReflectValue(mv1, mv2) {
				return false
			}
		}
		return true

	case reflect.Struct:
		for i := 0; i < v1.NumField(); i++ {
			if !equalReflectValue(v1.Field(i), v2.Field(i)) {
				return false
			}
		}
		return true

	case reflect.Func:
		return v1.UnsafePointer() == v2.UnsafePointer()

	case reflect.Pointer:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		return v1.IsNil() || equalReflectValue(v1.Elem(), v2.Elem())

	default:
		panic(fmt.Sprintf("not implemented: comparison of reflect.Value with type %s", v1.Type()))
	}
}

func assertRoundTrip[T any](t *testing.T, orig T) T {
	t.Helper()

	b, err := Serialize(orig)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Deserialize(b)
	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, orig, out)

	return out.(T)
}

func testReflect(t *testing.T, name string, f func(t *testing.T)) {
	t.Helper()
	t.Run(name, f)
}
