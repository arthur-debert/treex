package commands

import (
	_ "embed"
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	ignoreFile string
	noIgnore   bool
	infoFile   string
	maxDepth   int
	ignoreWarnings bool
	// Format is defined in show.go since it's shared
	// showMode is also defined in show.go since it's shared between root and show commands
)

//go:embed formats.help.txt
var formatsHelp string

// SetVersion allows the main package to set the version
func SetVersion(v string) {
	rootCmd.Version = v // Set the version on the root command
}

var rootCmd = &cobra.Command{
	Use:   "treex [path...]",
	Short: "treex : renders a documented file tree in your shell.",
	Long: `treex displays directory trees with annotations from .info files.

Multiple paths can be specified to show multiple directories:
  treex docs src                  # Show docs and src directories
  treex dir1 dir2 dir3           # Show multiple directories`,
	Args: cobra.ArbitraryArgs,
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

// formatsCmd is a special command to list all available output formats
var formatsCmd = &cobra.Command{
	Use:     "formats",
	GroupID: "help",
	Short:   "List available output formats (--format=NAME)",
	Long:    formatsHelp,
	// This command doesn't actually do anything when run
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute executes the root command.
func Execute() error {
	// Disable the completion command output in help
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Use our custom help command
	rootCmd.SetHelpCommand(helpCmd)

	// Set up a PersistentPreRun to handle theme detection
	// Remove the persistent pre-run since lipgloss handles theme detection automatically

	return rootCmd.Execute()
}

// GetRootCommand returns the root command for use by build scripts
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func init() {
	// Define command groups
	rootCmd.AddGroup(&cobra.Group{
		ID:    "info",
		Title: "Authoring Annotations:",
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
	_ = rootCmd.Flags().MarkHidden("verbose")

	// New format system
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "color",
		"color, no-color, markdown (see formats command)")

	// View mode flag
	rootCmd.Flags().StringVar(&showMode, "show", "mix",
		"View mode: mix, annotated, all")

	// Other flags
	rootCmd.Flags().StringVar(&ignoreFile, "ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().BoolVar(&noIgnore, "no-ignore", false, "Don't use any ignore file")
	rootCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	rootCmd.Flags().BoolVar(&ignoreWarnings, "ignore-warnings", false, "Don't print warnings for non-existent paths in .info files")

	// Add formats command to the root
	rootCmd.AddCommand(formatsCmd)

	// Add subcommands
	// Note: completion and man page generation are handled by build scripts
}
