//go:build !durable

package main

import (
	"testing"

	"github.com/stealthrocket/coroutine"
)

func assertSerializable[R, S any](t *testing.T, g coroutine.Generator[R, S]) coroutine.Generator[R, S] {
	// Not serializable unless the "durable" build tag is present.
	return g
}
