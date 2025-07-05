package rendering

import (
	"io"
	
	"github.com/adebert/treex/pkg/display/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// StyleRenderer wraps a lipgloss renderer with our style system
type StyleRenderer struct {
	renderer *lipgloss.Renderer
	styles   *styles.TreeStyles
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

// NewStyleRendererWithAutoTheme creates a new style renderer with automatic theme detection
func NewStyleRendererWithAutoTheme(output io.Writer, verbose bool) *StyleRenderer {
	sr := NewStyleRenderer(output)
	
	// Auto-detect theme for this renderer
	_ = sr.AutoDetectTheme(verbose)
	
	return sr
}

// NewTreeStylesWithRenderer creates tree styles using a specific renderer
func NewTreeStylesWithRenderer(r *lipgloss.Renderer) *styles.TreeStyles {
	// Create base styles that use the renderer
	base := NewBaseStylesWithRenderer(r)
	
	return &styles.TreeStyles{
		Base: base,
		
		// Tree structure styles - inherit from base styles
		TreeLines: base.Structure.Faint(true),
		
		RootPath: r.NewStyle().
			Foreground(styles.Colors.TreeDirectory).
			Bold(true),
		
		AnnotatedPath: base.Text,  // Items with info use regular text
		
		UnannotatedPath: base.TextFaint,  // Items without info use faint text
		
		// Annotation styles - compose from base styles
		AnnotationText: base.TextTitle,  // Use title style for inline annotations
		
		AnnotationNotes: base.TextTitle,  // Use title style (bold) for notes
		
		AnnotationDescription: base.TextSubtle,  // Use subtle for descriptions
		
		AnnotationContainer: base.Text,  // No padding to align with title
		
		// Layout styles
		AnnotationSeparator: base.TextFaint.SetString("  "),
		
		MultiLineIndent: base.Border.
			Faint(true).
			PaddingLeft(1),
	}
}

// NewBaseStylesWithRenderer creates base styles using a specific renderer
func NewBaseStylesWithRenderer(r *lipgloss.Renderer) *styles.BaseStyles {
	return &styles.BaseStyles{
		// Base text styles
		Text:       r.NewStyle().Foreground(styles.Colors.Text),
		TextBold:   r.NewStyle().Foreground(styles.Colors.TextBold).Bold(true),
		TextFaint:  r.NewStyle().Foreground(styles.Colors.TextMuted).Faint(true),
		TextSubtle: r.NewStyle().Foreground(styles.Colors.TextSubtle),
		TextTitle:  r.NewStyle().Foreground(styles.Colors.TextTitle).Bold(true),
		
		// Base semantic styles
		Primary:   r.NewStyle().Foreground(styles.Colors.Primary),
		Secondary: r.NewStyle().Foreground(styles.Colors.Secondary),
		Success:   r.NewStyle().Foreground(styles.Colors.Success),
		Warning:   r.NewStyle().Foreground(styles.Colors.Warning),
		Error:     r.NewStyle().Foreground(styles.Colors.Error),
		Info:      r.NewStyle().Foreground(styles.Colors.Info),
		
		// Base structure styles
		Structure: r.NewStyle().Foreground(styles.Colors.TreeConnector),
		Border:    r.NewStyle().Foreground(styles.Colors.Border),
	}
}

// Renderer returns the underlying lipgloss renderer
func (sr *StyleRenderer) Renderer() *lipgloss.Renderer {
	return sr.renderer
}

// Styles returns the tree styles for this renderer
func (sr *StyleRenderer) Styles() *styles.TreeStyles {
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

// AutoDetectTheme detects and sets the terminal theme for this renderer
func (sr *StyleRenderer) AutoDetectTheme(verbose bool) error {
	// Lipgloss already handles theme detection internally
	// The renderer will automatically detect the terminal's background color
	// when created, so we don't need to do anything special here
	
	// We can still allow manual override if needed
	sr.renderer.SetHasDarkBackground(sr.renderer.HasDarkBackground())
	
	return nil
}

// NewMinimalStyleRenderer creates a style renderer with minimal color support
func NewMinimalStyleRenderer(output io.Writer) *StyleRenderer {
	renderer := lipgloss.NewRenderer(output)
	
	// Force ANSI color profile for minimal colors
	renderer.SetColorProfile(termenv.ANSI)
	
	// Create minimal styles
	base := NewMinimalBaseStylesWithRenderer(renderer)
	styles := &styles.TreeStyles{
		Base: base,
		// Tree structure styles - using minimal base styles
		TreeLines:       base.Structure,
		RootPath:        base.TextBold,
		AnnotatedPath:   base.Text,
		UnannotatedPath: base.Structure,
		
		// Annotation styles
		AnnotationText:        base.TextBold,
		AnnotationNotes:       base.TextBold,
		AnnotationDescription: base.Text,
		AnnotationContainer:   base.Text.PaddingLeft(1),
		
		// Layout styles
		AnnotationSeparator: base.Text.SetString("  "),
		MultiLineIndent:     base.Text.PaddingLeft(1),
	}
	
	return &StyleRenderer{
		renderer: renderer,
		styles:   styles,
	}
}

// NewNoColorStyleRenderer creates a style renderer without any colors
func NewNoColorStyleRenderer(output io.Writer) *StyleRenderer {
	renderer := lipgloss.NewRenderer(output)
	
	// Force ASCII profile for no colors
	renderer.SetColorProfile(termenv.Ascii)
	
	// Create no-color styles
	base := NewNoColorBaseStylesWithRenderer(renderer)
	styles := &styles.TreeStyles{
		Base: base,
		// Tree structure styles - using no-color base styles
		TreeLines:       base.Structure,
		RootPath:        base.TextBold,
		AnnotatedPath:   base.Text,
		UnannotatedPath: base.Structure,
		
		// Annotation styles
		AnnotationText:        base.TextBold,
		AnnotationNotes:       base.TextBold,
		AnnotationDescription: base.Text,
		AnnotationContainer:   base.Text,
		
		// Layout styles
		AnnotationSeparator: base.Text.SetString("  "),
		MultiLineIndent:     base.Text,
	}
	
	return &StyleRenderer{
		renderer: renderer,
		styles:   styles,
	}
}

// NewMinimalBaseStylesWithRenderer creates minimal base styles with a specific renderer
func NewMinimalBaseStylesWithRenderer(r *lipgloss.Renderer) *styles.BaseStyles {
	gray := lipgloss.Color("8")
	return &styles.BaseStyles{
		Text:       r.NewStyle(),
		TextBold:   r.NewStyle().Bold(true),
		TextFaint:  r.NewStyle().Foreground(gray),
		TextSubtle: r.NewStyle(),
		TextTitle:  r.NewStyle().Bold(true),
		Primary:    r.NewStyle(),
		Secondary:  r.NewStyle(),
		Success:    r.NewStyle(),
		Warning:    r.NewStyle(),
		Error:      r.NewStyle(),
		Info:       r.NewStyle(),
		Structure:  r.NewStyle().Foreground(gray),
		Border:     r.NewStyle().Foreground(gray),
	}
}

// NewNoColorBaseStylesWithRenderer creates base styles without colors with a specific renderer
func NewNoColorBaseStylesWithRenderer(r *lipgloss.Renderer) *styles.BaseStyles {
	return &styles.BaseStyles{
		Text:       r.NewStyle(),
		TextBold:   r.NewStyle().Bold(true),
		TextFaint:  r.NewStyle(),
		TextSubtle: r.NewStyle(),
		TextTitle:  r.NewStyle().Bold(true),
		Primary:    r.NewStyle(),
		Secondary:  r.NewStyle(),
		Success:    r.NewStyle(),
		Warning:    r.NewStyle(),
		Error:      r.NewStyle(),
		Info:       r.NewStyle(),
		Structure:  r.NewStyle(),
		Border:     r.NewStyle(),
	}
}