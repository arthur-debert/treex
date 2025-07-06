package commands

import (
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	ignoreFile string
	maxDepth   int
	// Format is defined in show.go since it's shared
	// showMode is also defined in show.go since it's shared between root and show commands
)

// SetVersion allows the main package to set the version
func SetVersion(v string) {
	rootCmd.Version = v // Set the version on the root command
}

var rootCmd = &cobra.Command{
	Use:   "treex [path...]",
	Short: "Vizualize project documentation through the file tree.",
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
	Long: `Available output formats for the --format flag:

Terminal formats (for display):
  color           Full color terminal output with beautiful styling (default)
                  Aliases: colorful, full
  minimal         Minimal color styling for basic terminals  
                  Aliases: simple
  no-color        Plain text output without colors
                  Aliases: plain, text

Data formats (for automation and processing):
  json            JSON structured data format
  yaml            YAML structured data format
                  Aliases: yml
  compact-json    Compact JSON format (no indentation)
                  Aliases: compact
  flat-json       Flat JSON array of paths with metadata
                  Aliases: flat

Markdown formats (for documentation):
  markdown        Markdown format with clickable file links
                  Aliases: md
  nested-markdown Nested Markdown with sections and table of contents
                  Aliases: nested-md
  table-markdown  Markdown with table layout
                  Aliases: table-md

HTML formats (for web display):
  html            Interactive HTML with expandable tree
                  Aliases: interactive
  compact-html    Compact HTML format
                  Aliases: compact-web
  table-html      HTML with table layout

Special formats:
  simplelist      Simple indented list of file and directory names
                  Aliases: slist

Examples:
  treex                           # Default color output
  treex --format=json > tree.json # Export as JSON
  treex --format=minimal .        # Minimal colors for basic terminals
  treex --format=markdown > README.md  # Generate markdown documentation
  treex --format=no-color > tree.txt   # Plain text for files
  treex --format=yaml | less      # YAML output with pager`,
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

	// New format system
	rootCmd.Flags().StringVar(&outputFormat, "format", "color",
		"Output format: color, minimal, no-color, json, yaml, markdown, html, etc. (see formats command)")

	// View mode flag
	rootCmd.Flags().StringVar(&showMode, "show", "mix",
		"View mode: mix, annotated, all (default 'mix')")

	// Other flags
	rootCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")

	// Add formats command to the root
	rootCmd.AddCommand(formatsCmd)

	// Add subcommands
	// Note: completion and man page generation are handled by build scripts
}
