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
	Long: `treex displays directory trees with annotations from .info files.
	
Annotations are read from .info files in directories and displayed
alongside the file tree structure, similar to the unix tree command
but with additional context and descriptions for files and directories.

.INFO FILES:

.info files are simple text files that describe the contents of directories.
Each directory can contain its own .info file to document files and subdirectories.

Basic format:
    filename
    Description of the file

    directory/
    Description of the directory

Example .info file:
    README.md
    Main project documentation

    src/main.go
    Application Entry Point
    Handles command line arguments and initializes the application.

    config/
    Configuration files and settings

OUTPUT FORMATS:

treex supports multiple output formats for different use cases:

Terminal formats:
  --format=color    Full color terminal output with beautiful styling (default)
  --format=minimal  Minimal color styling for basic terminals
  --format=no-color Plain text output without colors

Examples:
  treex                           # Full color output (default)
  treex --format=minimal .        # Minimal colors for basic terminals
  treex --format=no-color > tree.txt  # Plain text suitable for files
  treex --format=plain .          # Alternative alias for no-color


NESTED .INFO FILES:

treex supports nested .info files - any directory can have its own .info file:
    project/.info          # Describes project/ contents
    project/src/.info      # Describes src/ contents  
    project/docs/.info     # Describes docs/ contents

Each .info file can only describe paths within its own directory for security.

GENERATING .INFO FILES:

Use 'treex import <file>' to generate .info files from annotated tree structures.
The input can be simple:
    myproject/cmd The CLI utilities
    myproject/docs Documentation

Use 'treex info-files' for a quick reference guide.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to the show command for default behavior
		return runShowCmd(cmd, args)
	},
}

// Execute executes the root command.
func Execute() error {
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
		Title: "New .info files:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "filesystem",
		Title: "File-system:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "help",
		Title: "Help and learning:",
	})

	// Set custom help template to show only short description
	rootCmd.SetHelpTemplate(`{{.Short}}

{{.UsageString}}`)

	// Set custom usage template to match desired format
	rootCmd.SetUsageTemplate(`Usage: 
  $ {{.CommandPath}}{{if .HasAvailableSubCommands}}
  $ {{.CommandPath}} add <path> <description>{{end}}
  {{range $group := .Groups}}
{{.Title}}{{range $cmd := $.Commands}}{{if (and (eq $cmd.GroupID $group.ID) (or $cmd.IsAvailableCommand (eq $cmd.Name "help")))}}
    {{rpad $cmd.Name $cmd.NamePadding }} {{$cmd.Short}}{{end}}{{end}}{{if eq $group.ID "help"}}
    {{rpad "help" $.NamePadding }} Help about any command{{end}}
{{end}}
`)

	// Add our flags
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")

	// New format system
	rootCmd.Flags().StringVar(&outputFormat, "format", "color",
		"Output format: color, minimal, no-color")


	// Other flags
	rootCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	rootCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")

	// Add subcommands
	// Note: completion and man page generation are handled by build scripts
}
