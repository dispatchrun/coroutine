package examples

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestStruct1(t *testing.T) {
	str := "pointed at"
	//	myint := 999
	s := Struct1{
		Str:  "hello",
		Int:  42,
		Ints: []int64{1, 2, 3},

		Bool:       true,
		Uint64:     11,
		Uint32:     22,
		Uint16:     33,
		Uint8:      44,
		Int64:      -11,
		Int32:      -22,
		Int16:      -33,
		Int8:       -44,
		Float32:    42.11,
		Float64:    420.110,
		Complex64:  42 + 11i,
		Complex128: 420 + 110i,

		FooSer: Foo{t: time.Now()},
		StrPtr: &str,
		//		IntPtr: &myint,
	}

	var b []byte
	b = Serialize_Struct1(nil, s, b)
	s2, b := Deserialize_Struct1(nil, b)

	opts := []cmp.Option{
		cmp.AllowUnexported(Foo{}),
	}

	if diff := cmp.Diff(s, s2, opts...); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}

	if len(b) > 0 {
		t.Fatalf("leftover bytes: %d", len(b))
	}
}
