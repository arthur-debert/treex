package commands

import (
	"fmt"
	"path/filepath"

	"github.com/adebert/treex/pkg/app"
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/display/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Format-based flags (new system)
	outputFormat string
	// View mode flag
	showMode string
)

// showCmd represents the main tree display functionality
// This is also the default command when no subcommand is specified
var showCmd = &cobra.Command{
	Use:     "show [path...]",
	Short:   "Display annotated file tree (default command)",
	GroupID: "main",
	Hidden:  true,
	Long: `Display directory trees with annotations from .info files.

This is the main functionality of treex. When no command is specified,
this command runs by default.

The command looks for .info files in the directory tree and displays
an annotated view of the file structure with descriptions.

Multiple paths can be specified to show multiple directories, similar to 
the Unix tree command:
  treex docs src                  # Show docs and src directories
  treex dir1 dir2 dir3           # Show multiple directories

OUTPUT FORMATS:

treex supports multiple output formats:
  --format=color    Full color terminal output (default)
  --format=minimal  Minimal color styling for basic terminals  
  --format=no-color Plain text output without colors

VIEW MODES:

Control which paths are displayed:
  --show=mix        Show annotations with contextual paths (default)
  --show=annotated  Show only annotated paths
  --show=all        Show all paths


Examples:
  treex                           # Full color output (default)
  treex --format=minimal .        # Minimal colors
  treex --format=no-color > tree.txt  # Plain text for files
  treex --show=annotated          # Show only annotated paths
  treex --show=all               # Show all paths
  treex docs src bin              # Show multiple directories
`,
	Args: cobra.ArbitraryArgs,
	RunE: runShowCmd,
}

func init() {
	// Add flags specific to the show command
	showCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	showCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")

	// New format system
	showCmd.Flags().StringVar(&outputFormat, "format", "color",
		"Output format: color, minimal, no-color (use --help for details)")

	// View mode flag
	showCmd.Flags().StringVar(&showMode, "show", "mix",
		"View mode: mix, annotated, all (use --help for details)")

	// Other flags
	showCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	showCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	showCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode (useful for terminals with rendering issues)")

	// Register the command with root
	rootCmd.AddCommand(showCmd)
}

// runShowCmd handles the CLI interface for the show command
func runShowCmd(cmd *cobra.Command, args []string) error {
	// Determine target paths
	var targetPaths []string

	// If --path flag is used, use that
	if path != "" {
		targetPaths = []string{path}
	} else if len(args) > 0 {
		// Use command line arguments
		targetPaths = args
	} else {
		// Default to current directory
		targetPaths = []string{"."}
	}

	// Validate format
	if outputFormat != "" {
		if _, err := format.ParseFormatString(outputFormat); err != nil {
			// Print available formats on error
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s", format.GetFormatHelp())
			return fmt.Errorf("invalid format: %s", outputFormat)
		}
	}

	// Validate view mode
	if showMode != "" {
		validModes := []string{"mix", "annotated", "all"}
		isValid := false
		for _, mode := range validModes {
			if showMode == mode {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid view mode: %s (must be 'mix', 'annotated', or 'all')", showMode)
		}
	}

	// Process each target path
	for i, targetPath := range targetPaths {
		// Add separator between multiple paths (like Unix tree command)
		if i > 0 {
			_, _ = cmd.OutOrStdout().Write([]byte("\n"))
		}

		// Resolve ignore file path relative to target path if it's a relative path
		resolvedIgnoreFile := ignoreFile
		if ignoreFile != "" && !filepath.IsAbs(ignoreFile) {
			resolvedIgnoreFile = filepath.Join(targetPath, ignoreFile)
		}

		options := app.RenderOptions{
			Verbose:      verbose,
			Format:       outputFormat, // New format system
			ViewMode:     showMode,
			IgnoreFile:   resolvedIgnoreFile,
			MaxDepth:     maxDepth,
			SafeMode:     safeMode,
		}

		// Call the main business logic
		result, err := app.RenderAnnotatedTree(targetPath, options)
		if err != nil {
			return fmt.Errorf("failed to display tree for %s: %w", targetPath, err)
		}

		// Output the result (conditionally handling verbose output)
		if options.Verbose && result.VerboseOutput != nil {
			printVerboseOutput(cmd, result.VerboseOutput)
		}

		// Write the tree output
		_, err = cmd.OutOrStdout().Write([]byte(result.Output))
		if err != nil {
			// If we can't write to output, return an error
			return fmt.Errorf("failed to write output for %s: %w", targetPath, err)
		}
		
		// Display warnings if any
		if len(result.Warnings) > 0 {
			printWarnings(cmd, result.Warnings)
		}
	}

	return nil
}

// printVerboseOutput formats and prints the structured verbose information
func printVerboseOutput(cmd *cobra.Command, verboseData *app.VerboseOutput) {
	// For verbose output, we'll ignore errors since they're not critical to functionality
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Analyzing directory: %s\n", verboseData.AnalyzedPath)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Verbose mode enabled - will show parsed .info structure")
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "=== Parsed Annotations ===")
	if len(verboseData.ParsedAnnotations) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No annotations found (no .info file or empty file)")
	} else {
		for path, annotation := range verboseData.ParsedAnnotations {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", path)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Notes: %s\n", annotation.Notes)
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "=== End Annotations ===")
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	if verboseData.TreeStructure != "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "=== File Tree Structure ===")
		_, _ = fmt.Fprint(cmd.OutOrStdout(), verboseData.TreeStructure) // Error can be ignored here as it's verbose output
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "=== End Tree Structure ===")
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "treex analysis of: %s\n", verboseData.AnalyzedPath)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Found %d annotations\n", verboseData.FoundAnnotations)
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
}

// getFormatListCmd creates a hidden command to list available formats
func getFormatListCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "list-formats",
		Hidden: true,
		Short:  "List all available output formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(format.GetFormatHelp())
			return nil
		},
	}
}

func init() {
	// Add hidden format listing command for development/debugging
	showCmd.AddCommand(getFormatListCmd())
}

// printWarnings displays warnings in a formatted way
func printWarnings(cmd *cobra.Command, warnings []string) {
	// Print a newline to separate from tree output
	_, _ = fmt.Fprintln(cmd.OutOrStderr())
	
	// Create a warning style using lipgloss
	warningStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Warning)
	
	// Create the warning header
	header := warningStyle.Render("⚠️  Warnings found in .info files:")
	_, _ = fmt.Fprintln(cmd.OutOrStderr(), header)
	
	// Print each warning with bullet point
	for _, warning := range warnings {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  • %s\n", warning)
	}
}
