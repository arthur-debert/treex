package cmd

import (
	_ "embed" // keep: required for go:embed

	"github.com/spf13/cobra"
)

//go:embed templates/help.tpl
var helpTemplate string

//go:embed templates/long.tpl
var longTemplate string

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
	Long:  longTemplate,
	Args:  cobra.MaximumNArgs(1),
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
	rootCmd.SetHelpTemplate(helpTemplate)

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
