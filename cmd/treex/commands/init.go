package commands

import (
	_ "embed"
	"fmt"
	"strings"

	editinit "github.com/adebert/treex/pkg/edit/init"
	"github.com/spf13/cobra"
)

//go:embed init.help.txt
var initHelp string

var (
	forceInit bool
)

var initCmd = &cobra.Command{
	Use:     "init [path...]",
	Short:   "Initialize a .info file for a directory or specific paths",
	GroupID: "info",
	Long:    initHelp,
	Args:    cobra.ArbitraryArgs,
	RunE:    runInitCmd,
}

func init() {
	// Add flags specific to init command
	initCmd.Flags().IntP("depth", "d", 3, "Maximum depth to scan (default: 3)")
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing .info file without confirmation")

	// Register the command with root
	rootCmd.AddCommand(initCmd)
}

// CLIUserInteraction implements the UserInteraction interface for command line usage
type CLIUserInteraction struct{
	force bool
}

// ConfirmOverwrite prompts the user for confirmation to overwrite existing .info file
func (c *CLIUserInteraction) ConfirmOverwrite(targetPath string) (bool, error) {
	if c.force {
		return true, nil
	}
	
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
	userInteraction := &CLIUserInteraction{force: forceInit}

	// Determine mode based on number of arguments
	if len(args) <= 1 {
		// Directory scanning mode
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		// Create options
		options := editinit.InitOptions{
			Depth: depth,
		}

		// Delegate to existing business logic
		return editinit.InitializeInfoFile(targetPath, options, userInteraction)
	} else {
		// Specific paths mode
		return editinit.InitializeInfoFileWithPaths(".", args, userInteraction)
	}
}
