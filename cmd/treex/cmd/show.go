package cmd

import (
	"fmt"
	"os"

	"github.com/adebert/treex/internal/info"
	"github.com/adebert/treex/internal/tree"
	"github.com/adebert/treex/internal/tui"
	"github.com/spf13/cobra"
)

// showCmd represents the main tree display functionality
// This is also the default command when no subcommand is specified
var showCmd = &cobra.Command{
	Use:   "show [path]",
	Short: "Display annotated file tree (default command)",
	Long: `Display directory trees with annotations from .info files.

This is the main functionality of treex. When no command is specified,
this command runs by default.

The command looks for .info files in the directory tree and displays
an annotated view of the file structure with descriptions.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShowCmd,
}

func init() {
	// Add flags specific to the show command
	showCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	showCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")
	showCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	showCmd.Flags().BoolVar(&minimal, "minimal", false, "Use minimal styling (fewer colors)")
	showCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	showCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	showCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")
	
	// Register the command with root
	rootCmd.AddCommand(showCmd)
}

// runShowCmd handles the CLI interface for the show command
func runShowCmd(cmd *cobra.Command, args []string) error {
	// Determine the target path
	targetPath := path
	if len(args) > 0 {
		targetPath = args[0]
	}
	if targetPath == "" {
		targetPath = "."
	}

	// Create configuration from flags
	config := &tree.DisplayConfig{
		Verbose:    verbose,
		NoColor:    noColor,
		Minimal:    minimal,
		IgnoreFile: ignoreFile,
		MaxDepth:   maxDepth,
		SafeMode:   safeMode,
	}

	// Delegate to business logic  
	if err := displayAnnotatedTree(targetPath, config, os.Stdout); err != nil {
		return fmt.Errorf("failed to display tree: %w", err)
	}

	return nil
}

// displayAnnotatedTree handles the complete business logic for displaying an annotated tree
func displayAnnotatedTree(targetPath string, config *tree.DisplayConfig, output *os.File) error {
	if config.Verbose {
		fmt.Fprintf(output, "Analyzing directory: %s\n", targetPath)
		fmt.Fprintln(output, "Verbose mode enabled - will show parsed .info structure")
		fmt.Fprintln(output)
	}

	// Phase 1 - Parse .info files (nested)
	annotations, err := info.ParseDirectoryTree(targetPath)
	if err != nil {
		return fmt.Errorf("failed to parse .info files: %w", err)
	}

	if config.Verbose {
		fmt.Fprintln(output, "=== Parsed Annotations ===")
		if len(annotations) == 0 {
			fmt.Fprintln(output, "No annotations found (no .info file or empty file)")
		} else {
			for path, annotation := range annotations {
				fmt.Fprintf(output, "Path: %s\n", path)
				if annotation.Title != "" {
					fmt.Fprintf(output, "  Title: %s\n", annotation.Title)
				}
				fmt.Fprintf(output, "  Description: %s\n", annotation.Description)
				fmt.Fprintln(output)
			}
		}
		fmt.Fprintln(output, "=== End Annotations ===")
		fmt.Fprintln(output)
	}

	// Phase 2 - Build file tree (using nested annotations with filtering options)
	var root *tree.Node
	if config.IgnoreFile != "" || config.MaxDepth != -1 {
		// Build tree with filtering options
		root, err = tree.BuildTreeNestedWithOptions(targetPath, config.IgnoreFile, config.MaxDepth)
		if err != nil {
			return fmt.Errorf("failed to build file tree with options: %w", err)
		}
	} else {
		// Build tree without filtering
		root, err = tree.BuildTreeNested(targetPath)
		if err != nil {
			return fmt.Errorf("failed to build file tree: %w", err)
		}
	}

	if config.Verbose {
		fmt.Fprintln(output, "=== File Tree Structure ===")
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
			
			fmt.Fprintf(output, "%s%s (%s)%s\n", indent, node.Name, nodeType, annotationInfo)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk tree: %w", err)
		}
		fmt.Fprintln(output, "=== End Tree Structure ===")
		fmt.Fprintln(output)
	}

	// Phase 3 - Render tree with beautiful styling
	if config.Verbose {
		fmt.Fprintf(output, "treex analysis of: %s\n", targetPath)
		fmt.Fprintf(output, "Found %d annotations\n", len(annotations))
		fmt.Fprintln(output)
	}
	
	// Choose the appropriate renderer based on flags
	if config.NoColor {
		// Use plain renderer without colors
		err = tui.RenderPlainTree(output, root, true)
	} else if config.Minimal {
		// Use minimal styling
		err = tui.RenderMinimalStyledTreeWithSafeMode(output, root, true, config.SafeMode)
	} else {
		// Use full beautiful styling
		err = tui.RenderStyledTreeWithSafeMode(output, root, true, config.SafeMode)
	}
	
	if err != nil {
		return fmt.Errorf("failed to render tree: %w", err)
	}
	
	return nil
} 