GO ?= go

generate: coroc
	PATH="$(PWD):$$PATH" $(GO) generate ./

coroc:
	$(GO) build -o $@ ../cmd/coroc

test: generate
	$(GO) test ./
	$(GO) test -tags durable ./

clean:
	$(RM) coroc

.PHONY: clean generate
