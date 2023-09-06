package examples

import "time"

type Foo struct {
	t time.Time
}

func (f *Foo) MarshalAppend(b []byte) ([]byte, error) {
	out := f.t.Format(time.RFC3339Nano)
	return append(b, []byte(out)...), nil
}

func (f *Foo) Unmarshal(b []byte) (int, error) {
	n := len(time.RFC3339Nano)
	v := string(b[:n])
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return n, err
	}
	f.t = t
	return n, nil
}

type Struct1 struct {
	Str  string
	Int  int64
	Ints []int64

	Bool       bool
	Uint64     uint64
	Uint32     uint32
	Uint16     uint16
	Uint8      uint8
	Int64      int64
	Int32      int32
	Int16      int16
	Int8       int8
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128

	FooSer Foo
}
