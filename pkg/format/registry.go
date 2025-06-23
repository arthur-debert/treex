package format

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/tree"
)

// OutputFormat represents a specific output format
type OutputFormat string

const (
	// Terminal-based formats
	FormatColor   OutputFormat = "color"
	FormatMinimal OutputFormat = "minimal"
	FormatNoColor OutputFormat = "no-color"

	// Structured data formats (for future implementation)
	FormatJSON OutputFormat = "json"
	FormatYAML OutputFormat = "yaml"
)

// RenderOptions contains configuration options for rendering
type RenderOptions struct {
	Format        OutputFormat
	Verbose       bool
	ShowStats     bool
	IgnoreFile    string
	MaxDepth      int
	SafeMode      bool
	TerminalWidth int
}

// Renderer interface defines the contract for all output format renderers
type Renderer interface {
	// Render takes a tree and options and returns the formatted output
	Render(root *tree.Node, options RenderOptions) (string, error)

	// Format returns the format this renderer handles
	Format() OutputFormat

	// Description returns a human-readable description of the format
	Description() string

	// IsTerminalFormat returns true if this format is designed for terminal output
	IsTerminalFormat() bool
}

// RendererRegistry manages the available output format renderers
type RendererRegistry struct {
	renderers map[OutputFormat]Renderer
	aliases   map[string]OutputFormat // Alternative names for formats
}

// NewRendererRegistry creates a new renderer registry
func NewRendererRegistry() *RendererRegistry {
	return &RendererRegistry{
		renderers: make(map[OutputFormat]Renderer),
		aliases: map[string]OutputFormat{
			// Standard format names
			"color":    FormatColor,
			"minimal":  FormatMinimal,
			"no-color": FormatNoColor,
			"json":     FormatJSON,
			"yaml":     FormatYAML,

			// Alternative aliases
			"colorful": FormatColor,
			"full":     FormatColor,
			"plain":    FormatNoColor,
			"text":     FormatNoColor,
			"simple":   FormatMinimal,
		},
	}
}

// Register adds a renderer to the registry
func (r *RendererRegistry) Register(renderer Renderer) error {
	if renderer == nil {
		return fmt.Errorf("renderer cannot be nil")
	}

	format := renderer.Format()
	if format == "" {
		return fmt.Errorf("renderer format cannot be empty")
	}

	r.renderers[format] = renderer
	return nil
}

// GetRenderer returns the renderer for the specified format
func (r *RendererRegistry) GetRenderer(format OutputFormat) (Renderer, error) {
	renderer, exists := r.renderers[format]
	if !exists {
		return nil, fmt.Errorf("no renderer registered for format: %s", format)
	}
	return renderer, nil
}

// ParseFormat converts a string to an OutputFormat, handling aliases
func (r *RendererRegistry) ParseFormat(formatStr string) (OutputFormat, error) {
	formatStr = strings.ToLower(strings.TrimSpace(formatStr))

	// Try direct format match first
	format := OutputFormat(formatStr)
	if _, exists := r.renderers[format]; exists {
		return format, nil
	}

	// Try aliases
	if aliasFormat, exists := r.aliases[formatStr]; exists {
		if _, rendererExists := r.renderers[aliasFormat]; rendererExists {
			return aliasFormat, nil
		}
	}

	return "", fmt.Errorf("unknown format: %s", formatStr)
}

// ListFormats returns all available formats with their descriptions
func (r *RendererRegistry) ListFormats() map[OutputFormat]string {
	formats := make(map[OutputFormat]string)
	for format, renderer := range r.renderers {
		formats[format] = renderer.Description()
	}
	return formats
}

// GetTerminalFormats returns only the terminal-based formats
func (r *RendererRegistry) GetTerminalFormats() []OutputFormat {
	var terminalFormats []OutputFormat
	for format, renderer := range r.renderers {
		if renderer.IsTerminalFormat() {
			terminalFormats = append(terminalFormats, format)
		}
	}
	return terminalFormats
}

// GetDataFormats returns only the structured data formats
func (r *RendererRegistry) GetDataFormats() []OutputFormat {
	var dataFormats []OutputFormat
	for format, renderer := range r.renderers {
		if !renderer.IsTerminalFormat() {
			dataFormats = append(dataFormats, format)
		}
	}
	return dataFormats
}

// ValidateFormat checks if a format is valid and registered
func (r *RendererRegistry) ValidateFormat(format OutputFormat) error {
	if _, exists := r.renderers[format]; !exists {
		available := make([]string, 0, len(r.renderers))
		for f := range r.renderers {
			available = append(available, string(f))
		}
		return fmt.Errorf("format %q is not available. Available formats: %s",
			format, strings.Join(available, ", "))
	}
	return nil
}

// DefaultFormat returns the default output format
func (r *RendererRegistry) DefaultFormat() OutputFormat {
	// Color is the default for the best user experience
	return FormatColor
}

// GetFormatHelp returns help text for format selection
func (r *RendererRegistry) GetFormatHelp() string {
	var help strings.Builder
	help.WriteString("Available output formats:\n\n")

	// Group by type
	terminalFormats := r.GetTerminalFormats()
	dataFormats := r.GetDataFormats()

	if len(terminalFormats) > 0 {
		help.WriteString("Terminal formats:\n")
		for _, format := range terminalFormats {
			if renderer, exists := r.renderers[format]; exists {
				help.WriteString(fmt.Sprintf("  %-10s - %s\n", format, renderer.Description()))
			}
		}
		help.WriteString("\n")
	}

	if len(dataFormats) > 0 {
		help.WriteString("Data formats:\n")
		for _, format := range dataFormats {
			if renderer, exists := r.renderers[format]; exists {
				help.WriteString(fmt.Sprintf("  %-10s - %s\n", format, renderer.Description()))
			}
		}
	}

	return help.String()
}
