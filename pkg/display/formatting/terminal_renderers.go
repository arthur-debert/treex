package formatting

import (
	"strings"
	
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/tree"
	"github.com/adebert/treex/pkg/display/rendering"
)

// ColorRenderer renders trees with full color styling
type ColorRenderer struct{}


func (r *ColorRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
	// Use RenderStyledTreeWithOptions to support extra spacing
	var builder strings.Builder
	renderer := rendering.NewStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode).
		WithExtraSpacing(options.ExtraSpacing)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

func (r *ColorRenderer) Format() format.OutputFormat {
	return format.FormatColor
}

func (r *ColorRenderer) Description() string {
	return "Full color terminal output with beautiful styling (default)"
}

func (r *ColorRenderer) IsTerminalFormat() bool {
	return true
}

// MinimalRenderer renders trees with minimal color styling
type MinimalRenderer struct{}


func (r *MinimalRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
	// Use minimal renderer with extra spacing support
	var builder strings.Builder
	renderer := rendering.NewMinimalStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode).
		WithExtraSpacing(options.ExtraSpacing)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

func (r *MinimalRenderer) Format() format.OutputFormat {
	return format.FormatMinimal
}

func (r *MinimalRenderer) Description() string {
	return "Minimal color styling for basic terminals"
}

func (r *MinimalRenderer) IsTerminalFormat() bool {
	return true
}

// NoColorRenderer renders trees without any color styling
type NoColorRenderer struct{}


func (r *NoColorRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
	// Use no-color renderer with extra spacing support
	var builder strings.Builder
	renderer := rendering.NewNoColorStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode).
		WithExtraSpacing(options.ExtraSpacing)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

func (r *NoColorRenderer) Format() format.OutputFormat {
	return format.FormatNoColor
}

func (r *NoColorRenderer) Description() string {
	return "Plain text output without colors"
}

func (r *NoColorRenderer) IsTerminalFormat() bool {
	return true
}
