package rendering

import (
	"fmt"
	"io"
	"strings"

	"github.com/adebert/treex/pkg/core/tree"
	"github.com/adebert/treex/pkg/core/info"
)

// TreeRenderer handles rendering file trees with annotations
type TreeRenderer struct {
	writer io.Writer
	showAnnotations bool
}

// NewTreeRenderer creates a new tree renderer
func NewTreeRenderer(writer io.Writer, showAnnotations bool) *TreeRenderer {
	return &TreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
	}
}

// Render renders the tree starting from the root node
func (r *TreeRenderer) Render(root *tree.Node) error {
	// Print the root directory name
	if _, err := fmt.Fprintf(r.writer, "%s\n", root.Name); err != nil {
		return err
	}
	
	// Render children
	return r.renderChildren(root.Children, "")
}

// renderChildren renders a list of child nodes with proper tree formatting
func (r *TreeRenderer) renderChildren(children []*tree.Node, prefix string) error {
	for i, child := range children {
		isLast := i == len(children)-1
		
		// Determine the connector and continuation prefix for multi-line content
		var connector, continuationPrefix, nextPrefix string
		if isLast {
			connector = "└── "
			continuationPrefix = prefix + "    " // No more siblings, use spaces
			nextPrefix = prefix + "    "
		} else {
			connector = "├── "
			continuationPrefix = prefix + "│   " // Has siblings, maintain connector
			nextPrefix = prefix + "│   "
		}
		
		// Render the current node with proper prefixes
		if err := r.renderNodeWithPrefix(child, prefix, connector, continuationPrefix); err != nil {
			return err
		}
		
		// Recursively render children if this is a directory
		if child.IsDir && len(child.Children) > 0 {
			if err := r.renderChildren(child.Children, nextPrefix); err != nil {
				return err
			}
		}
	}
	
	return nil
}


// renderNodeWithPrefix renders a single node with proper prefix handling for multi-line content
func (r *TreeRenderer) renderNodeWithPrefix(node *tree.Node, treePrefix, connector, continuationPrefix string) error {
	// Start with the full prefix (tree prefix + connector) and node name
	line := treePrefix + connector + node.Name
	
	// Add annotation if present and enabled
	if r.showAnnotations && node.Annotation != nil {
		annotation := r.formatAnnotation(node.Annotation)
		if annotation != "" {
			// Calculate padding to align annotations
			padding := r.calculatePadding(len(line))
			line += padding + annotation
		}
	}
	
	// Write the line
	if _, err := fmt.Fprintf(r.writer, "%s\n", line); err != nil {
		return err
	}
	
	
	return nil
}

// formatAnnotation formats an annotation for display
func (r *TreeRenderer) formatAnnotation(annotation *info.Annotation) string {
	if annotation == nil {
		return ""
	}
	
	// Use Notes if available, otherwise Description for backwards compatibility
	if annotation.Notes != "" {
		return annotation.Notes
	}
	
	// Fall back to Description (already single line after parser)
	return annotation.Description
}


// calculatePadding calculates padding to align annotations
func (r *TreeRenderer) calculatePadding(lineLength int) string {
	// Target column for annotations (adjust as needed)
	targetColumn := 40
	
	if lineLength >= targetColumn {
		return "  " // Minimum spacing
	}
	
	return strings.Repeat(" ", targetColumn-lineLength)
}

// RenderTree is a convenience function that renders a tree to a writer
func RenderTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewTreeRenderer(writer, showAnnotations)
	return renderer.Render(root)
}

// RenderTreeToString renders a tree to a string
func RenderTreeToString(root *tree.Node, showAnnotations bool) (string, error) {
	var builder strings.Builder
	renderer := NewTreeRenderer(&builder, showAnnotations)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
} 