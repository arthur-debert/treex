// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"treex/treex/pattern"
)

// PathInfo represents collected information about a file or directory
type PathInfo struct {
	Path         string // Relative path from root
	AbsolutePath string // Absolute filesystem path
	IsDir        bool   // True if this is a directory
	Size         int64  // File size in bytes (0 for directories)
	Depth        int    // Depth from collection root (root = 0)
}

// Logger interface for error reporting during path collection
type Logger interface {
	Printf(format string, v ...interface{})
}

// CollectionOptions configures the path collection process
type CollectionOptions struct {
	Root      string                   // Root directory to start collection from
	MaxDepth  int                      // Maximum depth to traverse (0 = no limit)
	Filter    *pattern.CompositeFilter // Pattern filter for early pruning
	DirsOnly  bool                     // If true, collect only directories
	FilesOnly bool                     // If true, collect only files
	Logger    Logger                   // Optional logger for error reporting (uses log.Printf if nil)
}

// Collector handles filesystem traversal with early pruning
type Collector struct {
	fs      afero.Fs
	options CollectionOptions
	results []PathInfo
}

// NewCollector creates a new path collector
func NewCollector(fs afero.Fs, options CollectionOptions) *Collector {
	return &Collector{
		fs:      fs,
		options: options,
		results: make([]PathInfo, 0),
	}
}

// Collect performs the filesystem walk and returns collected paths
func (c *Collector) Collect() ([]PathInfo, error) {
	// Reset results for fresh collection
	c.results = make([]PathInfo, 0)

	// Convert root to absolute path for consistent handling
	absRoot, err := filepath.Abs(c.options.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for root %q: %w", c.options.Root, err)
	}

	// Ensure root directory exists
	rootInfo, err := c.fs.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("root directory %q not accessible: %w", absRoot, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("root %q is not a directory", absRoot)
	}

	// Start filesystem walk from root
	err = afero.Walk(c.fs, absRoot, func(path string, info fs.FileInfo, err error) error {
		return c.walkFunc(absRoot, path, info, err)
	})

	if err != nil {
		return nil, fmt.Errorf("filesystem walk failed: %w", err)
	}

	return c.results, nil
}

// logf logs a message using the configured logger, or log.Printf if no logger is set
func (c *Collector) logf(format string, v ...interface{}) {
	if c.options.Logger != nil {
		c.options.Logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// walkFunc is called for each file/directory during filesystem traversal
func (c *Collector) walkFunc(rootPath, currentPath string, info fs.FileInfo, err error) error {
	// Handle errors encountered during traversal
	// According to architecture.txt: "Permission errors during walk: log and continue"
	if err != nil {
		// Log the error but continue traversal for robustness
		// This handles permission errors, broken symlinks, etc.
		c.logf("pathcollection: skipping path %q due to error: %v", currentPath, err)
		return nil
	}

	// Calculate relative path from root
	// Example: if root="/home/user/project" and currentPath="/home/user/project/src/main.go"
	// then relativePath="src/main.go"
	relativePath, err := filepath.Rel(rootPath, currentPath)
	if err != nil {
		// This shouldn't happen in normal cases, but handle gracefully
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Handle root directory specially - use "." as canonical root path
	// This avoids ambiguity between empty string and actual root directory
	if relativePath == "." {
		relativePath = "."
	}

	// Calculate depth from root
	// Root directory has depth 0, immediate children have depth 1, etc.
	depth := 0
	if relativePath != "." && relativePath != "" {
		// Count path separators to determine depth
		// Examples: "file.txt" = 0 separators = depth 1
		//          "src/main.go" = 1 separator = depth 2
		//          "src/lib/helper.go" = 2 separators = depth 3
		depth = strings.Count(relativePath, string(filepath.Separator)) + 1
	}

	// Apply depth limiting BEFORE other checks for efficiency
	// If we're beyond max depth and this is a directory, skip entire subtree
	if c.options.MaxDepth > 0 && depth > c.options.MaxDepth {
		if info.IsDir() {
			// Return filepath.SkipDir to prevent recursion into this directory
			// This is the key optimization - we don't traverse deeper than needed
			return filepath.SkipDir
		}
		// For files beyond max depth, just skip but don't affect directory traversal
		return nil
	}

	// Apply pattern filtering with early pruning
	if c.options.Filter != nil && c.options.Filter.ShouldExclude(relativePath, info.IsDir()) {
		if info.IsDir() {
			// CRITICAL: If a directory is excluded by patterns (e.g., "node_modules", ".git")
			// we must return filepath.SkipDir to prevent traversing into it
			// This implements the "early pruning" strategy - we don't waste time
			// walking through thousands of files in excluded directories
			return filepath.SkipDir
		}
		// For excluded files, just skip but continue directory traversal
		return nil
	}

	// Apply file/directory type filtering
	if c.options.DirsOnly && !info.IsDir() {
		return nil // Skip files when we only want directories
	}
	if c.options.FilesOnly && info.IsDir() {
		return nil // Skip directories when we only want files
	}

	// Collect file size information
	// For directories, size is typically 0 or filesystem-specific block size
	// We normalize directory sizes to 0 for consistency
	size := info.Size()
	if info.IsDir() {
		size = 0
	}

	// Create path info and add to results
	pathInfo := PathInfo{
		Path:         relativePath,
		AbsolutePath: currentPath,
		IsDir:        info.IsDir(),
		Size:         size,
		Depth:        depth,
	}

	c.results = append(c.results, pathInfo)

	// Continue traversal
	return nil
}

// GetPaths returns just the relative paths from collected results
// This is a convenience method for phases that only need path strings
func (c *Collector) GetPaths() []string {
	paths := make([]string, len(c.results))
	for i, result := range c.results {
		paths[i] = result.Path
	}
	return paths
}
