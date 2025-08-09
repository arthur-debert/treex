package commands

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/adebert/treex/pkg/app"
	"github.com/adebert/treex/pkg/config"
	"github.com/adebert/treex/pkg/core/firstuse"
	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/plugins"
	"github.com/adebert/treex/pkg/core/plugins/builtin"
	"github.com/adebert/treex/pkg/core/query"
	"github.com/adebert/treex/pkg/core/tree"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/display/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Format-based flags (new system)
	outputFormat string
	// View mode flag
	infoModeFlag string
	// Overlay plugins flag
	overlayPlugins []string
	// Query system integration
	queryCLI *query.CLIIntegration
)

//go:embed show.help.txt
var showHelp string

// showCmd represents the main tree display functionality
// This is also the default command when no subcommand is specified
var showCmd = &cobra.Command{
	Use:    "show [path...]",
	Short:  "Display annotated file tree (default command)",
	Hidden: true,
	Long:   showHelp,
	Args:   cobra.ArbitraryArgs,
	RunE:   runShowCmd,
}

func init() {
	// Add flags specific to the show command
	showCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output including parsed .info file structure")
	_ = showCmd.Flags().MarkHidden("verbose")

	// New format system
	showCmd.Flags().StringVarP(&outputFormat, "format", "f", "color",
		"Output format: color, no-color, markdown (use --help for details)")

	// View mode flag
	showCmd.Flags().StringVar(&infoModeFlag, "info-mode", "mix",
		"View mode: mix, annotated, all (use --help for details)")

	// Overlay plugins flag
	showCmd.Flags().StringSliceVar(&overlayPlugins, "overlay", []string{},
		"Show additional file information (size, date-created, date-modified, lc)")

	// Other flags
	showCmd.Flags().StringVar(&ignoreFile, "ignore-file", ".gitignore", "Use specified ignore file (default is .gitignore)")
	showCmd.Flags().BoolVar(&noIgnore, "no-ignore", false, "Don't use any ignore file")
	showCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	showCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	showCmd.Flags().BoolVar(&infoIgnoreWarnings, "info-ignore-warnings", false, "Don't print warnings for non-existent paths in .info files")

	// Initialize query system
	if err := query.InitializeQuerySystem(); err != nil {
		// Log error but don't fail - query system is optional
		_, _ = fmt.Fprintf(showCmd.ErrOrStderr(), "Warning: failed to initialize query system: %v\n", err)
	} else {
		// Register query flags
		queryCLI = query.NewCLIIntegration(query.GetGlobalRegistry())
		if err := queryCLI.RegisterFlags(showCmd); err != nil {
			_, _ = fmt.Fprintf(showCmd.ErrOrStderr(), "Warning: failed to register query flags: %v\n", err)
		}
	}

	// Register the command with root
	rootCmd.AddCommand(showCmd)
}

// runShowCmd handles the CLI interface for the show command
func runShowCmd(cmd *cobra.Command, args []string) error {
	// Load configuration from treex.yaml
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		// Continue with defaults silently
		cfg = config.DefaultConfig()
	}

	// Re-register renderers with loaded configuration
	app.RegisterDefaultRenderersWithConfig(cfg)

	// Determine target paths
	var targetPaths []string

	if len(args) > 0 {
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
	if infoModeFlag != "" {
		validModes := []string{"mix", "annotated", "all"}
		isValid := false
		for _, mode := range validModes {
			if infoModeFlag == mode {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid view mode: %s (must be 'mix', 'annotated', or 'all')", infoModeFlag)
		}
	}

	// Initialize and validate plugins
	if len(overlayPlugins) > 0 {
		registry := plugins.GetGlobalRegistry()
		
		// Register built-in plugins if not already registered
		_ = builtin.RegisterBuiltinPlugins(registry)
		// Ignore errors as plugins might already be registered
		
		// Validate that all requested plugins exist
		if err := registry.ValidatePlugins(overlayPlugins); err != nil {
			return fmt.Errorf("plugin validation failed: %w", err)
		}
	}

	// Parse query from flags
	var userQuery *query.Query
	if queryCLI != nil {
		parsedQuery, err := queryCLI.BuildQuery(cmd.Flags())
		if err != nil {
			return fmt.Errorf("failed to parse query: %w", err)
		}
		userQuery = parsedQuery
		
	}

	// Process each target path
	for i, targetPath := range targetPaths {
		// Add separator between multiple paths (like Unix tree command)
		if i > 0 {
			_, _ = cmd.OutOrStdout().Write([]byte("\n"))
		}

		// Resolve ignore file path relative to target path if it's a relative path
		resolvedIgnoreFile := ignoreFile
		if noIgnore {
			// If --no-ignore is set, use empty string to disable ignore file
			resolvedIgnoreFile = ""
		} else if ignoreFile != "" && !filepath.IsAbs(ignoreFile) {
			resolvedIgnoreFile = filepath.Join(targetPath, ignoreFile)
		}

		options := app.RenderOptions{
			Verbose:      verbose,
			Format:       outputFormat, // New format system
			ViewMode:     infoModeFlag,
			IgnoreFile:   resolvedIgnoreFile,
			InfoFileName: infoFile,
			MaxDepth:     maxDepth,
			Config:       cfg,
			OverlayPlugins: overlayPlugins,
			Query:         userQuery,
		}

		// Call the main business logic
		result, err := app.RenderAnnotatedTree(targetPath, options)
		if err != nil {
			return fmt.Errorf("failed to display tree for %s: %w", targetPath, err)
		}

		// Check if this is a first-time user scenario (no annotations found and no query)
		if result.Stats != nil && result.Stats.AnnotationsFound == 0 && infoModeFlag != "annotated" && userQuery == nil {
			// Generate first-use message using the template
			firstUseMessage, err := generateFirstUseMessageForPath(targetPath, options)
			if err == nil && firstUseMessage != "" {
				result.Output = firstUseMessage
			}
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
		if len(result.Warnings) > 0 && !infoIgnoreWarnings {
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

// generateFirstUseMessageForPath generates a first-time user message for the given path
func generateFirstUseMessageForPath(targetPath string, options app.RenderOptions) (string, error) {
	// Build the tree first to analyze directory content
	annotations := make(map[string]*types.Annotation) // Empty for first-time users
	
	// Parse view mode
	viewMode := types.ViewModeMix
	if options.ViewMode != "" {
		parsedMode, err := types.ParseViewMode(options.ViewMode)
		if err != nil {
			return "", err
		}
		viewMode = parsedMode
	}
	
	viewOptions := types.ViewOptions{
		Mode: viewMode,
	}
	
	// Build tree
	var root *types.Node
	var err error
	if options.IgnoreFile != "" || options.MaxDepth != -1 {
		builder, err := tree.NewViewBuilderWithOptions(targetPath, annotations, options.IgnoreFile, options.MaxDepth, viewOptions)
		if err != nil {
			return "", err
		}
		root, err = builder.Build()
		if err != nil {
			return "", err
		}
	} else {
		builder := tree.NewViewBuilder(targetPath, annotations, viewOptions)
		root, err = builder.Build()
		if err != nil {
			return "", err
		}
	}
	
	// Check if directory has any content
	hasContent := len(root.Children) > 0
	
	var examples []firstuse.CommonPath
	
	if hasContent {
		// Directory has content - use actual files/folders as examples
		examples = firstuse.FindExamplesInPath(targetPath, 2)
		if len(examples) == 0 {
			examples = firstuse.GetFallbackExamples(targetPath, 2)
		}
	} else {
		// Directory is empty - use generic examples
		examples = []firstuse.CommonPath{
			{Path: "src/", Annotation: "Source code directory"},
			{Path: "README.md", Annotation: "Project documentation"},
		}
	}
	
	// Use the template-based approach
	infoFileName := options.InfoFileName
	if infoFileName == "" {
		infoFileName = ".info"
	}
	
	return GenerateFirstUseMessage(root, examples, infoFileName, options.Format)
}
