package tui

import (
	"io"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// StyleRenderer wraps a lipgloss renderer with our style system
type StyleRenderer struct {
	renderer *lipgloss.Renderer
	styles   *TreeStyles
}

// NewStyleRenderer creates a new style renderer for the given output
func NewStyleRenderer(output io.Writer) *StyleRenderer {
	// Create a lipgloss renderer for the specific output
	renderer := lipgloss.NewRenderer(output)
	
	// Create styles using the renderer's color profile
	styles := NewTreeStylesWithRenderer(renderer)
	
	return &StyleRenderer{
		renderer: renderer,
		styles:   styles,
	}
}

// NewTreeStylesWithRenderer creates tree styles using a specific renderer
func NewTreeStylesWithRenderer(r *lipgloss.Renderer) *TreeStyles {
	// Create base styles that use the renderer
	base := NewBaseStylesWithRenderer(r)
	
	return &TreeStyles{
		Base: base,
		
		// Tree structure styles - inherit from base styles
		TreeLines: base.Structure.Faint(true),
		
		RootPath: r.NewStyle().
			Foreground(Colors.TreeDirectory).
			Bold(true),
		
		AnnotatedPath: base.Text.Foreground(Colors.TreeFile),
		
		UnannotatedPath: base.Structure.Faint(true),
		
		// Annotation styles - compose from base styles
		AnnotationText: base.Primary.Bold(true),
		
		AnnotationTitle: base.Warning.Bold(true),
		
		AnnotationDescription: base.Success,
		
		AnnotationContainer: base.Text.PaddingLeft(1),
		
		// Layout styles
		AnnotationSeparator: base.TextFaint.SetString("  "),
		
		MultiLineIndent: base.Border.
			Faint(true).
			PaddingLeft(1),
	}
}

// NewBaseStylesWithRenderer creates base styles using a specific renderer
func NewBaseStylesWithRenderer(r *lipgloss.Renderer) *BaseStyles {
	return &BaseStyles{
		// Base text styles
		Text:      r.NewStyle().Foreground(Colors.Text),
		TextBold:  r.NewStyle().Foreground(Colors.TextBold).Bold(true),
		TextFaint: r.NewStyle().Foreground(Colors.TextMuted).Faint(true),
		
		// Base semantic styles
		Primary:   r.NewStyle().Foreground(Colors.Primary),
		Secondary: r.NewStyle().Foreground(Colors.Secondary),
		Success:   r.NewStyle().Foreground(Colors.Success),
		Warning:   r.NewStyle().Foreground(Colors.Warning),
		Error:     r.NewStyle().Foreground(Colors.Error),
		Info:      r.NewStyle().Foreground(Colors.Info),
		
		// Base structure styles
		Structure: r.NewStyle().Foreground(Colors.TreeConnector),
		Border:    r.NewStyle().Foreground(Colors.Border),
	}
}

// Renderer returns the underlying lipgloss renderer
func (sr *StyleRenderer) Renderer() *lipgloss.Renderer {
	return sr.renderer
}

// Styles returns the tree styles for this renderer
func (sr *StyleRenderer) Styles() *TreeStyles {
	return sr.styles
}

// SetColorProfile sets the color profile for the renderer
func (sr *StyleRenderer) SetColorProfile(profile termenv.Profile) {
	sr.renderer.SetColorProfile(profile)
}

// SetHasDarkBackground sets whether the terminal has a dark background
func (sr *StyleRenderer) SetHasDarkBackground(dark bool) {
	sr.renderer.SetHasDarkBackground(dark)
}

// HasDarkBackground returns whether the terminal is set to have a dark background
func (sr *StyleRenderer) HasDarkBackground() bool {
	return sr.renderer.HasDarkBackground()
}