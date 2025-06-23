package app

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/format"
	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

// RenderOptions contains configuration options for rendering annotated trees
// DEPRECATED: Use format.RenderRequest instead - this is kept for backward compatibility
type RenderOptions struct {
	Verbose    bool
	NoColor    bool
	Minimal    bool
	IgnoreFile string
	MaxDepth   int
	SafeMode   bool
	// New field for format-based rendering
	Format string
}

// RenderResult contains the rendered output and optional verbose information
type RenderResult struct {
	Output        string
	Stats         *RenderStats
	VerboseOutput *VerboseOutput // New field for structured verbose info
}

// VerboseOutput holds structured information for verbose mode
type VerboseOutput struct {
	AnalyzedPath      string
	ParsedAnnotations map[string]*info.Annotation // Changed to *info.Annotation
	TreeStructure     string                      // Keep tree structure as string for now, could be structured further if needed
	FoundAnnotations  int
}

// RenderStats contains statistics about the rendering process
type RenderStats struct {
	AnnotationsFound int
	TreeGenerated    bool
}

// RenderAnnotatedTree is the main business logic function that generates an annotated tree
// This function handles all the core application logic and returns a complete rendered string
func RenderAnnotatedTree(targetPath string, options RenderOptions) (*RenderResult, error) {
	stats := &RenderStats{}
	verboseOutput := &VerboseOutput{}

	if options.Verbose {
		verboseOutput.AnalyzedPath = targetPath
	}

	// Phase 1 - Parse .info files (nested)
	annotations, err := info.ParseDirectoryTree(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .info files: %w", err)
	}

	stats.AnnotationsFound = len(annotations)
	if options.Verbose {
		verboseOutput.ParsedAnnotations = annotations
		verboseOutput.FoundAnnotations = len(annotations)
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
		var treeStructureBuilder strings.Builder
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

			fmt.Fprintf(&treeStructureBuilder, "%s%s (%s)%s\n", indent, node.Name, nodeType, annotationInfo)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk verbose tree: %w", err)
		}
		verboseOutput.TreeStructure = treeStructureBuilder.String()
	}

	// Phase 3 - Render tree using the new RendererManager
	renderRequest := format.RenderRequest{
		Tree:          root,
		Format:        parseFormat(options.Format),
		Verbose:       false, // Tree rendering verbosity is separate from app verbosity
		ShowStats:     false,
		SafeMode:      options.SafeMode,
		TerminalWidth: 80, // TODO: Consider making this dynamic or configurable
		LegacyNoColor: options.NoColor,
		LegacyMinimal: options.Minimal,
	}

	manager := format.GetDefaultManager()
	renderResponse, err := manager.RenderTree(renderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to render tree: %w", err)
	}

	result := &RenderResult{
		Output: renderResponse.Output,
		Stats:  stats,
	}

	if options.Verbose {
		result.VerboseOutput = verboseOutput
	}

	return result, nil
}

// parseFormat safely converts a format string to OutputFormat
func parseFormat(formatStr string) format.OutputFormat {
	if formatStr == "" {
		return "" // Let the manager use defaults
	}

	// Try to parse, but don't fail - let the manager handle validation
	if parsedFormat, err := format.ParseFormatString(formatStr); err == nil {
		return parsedFormat
	}

	// Return as-is and let the manager handle the error
	return format.OutputFormat(formatStr)
}
