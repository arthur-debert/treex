// see docs/dev/patterns.txt
package pattern

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Pattern represents a single pattern that can match file paths
type Pattern interface {
	// Matches returns true if the path should be excluded
	Matches(path string, isDir bool) bool
	// String returns a description of the pattern for debugging
	String() string
}

// CompositeFilter combines multiple patterns into a single filter
type CompositeFilter struct {
	patterns []Pattern
}

// NewCompositeFilter creates a new composite filter
func NewCompositeFilter(patterns ...Pattern) *CompositeFilter {
	return &CompositeFilter{patterns: patterns}
}

// ShouldExclude returns true if any pattern matches (excludes) the path
func (cf *CompositeFilter) ShouldExclude(path string, isDir bool) bool {
	for _, pattern := range cf.patterns {
		if pattern.Matches(path, isDir) {
			return true
		}
	}
	return false
}

// AddPattern adds a pattern to the filter
func (cf *CompositeFilter) AddPattern(pattern Pattern) {
	cf.patterns = append(cf.patterns, pattern)
}

// HiddenPattern matches hidden files/directories (starting with .)
type HiddenPattern struct {
	exclude bool // if true, exclude hidden files; if false, include them
}

// NewHiddenPattern creates a hidden file pattern
func NewHiddenPattern(exclude bool) *HiddenPattern {
	return &HiddenPattern{exclude: exclude}
}

// Matches returns true if the path should be excluded according to hidden file rules
func (hp *HiddenPattern) Matches(path string, isDir bool) bool {
	basename := filepath.Base(path)
	isHidden := strings.HasPrefix(basename, ".") && basename != "." && basename != ".."

	// If exclude=true, we want to exclude hidden files
	// If exclude=false, we want to include hidden files (so don't exclude anything)
	return hp.exclude && isHidden
}

// String returns a description of the pattern for debugging
func (hp *HiddenPattern) String() string {
	if hp.exclude {
		return "hidden:exclude"
	}
	return "hidden:include"
}

// FilterBuilder helps construct composite filters from options
type FilterBuilder struct {
	fs     afero.Fs
	filter *CompositeFilter
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder(fs afero.Fs) *FilterBuilder {
	return &FilterBuilder{
		fs:     fs,
		filter: NewCompositeFilter(),
	}
}

// AddUserExcludes adds user-specified exclude patterns using shell glob semantics
func (fb *FilterBuilder) AddUserExcludes(excludes []string) *FilterBuilder {
	for _, exclude := range excludes {
		fb.filter.AddPattern(NewShellPattern(exclude))
	}
	return fb
}

// AddHiddenFilter adds hidden file filtering
func (fb *FilterBuilder) AddHiddenFilter(showHidden bool) *FilterBuilder {
	// If showHidden=false, we want to exclude hidden files
	fb.filter.AddPattern(NewHiddenPattern(!showHidden))
	return fb
}

// AddGitignore adds patterns from .gitignore file using gitignore semantics
func (fb *FilterBuilder) AddGitignore(gitignorePath string, disabled bool) *FilterBuilder {
	if disabled {
		return fb
	}

	ignorePattern, err := NewIgnorefilePattern(fb.fs, gitignorePath)
	if err != nil {
		// Silently ignore missing .gitignore files
		return fb
	}

	fb.filter.AddPattern(ignorePattern)
	return fb
}

// Build returns the constructed composite filter
func (fb *FilterBuilder) Build() *CompositeFilter {
	return fb.filter
}
