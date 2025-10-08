// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"sort"
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/jwaldrip/treex/treex/pathcollection"
)

func TestBasicPathCollection(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create a simple directory structure
	fs.MustCreateTree("/project", map[string]interface{}{
		"file1.txt": "content1",
		"file2.go":  "package main",
		"src": map[string]interface{}{
			"main.go":  "package main",
			"utils.go": "package main",
			"lib": map[string]interface{}{
				"helper.go": "package lib",
			},
		},
		"docs": map[string]interface{}{
			"README.md": "# Project",
		},
	})

	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/project",
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	// Convert to paths for easier testing
	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
	}
	sort.Strings(paths)

	expected := []string{
		".", // root directory
		"docs",
		"docs/README.md",
		"file1.txt",
		"file2.go",
		"src",
		"src/lib",
		"src/lib/helper.go",
		"src/main.go",
		"src/utils.go",
	}
	sort.Strings(expected)

	if len(paths) != len(expected) {
		t.Errorf("Expected %d paths, got %d", len(expected), len(paths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", paths)
		return
	}

	for i, expectedPath := range expected {
		if paths[i] != expectedPath {
			t.Errorf("Path mismatch at index %d: expected %q, got %q", i, expectedPath, paths[i])
		}
	}
}

// Depth limiting and pattern filtering tests moved to dedicated files

// File vs directory filtering tests moved to filevsdir_test.go

func TestPathInfoDetails(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/test", map[string]interface{}{
		"root.txt": "root content",
		"subdir": map[string]interface{}{
			"nested.txt": "nested content",
		},
	})

	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/test",
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	// Verify depth calculations
	depthMap := make(map[string]int)
	for _, result := range results {
		depthMap[result.Path] = result.Depth
	}

	expectedDepths := map[string]int{
		".":                 0, // root
		"root.txt":          1,
		"subdir":            1,
		"subdir/nested.txt": 2,
	}

	for path, expectedDepth := range expectedDepths {
		if actualDepth, exists := depthMap[path]; !exists {
			t.Errorf("Path %q not found in results", path)
		} else if actualDepth != expectedDepth {
			t.Errorf("Path %q has depth %d, expected %d", path, actualDepth, expectedDepth)
		}
	}

	// Verify absolute paths are set
	for _, result := range results {
		if result.AbsolutePath == "" {
			t.Errorf("AbsolutePath not set for %q", result.Path)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	fs := testutil.NewTestFS()

	// Test with non-existent root
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/nonexistent",
	})

	_, err := collector.Collect()
	if err == nil {
		t.Error("Expected error for non-existent root directory")
	}

	// Test with file as root (should fail)
	fs.MustCreateTree("/test", map[string]interface{}{
		"file.txt": "content",
	})

	collector = pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/test/file.txt",
	})

	_, err = collector.Collect()
	if err == nil {
		t.Error("Expected error when root is a file, not directory")
	}
}
