package cmd

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/info"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init [path...]",
	Short:   "Initialize a .info file for a directory or specific paths",
	GroupID: "info",
	Long: `Generate a .info file for the specified directory (or current directory if not specified).

This command supports two modes:

1. Directory scanning (default - backward compatible):
   - Scan the directory structure up to a specified depth (default: 3)
   - Create a .info file with entries for all files and directories found
   - Skip files that are typically not documented (like .git, node_modules, etc.)

2. Specific paths mode (when multiple paths provided):
   - Create a .info file with entries only for the specified paths
   - Paths can be files or directories from anywhere in the project
   - Each path will be listed in the .info file for documentation

Examples:
  treex init                           # Initialize .info file for current directory
  treex init ./src                     # Initialize .info file for src directory  
  treex init --depth=2                 # Initialize with depth limit of 2
  treex init docs/dev/HELP src/main.go bin  # Initialize with specific paths only`,
	Args: cobra.ArbitraryArgs,
	RunE: runInitCmd,
}

func init() {
	// Add flags specific to init command
	initCmd.Flags().IntP("depth", "d", 3, "Maximum depth to scan (default: 3)")

	// Register the command with root
	rootCmd.AddCommand(initCmd)
}

// CLIUserInteraction implements the UserInteraction interface for command line usage
type CLIUserInteraction struct{}

// ConfirmOverwrite prompts the user for confirmation to overwrite existing .info file
func (c *CLIUserInteraction) ConfirmOverwrite(targetPath string) (bool, error) {
	fmt.Printf(".info file already exists in %s. Overwrite? [y/N]: ", targetPath)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// ShowSuccess displays the success message for directory scanning mode
func (c *CLIUserInteraction) ShowSuccess(targetPath string, depth int) {
	fmt.Printf("Initialized .info file for '%s' (depth: %d)\n", targetPath, depth)
}

// ShowSuccessWithPaths displays the success message for specific paths mode
func (c *CLIUserInteraction) ShowSuccessWithPaths(targetPath string, pathCount int) {
	fmt.Printf("Initialized .info file for '%s' (%d paths)\n", targetPath, pathCount)
}

// runInitCmd handles the CLI interface for init command
func runInitCmd(cmd *cobra.Command, args []string) error {
	// Get depth flag
	depth, err := cmd.Flags().GetInt("depth")
	if err != nil {
		return fmt.Errorf("failed to get depth flag: %w", err)
	}

	// Create CLI user interaction
	userInteraction := &CLIUserInteraction{}

	// Determine mode based on number of arguments
	if len(args) <= 1 {
		// Directory scanning mode (backward compatible)
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		// Create options
		options := info.InitOptions{
			Depth: depth,
		}

		// Delegate to existing business logic
		return info.InitializeInfoFile(targetPath, options, userInteraction)
	} else {
		// Specific paths mode
		return info.InitializeInfoFileWithPaths(".", args, userInteraction)
	}
}
