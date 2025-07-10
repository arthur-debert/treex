package formatting

import (
	"strings"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/display/rendering"
)

// ColorRenderer renders trees with full color styling
type ColorRenderer struct{}

func (r *ColorRenderer) Render(root *types.Node, options format.RenderOptions) (string, error) {
	var builder strings.Builder
	renderer := rendering.NewStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode)

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

// NoColorRenderer renders trees without any color styling
type NoColorRenderer struct{}

func (r *NoColorRenderer) Render(root *types.Node, options format.RenderOptions) (string, error) {
	// Use no-color renderer with extra spacing support
	var builder strings.Builder
	renderer := rendering.NewNoColorStyledTreeRenderer(&builder, true).
		WithSafeMode(options.SafeMode)

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
