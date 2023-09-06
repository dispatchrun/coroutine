package examples

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func enableDebugLogs() {
	removeTime := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		return a
	}

	var programLevel = new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:       programLevel,
		ReplaceAttr: removeTime,
	})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)
}

func TestStruct1Empty(t *testing.T) {
	enableDebugLogs()

	s := Struct1{}

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

func TestStruct1(t *testing.T) {
	enableDebugLogs()

	str := "pointed at"
	myint := 999
	myintptr := &myint

	bounce1 := &Bounce{
		Value: 1,
	}
	bounce2 := &Bounce{
		Value: 2,
	}
	bounce1.Other = bounce2
	bounce2.Other = bounce1

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

		FooSer:    Foo{t: time.Now()},
		StrPtr:    &str,
		IntPtr:    &myint,
		IntPtrPtr: &myintptr,

		InnerV: Inner{
			A: 53,
			B: "test",
		},

		InnerP: &Inner{
			A: 99,
			B: "hello",
		},

		Bounce1: bounce1,
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
