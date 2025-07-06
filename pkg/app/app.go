package app

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/tree"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/display/formatting"
)

// RenderOptions contains configuration options for rendering annotated trees
type RenderOptions struct {
	Verbose    bool
	IgnoreFile string
	MaxDepth   int
	SafeMode   bool
	// New field for format-based rendering
	Format string
	// View mode for controlling what paths are shown
	ViewMode string
}

// RenderResult contains the rendered output and optional verbose information
type RenderResult struct {
	Output        string
	Stats         *RenderStats
	VerboseOutput *VerboseOutput // New field for structured verbose info
	Warnings      []string       // Warnings from parsing .info files
}

// VerboseOutput holds structured information for verbose mode
type VerboseOutput struct {
	AnalyzedPath      string
	ParsedAnnotations map[string]*types.Annotation // Changed to *types.Annotation
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
	annotations, warnings, err := info.ParseDirectoryTreeWithWarnings(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .info files: %w", err)
	}

	stats.AnnotationsFound = len(annotations)
	if options.Verbose {
		verboseOutput.ParsedAnnotations = annotations
		verboseOutput.FoundAnnotations = len(annotations)
	}

	// Phase 2 - Build file tree with view mode support
	viewMode := types.ViewModeMix // default
	if options.ViewMode != "" {
		parsedMode, err := types.ParseViewMode(options.ViewMode)
		if err != nil {
			return nil, err
		}
		viewMode = parsedMode
	}

	viewOptions := types.ViewOptions{
		Mode: viewMode,
	}

	var root *types.Node
	if options.IgnoreFile != "" || options.MaxDepth != -1 {
		// Build tree with filtering options
		builder, err := tree.NewViewBuilderWithOptions(targetPath, annotations, options.IgnoreFile, options.MaxDepth, viewOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create view builder: %w", err)
		}
		root, err = builder.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build file tree with options: %w", err)
		}
	} else {
		// Build tree without filtering
		builder := tree.NewViewBuilder(targetPath, annotations, viewOptions)
		root, err = builder.Build()
		if err != nil {
			return nil, fmt.Errorf("failed to build file tree: %w", err)
		}
	}

	stats.TreeGenerated = true

	if options.Verbose {
		var treeStructureBuilder strings.Builder
		err = tree.WalkTree(root, func(node *types.Node, depth int) error {
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
				if node.Annotation.Notes != "" {
					// Get first line of notes for brief display
					lines := strings.Split(node.Annotation.Notes, "\n")
					if len(lines) > 0 && lines[0] != "" {
						annotationInfo = fmt.Sprintf(" [%s]", lines[0])
					}
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
	}

	manager := format.GetDefaultManager()
	renderResponse, err := manager.RenderTree(renderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to render tree: %w", err)
	}

	result := &RenderResult{
		Output:   renderResponse.Output,
		Stats:    stats,
		Warnings: warnings,
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

// RegisterDefaultRenderers registers all built-in renderers with the format registry
func RegisterDefaultRenderers() {
	registry := format.GetDefaultRegistry()
	
	// Register terminal renderers
	_ = registry.Register(&formatting.ColorRenderer{})
	_ = registry.Register(&formatting.MinimalRenderer{})
	_ = registry.Register(&formatting.NoColorRenderer{})
	
	// Register data format renderers
	_ = registry.Register(&format.JSONRenderer{})
	_ = registry.Register(&format.YAMLRenderer{})
	_ = registry.Register(&format.CompactJSONRenderer{})
	_ = registry.Register(&format.FlatJSONRenderer{})
	
	// Register markdown renderers
	_ = registry.Register(&formatting.MarkdownRenderer{})
	_ = registry.Register(&formatting.NestedMarkdownRenderer{})
	_ = registry.Register(&formatting.TableMarkdownRenderer{})
	
	// Register HTML renderers
	_ = registry.Register(&formatting.HTMLRenderer{})
	_ = registry.Register(&formatting.CompactHTMLRenderer{})
	_ = registry.Register(&formatting.TableHTMLRenderer{})
	
	// Register SimpleList renderer
	_ = registry.Register(&formatting.SimpleListRenderer{})
}

// init function ensures renderers are registered on package import
func init() {
	RegisterDefaultRenderers()
}
