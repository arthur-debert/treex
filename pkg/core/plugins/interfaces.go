package plugins

import (
	"github.com/adebert/treex/pkg/core/types"
)

// FileInfoPlugin defines the interface for plugins that collect additional file information
type FileInfoPlugin interface {
	// Name returns the unique identifier for this plugin (e.g., "size", "date")
	Name() string
	
	// Description returns a human-readable description of what this plugin provides
	Description() string
	
	// AppliesTo returns whether this plugin applies to files, directories, or both
	AppliesTo(node *types.Node) bool
	
	// Collect gathers metadata for a node and returns it as key-value pairs
	Collect(node *types.Node) (map[string]interface{}, error)
	
	// Format takes the collected metadata and formats it for display
	Format(metadata map[string]interface{}) string
}

// PluginInfo contains information about a registered plugin
type PluginInfo struct {
	Plugin      FileInfoPlugin
	Name        string
	Description string
}