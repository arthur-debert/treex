package commands

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

// drawCmd represents the draw command for creating tree diagrams from info files
var drawCmd = &cobra.Command{
	Use:     "draw [--info-file FILE]",
	GroupID: "info",
	Short:   "Draw tree diagrams from info files",
	Long:    drawHelp,
	RunE:    runDrawCmd,
}

func init() {
	// Add flags specific to the draw command
	drawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color",
		"Output format: color, no-color, markdown")
	drawCmd.Flags().StringVar(&infoFile, "info-file", "",
		"Info file to draw from (optional if piping from stdin)")

	// Register the command with root
	rootCmd.AddCommand(drawCmd)
}

// runDrawCmd handles the CLI interface for the draw command
func runDrawCmd(cmd *cobra.Command, args []string) error {
	// Get flag values from the command
	outputFormat, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("failed to get format flag: %w", err)
	}
	
	infoFile, err := cmd.Flags().GetString("info-file")
	if err != nil {
		return fmt.Errorf("failed to get info-file flag: %w", err)
	}

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
		annotations, parseWarnings, err = info.ParseInfoFromReader(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to parse input: %w", err)
		}
	} else if infoFile != "" {
		// Read from specified file
		parser := info.NewParser()
		annotations, parseWarnings, err = parser.ParseFileWithWarnings(infoFile)
		if err != nil {
			return fmt.Errorf("failed to parse info file: %w", err)
		}
	} else {
		return fmt.Errorf("no input provided: use --info-file or pipe data to stdin")
	}

	// Check if we have any annotations
	if len(annotations) == 0 {
		if infoFile != "" {
			return fmt.Errorf("no annotations found in %s", infoFile)
		} else {
			return fmt.Errorf("no annotations found in input")
		}
	}

	// For draw command, we always ignore filesystem warnings
	// since paths are conceptual, not real filesystem paths
	_ = parseWarnings

	// Build a virtual tree from the annotations
	root, err := BuildVirtualTree(annotations)
	if err != nil {
		return fmt.Errorf("failed to build virtual tree: %w", err)
	}

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

// BuildVirtualTree creates a virtual tree structure from annotations
func BuildVirtualTree(annotations map[string]*types.Annotation) (*types.Node, error) {
	if len(annotations) == 0 {
		return nil, fmt.Errorf("no annotations provided")
	}

	// Create root node
	root := &types.Node{
		Name:         "root",
		Path:         "",
		RelativePath: "",
		IsDir:        true,
		Annotation:   nil,
		Children:     make([]*types.Node, 0),
		Parent:       nil,
	}

	// Sort paths to ensure consistent ordering
	var paths []string
	for path := range annotations {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// Build the tree structure
	nodeMap := make(map[string]*types.Node)
	nodeMap[""] = root

	for _, path := range paths {
		annotation := annotations[path]
		
		// Clean the path
		cleanPath := strings.TrimSpace(path)
		if cleanPath == "" {
			continue
		}

		// Determine if this is a directory (ends with /)
		isDir := strings.HasSuffix(cleanPath, "/")
		if isDir {
			cleanPath = strings.TrimSuffix(cleanPath, "/")
		}

		// Create all parent directories if they don't exist
		if err := EnsureParentDirectories(cleanPath, nodeMap, root); err != nil {
			return nil, err
		}

		// Create the node
		node := &types.Node{
			Name:         filepath.Base(cleanPath),
			Path:         cleanPath,
			RelativePath: cleanPath,
			IsDir:        isDir,
			Annotation:   annotation,
			Children:     make([]*types.Node, 0),
		}

		// Find parent
		parentPath := filepath.Dir(cleanPath)
		if parentPath == "." {
			parentPath = ""
		}

		parent, exists := nodeMap[parentPath]
		if !exists {
			return nil, fmt.Errorf("parent directory not found for path: %s", cleanPath)
		}

		// Set parent and add to parent's children
		node.Parent = parent
		parent.Children = append(parent.Children, node)

		// Add to node map
		nodeMap[cleanPath] = node
	}

	return root, nil
}

// EnsureParentDirectories creates all parent directories for a given path
func EnsureParentDirectories(path string, nodeMap map[string]*types.Node, root *types.Node) error {
	if path == "" {
		return nil
	}

	// Get parent path
	parentPath := filepath.Dir(path)
	if parentPath == "." {
		parentPath = ""
	}

	// If parent already exists, we're done
	if _, exists := nodeMap[parentPath]; exists {
		return nil
	}

	// Recursively ensure parent's parent exists
	if err := EnsureParentDirectories(parentPath, nodeMap, root); err != nil {
		return err
	}

	// Create parent directory node
	parentNode := &types.Node{
		Name:         filepath.Base(parentPath),
		Path:         parentPath,
		RelativePath: parentPath,
		IsDir:        true,
		Annotation:   nil,
		Children:     make([]*types.Node, 0),
	}

	// Find grandparent
	grandParentPath := filepath.Dir(parentPath)
	if grandParentPath == "." {
		grandParentPath = ""
	}

	grandParent, exists := nodeMap[grandParentPath]
	if !exists {
		return fmt.Errorf("grandparent directory not found for path: %s", parentPath)
	}

	// Set parent and add to grandparent's children
	parentNode.Parent = grandParent
	grandParent.Children = append(grandParent.Children, parentNode)

	// Add to node map
	nodeMap[parentPath] = parentNode

	return nil
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