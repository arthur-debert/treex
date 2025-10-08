// see docs/dev/patterns.txt
package pattern

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/spf13/afero"
)

// IgnorefilePattern wraps go-git's gitignore implementation
// This handles .gitignore files using proper gitignore semantics
type IgnorefilePattern struct {
	matcher gitignore.Matcher
}

// NewIgnorefilePattern loads patterns from a .gitignore file using go-git
func NewIgnorefilePattern(fs afero.Fs, gitignorePath string) (*IgnorefilePattern, error) {
	content, err := afero.ReadFile(fs, gitignorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gitignore file: %w", err)
	}

	// Parse gitignore patterns line by line
	var patterns []gitignore.Pattern
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := gitignore.ParsePattern(line, nil)
		patterns = append(patterns, pattern)
	}

	matcher := gitignore.NewMatcher(patterns)

	return &IgnorefilePattern{matcher: matcher}, nil
}

// Matches returns true if the path should be excluded according to gitignore rules
func (ip *IgnorefilePattern) Matches(path string, isDir bool) bool {
	// go-git expects the path without leading slash
	cleanPath := strings.TrimPrefix(path, "/")

	// go-git's Match method returns true if the path should be ignored
	return ip.matcher.Match(strings.Split(cleanPath, "/"), isDir)
}

// String returns a description of the pattern for debugging
func (ip *IgnorefilePattern) String() string {
	return "ignorefile"
}
