// Package query provides Phase 5: User Query processing for filtering tree nodes
// See docs/dev/architecture.txt - Phase 5: User Queries
package query

import (
	"treex/treex/types"
)

// Query represents a user query that can filter tree nodes
type Query interface {
	// Match returns true if the node matches this query
	Match(node *types.Node) bool

	// Name returns a human-readable name for this query
	Name() string
}

// Processor processes queries against a tree structure
type Processor struct {
	queries []Query
}

// NewProcessor creates a new query processor
func NewProcessor() *Processor {
	return &Processor{
		queries: make([]Query, 0),
	}
}

// AddQuery adds a query to the processor
func (p *Processor) AddQuery(query Query) {
	p.queries = append(p.queries, query)
}

// Process filters the tree by applying all queries
// Returns a new tree with only nodes that match all queries
func (p *Processor) Process(root *types.Node) *types.Node {
	if root == nil {
		return nil
	}

	// If no queries, return the original tree
	if len(p.queries) == 0 {
		return root
	}

	return p.filterNode(root)
}

// filterNode recursively filters a node and its children
func (p *Processor) filterNode(node *types.Node) *types.Node {
	if node == nil {
		return nil
	}

	// Check if current node matches all queries
	matches := true
	for _, query := range p.queries {
		if !query.Match(node) {
			matches = false
			break
		}
	}

	// Process children
	var filteredChildren []*types.Node
	for _, child := range node.Children {
		if filteredChild := p.filterNode(child); filteredChild != nil {
			filteredChildren = append(filteredChildren, filteredChild)
		}
	}

	// If current node matches or has matching children, include it
	if matches || len(filteredChildren) > 0 {
		// Create a new node with filtered children
		filteredNode := &types.Node{
			Name:       node.Name,
			Path:       node.Path,
			IsDir:      node.IsDir,
			Size:       node.Size,
			Annotation: node.Annotation,
			Children:   filteredChildren,
			Parent:     nil, // Will be set when building the tree
		}

		// Set parent relationships for children
		for _, child := range filteredChildren {
			child.Parent = filteredNode
		}

		return filteredNode
	}

	return nil
}

// QueryOptions contains configuration for query processing
type QueryOptions struct {
	// PathPattern is a doublestar glob pattern for matching full paths
	PathPattern string

	// NamePattern is a doublestar glob pattern for matching file/directory names only
	NamePattern string
}

// NewQueryOptions creates default query options
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

// Builder provides a fluent interface for building query processors
type Builder struct {
	processor *Processor
}

// NewBuilder creates a new query builder
func NewBuilder() *Builder {
	return &Builder{
		processor: NewProcessor(),
	}
}

// WithPathPattern adds a path pattern query
func (b *Builder) WithPathPattern(pattern string) *Builder {
	if pattern != "" {
		b.processor.AddQuery(NewPathQuery(pattern))
	}
	return b
}

// WithNamePattern adds a name pattern query
func (b *Builder) WithNamePattern(pattern string) *Builder {
	if pattern != "" {
		b.processor.AddQuery(NewNameQuery(pattern))
	}
	return b
}

// Build returns the configured processor
func (b *Builder) Build() *Processor {
	return b.processor
}