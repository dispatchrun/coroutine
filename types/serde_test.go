package types

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
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
	b := Serialize(x)
	out, _ := Deserialize(b)

	if !x.Equal(out.(time.Time)) {
		t.Errorf("expected %v, got %v", x, out)
	}
}

type EasyStruct struct {
	A int
	B string
}

func TestReflect(t *testing.T) {
	withBlankTypeMap(func() {
		intv := int(100)
		intp := &intv
		intpp := &intp
		type ctxKey1 struct{}

		cases := []any{
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

			func() {},
			func(int) int { return 42 },

			[1]*int{intp},

			context.Background(),
			context.TODO(),
			context.WithValue(context.Background(), ctxKey1{}, "hello"),

			"",
			struct{}{},
			errors.New("test"),
		}

		for _, x := range cases {
			t := reflect.TypeOf(x)

			if t.Kind() == reflect.Func {
				a := FuncAddr(x)
				f := FuncByAddr(a)
				f.Type = t
			}
		}

		for i, x := range cases {
			x := x
			typ := reflect.TypeOf(x)
			t.Run(fmt.Sprintf("%d-%s", i, typ), func(t *testing.T) {
				b := Serialize(x)
				out, b := Deserialize(b)

				assertEqual(t, x, out)

				if len(b) > 0 {
					t.Fatalf("leftover bytes: %d", len(b))
				}
			})
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

		b := Serialize(p)

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

		b := Serialize(x)
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
		b := Serialize(x)
		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type in slice", func(t *testing.T) {
		Register[int](ser, des)
		x := []int{1, 2, 3, 42, 5, 6}
		assertRoundTrip(t, x)
		b := Serialize(x)
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

		b := Serialize(x)
		out, b := Deserialize(b)

		assertEqual(t, x.Timeout, out.(http.Client).Timeout)

		if len(b) > 0 {
			t.Fatalf("leftover bytes: %d", len(b))
		}

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

	if t1.Kind() == reflect.Func {
		return FuncAddr(v1) == FuncAddr(v2)
	}

	return reflect.DeepEqual(v1, v2)
}

func assertRoundTrip[T any](t *testing.T, orig T) T {
	t.Helper()

	b := Serialize(orig)
	out, b := Deserialize(b)

	assertEqual(t, orig, out)

	if len(b) > 0 {
		t.Fatalf("leftover bytes: %d", len(b))
	}

	return out.(T)
}

func withBlankTypeMap(f func()) {
	oldtm := types
	types = newTypemap()
	defer func() { types = oldtm }()

	f()
}

func testReflect(t *testing.T, name string, f func(t *testing.T)) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		withBlankTypeMap(func() {
			f(t)
		})
	})
}
