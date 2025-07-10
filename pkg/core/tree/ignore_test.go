package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
)

func TestIgnoreMatcher_BasicPatterns(t *testing.T) {
	// Create a temporary .gitignore file
	tempDir := t.TempDir()
	ignoreFile := filepath.Join(tempDir, ".gitignore")

	ignoreContent := `# Test ignore file
*.log
*.tmp
build/
node_modules/
!important.log
`

	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ignore file: %v", err)
	}

	// Create ignore matcher
	matcher, err := NewIgnoreMatcher(ignoreFile)
	if err != nil {
		t.Fatalf("Failed to create ignore matcher: %v", err)
	}

	// Test cases
	testCases := []struct {
		path     string
		isDir    bool
		expected bool
		desc     string
	}{
		{"test.log", false, true, "*.log pattern should match"},
		{"debug.log", false, true, "*.log pattern should match"},
		{"important.log", false, false, "!important.log should override *.log"},
		{"test.txt", false, false, "unmatched file should not be ignored"},
		{"build", true, true, "build/ directory should be ignored"},
		{"build/file.txt", false, true, "files in build/ should be ignored"},
		{"node_modules", true, true, "node_modules/ should be ignored"},
		{"src/test.log", false, true, "*.log should match in subdirectories"},
		{"temp.tmp", false, true, "*.tmp pattern should match"},
	}

	for _, tc := range testCases {
		result := matcher.ShouldIgnore(tc.path, tc.isDir)
		if result != tc.expected {
			t.Errorf("%s: path=%s, isDir=%v, expected=%v, got=%v",
				tc.desc, tc.path, tc.isDir, tc.expected, result)
		}
	}
}

func TestIgnoreMatcher_NonExistentFile(t *testing.T) {
	// Test with non-existent ignore file
	matcher, err := NewIgnoreMatcher("/non/existent/file")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}

	// Should not ignore anything
	if matcher.ShouldIgnore("any/file.txt", false) {
		t.Error("Empty matcher should not ignore any files")
	}
}

func TestIgnoreMatcher_ComplexPatterns(t *testing.T) {
	tempDir := t.TempDir()
	ignoreFile := filepath.Join(tempDir, ".gitignore")

	ignoreContent := `# Complex patterns
**/*.log
src/**/temp/
!src/important/
/root-only.txt
*.{tmp,bak}
`

	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ignore file: %v", err)
	}

	matcher, err := NewIgnoreMatcher(ignoreFile)
	if err != nil {
		t.Fatalf("Failed to create ignore matcher: %v", err)
	}

	testCases := []struct {
		path     string
		isDir    bool
		expected bool
		desc     string
	}{
		{"deep/nested/file.log", false, true, "**/*.log should match deeply nested logs"},
		{"src/module/temp", true, true, "src/**/temp/ should match nested temp dirs"},
		{"src/important", true, false, "!src/important/ should override patterns"},
		{"root-only.txt", false, true, "/root-only.txt should match at root"},
		{"sub/root-only.txt", false, false, "/root-only.txt should not match in subdirs"},
	}

	for _, tc := range testCases {
		result := matcher.ShouldIgnore(tc.path, tc.isDir)
		if result != tc.expected {
			t.Errorf("%s: path=%s, isDir=%v, expected=%v, got=%v",
				tc.desc, tc.path, tc.isDir, tc.expected, result)
		}
	}
}

func TestIgnoreMatcher_WithTreeBuilder(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()

	// Create test files and directories
	testFiles := []string{
		"README.md",
		"main.go",
		"debug.log",
		"important.log",
		"build/output.bin",
		"src/app.go",
		"src/test.log",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .gitignore
	ignoreFile := filepath.Join(tempDir, ".gitignore")
	ignoreContent := `*.log
build/
!important.log
`

	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Build tree with ignore support
	root, err := BuildTreeNestedWithIgnore(tempDir, ignoreFile)
	if err != nil {
		t.Fatalf("Failed to build tree with ignore: %v", err)
	}

	// Verify that ignored files are filtered out
	foundFiles := make(map[string]bool)
	err = WalkTree(root, func(node *types.Node, depth int) error {
		if !node.IsDir {
			foundFiles[node.RelativePath] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	// Should have these files
	expectedPresent := []string{
		"README.md",
		"main.go",
		"important.log", // not ignored due to !important.log
		"src/app.go",
	}

	// Should NOT have these files
	expectedAbsent := []string{
		"debug.log",        // ignored by *.log
		"build/output.bin", // ignored by build/
		"src/test.log",     // ignored by *.log
	}

	for _, file := range expectedPresent {
		if !foundFiles[file] {
			t.Errorf("Expected file %s to be present but it was filtered out", file)
		}
	}

	for _, file := range expectedAbsent {
		if foundFiles[file] {
			t.Errorf("Expected file %s to be filtered out but it was present", file)
		}
	}
}
