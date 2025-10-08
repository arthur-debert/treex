// see docs/dev/architecture.txt - Phase 3: Tree Construction
package treeconstruction

import (
	"path/filepath"
	"sort"

	"treex/treex/pathcollection"
	"treex/treex/types"
)

// TreeConstructor builds node trees from collected paths
type TreeConstructor struct {
	// pathToNode maps relative paths to their corresponding nodes
	// This enables fast lookup when establishing parent-child relationships
	pathToNode map[string]*types.Node
}

// NewTreeConstructor creates a new tree constructor
func NewTreeConstructor() *TreeConstructor {
	return &TreeConstructor{
		pathToNode: make(map[string]*types.Node),
	}
}

// BuildTree constructs a node tree from collected path information
// This is a fast in-memory operation that processes paths in a single pass
// Returns the root node of the constructed tree
func (tc *TreeConstructor) BuildTree(pathInfos []pathcollection.PathInfo) (*types.Node, error) {
	// Reset the path-to-node mapping for fresh construction
	tc.pathToNode = make(map[string]*types.Node)

	// Sort paths to ensure parents are processed before children
	// This is critical for establishing correct parent-child relationships
	// Example: [".", "src", "src/main.go"] ensures root and src exist before src/main.go
	sortedPaths := make([]pathcollection.PathInfo, len(pathInfos))
	copy(sortedPaths, pathInfos)
	sort.Slice(sortedPaths, func(i, j int) bool {
		// Sort by path length first (shorter paths = higher in tree hierarchy)
		// Then lexicographically for consistent ordering
		if len(sortedPaths[i].Path) != len(sortedPaths[j].Path) {
			return len(sortedPaths[i].Path) < len(sortedPaths[j].Path)
		}
		return sortedPaths[i].Path < sortedPaths[j].Path
	})

	var root *types.Node

	// Single pass through sorted paths to build the tree
	// This algorithm ensures O(n) complexity where n is the number of paths
	for _, pathInfo := range sortedPaths {
		// Create node from path information
		node := tc.createNodeFromPathInfo(pathInfo)

		// Store in lookup map for fast parent-child relationship establishment
		tc.pathToNode[pathInfo.Path] = node

		// Handle root node specially - it has no parent
		if pathInfo.Path == "." {
			root = node
			continue
		}

		// Find and establish parent-child relationship
		tc.establishParentChildRelationship(node, pathInfo.Path)
	}

	return root, nil
}

// createNodeFromPathInfo converts PathInfo to Node structure
// Extracts the filename/dirname and sets up basic node properties
func (tc *TreeConstructor) createNodeFromPathInfo(pathInfo pathcollection.PathInfo) *types.Node {
	// Extract just the filename or directory name from the path
	// Examples: "src/main.go" -> "main.go", "src" -> "src", "." -> "."
	name := filepath.Base(pathInfo.Path)
	if pathInfo.Path == "." {
		name = "." // Keep root as "." for clarity
	}

	return &types.Node{
		Name:         name,
		Path:         pathInfo.AbsolutePath, // Full filesystem path
		RelativePath: pathInfo.Path,         // Path relative to tree root
		IsDir:        pathInfo.IsDir,
		Annotation:   nil, // Will be set later by annotation plugin
		Children:     nil, // Will be populated as children are processed
		Parent:       nil, // Will be set when establishing relationships
	}
}

// establishParentChildRelationship finds the parent of a node and establishes the relationship
// This function relies on the fact that paths are processed in sorted order (parents before children)
func (tc *TreeConstructor) establishParentChildRelationship(node *types.Node, nodePath string) {
	// Calculate parent path by removing the last path component
	// Examples: "src/main.go" -> "src", "src" -> ".", "level1/level2/file.txt" -> "level1/level2"
	parentPath := filepath.Dir(nodePath)

	// Handle special case where filepath.Dir() returns different values on different systems
	// Normalize to use "." for immediate children of root
	if parentPath == "/" || parentPath == "\\" || parentPath == nodePath {
		parentPath = "."
	}

	// Look up parent node in our path-to-node mapping
	// Since we process paths in sorted order, parent must already exist
	parentNode, exists := tc.pathToNode[parentPath]
	if !exists {
		// This should not happen with properly sorted paths, but handle gracefully
		// Create missing parent directory nodes on demand
		parentNode = tc.createMissingParentNode(parentPath)
		tc.pathToNode[parentPath] = parentNode

		// Recursively establish parent's relationships if needed
		if parentPath != "." {
			tc.establishParentChildRelationship(parentNode, parentPath)
		}
	}

	// Establish bidirectional parent-child relationship
	node.Parent = parentNode

	// Initialize parent's children slice if this is the first child
	if parentNode.Children == nil {
		parentNode.Children = make([]*types.Node, 0)
	}

	// Add this node to parent's children list
	parentNode.Children = append(parentNode.Children, node)
}

// createMissingParentNode creates a directory node for missing parent paths
// This handles cases where intermediate directories might not be in the collected paths
// (though this should be rare with proper path collection)
func (tc *TreeConstructor) createMissingParentNode(parentPath string) *types.Node {
	// Extract name from path
	name := filepath.Base(parentPath)
	if parentPath == "." {
		name = "."
	}

	// Create directory node with minimal information
	// AbsolutePath will be empty since we don't have the full filesystem path
	return &types.Node{
		Name:         name,
		Path:         "", // Empty since we don't have absolute path
		RelativePath: parentPath,
		IsDir:        true, // Missing parents are always directories
		Annotation:   nil,
		Children:     nil,
		Parent:       nil,
	}
}

// GetNodeByPath returns the node at the specified relative path
// This is useful for testing and for other phases that need to access specific nodes
func (tc *TreeConstructor) GetNodeByPath(relativePath string) *types.Node {
	return tc.pathToNode[relativePath]
}

// GetAllNodes returns all nodes in the constructed tree
// Useful for debugging and for phases that need to iterate over all nodes
func (tc *TreeConstructor) GetAllNodes() []*types.Node {
	nodes := make([]*types.Node, 0, len(tc.pathToNode))
	for _, node := range tc.pathToNode {
		nodes = append(nodes, node)
	}
	return nodes
}

// WalkTree performs a depth-first traversal of the tree starting from the given node
// Calls the provided function for each node in the tree
// This is useful for implementing tree operations and for other phases
func WalkTree(root *types.Node, fn func(*types.Node) error) error {
	// Visit current node
	if err := fn(root); err != nil {
		return err
	}

	// Recursively visit all children
	for _, child := range root.Children {
		if err := WalkTree(child, fn); err != nil {
			return err
		}
	}

	return nil
}

// GetTreeStats returns statistics about the constructed tree
// Useful for debugging and performance analysis
type TreeStats struct {
	TotalNodes      int
	DirectoryNodes  int
	FileNodes       int
	MaxDepth        int
	AverageChildren float64
}

// CalculateTreeStats computes statistics for the given tree
func CalculateTreeStats(root *types.Node) TreeStats {
	stats := TreeStats{}

	// Track depth during traversal
	var calculateStats func(*types.Node, int)
	var totalChildrenCount int
	var directoriesWithChildren int

	calculateStats = func(node *types.Node, depth int) {
		stats.TotalNodes++

		if node.IsDir {
			stats.DirectoryNodes++
			childrenCount := len(node.Children)
			if childrenCount > 0 {
				totalChildrenCount += childrenCount
				directoriesWithChildren++
			}
		} else {
			stats.FileNodes++
		}

		if depth > stats.MaxDepth {
			stats.MaxDepth = depth
		}

		// Recursively process children
		for _, child := range node.Children {
			calculateStats(child, depth+1)
		}
	}

	calculateStats(root, 0)

	// Calculate average children per directory (only for directories that have children)
	if directoriesWithChildren > 0 {
		stats.AverageChildren = float64(totalChildrenCount) / float64(directoriesWithChildren)
	}

	return stats
}
