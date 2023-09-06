GO ?= go

test: generate
	$(GO) test ./
	$(GO) test -tags durable ./

generate: coroc
	PATH="$$(pwd):$$PATH" $(GO) generate ./testdata

coroc:
	$(GO) build -o $@ .

clean:
	$(RM) coroc

.PHONY: clean generate
