package query

import (
	"github.com/adebert/treex/pkg/core/limits"
	"github.com/adebert/treex/pkg/core/types"
)

// FilterTreeWithMatches filters a tree and collects matching lines
func FilterTreeWithMatches(node *types.Node, collector *MatchCollector, limiter *limits.Limiter) (*types.Node, error) {
	// Check if we've exceeded the runtime limit
	if !limiter.CheckTimeout() {
		return nil, nil
	}
	
	// First, check if this node matches and get matches
	matches, matchLines, err := collector.MatchesWithDetail(node)
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
			
			filteredChild, err := FilterTreeWithMatches(child, collector, limiter)
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
				Matches:      matchLines,
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
			Matches:      matchLines,
		}
		return filteredNode, nil
	}
	
	return nil, nil
}