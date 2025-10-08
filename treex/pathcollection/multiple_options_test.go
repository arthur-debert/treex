// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
	"treex/treex/pattern"
)

func TestMultipleOptionsConfigurator(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/project", map[string]interface{}{
		"main.go":   "package main",
		"main.tmp":  "temporary",
		".hidden":   "secret",
		"README.md": "docs",
		"src": map[string]interface{}{
			"app.go": "package app",
			"deep": map[string]interface{}{
				"nested": map[string]interface{}{
					"file.go": "package nested",
				},
			},
		},
	})

	// Create a filter for testing
	filter := pattern.NewFilterBuilder(fs).
		AddUserExcludes([]string{"*.tmp"}).
		AddHiddenFilter(false).
		Build()

	// Test configurator fluent interface
	results, err := pathcollection.NewConfigurator(fs).
		WithRoot("/project").
		WithMaxDepth(2).
		WithFilter(filter).
		Collect()

	if err != nil {
		t.Fatalf("Configurator collection failed: %v", err)
	}

	// Verify results
	foundDeep := false
	foundNested := false
	foundTemp := false
	foundHidden := false

	for _, result := range results {
		switch result.Path {
		case "src/deep":
			foundDeep = true
			if result.Depth != 2 {
				t.Errorf("Expected depth 2 for src/deep, got %d", result.Depth)
			}
		case "src/deep/nested":
			foundNested = true
			t.Error("src/deep/nested should be excluded by depth limit")
		case "main.tmp":
			foundTemp = true
			t.Error("main.tmp should be excluded by pattern filter")
		case ".hidden":
			foundHidden = true
			t.Error(".hidden should be excluded by hidden filter")
		}
	}

	if !foundDeep {
		t.Error("Expected to find src/deep within depth limit")
	}
	if foundNested {
		t.Error("Should not find nested paths beyond depth limit")
	}
	if foundTemp {
		t.Error("Should not find temp files due to pattern filter")
	}
	if foundHidden {
		t.Error("Should not find hidden files due to hidden filter")
	}
}
