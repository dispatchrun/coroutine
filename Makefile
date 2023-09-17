all: test

test:
	go test ./...
	go test --tags=durable ./...
	$(MAKE) -C compiler $@

clean:
	$(MAKE) -C compiler $@

.PHONY: all clean test
