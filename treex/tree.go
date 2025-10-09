// Package treex provides core tree building functionality.
package treex

import (
	"path/filepath"

	"github.com/jwaldrip/treex/treex/pathcollection"
	"github.com/jwaldrip/treex/treex/pattern"
	"github.com/jwaldrip/treex/treex/plugins"
	"github.com/jwaldrip/treex/treex/treeconstruction"
	"github.com/jwaldrip/treex/treex/types"
	"github.com/spf13/afero"
)

// TreeConfig represents configuration for tree building operations
type TreeConfig struct {
	// Root directory to start tree building from
	Root string

	// Filesystem interface (allows for testing with mock filesystems)
	Filesystem afero.Fs

	// Basic options (start simple as instructed)
	MaxDepth int // Maximum depth to traverse (0 = no limit)

	// Path filtering options (added incrementally)
	// Multiple exclusion mechanisms work together:
	// 1. BuiltinIgnores - default patterns for VCS/build artifacts (can be disabled)
	// 2. ExcludeGlobs - user-specified patterns via --exclude
	// 3. Gitignore files - .gitignore pattern support
	// 4. IncludeHidden - hidden file visibility control
	// 5. PluginFilters - filter by plugin categories (e.g., --git-staged, --info-annotated)
	BuiltinIgnores  bool                       // Whether to apply built-in ignore patterns (default: true)
	ExcludeGlobs    []string                   // User-specified exclude patterns
	IncludeHidden   bool                       // Whether to include hidden files (default: true)
	DirectoriesOnly bool                       // Whether to show directories only (default: false)
	PluginFilters   map[string]map[string]bool // Plugin category filters: plugin -> category -> enabled
}

// TreeResult represents the result of tree building operations
type TreeResult struct {
	// Root node of the built tree
	Root *types.Node

	// Statistics about the tree building process
	Stats TreeStats

	// Plugin results (if any plugins were applied)
	PluginResults map[string][]*plugins.Result
}

// TreeStats provides statistics about the tree building process
type TreeStats struct {
	TotalFiles       int
	TotalDirectories int
	MaxDepthReached  int
	FilteredOut      int // Number of files/directories filtered out
}

// BuildTree constructs a file tree based on the provided configuration.
// This is the main tree building function that orchestrates the entire process.
func BuildTree(config TreeConfig) (*TreeResult, error) {
	// Set default filesystem if not provided
	if config.Filesystem == nil {
		config.Filesystem = afero.NewOsFs()
	}

	// Phase 1: Pattern Matching - Build composite filter combining multiple exclusion mechanisms
	// This coordinates: built-in ignores, user excludes, gitignore files, and hidden file filtering
	var compositeFilter *pattern.CompositeFilter
	if config.BuiltinIgnores || len(config.ExcludeGlobs) > 0 || !config.IncludeHidden {
		filterBuilder := pattern.NewFilterBuilder(config.Filesystem)

		// 1. Add built-in ignore patterns (VCS dirs, build artifacts, etc.)
		filterBuilder.AddBuiltinIgnores(config.BuiltinIgnores)

		// 2. Add user exclude patterns (--exclude flags)
		if len(config.ExcludeGlobs) > 0 {
			filterBuilder.AddUserExcludes(config.ExcludeGlobs)
		}

		// 3. Add gitignore support (automatic .gitignore detection)
		filterBuilder.AddGitignore(".gitignore", false) // TODO: Make gitignore configurable

		// 4. Add hidden file filtering (--hidden flag control)
		filterBuilder.AddHiddenFilter(config.IncludeHidden)

		compositeFilter = filterBuilder.Build()
	}

	// Phase 2: Path Collection - Basic collection with depth limit and optional filtering
	collector := pathcollection.NewConfigurator(config.Filesystem).
		WithRoot(config.Root).
		WithMaxDepth(config.MaxDepth)

	if compositeFilter != nil {
		collector = collector.WithFilter(compositeFilter)
	}

	// Apply directories only filter if requested
	if config.DirectoriesOnly {
		collector = collector.WithDirsOnly()
	}

	// Phase 3: Plugin Filtering - Apply plugin filtering during path collection
	pluginResults := make(map[string][]*plugins.Result)
	if len(config.PluginFilters) > 0 {
		pluginFilter, results, err := createPluginFilter(config.Filesystem, config.Root, config.PluginFilters)
		if err != nil {
			return nil, err
		}
		if pluginFilter != nil {
			collector = collector.WithFilter(pluginFilter)
		}
		pluginResults = results
	}

	pathInfos, err := collector.Collect()
	if err != nil {
		return nil, err
	}

	// Phase 4: Tree Construction - Build tree structure from collected paths
	constructor := treeconstruction.NewConstructor()
	root := constructor.BuildTree(pathInfos)

	// Phase 5: Data Enrichment - Enrich surviving nodes with plugin data
	// This runs after filtering to avoid expensive operations on filtered-out files
	err = applyDataEnrichment(config.Filesystem, root, pluginResults)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	stats := calculateStats(pathInfos)

	return &TreeResult{
		Root:          root,
		Stats:         stats,
		PluginResults: pluginResults,
	}, nil
}

// calculateStats computes statistics about the collected paths
func calculateStats(pathInfos []pathcollection.PathInfo) TreeStats {
	stats := TreeStats{}

	for _, pathInfo := range pathInfos {
		if pathInfo.IsDir {
			stats.TotalDirectories++
		} else {
			stats.TotalFiles++
		}

		if pathInfo.Depth > stats.MaxDepthReached {
			stats.MaxDepthReached = pathInfo.Depth
		}
	}

	return stats
}

// createPluginFilter creates a filter that includes only paths matching plugin categories
// Returns the filter and plugin results for metadata
func createPluginFilter(fs afero.Fs, rootPath string, pluginFilters map[string]map[string]bool) (*pattern.CompositeFilter, map[string][]*plugins.Result, error) {
	registry := plugins.GetDefaultRegistry()
	pluginResults := make(map[string][]*plugins.Result)
	allowedPaths := make(map[string]bool)

	// Collect plugin results for active filters
	for pluginName := range pluginFilters {
		plugin := registry.GetPlugin(pluginName)
		if plugin == nil {
			continue // Skip if plugin not found
		}

		// Find plugin roots
		roots, err := plugin.FindRoots(fs, rootPath)
		if err != nil {
			continue // Skip plugin on error
		}

		// Process each root
		for _, pluginRoot := range roots {
			// Make plugin root absolute by joining with search root
			var absolutePluginRoot string
			if pluginRoot != "." {
				absolutePluginRoot = filepath.Join(rootPath, pluginRoot)
			} else {
				absolutePluginRoot = rootPath
			}

			result, err := plugin.ProcessRoot(fs, absolutePluginRoot)
			if err != nil {
				continue // Skip this root on error
			}

			pluginResults[pluginName] = append(pluginResults[pluginName], result)

			// Add matching files to allowed paths
			for categoryName, enabled := range pluginFilters[pluginName] {
				if !enabled {
					continue
				}

				if files, exists := result.Categories[categoryName]; exists {
					for _, filePath := range files {
						allowedPaths[filePath] = true
					}
				}
			}
		}
	}

	// If no paths matched filters, return nil filter (no filtering)
	if len(allowedPaths) == 0 {
		return nil, pluginResults, nil
	}

	// Create a filter that only allows matching paths and their parent directories
	filterBuilder := pattern.NewFilterBuilder(fs)
	filterBuilder.AddPluginFilter(allowedPaths)
	pluginFilter := filterBuilder.Build()

	return pluginFilter, pluginResults, nil
}

// applyDataEnrichment enriches tree nodes with plugin data
// Runs through all registered DataPlugin implementations and enriches matching nodes
// Uses cached plugin results when available to avoid expensive re-computation
func applyDataEnrichment(fs afero.Fs, root *types.Node, pluginResults map[string][]*plugins.Result) error {
	if root == nil {
		return nil
	}

	registry := plugins.GetDefaultRegistry()
	registeredPlugins := registry.GetPlugins()

	// Collect all DataPlugin implementations
	var dataPlugins []plugins.DataPlugin
	cachedDataPlugins := make(map[string]plugins.CachedDataPlugin)

	for _, plugin := range registeredPlugins {
		if dataPlugin, ok := plugin.(plugins.DataPlugin); ok {
			dataPlugins = append(dataPlugins, dataPlugin)

			// Check if this plugin also supports cached enrichment
			if cachedDataPlugin, ok := plugin.(plugins.CachedDataPlugin); ok {
				cachedDataPlugins[plugin.Name()] = cachedDataPlugin
			}
		}
	}

	// Apply data enrichment to the tree
	return enrichNodeRecursively(fs, root, dataPlugins, cachedDataPlugins, pluginResults)
}

// enrichNodeRecursively applies data enrichment to a node and all its children
func enrichNodeRecursively(fs afero.Fs, node *types.Node, dataPlugins []plugins.DataPlugin, cachedDataPlugins map[string]plugins.CachedDataPlugin, pluginResults map[string][]*plugins.Result) error {
	if node == nil {
		return nil
	}

	// Enrich this node with data from all DataPlugin implementations
	for _, dataPlugin := range dataPlugins {
		pluginName := dataPlugin.Name()

		// Check if we have cached results for this plugin and it supports cached enrichment
		if cachedDataPlugin, hasCached := cachedDataPlugins[pluginName]; hasCached {
			if results, hasResults := pluginResults[pluginName]; hasResults {
				// Use cached enrichment for better performance
				err := cachedDataPlugin.EnrichNodeWithCache(fs, node, results)
				if err != nil {
					// Log error but continue with other plugins
					// TODO: Add proper logging when available
					continue
				}
			} else {
				// No cached results, fall back to regular enrichment
				err := dataPlugin.EnrichNode(fs, node)
				if err != nil {
					// Log error but continue with other plugins
					// TODO: Add proper logging when available
					continue
				}
			}
		} else {
			// Regular enrichment for plugins without cache support
			err := dataPlugin.EnrichNode(fs, node)
			if err != nil {
				// Log error but continue with other plugins
				// TODO: Add proper logging when available
				continue
			}
		}
	}

	// Recursively enrich children
	for _, child := range node.Children {
		err := enrichNodeRecursively(fs, child, dataPlugins, cachedDataPlugins, pluginResults)
		if err != nil {
			return err
		}
	}

	return nil
}

// DefaultTreeConfig returns a TreeConfig with sensible defaults
func DefaultTreeConfig(root string) TreeConfig {
	return TreeConfig{
		Root:            root,
		Filesystem:      nil,                              // Will use OS filesystem
		MaxDepth:        0,                                // No depth limit
		BuiltinIgnores:  true,                             // Enable built-in ignores by default (.git, node_modules, etc.)
		ExcludeGlobs:    []string{},                       // No user excludes by default
		IncludeHidden:   true,                             // Show hidden files by default (as per options.txt)
		DirectoriesOnly: false,                            // Show both files and directories by default
		PluginFilters:   make(map[string]map[string]bool), // No plugin filters by default
	}
}
