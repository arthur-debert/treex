package types

// Node represents a file or directory in the tree
type Node struct {
	Name         string      // Just the filename/dirname
	Path         string      // Full path from root
	RelativePath string      // Path relative to the tree root
	IsDir        bool        // Whether this is a directory
	Annotation   *Annotation // Associated annotation if any
	Children     []*Node     // Child nodes (for directories)
	Parent       *Node       // Parent node (nil for root)
}

