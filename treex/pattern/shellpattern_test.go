// see docs/dev/patterns.txt
package pattern_test

import (
	"testing"

	"github.com/jwaldrip/treex/treex/pattern"
)

func TestShellPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		path     string
		isDir    bool
		expected bool
		desc     string
	}{
		// Basic glob patterns
		{"*.txt", "file.txt", false, true, "simple glob matches file"},
		{"*.txt", "file.go", false, false, "simple glob doesn't match different extension"},
		{"*.txt", "src/file.txt", false, true, "simple glob matches file in subdirectory"},

		// Directory patterns
		{"node_modules", "node_modules", true, true, "exact directory name"},
		{"node_modules", "src/node_modules", true, true, "directory in subdirectory (path match)"},

		// Path patterns
		{"src/*.go", "src/main.go", false, true, "path pattern matches"},
		{"src/*.go", "main.go", false, false, "path pattern doesn't match without path"},
		{"src/**/*.go", "src/deep/nested/file.go", false, true, "recursive pattern matches"},

		// Hidden files (handled by HiddenPattern, but testing glob behavior)
		{".*", ".hidden", false, true, "dot pattern matches hidden file"},
		{".*", "visible", false, false, "dot pattern doesn't match visible file"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			shellPattern := pattern.NewShellPattern(tt.pattern)
			result := shellPattern.Matches(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("Pattern %q on path %q: expected %v, got %v",
					tt.pattern, tt.path, tt.expected, result)
			}
		})
	}
}
