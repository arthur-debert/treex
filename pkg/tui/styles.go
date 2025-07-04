package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// TreeStyles contains all the styling for our tree renderer
type TreeStyles struct {
	// Tree structure styles
	TreeLines       lipgloss.Style // For tree connectors (├── └──)
	RootPath        lipgloss.Style // For the root directory name
	AnnotatedPath   lipgloss.Style // For paths that have annotations
	UnannotatedPath lipgloss.Style // For paths without annotations

	// Annotation styles
	AnnotationText        lipgloss.Style // For annotation content
	AnnotationTitle       lipgloss.Style // For annotation titles (inline)
	AnnotationDescription lipgloss.Style // For annotation descriptions (multi-line)
	AnnotationContainer   lipgloss.Style // For annotation formatting/borders

	// Layout styles
	AnnotationSeparator lipgloss.Style
	MultiLineIndent     lipgloss.Style
}

// NewTreeStyles creates a new set of tree styles with adaptive colors
func NewTreeStyles() *TreeStyles {
	return &TreeStyles{
		// Tree structure styles
		TreeLines: lipgloss.NewStyle().
			Foreground(Colors.TreeConnector).
			Faint(true),

		RootPath: lipgloss.NewStyle().
			Foreground(Colors.TreeDirectory).
			Bold(true),

		AnnotatedPath: lipgloss.NewStyle().
			Foreground(Colors.TreeFile).
			Bold(false),

		UnannotatedPath: lipgloss.NewStyle().
			Foreground(Colors.TreeConnector).
			Faint(true),

		// Annotation styles
		AnnotationText: lipgloss.NewStyle().
			Foreground(Colors.Primary).
			Bold(true),

		AnnotationTitle: lipgloss.NewStyle().
			Foreground(Colors.Warning).
			Bold(true),

		AnnotationDescription: lipgloss.NewStyle().
			Foreground(Colors.Success).
			Bold(false),

		AnnotationContainer: lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(Colors.Text),

		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().
			Foreground(Colors.TextMuted).
			Faint(true).
			SetString("  "),

		MultiLineIndent: lipgloss.NewStyle().
			Foreground(Colors.Border).
			Faint(true).
			PaddingLeft(1),
	}
}

// NewMinimalTreeStyles creates a minimal color scheme for environments with limited color support
func NewMinimalTreeStyles() *TreeStyles {
	return &TreeStyles{
		// Tree structure styles
		TreeLines: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")), // Dark gray

		RootPath: lipgloss.NewStyle().
			Bold(true),

		AnnotatedPath: lipgloss.NewStyle(),

		UnannotatedPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")), // Dark gray

		// Annotation styles
		AnnotationText: lipgloss.NewStyle().
			Bold(true),

		AnnotationTitle: lipgloss.NewStyle().
			Bold(true),

		AnnotationDescription: lipgloss.NewStyle(),

		AnnotationContainer: lipgloss.NewStyle().
			PaddingLeft(1),

		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().
			SetString("  "),

		MultiLineIndent: lipgloss.NewStyle().
			PaddingLeft(1),
	}
}

// NewNoColorTreeStyles creates styles without any colors for plain text output
func NewNoColorTreeStyles() *TreeStyles {
	return &TreeStyles{
		// Tree structure styles
		TreeLines:       lipgloss.NewStyle(),
		RootPath:        lipgloss.NewStyle().Bold(true),
		AnnotatedPath:   lipgloss.NewStyle(),
		UnannotatedPath: lipgloss.NewStyle(),

		// Annotation styles
		AnnotationText:        lipgloss.NewStyle().Bold(true),
		AnnotationTitle:       lipgloss.NewStyle().Bold(true),
		AnnotationDescription: lipgloss.NewStyle(),
		AnnotationContainer:   lipgloss.NewStyle(),

		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().SetString("  "),
		MultiLineIndent:     lipgloss.NewStyle(),
	}
}
