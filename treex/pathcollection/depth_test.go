// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
)

func TestDepthLimiting(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create nested directory structure
	fs.MustCreateTree("/deep", map[string]interface{}{
		"level0.txt": "content",
		"level1": map[string]interface{}{
			"level1.txt": "content",
			"level2": map[string]interface{}{
				"level2.txt": "content",
				"level3": map[string]interface{}{
					"level3.txt": "content",
					"level4": map[string]interface{}{
						"level4.txt": "should not appear",
					},
				},
			},
		},
	})

	// Test depth limiting at level 2
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root:     "/deep",
		MaxDepth: 2,
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
	}

	// Should include up to depth 2, but not deeper
	expectedPresent := []string{
		"",                  // root (depth 0)
		"level0.txt",        // depth 1
		"level1",            // depth 1
		"level1/level1.txt", // depth 2
		"level1/level2",     // depth 2
	}

	expectedAbsent := []string{
		"level1/level2/level2.txt",               // depth 3 - should be excluded
		"level1/level2/level3",                   // depth 3 - should be excluded
		"level1/level2/level3/level3.txt",        // depth 4 - should be excluded
		"level1/level2/level3/level4",            // depth 4 - should be excluded
		"level1/level2/level3/level4/level4.txt", // depth 5 - should be excluded
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
			t.Errorf("Expected path %q to be present in results", expectedPath)
		}
	}

	for _, unexpectedPath := range expectedAbsent {
		for _, actualPath := range paths {
			if actualPath == unexpectedPath {
				t.Errorf("Path %q should not be present due to depth limit", unexpectedPath)
			}
		}
	}
}
