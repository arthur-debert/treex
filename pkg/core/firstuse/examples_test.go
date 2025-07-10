package firstuse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetCommonPaths(t *testing.T) {
	paths := GetCommonPaths()

	// Check we have a reasonable number of paths
	if len(paths) < 10 {
		t.Errorf("Expected at least 10 common paths, got %d", len(paths))
	}

	// Check some essential paths are included
	essentialPaths := map[string]bool{
		"src/":         false,
		"README.md":    false,
		"package.json": false,
		"tests/":       false,
		"docs/":        false,
	}

	for _, path := range paths {
		if _, exists := essentialPaths[path.Path]; exists {
			essentialPaths[path.Path] = true
		}
	}

	for path, found := range essentialPaths {
		if !found {
			t.Errorf("Essential path %s not found in common paths", path)
		}
	}

	// Check all paths have annotations
	for _, path := range paths {
		if path.Annotation == "" {
			t.Errorf("Path %s has empty annotation", path.Path)
		}
		if path.Priority <= 0 {
			t.Errorf("Path %s has invalid priority %d", path.Path, path.Priority)
		}
	}
}

func TestFindExamplesInPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create some common paths
	testPaths := []string{
		"src",
		"tests",
		"docs",
		"build",
	}

	for _, path := range testPaths {
		err := os.MkdirAll(filepath.Join(tempDir, path), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", path, err)
		}
	}

	// Create some common files
	testFiles := []string{
		"README.md",
		"package.json",
		"Makefile",
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tempDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		_ = f.Close()
	}

	// Test finding examples
	examples := FindExamplesInPath(tempDir, 5)

	// Should find at least some examples
	if len(examples) == 0 {
		t.Error("Expected to find some examples, got none")
	}

	// Should not exceed max examples
	if len(examples) > 5 {
		t.Errorf("Expected at most 5 examples, got %d", len(examples))
	}

	// Verify found examples exist in our test setup
	foundPaths := make(map[string]bool)
	for _, ex := range examples {
		foundPaths[ex.Path] = true
	}

	// Check that high-priority items are found
	highPriorityPaths := []string{"src/", "README.md", "tests/"}
	foundCount := 0
	for _, path := range highPriorityPaths {
		if foundPaths[path] {
			foundCount++
		}
	}

	if foundCount < 2 {
		t.Errorf("Expected to find at least 2 high-priority paths, found %d", foundCount)
	}
}

func TestGetFallbackExamples(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "treex-test-fallback-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create some non-standard directories and files
	testItems := []struct {
		path  string
		isDir bool
	}{
		{"my-custom-dir", true},
		{"another-folder", true},
		{"custom-file.txt", false},
		{"data.csv", false},
		{".hidden-file", false}, // Should be excluded
	}

	for _, item := range testItems {
		fullPath := filepath.Join(tempDir, item.path)
		if item.isDir {
			err := os.MkdirAll(fullPath, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", item.path, err)
			}
		} else {
			f, err := os.Create(fullPath)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", item.path, err)
			}
			_ = f.Close()
		}
	}

	// Get fallback examples
	examples := GetFallbackExamples(tempDir, 3)

	// Should find some examples
	if len(examples) == 0 {
		t.Error("Expected to find some fallback examples, got none")
	}

	// Should not exceed max
	if len(examples) > 3 {
		t.Errorf("Expected at most 3 examples, got %d", len(examples))
	}

	// Should not include hidden files (except .gitignore, .env)
	for _, ex := range examples {
		if ex.Path == ".hidden-file" {
			t.Error("Fallback examples should not include regular hidden files")
		}
	}

	// All examples should have annotations
	for _, ex := range examples {
		if ex.Annotation == "" {
			t.Errorf("Example %s has empty annotation", ex.Path)
		}
	}
}

func TestGenerateGenericAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		isDir    bool
		expected string
	}{
		{"test-folder", true, "Test files and utilities"},
		{"docs", true, "Documentation files"},
		{"src", true, "Source code files"},
		{"lib", true, "Library code"},
		{"assets", true, "Static assets"},
		{"config", true, "Configuration files"},
		{"random-dir", true, "random-dir directory"},
		{"README.md", false, "Markdown documentation"},
		{"config.json", false, "JSON configuration file"},
		{"script.sh", false, "Shell script"},
		{"style.css", false, "Stylesheet file"},
		{"index.html", false, "HTML template"},
		{"main.go", false, "Go source file"},
		{"app.js", false, "JavaScript/TypeScript source file"},
		{"test.py", false, "Python source file"},
		{"unknown.xyz", false, "xyz file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateGenericAnnotation(tt.name, tt.isDir)
			if result != tt.expected {
				t.Errorf("generateGenericAnnotation(%s, %v) = %s, want %s",
					tt.name, tt.isDir, result, tt.expected)
			}
		})
	}
}
