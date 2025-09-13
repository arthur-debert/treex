package tree

import (
	"github.com/adebert/treex/pkg/core/types"
)

// AggregatePluginData aggregates plugin data from visible children to parent directories
// This should be called after the tree is built and view mode filtering is applied
func AggregatePluginData(root *types.Node) {
	if root == nil {
		return
	}
	
	// Post-order traversal to aggregate from leaves up
	aggregatePluginDataRecursive(root)
}

func aggregatePluginDataRecursive(node *types.Node) {
	if node == nil || !node.IsDir {
		return
	}
	
	// First, recursively process all child directories
	for _, child := range node.Children {
		if child.IsDir {
			aggregatePluginDataRecursive(child)
		}
	}
	
	// Now aggregate data from visible children
	// Skip if this is a "more items" indicator
	if node.Name != "" && node.Path != "" {
		aggregateFromChildren(node)
	}
}

func aggregateFromChildren(dir *types.Node) {
	if !dir.IsDir || len(dir.Children) == 0 {
		return
	}
	
	// Initialize aggregated values
	totalSize := int64(0)
	totalLines := int64(0)
	totalCloc := int64(0)
	hasSize := false
	hasLines := false
	hasCloc := false
	
	// Aggregate from all visible children
	for _, child := range dir.Children {
		// Skip "more items" indicators
		if child.Path == "" {
			continue
		}
		
		// Get size data (check with plugin prefix)
		if bytesVal, exists := child.Metadata["size_bytes"]; exists {
			if bytes, ok := bytesVal.(int64); ok {
				totalSize += bytes
				hasSize = true
			}
		}
		
		// Get line count data (check with plugin prefix)
		if linesVal, exists := child.Metadata["lc_lines"]; exists {
			if lines, ok := linesVal.(int64); ok {
				totalLines += lines
				hasLines = true
			}
		}
		
		// Get cloc data (check with plugin prefix)
		if clocVal, exists := child.Metadata["cloc_cloc"]; exists {
			if cloc, ok := clocVal.(int64); ok {
				totalCloc += cloc
				hasCloc = true
			}
		}
	}
	
	// Ensure metadata map exists
	if dir.Metadata == nil {
		dir.Metadata = make(map[string]interface{})
	}
	
	// Store aggregated values in directory metadata with plugin prefix
	if hasSize {
		dir.Metadata["size_bytes"] = totalSize
		dir.Metadata["size_is_aggregate"] = true
	}
	
	if hasLines {
		dir.Metadata["lc_lines"] = totalLines
		dir.Metadata["lc_is_aggregate"] = true
	}
	
	if hasCloc {
		dir.Metadata["cloc_cloc"] = totalCloc
		dir.Metadata["cloc_is_aggregate"] = true
	}
}