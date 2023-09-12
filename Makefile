all: test

generate:
	go generate
	go generate --tags=durable
	$(MAKE) -C coroc $@

test:
	go test ./...
	go test --tags=durable ./...
	$(MAKE) -C coroc $@

clean:
	$(MAKE) -C coroc $@
