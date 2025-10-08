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
	ExcludeGlobs    []string // User-specified exclude patterns
	IncludeHidden   bool     // Whether to include hidden files (default: true)
	DirectoriesOnly bool     // Whether to show directories only (default: false)
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

	// Phase 1: Pattern Matching - Build composite filter if filtering is needed
	var compositeFilter *pattern.CompositeFilter
	if len(config.ExcludeGlobs) > 0 || !config.IncludeHidden {
		filterBuilder := pattern.NewFilterBuilder(config.Filesystem)

		// Add user exclude patterns
		if len(config.ExcludeGlobs) > 0 {
			filterBuilder.AddUserExcludes(config.ExcludeGlobs)
		}

		// Add hidden file filtering
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

	// Phase 4: Plugin Processing - Apply any plugins (for future implementation)
	// Currently plugins are not integrated with basic tree building
	// This will be expanded in Phase 4 of the architecture plan

	// Calculate statistics
	stats := calculateStats(pathInfos)

	return &TreeResult{
		Root:          root,
		Stats:         stats,
		PluginResults: make(map[string][]*plugins.Result),
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

// DefaultTreeConfig returns a TreeConfig with sensible defaults
func DefaultTreeConfig(root string) TreeConfig {
	return TreeConfig{
		Root:            root,
		Filesystem:      nil,        // Will use OS filesystem
		MaxDepth:        0,          // No depth limit
		ExcludeGlobs:    []string{}, // No excludes by default
		IncludeHidden:   true,       // Show hidden files by default (as per options.txt)
		DirectoriesOnly: false,      // Show both files and directories by default
	}
}
