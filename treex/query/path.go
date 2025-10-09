package query

import (
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"treex/treex/types"
)

// PathQuery filters nodes based on their full path using doublestar glob patterns
type PathQuery struct {
	pattern string
}

// NewPathQuery creates a new path query with the given doublestar glob pattern
func NewPathQuery(pattern string) *PathQuery {
	return &PathQuery{
		pattern: pattern,
	}
}

// Match returns true if the node's path matches the glob pattern
func (q *PathQuery) Match(node *types.Node) bool {
	if node == nil {
		return false
	}

	// Normalize the path for consistent matching
	nodePath := filepath.ToSlash(node.Path)

	// Use doublestar for gitignore-compatible glob matching
	matched, err := doublestar.Match(q.pattern, nodePath)
	if err != nil {
		// If pattern is invalid, don't match anything
		return false
	}

	return matched
}

// Name returns the name of this query type
func (q *PathQuery) Name() string {
	return "path:" + q.pattern
}
