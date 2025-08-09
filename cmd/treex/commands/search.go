package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "info-search <term>",
	Short: "Search for a term in all .info files",
	Long:  "Search recursively finds all .info files and searches for the given term in both paths and annotations",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

type searchResult struct {
	infoFile   string
	path       string
	annotation string
	score      int // Higher score = better match
}

func init() {
	searchCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	searchTerm := strings.ToLower(args[0])

	// Find all info files
	infoFiles, err := findSearchInfoFiles(".", infoFile)
	if err != nil {
		return fmt.Errorf("failed to find .info files: %w", err)
	}

	if len(infoFiles) == 0 {
		cmd.Printf("No %s files found\n", infoFile)
		return nil
	}

	// Search in all .info files
	var results []searchResult
	parser := info.NewParser()

	for _, infoFile := range infoFiles {
		annotations, _, err := parser.ParseFileWithWarnings(infoFile)
		if err != nil {
			cmd.PrintErrln(fmt.Sprintf("Warning: failed to parse %s: %v", infoFile, err))
			continue
		}

		// Search in each annotation
		for path, ann := range annotations {
			score := 0
			pathLower := strings.ToLower(path)
			annotationLower := strings.ToLower(ann.Notes)

			// Check if term is in path (higher priority)
			if strings.Contains(pathLower, searchTerm) {
				score += 10
				// Exact path match gets even higher score
				if pathLower == searchTerm {
					score += 20
				}
				// Path basename match gets bonus
				if strings.ToLower(filepath.Base(path)) == searchTerm {
					score += 10
				}
			}

			// Check if term is in annotation
			if strings.Contains(annotationLower, searchTerm) {
				score += 5
			}

			// Add result if there's a match
			if score > 0 {
				results = append(results, searchResult{
					infoFile:   infoFile,
					path:       path,
					annotation: ann.Notes,
					score:      score,
				})
			}
		}
	}

	// Check if no results
	if len(results) == 0 {
		cmd.Printf("No matches found for '%s'\n", searchTerm)
		return nil
	}

	// Sort results by score (highest first), then by path
	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].path < results[j].path
	})

	// Display results
	cmd.Printf("Found %d matches for '%s':\n\n", len(results), searchTerm)

	currentInfoFile := ""
	for _, result := range results {
		// Show info file header when it changes
		if result.infoFile != currentInfoFile {
			currentInfoFile = result.infoFile
			cmd.Printf("In %s:\n", result.infoFile)
		}

		// Highlight the search term in path and annotation
		highlightedPath := highlightTerm(result.path, searchTerm)
		highlightedAnnotation := highlightTerm(result.annotation, searchTerm)

		cmd.Printf("  %s: %s\n", highlightedPath, highlightedAnnotation)
	}

	return nil
}

// findSearchInfoFiles recursively finds all info files
func findSearchInfoFiles(root string, infoFileName string) ([]string, error) {
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

// highlightTerm highlights the search term in the text
func highlightTerm(text, term string) string {
	// Simple case-insensitive highlighting
	// In a terminal with color support, we could use ANSI codes
	// For now, we'll use asterisks to highlight

	// Find all occurrences case-insensitively
	lowerText := strings.ToLower(text)
	lowerTerm := strings.ToLower(term)

	result := text
	offset := 0

	for {
		index := strings.Index(lowerText[offset:], lowerTerm)
		if index == -1 {
			break
		}

		realIndex := offset + index
		// Insert highlighting around the match
		before := result[:realIndex]
		match := result[realIndex : realIndex+len(term)]
		after := result[realIndex+len(term):]

		result = before + "*" + match + "*" + after
		offset = realIndex + len(term) + 2 // +2 for the added asterisks
		lowerText = strings.ToLower(result)
	}

	return result
}
