# treex

A file tree explorer for annotated files

## Installation

```bash
go install github.com/adebert/treex/cmd/treex@latest
```

## Usage

```bash
treex --help
```

## Development

### Prerequisites

- Go 1.21 or higher

### Building from source

```bash
# Clone the repository
git clone https://github.com/adebert/treex.git
cd treex

# Build the binary
go build -o treex ./cmd/treex

# Run the binary
./treex
```

### Running tests

```bash
go test ./...
```

## License

[MIT](LICENSE)

