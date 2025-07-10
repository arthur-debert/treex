package format

import (
	"github.com/adebert/treex/pkg/core/types"
)

// Render is a convenience function that renders using the default registry
func Render(root *types.Node, options RenderOptions) (string, error) {
	registry := GetDefaultRegistry()

	// Use default format if none specified
	if options.Format == "" {
		options.Format = registry.DefaultFormat()
	}

	renderer, err := registry.GetRenderer(options.Format)
	if err != nil {
		return "", err
	}

	return renderer.Render(root, options)
}

// ParseFormatString is a convenience function for parsing format strings
func ParseFormatString(formatStr string) (OutputFormat, error) {
	registry := GetDefaultRegistry()
	return registry.ParseFormat(formatStr)
}

// ListAvailableFormats returns all available formats with descriptions
func ListAvailableFormats() map[OutputFormat]string {
	registry := GetDefaultRegistry()
	return registry.ListFormats()
}

// GetFormatHelp returns help text for all available formats
func GetFormatHelp() string {
	registry := GetDefaultRegistry()
	return registry.GetFormatHelp()
}
