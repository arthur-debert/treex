# Development Guide

## Codebase Structure

```
treex/
├── cmd/treex/           # CLI application entry point
│   ├── main.go         # Main executable
│   └── cmd/root.go     # Cobra command definitions & flags
├── internal/           # Internal packages (not importable)
│   ├── info/           # .info file parsing
│   │   ├── parser.go   # Annotation parser logic
│   │   └── parser_test.go
│   ├── tree/           # File tree building
│   │   ├── builder.go  # Tree construction from filesystem
│   │   └── builder_test.go
│   └── tui/            # Terminal UI rendering
│       ├── renderer.go      # Plain text renderer
│       ├── styled_renderer.go # Styled renderer with colors
│       ├── styles.go        # Color schemes & styling
│       └── *_test.go        # Tests
├── pkg/                # Public packages (empty)
└── docs/               # Documentation
```

**Key Components:**

- **Parser** (`internal/info/`): Handles `.info` file parsing and annotation extraction
- **Builder** (`internal/tree/`): Creates tree structures from filesystem with annotations
- **Renderers** (`internal/tui/`): Multiple rendering modes (plain, styled, minimal)

## Styling System

Styles are located in `internal/tui/styles.go` with three rendering modes:

### Style Names (Semantic)
- **TreeLines**: Tree connectors (├── └──)
- **RootPath**: Root directory name
- **AnnotatedPath**: Paths that have annotations (regular full color)
- **UnannotatedPath**: Paths without annotations (subdued gray)
- **AnnotationText**: Annotation content (blue)
- **AnnotationContainer**: Annotation formatting/borders

### Color Scheme
- **Paths with annotations**: Regular full color (`AnnotatedPath` - light gray)
- **Paths without annotations**: Subdued gray (`UnannotatedPath`)
- **Annotations**: Blue (`AnnotationText`)
- **Tree connectors**: Subtle gray (`TreeLines`)

### Rendering Modes
1. **Full Styled** (`NewTreeStyles()`): Full color palette with Lip Gloss styling
2. **Minimal** (`NewMinimalTreeStyles()`): Limited colors for basic terminals
3. **Plain** (`NewNoColorTreeStyles()`): No colors, plain text only

### Legacy Compatibility
The old style names (`TreeConnector`, `Directory`, `File`, `AnnotationTitle`, `AnnotationDescription`) are still available for backward compatibility but are deprecated.

## Command Options

```bash
treex [path] [flags]
```

### Available Flags

- `--verbose, -v`: Show verbose output including parsed .info structure
- `--path, -p <path>`: Specify path to analyze (defaults to current directory)
- `--no-color`: Disable colored output (uses plain renderer)
- `--minimal`: Use minimal styling with fewer colors

### Examples

```bash
treex .                    # Analyze current directory with full styling
treex /path/to/project     # Analyze specific path
treex --verbose .          # Show parsing details and tree structure
treex --no-color .         # Plain text output
treex --minimal .          # Minimal colors for limited terminals
```

## Development Workflow

### Prerequisites

- Go 1.24.4 or higher
- Terminal with color support (for styled output)

### Building

```bash
# Build for current platform
go build -o treex ./cmd/treex

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o treex-linux ./cmd/treex

# Build with styling support
go build -o treex-styled ./cmd/treex
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/info
go test ./internal/tree
go test ./internal/tui

# Run with coverage
go test -cover ./...

# Verbose test output
go test -v ./...
```

### Development Testing

```bash
# Test current implementation
./treex .

# Test with different flags
./treex --verbose .
./treex --no-color .
./treex --minimal .

# Test on different directories
./treex /path/to/test/project
```

### Distribution

#### Local Installation

```bash
go install github.com/adebert/treex/cmd/treex@latest
```

#### Manual Distribution

```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o treex-linux ./cmd/treex
GOOS=darwin GOARCH=amd64 go build -o treex-darwin ./cmd/treex
GOOS=windows GOARCH=amd64 go build -o treex-windows.exe ./cmd/treex
```

## Architecture Notes

### Annotation System

- `.info` files contain path-description pairs
- Supports single-line and multi-line descriptions
- Multi-line descriptions use first line as title if followed by content
- Paths are relative to the `.info` file location

### Tree Building

- Filesystem traversal with annotation mapping
- Directories sorted before files, both alphabetically
- Hidden files/directories skipped unless annotated
- Recursive traversal for directory structures

### Rendering Pipeline

1. Parse `.info` files for annotations
2. Build tree structure from filesystem
3. Map annotations to tree nodes
4. Render with chosen styling mode

### Testing Strategy

- Unit tests for each component
- Integration tests for full pipeline
- Style testing with different color modes
- Edge case testing (empty dirs, missing files, etc.)
