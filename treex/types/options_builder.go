package types

// OptionsBuilder provides a fluent interface for building TreeOptions
type OptionsBuilder struct {
	opts TreeOptions
}

// NewOptionsBuilder creates a new builder with defaults
func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{
		opts: DefaultTreeOptions(),
	}
}

// WithRoot sets the root directory
func (b *OptionsBuilder) WithRoot(root string) *OptionsBuilder {
	b.opts.Root = root
	return b
}

// WithMaxDepth sets the maximum traversal depth
func (b *OptionsBuilder) WithMaxDepth(depth int) *OptionsBuilder {
	b.opts.Tree.MaxDepth = depth
	return b
}

// WithDirsOnly enables directory-only mode
func (b *OptionsBuilder) WithDirsOnly() *OptionsBuilder {
	b.opts.Tree.DirsOnly = true
	return b
}

// WithHidden enables showing hidden files
func (b *OptionsBuilder) WithHidden() *OptionsBuilder {
	b.opts.Tree.ShowHidden = true
	return b
}

// WithExclude adds an exclude pattern
func (b *OptionsBuilder) WithExclude(pattern string) *OptionsBuilder {
	b.opts.Patterns.Excludes = append(b.opts.Patterns.Excludes, pattern)
	return b
}

// WithExcludes adds multiple exclude patterns
func (b *OptionsBuilder) WithExcludes(patterns ...string) *OptionsBuilder {
	b.opts.Patterns.Excludes = append(b.opts.Patterns.Excludes, patterns...)
	return b
}

// WithIgnoreFile sets a custom ignore file path
func (b *OptionsBuilder) WithIgnoreFile(path string) *OptionsBuilder {
	b.opts.Patterns.IgnoreFilePath = path
	return b
}

// WithoutIgnoreFile disables ignore file processing
func (b *OptionsBuilder) WithoutIgnoreFile() *OptionsBuilder {
	b.opts.Patterns.NoIgnoreFile = true
	return b
}

// WithSearch adds search terms
func (b *OptionsBuilder) WithSearch(terms ...string) *OptionsBuilder {
	b.opts.Search = append(b.opts.Search, terms...)
	return b
}

// Build returns the constructed options
func (b *OptionsBuilder) Build() TreeOptions {
	return b.opts
}

// Validate checks if options are valid and fixes defaults
func (opts *TreeOptions) Validate() error {
	if opts.Root == "" {
		opts.Root = "."
	}

	if opts.Tree.MaxDepth <= 0 {
		opts.Tree.MaxDepth = 3
	}

	return nil
}

// ErrInvalidOptions represents an options validation error
type ErrInvalidOptions struct {
	Message string
}

func (e ErrInvalidOptions) Error() string {
	return e.Message
}
