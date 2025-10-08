// see docs/dev/patterns.txt
package pattern

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ShellPattern implements glob pattern matching using doublestar
// This handles user-specified exclude patterns using shell glob semantics
type ShellPattern struct {
	pattern string
}

// NewShellPattern creates a new shell glob pattern
func NewShellPattern(pattern string) *ShellPattern {
	return &ShellPattern{pattern: pattern}
}

// Matches returns true if the path should be excluded according to shell glob rules
func (sp *ShellPattern) Matches(path string, isDir bool) bool {
	// For user patterns, use doublestar's natural behavior
	// Try full path match first
	if matched, err := doublestar.PathMatch(sp.pattern, path); err == nil && matched {
		return true
	}

	// Then try basename match for simple patterns (no path separators)
	if !strings.Contains(sp.pattern, "/") {
		basename := filepath.Base(path)
		if matched, err := doublestar.Match(sp.pattern, basename); err == nil && matched {
			return true
		}
	}

	return false
}

// String returns a description of the pattern for debugging
func (sp *ShellPattern) String() string {
	return fmt.Sprintf("shell-glob:%s", sp.pattern)
}
