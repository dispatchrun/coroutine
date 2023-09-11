//go:build durable

package coroutine

import (
	"math"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestContextSerialization(t *testing.T) {
	original := Context[string, rune]{
		Stack: Stack{
			Frames: []Frame{
				{
					IP: 3,
					Storage: NewStorage([]any{
						1: int(3),
						5: int(-1),
					}),
				},
				{
					IP: 5,
					Storage: NewStorage([]any{
						0: int(4),
						2: int(math.MaxInt),
					}),
				},
			},
		},
		// TODO: heap ignored for now
		// Heap: NewStorage([]any{
		// 	32: int(11),
		// }),
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

	if diff := cmp.Diff(original, reconstructed, cmp.AllowUnexported(reconstructed, Storage{})); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}

	if !reflect.DeepEqual(original, reconstructed) {
		t.Error("unexpected context")
		t.Logf("   got: %#v", reconstructed)
		t.Logf("expect: %#v", original)
	}
}
