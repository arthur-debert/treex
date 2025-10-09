// Package treex provides core tree building functionality.
package treex

import (
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

	pathInfos, err := collector.Collect()
	if err != nil {
		return nil, err
	}

	// Phase 3: Tree Construction - Build tree structure from collected paths
	constructor := treeconstruction.NewConstructor()
	root := constructor.BuildTree(pathInfos)

	// Phase 4: Plugin Processing - Apply plugin filtering if configured
	pluginResults := make(map[string][]*plugins.Result)
	if len(config.PluginFilters) > 0 {
		filteredRoot, results, err := applyPluginFiltering(config.Filesystem, root, config.PluginFilters)
		if err != nil {
			return nil, err
		}
		root = filteredRoot
		pluginResults = results
	}

	// Phase 5: Data Enrichment - Enrich surviving nodes with plugin data
	// This runs after filtering to avoid expensive operations on filtered-out files
	err = applyDataEnrichment(config.Filesystem, root)
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

// applyPluginFiltering processes the tree with plugin filters to include only matching files
// Returns the filtered tree and plugin results
func applyPluginFiltering(fs afero.Fs, root *types.Node, pluginFilters map[string]map[string]bool) (*types.Node, map[string][]*plugins.Result, error) {
	registry := plugins.GetDefaultRegistry()
	pluginResults := make(map[string][]*plugins.Result)

	// Collect plugin results for active filters
	for pluginName := range pluginFilters {
		plugin := registry.GetPlugin(pluginName)
		if plugin == nil {
			continue // Skip if plugin not found
		}

		// Find plugin roots
		roots, err := plugin.FindRoots(fs, root.Path)
		if err != nil {
			continue // Skip plugin on error
		}

		// Process each root
		for _, pluginRoot := range roots {
			result, err := plugin.ProcessRoot(fs, pluginRoot)
			if err != nil {
				continue // Skip this root on error
			}

			pluginResults[pluginName] = append(pluginResults[pluginName], result)
		}
	}

	// Filter the tree based on plugin results
	filteredRoot := filterTreeByPluginResults(root, pluginFilters, pluginResults)

	return filteredRoot, pluginResults, nil
}

// filterTreeByPluginResults filters tree nodes based on plugin category results
// Only keeps nodes that match the enabled plugin filters
func filterTreeByPluginResults(node *types.Node, pluginFilters map[string]map[string]bool, pluginResults map[string][]*plugins.Result) *types.Node {
	if node == nil {
		return nil
	}

	// Check if this node should be included based on plugin filters
	shouldInclude := false

	// If no plugin filters are active, include all nodes
	if len(pluginFilters) == 0 {
		shouldInclude = true
	} else {
		// Check each plugin filter
		for pluginName, categories := range pluginFilters {
			results, exists := pluginResults[pluginName]
			if !exists {
				continue
			}

			// Check if this node matches any enabled category for this plugin
			for _, result := range results {
				for categoryName, enabled := range categories {
					if !enabled {
						continue
					}

					// Check if node path is in this category
					if files, exists := result.Categories[categoryName]; exists {
						for _, filePath := range files {
							if node.Path == filePath {
								shouldInclude = true
								break
							}
						}
					}
					if shouldInclude {
						break
					}
				}
				if shouldInclude {
					break
				}
			}
			if shouldInclude {
				break
			}
		}
	}

	// If this node should be included, create a copy and filter its children
	if shouldInclude {
		filtered := &types.Node{
			Name:     node.Name,
			Path:     node.Path,
			IsDir:    node.IsDir,
			Children: make([]*types.Node, 0),
			Data:     make(map[string]interface{}),
		}

		// Copy data
		for k, v := range node.Data {
			filtered.Data[k] = v
		}

		// Recursively filter children
		for _, child := range node.Children {
			filteredChild := filterTreeByPluginResults(child, pluginFilters, pluginResults)
			if filteredChild != nil {
				filtered.Children = append(filtered.Children, filteredChild)
			}
		}

		return filtered
	}

	// If this node shouldn't be included, still check children (in case they should be included)
	// This handles cases where parent directories aren't explicitly in categories but children are
	var filteredChildren []*types.Node
	for _, child := range node.Children {
		filteredChild := filterTreeByPluginResults(child, pluginFilters, pluginResults)
		if filteredChild != nil {
			filteredChildren = append(filteredChildren, filteredChild)
		}
	}

	// If we have filtered children but this node isn't explicitly included,
	// include it as a parent container
	if len(filteredChildren) > 0 {
		filtered := &types.Node{
			Name:     node.Name,
			Path:     node.Path,
			IsDir:    node.IsDir,
			Children: filteredChildren,
			Data:     make(map[string]interface{}),
		}

		// Copy data
		for k, v := range node.Data {
			filtered.Data[k] = v
		}

		return filtered
	}

	return nil
}

// applyDataEnrichment enriches tree nodes with plugin data
// Runs through all registered DataPlugin implementations and enriches matching nodes
func applyDataEnrichment(fs afero.Fs, root *types.Node) error {
	if root == nil {
		return nil
	}

	registry := plugins.GetDefaultRegistry()
	registeredPlugins := registry.GetPlugins()

	// Collect all DataPlugin implementations
	var dataPlugins []plugins.DataPlugin
	for _, plugin := range registeredPlugins {
		if dataPlugin, ok := plugin.(plugins.DataPlugin); ok {
			dataPlugins = append(dataPlugins, dataPlugin)
		}
	}

	// Apply data enrichment to the tree
	return enrichNodeRecursively(fs, root, dataPlugins)
}

// enrichNodeRecursively applies data enrichment to a node and all its children
func enrichNodeRecursively(fs afero.Fs, node *types.Node, dataPlugins []plugins.DataPlugin) error {
	if node == nil {
		return nil
	}

	// Enrich this node with data from all DataPlugin implementations
	for _, dataPlugin := range dataPlugins {
		err := dataPlugin.EnrichNode(fs, node)
		if err != nil {
			// Log error but continue with other plugins
			// TODO: Add proper logging when available
			continue
		}
	}

	// Recursively enrich children
	for _, child := range node.Children {
		err := enrichNodeRecursively(fs, child, dataPlugins)
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
