package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/adebert/treex/internal/info"
	"github.com/adebert/treex/internal/tree"
)

// StyledTreeRenderer handles rendering file trees with beautiful Lip Gloss styling
type StyledTreeRenderer struct {
	writer          io.Writer
	showAnnotations bool
	styles          *TreeStyles
	terminalWidth   int
}

// NewStyledTreeRenderer creates a new styled tree renderer
func NewStyledTreeRenderer(writer io.Writer, showAnnotations bool) *StyledTreeRenderer {
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          NewTreeStyles(),
		terminalWidth:   80, // Default width, can be detected
	}
}

// WithStyles sets custom styles for the renderer
func (r *StyledTreeRenderer) WithStyles(styles *TreeStyles) *StyledTreeRenderer {
	r.styles = styles
	return r
}

// WithTerminalWidth sets the terminal width for better formatting
func (r *StyledTreeRenderer) WithTerminalWidth(width int) *StyledTreeRenderer {
	r.terminalWidth = width
	return r
}

// Render renders the tree starting from the root node with beautiful styling
func (r *StyledTreeRenderer) Render(root *tree.Node) error {
	// Render the root directory name with styling
	rootName := r.styles.RootPath.Render(root.Name)
	fmt.Fprintf(r.writer, "%s\n", rootName)
	
	// Render children
	return r.renderChildren(root.Children, "")
}

// renderChildren renders a list of child nodes with proper tree formatting and styling
func (r *StyledTreeRenderer) renderChildren(children []*tree.Node, prefix string) error {
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
		
		// Style the connector
		styledConnector := r.styles.TreeLines.Render(connector)
		
		// Render the current node
		if err := r.renderNode(child, prefix+styledConnector, nextPrefix); err != nil {
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

// renderNode renders a single node with its annotation using beautiful styling
func (r *StyledTreeRenderer) renderNode(node *tree.Node, prefix, nextPrefix string) error {
	// Style the node name based on whether it has annotations
	var styledName string
	if node.Annotation != nil {
		// Paths with annotations: use regular full color
		styledName = r.styles.AnnotatedPath.Render(node.Name)
	} else {
		// Paths with no annotations: use subdued gray
		styledName = r.styles.UnannotatedPath.Render(node.Name)
	}
	
	// Build the main line
	mainLine := prefix + styledName
	
	// Add inline annotation if present and enabled
	if r.showAnnotations && node.Annotation != nil {
		inlineAnnotation := r.formatInlineAnnotation(node.Annotation)
		if inlineAnnotation != "" {
			separator := r.styles.AnnotationSeparator.Render("")
			mainLine += separator + inlineAnnotation
		}
	}
	
	// Write the main line
	fmt.Fprintf(r.writer, "%s\n", mainLine)
	
	// Render multi-line annotation if present
	if r.showAnnotations && node.Annotation != nil {
		if err := r.renderMultiLineAnnotation(node.Annotation, nextPrefix); err != nil {
			return err
		}
	}
	
	return nil
}

// formatInlineAnnotation formats the primary annotation text for inline display
func (r *StyledTreeRenderer) formatInlineAnnotation(annotation *info.Annotation) string {
	if annotation == nil {
		return ""
	}
	
	// If we have a title, use it as the primary annotation
	if annotation.Title != "" {
		return r.styles.AnnotationText.Render(annotation.Title)
	}
	
	// Otherwise, use the first line of the description
	lines := strings.Split(annotation.Description, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		firstLine := strings.TrimSpace(lines[0])
		return r.styles.AnnotationText.Render(firstLine)
	}
	
	return ""
}

// renderMultiLineAnnotation renders additional annotation lines with beautiful formatting
func (r *StyledTreeRenderer) renderMultiLineAnnotation(annotation *info.Annotation, basePrefix string) error {
	if annotation == nil {
		return nil
	}
	
	var additionalLines []string
	lines := strings.Split(annotation.Description, "\n")
	
	// Determine which lines to show as additional content
	startIndex := 1
	if annotation.Title != "" {
		// If we used the title as inline annotation, show all description lines
		startIndex = 0
		
		// But skip empty first lines
		for startIndex < len(lines) && strings.TrimSpace(lines[startIndex]) == "" {
			startIndex++
		}
	}
	
	// Collect non-empty additional lines
	for i := startIndex; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			additionalLines = append(additionalLines, line)
		}
	}
	
	// Render additional lines if we have any
	if len(additionalLines) > 0 {
		// Create the annotation block
		annotationContent := strings.Join(additionalLines, "\n")
		
		// Style the content with annotation text styling
		styledContent := r.styles.AnnotationText.Render(annotationContent)
		
		// Create indentation that matches the tree structure
		indent := r.createAnnotationIndent(basePrefix)
		
		// Apply container styling
		containerStyle := r.styles.AnnotationContainer.Copy()
		
		// Split into lines and render each with proper indentation
		contentLines := strings.Split(styledContent, "\n")
		for _, line := range contentLines {
			if strings.TrimSpace(line) != "" {
				styledLine := containerStyle.Render(line)
				fmt.Fprintf(r.writer, "%s%s\n", indent, styledLine)
			}
		}
	}
	
	return nil
}

// createAnnotationIndent creates proper indentation for annotation continuation lines
func (r *StyledTreeRenderer) createAnnotationIndent(basePrefix string) string {
	// Convert tree characters to spaces while preserving structure
	indent := ""
	
	// Count the visual width of the prefix to maintain alignment
	visualWidth := lipgloss.Width(basePrefix)
	
	// Create spacing that aligns with the tree structure
	for i := 0; i < visualWidth; i++ {
		indent += " "
	}
	
	return indent
}

// RenderStyledTree is a convenience function that renders a tree with beautiful styling
func RenderStyledTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations)
	return renderer.Render(root)
}

// RenderStyledTreeToString renders a styled tree to a string
func RenderStyledTreeToString(root *tree.Node, showAnnotations bool) (string, error) {
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, showAnnotations)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

// RenderMinimalStyledTree renders a tree with minimal styling for limited color environments
func RenderMinimalStyledTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(NewMinimalTreeStyles())
	return renderer.Render(root)
}

// RenderPlainTree renders a tree without colors for plain text output
func RenderPlainTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(NewNoColorTreeStyles())
	return renderer.Render(root)
} 