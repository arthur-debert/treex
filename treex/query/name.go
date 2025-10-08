package query

import (
	"github.com/bmatcuk/doublestar/v4"
	"treex/treex/types"
)

// NameQuery filters nodes based on their name (file or directory name only) using doublestar glob patterns
type NameQuery struct {
	pattern string
}

// NewNameQuery creates a new name query with the given doublestar glob pattern
func NewNameQuery(pattern string) *NameQuery {
	return &NameQuery{
		pattern: pattern,
	}
}

// Match returns true if the node's name matches the glob pattern
func (q *NameQuery) Match(node *types.Node) bool {
	if node == nil {
		return false
	}

	// Use the node's name (file or directory name without path)
	nodeName := node.Name

	// Use doublestar for gitignore-compatible glob matching
	matched, err := doublestar.Match(q.pattern, nodeName)
	if err != nil {
		// If pattern is invalid, don't match anything
		return false
	}

	return matched
}

// Name returns the name of this query type
func (q *NameQuery) Name() string {
	return "name:" + q.pattern
}