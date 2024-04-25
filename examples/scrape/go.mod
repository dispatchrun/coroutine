module scrape

go 1.21.0
toolchain go1.22.2

require github.com/stealthrocket/coroutine v0.0.0-00000000000000-000000000000

require google.golang.org/protobuf v1.33.0 // indirect

replace github.com/stealthrocket/coroutine => ../../
