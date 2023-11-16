package coroutine

// TOOD: go generate this file and the getg_{arch}.go files containing the
// sizes of g structs.

// This file contains parts of the Go runtime source needed to implement the
// goroutine local storage capabilities that use use in this package.
//
// See: https://github.com/golang/go/blob/master/src/runtime/runtime2.go

type stack struct {
	_, hi uintptr
}

type g struct {
	stack stack
}

func getg() *g
