package rendering

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/tree"
	"github.com/adebert/treex/pkg/display/styles"
)

// StyledTreeRenderer handles rendering file trees with beautiful Lip Gloss styling
type StyledTreeRenderer struct {
	writer          io.Writer
	showAnnotations bool
	styles          *styles.TreeStyles
	styleRenderer   *StyleRenderer // Optional renderer-based styles
	terminalWidth   int
	tabstop         int  // Calculated tabstop for annotation alignment
	safeMode        bool // Use safe width calculations for problematic terminals
	extraSpacing    bool // Add extra vertical spacing between annotated items
}

// NewStyledTreeRenderer creates a new styled tree renderer
func NewStyledTreeRenderer(writer io.Writer, showAnnotations bool) *StyledTreeRenderer {
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          styles.NewTreeStyles(),
		terminalWidth:   80, // Default width, can be detected
		tabstop:         0,  // Will be calculated during rendering
		safeMode:        isProblematicTerminal(),
		extraSpacing:    true, // Default to true for better readability
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
		extraSpacing:    true, // Default to true for better readability
	}
}

// NewStyledTreeRendererWithAutoTheme creates a new styled tree renderer with automatic theme detection
func NewStyledTreeRendererWithAutoTheme(writer io.Writer, showAnnotations bool, verbose bool) *StyledTreeRenderer {
	styleRenderer := NewStyleRendererWithAutoTheme(writer, verbose)
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          styleRenderer.Styles(),
		styleRenderer:   styleRenderer,
		terminalWidth:   80, // Default width, can be detected
		tabstop:         0,  // Will be calculated during rendering
		safeMode:        isProblematicTerminal(),
		extraSpacing:    true, // Default to true for better readability
	}
}

// NewMinimalStyledTreeRenderer creates a styled tree renderer with minimal color support
func NewMinimalStyledTreeRenderer(writer io.Writer, showAnnotations bool) *StyledTreeRenderer {
	styleRenderer := NewMinimalStyleRenderer(writer)
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          styleRenderer.Styles(),
		styleRenderer:   styleRenderer,
		terminalWidth:   80,
		tabstop:         0,
		safeMode:        isProblematicTerminal(),
		extraSpacing:    true, // Default to true for better readability
	}
}

// NewNoColorStyledTreeRenderer creates a styled tree renderer without any colors
func NewNoColorStyledTreeRenderer(writer io.Writer, showAnnotations bool) *StyledTreeRenderer {
	styleRenderer := NewNoColorStyleRenderer(writer)
	return &StyledTreeRenderer{
		writer:          writer,
		showAnnotations: showAnnotations,
		styles:          styleRenderer.Styles(),
		styleRenderer:   styleRenderer,
		terminalWidth:   80,
		tabstop:         0,
		safeMode:        isProblematicTerminal(),
		extraSpacing:    true, // Default to true for better readability
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
func (r *StyledTreeRenderer) WithStyles(styles *styles.TreeStyles) *StyledTreeRenderer {
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

// WithExtraSpacing enables or disables extra vertical spacing between annotated items
func (r *StyledTreeRenderer) WithExtraSpacing(enabled bool) *StyledTreeRenderer {
	r.extraSpacing = enabled
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
		// Check if we should add spacing after this node
		// Add spacing if it has annotations AND either has children or is not the last sibling
		shouldAddSpacing := child.Annotation != nil && (!isLast || (child.IsDir && len(child.Children) > 0))
		
		if err := r.renderNode(child, prefix+styledConnector, continuationPrefix, shouldAddSpacing); err != nil {
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
func (r *StyledTreeRenderer) renderNode(node *tree.Node, prefix, continuationPrefix string, shouldAddSpacing bool) error {
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
	
	
	// Add extra spacing after items with annotations, maintaining tree structure
	if shouldAddSpacing && r.showAnnotations && r.extraSpacing {
		// Print the continuation prefix to maintain tree lines
		if _, err := fmt.Fprintln(r.writer, continuationPrefix); err != nil {
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
	
	// Use Notes field
	notes := annotation.Notes
	
	if notes != "" {
		// Check if the text contains markdown syntax
		hasMarkdown := strings.Contains(notes, "**") || strings.Contains(notes, "*") || strings.Contains(notes, "`")
		
		// Check if we're in no-color mode by examining the annotation text style
		isNoColor := r.styleRenderer != nil && r.styleRenderer.renderer != nil && 
			r.styleRenderer.renderer.ColorProfile() == termenv.Ascii
		
		// Only use Glamour if we have markdown and we're not in no-color mode
		if hasMarkdown && !isNoColor {
			// Create a glamour renderer with appropriate style
			// Use "auto" to automatically detect dark/light mode
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(0), // Disable word wrap for inline display
			)
			if err != nil {
				// Fallback to plain text with lipgloss styling if glamour fails
				return r.styles.AnnotationText.Render(notes)
			}
			
			// Render the markdown
			rendered, err := renderer.Render(notes)
			if err != nil {
				// Fallback to plain text with lipgloss styling if rendering fails
				return r.styles.AnnotationText.Render(notes)
			}
			
			// Glamour adds newlines, so trim them for inline display
			rendered = strings.TrimSpace(rendered)
			
			// Return the Glamour-rendered text directly to preserve markdown formatting
			return rendered
		}
		
		// For plain text or no-color mode, use lipgloss styling
		return r.styles.AnnotationText.Render(notes)
	}
	
	return ""
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

// RenderStyledTreeWithOptions renders a tree with beautiful styling and configurable options
func RenderStyledTreeWithOptions(writer io.Writer, root *tree.Node, showAnnotations bool, safeMode bool, extraSpacing bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithSafeMode(safeMode).
		WithExtraSpacing(extraSpacing)
	return renderer.Render(root)
}

// RenderMinimalStyledTree is a convenience function that renders a tree with minimal styling for limited color environments
func RenderMinimalStyledTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(styles.NewMinimalTreeStyles())
	return renderer.Render(root)
}

// RenderMinimalStyledTreeWithSafeMode renders a tree with minimal styling and explicit safe mode control
func RenderMinimalStyledTreeWithSafeMode(writer io.Writer, root *tree.Node, showAnnotations bool, safeMode bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(styles.NewMinimalTreeStyles())
	if safeMode {
		renderer = renderer.WithSafeMode(true)
	}
	return renderer.Render(root)
}

// RenderPlainTree renders a tree without colors for plain text output
func RenderPlainTree(writer io.Writer, root *tree.Node, showAnnotations bool) error {
	renderer := NewStyledTreeRenderer(writer, showAnnotations).
		WithStyles(styles.NewNoColorTreeStyles())
	return renderer.Render(root)
}

// RenderPlainTreeToString renders a tree without colors to a string
func RenderPlainTreeToString(root *tree.Node, showAnnotations bool) (string, error) {
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, showAnnotations).
		WithStyles(styles.NewNoColorTreeStyles())
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

// RenderMinimalStyledTreeToString renders a tree with minimal styling to a string
func RenderMinimalStyledTreeToString(root *tree.Node, showAnnotations bool, safeMode bool) (string, error) {
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, showAnnotations).
		WithStyles(styles.NewMinimalTreeStyles())
	if safeMode {
		renderer = renderer.WithSafeMode(true)
	}
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
} 