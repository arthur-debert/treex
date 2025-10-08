// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection

import (
	"github.com/spf13/afero"
	"github.com/jwaldrip/treex/treex/pattern"
)

// OptionsConfigurator provides a fluent interface for configuring path collection
type OptionsConfigurator struct {
	fs      afero.Fs
	options CollectionOptions
}

// NewConfigurator creates a new path collection options configurator
func NewConfigurator(fs afero.Fs) *OptionsConfigurator {
	return &OptionsConfigurator{
		fs: fs,
		options: CollectionOptions{
			MaxDepth: 0, // 0 means no limit
		},
	}
}

// WithRoot sets the root directory for collection
func (c *OptionsConfigurator) WithRoot(root string) *OptionsConfigurator {
	c.options.Root = root
	return c
}

// WithMaxDepth sets the maximum traversal depth
// Depth 0 is the root directory, depth 1 are immediate children, etc.
// Setting maxDepth to 0 removes depth limiting
func (c *OptionsConfigurator) WithMaxDepth(maxDepth int) *OptionsConfigurator {
	c.options.MaxDepth = maxDepth
	return c
}

// WithFilter sets the pattern filter for exclusion
func (c *OptionsConfigurator) WithFilter(filter *pattern.CompositeFilter) *OptionsConfigurator {
	c.options.Filter = filter
	return c
}

// WithDirsOnly configures collection to include only directories
func (c *OptionsConfigurator) WithDirsOnly() *OptionsConfigurator {
	c.options.DirsOnly = true
	c.options.FilesOnly = false // ensure mutual exclusion
	return c
}

// WithFilesOnly configures collection to include only files
func (c *OptionsConfigurator) WithFilesOnly() *OptionsConfigurator {
	c.options.FilesOnly = true
	c.options.DirsOnly = false // ensure mutual exclusion
	return c
}

// WithLogger sets a custom logger for error reporting during collection
func (c *OptionsConfigurator) WithLogger(logger Logger) *OptionsConfigurator {
	c.options.Logger = logger
	return c
}

// NewCollector creates and returns a configured collector
func (c *OptionsConfigurator) NewCollector() *Collector {
	return NewCollector(c.fs, c.options)
}

// Collect is a convenience method that creates a collector and immediately runs collection
func (c *OptionsConfigurator) Collect() ([]PathInfo, error) {
	collector := c.NewCollector()
	return collector.Collect()
}