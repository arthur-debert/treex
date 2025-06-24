package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/maketree"
	"github.com/spf13/cobra"
)

var makeTreeCmd = &cobra.Command{
	Use:   "make-tree [input-file] [target-directory]",
	Short: "Create file/directory structure from tree text or .info file",
	Long: `Create actual file and directory structure from either a tree-like text file, .info file, or stdin.

The input can come from a file or be piped via stdin. Use "-" or omit the file argument to read from stdin.

This command can read from two types of input:

1. Tree-like text format (similar to 'treex import'):
   myproject
   ├── cmd/ Command line utilities
   ├── docs/ All documentation
   │   └── guides/ User guides and tutorials
   ├── pkg/ Core application code
   └── README.md Main project documentation

2. .info file format:
   cmd/
   Command line utilities
   
   pkg/
   Core application code
   
   README.md
   Main project documentation

The command will:
- Create the actual directory and file structure
- Create empty files (you can then populate them with content)
- Generate a master .info file in the root with all the descriptions (unless --no-info is used)

Examples:
  treex make-tree project-structure.txt ./my-project    # Read from file
  treex make-tree                                       # Read from stdin, create in current dir
  treex make-tree - ./my-project                        # Read from stdin, create in my-project
  echo "app/main.go Entry" | treex make-tree            # Pipe content
  treex make-tree .info /path/to/new/project            # Create from .info file
  treex make-tree tree.txt ./my-new-project --force     # Overwrite existing files
  treex make-tree structure.txt . --dry-run             # Preview without creating
  treex make-tree template.txt ./project --no-info      # Don't create .info file`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runMakeTreeCmd,
}

func init() {
	// Add flags specific to make-tree command
	makeTreeCmd.Flags().Bool("force", false, "Overwrite existing files and directories")
	makeTreeCmd.Flags().Bool("dry-run", false, "Show what would be created without actually creating files")
	makeTreeCmd.Flags().Bool("no-info", false, "Don't create a master .info file")

	// Register the command with root
	rootCmd.AddCommand(makeTreeCmd)
}

// runMakeTreeCmd handles the CLI interface for make-tree command
func runMakeTreeCmd(cmd *cobra.Command, args []string) error {
	var inputFile string
	var targetDir string
	var useStdin bool

	// Parse arguments based on number provided
	switch len(args) {
	case 0:
		// No arguments: read from stdin, create in current directory
		useStdin = true
		targetDir = "."
	case 1:
		// One argument: could be input file or target directory
		if args[0] == "-" {
			// Explicit stdin marker
			useStdin = true
			targetDir = "."
		} else {
			// Check if it looks like a target directory (no extension or is an existing directory)
			if filepath.Ext(args[0]) == "" {
				// No extension, treat as target directory and read from stdin
				useStdin = true
				targetDir = args[0]
			} else {
				// Has extension, treat as input file
				inputFile = args[0]
				targetDir = "."
			}
		}
	case 2:
		// Two arguments: input file and target directory
		if args[0] == "-" {
			useStdin = true
		} else {
			inputFile = args[0]
		}
		targetDir = args[1]
	}

	// Get flags
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry-run flag: %w", err)
	}

	noInfo, err := cmd.Flags().GetBool("no-info")
	if err != nil {
		return fmt.Errorf("failed to get no-info flag: %w", err)
	}

	// Create options
	options := maketree.MakeTreeOptions{
		Force:      force,
		DryRun:     dryRun,
		CreateInfo: !noInfo, // Invert the flag since default is to create .info
	}

	// Delegate to business logic
	var result *maketree.MakeResult
	if useStdin {
		result, err = maketree.MakeTreeFromReader(os.Stdin, targetDir, options)
		if err != nil {
			return fmt.Errorf("failed to make tree structure from stdin: %w", err)
		}
	} else {
		result, err = maketree.MakeTreeFromFile(inputFile, targetDir, options)
		if err != nil {
			return fmt.Errorf("failed to make tree structure: %w", err)
		}
	}

	// Display results
	return displayMakeTreeResult(cmd, result, targetDir, dryRun)
}

// displayMakeTreeResult formats and displays the result of the make-tree operation
func displayMakeTreeResult(cmd *cobra.Command, result *maketree.MakeResult, targetDir string, dryRun bool) error {
	if dryRun {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "DRY RUN - showing what would be created in %s:\n\n", targetDir); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Created file structure in %s:\n\n", targetDir); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	// Display created directories
	if len(result.CreatedDirs) > 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "📁 Directories:"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		for _, dir := range result.CreatedDirs {
			// Make path relative to target directory for cleaner output
			relativePath, err := filepath.Rel(targetDir, strings.TrimSuffix(dir, " (dry run)"))
			if err != nil {
				relativePath = dir
			} else if dryRun {
				relativePath += " (dry run)"
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", relativePath); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	// Display created files
	if len(result.CreatedFiles) > 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "📄 Files:"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		for _, file := range result.CreatedFiles {
			// Make path relative to target directory for cleaner output
			relativePath, err := filepath.Rel(targetDir, strings.TrimSuffix(file, " (dry run)"))
			if err != nil {
				relativePath = file
			} else if dryRun {
				relativePath += " (dry run)"
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", relativePath); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	// Display skipped paths
	if len(result.SkippedPaths) > 0 {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "⏭️  Skipped (already exists):"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		for _, skipped := range result.SkippedPaths {
			// Make path relative to target directory for cleaner output
			originalPath := strings.TrimSuffix(skipped, " (already exists)")
			relativePath, err := filepath.Rel(targetDir, originalPath)
			if err != nil {
				relativePath = originalPath
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", relativePath); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	// Display .info file creation
	if result.InfoFileCreated {
		if dryRun {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "📋 Would create master .info file"); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		} else {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "📋 Created master .info file"); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	// Summary
	if dryRun {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Summary: Would create %d directories, %d files",
			len(result.CreatedDirs), len(result.CreatedFiles)); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Summary: Created %d directories, %d files",
			len(result.CreatedDirs), len(result.CreatedFiles)); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	if len(result.SkippedPaths) > 0 {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), ", skipped %d existing items", len(result.SkippedPaths)); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if !dryRun && (len(result.CreatedDirs) > 0 || len(result.CreatedFiles) > 0) {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "✅ File structure created successfully!"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		if result.InfoFileCreated {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "💡 Use 'treex .' to view the annotated structure"); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
	}

	return nil
}
