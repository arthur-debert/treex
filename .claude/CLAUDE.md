# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**treex** is a CLI file viewer that displays directory trees with annotations from `.info` files. It's written in Go and helps developers understand project structure by showing file descriptions alongside the tree view.

## Common Development Commands

### Building

```bash
# Build for current platform (recommended)
./scripts/build

# Or build directly to bin/
go build -o ./bin/treex ./cmd/treex
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage (recommended for development)
./scripts/test-with-cov

# Run specific package tests
go test ./pkg/info
go test ./pkg/tree
go test ./pkg/tui
go test ./pkg/app

# CI mode with race detection and coverage
./scripts/test --ci
```

### Linting

```bash
# Run golangci-lint
./scripts/lint
```

### Development Testing

```bash
# Build and test current implementation
./scripts/build
./bin/treex .

# Test with different options
./bin/treex --verbose .
./bin/treex --no-color .
./bin/treex --depth 2 .
```

## Architecture

### Clean Architecture Pattern

The codebase follows clean architecture with proper separation of concerns:

- **CLI Layer** (`cmd/treex/cmd/`): Thin interface layer handling argument parsing and calling business logic
- **Business Logic** (`pkg/app/`): Core `RenderAnnotatedTree()` function coordinating all operations
- **Support Packages**:
  - `pkg/info/` - .info file parsing and annotation handling
  - `pkg/tree/` - Tree building with filesystem filtering
  - `pkg/tui/` - Terminal rendering with multiple style modes
  - `pkg/format/` - Output format management

### Main CLI Commands

1. **Default/Show Command** (`treex [path]`): Main functionality displaying annotated tree
2. **init** (`treex init`): Initialize .info files for directory structure
3. **add** (`treex add <path> <description>`): Add descriptions to .info files
4. **import** (`treex import <file>`): Generate .info files from annotated tree structure
5. **check** (`treex check`): Validate .info files
6. **make-tree** (`treex make-tree`): Create directory structure from specification
7. **info-files** (`treex info-files`): Quick reference guide for .info format

### Key Components

**Main Business Logic** (`pkg/app/app.go`):

- Central `RenderAnnotatedTree()` function coordinates all operations
- Returns structured `RenderResult` with output string and stats
- Highly testable with no I/O dependencies

**Annotation System** (`pkg/info/`):

- Supports nested .info files in any directory
- Path resolution limited to current directory for security
- Automatic merging of all .info files in directory tree
- Handles both single-line and multi-line descriptions

**Tree Building** (`pkg/tree/`):

- Recursive filesystem discovery with .gitignore-style filtering
- Intelligent filtering with depth limits and max files protection
- Annotation priority (annotated files always shown)
- Smart ordering: annotated files first, directories before files, alphabetical

**Rendering System** (`pkg/tui/`):

- Multiple rendering modes: full color, minimal, plain text
- Tabstop-based annotation alignment for clean output
- Styling system with semantic color names
- Professional layout optimized for terminal display

### .info File Format

Simple text format for file/directory descriptions:

```text
filename
Description of the file

directory/
Description of the directory

complex-file.go
Title for the file
Multi-line description starts here and can
continue on multiple lines.
```

### Testing Strategy

The project has comprehensive test coverage:

- Unit tests for each component (`*_test.go` files)
- Integration tests for full pipeline
- Style testing with different color modes
- Edge case testing (empty dirs, missing files, etc.)
- Test coverage threshold checking in CI

### Output Formats

treex supports multiple output formats:

- `--format=color` - Full color terminal output (default)
- `--format=minimal` - Minimal color styling for basic terminals
- `--format=no-color` - Plain text output without colors

Legacy flags (`--no-color`, `--minimal`) are deprecated but still supported.

### Security Features

- Directory traversal protection in .info file paths
- Paths with `..` automatically filtered out
- .info files can only describe paths within their own directory
- Safe ignore file pattern processing

## Development Tips

- Always run `./scripts/build` before testing changes
- Use `./scripts/test-with-cov` for development to ensure good test coverage
- The main business logic is in `pkg/app/app.go` - start there for understanding flow
- CLI commands are thin wrappers that delegate to business logic
- When adding new features, follow the clean architecture pattern
- Test with different output formats to ensure compatibility
- Use the example/ directory for testing complex scenarios
