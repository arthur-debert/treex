package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/adebert/treex/pkg/app"
	"github.com/adebert/treex/pkg/config"
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/spf13/cobra"
)

var drawCmd = &cobra.Command{
	Use:   "draw",
	Short: "Draw tree diagrams from .info file content",
	Long: `The draw command treats .info file content as tree data rather than filesystem paths.

Examples:
  treex draw --info-file family.txt      # Draw tree from family.txt
  cat family.txt | treex draw           # Draw tree from stdin
  
The .info file format remains the same, but paths are treated as tree nodes:
  Dad Chill, dad
  Mom Listen to your mother  
  kids/Sam Little Sam`,
	Args: cobra.NoArgs,
	RunE: runDrawCmd,
}

func init() {
	// Add flags specific to draw command
	drawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file (default: stdin)")
	drawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format: color, no-color, markdown")
	drawCmd.Flags().BoolVar(&ignoreWarnings, "ignore-warnings", false, "Don't print warnings for invalid entries")
	
	// Register the command with root
	rootCmd.AddCommand(drawCmd)
}

func runDrawCmd(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Re-register renderers with loaded configuration
	app.RegisterDefaultRenderersWithConfig(cfg)

	// Validate format
	if outputFormat != "" {
		if _, err := format.ParseFormatString(outputFormat); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s", format.GetFormatHelp())
			return fmt.Errorf("invalid format: %s", outputFormat)
		}
	}

	// Read tree data from file or stdin
	var reader io.Reader
	var sourceName string
	
	if infoFile != "" {
		// Read from specified file
		file, err := os.Open(infoFile)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", infoFile, err)
		}
		defer file.Close()
		reader = file
		sourceName = infoFile
	} else {
		// Read from stdin
		reader = os.Stdin
		sourceName = "stdin"
	}

	// Parse the info file content from reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read tree data: %w", err)
	}

	// Create a temporary file to parse (since ParseInfoFile expects a file path)
	tmpFile, err := os.CreateTemp("", "treex-draw-*.info")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(content); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Parse the info file content
	annotations, warnings, err := info.ParseInfoFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to parse tree data: %w", err)
	}

	// Build tree structure from annotations
	root := buildTreeFromAnnotations(annotations)
	
	// Determine format
	formatName := outputFormat
	if formatName == "" {
		formatName = "color"
	}

	// Get the formatter
	mgr := format.GetManager()
	formatter, exists := mgr.GetFormat(formatName)
	if !exists {
		return fmt.Errorf("unknown format: %s", formatName)
	}

	// Create renderer instance
	rendererFunc := formatter.Renderer
	renderer := rendererFunc(cfg, cmd.OutOrStdout())

	// Render the tree
	output := renderer.Render(root)
	
	// Write output
	_, err = cmd.OutOrStdout().Write([]byte(output))
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Display warnings if any (unless suppressed)
	if len(warnings) > 0 && !ignoreWarnings {
		printDrawWarnings(cmd, warnings)
	}

	return nil
}

// buildTreeFromAnnotations builds a tree structure from annotations treating paths as tree nodes
func buildTreeFromAnnotations(annotations map[string]*types.Annotation) *types.Node {
	root := &types.Node{
		Name:         ".",
		RelativePath: ".",
		IsDir:        true,
		Children:     make([]*types.Node, 0),
	}

	// Build tree structure from paths
	for path, annotation := range annotations {
		addPathToTree(root, path, annotation)
	}

	return root
}

// addPathToTree adds a path to the tree structure
func addPathToTree(root *types.Node, path string, annotation *types.Annotation) {
	// Split path into components
	components := splitPath(path)
	
	current := root
	currentPath := "."
	
	// Navigate/create path in tree
	for i, component := range components {
		// Update current path
		if currentPath == "." {
			currentPath = component
		} else {
			currentPath = currentPath + "/" + component
		}
		
		// Check if this component already exists
		var found *types.Node
		for _, child := range current.Children {
			if child.Name == component {
				found = child
				break
			}
		}
		
		if found == nil {
			// Create new node
			isDir := i < len(components)-1 || isDirectory(path)
			node := &types.Node{
				Name:         component,
				RelativePath: currentPath,
				IsDir:        isDir,
				Children:     make([]*types.Node, 0),
			}
			
			// Add annotation if this is the final component
			if i == len(components)-1 {
				node.Annotation = annotation
			}
			
			current.Children = append(current.Children, node)
			current = node
		} else {
			// Use existing node
			current = found
			
			// Update annotation if this is the final component
			if i == len(components)-1 && annotation != nil {
				current.Annotation = annotation
			}
		}
	}
}

// splitPath splits a path into components, handling various formats
func splitPath(path string) []string {
	// Remove trailing slash if present
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	
	// Split by forward slash
	components := []string{}
	current := ""
	
	for _, ch := range path {
		if ch == '/' {
			if current != "" {
				components = append(components, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	
	if current != "" {
		components = append(components, current)
	}
	
	return components
}

// isDirectory checks if a path represents a directory (ends with /)
func isDirectory(path string) bool {
	return len(path) > 0 && path[len(path)-1] == '/'
}

// printDrawWarnings displays warnings for draw command
func printDrawWarnings(cmd *cobra.Command, warnings []string) {
	// Print a newline to separate from tree output
	_, _ = fmt.Fprintln(cmd.OutOrStderr())

	// Create the warning header (simple output without style for now)
	header := "⚠️  Warnings found in tree data:"
	_, _ = fmt.Fprintln(cmd.OutOrStderr(), header)

	// Print each warning with bullet point
	for _, warning := range warnings {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  • %s\n", warning)
	}
}