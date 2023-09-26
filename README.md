[![build](https://github.com/stealthrocket/coroutine/actions/workflows/build.yml/badge.svg)](https://github.com/stealthrocket/coroutine/actions/workflows/build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/stealthrocket/coroutine.svg)](https://pkg.go.dev/github.com/stealthrocket/coroutine)
[![Apache 2 License](https://img.shields.io/badge/license-Apache%202-blue.svg)](LICENSE)

# coroutine

This project contains a durable coroutine compiler and runtime library for Go.

## Usage

The `coroutine` package can be used as a simple library to create coroutines in
a Go program, allowing the function passed as entry point to the coroutine to
be paused by at yield points and later resumed by the caller.

When pausing, the coroutine yields a value that is received by the caller, and
on resumption the caller can send back a value that the coroutine obtains as
result.

### Creating Coroutines

The following code example shows how to create a coroutine and yield values to
the caller:

```go
// main.go
package main

import "github.com/stealthrocket/coroutine"

func main() {
    coro := coroutine.New[int, any](func() {
        for i := 0; i < 3; i++ {
            coroutine.Yield[int, any](i)
        }
    })

    for coro.Next() {
        println(coro.Recv())
    }
}
```
Executing the program produces the following output:
```
$ go run main.go
0
1
2
```

The `coroutine.New` and `coroutine.Yield` functions are the main functions that
applications would use to create coroutines and declare yield points.
An important observation to make here is the fact that the functions have two
generic type parameters (we name then `R` and `S`) to declare the type of values
that the program can receive from and send to the coroutine. The types passed to
`Yield` must correspond to the types used when creating the coroutine, if the
types mismatch, the coroutine panics at the yield point.

### Terminating Coroutines

Coroutines hold state in order to resume from the yield point and ensure that
operations such as deferred function calls will be executed, therefore they need
to be driven to completion by calling `Next` until it returns `false`, which
indicates that the entry point of the coroutine as exited.

Sometimes, a program may need to prematurely interrupt a coroutine before it
reached completion; this can be done by calling `Stop`, which prevents returning
from the current yield point. `Stop` only marks the coroutine as interrupted, it
is still necessary for the program to call `Next` in order to drive the code to
completion, running deferred function calls, and returning from the entry point.

> **Note**
> yielding from a coroutine after it was stopped is a fatal error;
> therefore, it is not advised to yield from defers as it would result in
> a crash if the defer was executed after stopping the coroutine.

Often times, the simplest construct to drive coroutine executions is to use the
`Run` function:
```go
package main

import "github.com/stealthrocket/coroutine"

func main() {
    coro := coroutine.New[int, any](func() {
        for i := 0; i < 3; i++ {
            coroutine.Yield[int, any](i)
        }
    })

    // The callback function is invoked for each value received from the
    // coroutine, and returns the value that will be sent back.
    coroutine.Run(coro, func(v int) any {
        println(v)
        return nil
    })
}
```

### Using Coroutines

Coroutines can be a powerful building block to represent cooperative scheduling
constructs, the program has complete autonomy when it comes to deciding when
a coroutine is resumed and gets to carry on to its next operation.

The part of the program driving the execution of coroutines can be seen as a
local scheduler for a sub-part of the code.

Another useful property of coroutines is that, just like functions, they can be
composed. A parent coroutine can create more coroutines for which it can drive
execution in a sub context of the program, and yield values to its caller that
applied computations from the values received from the sub-routines.

## Durable Coroutines

_Coroutines are functions that can be suspended and resumed. Durable coroutines
are functions that can be suspended, serialized and resumed in another process._

> **Warning**
> This section documents highly experimental capabilities of the coroutine
> package, changes will likely be made, use at your own risks!

The project contains a source-to-source Go compiler named `coroc` which compiles
**volatile** coroutines into **durable** versions where the state captured by a
coroutine can be encoded into a binary representation that can be saved to a
storage medium and later reloaded to restart computation where it was left off.

### Generate durable programs

To install the `coroc` compiler:
```
go install github.com/stealthrocket/coroutine/compiler/cmd/coroc
```

Then to compile a package:
```
coroc ./path/to/package
```
This will generate files named `*_durable.go` and set build tags on source files
that need to be excluded when building in durable mode. Because the compiler may
need to generate coroutines in code paths of the standard Go library, it creates
a vendors a copy under `vendor/goroot`.

The standard Go toolchain can then be used to compile the application in durable
mode:
```
GOROOT=$PWD/vendor/goroot go build -tags durable .
```

**Pro tip**
A common pattern is to use a `go:generate` directive in the main application
package to trigger the compilation of the durable files:
```go
//go:generate coroc

package main

func main() {
    ...
}
```
This creates a tighter integration with the Go toolchain, as the compilation of
durable files can be executed with:
```
go generate
```

### Saving and Restoring

In durable mode, the state of `coroutine.Coroutine` values can be marshaled into
a byte slice for storage beyond the lifetime of the application, and resumed on
restart.

The following program is adapted from the previous example to save and restore
the coroutine state from a file. In volatile mode, it only prints the first
yielded value; but in durable mode, progress made by the coroutine is captured
and restored over multiple executions.

```go
//go:generate coroc

package main

import (
    "errors"
    "flag"
    "io/fs"
    "log"

    "github.com/stealthrocket/coroutine"
)

func main() {
    var state string
    flag.StringVar(&state, "state", "coroutine.state", "Location of the coroutine state file")
    flag.Parse()

    coro := coroutine.New[int, any](func() {
        for i := 0; i < 3; i++ {
            coroutine.Yield[int, any](i)
        }
    })

    if coroutine.Durable {
        b, err := os.ReadFile(state)
        if err != nil {
            if !errors.Is(err, fs.ErrNotExist) {
                log.Fatal(err)
            }
        } else {
            if err := coro.Context().Unmarshal(b); err != nil {
                log.Fatal(err)
            }
        }

        defer func() {
            if b, err := coro.Context().Marshal(); err != nil {
                log.Fatal(err)
            } else if err := os.WriteFile(state, b, 0666); err != nil {
                log.Fatal(err)
            }
        }()
    }

    if coro.Next() {
        println("yield:", coro.Recv())
    }
}
```

When building in volatile mode (the default), the program runs a single step of
the coroutine and loses its state, each run of the application starts back at
the beginning:
```
$ go build
$ ./main
yield: 0
$ ./main
yield: 0
$ ./main
yield: 0
```
However, when building in durable mode, the program saves the coroutine state
and restores it for each run, it keeps making progress across executions:
```
$ go generate && GOROOT=$PWD/vendor/goroot go build -tags durable
$ ./main
yield: 0
$ ./main
yield: 1
$ ./main
yield: 2
```

> **Warning**
> At this time, the state of a coroutine is bound to a specific version of the
> program, attempting to resume a state on a different version is not supported.

### Scheduling

Pausing, marshaling, unmarshalling, and resuming durable coroutines is work for
a scheduler which is not included in this package. The `coroutine` project only
provides the building blocks needed to create those types of systems.

### Language Support

The `coroc` compiler currently supports a subset of Go when compiling coroutines
to durable mode.

The compiler currently does not support compiling coroutines that contain the
`go` keyword, control structures with `goto` or `fallthrough` statements, or
`for` loop post statements with function calls. Those limitations will be lifted
in the future but as of now have not proven necessary to support compiling
durable coroutines in common Go programs.

Note that none of those restrictions apply to code that is not on the call path
of coroutines.

### Performance

The code generated by `coroc` has been tested for correctness but has not been
extensively benchmarked, it is expected that the durable form of coroutines will
have compute and memory footprint than when running in volatile mode.