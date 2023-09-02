package coroutine

import (
	"fmt"
	"math"
	"testing"
)

func TestBuiltinSerialization(t *testing.T) {
	for _, s := range []Serializable{
		Int(0),
		Int(-1),
		Int(math.MaxInt),
	} {
		t.Run(fmt.Sprintf("%#v", s), func(t *testing.T) {
			b, err := MarshalAppend(nil, s)
			if err != nil {
				t.Fatal(err)
			}
			reconstructed, n, err := Unmarshal(b)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(b) {
				t.Errorf("not all bytes in the buffer were used to reconstruct value: got %d, expect %d", n, len(b))
			}
			if s != reconstructed {
				t.Fatalf("reconstructed value mismatch: got %d, expect %d", reconstructed, s)
			}
		})
	}
}
