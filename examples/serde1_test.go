package examples

import (
	"reflect"
	"testing"
)

func TestStruct1(t *testing.T) {
	s := Struct1{
		Str: "hello",
		Int: 42,
	}

	var b []byte
	b = Serialize_gen0(s, b)
	s2, b := Deserialize_gen0(b)

	if s != s2 {
		t.Fatal("s != s2")
	}
	t1 := reflect.TypeOf(s)
	t2 := reflect.TypeOf(s2)

	if t1 != t2 {
		t.Fatalf("reflect types don't match: %s; expected %s", t2, t1)
	}

	if len(b) > 0 {
		t.Fatalf("leftover bytes: %d", len(b))
	}
}
