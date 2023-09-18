//go:build durable

package compiler

import (
	"testing"

	"github.com/stealthrocket/coroutine"
)

func assertSerializable[R, S any](t *testing.T, g coroutine.Coroutine[R, S]) coroutine.Coroutine[R, S] {
	c := g.Context()
	b, err := c.MarshalAppend(nil)
	if err != nil {
		t.Fatal(err)
	}
	var reconstructed coroutine.Context[R, S]
	if n, err := reconstructed.Unmarshal(b); err != nil {
		t.Fatal(err)
	} else if n != len(b) {
		t.Fatal("invalid number of bytes read when reconstructing context")
	}
	*c = reconstructed
	return g
}
