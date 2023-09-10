//go:build durable

package coroutine

import (
	"math"
	"reflect"
	"testing"
)

func TestContextSerialization(t *testing.T) {
	original := Context[string, rune]{
		Stack: Stack{
			Frames: []Frame{
				{
					IP: 3,
					Storage: NewStorage([]any{
						1: Int(3),
						5: Int(-1),
					}),
				},
				{
					IP: 5,
					Storage: NewStorage([]any{
						0: Int(4),
						2: Int(math.MaxInt),
					}),
				},
			},
		},
		Heap: NewStorage([]any{
			32: Int(11),
		}),
	}

	b, err := original.MarshalAppend(nil)
	if err != nil {
		t.Fatal(err)
	}

	var reconstructed Context[string, rune]
	if n, err := reconstructed.Unmarshal(b); err != nil {
		t.Fatal(err)
	} else if n != len(b) {
		t.Errorf("not all bytes were consumed when reconstructing the Context: got %d, expected %d", n, len(b))
	}
	if !reflect.DeepEqual(original, reconstructed) {
		t.Error("unexpected context")
		t.Logf("   got: %#v", reconstructed)
		t.Logf("expect: %#v", original)
	}
}
