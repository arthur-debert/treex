package commands

import (
	"fmt"
	"os"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:     "check [path]",
	Short:   "Validate .info files in a directory",
	GroupID: "info",
	Long: `Validate .info files in the specified directory (or current directory if not specified).

This command will:
- Parse all .info files in the directory tree
- Check for syntax errors and formatting issues
- Verify that referenced paths exist
- Exit with code 0 if all .info files are valid (prints nothing)
- Exit with code 1 if any .info files have errors (prints error details)

Examples:
  treex check              # Check .info files in current directory
  treex check ./src        # Check .info files in src directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheckCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(checkCmd)
}

// runCheckCmd handles the CLI interface for check command
func runCheckCmd(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Delegate to business logic
	err := info.ValidateInfoFiles(targetPath)
	if err != nil {
		// Print the error and exit with code 1
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}

	// Success - print nothing and exit with code 0
	return nil
}
