package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adebert/treex/internal/info"
)

// MAX_FILES_PER_DIR limits the number of unannotated files shown per directory
const MAX_FILES_PER_DIR = 10

// Node represents a file or directory in the tree
type Node struct {
	Name        string            // Just the filename/dirname
	Path        string            // Full path from root
	RelativePath string           // Path relative to the tree root
	IsDir       bool              // Whether this is a directory
	Annotation  *info.Annotation  // Associated annotation if any
	Children    []*Node           // Child nodes (for directories)
	Parent      *Node             // Parent node (nil for root)
}

// Builder handles building file trees with annotations
type Builder struct {
	rootPath      string
	annotations   map[string]*info.Annotation
	ignoreMatcher *IgnoreMatcher
	maxDepth      int
}

// NewBuilder creates a new tree builder
func NewBuilder(rootPath string, annotations map[string]*info.Annotation) *Builder {
	return &Builder{
		rootPath:      rootPath,
		annotations:   annotations,
		ignoreMatcher: nil, // No filtering by default
		maxDepth:      -1,  // No depth limit by default
	}
}

// NewBuilderWithIgnore creates a new tree builder with ignore file support
func NewBuilderWithIgnore(rootPath string, annotations map[string]*info.Annotation, ignoreFilePath string) (*Builder, error) {
	ignoreMatcher, err := NewIgnoreMatcher(ignoreFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ignore file: %w", err)
	}
	
	return &Builder{
		rootPath:      rootPath,
		annotations:   annotations,
		ignoreMatcher: ignoreMatcher,
		maxDepth:      -1, // No depth limit by default
	}, nil
}

// NewBuilderWithOptions creates a new tree builder with all options
func NewBuilderWithOptions(rootPath string, annotations map[string]*info.Annotation, ignoreFilePath string, maxDepth int) (*Builder, error) {
	var ignoreMatcher *IgnoreMatcher
	var err error
	
	if ignoreFilePath != "" {
		ignoreMatcher, err = NewIgnoreMatcher(ignoreFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load ignore file: %w", err)
		}
	}
	
	return &Builder{
		rootPath:      rootPath,
		annotations:   annotations,
		ignoreMatcher: ignoreMatcher,
		maxDepth:      maxDepth,
	}, nil
}

// Build constructs the file tree starting from the root path
func (b *Builder) Build() (*Node, error) {
	// Clean the root path
	absRoot, err := filepath.Abs(b.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create the root node
	rootInfo, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root path: %w", err)
	}

	rootNode := &Node{
		Name:         filepath.Base(absRoot),
		Path:         absRoot,
		RelativePath: ".",
		IsDir:        rootInfo.IsDir(),
		Annotation:   b.getAnnotation("."),
		Children:     []*Node{},
		Parent:       nil,
	}

	// If root is a directory, build its children (starting at depth 0)
	if rootNode.IsDir {
		if err := b.buildChildren(rootNode, 0); err != nil {
			return nil, fmt.Errorf("failed to build children: %w", err)
		}
	}

	return rootNode, nil
}

// buildChildren recursively builds child nodes for a directory
func (b *Builder) buildChildren(parent *Node, depth int) error {
	entries, err := os.ReadDir(parent.Path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", parent.Path, err)
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir() // directories first
		}
		return entries[i].Name() < entries[j].Name()
	})

	// Separate entries into annotated and unannotated groups
	var annotatedEntries []os.DirEntry
	var unannotatedEntries []os.DirEntry
	var filteredEntries []os.DirEntry

	for _, entry := range entries {
		// Skip hidden files and directories (starting with .)
		// except for explicitly annotated paths
		if strings.HasPrefix(entry.Name(), ".") {
			// Always skip .info files as they're metadata
			if entry.Name() == ".info" {
				continue
			}
			
			// Check if this hidden file/dir has an annotation
			relativePath := filepath.Join(parent.RelativePath, entry.Name())
			if parent.RelativePath == "." {
				relativePath = entry.Name()
			}
			if _, hasAnnotation := b.annotations[relativePath]; !hasAnnotation {
				continue // Skip hidden files without annotations
			}
		}

		childRelativePath := filepath.Join(parent.RelativePath, entry.Name())
		
		// Normalize relative path for root directory
		if parent.RelativePath == "." {
			childRelativePath = entry.Name()
		}

		// Check ignore patterns if ignore matcher is configured
		if b.ignoreMatcher != nil && b.ignoreMatcher.ShouldIgnore(childRelativePath, entry.IsDir()) {
			// Skip ignored files unless they have annotations
			if _, hasAnnotation := b.annotations[childRelativePath]; !hasAnnotation {
				continue
			}
		}

		// Check if entry has annotation
		if _, hasAnnotation := b.annotations[childRelativePath]; hasAnnotation {
			annotatedEntries = append(annotatedEntries, entry)
		} else {
			unannotatedEntries = append(unannotatedEntries, entry)
		}
	}

	// Add all annotated entries (always show these)
	filteredEntries = append(filteredEntries, annotatedEntries...)

	// Add unannotated entries up to the limit
	unannotatedCount := len(unannotatedEntries)
	if unannotatedCount <= MAX_FILES_PER_DIR {
		// Under the limit, add all unannotated entries
		filteredEntries = append(filteredEntries, unannotatedEntries...)
	} else {
		// Over the limit, add only MAX_FILES_PER_DIR entries
		filteredEntries = append(filteredEntries, unannotatedEntries[:MAX_FILES_PER_DIR]...)
	}

	// Build child nodes from filtered entries
	for _, entry := range filteredEntries {
		childPath := filepath.Join(parent.Path, entry.Name())
		childRelativePath := filepath.Join(parent.RelativePath, entry.Name())
		
		// Normalize relative path for root directory
		if parent.RelativePath == "." {
			childRelativePath = entry.Name()
		}

		childNode := &Node{
			Name:         entry.Name(),
			Path:         childPath,
			RelativePath: childRelativePath,
			IsDir:        entry.IsDir(),
			Annotation:   b.getAnnotation(childRelativePath),
			Children:     []*Node{},
			Parent:       parent,
		}

		parent.Children = append(parent.Children, childNode)

		// Recursively build children for directories
		if entry.IsDir() {
			// Only recurse if the child's children would not exceed maxDepth
			if b.maxDepth == -1 || depth+1 < b.maxDepth {
				if err := b.buildChildren(childNode, depth+1); err != nil {
					return err
				}
			}
		}
	}

	// Add "more files..." indicator if we exceeded the limit
	if unannotatedCount > MAX_FILES_PER_DIR {
		hiddenCount := unannotatedCount - MAX_FILES_PER_DIR
		moreFilesNode := &Node{
			Name:         fmt.Sprintf("... %d more files not shown", hiddenCount),
			Path:         "",
			RelativePath: "",
			IsDir:        false,
			Annotation:   nil,
			Children:     []*Node{},
			Parent:       parent,
		}
		parent.Children = append(parent.Children, moreFilesNode)
	}

	return nil
}

// getAnnotation retrieves the annotation for a given relative path
func (b *Builder) getAnnotation(relativePath string) *info.Annotation {
	// Try exact match first
	if annotation, exists := b.annotations[relativePath]; exists {
		return annotation
	}

	// Try with forward slashes (in case of path separator differences)
	normalizedPath := filepath.ToSlash(relativePath)
	if annotation, exists := b.annotations[normalizedPath]; exists {
		return annotation
	}

	return nil
}

// BuildTree is a convenience function that combines parsing and building (single .info file)
func BuildTree(rootPath string) (*Node, error) {
	// Parse annotations from the root directory only
	annotations, err := info.ParseDirectory(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotations: %w", err)
	}

	// Build the tree
	builder := NewBuilder(rootPath, annotations)
	return builder.Build()
}

// BuildTreeNested is a convenience function that combines nested parsing and building
// This looks for .info files in all subdirectories and merges their annotations
func BuildTreeNested(rootPath string) (*Node, error) {
	// Parse annotations from the entire directory tree
	annotations, err := info.ParseDirectoryTree(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nested annotations: %w", err)
	}

	// Build the tree
	builder := NewBuilder(rootPath, annotations)
	return builder.Build()
}

// BuildTreeNestedWithIgnore is a convenience function that combines nested parsing and building with ignore support
func BuildTreeNestedWithIgnore(rootPath, ignoreFilePath string) (*Node, error) {
	// Parse annotations from the entire directory tree
	annotations, err := info.ParseDirectoryTree(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nested annotations: %w", err)
	}

	// Build the tree with ignore support
	builder, err := NewBuilderWithIgnore(rootPath, annotations, ignoreFilePath)
	if err != nil {
		return nil, err
	}
	
	return builder.Build()
}

// BuildTreeNestedWithOptions is a convenience function that combines nested parsing and building with all options
func BuildTreeNestedWithOptions(rootPath, ignoreFilePath string, maxDepth int) (*Node, error) {
	// Parse annotations from the entire directory tree
	annotations, err := info.ParseDirectoryTree(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nested annotations: %w", err)
	}

	// Build the tree with all options
	builder, err := NewBuilderWithOptions(rootPath, annotations, ignoreFilePath, maxDepth)
	if err != nil {
		return nil, err
	}
	
	return builder.Build()
}

// WalkTree traverses the tree and calls the provided function for each node
func WalkTree(root *Node, fn func(*Node, int) error) error {
	return walkTree(root, 0, fn)
}

// walkTree is the internal recursive implementation
func walkTree(node *Node, depth int, fn func(*Node, int) error) error {
	if err := fn(node, depth); err != nil {
		return err
	}
	
	for _, child := range node.Children {
		if err := walkTree(child, depth+1, fn); err != nil {
			return err
		}
	}
	
	return nil
}

// DisplayConfig holds configuration for displaying annotated trees
type DisplayConfig struct {
	Verbose    bool
	NoColor    bool
	Minimal    bool
	IgnoreFile string
	MaxDepth   int
	SafeMode   bool
} 