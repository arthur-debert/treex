package types

// Node represents a file or directory in the tree
type Node struct {
	Name       string      // Just the filename/dirname, e.g., "main.go"
	Path       string      // The unique, relative path from the tree root, e.g., "src/main.go"
	IsDir      bool        // Whether this is a directory
	Size       int64       // File size in bytes (0 for directories)
	Annotation *Annotation // Associated annotation if any
	Children   []*Node     // Child nodes (for directories)
	Parent     *Node       // Parent node (nil for root)
}
