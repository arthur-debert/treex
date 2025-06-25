package cmd

import (
	"fmt"

	"github.com/adebert/treex/pkg/app"
	"github.com/adebert/treex/pkg/format"
	"github.com/spf13/cobra"
)

var (
	// Format-based flags (new system)
	outputFormat string
)

// showCmd represents the main tree display functionality
// This is also the default command when no subcommand is specified
var showCmd = &cobra.Command{
	Use:     "show [path]",
	Short:   "Display annotated file tree (default command)",
	GroupID: "main",
	Long: `Display directory trees with annotations from .info files.

This is the main functionality of treex. When no command is specified,
this command runs by default.

The command looks for .info files in the directory tree and displays
an annotated view of the file structure with descriptions.

OUTPUT FORMATS:

treex supports multiple output formats:
  --format=color    Full color terminal output (default)
  --format=minimal  Minimal color styling for basic terminals  
  --format=no-color Plain text output without colors

Legacy format flags (deprecated but supported):
  --no-color        Same as --format=no-color
  --minimal         Same as --format=minimal

Examples:
  treex                           # Full color output (default)
  treex --format=minimal .        # Minimal colors
  treex --format=no-color > tree.txt  # Plain text for files
  treex --no-color .              # Legacy flag (still works)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShowCmd,
}

func init() {
	// Add flags specific to the show command
	showCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	showCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze (defaults to current directory)")

	// New format system
	showCmd.Flags().StringVar(&outputFormat, "format", "color",
		"Output format: color, minimal, no-color (use --help for details)")

	// Legacy format flags (deprecated but supported for backward compatibility)
	showCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output (deprecated: use --format=no-color)")
	showCmd.Flags().BoolVar(&minimal, "minimal", false, "Use minimal styling (deprecated: use --format=minimal)")

	// Other flags
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

	// Validate format
	if outputFormat != "" {
		if _, err := format.ParseFormatString(outputFormat); err != nil {
			// Print available formats on error
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n\n", err)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s", format.GetFormatHelp())
			return fmt.Errorf("invalid format: %s", outputFormat)
		}
	}

	// Warn about deprecated flags
	if cmd.Flags().Changed("no-color") {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --no-color is deprecated, use --format=no-color instead\n")
	}
	if cmd.Flags().Changed("minimal") {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --minimal is deprecated, use --format=minimal instead\n")
	}

	// Create configuration from flags
	options := app.RenderOptions{
		Verbose:    verbose,
		NoColor:    noColor,      // Legacy support
		Minimal:    minimal,      // Legacy support
		Format:     outputFormat, // New format system
		IgnoreFile: ignoreFile,
		MaxDepth:   maxDepth,
		SafeMode:   safeMode,
	}

	// Call the main business logic
	result, err := app.RenderAnnotatedTree(targetPath, options)
	if err != nil {
		return fmt.Errorf("failed to display tree: %w", err)
	}

	// Output the result (conditionally handling verbose output)
	if options.Verbose && result.VerboseOutput != nil {
		printVerboseOutput(cmd, result.VerboseOutput)
	}
	// Use cmd.Print or fmt.Fprint(cmd.OutOrStdout(), ...) to respect output redirection
	_, err = cmd.OutOrStdout().Write([]byte(result.Output))
	if err != nil {
		// If we can't write to output, return an error
		return fmt.Errorf("failed to write output: %w", err)
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
			if annotation.Title != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Title: %s\n", annotation.Title)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Description: %s\n", annotation.Description)
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
