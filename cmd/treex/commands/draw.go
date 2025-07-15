package commands

import (
	_ "embed"
	"fmt"
	"io"
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
	Use:     "draw [--info-file FILE | -]",
	Short:   "Draw tree diagrams from info files without filesystem validation",
	Long:    drawHelp,
	GroupID: "info",
	Args:    cobra.NoArgs,
	RunE:    runDrawCmd,
}

func init() {
	// Add flags specific to the draw command
	drawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color",
		"Output format: color, no-color, markdown")
	drawCmd.Flags().StringVar(&infoFile, "info-file", "",
		"Info file to read from (required, or use '-' for stdin)")
	drawCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	drawCmd.Flags().StringVar(&showMode, "show", "all",
		"View mode: all (draw shows all paths)")

	// Mark info-file as required
	_ = drawCmd.MarkFlagRequired("info-file")

	// Register the command with root
	rootCmd.AddCommand(drawCmd)
}

// runDrawCmd handles the CLI interface for the draw command
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

	// Parse annotations from the info file
	var annotations map[string]*types.Annotation
	if infoFile == "-" {
		// Read from stdin
		annotations, err = parseInfoFromReader(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to parse info from stdin: %w", err)
		}
	} else {
		// Read from file
		annotations, err = parseInfoFromFile(infoFile)
		if err != nil {
			return fmt.Errorf("failed to parse info from file %s: %w", infoFile, err)
		}
	}

	if len(annotations) == 0 {
		return fmt.Errorf("no annotations found in info file")
	}

	// Create a tree structure from annotations
	root, err := buildTreeFromAnnotations(annotations)
	if err != nil {
		return fmt.Errorf("failed to build tree from annotations: %w", err)
	}

	// Render the tree using the same pipeline as the show command
	renderRequest := format.RenderRequest{
		Tree:          root,
		Format:        parseFormat(outputFormat),
		Verbose:       verbose,
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

// parseInfoFromFile parses annotations from a file
func parseInfoFromFile(filename string) (map[string]*types.Annotation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return parseInfoFromReader(file)
}

// parseInfoFromReader parses annotations from an io.Reader
func parseInfoFromReader(reader io.Reader) (map[string]*types.Annotation, error) {
	parser := info.NewParser()
	
	// Create a temporary file to use the existing parser
	tempFile, err := os.CreateTemp("", "treex-draw-*.info")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// Copy reader content to temp file
	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy input: %w", err)
	}

	// Parse the temp file
	annotations, err := parser.ParseFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to parse info: %w", err)
	}

	return annotations, nil
}

// buildTreeFromAnnotations creates a tree structure from annotations without filesystem validation
func buildTreeFromAnnotations(annotations map[string]*types.Annotation) (*types.Node, error) {
	// Create root node
	root := &types.Node{
		Name:       ".",
		Path:       ".",
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
		parentPath := currentPath
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