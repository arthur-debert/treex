package format

import (
	"strings"
	
	"github.com/adebert/treex/pkg/tree"
	"github.com/adebert/treex/pkg/tui"
)

// ColorRenderer renders trees with full color styling
type ColorRenderer struct{}


func (r *ColorRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	// Use RenderStyledTreeWithOptions to support extra spacing
	var builder strings.Builder
	renderer := tui.NewStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode).
		WithExtraSpacing(options.ExtraSpacing)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

func (r *ColorRenderer) Format() OutputFormat {
	return FormatColor
}

func (r *ColorRenderer) Description() string {
	return "Full color terminal output with beautiful styling (default)"
}

func (r *ColorRenderer) IsTerminalFormat() bool {
	return true
}

// MinimalRenderer renders trees with minimal color styling
type MinimalRenderer struct{}


func (r *MinimalRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	// Use minimal renderer with extra spacing support
	var builder strings.Builder
	renderer := tui.NewMinimalStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode).
		WithExtraSpacing(options.ExtraSpacing)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

func (r *MinimalRenderer) Format() OutputFormat {
	return FormatMinimal
}

func (r *MinimalRenderer) Description() string {
	return "Minimal color styling for basic terminals"
}

func (r *MinimalRenderer) IsTerminalFormat() bool {
	return true
}

// NoColorRenderer renders trees without any color styling
type NoColorRenderer struct{}


func (r *NoColorRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	// Use no-color renderer with extra spacing support
	var builder strings.Builder
	renderer := tui.NewNoColorStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode).
		WithExtraSpacing(options.ExtraSpacing)
	
	if err := renderer.Render(root); err != nil {
		return "", err
	}
	
	return builder.String(), nil
}

func (r *NoColorRenderer) Format() OutputFormat {
	return FormatNoColor
}

func (r *NoColorRenderer) Description() string {
	return "Plain text output without colors"
}

func (r *NoColorRenderer) IsTerminalFormat() bool {
	return true
}
