package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/createtree"
	"github.com/spf13/cobra"
)

var createTreeCmd = &cobra.Command{
	Use:   "create-tree <input-file> [target-directory]",
	Short: "Create file/directory structure from tree text or .info file",
	Long: `Create actual file and directory structure from either a tree-like text file or a .info file.

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
  treex create-tree project-structure.txt                    # Create in current directory
  treex create-tree .info /path/to/new/project               # Create from .info file
  treex create-tree tree.txt ./my-new-project --force        # Overwrite existing files
  treex create-tree structure.txt . --dry-run                # Preview without creating
  treex create-tree template.txt ./project --no-info         # Don't create .info file`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runCreateTreeCmd,
}

func init() {
	// Add flags specific to create-tree command
	createTreeCmd.Flags().Bool("force", false, "Overwrite existing files and directories")
	createTreeCmd.Flags().Bool("dry-run", false, "Show what would be created without actually creating files")
	createTreeCmd.Flags().Bool("no-info", false, "Don't create a master .info file")

	// Register the command with root
	rootCmd.AddCommand(createTreeCmd)
}

// runCreateTreeCmd handles the CLI interface for create-tree command
func runCreateTreeCmd(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Determine target directory
	targetDir := "."
	if len(args) > 1 {
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
	options := createtree.CreateTreeOptions{
		Force:      force,
		DryRun:     dryRun,
		CreateInfo: !noInfo, // Invert the flag since default is to create .info
	}

	// Delegate to business logic
	result, err := createtree.CreateTreeFromFile(inputFile, targetDir, options)
	if err != nil {
		return fmt.Errorf("failed to create tree structure: %w", err)
	}

	// Display results
	return displayCreateTreeResult(cmd, result, targetDir, dryRun)
}

// displayCreateTreeResult formats and displays the result of the create-tree operation
func displayCreateTreeResult(cmd *cobra.Command, result *createtree.CreateResult, targetDir string, dryRun bool) error {
	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "DRY RUN - showing what would be created in %s:\n\n", targetDir)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Created file structure in %s:\n\n", targetDir)
	}

	// Display created directories
	if len(result.CreatedDirs) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "📁 Directories:")
		for _, dir := range result.CreatedDirs {
			// Make path relative to target directory for cleaner output
			relativePath, err := filepath.Rel(targetDir, strings.TrimSuffix(dir, " (dry run)"))
			if err != nil {
				relativePath = dir
			} else if dryRun {
				relativePath += " (dry run)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", relativePath)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Display created files
	if len(result.CreatedFiles) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "📄 Files:")
		for _, file := range result.CreatedFiles {
			// Make path relative to target directory for cleaner output
			relativePath, err := filepath.Rel(targetDir, strings.TrimSuffix(file, " (dry run)"))
			if err != nil {
				relativePath = file
			} else if dryRun {
				relativePath += " (dry run)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", relativePath)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Display skipped paths
	if len(result.SkippedPaths) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "⏭️  Skipped (already exists):")
		for _, skipped := range result.SkippedPaths {
			// Make path relative to target directory for cleaner output
			originalPath := strings.TrimSuffix(skipped, " (already exists)")
			relativePath, err := filepath.Rel(targetDir, originalPath)
			if err != nil {
				relativePath = originalPath
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", relativePath)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Display .info file creation
	if result.InfoFileCreated {
		if dryRun {
			fmt.Fprintln(cmd.OutOrStdout(), "📋 Would create master .info file")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "📋 Created master .info file")
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Summary
	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Summary: Would create %d directories, %d files",
			len(result.CreatedDirs), len(result.CreatedFiles))
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Summary: Created %d directories, %d files",
			len(result.CreatedDirs), len(result.CreatedFiles))
	}

	if len(result.SkippedPaths) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), ", skipped %d existing items", len(result.SkippedPaths))
	}

	fmt.Fprintln(cmd.OutOrStdout())

	if !dryRun && (len(result.CreatedDirs) > 0 || len(result.CreatedFiles) > 0) {
		fmt.Fprintln(cmd.OutOrStdout(), "✅ File structure created successfully!")
		if result.InfoFileCreated {
			fmt.Fprintln(cmd.OutOrStdout(), "💡 Use 'treex .' to view the annotated structure")
		}
	}

	return nil
}
