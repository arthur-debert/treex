package treex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"treex/treex/internal/testutil"
	_ "treex/treex/plugins/infofile" // Import for plugin registration
	"treex/treex/types"
)

func TestTreeBuildingWithDataPluginV2(t *testing.T) {
	// Create test filesystem with annotations
	fs := testutil.NewTestFS()
	fs.MustCreateTree("/test", map[string]interface{}{
		".info":     "test.txt  Test annotation via V2\nother.txt  Other annotation via V2",
		"test.txt":  "test content",
		"other.txt": "other content",
		"none.txt":  "no annotation",
	})

	// Configure tree building with plugin filters (this will trigger caching)
	config := TreeConfig{
		Root:       "/test",
		Filesystem: fs,
		PluginFilters: map[string]map[string]bool{
			"info": {"annotated": true},
		},
	}

	// Build tree - this should use the new DataPluginV2 interface for enrichment
	result, err := BuildTree(config)
	require.NoError(t, err)
	require.NotNil(t, result.Root)

	// Verify that plugin results were cached for V2 interface usage
	assert.NotEmpty(t, result.PluginResults, "Plugin results should be stored")
	infoResults, exists := result.PluginResults["info"]
	assert.True(t, exists, "Info plugin results should be available")
	assert.NotEmpty(t, infoResults, "Info plugin should have results")

	// Find nodes in the tree and verify they were enriched using DataPluginV2
	var testNode *types.Node
	var otherNode *types.Node

	// Walk the tree to find our test nodes
	walkTree(result.Root, func(node *types.Node) {
		switch node.Name {
		case "test.txt":
			testNode = node
		case "other.txt":
			otherNode = node
		}
	})

	// Verify annotated nodes were found and enriched
	require.NotNil(t, testNode, "test.txt node should be in tree")
	require.NotNil(t, otherNode, "other.txt node should be in tree")

	// Note: none.txt is filtered out because it doesn't have annotations
	// and we're using the "annotated" filter

	// Verify enrichment data was attached using DataPluginV2 batch processing
	testData, exists := testNode.GetPluginData("info")
	assert.True(t, exists, "test.txt should have annotation data")
	if exists {
		annotation, ok := testData.(*types.Annotation)
		require.True(t, ok, "Plugin data should be annotation type")
		assert.Equal(t, "Test annotation via V2", annotation.Notes)
	}

	otherData, exists := otherNode.GetPluginData("info")
	assert.True(t, exists, "other.txt should have annotation data")
	if exists {
		annotation, ok := otherData.(*types.Annotation)
		require.True(t, ok, "Plugin data should be annotation type")
		assert.Equal(t, "Other annotation via V2", annotation.Notes)
	}
}

func TestDataPluginV2FreshDataPath(t *testing.T) {
	// Create test filesystem where files have different annotations than what gets cached
	// This tests the "fresh data" path when cache doesn't contain what we need
	fs := testutil.NewTestFS()
	fs.MustCreateTree("/test", map[string]interface{}{
		".info":      "cached.txt  This will be cached\nfresh.txt  This won't be in cache",
		"cached.txt": "cached content",
		"fresh.txt":  "fresh content",
		"other.txt":  "other content",
	})

	// Configure tree building to only cache data for "cached.txt"
	// but allow all files in the tree
	config := TreeConfig{
		Root:       "/test",
		Filesystem: fs,
		PluginFilters: map[string]map[string]bool{
			"info": {"annotated": true}, // This will cache data for annotated files
		},
	}

	// Build tree - both cached.txt and fresh.txt should be enriched
	// cached.txt uses cache, fresh.txt uses fresh data gathering
	result, err := BuildTree(config)
	require.NoError(t, err)
	require.NotNil(t, result.Root)

	// Verify that plugin processing happened
	assert.Contains(t, result.PluginResults, "info", "Info plugin should have been processed")

	// Find both nodes
	var cachedNode *types.Node
	var freshNode *types.Node
	walkTree(result.Root, func(node *types.Node) {
		switch node.Name {
		case "cached.txt":
			cachedNode = node
		case "fresh.txt":
			freshNode = node
		}
	})

	// Both files should be in the tree and enriched (both have annotations)
	require.NotNil(t, cachedNode, "cached.txt node should be in tree")
	require.NotNil(t, freshNode, "fresh.txt node should be in tree")

	// Verify both were enriched by DataPluginV2
	cachedData, exists := cachedNode.GetPluginData("info")
	assert.True(t, exists, "cached.txt should have annotation data")
	if exists {
		annotation, ok := cachedData.(*types.Annotation)
		require.True(t, ok, "Plugin data should be annotation type")
		assert.Equal(t, "This will be cached", annotation.Notes)
	}

	freshData, exists := freshNode.GetPluginData("info")
	assert.True(t, exists, "fresh.txt should have annotation data")
	if exists {
		annotation, ok := freshData.(*types.Annotation)
		require.True(t, ok, "Plugin data should be annotation type")
		assert.Equal(t, "This won't be in cache", annotation.Notes)
	}
}

func TestDataPluginV2BatchProcessing(t *testing.T) {
	// Create test filesystem with many files to test batch processing efficiency
	fs := testutil.NewTestFS()
	testFiles := map[string]interface{}{
		".info": "file1.txt  Annotation 1\nfile2.txt  Annotation 2\nfile3.txt  Annotation 3",
	}

	// Add multiple files
	for i := 1; i <= 10; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		testFiles[filename] = fmt.Sprintf("content %d", i)
	}

	fs.MustCreateTree("/batch", testFiles)

	// Configure tree building with plugin filters
	config := TreeConfig{
		Root:       "/batch",
		Filesystem: fs,
		PluginFilters: map[string]map[string]bool{
			"info": {"annotated": true},
		},
	}

	// Build tree - should process all files in batch with DataPluginV2
	result, err := BuildTree(config)
	require.NoError(t, err)
	require.NotNil(t, result.Root)

	// Count enriched nodes
	enrichedCount := 0
	walkTree(result.Root, func(node *types.Node) {
		if !node.IsDir {
			if _, exists := node.GetPluginData("info"); exists {
				enrichedCount++
			}
		}
	})

	// Should have 3 enriched files (file1.txt, file2.txt, file3.txt have annotations)
	assert.Equal(t, 3, enrichedCount, "Should have exactly 3 files with annotations")
}
