package format

import (
	"sync"

	"github.com/adebert/treex/pkg/tree"
)

var (
	defaultRegistry *RendererRegistry
	registryOnce    sync.Once
)

// GetDefaultRegistry returns the default renderer registry with all built-in renderers
func GetDefaultRegistry() *RendererRegistry {
	registryOnce.Do(func() {
		defaultRegistry = NewRendererRegistry()

		// Register all built-in terminal renderers
		_ = defaultRegistry.Register(NewColorRenderer())
		_ = defaultRegistry.Register(NewMinimalRenderer())
		_ = defaultRegistry.Register(NewNoColorRenderer())

		// Register data format renderers
		_ = defaultRegistry.Register(NewJSONRenderer())
		_ = defaultRegistry.Register(NewYAMLRenderer())
		_ = defaultRegistry.Register(NewCompactJSONRenderer())
		_ = defaultRegistry.Register(NewFlatJSONRenderer())

		// Register markdown renderers
		_ = defaultRegistry.Register(NewMarkdownRenderer())
		_ = defaultRegistry.Register(NewNestedMarkdownRenderer())
		_ = defaultRegistry.Register(NewTableMarkdownRenderer())
	})

	return defaultRegistry
}

// Render is a convenience function that renders using the default registry
func Render(root *tree.Node, options RenderOptions) (string, error) {
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
