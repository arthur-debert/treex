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

FORMATS:

treex supports multiple output formats for different use cases. Use --format=<name> to specify:

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
  treex --format=yaml | less      # YAML output with pager

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
		"Output format: color, minimal, no-color, json, yaml, markdown, html, etc. (see help for all formats)")

	// Other flags
	rootCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	rootCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")

	// Add subcommands
	// Note: completion and man page generation are handled by build scripts
}
