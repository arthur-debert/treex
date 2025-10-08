// Package cmd provides the command-line interface for treex.
// It implements thin wrappers around the core API with argument parsing and output rendering.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jwaldrip/treex/treex"
	"github.com/jwaldrip/treex/treex/rendering"
	"github.com/spf13/cobra"
)

var (
	// Basic options
	maxLevel int

	// Path filtering options (added incrementally)
	excludeGlobs    []string
	includeHidden   bool
	directoriesOnly bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "treex [path]",
	Short: "A modern tree command for displaying file hierarchies",
	Long: `treex is a modernized version of the classic tree command that displays
directory structures in a tree format.

When called without arguments, treex displays the current directory tree.
You can specify a different path as an argument.`,
	Example: `  treex                    # Show current directory tree
  treex /home/user/project # Show specific directory tree
  treex -l 2               # Limit depth to 2 levels
  treex -d                 # Show directories only`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTreeCommand,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add our flags first to claim the shortcuts
	// Basic options
	rootCmd.PersistentFlags().IntVarP(&maxLevel, "level", "l", 0,
		"Maximum depth to traverse (0 = no limit)")

	// Path filtering options (added incrementally)
	rootCmd.PersistentFlags().StringSliceVarP(&excludeGlobs, "exclude", "e", []string{},
		"Exclude paths matching these glob patterns (can be used multiple times)")
	rootCmd.PersistentFlags().BoolVarP(&includeHidden, "hidden", "h", true,
		"Include hidden files and directories (default: true)")
	rootCmd.PersistentFlags().BoolVarP(&directoriesOnly, "directory", "d", false,
		"Show directories only")

	// Override default help flag to avoid conflict with our -h flag
	rootCmd.PersistentFlags().Bool("help", false, "help for treex")
	rootCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		// Use the default help template but with long-form help flag only
		command.Print(command.UsageString())
	})
}

// runTreeCommand executes the tree command with the provided arguments and flags
func runTreeCommand(cmd *cobra.Command, args []string) error {
	// Determine root path
	rootPath := "."
	if len(args) > 0 {
		rootPath = args[0]
	}

	// Convert to absolute path for consistent handling
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path %q: %w", rootPath, err)
	}

	// Verify the root path exists
	if _, err := os.Stat(absRoot); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", rootPath)
		}
		return fmt.Errorf("cannot access path %q: %w", rootPath, err)
	}

	// Build tree configuration from command-line flags
	config := buildTreeConfig(absRoot)

	// Call core API to build the tree
	result, err := treex.BuildTree(config)
	if err != nil {
		return fmt.Errorf("failed to build tree: %w", err)
	}

	// Handle empty results
	if result.Root == nil {
		fmt.Fprintf(os.Stderr, "No files found\n")
		return nil
	}

	// Configure renderer with basic terminal output (no fancy formats for now)
	renderer := rendering.NewRenderer(rendering.RenderConfig{
		Format:     rendering.FormatTerm,
		Writer:     os.Stdout,
		AutoDetect: false,
		NoColor:    false,
		ShowStats:  false,
	})

	// Render the tree
	err = renderer.RenderTree(result)
	if err != nil {
		return fmt.Errorf("failed to render tree: %w", err)
	}

	return nil
}

// buildTreeConfig creates a TreeConfig from command-line flags
func buildTreeConfig(rootPath string) treex.TreeConfig {
	config := treex.DefaultTreeConfig(rootPath)

	// Apply parsed flags
	config.MaxDepth = maxLevel
	config.ExcludeGlobs = excludeGlobs
	config.IncludeHidden = includeHidden
	config.DirectoriesOnly = directoriesOnly

	return config
}

// Version information (would be set at build time)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of treex",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("treex version %s (commit %s, built %s)\n", Version, Commit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
