package coroutine

import (
	"fmt"
	"reflect"
	"testing"
)

type EasyStruct struct {
	A int
	B string
}

func TestReflect(t *testing.T) {
	// override type map and manually register types to avoid using the
	// genereated codecs
	oldtm := tm
	tm = newTypeMap()
	defer func() {
		tm = oldtm
	}()

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
	}

	for _, x := range cases {
		tm.Add(reflect.TypeOf(x))
	}

	for i, x := range cases {
		x := x
		typ := reflect.TypeOf(x)
		t.Run(fmt.Sprintf("%d-%s", i, typ), func(t *testing.T) {
			var b []byte

			b = Serialize(x, b)
			out, b := Deserialize(b)

			if !reflect.DeepEqual(x, out) {
				t.Error("unexpected context")
				t.Logf("   got: %#v", out)
				t.Logf("expect: %#v", x)
			}

			if len(b) > 0 {
				t.Fatalf("leftover bytes: %d", len(b))
			}
		})
	}
}
