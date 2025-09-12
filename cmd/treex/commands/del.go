package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/spf13/cobra"
)

var delCmd = &cobra.Command{
	Use:   "del <path>",
	Short: "Delete annotation for a path from .info file",
	Long:  "Delete the annotation for a specific path from the .info file without affecting the actual file or directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runDel,
}

func init() {
	delCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	infoCmd.AddCommand(delCmd)
}

func runDel(cmd *cobra.Command, args []string) error {
	pathToRemove := args[0]

	// Find the info file in the current directory
	infoPath := filepath.Join(".", infoFile)

	// Check if .info file exists
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		return fmt.Errorf("no %s file found in current directory", infoFile)
	}

	// Parse existing annotations
	parser := info.NewParser()
	annotations, _, err := parser.ParseFileWithWarnings(infoPath)
	if err != nil {
		return fmt.Errorf("failed to parse .info file: %w", err)
	}

	// Check if the path exists in annotations
	found := false
	for path := range annotations {
		if path == pathToRemove {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no annotation found for path: %s", pathToRemove)
	}

	// Remove the annotation
	delete(annotations, pathToRemove)

	// Write back to file
	if err := writeInfoFile(infoPath, annotations); err != nil {
		return fmt.Errorf("failed to write .info file: %w", err)
	}

	fmt.Printf("Deleted annotation for: %s\n", pathToRemove)
	return nil
}

// writeInfoFile writes annotations back to .info file
func writeInfoFile(path string, annotations map[string]*types.Annotation) error {
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
