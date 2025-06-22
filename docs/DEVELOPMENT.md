# Development Guide

## Documentation

- **[OPTIONS.md](OPTIONS.md)** - Complete command-line options reference
- **[INFO-FILES.md](INFO-FILES.md)** - .info file format specification and examples

## Codebase Structure

```text
treex/
├── cmd/treex/           # CLI application entry point
│   ├── main.go         # Main executable
│   └── cmd/            # Cobra command definitions & flags
│       ├── root.go     # Root command setup
│       ├── show.go     # Show command (main functionality) - THIN CLI LAYER
│       ├── add_info.go # Add info command
│       ├── gen_info.go # Generate info command
│       └── info_files.go # Info files command
├── pkg/                # Public packages
│   ├── app/            # 🆕 MAIN BUSINESS LOGIC
│   │   ├── app.go      # Core RenderAnnotatedTree() function
│   │   └── app_test.go # Business logic tests
│   ├── info/           # .info file parsing
│   │   ├── parser.go   # Annotation parser logic
│   │   └── *_test.go   # Parser tests
│   ├── tree/           # File tree building
│   │   ├── builder.go  # Tree construction from filesystem
│   │   ├── ignore.go   # .gitignore-style filtering
│   │   └── *_test.go   # Tree building tests
│   └── tui/            # Terminal UI rendering
│       ├── renderer.go      # Plain text renderer
│       ├── styled_renderer.go # Styled renderer with colors
│       ├── styles.go        # Color schemes & styling
│       └── *_test.go        # Rendering tests
├── scripts/            # Development and build scripts
│   ├── build          # Build binary to bin/
│   ├── gen-completion # Generate shell completions
│   ├── gen-manpage    # Generate man pages
│   ├── test-with-cov  # Run tests with coverage
│   └── release-new    # Create new releases
└── docs/               # Documentation
```

**Key Components:**

- **App** (`pkg/app/`): 🆕 **MAIN BUSINESS LOGIC** - Central `RenderAnnotatedTree()` function that coordinates all operations
- **Parser** (`pkg/info/`): Handles `.info` file parsing, annotation extraction, and `.info` file generation  
- **Builder** (`pkg/tree/`): Creates tree structures from filesystem with annotations and filtering
- **Renderers** (`pkg/tui/`): Multiple rendering modes (plain, styled, minimal)

## Styling System

Styles are located in `pkg/tui/styles.go` with three rendering modes:

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

- Go 1.21 or higher
- Terminal with color support (for styled output)

### Building

```bash
# Build for current platform (recommended)
./scripts/build

# Or build directly
go build -o ./bin/treex ./cmd/treex
```

The binary will be created at `./bin/treex`.

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage (recommended for development)
./scripts/test-with-cov

# Run specific package tests
go test ./pkg/info
go test ./pkg/tree
go test ./pkg/tui
go test ./pkg/app

# Verbose test output
go test -v ./...
```

### Development Testing

```bash
# Build first
./scripts/build

# Test current implementation
./bin/treex .

# Test with different options (see OPTIONS.md for full reference)
./bin/treex --verbose .
./bin/treex --no-color .
./bin/treex --depth 2 .

# Or install for global use
go install ./cmd/treex
treex .
```

### Distribution

#### Local Installation

```bash
go install github.com/adebert/treex/cmd/treex@latest
```

#### Local Development Build

```bash
./scripts/build
# Binary available at ./bin/treex
```

#### Release Process

```bash
# Create a new release (interactive)
./scripts/release-new

# Or automatic patch release
./scripts/release-new --patch --yes
```

This uses GoReleaser to build for all platforms and create GitHub releases.

## Architecture Notes

### 🆕 Clean Architecture (Recent Refactoring)

The codebase follows clean architecture principles with proper separation of concerns:

**CLI Layer** (`cmd/treex/cmd/show.go`):

- **Thin interface layer** (~25 lines of code)
- Only handles: argument parsing, calling business logic, outputting result
- **No business logic** - purely CLI concerns

**Business Logic** (`pkg/app/app.go`):

- **Central `RenderAnnotatedTree()` function** - main application logic
- Coordinates all operations: parsing, building, rendering, verbose output
- **Returns structured results** (`RenderResult` with output string and stats)
- **Highly testable** - pure functions with no I/O dependencies
- **Reusable** - can be used by web APIs, other interfaces, etc.

**Support Packages**:

- `pkg/info` - Annotation parsing
- `pkg/tree` - Tree building with filtering
- `pkg/tui` - Rendering with string output support

**Benefits Achieved**:

- ✅ **Testable**: Business logic is unit tested in isolation
- ✅ **Reusable**: Core functionality can be used by different interfaces  
- ✅ **Maintainable**: Single responsibility principle enforced
- ✅ **Clean**: Proper separation between CLI, business logic, and infrastructure

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

### Main Application Flow

```go
// This is what happens when you run `treex show`:
options := app.RenderOptions{
    Verbose: true,
    NoColor: false,
    // ... other options from CLI flags
}

result, err := app.RenderAnnotatedTree(targetPath, options)
if err != nil {
    return err
}

fmt.Print(result.Output) // CLI just outputs the result
```

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
