package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// ColorTheme represents a color theme (dark or light)
type ColorTheme struct {
	// Tree structure colors
	TreeConnectorColor lipgloss.Color
	DirectoryColor     lipgloss.Color
	FileColor          lipgloss.Color
	
	// Annotation colors
	AnnotationTitleColor       lipgloss.Color
	AnnotationDescriptionColor lipgloss.Color
	AnnotationBorderColor      lipgloss.Color
	
	// Accent colors
	HighlightColor lipgloss.Color
	MutedColor     lipgloss.Color
}

// Dark theme colors - optimized for dark backgrounds
var DarkTheme = ColorTheme{
	// Tree structure colors
	TreeConnectorColor: lipgloss.Color("#6C7086"), // Subtle gray for tree lines
	DirectoryColor:     lipgloss.Color("#89B4FA"), // Blue for directories
	FileColor:          lipgloss.Color("#CDD6F4"), // Light gray for files
	
	// Annotation colors
	AnnotationTitleColor:       lipgloss.Color("#F9E2AF"), // Yellow for titles
	AnnotationDescriptionColor: lipgloss.Color("#A6E3A1"), // Green for descriptions
	AnnotationBorderColor:      lipgloss.Color("#585B70"), // Darker gray for borders
	
	// Accent colors
	HighlightColor: lipgloss.Color("#F38BA8"), // Pink for highlights
	MutedColor:     lipgloss.Color("#6C7086"), // Muted gray
}

// Light theme colors - optimized for light backgrounds
var LightTheme = ColorTheme{
	// Tree structure colors
	TreeConnectorColor: lipgloss.Color("#6B6B6B"), // Medium gray for tree lines
	DirectoryColor:     lipgloss.Color("#0969DA"), // Blue for directories
	FileColor:          lipgloss.Color("#1F2328"), // Dark gray for files
	
	// Annotation colors
	AnnotationTitleColor:       lipgloss.Color("#9A6700"), // Dark yellow for titles
	AnnotationDescriptionColor: lipgloss.Color("#1A7F37"), // Dark green for descriptions
	AnnotationBorderColor:      lipgloss.Color("#D1D9E0"), // Light gray for borders
	
	// Accent colors
	HighlightColor: lipgloss.Color("#CF222E"), // Red for highlights
	MutedColor:     lipgloss.Color("#656D76"), // Muted gray
}

// Current active theme - defaults to dark
var activeTheme = DarkTheme

// SetTheme sets the active color theme
func SetTheme(dark bool) {
	if dark {
		activeTheme = DarkTheme
	} else {
		activeTheme = LightTheme
	}
}

// GetTheme returns the current active theme
func GetTheme() ColorTheme {
	return activeTheme
}

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

// NewTreeStyles creates a new set of tree styles using the active theme
func NewTreeStyles() *TreeStyles {
	theme := GetTheme()
	return &TreeStyles{
		// Tree structure styles
		TreeLines: lipgloss.NewStyle().
			Foreground(theme.TreeConnectorColor).
			Bold(false),

		RootPath: lipgloss.NewStyle().
			Foreground(theme.DirectoryColor).
			Bold(true),

		AnnotatedPath: lipgloss.NewStyle().
			Foreground(theme.FileColor),

		UnannotatedPath: lipgloss.NewStyle().
			Foreground(theme.TreeConnectorColor),

		// Annotation styles
		AnnotationText: lipgloss.NewStyle().
			Foreground(theme.DirectoryColor).
			Bold(true),

		AnnotationContainer: lipgloss.NewStyle().
			PaddingLeft(1), // Just a small padding, no border since we maintain tree structure

		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().
			Foreground(theme.MutedColor).
			SetString("  "),

		MultiLineIndent: lipgloss.NewStyle().
			Foreground(theme.AnnotationBorderColor).
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
