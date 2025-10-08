// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
	"treex/treex/pattern"
)

func TestHiddenFileFiltering(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create directory structure with hidden files and directories
	fs.MustCreateTree("/project", map[string]interface{}{
		"visible.txt":  "visible content",
		".hidden_file": "hidden content",
		".hidden_dir": map[string]interface{}{
			"nested.txt":     "nested in hidden dir",
			".double_hidden": "double hidden",
		},
		"normal_dir": map[string]interface{}{
			"normal.txt":        "normal content",
			".hidden_in_normal": "hidden in normal dir",
		},
	})

	// Test with hidden files excluded (default behavior)
	filterExcludeHidden := pattern.NewFilterBuilder(fs).
		AddHiddenFilter(false). // showHidden=false means exclude hidden
		Build()

	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root:   "/project",
		Filter: filterExcludeHidden,
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
	}

	// Should be present (visible files/dirs)
	expectedPresent := []string{
		".", // root
		"visible.txt",
		"normal_dir",
		"normal_dir/normal.txt",
	}

	// Should be excluded (hidden files/dirs and their contents)
	expectedAbsent := []string{
		".hidden_file",
		".hidden_dir",
		".hidden_dir/nested.txt",       // excluded because parent is excluded
		".hidden_dir/.double_hidden",   // excluded because parent is excluded
		"normal_dir/.hidden_in_normal", // excluded by hidden filter
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
			t.Errorf("Expected visible path %q to be present", expectedPath)
		}
	}

	for _, unexpectedPath := range expectedAbsent {
		for _, actualPath := range paths {
			if actualPath == unexpectedPath {
				t.Errorf("Hidden path %q should be excluded", unexpectedPath)
			}
		}
	}
}

func TestHiddenFileIncluding(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create directory structure with hidden files
	fs.MustCreateTree("/project", map[string]interface{}{
		"visible.txt":  "visible content",
		".hidden_file": "hidden content",
		".hidden_dir": map[string]interface{}{
			"nested.txt": "nested in hidden dir",
		},
	})

	// Test with hidden files included
	filterIncludeHidden := pattern.NewFilterBuilder(fs).
		AddHiddenFilter(true). // showHidden=true means include hidden
		Build()

	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root:   "/project",
		Filter: filterIncludeHidden,
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
	}

	// Should include both visible and hidden files
	expectedAll := []string{
		".", // root
		"visible.txt",
		".hidden_file",
		".hidden_dir",
		".hidden_dir/nested.txt",
	}

	for _, expectedPath := range expectedAll {
		found := false
		for _, actualPath := range paths {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected path %q to be present when including hidden files", expectedPath)
		}
	}
}
