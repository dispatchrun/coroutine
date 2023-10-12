GO ?= go

all: test

test:
	$(GO) test ./...
	$(GO) test --tags=durable ./...
	$(MAKE) -C compiler $@

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: lint
lint:
	which golangci-lint >/dev/null && golangci-lint run

clean:
	$(MAKE) -C compiler $@

.PHONY: all clean test
