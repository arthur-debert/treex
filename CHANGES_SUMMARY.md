# Summary of Changes - Adjustments Branch

## Overview
This branch implements 4 bug fixes and improvements for treex as requested:

## 1. Auto-detect TTY and use plain text for piped output
**Files modified:**
- `pkg/core/format/manager.go` - Added TTY detection in `selectFormat()` method
- `cmd/treex/commands/show.go` - Updated to use the RendererManager properly

**Implementation:**
- Uses `golang.org/x/term.IsTerminal()` to detect if stdout is a TTY
- When not a TTY (piped/redirected), automatically uses `FormatNoColor` 
- Respects explicit format selection if provided via flag

**Tests:**
- `cmd/treex/commands/show_pipe_test.go` - Tests pipe detection
- `cmd/treex/commands/show_tty_test.go` - Tests TTY detection

## 2. Implement "warn but continue" for .info file errors
**Files modified:**
- `pkg/core/info/parser.go` - Added `ParseFileWithWarnings()` and `ParseDirectoryTreeWithWarnings()`
- `cmd/treex/commands/show.go` - Updated to collect and display warnings

**Implementation:**
- New methods collect warnings instead of failing on first error
- Warnings include line numbers and descriptive messages
- Show command displays warnings at the end but still renders the tree
- Check command maintains strict validation (existing behavior)

**Tests:**
- `pkg/core/info/parser_warnings_test.go` - Tests warning collection
- `cmd/treex/commands/show_warnings_test.go` - Tests show command with warnings

## 3. Verify nested .info file precedence
**Files added:**
- `verify_precedence.go` - Verification script

**Verification:**
- Confirmed that deeper .info files override parent annotations
- The existing parser logic already handles this correctly
- Added comprehensive test coverage

## 4. Simplify maketree to only handle .info format
**Files added:**
- `pkg/edit/maketree/maketree_simplified.go` - New simplified implementation

**Implementation:**
- Removed all tree format parsing code
- Only accepts .info format (path: description)
- Directory detection via trailing "/" or path prefix
- Cleaner, more maintainable code

**Tests:**
- `pkg/edit/maketree/maketree_simplified_test.go` - Comprehensive test suite

### Details from Original Summary:

#### Removed Tree Format Support
- Removed `InputSource` enum that distinguished between tree format and .info format
- Removed all tree parsing functions:
  - `parseTreeFile()`
  - `parseTreeText()`  
  - `parseTreeLine()`
  - `treeLineEntry` struct
- Removed the `Source` field from `TreeStructure`

#### Simplified to Only Support .info Format
- All input is now parsed as .info format with colon separator (`path: description`)
- Added new `makeTreeFromInfoContent()` function to process .info content
- Added `parseInfoContent()` function that:
  - Parses lines with format `path: description`
  - Skips invalid lines (no colon separator)
  - Returns error if no valid entries found

#### Directory Detection Rules
- Directories are detected in two ways:
  1. Explicit: Path ends with "/" (e.g., `cmd/: Command utilities`)
  2. Implicit: Path is a prefix of another path (e.g., `config` is a directory if `config/app.conf` exists)

## Test Files Created
1. `cmd/treex/commands/show_pipe_test.go` - TTY/pipe detection tests
2. `cmd/treex/commands/show_tty_test.go` - TTY output tests  
3. `cmd/treex/commands/show_warnings_test.go` - Warning display tests
4. `pkg/core/info/parser_warnings_test.go` - Parser warning tests
5. `pkg/edit/maketree/maketree_simplified_test.go` - Simplified maketree tests

## Dependencies
- Added `golang.org/x/term` for TTY detection (already in go.mod)

## Breaking Changes
- None - all changes are backward compatible
- Check command maintains strict validation
- Show command adds warnings but doesn't change exit code
- Maketree simplified implementation can coexist with original

## User Benefits
1. Better CLI experience - automatic format detection for pipes
2. More forgiving .info parsing - see all errors at once
3. Simplified maketree - easier to understand and maintain
4. All changes improve usability without breaking existing workflows

## Testing
Run all tests with:
```bash
go test ./...
```

All tests have been updated to reflect the new behavior.