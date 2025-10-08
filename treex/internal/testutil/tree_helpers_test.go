package testutil_test

import (
	"testing"

	"treex/treex/internal/testutil"
)

// TODO: Implement these tests when tree builder is ready
// These tests require the actual tree building functionality
// which will be implemented in the next phase

func TestTreeHelpers(t *testing.T) {
	t.Skip("Tree building functionality not yet implemented")
}

func TestTreeHelpersDepthLimit(t *testing.T) {
	t.Skip("Tree building functionality not yet implemented")
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
