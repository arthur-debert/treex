package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	path       string
	ignoreFile string
	maxDepth   int
	safeMode   bool
	// Format is defined in show.go since it's shared
)

// SetVersion allows the main package to set the version
func SetVersion(v string) {
	rootCmd.Version = v // Set the version on the root command
}

var rootCmd = &cobra.Command{
	Use:   "treex [path]",
	Short: "Vizualize project documentation through the file tree.",
	Long:  "treex displays directory trees with annotations from .info files.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to the show command for default behavior
		return runShowCmd(cmd, args)
	},
}

// Create a custom help command to control where it appears in groups
var helpCmd = &cobra.Command{
	Use:     "help [command]",
	GroupID: "help",
	Short:   "Help about any command",
	Long:    `Help provides help for any command in the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no args, show help for root command
		if len(args) == 0 {
			rootCmd.HelpFunc()(rootCmd, nil)
			return
		}

		// Find the command and show its help
		c, _, e := rootCmd.Find(args)
		if c == nil || e != nil {
			c = rootCmd
		}
		c.HelpFunc()(c, nil)
	},
}

// Execute executes the root command.
func Execute() error {
	// Disable the completion command output in help
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Use our custom help command
	rootCmd.SetHelpCommand(helpCmd)

	return rootCmd.Execute()
}

// GetRootCommand returns the root command for use by build scripts
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func init() {
	// Define command groups
	rootCmd.AddGroup(&cobra.Group{
		ID:    "main",
		Title: "Available Commands:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "info",
		Title: "Info files:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "filesystem",
		Title: "File-system:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "help",
		Title: "Help and learning:",
	})

	// Add our flags
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")

	// New format system
	rootCmd.Flags().StringVar(&outputFormat, "format", "color",
		"Output format: color, minimal, no-color, json, yaml, markdown, html, etc. (see help for all formats)")

	// Other flags
	rootCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	rootCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")

	// Add subcommands
	// Note: completion and man page generation are handled by build scripts
}
