.PHONY: all build clean completions man-page release

BINARY_NAME=treex
VERSION?=$(shell git describe --tags --abbrev=0)

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o ./bin/$(BINARY_NAME) ./cmd/treex

clean:
	@echo "Cleaning up..."
	@rm -rf ./bin
	@rm -rf ./dist
	@rm -rf ./completions
	@rm -rf ./man

completions:
	@echo "Generating shell completions..."
	@mkdir -p ./completions
	@go run ./cmd/treex completion bash > ./completions/treex.bash
	@go run ./cmd/treex completion zsh > ./completions/_treex
	@go run ./cmd/treex completion fish > ./completions/treex.fish

man-page:
	@echo "Generating man page..."
	@mkdir -p ./man/man1
	@go run ./cmd/treex man --path ./man/man1

release: completions man-page
	@echo "Creating a release..."
	@docker run --rm -it \
		-v "$(CURDIR):/go/src/github.com/arthur-debert/treex" \
		-w /go/src/github.com/arthur-debert/treex \
		goreleaser/goreleaser release --snapshot --clean 