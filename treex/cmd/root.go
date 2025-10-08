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
	// Multiple exclusion mechanisms work together:
	// 1. Built-in ignores (default VCS/build patterns, disable with --no-builtin-ignores)
	// 2. User excludes (--exclude patterns)
	// 3. Gitignore files (automatic .gitignore support)
	// 4. Hidden files (--hidden flag control)
	noBuiltinIgnores bool     // Disable built-in ignore patterns
	excludeGlobs     []string // User-specified exclude patterns
	includeHidden    bool     // Include hidden files
	directoriesOnly  bool     // Show directories only
)

// rootCmd represents the base command when called without any subcommands
// According to cli-architecture.txt, "treex" should default to tree rendering
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

// treeCmd represents the explicit tree command
// This provides "treex tree" as an explicit alternative to naked "treex"
var treeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Display directory tree structure",
	Long: `Display directory tree structure in a hierarchical format.

This is the explicit form of the default treex command.`,
	Example: `  treex tree                    # Show current directory tree
  treex tree /path          # Show specific directory tree
  treex tree -l 2           # Limit depth to 2 levels`,
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
	// Add the explicit tree command as a subcommand
	rootCmd.AddCommand(treeCmd)

	// Configure flags for both root and tree commands
	setupTreeFlags(rootCmd)
	setupTreeFlags(treeCmd)
}

// setupTreeFlags configures the tree-related flags for a command
func setupTreeFlags(cmd *cobra.Command) {
	// Basic options
	cmd.PersistentFlags().IntVarP(&maxLevel, "level", "l", 0,
		"Maximum depth to traverse (0 = no limit)")

	// Path filtering options (added incrementally)
	// Multiple exclusion mechanisms work together for comprehensive filtering
	cmd.PersistentFlags().BoolVar(&noBuiltinIgnores, "no-builtin-ignores", false,
		"Disable built-in ignore patterns (.git, node_modules, __pycache__, etc.)")
	cmd.PersistentFlags().StringSliceVarP(&excludeGlobs, "exclude", "e", []string{},
		"Exclude paths matching these glob patterns (can be used multiple times)")
	cmd.PersistentFlags().BoolVarP(&includeHidden, "hidden", "h", true,
		"Include hidden files and directories (default: true)")
	cmd.PersistentFlags().BoolVarP(&directoriesOnly, "directory", "d", false,
		"Show directories only")

	// Override default help flag to avoid conflict with our -h flag
	cmd.PersistentFlags().Bool("help", false, "help for treex")
	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		// Use the default help template but with long-form help flag only
		command.Print(command.UsageString())
	})
}

// runTreeCommand executes the tree command with the provided arguments and flags
// This is the core CLI logic that both "treex" and "treex tree" use
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
// Maps CLI flags to TreeConfig, coordinating multiple exclusion mechanisms
func buildTreeConfig(rootPath string) treex.TreeConfig {
	config := treex.DefaultTreeConfig(rootPath)

	// Apply parsed flags - coordinate all exclusion mechanisms:
	// 1. Built-in ignores (disabled by --no-builtin-ignores)
	config.BuiltinIgnores = !noBuiltinIgnores
	// 2. User exclude patterns (--exclude flags)
	config.ExcludeGlobs = excludeGlobs
	// 3. Hidden file control (--hidden flag)
	config.IncludeHidden = includeHidden
	// 4. Directory filtering (--directory flag)
	config.DirectoriesOnly = directoriesOnly

	config.MaxDepth = maxLevel

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
