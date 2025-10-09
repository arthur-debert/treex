package treex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"treex/treex/internal/testutil"
	_ "treex/treex/plugins/infofile" // Import for plugin registration
	"treex/treex/types"
)

func TestTreeBuildingWithPluginFiltering(t *testing.T) {
	tests := []struct {
		name          string
		fsStructure   map[string]interface{}
		pluginFilters map[string]map[string]bool
		expectedFiles []string // Files that should appear in final tree
	}{
		{
			name: "info plugin filtering - only annotated files",
			fsStructure: map[string]interface{}{
				".info":     "test.txt  This is annotated",
				"test.txt":  "test content",
				"other.txt": "other content",
				"README.md": "readme content",
			},
			pluginFilters: map[string]map[string]bool{
				"info": {"annotated": true},
			},
			expectedFiles: []string{"test.txt"}, // Only annotated file should remain
		},
		{
			name: "no plugin filters - all files shown",
			fsStructure: map[string]interface{}{
				".info":     "test.txt  This is annotated",
				"test.txt":  "test content",
				"other.txt": "other content",
				"README.md": "readme content",
			},
			pluginFilters: map[string]map[string]bool{},
			expectedFiles: []string{"test.txt", "other.txt", "README.md"}, // All files shown (excluding .info)
		},
		{
			name: "info plugin filtering with multiple files",
			fsStructure: map[string]interface{}{
				".info":     "doc.txt  Documentation\nother.txt  Other file annotation",
				"doc.txt":   "documentation",
				"other.txt": "other content",
				"script.py": "print('hello')",
				"README.md": "readme",
			},
			pluginFilters: map[string]map[string]bool{
				"info": {"annotated": true},
			},
			expectedFiles: []string{"doc.txt", "other.txt"}, // Only annotated files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test filesystem
			fs := testutil.NewTestFS()
			fs.MustCreateTree("/test", tt.fsStructure)

			// Configure tree building with plugin filters
			config := TreeConfig{
				Root:          "/test",
				Filesystem:    fs,
				MaxDepth:      0,
				PluginFilters: tt.pluginFilters,
			}

			// Build tree
			result, err := BuildTree(config)
			require.NoError(t, err)
			require.NotNil(t, result.Root)

			// Collect all file names from the tree
			actualFiles := collectFileNames(result.Root)

			// Verify only expected files are present
			assert.ElementsMatch(t, tt.expectedFiles, actualFiles,
				"Tree should contain only the files matching enabled plugin categories")
		})
	}
}

// collectFileNames recursively collects all file names from a tree node
func collectFileNames(node *types.Node) []string {
	if node == nil {
		return nil
	}

	var files []string

	// Add this node if it's a file
	if !node.IsDir {
		files = append(files, node.Name)
	}

	// Recursively collect from children
	for _, child := range node.Children {
		files = append(files, collectFileNames(child)...)
	}

	return files
}
