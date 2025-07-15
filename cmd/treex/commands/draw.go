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
	Short: "Draw tree diagrams from info file format without filesystem",
	Long: `Draw renders tree diagrams from .info format files without requiring
the paths to exist in the filesystem. This allows creating documentation
diagrams for conceptual structures.

Examples:
  treex draw --info-file family.txt    # Draw from specific file
  treex draw < organization.txt        # Draw from stdin
  cat diagram.info | treex draw        # Draw from pipe`,
	Args: cobra.NoArgs,
	RunE: runDrawCmd,
}

func init() {
	// Add flags specific to draw command
	drawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file (required unless piping)")
	drawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", 
		"Output format: color, no-color, markdown")
	
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

	// Determine input source
	var annotations map[string]*types.Annotation
	var parseWarnings []string

	// Check if we're reading from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		annotations, parseWarnings, err = parseInfoFromReader(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to parse input: %w", err)
		}
	} else if infoFile != "" {
		// Read from specified file
		annotations, parseWarnings, err = info.ParseInfoFile(infoFile)
		if err != nil {
			return fmt.Errorf("failed to parse info file: %w", err)
		}
	} else {
		return fmt.Errorf("no input provided: use --info-file or pipe data to stdin")
	}

	// For draw command, we always ignore filesystem warnings
	// since paths are conceptual, not real filesystem paths
	_ = parseWarnings

	// Build a virtual tree from the annotations
	root := buildVirtualTree(annotations)

	// Validate format
	if outputFormat != "" {
		if _, err := format.ParseFormatString(outputFormat); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s", format.GetFormatHelp())
			return fmt.Errorf("invalid format: %s", outputFormat)
		}
	}

	// Render the tree
	renderRequest := format.RenderRequest{
		Tree:          root,
		Format:        parseFormat(outputFormat),
		Verbose:       false,
		ShowStats:     false,
		SafeMode:      false,
		TerminalWidth: 80,
	}

	manager := format.GetDefaultManager()
	renderResponse, err := manager.RenderTree(renderRequest)
	if err != nil {
		return fmt.Errorf("failed to render tree: %w", err)
	}

	// Output the result
	_, err = cmd.OutOrStdout().Write([]byte(renderResponse.Output))
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// parseInfoFromReader parses info format from an io.Reader
func parseInfoFromReader(r io.Reader) (map[string]*types.Annotation, []string, error) {
	// Read all content
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Create a temporary file to use the existing parser
	tmpFile, err := os.CreateTemp("", "treex-draw-*.info")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write content to temp file
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return nil, nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Parse using existing parser
	return info.ParseInfoFile(tmpFile.Name())
}

// buildVirtualTree builds a tree structure from annotations without filesystem
func buildVirtualTree(annotations map[string]*types.Annotation) *types.Node {
	root := &types.Node{
		Name:     ".",
		Path:     ".",
		IsDir:    true,
		Children: []*types.Node{},
	}

	// Create a map to store nodes by path
	nodeMap := make(map[string]*types.Node)
	nodeMap["."] = root

	// Sort paths to ensure parent directories are created before children
	paths := make([]string, 0, len(annotations))
	for path := range annotations {
		paths = append(paths, path)
	}
	
	// Simple bubble sort for stability
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[j] < paths[i] {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}

	// Process each annotation
	for _, path := range paths {
		annotation := annotations[path]
		
		// Skip empty paths
		if path == "" || path == "." {
			continue
		}

		// Determine if it's a directory (ends with /)
		isDir := false
		cleanPath := path
		if len(path) > 0 && path[len(path)-1] == '/' {
			isDir = true
			cleanPath = path[:len(path)-1]
		}

		// Create the node
		node := &types.Node{
			Name:       getBaseName(cleanPath),
			Path:       cleanPath,
			IsDir:      isDir,
			Annotation: annotation,
			Children:   []*types.Node{},
		}

		// Find or create parent
		parentPath := getParentPath(cleanPath)
		parent, exists := nodeMap[parentPath]
		if !exists {
			// Create parent directory nodes as needed
			parent = ensureParentExists(nodeMap, parentPath, root)
		}

		// Add to parent
		parent.Children = append(parent.Children, node)
		nodeMap[cleanPath] = node
	}

	return root
}

// ensureParentExists creates parent directory nodes as needed
func ensureParentExists(nodeMap map[string]*types.Node, path string, root *types.Node) *types.Node {
	if path == "." || path == "" {
		return root
	}

	// Check if already exists
	if node, exists := nodeMap[path]; exists {
		return node
	}

	// Create parent first
	parentPath := getParentPath(path)
	parent := ensureParentExists(nodeMap, parentPath, root)

	// Create this directory node
	node := &types.Node{
		Name:     getBaseName(path),
		Path:     path,
		IsDir:    true,
		Children: []*types.Node{},
	}

	parent.Children = append(parent.Children, node)
	nodeMap[path] = node

	return node
}

// getParentPath returns the parent directory path
func getParentPath(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "."
			}
			return path[:i]
		}
	}
	return "."
}

// getBaseName returns the base name of a path
func getBaseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
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