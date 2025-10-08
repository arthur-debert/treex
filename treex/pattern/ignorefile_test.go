package pattern_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pattern"
)

func TestIgnorefilePattern(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create a .gitignore file with go-git compatible patterns
	gitignoreContent := `# This is a comment
*.log
*.tmp

# Directory patterns
node_modules/
dist/

# Negation (let go-git handle this properly)
!important.log

src/**/*.test.js
`

	fs.MustCreateTree("/project", map[string]interface{}{
		".gitignore": gitignoreContent,
	})

	ignorePattern, err := pattern.NewIgnorefilePattern(fs, "/project/.gitignore")
	if err != nil {
		t.Fatalf("Failed to create ignorefile pattern: %v", err)
	}

	tests := []struct {
		path     string
		isDir    bool
		expected bool
		desc     string
	}{
		{"debug.log", false, true, "matches *.log pattern"},
		{"temp.tmp", false, true, "matches *.tmp pattern"},
		{"src/main.go", false, false, "doesn't match non-matching file"},
		{"node_modules", true, true, "matches directory pattern"},
		{"dist", true, true, "matches directory pattern"},
		{"important.log", false, false, "negated pattern should not match"},
		{"src/component/test.test.js", false, true, "matches recursive pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := ignorePattern.Matches(tt.path, tt.isDir)
			if result != tt.expected {
				t.Errorf("Ignorefile pattern on %q (isDir=%v): expected %v, got %v",
					tt.path, tt.isDir, tt.expected, result)
			}
		})
	}
}

func TestIgnorefilePatternMissingFile(t *testing.T) {
	fs := testutil.NewTestFS()

	_, err := pattern.NewIgnorefilePattern(fs, "/nonexistent/.gitignore")
	if err == nil {
		t.Error("Expected error for missing gitignore file")
	}
}
