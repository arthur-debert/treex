package plugins

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/core/types"
)

// Registry manages all available file info plugins
type Registry struct {
	plugins map[string]*PluginInfo
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*PluginInfo),
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(plugin FileInfoPlugin) error {
	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin with name '%s' already registered", name)
	}
	
	r.plugins[name] = &PluginInfo{
		Plugin:      plugin,
		Name:        name,
		Description: plugin.Description(),
	}
	
	return nil
}

// GetPlugin returns a plugin by name
func (r *Registry) GetPlugin(name string) (FileInfoPlugin, bool) {
	info, exists := r.plugins[name]
	if !exists {
		return nil, false
	}
	return info.Plugin, true
}

// ListPlugins returns all registered plugin names
func (r *Registry) ListPlugins() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetPluginInfo returns information about a plugin
func (r *Registry) GetPluginInfo(name string) (*PluginInfo, bool) {
	info, exists := r.plugins[name]
	return info, exists
}

// GetAllPluginInfo returns information about all registered plugins
func (r *Registry) GetAllPluginInfo() []*PluginInfo {
	infos := make([]*PluginInfo, 0, len(r.plugins))
	for _, info := range r.plugins {
		infos = append(infos, info)
	}
	return infos
}

// CollectMetadata runs the specified plugins on a node and returns collected metadata
func (r *Registry) CollectMetadata(node *types.Node, pluginNames []string) error {
	if node.Metadata == nil {
		node.Metadata = make(map[string]interface{})
	}
	
	for _, pluginName := range pluginNames {
		plugin, exists := r.GetPlugin(pluginName)
		if !exists {
			return fmt.Errorf("plugin '%s' not found", pluginName)
		}
		
		// Check if plugin applies to this node type
		if !plugin.AppliesTo(node) {
			continue
		}
		
		// Collect metadata from the plugin
		metadata, err := plugin.Collect(node)
		if err != nil {
			return fmt.Errorf("plugin '%s' failed to collect metadata: %w", pluginName, err)
		}
		
		// Store metadata with plugin prefix to avoid conflicts
		for key, value := range metadata {
			node.Metadata[pluginName+"_"+key] = value
		}
	}
	
	return nil
}

// FormatMetadata formats the metadata from specified plugins for display
func (r *Registry) FormatMetadata(node *types.Node, pluginNames []string) []string {
	var formatted []string
	
	if node.Metadata == nil {
		return formatted
	}
	
	for _, pluginName := range pluginNames {
		plugin, exists := r.GetPlugin(pluginName)
		if !exists {
			continue
		}
		
		// Check if plugin applies to this node type
		if !plugin.AppliesTo(node) {
			continue
		}
		
		// Extract metadata for this plugin
		pluginMetadata := make(map[string]interface{})
		prefix := pluginName + "_"
		for key, value := range node.Metadata {
			if strings.HasPrefix(key, prefix) {
				// Remove the plugin prefix from the key
				cleanKey := strings.TrimPrefix(key, prefix)
				pluginMetadata[cleanKey] = value
			}
		}
		
		// Format the metadata using the plugin
		if len(pluginMetadata) > 0 {
			formattedData := plugin.Format(pluginMetadata)
			if formattedData != "" {
				formatted = append(formatted, formattedData)
			}
		}
	}
	
	return formatted
}

// ValidatePlugins validates that all requested plugin names are available
func (r *Registry) ValidatePlugins(pluginNames []string) error {
	for _, name := range pluginNames {
		if _, exists := r.plugins[name]; !exists {
			available := r.ListPlugins()
			return fmt.Errorf("unknown plugin '%s'. Available plugins: %s", name, strings.Join(available, ", "))
		}
	}
	return nil
}