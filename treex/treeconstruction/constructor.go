// see docs/dev/architecture.txt - Phase 3: Tree Construction
package treeconstruction

import (
	"path/filepath"
	"sort"

	"github.com/jwaldrip/treex/treex/pathcollection"
	"github.com/jwaldrip/treex/treex/types"
)

// Constructor builds a tree of nodes from a flat list of paths.
type Constructor struct{}

// NewConstructor creates a new tree constructor.
func NewConstructor() *Constructor {
	return &Constructor{}
}

// BuildTree constructs a tree from a list of PathInfo objects.
// The algorithm relies on the input paths being sorted to ensure that parent
// directories are always processed before their children.
func (c *Constructor) BuildTree(paths []pathcollection.PathInfo) *types.Node {
	if len(paths) == 0 {
		return nil
	}

	// A defensive sort to ensure parent directories come before their children.
	// For example, "a" will always come before "a/b".
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})

	// nodeMap provides fast O(1) average time complexity access to any node by its path.
	nodeMap := make(map[string]*types.Node)
	var root *types.Node

	for _, p := range paths {
		node := &types.Node{
			Name:  filepath.Base(p.Path),
			Path:  p.Path,
			IsDir: p.IsDir,
			Size:  p.Size,
			Data:  make(map[string]interface{}),
		}

		// Store the newly created node in the map for future lookups.
		nodeMap[p.Path] = node

		// The first path in the sorted slice is the root.
		if root == nil {
			root = node
			continue
		}

		// Determine the parent's path. For a path like "a/b/c", the parent is "a/b".
		// For a top-level path like "a", the parent is ".".
		parentPath := filepath.Dir(p.Path)

		// The parent must exist in the map due to the lexicographical sort order.
		if parent, ok := nodeMap[parentPath]; ok {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}
	}

	return root
}
