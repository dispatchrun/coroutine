GO ?= go

TARGET = testdata/coroutine_durable.go

test: clean generate
	$(GO) test ./...
	$(GO) test -tags durable ./...

generate: coroc
	PATH="$$(pwd):$$PATH" $(GO) generate ./testdata
	$(GO) fmt $(TARGET) || exit 0

coroc:
	$(GO) build -o $@ .

clean:
	$(RM) coroc $(TARGET)

.PHONY: clean generate test
