// Package treex provides shell-agnostic core API functions for tree building and manipulation.
// All functions accept structured input and return structured output without performing I/O.
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

// TreeStats provides statistics about tree building operations
type TreeStats struct {
	TotalFiles       int // Total number of files processed
	TotalDirectories int // Total number of directories processed
	FilteredOut      int // Number of items filtered out
	MaxDepthReached  int // Maximum depth reached during traversal
}

// BuildTree constructs a file tree based on the provided configuration.
// This is the main API function for tree building operations.
func BuildTree(config TreeConfig) (*TreeResult, error) {
	// Use default filesystem if none provided
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

	// Phase 3: Tree Construction - Build node tree from paths
	constructor := treeconstruction.NewConstructor()
	rootNode := constructor.BuildTree(pathInfos)

	// Calculate basic statistics
	stats := calculateStats(pathInfos)

	// Phase 4: Plugin Filtering - Skip for now
	// Phase 5: User Queries - Skip for now

	return &TreeResult{
		Root:          rootNode,
		Stats:         stats,
		PluginResults: make(map[string][]*plugins.Result), // Empty for now
	}, nil
}

// calculateStats computes statistics from the collected path information
func calculateStats(pathInfos []pathcollection.PathInfo) TreeStats {
	stats := TreeStats{}

	for _, info := range pathInfos {
		if info.IsDir {
			stats.TotalDirectories++
		} else {
			stats.TotalFiles++
		}

		if info.Depth > stats.MaxDepthReached {
			stats.MaxDepthReached = info.Depth
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
