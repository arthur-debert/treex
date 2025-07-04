package tree

import "fmt"

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