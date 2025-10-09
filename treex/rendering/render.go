// Package rendering handles output formatting for treex CLI commands.
// It supports multiple output formats with consistent styling.
package rendering

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"treex/treex"
	"treex/treex/types"
)

// OutputFormat represents the different output formats supported
type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatPlain OutputFormat = "plain"
	FormatTerm  OutputFormat = "term"
)

// RenderConfig configures the rendering process
type RenderConfig struct {
	Format     OutputFormat // Output format to use
	Writer     io.Writer    // Where to write output
	AutoDetect bool         // Whether to auto-detect terminal capabilities
	NoColor    bool         // Force disable colors
	ShowStats  bool         // Whether to show statistics
	ShowNotes  bool         // Whether to show annotation notes
}

// Renderer handles output formatting for tree results
type Renderer struct {
	config RenderConfig
	styles *StyleManager
}

// NewRenderer creates a new renderer with the specified configuration
func NewRenderer(config RenderConfig) *Renderer {
	// Auto-detect format if not specified
	if config.Format == "" {
		config.Format = detectOutputFormat(config.Writer, config.AutoDetect)
	}

	// Default to stdout if no writer specified
	if config.Writer == nil {
		config.Writer = os.Stdout
	}

	return &Renderer{
		config: config,
		styles: NewStyleManager(config.Format == FormatTerm && !config.NoColor),
	}
}

// RenderTree renders a tree result according to the configured format
func (r *Renderer) RenderTree(result *treex.TreeResult) error {
	switch r.config.Format {
	case FormatJSON:
		return r.renderJSON(result)
	case FormatPlain, FormatTerm:
		return r.renderText(result)
	default:
		return r.renderText(result) // Default to text rendering
	}
}

// renderJSON outputs the tree result as JSON
func (r *Renderer) renderJSON(result *treex.TreeResult) error {
	// Create a JSON-friendly representation
	output := map[string]interface{}{
		"tree":  nodeToJSON(result.Root),
		"stats": result.Stats,
	}

	if len(result.PluginResults) > 0 {
		output["plugins"] = result.PluginResults
	}

	encoder := json.NewEncoder(r.config.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// renderText outputs the tree result as formatted text
func (r *Renderer) renderText(result *treex.TreeResult) error {
	if result.Root == nil {
		return nil
	}

	// Render the tree structure
	err := r.renderNode(result.Root, "", true)
	if err != nil {
		return err
	}

	// Render statistics if requested
	if r.config.ShowStats {
		err = r.renderStats(result.Stats)
		if err != nil {
			return err
		}
	}

	return nil
}

// renderNode recursively renders a node and its children
func (r *Renderer) renderNode(node *types.Node, prefix string, isLast bool) error {
	if node == nil {
		return nil
	}

	// Determine the tree connector
	var connector string
	if node.Parent == nil {
		// Root node
		connector = ""
	} else if isLast {
		connector = "└─ "
	} else {
		connector = "├─ "
	}

	// Apply styling
	styledConnector := r.styles.TreeConnector(connector)
	styledName := r.styles.FileName(node.Name)

	// Build the node line with optional annotation notes
	line := prefix + styledConnector + styledName

	// Add annotation notes if ShowNotes is enabled and node has annotation
	if r.config.ShowNotes {
		if annotation := node.GetAnnotation(); annotation != nil && annotation.Notes != "" {
			styledNotes := r.styles.Annotation("   " + annotation.Notes)
			line += styledNotes
		}
	}

	line += "\n"

	// Write the node line
	_, err := r.config.Writer.Write([]byte(line))
	if err != nil {
		return err
	}

	// Render children
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1

		// Calculate prefix for child
		var childPrefix string
		if node.Parent == nil {
			// Root node children don't get additional prefix
			childPrefix = ""
		} else if isLast {
			childPrefix = prefix + "   "
		} else {
			childPrefix = prefix + "│  "
		}

		err = r.renderNode(child, childPrefix, childIsLast)
		if err != nil {
			return err
		}
	}

	return nil
}

// renderStats renders statistics information
func (r *Renderer) renderStats(stats treex.TreeStats) error {
	statsText := r.styles.StatsHeader("\nStatistics:\n") +
		r.styles.StatsItem("  Files: ") + r.styles.StatsValue(formatNumber(stats.TotalFiles)) + "\n" +
		r.styles.StatsItem("  Directories: ") + r.styles.StatsValue(formatNumber(stats.TotalDirectories)) + "\n" +
		r.styles.StatsItem("  Max Depth: ") + r.styles.StatsValue(formatNumber(stats.MaxDepthReached)) + "\n"

	if stats.FilteredOut > 0 {
		statsText += r.styles.StatsItem("  Filtered Out: ") + r.styles.StatsValue(formatNumber(stats.FilteredOut)) + "\n"
	}

	_, err := r.config.Writer.Write([]byte(statsText))
	return err
}

// detectOutputFormat automatically determines the appropriate output format
func detectOutputFormat(writer io.Writer, autoDetect bool) OutputFormat {
	if !autoDetect {
		return FormatTerm
	}

	// Check if output is being piped or redirected
	if file, ok := writer.(*os.File); ok {
		stat, err := file.Stat()
		if err != nil {
			return FormatPlain
		}

		// If not a character device (i.e., piped or redirected), use plain format
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return FormatPlain
		}
	}

	return FormatTerm
}

// nodeToJSON converts a node tree to JSON-serializable format
func nodeToJSON(node *types.Node) interface{} {
	if node == nil {
		return nil
	}

	result := map[string]interface{}{
		"name":  node.Name,
		"path":  node.Path,
		"isDir": node.IsDir,
		"size":  node.Size,
	}

	// Include annotation notes if present
	if annotation := node.GetAnnotation(); annotation != nil && annotation.Notes != "" {
		result["notes"] = annotation.Notes
	}

	if len(node.Children) > 0 {
		children := make([]interface{}, len(node.Children))
		for i, child := range node.Children {
			children[i] = nodeToJSON(child)
		}
		result["children"] = children
	}

	return result
}

// formatNumber formats a number for display
func formatNumber(n int) string {
	return fmt.Sprintf("%d", n)
}
