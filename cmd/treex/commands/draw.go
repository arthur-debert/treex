package commands

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/app"
	"github.com/adebert/treex/pkg/config"
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/spf13/cobra"
)

//go:embed draw.help.txt
var drawHelp string

// drawCmd represents the draw command
var drawCmd = &cobra.Command{
	Use:   "draw [--info-file FILE | -]",
	Short: "Draw tree diagrams from info files without filesystem validation",
	Long:  drawHelp,
	Args:  cobra.NoArgs,
	RunE:  runDrawCmd,
}

func init() {
	// Add flags specific to the draw command
	drawCmd.Flags().StringP("format", "f", "color",
		"Output format: color, no-color, markdown")
	drawCmd.Flags().String("info-file", "",
		"Info file to read from (optional, reads from stdin if not provided)")

	// Register the command with root
	infoCmd.AddCommand(drawCmd)
}

// runDrawCmd handles the CLI interface for the draw command
func runDrawCmd(cmd *cobra.Command, args []string) error {
	// Get flag values from the command
	infoFile, err := cmd.Flags().GetString("info-file")
	if err != nil {
		return fmt.Errorf("failed to get info-file flag: %w", err)
	}
	
	outputFormat, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("failed to get format flag: %w", err)
	}

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

	// Parse annotations from the info file or stdin
	var annotations map[string]*types.Annotation
	
	// Check if we're reading from stdin
	stat, _ := os.Stdin.Stat()
	isStdinPipe := (stat.Mode() & os.ModeCharDevice) == 0

	if infoFile == "-" || (infoFile == "" && isStdinPipe) {
		// Read from stdin
		annotations, _, err = info.ParseInfoFromReader(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to parse info from stdin: %w", err)
		}
	} else if infoFile != "" {
		// Read from file
		annotations, _, err = info.ParseInfoFile(infoFile)
		if err != nil {
			return fmt.Errorf("failed to parse info file %s: %w", infoFile, err)
		}
	} else {
		return fmt.Errorf("no input provided: use --info-file or pipe data to stdin")
	}

	if len(annotations) == 0 {
		return fmt.Errorf("no annotations found in info file")
	}

	// For draw command, we always ignore filesystem warnings
	// since paths are conceptual, not real filesystem paths

	// Create a tree structure from annotations
	root, err := BuildVirtualTree(annotations)
	if err != nil {
		return fmt.Errorf("failed to build tree from annotations: %w", err)
	}

	// Sort the tree nodes to ensure deterministic output
	sortNodeChildren(root)

	// Render the tree using the same pipeline as the show command
	renderRequest := format.RenderRequest{
		Tree:          root,
		Format:        parseFormat(outputFormat),
		Verbose:       false, // Draw command doesn't need verbose output
		ShowStats:     false,
		SafeMode:      false,
		TerminalWidth: 80,
	}

	manager := format.GetDefaultManager()
	renderResponse, err := manager.RenderTree(renderRequest)
	if err != nil {
		return fmt.Errorf("failed to render tree: %w", err)
	}

	// Write the result
	_, err = cmd.OutOrStdout().Write([]byte(renderResponse.Output))
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}



// BuildVirtualTree creates a tree structure from annotations without filesystem validation
func BuildVirtualTree(annotations map[string]*types.Annotation) (*types.Node, error) {
	if len(annotations) == 0 {
		return nil, fmt.Errorf("no annotations provided")
	}
	// Create root node
	root := &types.Node{
		Name:       "root",
		Path:       "",
		IsDir:      true,
		Children:   make([]*types.Node, 0),
		Annotation: nil,
	}

	// Build node map for efficient lookup
	nodeMap := make(map[string]*types.Node)
	nodeMap["."] = root

	// Process each annotation to build the tree
	for path, annotation := range annotations {
		err := addPathToTree(root, nodeMap, path, annotation)
		if err != nil {
			return nil, fmt.Errorf("failed to add path %s to tree: %w", path, err)
		}
	}

	// If root has only one child and no annotation, return that child as the root
	if len(root.Children) == 1 && root.Annotation == nil {
		return root.Children[0], nil
	}

	return root, nil
}

// addPathToTree adds a path and its annotation to the tree
func addPathToTree(root *types.Node, nodeMap map[string]*types.Node, path string, annotation *types.Annotation) error {
	// Clean and normalize the path
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		// Root annotation
		root.Annotation = annotation
		return nil
	}

	// Split path into components
	parts := strings.Split(cleanPath, string(filepath.Separator))
	if parts[0] == "." {
		parts = parts[1:]
	}

	// Build the path incrementally
	currentPath := "."
	currentNode := root

	for i, part := range parts {
		if currentPath == "." {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		// Check if node already exists
		if existingNode, exists := nodeMap[currentPath]; exists {
			currentNode = existingNode
		} else {
			// Create new node
			isDir := i < len(parts)-1 || strings.HasSuffix(path, "/")
			newNode := &types.Node{
				Name:     part,
				Path:     currentPath,
				IsDir:    isDir,
				Children: make([]*types.Node, 0),
			}

			// Add to parent's children
			currentNode.Children = append(currentNode.Children, newNode)
			nodeMap[currentPath] = newNode
			currentNode = newNode
		}
	}

	// Set the annotation on the final node
	currentNode.Annotation = annotation

	return nil
}

// parseFormat safely converts a format string to OutputFormat (copied from app.go)
func parseFormat(formatStr string) format.OutputFormat {
	if formatStr == "" {
		return ""
	}

	if parsedFormat, err := format.ParseFormatString(formatStr); err == nil {
		return parsedFormat
	}

	return format.OutputFormat(formatStr)
}

// sortNodeChildren sorts the children of a node alphabetically
func sortNodeChildren(node *types.Node) {
	if node == nil || !node.IsDir {
		return
	}

	// Sort children
	node.SortChildren()

	// Recursively sort children of subdirectories
	for _, child := range node.Children {
		if child.IsDir {
			sortNodeChildren(child)
		}
	}
}
