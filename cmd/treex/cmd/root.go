package cmd

import (
	"fmt"
	"os"

	"github.com/adebert/treex/internal/info"
	"github.com/adebert/treex/internal/tree"
	"github.com/adebert/treex/internal/tui"
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	path       string
	noColor    bool
	minimal    bool
	ignoreFile string
	maxDepth   int
	safeMode   bool
	version    string // Holds the version string
)

// SetVersion allows the main package to set the version
func SetVersion(v string) {
	version = v
	rootCmd.Version = v // Set the version on the root command
}

var rootCmd = &cobra.Command{
	Use:   "treex [path]",
	Short: "treex is a CLI file viewer for annotated file trees",
	Long: `treex displays directory trees with annotations from .info files.
	
Annotations are read from .info files in directories and displayed
alongside the file tree structure, similar to the unix tree command
but with additional context and descriptions for files and directories.

.INFO FILES:

.info files are simple text files that describe the contents of directories.
Each directory can contain its own .info file to document files and subdirectories.

Basic format:
    filename
    Description of the file

    directory/
    Description of the directory

Example .info file:
    README.md
    Main project documentation

    src/main.go
    Application Entry Point
    Handles command line arguments and initializes the application.

    config/
    Configuration files and settings

NESTED .INFO FILES:

treex supports nested .info files - any directory can have its own .info file:
    project/.info          # Describes project/ contents
    project/src/.info      # Describes src/ contents  
    project/docs/.info     # Describes docs/ contents

Each .info file can only describe paths within its own directory for security.

GENERATING .INFO FILES:

Use 'treex gen-info <file>' to generate .info files from annotated tree structures.
The input can be simple:
    myproject/cmd The CLI utilities
    myproject/docs Documentation

Use 'treex info-files' for a quick reference guide.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Determine the target path
		targetPath := path
		if len(args) > 0 {
			targetPath = args[0]
		}
		if targetPath == "" {
			targetPath = "."
		}

		// Run the main treex logic
		if err := runTreex(targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// GetRootCommand returns the root command for use by build scripts
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func init() {
	// Add our flags
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.Flags().BoolVar(&minimal, "minimal", false, "Use minimal styling (fewer colors)")
	rootCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	rootCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")

	// Add subcommands
	// Note: completion and man page generation are handled by build scripts
}

// runTreex contains the main logic for the treex command
func runTreex(targetPath string) error {
	if verbose {
		fmt.Printf("Analyzing directory: %s\n", targetPath)
		fmt.Println("Verbose mode enabled - will show parsed .info structure")
		fmt.Println()
	}

	// Phase 1 - Parse .info files (nested)
	annotations, err := info.ParseDirectoryTree(targetPath)
	if err != nil {
		return fmt.Errorf("failed to parse .info files: %w", err)
	}

	if verbose {
		fmt.Println("=== Parsed Annotations ===")
		if len(annotations) == 0 {
			fmt.Println("No annotations found (no .info file or empty file)")
		} else {
			for path, annotation := range annotations {
				fmt.Printf("Path: %s\n", path)
				if annotation.Title != "" {
					fmt.Printf("  Title: %s\n", annotation.Title)
				}
				fmt.Printf("  Description: %s\n", annotation.Description)
				fmt.Println()
			}
		}
		fmt.Println("=== End Annotations ===")
		fmt.Println()
	}

	// Phase 2 - Build file tree (using nested annotations with filtering options)
	var root *tree.Node
	if ignoreFile != "" || maxDepth != -1 {
		// Build tree with filtering options
		root, err = tree.BuildTreeNestedWithOptions(targetPath, ignoreFile, maxDepth)
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

	if verbose {
		fmt.Println("=== File Tree Structure ===")
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
			
			fmt.Printf("%s%s (%s)%s\n", indent, node.Name, nodeType, annotationInfo)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk tree: %w", err)
		}
		fmt.Println("=== End Tree Structure ===")
		fmt.Println()
	}

	// Phase 3 - Render tree with beautiful styling
	if verbose {
		fmt.Printf("treex analysis of: %s\n", targetPath)
		fmt.Printf("Found %d annotations\n", len(annotations))
		fmt.Println()
	}
	
	// Choose the appropriate renderer based on flags
	if noColor {
		// Use plain renderer without colors
		err = tui.RenderPlainTree(os.Stdout, root, true)
	} else if minimal {
		// Use minimal styling
		err = tui.RenderMinimalStyledTreeWithSafeMode(os.Stdout, root, true, safeMode)
	} else {
		// Use full beautiful styling
		err = tui.RenderStyledTreeWithSafeMode(os.Stdout, root, true, safeMode)
	}
	
	if err != nil {
		return fmt.Errorf("failed to render tree: %w", err)
	}
	
	return nil
}

