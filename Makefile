all: test

generate:
	go generate
	go generate --tags=durable

test:
	go test ./...
	go test --tags=durable ./...
