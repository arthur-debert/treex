package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/spf13/cobra"
)

var (
	forceSync bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Remove annotations for non-existent paths from .info files",
	Long:  "Sync scans all .info files and removes annotations for paths that no longer exist",
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&forceSync, "force", false, "Remove stale annotations without confirmation")
	syncCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	infoCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Find all info files in the current directory tree
	infoFiles, err := findInfoFiles(".", infoFile)
	if err != nil {
		return fmt.Errorf("failed to find .info files: %w", err)
	}

	if len(infoFiles) == 0 {
		cmd.Printf("No %s files found\n", infoFile)
		return nil
	}

	// Check each .info file for stale annotations
	staleAnnotations := make(map[string][]string) // infoFile -> []stalePaths

	parser := info.NewParser()
	for _, infoFile := range infoFiles {
		annotations, _, err := parser.ParseFileWithWarnings(infoFile)
		if err != nil {
			cmd.PrintErrln(fmt.Sprintf("Warning: failed to parse %s: %v", infoFile, err))
			continue
		}

		// Get directory of the .info file
		infoDir := filepath.Dir(infoFile)

		// Check each annotation path
		var stalePaths []string
		for annotPath := range annotations {
			// Resolve the full path relative to the .info file location
			fullPath := filepath.Join(infoDir, annotPath)

			// Check if the path exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				stalePaths = append(stalePaths, annotPath)
			}
		}

		if len(stalePaths) > 0 {
			staleAnnotations[infoFile] = stalePaths
		}
	}

	// If no stale annotations found
	if len(staleAnnotations) == 0 {
		cmd.Println("No stale annotations found")
		return nil
	}

	// Display stale annotations
	cmd.Printf("Found %d stale annotations:\n", countStaleAnnotations(staleAnnotations))
	for infoFile, paths := range staleAnnotations {
		for _, path := range paths {
			cmd.Printf("  - %s: for %s annotation\n", infoFile, path)
		}
	}
	cmd.Println()

	// Ask for confirmation unless --force is used
	if !forceSync {
		cmd.Print("Remove them:\n  (Y) Yes\n  (N) No\n? ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" {
			cmd.Println("Cancelled")
			return nil
		}
	}

	// Remove stale annotations
	for infoFile, stalePaths := range staleAnnotations {
		if err := removeStaleAnnotations(infoFile, stalePaths); err != nil {
			cmd.PrintErrln(fmt.Sprintf("Error updating %s: %v", infoFile, err))
			continue
		}
		cmd.Printf("Updated %s\n", infoFile)
	}

	return nil
}

// findInfoFiles recursively finds all info files in the given directory
func findInfoFiles(root string, infoFileName string) ([]string, error) {
	var infoFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		// Check if it's an info file
		if !info.IsDir() && info.Name() == infoFileName {
			infoFiles = append(infoFiles, path)
		}

		return nil
	})

	return infoFiles, err
}

// countStaleAnnotations counts total number of stale annotations
func countStaleAnnotations(staleAnnotations map[string][]string) int {
	count := 0
	for _, paths := range staleAnnotations {
		count += len(paths)
	}
	return count
}

// removeStaleAnnotations removes the specified paths from the .info file
func removeStaleAnnotations(infoFile string, stalePaths []string) error {
	parser := info.NewParser()
	annotations, _, err := parser.ParseFileWithWarnings(infoFile)
	if err != nil {
		return err
	}

	// Remove stale paths
	for _, path := range stalePaths {
		delete(annotations, path)
	}

	// Write back to file
	return writeSyncInfoFile(infoFile, annotations)
}

// writeSyncInfoFile writes annotations back to .info file
func writeSyncInfoFile(path string, annotations map[string]*types.Annotation) error {
	var lines []string

	// Sort paths for consistent output
	var paths []string
	for p := range annotations {
		paths = append(paths, p)
	}

	// Simple alphabetical sort
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[i] > paths[j] {
				paths[i], paths[j] = paths[j], paths[i]
			}
		}
	}

	// Build content
	for _, p := range paths {
		ann := annotations[p]
		if ann.Notes == "" {
			continue
		}

		// Check if path contains spaces
		if strings.Contains(p, " ") {
			lines = append(lines, fmt.Sprintf("%s: %s", p, ann.Notes))
		} else {
			lines = append(lines, fmt.Sprintf("%s %s", p, ann.Notes))
		}
	}

	// Write to file
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}
