// Package plugins - Plugin option service for platform-agnostic plugin capability discovery
package plugins

import (
	"fmt"
)

// PluginOptionDefinition represents a single plugin filter option
// This is platform-agnostic and doesn't contain UI-specific details
type PluginOptionDefinition struct {
	PluginName   string // Name of the plugin (e.g., "git", "info")
	CategoryName string // Name of the category (e.g., "staged", "annotated")
	Description  string // Human-readable description
}

// PluginOptionService provides platform-agnostic access to plugin filtering capabilities
// This abstracts plugin capability discovery from UI concerns (CLI flags, API parameters, etc.)
type PluginOptionService interface {
	// GetAvailableOptions returns all available plugin filter options
	GetAvailableOptions() []PluginOptionDefinition

	// ValidatePluginFilters checks if the provided filters reference valid plugin options
	ValidatePluginFilters(filters map[string]map[string]bool) error

	// GetPluginNames returns the names of all registered plugins that support filtering
	GetPluginNames() []string

	// GetPluginCategories returns the categories for a specific plugin
	GetPluginCategories(pluginName string) ([]FilterPluginCategory, error)
}

// DefaultPluginOptionService implements PluginOptionService using the plugin registry
type DefaultPluginOptionService struct {
	registry *Registry
}

// NewPluginOptionService creates a new plugin option service
func NewPluginOptionService(registry *Registry) PluginOptionService {
	return &DefaultPluginOptionService{
		registry: registry,
	}
}

// GetAvailableOptions returns all available plugin filter options
func (s *DefaultPluginOptionService) GetAvailableOptions() []PluginOptionDefinition {
	var options []PluginOptionDefinition

	// Get all registered plugins
	plugins := s.registry.GetAllPlugins()

	for _, plugin := range plugins {
		// Check if plugin supports filtering
		if filterPlugin, ok := plugin.(FilterPlugin); ok {
			pluginName := plugin.Name()
			categories := filterPlugin.GetCategories()

			// Convert each category to a platform-agnostic option definition
			for _, category := range categories {
				option := PluginOptionDefinition{
					PluginName:   pluginName,
					CategoryName: category.Name,
					Description:  category.Description,
				}
				options = append(options, option)
			}
		}
	}

	return options
}

// ValidatePluginFilters checks if the provided filters reference valid plugin options
func (s *DefaultPluginOptionService) ValidatePluginFilters(filters map[string]map[string]bool) error {
	// Build a map of valid options for quick lookup
	validOptions := make(map[string]map[string]bool)
	availableOptions := s.GetAvailableOptions()

	for _, option := range availableOptions {
		if validOptions[option.PluginName] == nil {
			validOptions[option.PluginName] = make(map[string]bool)
		}
		validOptions[option.PluginName][option.CategoryName] = true
	}

	// Validate each filter
	for pluginName, categories := range filters {
		if validOptions[pluginName] == nil {
			return fmt.Errorf("unknown plugin: %s", pluginName)
		}

		for categoryName, enabled := range categories {
			// Only validate enabled filters (disabled filters can reference any category)
			if enabled && !validOptions[pluginName][categoryName] {
				return fmt.Errorf("unknown category %s for plugin %s", categoryName, pluginName)
			}
		}
	}

	return nil
}

// GetPluginNames returns the names of all registered plugins that support filtering
func (s *DefaultPluginOptionService) GetPluginNames() []string {
	var names []string
	plugins := s.registry.GetAllPlugins()

	for _, plugin := range plugins {
		if _, ok := plugin.(FilterPlugin); ok {
			names = append(names, plugin.Name())
		}
	}

	return names
}

// GetPluginCategories returns the categories for a specific plugin
func (s *DefaultPluginOptionService) GetPluginCategories(pluginName string) ([]FilterPluginCategory, error) {
	plugin := s.registry.GetPlugin(pluginName)
	if plugin == nil {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	filterPlugin, ok := plugin.(FilterPlugin)
	if !ok {
		return nil, fmt.Errorf("plugin %s does not support filtering", pluginName)
	}

	return filterPlugin.GetCategories(), nil
}

// GetDefaultPluginOptionService provides a global service instance using the default registry
func GetDefaultPluginOptionService() PluginOptionService {
	return NewPluginOptionService(GetDefaultRegistry())
}
