GO ?= go

testdata.source = testdata/coroutine.go testdata/testdata.go
testdata.target = $(testdata.source:.go=_durable.go)

test: clean generate
	$(GO) test ./...
	$(GO) test -tags durable ./...

generate: coroc
	PATH="$$(pwd):$$PATH" $(GO) generate ./testdata
	$(GO) fmt $(testdata.target) || exit 0

coroc:
	$(GO) build -o $@ ./cmd/coroc

clean:
	$(RM) coroc $(testdata.target)

.PHONY: clean generate test
