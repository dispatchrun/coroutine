all: test

test:
	go test ./...
	go test --tags=durable ./...
	$(MAKE) -C coroc $@

clean:
	$(MAKE) -C coroc $@

.PHONY: all clean test
