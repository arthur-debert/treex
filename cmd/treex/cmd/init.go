package cmd

import (
	"fmt"
	"strings"

	"github.com/adebert/treex/pkg/info"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a .info file for a directory",
	Long: `Generate a .info file for the specified directory (or current directory if not specified).

This command will:
- Scan the directory structure up to a specified depth (default: 3)
- Create a .info file with entries for all files and directories found
- Skip files that are typically not documented (like .git, node_modules, etc.)

The generated .info file will contain empty descriptions that you can fill in later.

Examples:
  treex init              # Initialize .info file for current directory
  treex init ./src        # Initialize .info file for src directory
  treex init --depth=2    # Initialize with depth limit of 2`,
	Args: cobra.MaximumNArgs(1),
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

// ShowSuccess displays the success message
func (c *CLIUserInteraction) ShowSuccess(targetPath string, depth int) {
	fmt.Printf("Initialized .info file for '%s' (depth: %d)\n", targetPath, depth)
}

// runInitCmd handles the CLI interface for init command
func runInitCmd(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}
	
	// Get depth flag
	depth, err := cmd.Flags().GetInt("depth")
	if err != nil {
		return fmt.Errorf("failed to get depth flag: %w", err)
	}
	
	// Create options
	options := info.InitOptions{
		Depth: depth,
	}
	
	// Create CLI user interaction
	userInteraction := &CLIUserInteraction{}
	
	// Delegate to business logic
	return info.InitializeInfoFile(targetPath, options, userInteraction)
} 