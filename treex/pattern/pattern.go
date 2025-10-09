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

// PluginIncludePattern implements include-only filtering for plugin results
// It excludes everything that's NOT in the allowed paths (inverse logic)
type PluginIncludePattern struct {
	allowedPaths   map[string]bool
	allowedParents map[string]bool // Cache of parent directories
}

// NewPluginIncludePattern creates a plugin include pattern
// allowedPaths contains the exact file paths that should be included
func NewPluginIncludePattern(allowedPaths map[string]bool) *PluginIncludePattern {
	// Pre-compute parent directories that should also be included
	allowedParents := make(map[string]bool)

	for path := range allowedPaths {
		// Add all parent directories to ensure they're included
		dir := filepath.Dir(path)
		for dir != "." && dir != "/" && dir != "" {
			allowedParents[dir] = true
			dir = filepath.Dir(dir)
		}
	}

	return &PluginIncludePattern{
		allowedPaths:   allowedPaths,
		allowedParents: allowedParents,
	}
}

// Matches returns true if the path should be excluded (not in allowed paths or parent dirs)
func (pip *PluginIncludePattern) Matches(path string, isDir bool) bool {
	// Normalize path separators for consistent matching
	path = filepath.ToSlash(path)

	// Include if path is explicitly allowed
	if pip.allowedPaths[path] {
		return false // Don't exclude
	}

	// Include if path is a parent directory of allowed paths
	if pip.allowedParents[path] {
		return false // Don't exclude
	}

	// Include the root directory
	if path == "." || path == "" {
		return false // Don't exclude
	}

	// Exclude everything else
	return true
}

// String returns a description of the pattern for debugging
func (pip *PluginIncludePattern) String() string {
	return "plugin:include-only"
}

// BuiltinIgnorePatterns contains patterns that are ignored by default
// These represent common directories and files that users typically don't want in tree output:
// - Version control directories (.git, .svn, .hg)
// - Package manager caches (node_modules, __pycache__)
// - OS-specific files (.DS_Store)
// - Common temporary/log files
// Note: This works alongside user --exclude patterns, .gitignore files, and hidden file filtering
var BuiltinIgnorePatterns = []string{
	".git",         // Git repository directory
	".svn",         // Subversion directory
	".hg",          // Mercurial directory
	"node_modules", // Node.js package directory
	"__pycache__",  // Python bytecode cache
	".DS_Store",    // macOS directory metadata
	"*.tmp",        // Temporary files
	"*.log",        // Log files
}

// FilterBuilder helps construct composite filters from options
// It coordinates multiple exclusion mechanisms:
// 1. Built-in ignore patterns (BuiltinIgnorePatterns) - can be disabled with --no-builtin-ignores
// 2. User exclude patterns (--exclude flag) - shell glob patterns
// 3. Gitignore files (.gitignore) - gitignore format patterns
// 4. Hidden file filtering (--hidden flag) - files starting with '.'
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

// AddBuiltinIgnores adds default ignore patterns for common VCS and build artifacts
// These patterns work alongside user excludes, gitignore, and hidden file filtering.
// Can be disabled with --no-builtin-ignores flag in CLI.
func (fb *FilterBuilder) AddBuiltinIgnores(enabled bool) *FilterBuilder {
	if !enabled {
		return fb
	}

	// Add each built-in pattern as a shell pattern for consistent behavior
	for _, pattern := range BuiltinIgnorePatterns {
		fb.filter.AddPattern(NewShellPattern(pattern))
	}
	return fb
}

// AddUserExcludes adds user-specified exclude patterns using shell glob semantics
// These patterns are specified via --exclude flags and work alongside built-in ignores,
// gitignore files, and hidden file filtering.
func (fb *FilterBuilder) AddUserExcludes(excludes []string) *FilterBuilder {
	for _, exclude := range excludes {
		fb.filter.AddPattern(NewShellPattern(exclude))
	}
	return fb
}

// AddHiddenFilter adds hidden file filtering (files starting with '.')
// This works alongside built-in ignores, user excludes, and gitignore patterns.
// Controlled by --hidden flag in CLI (default: show hidden files).
func (fb *FilterBuilder) AddHiddenFilter(showHidden bool) *FilterBuilder {
	// If showHidden=false, we want to exclude hidden files
	fb.filter.AddPattern(NewHiddenPattern(!showHidden))
	return fb
}

// AddGitignore adds patterns from .gitignore file using gitignore semantics
// This works alongside built-in ignores, user excludes, and hidden file filtering.
// Automatically looks for .gitignore files and applies their patterns.
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

// AddPluginFilter adds plugin-based inclusion filtering (inverse of typical exclusion)
// Only paths in allowedPaths (and their parent directories) will be included
// This implements "include-only" filtering by excluding everything else
func (fb *FilterBuilder) AddPluginFilter(allowedPaths map[string]bool) *FilterBuilder {
	if len(allowedPaths) == 0 {
		return fb
	}

	fb.filter.AddPattern(NewPluginIncludePattern(allowedPaths))
	return fb
}

// Build returns the constructed composite filter
// The final filter combines all exclusion mechanisms that were added:
// built-in ignores, user excludes, gitignore patterns, hidden file filtering, and plugin filters
func (fb *FilterBuilder) Build() *CompositeFilter {
	return fb.filter
}
