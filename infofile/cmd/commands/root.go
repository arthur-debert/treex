package commands

import (
	"encoding/json"
	"fmt"
	"os"

	infofile "github.com/jwaldrip/treex/infofile"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Global flags
var (
	outputJSON bool
	infoFile   string
)

// Version information (set by main)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// SetVersionInfo sets version information for the CLI
func SetVersionInfo(version, commit, buildDate string) {
	Version = version
	Commit = commit
	BuildDate = buildDate
}

// Common functionality for all commands
type InfoCLI struct {
	api *infofile.InfoAPI
}

func NewInfoCLI() *InfoCLI {
	return &InfoCLI{
		api: infofile.NewInfoAPI(afero.NewOsFs()),
	}
}

// NewRootCommand creates the root infofile command
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "infofile",
		Short: "Manage .info file annotations",
		Long: `infofile is a command-line tool for managing file annotations in .info files.
It provides commands to add, remove, update, and query annotations that describe
files and directories in your project.`,
		SilenceUsage: true,
		Version:      getVersionString(),
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVar(&infoFile, "info-file", ".info", "Specify .info file for add/remove/update operations")

	// Add all subcommands
	cli := NewInfoCLI()
	rootCmd.AddCommand(
		cli.newGatherCommand(),
		cli.newDistributeCommand(),
		cli.newShowCommand(),
		cli.newAddCommand(),
		cli.newRemoveCommand(),
		cli.newUpdateCommand(),
		cli.newValidateCommand(),
		cli.newCleanCommand(),
		cli.newGetCommand(),
	)

	return rootCmd
}

// Helper functions for output formatting

func outputResult(data interface{}) error {
	if outputJSON {
		return outputJSON_format(data)
	}
	return outputText(data)
}

func outputJSON_format(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputText(data interface{}) error {
	switch v := data.(type) {
	case map[string]infofile.Annotation:
		return outputAnnotationsMap(v)
	case []infofile.Annotation:
		return outputAnnotationsList(v)
	case *infofile.ValidationResult:
		return outputValidationResult(v)
	case *infofile.CleanResult:
		return outputCleanResult(v)
	case *infofile.Annotation:
		return outputSingleAnnotation(v)
	case string:
		fmt.Println(v)
		return nil
	default:
		return fmt.Errorf("unsupported output type: %T", data)
	}
}

func outputAnnotationsMap(annotations map[string]infofile.Annotation) error {
	if len(annotations) == 0 {
		fmt.Println("No annotations found.")
		return nil
	}

	for path, annotation := range annotations {
		fmt.Printf("%-40s %s\n", path, annotation.Annotation)
	}
	return nil
}

func outputAnnotationsList(annotations []infofile.Annotation) error {
	if len(annotations) == 0 {
		fmt.Println("No annotations found.")
		return nil
	}

	for _, annotation := range annotations {
		fmt.Printf("%-40s %s\n", annotation.Path, annotation.Annotation)
	}
	return nil
}

func outputSingleAnnotation(annotation *infofile.Annotation) error {
	fmt.Printf("%s: %s\n", annotation.Path, annotation.Annotation)
	return nil
}

func outputValidationResult(result *infofile.ValidationResult) error {
	if len(result.Issues) == 0 {
		fmt.Println("✓ All .info files are valid.")
		return nil
	}

	fmt.Printf("Found %d validation issues:\n\n", len(result.Issues))
	for _, issue := range result.Issues {
		fmt.Printf("❌ %s:%d - %s\n", issue.InfoFile, issue.LineNum, issue.Message)
		if issue.Path != "" {
			fmt.Printf("   Path: %s\n", issue.Path)
		}
		if issue.Suggestion != "" {
			fmt.Printf("   💡 %s\n", issue.Suggestion)
		}
		fmt.Println()
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("  Valid files: %d\n", len(result.ValidFiles))
	fmt.Printf("  Invalid files: %d\n", len(result.InvalidFiles))
	return nil
}

func outputCleanResult(result *infofile.CleanResult) error {
	if len(result.RemovedAnnotations) == 0 {
		fmt.Println("✓ No cleanup needed - all annotations are valid.")
		return nil
	}

	fmt.Printf("Cleaned %d annotations:\n\n", len(result.RemovedAnnotations))
	for _, annotation := range result.RemovedAnnotations {
		fmt.Printf("❌ Removed: %s (%s)\n", annotation.Path, annotation.Annotation)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Invalid paths removed: %d\n", result.Summary.InvalidPathsRemoved)
	fmt.Printf("  Duplicates removed: %d\n", result.Summary.DuplicatesRemoved)
	fmt.Printf("  Files updated: %d\n", len(result.UpdatedFiles))

	if len(result.UpdatedFiles) > 0 {
		fmt.Printf("\nUpdated files:\n")
		for _, file := range result.UpdatedFiles {
			fmt.Printf("  %s\n", file)
		}
	}

	return nil
}

// Helper to get path argument with default
func getPathArg(args []string, defaultPath string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return defaultPath
}

// getVersionString returns a formatted version string
func getVersionString() string {
	if Version == "dev" {
		return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
	}
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
}
