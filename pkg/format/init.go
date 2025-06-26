package format

import (
	"sync"
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

		// Register HTML renderers
		_ = defaultRegistry.Register(NewHTMLRenderer())
		_ = defaultRegistry.Register(NewCompactHTMLRenderer())
		_ = defaultRegistry.Register(NewTableHTMLRenderer())

		// Register SimpleList renderer
		_ = defaultRegistry.Register(NewSimpleListRenderer())
	})

	return defaultRegistry
}