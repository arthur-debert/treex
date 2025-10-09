package pattern_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pattern"
)

func TestHiddenPattern(t *testing.T) {
	tests := []struct {
		exclude  bool
		path     string
		expected bool
		desc     string
	}{
		{true, ".hidden", true, "exclude hidden file"},
		{true, "visible.txt", false, "don't exclude visible file"},
		{true, ".", false, "don't exclude current directory"},
		{true, "..", false, "don't exclude parent directory"},
		{false, ".hidden", false, "include mode doesn't exclude hidden"},
		{false, "visible.txt", false, "include mode doesn't exclude visible"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			hiddenPattern := pattern.NewHiddenPattern(tt.exclude)
			result := hiddenPattern.Matches(tt.path, false)
			if result != tt.expected {
				t.Errorf("HiddenPattern(exclude=%v) on %q: expected %v, got %v",
					tt.exclude, tt.path, tt.expected, result)
			}
		})
	}
}

func TestCompositeFilter(t *testing.T) {
	// Create a composite filter with multiple patterns
	filter := pattern.NewCompositeFilter(
		pattern.NewShellPattern("*.tmp"),        // exclude temp files
		pattern.NewShellPattern("node_modules"), // exclude node_modules
		pattern.NewHiddenPattern(true),          // exclude hidden files
	)

	tests := []struct {
		path     string
		isDir    bool
		expected bool
		desc     string
	}{
		{"file.txt", false, false, "normal file not excluded"},
		{"temp.tmp", false, true, "temp file excluded by glob"},
		{"node_modules", true, true, "node_modules excluded by glob"},
		{".hidden", false, true, "hidden file excluded by hidden pattern"},
		{"src/temp.tmp", false, true, "temp file in subdirectory excluded"},
		{"src/.hidden", false, true, "hidden file in subdirectory excluded"},
		{"src/main.go", false, false, "normal file in subdirectory not excluded"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := filter.ShouldExclude(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("CompositeFilter on %q: expected %v, got %v",
					tt.path, tt.expected, result)
			}
		})
	}
}

func TestFilterBuilder(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create a .gitignore file
	fs.MustCreateTree("/project", map[string]interface{}{
		".gitignore": "*.log\nnode_modules/\n",
	})

	// Build a filter using the FilterBuilder
	filter := pattern.NewFilterBuilder(fs).
		AddUserExcludes([]string{"*.tmp", "dist"}).
		AddHiddenFilter(false). // showHidden=false means exclude hidden
		AddGitignore("/project/.gitignore", false).
		Build()

	tests := []struct {
		path     string
		isDir    bool
		expected bool
		desc     string
	}{
		{"main.go", false, false, "normal file not excluded"},
		{"temp.tmp", false, true, "excluded by user pattern"},
		{"dist", true, true, "excluded by user pattern"},
		{"debug.log", false, true, "excluded by gitignore"},
		{"node_modules", true, true, "excluded by gitignore"},
		{".hidden", false, true, "excluded by hidden filter"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := filter.ShouldExclude(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("FilterBuilder result on %q (isDir=%v): expected %v, got %v",
					tt.path, tt.isDir, tt.expected, result)
			}
		})
	}
}

func TestFilterBuilderMissingGitignore(t *testing.T) {
	fs := testutil.NewTestFS()

	// Build filter with non-existent .gitignore (should not error)
	filter := pattern.NewFilterBuilder(fs).
		AddGitignore("/nonexistent/.gitignore", false).
		Build()

	// Should work fine with missing gitignore
	result := filter.ShouldExclude("test.txt", false)
	if result {
		t.Error("Missing gitignore should not exclude anything")
	}
}

func TestFilterBuilderDisabledGitignore(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/project", map[string]interface{}{
		".gitignore": "*.log\n",
	})

	// Build filter with disabled gitignore
	filter := pattern.NewFilterBuilder(fs).
		AddGitignore("/project/.gitignore", true). // disabled=true
		Build()

	// Should not exclude files that would be excluded by gitignore
	result := filter.ShouldExclude("debug.log", false)
	if result {
		t.Error("Disabled gitignore should not exclude anything")
	}
}
