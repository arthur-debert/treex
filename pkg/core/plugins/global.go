package plugins

import (
	"sync"
)

var (
	globalRegistry *Registry
	once           sync.Once
)

// GetGlobalRegistry returns the global plugin registry, initializing it if necessary
func GetGlobalRegistry() *Registry {
	once.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}

// InitializeGlobalRegistry initializes the global registry with built-in plugins
// This should be called once during application startup
func InitializeGlobalRegistry() error {
	_ = GetGlobalRegistry()
	
	// Import built-in plugins - we'll do this in the caller to avoid circular imports
	return nil
}