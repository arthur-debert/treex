// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
)

func TestConfiguratorDirsOnly(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/test", map[string]interface{}{
		"file.txt": "content",
		"subdir": map[string]interface{}{
			"nested.txt": "nested",
		},
	})

	results, err := pathcollection.NewConfigurator(fs).
		WithRoot("/test").
		WithDirsOnly().
		Collect()

	if err != nil {
		t.Fatalf("Configurator collection failed: %v", err)
	}

	// All results should be directories
	for _, result := range results {
		if !result.IsDir {
			t.Errorf("Expected only directories, but found file: %s", result.Path)
		}
	}

	// Should have root and subdir
	expectedDirs := []string{"", "subdir"}
	if len(results) != len(expectedDirs) {
		t.Errorf("Expected %d directories, got %d", len(expectedDirs), len(results))
	}
}

func TestConfiguratorFilesOnly(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/test", map[string]interface{}{
		"file.txt": "content",
		"subdir": map[string]interface{}{
			"nested.txt": "nested",
		},
	})

	results, err := pathcollection.NewConfigurator(fs).
		WithRoot("/test").
		WithFilesOnly().
		Collect()

	if err != nil {
		t.Fatalf("Configurator collection failed: %v", err)
	}

	// All results should be files
	for _, result := range results {
		if result.IsDir {
			t.Errorf("Expected only files, but found directory: %s", result.Path)
		}
	}

	// Should have both files
	expectedFiles := []string{"file.txt", "subdir/nested.txt"}
	if len(results) != len(expectedFiles) {
		t.Errorf("Expected %d files, got %d", len(expectedFiles), len(results))
	}
}

func TestConfiguratorMutualExclusion(t *testing.T) {
	fs := testutil.NewTestFS()

	// Test that DirsOnly and FilesOnly are mutually exclusive
	configurator := pathcollection.NewConfigurator(fs).
		WithRoot("/test").
		WithDirsOnly().
		WithFilesOnly() // This should override DirsOnly

	collector := configurator.NewCollector()

	// Check that the final configuration has FilesOnly=true and DirsOnly=false
	// We can't directly access options, so we'll test the behavior instead
	fs.MustCreateTree("/test", map[string]interface{}{
		"file.txt": "content",
		"subdir":   map[string]interface{}{},
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	// Should only contain files (FilesOnly should have overridden DirsOnly)
	for _, result := range results {
		if result.IsDir {
			t.Errorf("Expected only files due to mutual exclusion, but found directory: %s", result.Path)
		}
	}
}
