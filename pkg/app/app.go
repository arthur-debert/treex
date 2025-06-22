package app

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
	"github.com/adebert/treex/pkg/tui"
)

// RenderOptions contains configuration options for rendering annotated trees
type RenderOptions struct {
	Verbose    bool
	NoColor    bool
	Minimal    bool
	IgnoreFile string
	MaxDepth   int
	SafeMode   bool
}

// RenderResult contains the rendered output and optional verbose information
type RenderResult struct {
	Output string
	Stats  *RenderStats
}

// RenderStats contains statistics about the rendering process
type RenderStats struct {
	AnnotationsFound int
	TreeGenerated    bool
}

// RenderAnnotatedTree is the main business logic function that generates an annotated tree
// This function handles all the core application logic and returns a complete rendered string
func RenderAnnotatedTree(targetPath string, options RenderOptions) (*RenderResult, error) {
	var outputBuilder strings.Builder
	stats := &RenderStats{}

	if options.Verbose {
		fmt.Fprintf(&outputBuilder, "Analyzing directory: %s\n", targetPath)
		fmt.Fprintln(&outputBuilder, "Verbose mode enabled - will show parsed .info structure")
		fmt.Fprintln(&outputBuilder)
	}

	// Phase 1 - Parse .info files (nested)
	annotations, err := info.ParseDirectoryTree(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .info files: %w", err)
	}

	stats.AnnotationsFound = len(annotations)

	if options.Verbose {
		fmt.Fprintln(&outputBuilder, "=== Parsed Annotations ===")
		if len(annotations) == 0 {
			fmt.Fprintln(&outputBuilder, "No annotations found (no .info file or empty file)")
		} else {
			for path, annotation := range annotations {
				fmt.Fprintf(&outputBuilder, "Path: %s\n", path)
				if annotation.Title != "" {
					fmt.Fprintf(&outputBuilder, "  Title: %s\n", annotation.Title)
				}
				fmt.Fprintf(&outputBuilder, "  Description: %s\n", annotation.Description)
				fmt.Fprintln(&outputBuilder)
			}
		}
		fmt.Fprintln(&outputBuilder, "=== End Annotations ===")
		fmt.Fprintln(&outputBuilder)
	}

	// Phase 2 - Build file tree (using nested annotations with filtering options)
	var root *tree.Node
	if options.IgnoreFile != "" || options.MaxDepth != -1 {
		// Build tree with filtering options
		root, err = tree.BuildTreeNestedWithOptions(targetPath, options.IgnoreFile, options.MaxDepth)
		if err != nil {
			return nil, fmt.Errorf("failed to build file tree with options: %w", err)
		}
	} else {
		// Build tree without filtering
		root, err = tree.BuildTreeNested(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build file tree: %w", err)
		}
	}

	stats.TreeGenerated = true

	if options.Verbose {
		fmt.Fprintln(&outputBuilder, "=== File Tree Structure ===")
		err = tree.WalkTree(root, func(node *tree.Node, depth int) error {
			indent := ""
			for i := 0; i < depth; i++ {
				indent += "  "
			}
			
			nodeType := "file"
			if node.IsDir {
				nodeType = "dir"
			}
			
			annotationInfo := ""
			if node.Annotation != nil {
				if node.Annotation.Title != "" {
					annotationInfo = fmt.Sprintf(" [%s]", node.Annotation.Title)
				} else {
					annotationInfo = " [annotated]"
				}
			}
			
			fmt.Fprintf(&outputBuilder, "%s%s (%s)%s\n", indent, node.Name, nodeType, annotationInfo)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk tree: %w", err)
		}
		fmt.Fprintln(&outputBuilder, "=== End Tree Structure ===")
		fmt.Fprintln(&outputBuilder)
	}

	// Phase 3 - Render tree with beautiful styling
	if options.Verbose {
		fmt.Fprintf(&outputBuilder, "treex analysis of: %s\n", targetPath)
		fmt.Fprintf(&outputBuilder, "Found %d annotations\n", len(annotations))
		fmt.Fprintln(&outputBuilder)
	}
	
	// Choose the appropriate renderer based on options and render to string
	var renderedTree string
	if options.NoColor {
		// Use plain renderer without colors
		renderedTree, err = tui.RenderPlainTreeToString(root, true)
	} else if options.Minimal {
		// Use minimal styling
		renderedTree, err = tui.RenderMinimalStyledTreeToString(root, true, options.SafeMode)
	} else {
		// Use full beautiful styling
		renderedTree, err = tui.RenderStyledTreeToStringWithSafeMode(root, true, options.SafeMode)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to render tree: %w", err)
	}
	
	// Append the rendered tree to our output
	outputBuilder.WriteString(renderedTree)
	
	return &RenderResult{
		Output: outputBuilder.String(),
		Stats:  stats,
	}, nil
} 