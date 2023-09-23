module github.com/stealthrocket/coroutine/compiler

go 1.21.0

require (
	github.com/google/go-cmp v0.5.9
	github.com/stealthrocket/coroutine v0.0.0-20230906012022-7474cda88ddc
	golang.org/x/sync v0.3.0
	golang.org/x/tools v0.13.0
)

require (
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

replace github.com/stealthrocket/coroutine => ../
