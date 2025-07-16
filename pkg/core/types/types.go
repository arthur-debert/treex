package types

import (
	"fmt"
	"sort"
)

// Annotation represents a single file/directory annotation
type Annotation struct {
	Path  string
	Notes string // Complete notes for the file/directory
}

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

// SortChildren sorts the children of a node alphabetically by name.
func (n *Node) SortChildren() {
	if n.Children == nil {
		return
	}
	sort.Slice(n.Children, func(i, j int) bool {
		return n.Children[i].Name < n.Children[j].Name
	})
}

// ViewMode represents the different ways to display the tree
type ViewMode string

const (
	// ViewModeAll shows all paths including unannotated ones
	ViewModeAll ViewMode = "all"
	// ViewModeAnnotated shows only paths with annotations
	ViewModeAnnotated ViewMode = "annotated"
	// ViewModeMix shows annotations plus contextual unannotated paths
	ViewModeMix ViewMode = "mix"
)

// ViewOptions contains options for controlling tree view behavior
type ViewOptions struct {
	Mode ViewMode
}

// DefaultViewOptions returns the default view options
func DefaultViewOptions() ViewOptions {
	return ViewOptions{
		Mode: ViewModeMix,
	}
}

// ParseViewMode converts a string to ViewMode
func ParseViewMode(mode string) (ViewMode, error) {
	switch mode {
	case "all":
		return ViewModeAll, nil
	case "annotated":
		return ViewModeAnnotated, nil
	case "mix":
		return ViewModeMix, nil
	default:
		return "", fmt.Errorf("invalid view mode: %s (must be 'all', 'annotated', or 'mix')", mode)
	}
}
