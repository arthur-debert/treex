.PHONY: build test clean run

# Build the application
build:
	go build -o bin/myapp ./cmd/myapp

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run the application in server mode
run:
	go run ./cmd/myapp server

# Run the application in parse mode with test fixture
parse:
	go run ./cmd/myapp parse ./internal/parser/testdata/fixture.json

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run with verbose output
test-verbose:
	go test -v ./... 