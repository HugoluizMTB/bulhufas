.PHONY: dev test build lint clean

dev:
	go run ./cmd/server

test:
	go test ./... -race -cover

build:
	go build -o bin/bulhufas ./cmd/server

lint:
	golangci-lint run

clean:
	rm -rf bin/
