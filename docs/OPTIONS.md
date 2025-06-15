# Command-Line Options

## Synopsis

```bash
treex [path] [flags]
```

## Arguments

### Path

- **Description**: Directory path to analyze
- **Default**: Current directory (`.`)
- **Example**: `treex /path/to/project`

## Flags

### General Options

#### `--verbose, -v`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Show verbose output including parsed .info file structure and tree building details
- **Example**: `treex --verbose .`

#### `--path, -p <path>`

- **Type**: String
- **Default**: Current directory
- **Description**: Specify the directory path to analyze (alternative to positional argument)
- **Example**: `treex --path /home/user/project`

### Display Options

#### `--no-color`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Disable colored output and use plain text renderer
- **Use Case**: For terminals that don't support colors or when redirecting output to files
- **Example**: `treex --no-color . > output.txt`

#### `--minimal`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Use minimal styling with fewer colors for basic terminals
- **Use Case**: For terminals with limited color support
- **Example**: `treex --minimal .`

### Filtering Options

#### `--use-ignore-file <file>`

- **Type**: String
- **Default**: `.gitignore`
- **Description**: Use specified ignore file to filter out files and directories
- **Format**: Supports .gitignore-style patterns including:
  - Wildcards (`*.log`, `*.tmp`)
  - Directory patterns (`build/`, `node_modules/`)
  - Negation patterns (`!important.log`)
  - Deep matching (`**/*.log`)
  - Root-relative patterns (`/root-only.txt`)
- **Special Behavior**:
  - Annotated files are always shown even if they match ignore patterns
  - Empty string disables ignore filtering
  - Missing ignore file is silently ignored (no filtering applied)
- **Examples**:
  - `treex --use-ignore-file .dockerignore .`
  - `treex --use-ignore-file "" .` (disable filtering)

#### `--depth, -d <depth>`

- **Type**: Integer
- **Default**: `10`
- **Description**: Maximum depth to traverse in the directory tree
- **Depth Calculation**: Root directory is depth 0, immediate children are depth 1, etc.
- **Use Case**: Limit output for very deep directory structures
- **Examples**:
  - `treex -d 3 .` (show only 3 levels deep)
  - `treex --depth=1 .` (show only immediate children)

### Max Files Protection

- **Constant**: `MAX_FILES_PER_DIR = 10` (not configurable via flag)
- **Behavior**:
  - Always shows all annotated files (no limit)
  - Limits unannotated files to 10 per directory
  - Shows "... X more files not shown" when limit exceeded
- **Purpose**: Improve UI usability for directories with many files

## Usage Examples

### Basic Usage

```bash
# Analyze current directory with default settings
treex

# Analyze specific directory
treex /path/to/project

# Verbose output showing parsing details
treex --verbose .
```

### Display Modes

```bash
# Plain text output (no colors)
treex --no-color .

# Minimal colors for basic terminals
treex --minimal .

# Save output to file
treex --no-color . > project-structure.txt
```

### Filtering Options

```bash
# Use custom ignore file
treex --use-ignore-file .dockerignore .

# Disable ignore filtering
treex --use-ignore-file="" .

# Limit depth to 2 levels
treex --depth 2 .
treex -d 2 .

# Combine filtering options
treex --use-ignore-file .gitignore --depth 3 --minimal .
```

### Advanced Examples

```bash
# Deep analysis with verbose output
treex --verbose --depth 5 /large/codebase

# Clean output for documentation
treex --no-color --depth 3 . > project-docs.txt

# Quick overview (shallow depth)
treex -d 1 .

# Ignore build artifacts but show structure
treex --use-ignore-file .gitignore --depth 4 .
```

## Flag Combinations

All flags can be combined freely:

```bash
# Comprehensive analysis
treex --verbose --minimal --depth 4 --use-ignore-file .gitignore /project

# Documentation generation
treex --no-color --depth 2 --use-ignore-file="" . > structure.md

# Quick filtered view
treex -d 2 --minimal .
```

## Output Behavior

### Color Schemes

- **Full styling** (default): Complete color palette with tree connectors
- **Minimal styling** (`--minimal`): Limited colors for compatibility
- **No styling** (`--no-color`): Plain text only

### Filtering Priority

1. **Annotated files**: Always shown regardless of ignore patterns or file limits
2. **Depth limits**: Applied to all files and directories
3. **Ignore patterns**: Applied to unannotated files only
4. **Max files**: Applied to unannotated files per directory (10 max)

### File Ordering

- **Annotated files**: Shown first (highlighted importance)
- **Directories**: Shown before files within each group
- **Alphabetical**: Within each category (annotated/unannotated, dir/file)

## Technical Notes

### Ignore File Format

Supports standard .gitignore syntax:

- `#` for comments
- `*` matches any characters except `/`
- `**` matches any characters including `/`
- `?` matches single character except `/`
- `!` negates a pattern
- Trailing `/` indicates directory-only patterns
- Leading `/` makes patterns root-relative

### Performance Considerations

- **Large directories**: Use `--depth` to limit traversal
- **Many files**: Max files protection automatically applies
- **Complex ignore patterns**: May slightly impact performance

### Exit Codes

- `0`: Success
- `1`: Error (invalid path, permission denied, etc.)

## See Also

- [INFO-FILES.md](INFO-FILES.md) - .info file format and examples
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development setup and architecture
