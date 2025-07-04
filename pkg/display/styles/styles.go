package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// BaseStyles contains reusable base style components
type BaseStyles struct {
	// Base text styles
	Text       lipgloss.Style // Base text style
	TextBold   lipgloss.Style // Bold text
	TextFaint  lipgloss.Style // Faint/muted text
	TextSubtle lipgloss.Style // Subtle text
	TextTitle  lipgloss.Style // Title text (bold)
	
	// Base semantic styles
	Primary   lipgloss.Style // Primary color items
	Secondary lipgloss.Style // Secondary color items
	Success   lipgloss.Style // Success state
	Warning   lipgloss.Style // Warning state
	Error     lipgloss.Style // Error state
	Info      lipgloss.Style // Info state
	
	// Base structure styles
	Structure lipgloss.Style // Tree structure elements
	Border    lipgloss.Style // Borders and dividers
}

// TreeStyles contains all the styling for our tree renderer
type TreeStyles struct {
	// Base styles for inheritance
	Base *BaseStyles
	
	// Tree structure styles
	TreeLines       lipgloss.Style // For tree connectors (├── └──)
	RootPath        lipgloss.Style // For the root directory name
	AnnotatedPath   lipgloss.Style // For paths that have annotations
	UnannotatedPath lipgloss.Style // For paths without annotations

	// Annotation styles
	AnnotationText        lipgloss.Style // For annotation content (inline)
	AnnotationNotes       lipgloss.Style // For annotation notes
	AnnotationDescription lipgloss.Style // For annotation descriptions (multi-line)
	AnnotationContainer   lipgloss.Style // For annotation formatting/borders

	// Layout styles
	AnnotationSeparator lipgloss.Style
	MultiLineIndent     lipgloss.Style
}

// NewBaseStyles creates base style components
func NewBaseStyles() *BaseStyles {
	return &BaseStyles{
		// Base text styles
		Text:       lipgloss.NewStyle().Foreground(Colors.Text),
		TextBold:   lipgloss.NewStyle().Foreground(Colors.TextBold).Bold(true),
		TextFaint:  lipgloss.NewStyle().Foreground(Colors.TextMuted).Faint(true),
		TextSubtle: lipgloss.NewStyle().Foreground(Colors.TextSubtle),
		TextTitle:  lipgloss.NewStyle().Foreground(Colors.TextTitle).Bold(true),
		
		// Base semantic styles
		Primary:   lipgloss.NewStyle().Foreground(Colors.Primary),
		Secondary: lipgloss.NewStyle().Foreground(Colors.Secondary),
		Success:   lipgloss.NewStyle().Foreground(Colors.Success),
		Warning:   lipgloss.NewStyle().Foreground(Colors.Warning),
		Error:     lipgloss.NewStyle().Foreground(Colors.Error),
		Info:      lipgloss.NewStyle().Foreground(Colors.Info),
		
		// Base structure styles
		Structure: lipgloss.NewStyle().Foreground(Colors.TreeConnector),
		Border:    lipgloss.NewStyle().Foreground(Colors.Border),
	}
}

// NewTreeStyles creates a new set of tree styles with adaptive colors
func NewTreeStyles() *TreeStyles {
	base := NewBaseStyles()
	
	return &TreeStyles{
		Base: base,
		// Tree structure styles - inherit from base styles
		TreeLines: base.Structure.Faint(true),
		
		RootPath: lipgloss.NewStyle().
			Foreground(Colors.TreeDirectory).
			Bold(true),
		
		AnnotatedPath: base.TextBold,  // Items with info use bold text
		
		UnannotatedPath: base.TextFaint.Bold(true),  // Items without info use faint bold text
		
		// Annotation styles - compose from base styles
		AnnotationText: base.Text,  // Use regular text for inline annotations
		
		AnnotationNotes: base.Text,  // Use regular text for notes
		
		AnnotationDescription: base.TextSubtle,  // Use subtle for descriptions
		
		AnnotationContainer: base.Text,  // No padding to align with title
		
		// Layout styles
		AnnotationSeparator: base.TextFaint.SetString("  "),
		
		MultiLineIndent: base.Border.
			Faint(true).
			PaddingLeft(1),
	}
}

// NewMinimalBaseStyles creates minimal base style components
func NewMinimalBaseStyles() *BaseStyles {
	gray := lipgloss.Color("8")
	return &BaseStyles{
		Text:       lipgloss.NewStyle(),
		TextBold:   lipgloss.NewStyle().Bold(true),
		TextFaint:  lipgloss.NewStyle().Foreground(gray),
		TextSubtle: lipgloss.NewStyle(),
		TextTitle:  lipgloss.NewStyle().Bold(true),
		Primary:    lipgloss.NewStyle(),
		Secondary:  lipgloss.NewStyle(),
		Success:    lipgloss.NewStyle(),
		Warning:    lipgloss.NewStyle(),
		Error:      lipgloss.NewStyle(),
		Info:       lipgloss.NewStyle(),
		Structure:  lipgloss.NewStyle().Foreground(gray),
		Border:     lipgloss.NewStyle().Foreground(gray),
	}
}

// NewMinimalTreeStyles creates a minimal color scheme for environments with limited color support
func NewMinimalTreeStyles() *TreeStyles {
	base := NewMinimalBaseStyles()
	
	return &TreeStyles{
		Base: base,
		// Tree structure styles - using minimal base styles
		TreeLines:       base.Structure,
		RootPath:        base.TextBold,
		AnnotatedPath:   base.TextBold,
		UnannotatedPath: base.Structure.Bold(true),
		
		// Annotation styles
		AnnotationText:        base.Text,
		AnnotationNotes:       base.Text,
		AnnotationDescription: base.Text,
		AnnotationContainer:   base.Text,
		
		// Layout styles
		AnnotationSeparator: base.Text.SetString("  "),
		MultiLineIndent:     base.Text.PaddingLeft(1),
	}
}

// NewNoColorBaseStyles creates base styles without colors
func NewNoColorBaseStyles() *BaseStyles {
	return &BaseStyles{
		Text:       lipgloss.NewStyle(),
		TextBold:   lipgloss.NewStyle().Bold(true),
		TextFaint:  lipgloss.NewStyle(),
		TextSubtle: lipgloss.NewStyle(),
		TextTitle:  lipgloss.NewStyle().Bold(true),
		Primary:    lipgloss.NewStyle(),
		Secondary:  lipgloss.NewStyle(),
		Success:    lipgloss.NewStyle(),
		Warning:    lipgloss.NewStyle(),
		Error:      lipgloss.NewStyle(),
		Info:       lipgloss.NewStyle(),
		Structure:  lipgloss.NewStyle(),
		Border:     lipgloss.NewStyle(),
	}
}

// NewNoColorTreeStyles creates styles without any colors for plain text output
func NewNoColorTreeStyles() *TreeStyles {
	base := NewNoColorBaseStyles()
	
	return &TreeStyles{
		Base: base,
		// Tree structure styles - using no-color base styles
		TreeLines:       base.Structure,
		RootPath:        base.TextBold,
		AnnotatedPath:   base.TextBold,
		UnannotatedPath: base.Structure.Bold(true),
		
		// Annotation styles
		AnnotationText:        base.Text,
		AnnotationNotes:       base.Text,
		AnnotationDescription: base.Text,
		AnnotationContainer:   base.Text,
		
		// Layout styles
		AnnotationSeparator: base.Text.SetString("  "),
		MultiLineIndent:     base.Text,
	}
}
