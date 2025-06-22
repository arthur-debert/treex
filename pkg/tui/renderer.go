package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/adebert/treex/pkg/tree"
	"github.com/adebert/treex/pkg/info"
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
		
		// Determine the connector and next prefix
		var connector, nextPrefix string
		if isLast {
			connector = "└── "
			nextPrefix = prefix + "    "
		} else {
			connector = "├── "
			nextPrefix = prefix + "│   "
		}
		
		// Render the current node
		if err := r.renderNode(child, prefix+connector); err != nil {
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

// renderNode renders a single node with its annotation
func (r *TreeRenderer) renderNode(node *tree.Node, prefix string) error {
	// Start with the prefix and node name
	line := prefix + node.Name
	
	// Add annotation if present and enabled
	if r.showAnnotations && node.Annotation != nil {
		annotation := r.formatAnnotation(node.Annotation)
		if annotation != "" {
			// Calculate padding to align annotations
			// We'll use a simple approach: add some spaces and then the annotation
			padding := r.calculatePadding(len(line))
			line += padding + annotation
		}
	}
	
	// Write the line
	if _, err := fmt.Fprintf(r.writer, "%s\n", line); err != nil {
		return err
	}
	
	// If we have a multi-line annotation, render the additional lines
	if r.showAnnotations && node.Annotation != nil {
		additionalLines := r.getAdditionalAnnotationLines(node.Annotation, prefix)
		for _, additionalLine := range additionalLines {
			if _, err := fmt.Fprintf(r.writer, "%s\n", additionalLine); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// formatAnnotation formats an annotation for display
func (r *TreeRenderer) formatAnnotation(annotation *info.Annotation) string {
	if annotation == nil {
		return ""
	}
	
	// If we have a title, use it as the primary annotation
	if annotation.Title != "" {
		return annotation.Title
	}
	
	// Otherwise, use the first line of the description
	lines := strings.Split(annotation.Description, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		return strings.TrimSpace(lines[0])
	}
	
	return ""
}

// getAdditionalAnnotationLines returns additional lines for multi-line annotations
func (r *TreeRenderer) getAdditionalAnnotationLines(annotation *info.Annotation, prefix string) []string {
	if annotation == nil {
		return nil
	}
	
	var additionalLines []string
	lines := strings.Split(annotation.Description, "\n")
	
	// If we used the title as the main annotation, show all description lines
	// If we used the first line of description, show the remaining lines
	startIndex := 1
	if annotation.Title != "" {
		startIndex = 0 // Show all description lines if we used title
	}
	
	// Create the indentation for additional lines
	// We need to match the tree structure indentation
	indent := r.createAnnotationIndent(prefix)
	
	for i := startIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			additionalLines = append(additionalLines, indent+line)
		}
	}
	
	return additionalLines
}

// createAnnotationIndent creates proper indentation for annotation continuation lines
func (r *TreeRenderer) createAnnotationIndent(prefix string) string {
	// Replace tree characters with spaces to maintain alignment
	indent := ""
	for _, char := range prefix {
		switch char {
		case '├', '└':
			indent += " "
		case '─':
			indent += " "
		case '│':
			indent += "│"
		default:
			indent += string(char)
		}
	}
	
	// Add some extra spaces to align with the annotation text
	return indent + strings.Repeat(" ", 20) // Adjust this value as needed
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