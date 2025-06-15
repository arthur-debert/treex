package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adebert/treex/internal/info"
)

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
	rootPath    string
	annotations map[string]*info.Annotation
}

// NewBuilder creates a new tree builder
func NewBuilder(rootPath string, annotations map[string]*info.Annotation) *Builder {
	return &Builder{
		rootPath:    rootPath,
		annotations: annotations,
	}
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

	// If root is a directory, build its children
	if rootNode.IsDir {
		if err := b.buildChildren(rootNode); err != nil {
			return nil, fmt.Errorf("failed to build children: %w", err)
		}
	}

	return rootNode, nil
}

// buildChildren recursively builds child nodes for a directory
func (b *Builder) buildChildren(parent *Node) error {
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
			if err := b.buildChildren(childNode); err != nil {
				return err
			}
		}
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

// BuildTree is a convenience function that combines parsing and building
func BuildTree(rootPath string) (*Node, error) {
	// Parse annotations from the root directory
	annotations, err := info.ParseDirectory(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse annotations: %w", err)
	}

	// Build the tree
	builder := NewBuilder(rootPath, annotations)
	return builder.Build()
}

// WalkTree performs a depth-first traversal of the tree, calling the visitor function for each node
func WalkTree(root *Node, visitor func(*Node, int) error) error {
	return walkTreeRecursive(root, 0, visitor)
}

func walkTreeRecursive(node *Node, depth int, visitor func(*Node, int) error) error {
	if err := visitor(node, depth); err != nil {
		return err
	}

	for _, child := range node.Children {
		if err := walkTreeRecursive(child, depth+1, visitor); err != nil {
			return err
		}
	}

	return nil
} 