package pattern

import (
	"testing"

	"github.com/spf13/afero"
)

func TestBuiltinIgnorePatterns(t *testing.T) {
	// Test that all built-in patterns are valid and work as expected
	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool // true if should be excluded
	}{
		// Version control directories
		{"git directory", ".git", true, true},
		{"git subdirectory", "project/.git", true, true},
		{"svn directory", ".svn", true, true},
		{"mercurial directory", ".hg", true, true},

		// Package manager caches
		{"node_modules directory", "node_modules", true, true},
		{"nested node_modules", "project/node_modules", true, true},
		{"python cache", "__pycache__", true, true},
		{"nested python cache", "src/__pycache__", true, true},

		// OS-specific files
		{"DS_Store file", ".DS_Store", false, true},
		{"nested DS_Store", "folder/.DS_Store", false, true},

		// Temporary/log files
		{"temp file", "test.tmp", false, true},
		{"log file", "app.log", false, true},
		{"nested temp", "build/output.tmp", false, true},

		// Files that should NOT be excluded
		{"regular file", "README.md", false, false},
		{"regular directory", "src", true, false},
		{"git-like but not git", "git-repo", true, false},
		{"contains git", "my.git.file", false, false},
		{"log-like but not log", "blog.txt", false, false},
	}

	fs := afero.NewMemMapFs()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create filter with only built-in ignores enabled
			builder := NewFilterBuilder(fs)
			builder.AddBuiltinIgnores(true)
			filter := builder.Build()

			result := filter.ShouldExclude(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("BuiltinIgnores for %q: expected %v, got %v", tt.path, tt.expected, result)
			}
		})
	}
}

func TestBuiltinIgnoresPatternsDisabled(t *testing.T) {
	// Test that built-in patterns can be disabled
	fs := afero.NewMemMapFs()

	// Create filter with built-in ignores disabled
	builder := NewFilterBuilder(fs)
	builder.AddBuiltinIgnores(false)
	filter := builder.Build()

	// Test that normally-ignored paths are not excluded when disabled
	testPaths := []struct {
		path  string
		isDir bool
	}{
		{".git", true},
		{"node_modules", true},
		{"__pycache__", true},
		{".DS_Store", false},
		{"test.tmp", false},
		{"app.log", false},
	}

	for _, tp := range testPaths {
		if filter.ShouldExclude(tp.path, tp.isDir) {
			t.Errorf("BuiltinIgnores disabled: %q should not be excluded", tp.path)
		}
	}
}

func TestBuiltinIgnoresWithOtherFilters(t *testing.T) {
	// Test that built-in ignores work alongside other filtering mechanisms
	fs := afero.NewMemMapFs()

	// Create filter with built-ins and user excludes
	builder := NewFilterBuilder(fs)
	builder.AddBuiltinIgnores(true)
	builder.AddUserExcludes([]string{"*.test"})
	builder.AddHiddenFilter(false) // exclude hidden files
	filter := builder.Build()

	tests := []struct {
		name     string
		path     string
		isDir    bool
		expected bool
		reason   string
	}{
		{"git dir excluded by builtin", ".git", true, true, "builtin ignore"},
		{"test file excluded by user", "main.test", false, true, "user exclude"},
		{"hidden file excluded by hidden filter", ".hidden", false, true, "hidden filter"},
		{"regular file not excluded", "main.go", false, false, "no exclusion"},
		{"visible git-like not excluded", "git-tools", true, false, "not exact match"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.ShouldExclude(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("Combined filters for %q: expected %v (reason: %s), got %v",
					tt.path, tt.expected, tt.reason, result)
			}
		})
	}
}

func TestBuiltinIgnorePatternsContent(t *testing.T) {
	// Test that the BuiltinIgnorePatterns slice contains expected patterns
	expectedPatterns := map[string]bool{
		".git":         true,
		".svn":         true,
		".hg":          true,
		"node_modules": true,
		"__pycache__":  true,
		".DS_Store":    true,
		"*.tmp":        true,
		"*.log":        true,
	}

	// Check that all expected patterns are present
	for _, pattern := range BuiltinIgnorePatterns {
		if !expectedPatterns[pattern] {
			t.Errorf("Unexpected pattern in BuiltinIgnorePatterns: %q", pattern)
		}
		delete(expectedPatterns, pattern)
	}

	// Check that no expected patterns are missing
	for missing := range expectedPatterns {
		t.Errorf("Missing expected pattern in BuiltinIgnorePatterns: %q", missing)
	}

	// Ensure we have a reasonable number of patterns (not empty, not too many)
	if len(BuiltinIgnorePatterns) < 5 {
		t.Errorf("BuiltinIgnorePatterns seems too short: %d patterns", len(BuiltinIgnorePatterns))
	}
	if len(BuiltinIgnorePatterns) > 20 {
		t.Errorf("BuiltinIgnorePatterns seems too long: %d patterns", len(BuiltinIgnorePatterns))
	}
}
