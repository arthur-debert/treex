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
	AnnotationText      lipgloss.Style // For annotation content
	AnnotationContainer lipgloss.Style // For annotation formatting/borders

	// Layout styles
	AnnotationSeparator lipgloss.Style
	MultiLineIndent     lipgloss.Style
}

// NewTreeStyles creates a new set of tree styles with adaptive colors
func NewTreeStyles() *TreeStyles {
	return &TreeStyles{
		// Tree structure styles
		TreeLines: lipgloss.NewStyle().
			Foreground(TreeConnectorColor).
			Bold(false),

		RootPath: lipgloss.NewStyle().
			Foreground(DirectoryColor).
			Bold(true),

		AnnotatedPath: lipgloss.NewStyle().
			Foreground(FileColor),

		UnannotatedPath: lipgloss.NewStyle().
			Foreground(TreeConnectorColor),

		// Annotation styles
		AnnotationText: lipgloss.NewStyle().
			Foreground(DirectoryColor).
			Bold(true),

		AnnotationContainer: lipgloss.NewStyle().
			PaddingLeft(1), // Just a small padding, no border since we maintain tree structure

		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().
			Foreground(MutedColor).
			SetString("  "),

		MultiLineIndent: lipgloss.NewStyle().
			Foreground(AnnotationBorderColor).
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
		AnnotationText:      lipgloss.NewStyle().Bold(true),
		AnnotationContainer: lipgloss.NewStyle(),

		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().SetString("  "),
		MultiLineIndent:     lipgloss.NewStyle(),
	}
}
