package cmd

import (
	"fmt"

	"github.com/adebert/treex/pkg/app"
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
	options := app.RenderOptions{
		Verbose:    verbose,
		NoColor:    noColor,
		Minimal:    minimal,
		IgnoreFile: ignoreFile,
		MaxDepth:   maxDepth,
		SafeMode:   safeMode,
	}

	// Call the main business logic
	result, err := app.RenderAnnotatedTree(targetPath, options)
	if err != nil {
		return fmt.Errorf("failed to display tree: %w", err)
	}

	// Output the result
	fmt.Print(result.Output)
	return nil
}

 