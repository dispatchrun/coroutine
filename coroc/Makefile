GO ?= go

TARGET = testdata/coroutine_durable.go

test: generate
	$(GO) test -v -count 1
	$(GO) test -tags durable -count 1 -v

generate: coroc
	PATH="$$(pwd):$$PATH" $(GO) generate ./testdata
	$(GO) fmt $(TARGET) || exit 0

coroc:
	$(GO) build -o $@ .

clean:
	$(RM) coroc $(TARGET)

.PHONY: clean generate
