package coroutine

import (
	"math"
	"reflect"
	"testing"
)

func TestContextSerialization(t *testing.T) {
	original := Context{
		Stack: Stack{
			Frames: []Frame{
				{
					IP: 3,
					Storage: NewStorage([]Serializable{
						1: Int(3),
						5: Int(-1),
					}),
				},
				{
					IP: 5,
					Storage: NewStorage([]Serializable{
						0: Int(4),
						2: Int(math.MaxInt),
					}),
				},
			},
		},
		Heap: NewStorage([]Serializable{
			32: Int(11),
		}),
	}

	b, err := original.MarshalAppend(nil)
	if err != nil {
		t.Fatal(err)
	}

	var reconstructed Context
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
