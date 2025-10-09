package treex

import (
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	_ "github.com/jwaldrip/treex/treex/plugins/infofile" // Import for plugin registration
	"github.com/jwaldrip/treex/treex/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeBuildingWithCachedEnrichment(t *testing.T) {
	// Create test filesystem with annotations
	fs := testutil.NewTestFS()
	fs.MustCreateTree("/test", map[string]interface{}{
		".info":     "test.txt  Test annotation\nother.txt  Other annotation",
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

	// Build tree - this should use cached enrichment for efficiency
	result, err := BuildTree(config)
	require.NoError(t, err)
	require.NotNil(t, result.Root)

	// Verify that plugin results were cached
	assert.NotEmpty(t, result.PluginResults, "Plugin results should be stored")
	infoResults, exists := result.PluginResults["info"]
	assert.True(t, exists, "Info plugin results should be available")
	assert.NotEmpty(t, infoResults, "Info plugin should have results")

	// Verify that cache was populated in plugin results
	for i, pluginResult := range infoResults {
		assert.NotNil(t, pluginResult.Cache, "Plugin result should have cache")
		if pluginResult.Cache != nil {
			cachedAnnotations, hasCachedAnnotations := pluginResult.Cache["annotations"]
			assert.True(t, hasCachedAnnotations, "Cache should contain annotations")
			t.Logf("Plugin result %d cache contents: %+v", i, pluginResult.Cache)
			if hasCachedAnnotations {
				t.Logf("Cached annotations type: %T, value: %+v", cachedAnnotations, cachedAnnotations)
			}
		}
	}

	// Find nodes in the tree and verify they were enriched with cached data
	var testNode *types.Node
	var otherNode *types.Node
	var allNodes []*types.Node

	// Walk the tree to find our test nodes
	walkTree(result.Root, func(node *types.Node) {
		allNodes = append(allNodes, node)
		switch node.Name {
		case "test.txt":
			testNode = node
		case "other.txt":
			otherNode = node
		}
	})

	// Debug: show all nodes found
	t.Logf("Found %d nodes in tree:", len(allNodes))
	for _, node := range allNodes {
		t.Logf("  Node: %s (path: %s, isDir: %v)", node.Name, node.Path, node.IsDir)
	}

	// Verify nodes were found and enriched
	require.NotNil(t, testNode, "test.txt node should be in tree")
	require.NotNil(t, otherNode, "other.txt node should be in tree")

	// Debug: Check what plugin data is available on nodes
	t.Logf("test.txt node data keys: %v", getNodeDataKeys(testNode))
	t.Logf("other.txt node data keys: %v", getNodeDataKeys(otherNode))

	// Verify enrichment data was attached using cached results
	testData, exists := testNode.GetPluginData("info")
	assert.True(t, exists, "test.txt should have annotation data")
	if exists {
		annotation, ok := testData.(*types.Annotation)
		require.True(t, ok, "Plugin data should be annotation type")
		assert.Equal(t, "Test annotation", annotation.Notes)
	}

	otherData, exists := otherNode.GetPluginData("info")
	assert.True(t, exists, "other.txt should have annotation data")
	if exists {
		annotation, ok := otherData.(*types.Annotation)
		require.True(t, ok, "Plugin data should be annotation type")
		assert.Equal(t, "Other annotation", annotation.Notes)
	}
}

// getNodeDataKeys returns the keys in a node's data map for debugging
func getNodeDataKeys(node *types.Node) []string {
	if node == nil || node.Data == nil {
		return nil
	}

	keys := make([]string, 0, len(node.Data))
	for k := range node.Data {
		keys = append(keys, k)
	}
	return keys
}

// walkTree recursively walks a tree and calls fn for each node
func walkTree(node *types.Node, fn func(*types.Node)) {
	if node == nil {
		return
	}

	fn(node)

	for _, child := range node.Children {
		walkTree(child, fn)
	}
}
