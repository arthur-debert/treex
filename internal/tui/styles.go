package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette for our tree UI
var (
	// Tree structure colors
	TreeConnectorColor = lipgloss.Color("#6C7086") // Subtle gray for tree lines
	DirectoryColor     = lipgloss.Color("#89B4FA") // Blue for directories
	FileColor          = lipgloss.Color("#CDD6F4") // Light gray for files
	
	// Annotation colors
	AnnotationTitleColor       = lipgloss.Color("#F9E2AF") // Yellow for titles
	AnnotationDescriptionColor = lipgloss.Color("#A6E3A1") // Green for descriptions
	AnnotationBorderColor      = lipgloss.Color("#585B70") // Darker gray for borders
	
	// Accent colors
	HighlightColor = lipgloss.Color("#F38BA8") // Pink for highlights
	MutedColor     = lipgloss.Color("#6C7086") // Muted gray
)

// TreeStyles contains all the styling for our tree renderer
type TreeStyles struct {
	// Tree structure styles
	TreeLines        lipgloss.Style // For tree connectors (├── └──)
	RootPath         lipgloss.Style // For the root directory name
	AnnotatedPath    lipgloss.Style // For paths that have annotations
	UnannotatedPath  lipgloss.Style // For paths without annotations
	
	// Annotation styles
	AnnotationText       lipgloss.Style // For annotation content
	AnnotationContainer  lipgloss.Style // For annotation formatting/borders
	
	// Layout styles
	AnnotationSeparator lipgloss.Style
	MultiLineIndent     lipgloss.Style
	
	// Legacy fields for backward compatibility (deprecated)
	TreeConnector lipgloss.Style // Deprecated: use TreeLines
	Directory     lipgloss.Style // Deprecated: use RootPath or AnnotationText  
	File          lipgloss.Style // Deprecated: use AnnotatedPath
	AnnotationTitle       lipgloss.Style // Deprecated: use AnnotationText
	AnnotationDescription lipgloss.Style // Deprecated: use AnnotationText
}

// NewTreeStyles creates a new set of tree styles
func NewTreeStyles() *TreeStyles {
	styles := &TreeStyles{
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
			PaddingLeft(2).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{
				Left: "│",
			}).
			BorderForeground(AnnotationBorderColor),
		
		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().
			Foreground(MutedColor).
			SetString("  "),
		
		MultiLineIndent: lipgloss.NewStyle().
			Foreground(AnnotationBorderColor).
			PaddingLeft(1),
	}
	
	// Set legacy fields for backward compatibility
	styles.TreeConnector = styles.TreeLines
	styles.Directory = styles.RootPath  
	styles.File = styles.AnnotatedPath
	styles.AnnotationTitle = styles.AnnotationText
	styles.AnnotationDescription = styles.AnnotationText
	
	return styles
}

// NewMinimalTreeStyles creates a minimal color scheme for environments with limited color support
func NewMinimalTreeStyles() *TreeStyles {
	styles := &TreeStyles{
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
	
	// Set legacy fields for backward compatibility
	styles.TreeConnector = styles.TreeLines
	styles.Directory = styles.RootPath
	styles.File = styles.AnnotatedPath
	styles.AnnotationTitle = styles.AnnotationText
	styles.AnnotationDescription = styles.AnnotationText
	
	return styles
}

// NewNoColorTreeStyles creates styles without any colors for plain text output
func NewNoColorTreeStyles() *TreeStyles {
	styles := &TreeStyles{
		// Tree structure styles
		TreeLines:       lipgloss.NewStyle(),
		RootPath:        lipgloss.NewStyle().Bold(true),
		AnnotatedPath:   lipgloss.NewStyle(),
		UnannotatedPath: lipgloss.NewStyle(),
		
		// Annotation styles
		AnnotationText:  lipgloss.NewStyle().Bold(true),
		AnnotationContainer: lipgloss.NewStyle(),
		
		// Layout styles
		AnnotationSeparator: lipgloss.NewStyle().SetString("  "),
		MultiLineIndent:     lipgloss.NewStyle(),
	}
	
	// Set legacy fields for backward compatibility
	styles.TreeConnector = styles.TreeLines
	styles.Directory = styles.RootPath
	styles.File = styles.AnnotatedPath
	styles.AnnotationTitle = styles.AnnotationText
	styles.AnnotationDescription = styles.AnnotationText
	
	return styles
} 