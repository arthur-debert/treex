package tui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

// StyledTreeRenderer handles rendering file trees with beautiful Lip Gloss styling
type StyledTreeRenderer struct {
	writer          io.Writer
	showAnnotations bool
	styles          *TreeStyles
	styleRenderer   *StyleRenderer // Optional renderer-based styles
	terminalWidth   int
	tabstop         int // Calculated tabstop for annotation alignment
	safeMode        bool // Use safe width calculations for problematic terminals
}

// NewStyledTreeRenderer creates a new styled tree renderer
func NewStyledTreeRenderer(writer io.Writer, showAnnotations bool) *StyledTreeRenderer {
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          NewTreeStyles(),
		terminalWidth:   80, // Default width, can be detected
		tabstop:         0,  // Will be calculated during rendering
		safeMode:        isProblematicTerminal(),
	}
}

// NewStyledTreeRendererWithRenderer creates a new styled tree renderer with a specific lipgloss renderer
func NewStyledTreeRendererWithRenderer(writer io.Writer, showAnnotations bool) *StyledTreeRenderer {
	styleRenderer := NewStyleRenderer(writer)
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          styleRenderer.Styles(),
		styleRenderer:   styleRenderer,
		terminalWidth:   80, // Default width, can be detected
		tabstop:         0,  // Will be calculated during rendering
		safeMode:        isProblematicTerminal(),
	}
}

// isProblematicTerminal detects terminals that might have issues with lipgloss.Width
func isProblematicTerminal() bool {
	// Check for explicit safe mode override
	if os.Getenv("TREEX_SAFE_MODE") == "1" || os.Getenv("TREEX_SAFE_MODE") == "true" {
		return true
	}
	
	termProgram := os.Getenv("TERM_PROGRAM")
	term := os.Getenv("TERM")
	
	// Known problematic terminals
	problematicTerms := []string{
		"ghostty",
		"GHOSTTY",
	}
	
	// Check TERM_PROGRAM (most reliable for Ghostty)
	for _, problematic := range problematicTerms {
		if strings.Contains(strings.ToLower(termProgram), strings.ToLower(problematic)) {
			return true
		}
	}
	
	// Check TERM variable as fallback
	for _, problematic := range problematicTerms {
		if strings.Contains(strings.ToLower(term), strings.ToLower(problematic)) {
			return true
		}
	}
	
	// Additional heuristics for Ghostty detection
	// Ghostty often sets TERM_PROGRAM to "ghostty"
	if strings.ToLower(termProgram) == "ghostty" {
		return true
	}
	
	return false
}

// safeWidth calculates the visual width of text with a timeout fallback
func (r *StyledTreeRenderer) safeWidth(text string) int {
	if r.safeMode {
		// In safe mode, just use string length (ignoring ANSI codes is better than hanging)
		return len(stripANSI(text))
	}
	
	// Use a channel to implement timeout
	result := make(chan int, 1)
	go func() {
		result <- lipgloss.Width(text)
	}()
	
	select {
	case width := <-result:
		return width
	case <-time.After(100 * time.Millisecond):
		// Timeout - fall back to safe calculation
		r.safeMode = true // Switch to safe mode for future calls
		return len(stripANSI(text))
	}
}

// stripANSI removes ANSI escape sequences from text for safe width calculation
func stripANSI(text string) string {
	// Simple ANSI escape sequence removal
	result := ""
	inEscape := false
	
	for i, char := range text {
		if char == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			inEscape = true
			continue
		}
		if inEscape {
			if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
				inEscape = false
			}
			continue
		}
		result += string(char)
	}
	
	return result
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

// WithSafeMode sets the safe mode for the renderer
func (r *StyledTreeRenderer) WithSafeMode(safe bool) *StyledTreeRenderer {
	r.safeMode = safe
	return r
}

// Render renders the tree starting from the root node with beautiful styling
func (r *StyledTreeRenderer) Render(root *tree.Node) error {
	// Calculate tabstop for annotation alignment
	if r.showAnnotations {
		r.calculateTabstop(root)
	}
	
	// Render the root directory name with styling
	rootName := r.styles.RootPath.Render(root.Name)
	if _, err := fmt.Fprintf(r.writer, "%s\n", rootName); err != nil {
		return err
	}
	
	// Render children
	return r.renderChildren(root.Children, "")
}

// calculateTabstop determines the tabstop position for annotation alignment
func (r *StyledTreeRenderer) calculateTabstop(root *tree.Node) {
	longestPath := r.findLongestRenderedPath(root, "")
	// Use the larger of longest path length or 40 as specified
	if longestPath < 40 {
		r.tabstop = 40
	} else {
		r.tabstop = longestPath
	}
}

// findLongestRenderedPath recursively finds the longest rendered path in the tree
func (r *StyledTreeRenderer) findLongestRenderedPath(node *tree.Node, prefix string) int {
	maxLength := 0
	
	// Check current node if it's not the root
	if prefix != "" {
		// Calculate the visual width of the rendered path
		var styledName string
		if node.Annotation != nil {
			styledName = r.styles.AnnotatedPath.Render(node.Name)
		} else {
			styledName = r.styles.UnannotatedPath.Render(node.Name)
		}
		
		currentPath := prefix + styledName
		currentLength := r.safeWidth(currentPath)
		if currentLength > maxLength {
			maxLength = currentLength
		}
	}
	
	// Check children
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1
		
		var connector string
		if isLast {
			connector = "└── "
		} else {
			connector = "├── "
		}
		
		styledConnector := r.styles.TreeLines.Render(connector)
		childPrefix := prefix + styledConnector
		
		childLength := r.findLongestRenderedPath(child, childPrefix)
		if childLength > maxLength {
			maxLength = childLength
		}
	}
	
	return maxLength
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
			styledVerticalConnector := r.styles.TreeLines.Render("│   ")
			nextPrefix = prefix + styledVerticalConnector
		}
		
		// Style the connector
		styledConnector := r.styles.TreeLines.Render(connector)
		
		// Render the current node
		// Pass the continuation prefix for multi-line content (prefix without the connector)
		var continuationPrefix string
		if !isLast {
			// For non-last children, add the vertical line
			continuationPrefix = prefix + r.styles.TreeLines.Render("│   ")
		} else {
			// For last children, add spaces
			continuationPrefix = prefix + "    "
		}
		if err := r.renderNode(child, prefix+styledConnector, continuationPrefix); err != nil {
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
func (r *StyledTreeRenderer) renderNode(node *tree.Node, prefix, continuationPrefix string) error {
	// Style the node name based on whether it has annotations
	var styledName string
	if node.Annotation != nil {
		// Paths with annotations: use regular full color
		styledName = r.styles.AnnotatedPath.Render(node.Name)
	} else {
		// Paths with no annotations: use subdued gray
		styledName = r.styles.UnannotatedPath.Render(node.Name)
	}
	
	// Build the main line with path
	pathLine := prefix + styledName
	
	// Add inline annotation if present and enabled
	if r.showAnnotations && node.Annotation != nil {
		inlineAnnotation := r.formatInlineAnnotation(node.Annotation)
		if inlineAnnotation != "" {
					// Calculate current path width and pad to tabstop
		currentWidth := r.safeWidth(pathLine)
		var padding string
		if currentWidth < r.tabstop {
			padding = strings.Repeat(" ", r.tabstop-currentWidth)
			} else {
				// If path is longer than tabstop, use minimum spacing
				padding = "  "
			}
			
			pathLine += padding + inlineAnnotation
		}
	}
	
	// Write the main line
	if _, err := fmt.Fprintf(r.writer, "%s\n", pathLine); err != nil {
		return err
	}
	
	// Render multi-line annotation if present
	if r.showAnnotations && node.Annotation != nil {
		// Use the continuation prefix that was passed in
		if err := r.renderMultiLineAnnotation(node.Annotation, continuationPrefix); err != nil {
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
		// Check if this is a single-line annotation (Title == Description)
		// If so, don't duplicate the content that's already shown inline
		if annotation.Title == annotation.Description {
			// For single-line annotations, don't show any additional lines
			// since the content is already displayed inline
			startIndex = len(lines) // This will skip all lines
		} else {
			// If we used the title as inline annotation, we need to skip
			// the first line of the description if it's the same as the title
			// (which is common when the parser includes the title in the description)
			if len(lines) > 0 && strings.TrimSpace(lines[0]) == annotation.Title {
				startIndex = 1
			} else {
				startIndex = 0
			}
			
			// Skip empty lines at the beginning
			for startIndex < len(lines) && strings.TrimSpace(lines[startIndex]) == "" {
				startIndex++
			}
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
		
		// Create proper indentation that maintains tree structure
		// The basePrefix already contains the tree connectors (│) we need
		// We need to add spacing to align with the annotation column
		treeIndent := basePrefix
		// Add spacing to reach the tabstop position
		spacingNeeded := r.tabstop - r.safeWidth(basePrefix)
		if spacingNeeded > 0 {
			treeIndent += strings.Repeat(" ", spacingNeeded)
		}
		
		// Apply container styling
		containerStyle := r.styles.AnnotationContainer
		
		// Split into lines and render each with proper tree-aware indentation
		contentLines := strings.Split(styledContent, "\n")
		for _, line := range contentLines {
			if strings.TrimSpace(line) != "" {
				styledLine := containerStyle.Render(line)
				if _, err := fmt.Fprintf(r.writer, "%s%s\n", treeIndent, styledLine); err != nil {
					return err
				}
			}
		}
	}
	
	return nil
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

// RenderStyledTreeToStringWithSafeMode renders a styled tree to a string with explicit safe mode control
func RenderStyledTreeToStringWithSafeMode(root *tree.Node, showAnnotations bool, safeMode bool) (string, error) {
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, showAnnotations)
	if safeMode {
		renderer = renderer.WithSafeMode(true)
	}
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

// RenderStyledTreeWithSafeMode renders a tree with beautiful styling and explicit safe mode control
func RenderStyledTreeWithSafeMode(writer io.Writer, root *tree.Node, showAnnotations bool, safeMode bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations)
	if safeMode {
		renderer = renderer.WithSafeMode(true)
	}
	return renderer.Render(root)
}

// RenderMinimalStyledTree is a convenience function that renders a tree with minimal styling for limited color environments
func RenderMinimalStyledTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(NewMinimalTreeStyles())
	return renderer.Render(root)
}

// RenderMinimalStyledTreeWithSafeMode renders a tree with minimal styling and explicit safe mode control
func RenderMinimalStyledTreeWithSafeMode(writer io.Writer, root *tree.Node, showAnnotations bool, safeMode bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(NewMinimalTreeStyles())
	if safeMode {
		renderer = renderer.WithSafeMode(true)
	}
	return renderer.Render(root)
}

// RenderPlainTree renders a tree without colors for plain text output
func RenderPlainTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(NewNoColorTreeStyles())
	return renderer.Render(root)
}

// RenderPlainTreeToString renders a tree without colors to a string
func RenderPlainTreeToString(root *tree.Node, showAnnotations bool) (string, error) {
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, showAnnotations).
		WithStyles(NewNoColorTreeStyles())
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

// RenderMinimalStyledTreeToString renders a tree with minimal styling to a string
func RenderMinimalStyledTreeToString(root *tree.Node, showAnnotations bool, safeMode bool) (string, error) {
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, showAnnotations).
		WithStyles(NewMinimalTreeStyles())
	if safeMode {
		renderer = renderer.WithSafeMode(true)
	}
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
} 