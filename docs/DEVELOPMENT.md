# Development Guide

## Documentation

- **[OPTIONS.md](OPTIONS.md)** - Complete command-line options reference
- **[INFO-FILES.md](INFO-FILES.md)** - .info file format specification and examples

## Codebase Structure

```text
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

# Test with different options (see OPTIONS.md for full reference)
./treex --verbose .
./treex --no-color .
./treex --depth 2 .
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

- **Nested .info files**: Any directory can contain a `.info` file describing its contents
- **Path resolution**: Each `.info` file can only describe paths within its own directory (no parent directory access)
- **Automatic merging**: All `.info` files in the directory tree are automatically discovered and merged
- **Security**: Paths with `..` are automatically filtered out to prevent directory traversal
- **Single-line and multi-line descriptions**: Supports both formats
- **Multi-line descriptions**: Use first line as title if followed by content
- **Path format**: Paths are relative to the `.info` file location

### Nested .info File Examples

```text
# Root .info file
README.md
Main project documentation

# internal/.info file  
parser.go
Handles parsing logic

utils.go
Utility functions

# internal/deep/.info file
config.json
Deep configuration settings
```

### Tree Building

- **Recursive discovery**: Walks entire directory tree looking for `.info` files
- **Path normalization**: Converts relative paths to tree-relative paths
- **Annotation merging**: Later files override earlier ones on conflicts
- **Intelligent filtering**: Supports .gitignore-style patterns, depth limits, and max files protection
- **File ordering**: Annotated files first, then directories before files, both alphabetically
- **Hidden files/directories**: Skipped unless annotated
- **Recursive traversal**: Full directory structure processing with filtering applied

### Parsing Modes

1. **Single .info file**: `info.ParseDirectory()` / `tree.BuildTree()` - Root directory only
2. **Nested .info files**: `info.ParseDirectoryTree()` / `tree.BuildTreeNested()` - All subdirectories (default)

### Intelligent Filtering System

- **Ignore file support**: .gitignore-style pattern matching with wildcards, negation, and directory patterns
- **Depth limiting**: Configurable maximum traversal depth (default: 10 levels)
- **Max files protection**: Automatic limiting of unannotated files per directory (10 max) with overflow indicators
- **Annotation priority**: Annotated files always shown regardless of filtering rules
- **Security**: Directory traversal protection in ignore patterns

### Rendering Pipeline

1. **Discovery phase**: Recursively find all `.info` files in directory tree
2. **Parsing phase**: Parse each `.info` file with proper path context  
3. **Merging phase**: Combine all annotations with path resolution
4. **Filtering phase**: Apply ignore patterns, depth limits, and max files protection
5. **Tabstop calculation**: Calculate optimal alignment position based on longest rendered path
6. **Tree building**: Build tree structure from filesystem with merged annotations and filtering
7. **Rendering**: Apply chosen styling mode to annotated tree with tabstop alignment

### Annotation Alignment

- **Tabstop-based**: Annotations are left-aligned at a consistent tabstop position
- **Dynamic calculation**: Tabstop = max(longest_rendered_path_length, 40)
- **Professional layout**: Creates clean, readable output optimized for 890-column displays
- **Multi-line support**: Continuation lines align with the tabstop position

### Testing Strategy

- Unit tests for each component
- Integration tests for full pipeline
- Style testing with different color modes
- Edge case testing (empty dirs, missing files, etc.)
