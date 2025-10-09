package dummy_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/plugins/dummy"
)

func TestDummyPluginName(t *testing.T) {
	plugin := dummy.NewDummyPlugin()
	if plugin.Name() != "dummy" {
		t.Errorf("Expected plugin name 'dummy', got %q", plugin.Name())
	}
}

func TestDummyPluginEmptyFilesystem(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	roots, err := plugin.FindRoots(fs, "/empty")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}
	if len(roots) != 0 {
		t.Errorf("Expected no roots in empty filesystem, got %d", len(roots))
	}
}

func TestDummyPluginSingleRoot(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	fs.MustCreateTree("/project", map[string]interface{}{
		".dummy":   "marker file",
		"main.go":  "package main",
		"test.txt": "test content",
	})

	roots, err := plugin.FindRoots(fs, "/project")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}

	if len(roots) != 1 {
		t.Fatalf("Expected 1 root, got %d", len(roots))
	}

	if roots[0] != "." {
		t.Errorf("Expected root '.', got %q", roots[0])
	}
}

func TestDummyPluginNestedRoots(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	fs.MustCreateTree("/workspace", map[string]interface{}{
		"project-a": map[string]interface{}{
			".dummy":  "marker",
			"main.go": "package main",
		},
		"project-b": map[string]interface{}{
			"subproject": map[string]interface{}{
				".dummy": "marker",
				"app.py": "print('hello')",
			},
		},
		"no-marker": map[string]interface{}{
			"file.txt": "no marker here",
		},
	})

	roots, err := plugin.FindRoots(fs, "/workspace")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}

	if len(roots) != 2 {
		t.Fatalf("Expected 2 roots, got %d", len(roots))
	}

	expectedRoots := map[string]bool{
		"project-a":            true,
		"project-b/subproject": true,
	}

	for _, root := range roots {
		if !expectedRoots[root] {
			t.Errorf("Unexpected root: %q", root)
		}
		delete(expectedRoots, root)
	}

	if len(expectedRoots) > 0 {
		t.Errorf("Missing expected roots: %v", expectedRoots)
	}
}

func TestDummyPluginProcessRootBasic(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	fs.MustCreateTree("/test", map[string]interface{}{
		".dummy":  "marker",
		"main.go": "package main",
		"test.py": "print('test')",
		"README":  "readme content", // no extension
	})

	result, err := plugin.ProcessRoot(fs, "/test")
	if err != nil {
		t.Fatalf("ProcessRoot failed: %v", err)
	}

	if result.PluginName != "dummy" {
		t.Errorf("Expected plugin name 'dummy', got %q", result.PluginName)
	}

	if result.RootPath != "/test" {
		t.Errorf("Expected root path '/test', got %q", result.RootPath)
	}

	// Check file categorization
	if len(result.Categories["go"]) != 1 {
		t.Errorf("Expected 1 .go file, got %d", len(result.Categories["go"]))
	}

	if len(result.Categories["py"]) != 1 {
		t.Errorf("Expected 1 .py file, got %d", len(result.Categories["py"]))
	}

	if len(result.Categories["no-extension"]) != 1 {
		t.Errorf("Expected 1 file with no extension, got %d", len(result.Categories["no-extension"]))
	}

	if len(result.Categories["dummy"]) != 1 {
		t.Errorf("Expected 1 .dummy file, got %d", len(result.Categories["dummy"]))
	}

	// Verify metadata
	totalFiles, ok := result.Metadata["total_files"].(int)
	if !ok || totalFiles != 4 { // .dummy, main.go, test.py, README
		t.Errorf("Expected 4 total files, got %v", totalFiles)
	}

	totalCategories, ok := result.Metadata["total_categories"].(int)
	if !ok || totalCategories != 4 { // go, py, no-extension, dummy
		t.Errorf("Expected 4 categories, got %v", totalCategories)
	}
}

func TestDummyPluginProcessRootWithSubdirectories(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	fs.MustCreateTree("/project", map[string]interface{}{
		".dummy": "marker",
		"src": map[string]interface{}{
			"main.go":  "package main",
			"utils.go": "package utils",
		},
		"tests": map[string]interface{}{
			"test_main.py": "import unittest",
			"helpers.py":   "def helper(): pass",
		},
		"docs": map[string]interface{}{
			"README.md": "# Project",
			"guide.txt": "User guide",
		},
		"config.yml": "version: 1.0",
	})

	result, err := plugin.ProcessRoot(fs, "/project")
	if err != nil {
		t.Fatalf("ProcessRoot failed: %v", err)
	}

	// Verify categories
	expectedCategoryCounts := map[string]int{
		"go":  2, // main.go, utils.go
		"py":  2, // test_main.py, helpers.py
		"md":  1, // README.md
		"txt": 1, // guide.txt
		"yml": 1, // config.yml
	}

	for category, expectedCount := range expectedCategoryCounts {
		actualCount := len(result.Categories[category])
		if actualCount != expectedCount {
			t.Errorf("Category %q: expected %d files, got %d", category, expectedCount, actualCount)
		}
	}

	// Verify file paths are relative and include subdirectories
	goFiles := result.Categories["go"]
	expectedGoFiles := map[string]bool{
		"src/main.go":  true,
		"src/utils.go": true,
	}

	for _, file := range goFiles {
		if !expectedGoFiles[file] {
			t.Errorf("Unexpected Go file: %q", file)
		}
		delete(expectedGoFiles, file)
	}

	if len(expectedGoFiles) > 0 {
		t.Errorf("Missing expected Go files: %v", expectedGoFiles)
	}
}

func TestDummyPluginProcessRootEmpty(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	// Create empty directory
	fs.MustCreateTree("/empty", map[string]interface{}{})

	result, err := plugin.ProcessRoot(fs, "/empty")
	if err != nil {
		t.Fatalf("ProcessRoot failed: %v", err)
	}

	if len(result.Categories) != 0 {
		t.Errorf("Expected no categories for empty directory, got %d", len(result.Categories))
	}

	totalFiles, ok := result.Metadata["total_files"].(int)
	if !ok || totalFiles != 0 {
		t.Errorf("Expected 0 total files, got %v", totalFiles)
	}
}
