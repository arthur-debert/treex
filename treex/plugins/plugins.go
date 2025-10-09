// see docs/dev/architecture.txt - Phase 4: Plugin Filtering
package plugins

import (
	"fmt"

	"github.com/jwaldrip/treex/treex/types"
	"github.com/spf13/afero"
)

// Plugin defines the interface that all plugins must implement
// Plugins operate on file lists, not tree structures, for simplicity
type Plugin interface {
	// Name returns the unique identifier for this plugin
	Name() string

	// FindRoots discovers relevant root directories for this plugin
	// For example, a git plugin would find directories containing .git folders
	// Returns relative paths from the search root
	FindRoots(fs afero.Fs, searchRoot string) ([]string, error)

	// ProcessRoot processes a single root directory and returns filtered results
	// The plugin analyzes files within this root and categorizes them
	ProcessRoot(fs afero.Fs, rootPath string) (*Result, error)
}

// FilterPluginCategory represents a filter category provided by a filter plugin
type FilterPluginCategory struct {
	// Name is the category identifier - must be a valid identifier
	// Used for CLI flags (--plugin-name), map keys, API parameters
	Name string

	// Description is human-readable text for CLI help and documentation
	Description string
}

// FilterPlugin extends Plugin with file categorization capabilities
// Filter plugins categorize files into named categories for visibility control
type FilterPlugin interface {
	Plugin

	// GetCategories returns the static list of filter categories this plugin provides
	// Called during plugin registration for CLI flag generation
	GetCategories() []FilterPluginCategory
}

// DataPlugin extends Plugin with node data enrichment capabilities
// Data plugins attach additional information to nodes after filtering
type DataPlugin interface {
	Plugin

	// EnrichNode attaches plugin-specific data to a node
	// Called only for nodes that survived filtering to avoid expensive operations
	// Data should be stored in node.Data[pluginName]
	EnrichNode(fs afero.Fs, node *types.Node) error
}

// CachedDataPlugin extends DataPlugin with cache-aware enrichment capabilities
// For plugins that implement both FilterPlugin and DataPlugin, this allows
// reusing data from the filtering phase to avoid expensive re-computation
type CachedDataPlugin interface {
	DataPlugin

	// EnrichNodeWithCache attaches plugin-specific data using cached results
	// from the filtering phase to avoid expensive re-computation
	// The pluginResults contain the data gathered during plugin filtering
	EnrichNodeWithCache(fs afero.Fs, node *types.Node, pluginResults []*Result) error
}

// DataEnrichmentMap represents the data that a plugin wants to attach to nodes
// Maps file paths (relative to root) to the data that should be attached to those nodes
type DataEnrichmentMap map[string]interface{}

// CacheMap represents cached data that can be passed between plugin phases
// This is a generic storage that treex doesn't interpret - purely for plugin internal use
type CacheMap map[string]interface{}

// DataPluginV2 extends Plugin with node data enrichment capabilities using map-based approach
// This replaces the current DataPlugin interface to eliminate direct node manipulation
type DataPluginV2 interface {
	Plugin

	// EnrichData returns a map of file paths to data that should be attached to nodes
	// The cache parameter contains any cached data from the filtering phase (empty if none available)
	// This eliminates the need for separate EnrichNode and EnrichNodeWithCache methods
	//
	// Parameters:
	//   fs: The filesystem interface for reading files
	//   rootPath: The root directory being processed
	//   filePaths: List of file paths that need enrichment (relative to rootPath)
	//   cache: Cached data from filtering phase (empty map if no cache available)
	//
	// Returns:
	//   DataEnrichmentMap: Maps file paths to data that should be attached to nodes
	//   error: Any error that occurred during enrichment
	//
	// Design principles:
	//   - Plugin returns data, treex handles node attachment
	//   - Plugin doesn't know about Node structure or internals
	//   - Single method handles both cached and non-cached scenarios
	//   - Plugin decides how to use (or ignore) the cache parameter
	EnrichData(fs afero.Fs, rootPath string, filePaths []string, cache CacheMap) (DataEnrichmentMap, error)
}

// Result represents the output of a plugin's processing
// Contains categorized file paths that the orchestrator can use for filtering
type Result struct {
	// PluginName identifies which plugin produced this result
	PluginName string

	// RootPath is the directory that was processed
	RootPath string

	// Categories maps filter names to lists of file paths
	// For example, git plugin might have: "staged", "unstaged", "untracked"
	// Paths are relative to the root directory
	Categories map[string][]string

	// Metadata can store additional information about the processing
	// For example, git branch name, commit hash, etc.
	Metadata map[string]interface{}

	// Cache is a generic storage for plugin-specific data that can be reused
	// during the data enrichment phase. Treex doesn't interpret this data -
	// it's purely for the plugin's internal use to avoid expensive re-computation
	Cache map[string]interface{}
}

// Registry manages all available plugins
// Plugins register themselves at initialization time
type Registry struct {
	plugins map[string]Plugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry
// Returns error if a plugin with the same name is already registered
func (r *Registry) Register(plugin Plugin) error {
	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q is already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// GetPlugin retrieves a plugin by name
// Returns nil if plugin is not found
func (r *Registry) GetPlugin(name string) Plugin {
	return r.plugins[name]
}

// ListPlugins returns the names of all registered plugins
func (r *Registry) ListPlugins() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetAllPlugins returns all registered plugins
// Useful for operations that need to process all plugins
func (r *Registry) GetAllPlugins() []Plugin {
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// Engine orchestrates plugin execution across multiple roots
// It coordinates the work between plugins and handles parallel processing
type Engine struct {
	registry *Registry
	fs       afero.Fs
}

// NewEngine creates a new plugin engine with the given registry and filesystem
func NewEngine(registry *Registry, fs afero.Fs) *Engine {
	return &Engine{
		registry: registry,
		fs:       fs,
	}
}

// ProcessOptions configures how the engine processes plugins
type ProcessOptions struct {
	// SearchRoot is the base directory to search for plugin roots
	SearchRoot string

	// EnabledPlugins lists which plugins to run (empty means all)
	EnabledPlugins []string

	// Parallel enables concurrent plugin processing (per root)
	Parallel bool
}

// ProcessResults aggregates all plugin results
type ProcessResults struct {
	// Results maps plugin names to their processing results
	// Each plugin can have multiple results (one per root)
	Results map[string][]*Result

	// Errors contains any errors encountered during processing
	// Maps plugin names to their errors
	Errors map[string]error
}

// Process executes all enabled plugins and returns aggregated results
// This is the main entry point for Phase 4 plugin filtering
func (e *Engine) Process(opts ProcessOptions) (*ProcessResults, error) {
	results := &ProcessResults{
		Results: make(map[string][]*Result),
		Errors:  make(map[string]error),
	}

	// Determine which plugins to run
	pluginsToRun := e.getPluginsToRun(opts.EnabledPlugins)
	if len(pluginsToRun) == 0 {
		return results, nil // No plugins to run
	}

	// Process each plugin
	for _, plugin := range pluginsToRun {
		pluginName := plugin.Name()

		// Find roots for this plugin
		roots, err := plugin.FindRoots(e.fs, opts.SearchRoot)
		if err != nil {
			results.Errors[pluginName] = fmt.Errorf("failed to find roots: %w", err)
			continue
		}

		// Process each root found by this plugin
		var pluginResults []*Result
		for _, root := range roots {
			result, err := plugin.ProcessRoot(e.fs, root)
			if err != nil {
				results.Errors[pluginName] = fmt.Errorf("failed to process root %q: %w", root, err)
				break // Stop processing this plugin on first error
			}
			if result != nil {
				pluginResults = append(pluginResults, result)
			}
		}

		// Store results for this plugin
		if len(pluginResults) > 0 {
			results.Results[pluginName] = pluginResults
		}
	}

	return results, nil
}

// getPluginsToRun determines which plugins should be executed based on options
func (e *Engine) getPluginsToRun(enabledPlugins []string) []Plugin {
	// If no specific plugins are enabled, run all registered plugins
	if len(enabledPlugins) == 0 {
		return e.registry.GetAllPlugins()
	}

	// Run only the specifically enabled plugins
	var plugins []Plugin
	for _, pluginName := range enabledPlugins {
		if plugin := e.registry.GetPlugin(pluginName); plugin != nil {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// DefaultRegistry provides a global registry that plugins can register with
// This is initialized with common plugins and can be extended
var DefaultRegistry = NewRegistry()

// RegisterPlugin is a convenience function to register with the default registry
func RegisterPlugin(plugin Plugin) error {
	return DefaultRegistry.Register(plugin)
}

// GetDefaultRegistry returns the default global registry
func GetDefaultRegistry() *Registry {
	return DefaultRegistry
}

// GetPlugins returns all registered plugins from the registry
// This is a convenience method for accessing all plugins
func (r *Registry) GetPlugins() []Plugin {
	return r.GetAllPlugins()
}

// NewDefaultEngine creates an engine using the default registry
func NewDefaultEngine(fs afero.Fs) *Engine {
	return NewEngine(DefaultRegistry, fs)
}
