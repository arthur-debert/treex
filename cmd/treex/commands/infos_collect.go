package commands

import (
	"fmt"
	"path/filepath"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/spf13/cobra"
)

var (
	collectDryRun bool
	collectPreserveOrder bool
)

// infosCollectCmd represents the infos-collect command
var infosCollectCmd = &cobra.Command{
	Use:     "infos-collect [path]",
	GroupID: "info",
	Short:   "Collect distributed .info files into a single root .info file",
	Long: `Collects all .info files from subdirectories and consolidates them into
a single .info file at the root directory.

This command:
- Recursively finds all .info files in subdirectories
- Merges their entries into the root .info file
- Adjusts paths to be relative to the root
- Resolves conflicts (same path in multiple files) by preferring the
  .info file closest to the referenced path
- Deletes the subdirectory .info files after successful collection

Example:
  treex infos-collect              # Collect in current directory
  treex infos-collect ~/project    # Collect in specific directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInfosCollect,
}

func init() {
	infosCollectCmd.Flags().BoolVar(&collectDryRun, "dry-run", false, 
		"Show what would be collected without modifying files")
	infosCollectCmd.Flags().BoolVar(&collectPreserveOrder, "preserve-order", false,
		"Preserve the original order of entries from each file")
	infosCollectCmd.Flags().StringVar(&infoFile, "info-file", ".info",
		"Use specified info file name instead of .info")
	
	rootCmd.AddCommand(infosCollectCmd)
}

func runInfosCollect(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Resolve to absolute path for clearer output
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Prepare options
	options := info.CollectOptions{
		InfoFileName:  infoFile,
		DryRun:        collectDryRun,
		PreserveOrder: collectPreserveOrder,
	}

	// Perform collection
	result, err := info.CollectInfoFiles(targetPath, options)
	if err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	// Display results
	if collectDryRun {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "DRY RUN - No files were modified")
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Collecting .info files in: %s\n", absPath)
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	if len(result.CollectedFiles) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No .info files found to collect.")
		return nil
	}

	// Show collected files
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Found %d .info file(s):\n", len(result.CollectedFiles))
	for _, file := range result.CollectedFiles {
		relPath, _ := filepath.Rel(absPath, file)
		if relPath == infoFile {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  • %s (root - will be updated)\n", relPath)
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  • %s\n", relPath)
		}
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// Show statistics
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Total entries collected: %d\n", result.TotalEntries)
	
	if len(result.ConflictResolutions) > 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Conflicts resolved: %d\n", len(result.ConflictResolutions))
		if verbose {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nConflict details:")
			for path, winner := range result.ConflictResolutions {
				winnerRel, _ := filepath.Rel(absPath, winner)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  • %s → kept from %s\n", path, winnerRel)
			}
		}
	}

	// Show errors if any
	if len(result.Errors) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStderr())
		_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Warnings:")
		for _, err := range result.Errors {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  • %v\n", err)
		}
	}

	// Show final status
	_, _ = fmt.Fprintln(cmd.OutOrStdout())
	if !collectDryRun {
		rootInfoPath := filepath.Join(absPath, infoFile)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Merged content written to: %s\n", infoFile)
		
		deletedCount := 0
		for _, file := range result.CollectedFiles {
			if file != rootInfoPath {
				deletedCount++
			}
		}
		if deletedCount > 0 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✓ Deleted %d child .info file(s)\n", deletedCount)
		}
	} else {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Preview of merged content:")
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "---")
		_, _ = fmt.Fprint(cmd.OutOrStdout(), result.MergedContent)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "---")
	}

	return nil
}