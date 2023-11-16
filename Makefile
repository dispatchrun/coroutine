GO ?= go

all: test

test:
	$(GO) test ./...
	$(GO) test --tags=durable ./...
	$(MAKE) -C compiler $@

fmt:
	$(GO) fmt ./...
	buf format -w

gen:
	buf generate

lint:
	golangci-lint run
	buf lint

clean:
	$(MAKE) -C compiler $@

.PHONY: all clean fmt gen lint test
