package formatting

import (
	"strings"

	"github.com/adebert/treex/pkg/config"
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/display/rendering"
)

// ColorRenderer renders trees with full color styling
type ColorRenderer struct {
	config *config.Config
}

// NewColorRenderer creates a new color renderer
func NewColorRenderer() *ColorRenderer {
	return &ColorRenderer{}
}

// NewColorRendererWithConfig creates a new color renderer with configuration
func NewColorRendererWithConfig(cfg *config.Config) *ColorRenderer {
	return &ColorRenderer{config: cfg}
}

func (r *ColorRenderer) Render(root *types.Node, options format.RenderOptions) (string, error) {
	var builder strings.Builder
	
	var renderer *rendering.StyledTreeRenderer
	if r.config != nil {
		renderer = rendering.NewStyledTreeRendererWithConfig(&builder, true, r.config).
			WithSafeMode(options.SafeMode)
	} else {
		renderer = rendering.NewStyledTreeRenderer(&builder, true).
			WithSafeMode(options.SafeMode)
	}

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
