// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/jwaldrip/treex/treex/pathcollection"
)

func TestOptionsConfiguratorInterface(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/simple", map[string]interface{}{
		"file.txt": "simple file",
		"subdir": map[string]interface{}{
			"nested.txt": "nested file",
		},
	})

	// Test that configurator interface works and produces a collector
	configurator := pathcollection.NewConfigurator(fs).
		WithRoot("/simple")

	// Test NewCollector method
	collector := configurator.NewCollector()
	if collector == nil {
		t.Fatal("NewCollector() returned nil")
	}

	// Test direct Collect method
	results, err := configurator.Collect()
	if err != nil {
		t.Fatalf("Configurator.Collect() failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Configurator should have collected some paths")
	}

	// Verify we can chain multiple configuration methods
	results2, err := pathcollection.NewConfigurator(fs).
		WithRoot("/simple").
		WithMaxDepth(1).
		WithFilesOnly().
		Collect()

	if err != nil {
		t.Fatalf("Chained configurator methods failed: %v", err)
	}

	// Should only contain files at depth 1
	expectedFiles := []string{"file.txt"}
	if len(results2) != len(expectedFiles) {
		t.Errorf("Expected %d files with chained options, got %d", len(expectedFiles), len(results2))
	}

	for _, result := range results2 {
		if result.IsDir {
			t.Errorf("Expected only files with FilesOnly(), got directory: %s", result.Path)
		}
		if result.Depth > 1 {
			t.Errorf("Expected max depth 1, got depth %d for path %s", result.Depth, result.Path)
		}
	}
}

// options_test.go now only contains basic configurator tests
// Specific option tests have been moved to dedicated files
