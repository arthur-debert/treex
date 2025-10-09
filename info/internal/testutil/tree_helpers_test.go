package testutil_test

import (
	"testing"

	"github.com/jwaldrip/treex/info/internal/testutil"
)

func TestTreeHelpers(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create a simple test structure
	fs.MustCreateTree("/test", map[string]interface{}{
		"README.md": "# Test Project",
		"src": map[string]interface{}{
			"main.go":  "package main",
			"utils.go": "package utils",
		},
		"docs": map[string]interface{}{
			"guide.txt": "User guide",
		},
	})

	// Test BuildFileTree with basic options
	opts := map[string]interface{}{
		"root": "/test",
	}

	tree, err := testutil.BuildFileTree(fs, opts)
	if err != nil {
		t.Fatalf("BuildFileTree failed: %v", err)
	}

	if tree == nil {
		t.Fatal("BuildFileTree returned nil tree")
	}

	// Test MustBuildFileTree (should not panic)
	tree2 := testutil.MustBuildFileTree(fs, opts)
	if tree2 == nil {
		t.Error("MustBuildFileTree returned nil tree")
	}

	// Test tree structure using AssertTreeMatchesMap
	expectedStructure := map[string]interface{}{
		"README.md": "file",
		"src": map[string]interface{}{
			"main.go":  "file",
			"utils.go": "file",
		},
		"docs": map[string]interface{}{
			"guide.txt": "file",
		},
	}

	// This should not panic or fail
	testutil.AssertTreeMatchesMap(t, tree, expectedStructure)
}

func TestTreeHelpersDepthLimit(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create a deeper test structure
	fs.MustCreateTree("/deep", map[string]interface{}{
		"level0.txt": "content",
		"level1": map[string]interface{}{
			"level1.txt": "content",
			"level2": map[string]interface{}{
				"level2.txt": "content",
				"level3": map[string]interface{}{
					"level3.txt": "deep content",
				},
			},
		},
	})

	// Test with depth limit
	opts := map[string]interface{}{
		"root":      "/deep",
		"max_depth": 2, // Should stop at level 2
	}

	tree, err := testutil.BuildFileTree(fs, opts)
	if err != nil {
		t.Fatalf("BuildFileTree with depth limit failed: %v", err)
	}

	if tree == nil {
		t.Fatal("BuildFileTree returned nil tree")
	}

	// Expected structure should only include up to depth 2
	// Note: level2.txt is at depth 3, so it should be excluded by depth 2 limit
	expectedStructure := map[string]interface{}{
		"level0.txt": "file",
		"level1": map[string]interface{}{
			"level1.txt": "file",
			"level2":     map[string]interface{}{
				// level2.txt is at depth 3, should be excluded by depth 2 limit
				// level3 directory is at depth 3, should be excluded
			},
		},
	}

	testutil.AssertTreeMatchesMap(t, tree, expectedStructure)
}

func TestHelperFunctions(t *testing.T) {
	// Test ExpectedFileMap
	fileMap := testutil.ExpectedFileMap("file1.txt", "file2.txt")
	expected := map[string]interface{}{
		"file1.txt": "file",
		"file2.txt": "file",
	}

	if len(fileMap) != len(expected) {
		t.Errorf("ExpectedFileMap length mismatch")
	}

	// Test CombineMaps
	map1 := testutil.ExpectedFileMap("file.txt")
	map2 := testutil.ExpectedDirMap(map[string]map[string]interface{}{
		"dir": map[string]interface{}{},
	})

	combined := testutil.CombineMaps(map1, map2)

	if len(combined) != 2 {
		t.Errorf("CombineMaps should have 2 items, got %d", len(combined))
	}

	if combined["file.txt"] != "file" {
		t.Error("File not properly combined")
	}
}
