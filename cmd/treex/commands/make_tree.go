package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/edit/maketree"
	"github.com/spf13/cobra"
)

var makeTreeCmd = &cobra.Command{
	Use:     "make-tree [input-file] [target-directory]",
	Short:   "Create file/directory structure from tree text or .info file",
	GroupID: "filesystem",
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
- Generate a master .info file in the root with all the descriptions

Examples:
  treex make-tree project-structure.txt ./my-project    # Read from file
  treex make-tree                                       # Read from stdin, create in current dir
  treex make-tree - ./my-project                        # Read from stdin, create in my-project
  echo "app/main.go Entry" | treex make-tree            # Pipe content
  treex make-tree .info /path/to/new/project            # Create from .info file`,
	Args: cobra.RangeArgs(0, 2),
	RunE: runMakeTreeCmd,
}

func init() {
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

	// Create options - always create .info files, skip existing files
	options := maketree.MakeTreeOptions{
		Force:      false, // Never overwrite existing files
		DryRun:     false, // No dry run mode
		CreateInfo: true,  // Always create .info files
		InfoHeader: "Now you can document your project, go crazy!",
	}

	// Delegate to business logic
	if useStdin {
		result, err := maketree.MakeTreeFromReader(os.Stdin, targetDir, options)
		if err != nil {
			return fmt.Errorf("failed to make tree structure from stdin: %w", err)
		}
		return displayMakeTreeResult(cmd, result, targetDir)
	} else {
		result, err := maketree.MakeTreeFromFile(inputFile, targetDir, options)
		if err != nil {
			return fmt.Errorf("failed to make tree structure: %w", err)
		}
		return displayMakeTreeResult(cmd, result, targetDir)
	}
}

// displayMakeTreeResult formats and displays the result of the make-tree operation
func displayMakeTreeResult(cmd *cobra.Command, result *maketree.MakeResult, targetDir string) error {
	// Collect all created paths (both directories and files)
	var allCreatedPaths []string
	for _, dir := range result.CreatedDirs {
		relativePath, err := filepath.Rel(targetDir, dir)
		if err != nil {
			relativePath = dir
		}
		allCreatedPaths = append(allCreatedPaths, relativePath)
	}
	for _, file := range result.CreatedFiles {
		relativePath, err := filepath.Rel(targetDir, file)
		if err != nil {
			relativePath = file
		}
		allCreatedPaths = append(allCreatedPaths, relativePath)
	}

	// Display the created paths on a single line in bold
	if len(allCreatedPaths) > 0 {
		// Bold text using ANSI escape codes
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "The following files/dirs were created: \033[1m%s\033[0m\n\n",
			strings.Join(allCreatedPaths, " ")); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	// Run the show command to display the tree
	if len(result.CreatedDirs) > 0 || len(result.CreatedFiles) > 0 {
		showCmd.SetOut(cmd.OutOrStdout())
		showCmd.SetErr(cmd.ErrOrStderr())
		if err := runShowCmd(showCmd, []string{targetDir}); err != nil {
			return fmt.Errorf("failed to show tree structure: %w", err)
		}
	}

	return nil
}
