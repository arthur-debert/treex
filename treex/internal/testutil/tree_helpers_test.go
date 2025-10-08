package testutil_test

import (
	"testing"

	"treex/treex/internal/testutil"
)

func TestTreeHelpers(t *testing.T) {
	// Create test filesystem
	fs := testutil.NewTestFS()
	fs.MustCreateTree("/project", map[string]interface{}{
		"src": map[string]interface{}{
			"main.go":  "package main",
			"utils.go": "package main",
			"lib": map[string]interface{}{
				"helper.go": "package lib",
			},
		},
		"docs": map[string]interface{}{
			"README.txt": "documentation",
		},
		".hidden":  "secret",
		"file.txt": "content",
	})

	// Test 1: Basic tree structure (no hidden files)
	tree := testutil.MustBuildFileTree(fs, map[string]interface{}{
		"root":   "/project",
		"hidden": false,
	})

	expected := testutil.CombineMaps(
		testutil.ExpectedDirMap(map[string]map[string]interface{}{
			"src": testutil.CombineMaps(
				testutil.ExpectedFileMap("main.go", "utils.go"),
				testutil.ExpectedDirMap(map[string]map[string]interface{}{
					"lib": testutil.ExpectedFileMap("helper.go"),
				}),
			),
			"docs": testutil.ExpectedFileMap("README.txt"),
		}),
		testutil.ExpectedFileMap("file.txt"),
	)

	testutil.AssertTreeMatchesMap(t, tree, expected)

	// Test 2: With hidden files
	treeWithHidden := testutil.MustBuildFileTree(fs, map[string]interface{}{
		"root":   "/project",
		"hidden": true,
	})

	expectedWithHidden := testutil.CombineMaps(
		expected,
		testutil.ExpectedFileMap(".hidden"),
	)

	testutil.AssertTreeMatchesMap(t, treeWithHidden, expectedWithHidden)

	// Test 3: Directories only
	dirsOnlyTree := testutil.MustBuildFileTree(fs, map[string]interface{}{
		"root":      "/project",
		"dirs_only": true,
	})

	expectedDirsOnly := testutil.ExpectedDirMap(map[string]map[string]interface{}{
		"src": testutil.ExpectedDirMap(map[string]map[string]interface{}{
			"lib": map[string]interface{}{},
		}),
		"docs": map[string]interface{}{},
	})

	testutil.AssertTreeMatchesMap(t, dirsOnlyTree, expectedDirsOnly)
}

func TestTreeHelpersDepthLimit(t *testing.T) {
	fs := testutil.NewTestFS()
	fs.MustCreateTree("/deep", map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"too-deep.txt": "should not appear",
					},
				},
			},
		},
	})

	// Test depth limiting
	tree := testutil.MustBuildFileTree(fs, map[string]interface{}{
		"root":      "/deep",
		"max_depth": 3,
	})

	// Should only go 3 levels deep
	expected := testutil.ExpectedDirMap(map[string]map[string]interface{}{
		"level1": testutil.ExpectedDirMap(map[string]map[string]interface{}{
			"level2": testutil.ExpectedDirMap(map[string]map[string]interface{}{
				"level3": map[string]interface{}{}, // Empty because level4 is beyond maxDepth
			}),
		}),
	})

	testutil.AssertTreeMatchesMap(t, tree, expected)
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
