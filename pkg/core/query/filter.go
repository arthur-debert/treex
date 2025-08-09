package query

import (
	"fmt"
	
	"github.com/adebert/treex/pkg/core/types"
)

// Matcher can evaluate if a node matches a query
type Matcher struct {
	registry *Registry
	query    *Query
}

// NewMatcher creates a new matcher with the given query
func NewMatcher(registry *Registry, query *Query) *Matcher {
	return &Matcher{
		registry: registry,
		query:    query,
	}
}

// Matches returns true if the node matches all filters in the query (AND logic)
func (m *Matcher) Matches(node *types.Node) (bool, error) {
	// Empty query matches everything
	if len(m.query.Filters) == 0 {
		return true, nil
	}
	
	// All filters must match (AND logic)
	for _, filter := range m.query.Filters {
		matches, err := m.matchesFilter(node, filter)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	
	return true, nil
}

// matchesFilter checks if a node matches a single filter
func (m *Matcher) matchesFilter(node *types.Node, filter Filter) (bool, error) {
	// Get the attribute
	attr, exists := m.registry.GetAttribute(filter.Attribute)
	if !exists {
		return false, fmt.Errorf("unknown attribute: %s", filter.Attribute)
	}
	
	// Get the operator
	op, exists := m.registry.GetOperator(filter.Operator)
	if !exists {
		return false, fmt.Errorf("unknown operator: %s", filter.Operator)
	}
	
	// Check if operator is valid for attribute type
	validForType := false
	for _, validType := range op.ValidTypes {
		if validType == attr.Type {
			validForType = true
			break
		}
	}
	if !validForType {
		typeName := "unknown"
		switch attr.Type {
		case StringType:
			typeName = "string"
		case NumericType:
			typeName = "numeric"
		case DateType:
			typeName = "date"
		case BoolType:
			typeName = "boolean"
		}
		return false, fmt.Errorf("operator %s is not valid for %s attributes", filter.Operator, typeName)
	}
	
	// Extract the value from the node
	nodeValue, err := attr.Extractor(node)
	if err != nil {
		// If we can't extract the value, consider it a non-match
		return false, nil
	}
	
	// Apply the comparator
	return op.Comparator(nodeValue, filter.Value)
}

// FilterTree recursively filters a tree based on the query
// It removes nodes that don't match and their non-matching ancestors
func FilterTree(node *types.Node, matcher *Matcher) (*types.Node, error) {
	// First, check if this node matches
	matches, err := matcher.Matches(node)
	if err != nil {
		return nil, err
	}
	
	// For directories, we need to check children first
	if node.IsDir && len(node.Children) > 0 {
		var filteredChildren []*types.Node
		childMatches := false
		
		for _, child := range node.Children {
			filteredChild, err := FilterTree(child, matcher)
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