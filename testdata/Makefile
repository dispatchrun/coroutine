GO ?= go

generate: coroc
	PATH="$(PWD):$$PATH" $(GO) generate ./

coroc:
	$(GO) build -o $@ ../cmd/coroc
