package types

// TreeOptions represents all configuration for tree building
type TreeOptions struct {
	// Root path to start tree from
	Root string

	// Tree display options
	Tree TreeDisplayOptions

	// Pattern-based filters
	Patterns PatternOptions

	// Search terms for path/name matching
	Search []string
}

// TreeDisplayOptions controls basic tree traversal
type TreeDisplayOptions struct {
	// Maximum depth to traverse (default: 3)
	MaxDepth int

	// Show only directories
	DirsOnly bool

	// Show hidden files/directories (starting with .)
	ShowHidden bool
}

// PatternOptions handles all pattern-based filtering
type PatternOptions struct {
	// User-supplied exclude patterns (can be multiple)
	Excludes []string

	// Path to ignore file (default: .gitignore)
	IgnoreFilePath string

	// Disable ignore file processing
	NoIgnoreFile bool
}

// DefaultTreeOptions returns options with sensible defaults
func DefaultTreeOptions() TreeOptions {
	return TreeOptions{
		Root: ".",
		Tree: TreeDisplayOptions{
			MaxDepth:   3,
			DirsOnly:   false,
			ShowHidden: false,
		},
		Patterns: PatternOptions{
			Excludes:       []string{},
			IgnoreFilePath: ".gitignore",
			NoIgnoreFile:   false,
		},
		Search: []string{},
	}
}
