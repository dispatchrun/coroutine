GO ?= go

TARGET = testdata/coroutine_durable.go

test: generate
	$(GO) test ./
	$(GO) test -tags durable ./

generate: coroc
	PATH="$$(pwd):$$PATH" $(GO) generate ./testdata
	$(GO) fmt $(TARGET)

coroc:
	$(GO) build -o $@ .

clean:
	$(RM) coroc $(TARGET)

.PHONY: clean generate
