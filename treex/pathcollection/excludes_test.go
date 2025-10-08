// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
	"treex/treex/pattern"
)

func TestPatternFiltering(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create directory structure with files that should be filtered
	fs.MustCreateTree("/project", map[string]interface{}{
		"main.go":   "package main",
		"main.tmp":  "temporary",
		".hidden":   "secret",
		"README.md": "docs",
		"node_modules": map[string]interface{}{
			"package.json": "should be excluded",
			"lib": map[string]interface{}{
				"index.js": "should be excluded",
			},
		},
		"src": map[string]interface{}{
			"app.go":  "package app",
			"app.tmp": "temp file",
		},
		".git": map[string]interface{}{
			"config": "git config",
		},
	})

	// Create filter that excludes tmp files, hidden files, and common directories
	filter := pattern.NewFilterBuilder(fs).
		AddUserExcludes([]string{"*.tmp", "node_modules", ".git"}).
		AddHiddenFilter(false). // showHidden=false means exclude hidden
		Build()

	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root:   "/project",
		Filter: filter,
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
	}

	// Should be present after filtering
	expectedPresent := []string{
		".", // root
		"main.go",
		"README.md",
		"src",
		"src/app.go",
	}

	// Should be excluded by patterns
	expectedAbsent := []string{
		"main.tmp",                  // excluded by *.tmp pattern
		"src/app.tmp",               // excluded by *.tmp pattern
		".hidden",                   // excluded by hidden filter
		"node_modules",              // excluded by explicit pattern
		"node_modules/package.json", // excluded because parent is excluded
		"node_modules/lib",          // excluded because parent is excluded
		"node_modules/lib/index.js", // excluded because parent is excluded
		".git",                      // excluded by explicit pattern
		".git/config",               // excluded because parent is excluded
	}

	for _, expectedPath := range expectedPresent {
		found := false
		for _, actualPath := range paths {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path %q to be present after filtering", expectedPath)
		}
	}

	for _, unexpectedPath := range expectedAbsent {
		for _, actualPath := range paths {
			if actualPath == unexpectedPath {
				t.Errorf("Path %q should be excluded by pattern filtering", unexpectedPath)
			}
		}
	}
}
