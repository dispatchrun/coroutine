GO ?= go

scrape: *.go coroc
	./coroc .
	$(GO) build -o $@ -tags durable .

coroc:
	$(GO) build -C ../../compiler -o $$(pwd)/coroc ./cmd/coroc

.PHONY: clean
clean:
	$(RM) scrape coroc
