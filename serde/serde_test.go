package serde_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
	"unsafe"

	serdeinternal "github.com/stealthrocket/coroutine/internal/serde"
	"github.com/stealthrocket/coroutine/serde"
)

type EasyStruct struct {
	A int
	B string
}

func TestReflect(t *testing.T) {
	withBlankTypeMap(func() {
		intv := int(100)
		intp := &intv
		intpp := &intp

		cases := []interface{}{
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
			[]interface{}{},
			*new([]interface{}),
			[]interface{}{nil, nil, nil},
			[]interface{}{nil, 42, nil},
			struct{ a *int }{intp},
			[1][2]int{{1, 2}},

			[]interface{}{(*int)(nil)},
		}

		for _, x := range cases {
			serdeinternal.Types.Add(reflect.TypeOf(x))
		}

		for i, x := range cases {
			x := x
			typ := reflect.TypeOf(x)
			t.Run(fmt.Sprintf("%d-%s", i, typ), func(t *testing.T) {
				b := serdeinternal.Serialize(x)
				out, b := serdeinternal.Deserialize(b)

				assertEqual(t, x, out)

				if len(b) > 0 {
					t.Fatalf("leftover bytes: %d", len(b))
				}
			})
		}
	})
}

func TestReflectCustom(t *testing.T) {
	ser := func(s *serde.Serializer, x *int) error {
		str := strconv.Itoa(*x)
		b := binary.BigEndian.AppendUint64(nil, uint64(len(str)))
		b = append(b, str...)
		serde.Serialize(s, b)
		return nil
	}

	des := func(d *serde.Deserializer, x *int) error {
		var b []byte
		serde.DeserializeTo(d, &b)

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
		serde.RegisterTypeWithSerde[int](ser, des)

		x := 42
		p := &x

		assertRoundTrip(t, p)

		b := serdeinternal.Serialize(p)

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

		serde.RegisterType[X]()
		serde.RegisterTypeWithSerde[int](ser, des)

		x := X{
			foo: "test",
			y:   Y{custom: 42},
		}

		assertRoundTrip(t, x)

		b := serdeinternal.Serialize(x)
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

		serde.RegisterType[X]()
		serde.RegisterTypeWithSerde[int](ser, des)

		x := &X{y: &Y{}}
		x.y.foo = "test"
		x.y.custom = 42
		x.int = &x.y.custom

		assertRoundTrip(t, x)
		b := serdeinternal.Serialize(x)
		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type in slice", func(t *testing.T) {
		serde.RegisterTypeWithSerde[int](ser, des)
		serde.RegisterType[[]int]()
		x := []int{1, 2, 3, 42, 5, 6}
		assertRoundTrip(t, x)
		b := serdeinternal.Serialize(x)
		if !bytes.Contains(b, int42) {
			t.Fatalf("custom serde was not used:\ngot: %v\nexpected: %v", b, int42)
		}
	})

	testReflect(t, "custom type of struct", func(t *testing.T) {
		ser := func(s *serde.Serializer, x *http.Client) error {
			i := uint64(x.Timeout)
			serde.Serialize(s, i)
			return nil
		}

		des := func(d *serde.Deserializer, x *http.Client) error {
			var i uint64
			serde.DeserializeTo(d, &i)
			x.Timeout = time.Duration(i)
			return nil
		}

		serde.RegisterTypeWithSerde[http.Client](ser, des)

		x := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return nil
			},
			Timeout: 42000,
		}

		// Without custom serializer, it would panic because of the
		// unserializable function in CheckRedirect.

		b := serdeinternal.Serialize(x)
		out, b := serdeinternal.Deserialize(b)

		assertEqual(t, x.Timeout, out.(http.Client).Timeout)

		if len(b) > 0 {
			t.Fatalf("leftover bytes: %d", len(b))
		}

	})
}

func TestReflectSharing(t *testing.T) {
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

		serde.RegisterType[X]()

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

		serde.RegisterType[X]()
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

		serde.RegisterType[X]()

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

		serde.RegisterType[X]()
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

		serde.RegisterType[X]()
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

		serde.RegisterType[X]()
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

		serde.RegisterType[X]()

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

		serde.RegisterType[map[string][]int]()
		out := assertRoundTrip(t, x)

		out["un"][0] = 100
		out["deux"][1] = 200
		out["trois"][2] = 300

		assertEqual(t, []int{100, 200, 300}, out["trois"])
	})
}

func assertEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Error("unexpected value:")
		t.Logf("   got: %#v", actual)
		t.Logf("expect: %#v", expected)
	}
}

func assertRoundTrip[T any](t *testing.T, orig T) T {
	t.Helper()

	b := serdeinternal.Serialize(orig)
	out, b := serdeinternal.Deserialize(b)

	assertEqual(t, orig, out)

	if len(b) > 0 {
		t.Fatalf("leftover bytes: %d", len(b))
	}

	return out.(T)
}

func withBlankTypeMap(f func()) {
	oldtm := serdeinternal.Types
	serdeinternal.Types = serdeinternal.NewTypeMap()
	defer func() { serdeinternal.Types = oldtm }()

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
