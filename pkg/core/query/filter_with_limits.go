package query

import (
	"github.com/adebert/treex/pkg/core/limits"
	"github.com/adebert/treex/pkg/core/types"
)

// FilterTreeWithLimits recursively filters a tree based on the query with performance limits
func FilterTreeWithLimits(node *types.Node, matcher *Matcher, limiter *limits.Limiter) (*types.Node, error) {
	// Check if we've exceeded the runtime limit
	if !limiter.CheckTimeout() {
		// Return what we have so far
		return nil, nil
	}
	
	// First, check if this node matches
	matches, err := matcher.Matches(node)
	if err != nil {
		return nil, err
	}
	
	// Record that we processed this file
	if !node.IsDir {
		limiter.RecordFileProcessed()
	}
	
	// For directories, we need to check children first
	if node.IsDir && len(node.Children) > 0 {
		var filteredChildren []*types.Node
		childMatches := false
		
		for _, child := range node.Children {
			// Check timeout before processing each child
			if !limiter.CheckTimeout() {
				break
			}
			
			filteredChild, err := FilterTreeWithLimits(child, matcher, limiter)
			if err != nil {
				return nil, err
			}
			if filteredChild != nil {
				filteredChildren = append(filteredChildren, filteredChild)
				childMatches = true
			}
		}
		
		// A directory is included if:
		// 1. It matches the query itself, OR
		// 2. It has at least one matching child
		if matches || childMatches {
			// Create a copy of the node with filtered children
			filteredNode := &types.Node{
				Name:         node.Name,
				Path:         node.Path,
				RelativePath: node.RelativePath,
				IsDir:        node.IsDir,
				Annotation:   node.Annotation,
				Metadata:     node.Metadata,
				Children:     filteredChildren,
				Parent:       node.Parent,
			}
			return filteredNode, nil
		}
		
		// Directory doesn't match and has no matching children
		return nil, nil
	}
	
	// For files, simply return the node if it matches
	if matches {
		// Create a copy to avoid modifying the original
		filteredNode := &types.Node{
			Name:         node.Name,
			Path:         node.Path,
			RelativePath: node.RelativePath,
			IsDir:        node.IsDir,
			Annotation:   node.Annotation,
			Metadata:     node.Metadata,
			Children:     nil,
			Parent:       node.Parent,
		}
		return filteredNode, nil
	}
	
	return nil, nil
}

// CountTotalFiles counts the total number of files in a tree (for estimation)
func CountTotalFiles(node *types.Node) int {
	if node == nil {
		return 0
	}
	
	count := 0
	if !node.IsDir {
		count = 1
	}
	
	for _, child := range node.Children {
		count += CountTotalFiles(child)
	}
	
	return count
}