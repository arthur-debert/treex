package geninfo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/edit/addinfo"
)

// TreeEntry represents a parsed entry from tree-like input
type TreeEntry struct {
	Path        string
	Description string
	IsDir       bool
}

// GenerateInfoFromTree parses a tree-like input file and generates .info files
func GenerateInfoFromTree(inputFile string) error {
	entries, err := ParseTreeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse tree file: %w", err)
	}

	return generateInfoFromEntries(entries)
}

// GenerateInfoFromReader parses tree-like content from a reader and generates .info files
func GenerateInfoFromReader(reader io.Reader) error {
	entries, err := parseTreeReader(reader)
	if err != nil {
		return fmt.Errorf("failed to parse tree content: %w", err)
	}

	return generateInfoFromEntries(entries)
}

// generateInfoFromEntries is the common logic for generating .info files from entries
func generateInfoFromEntries(entries []TreeEntry) error {
	// Group entries by their parent directories
	infoFiles := make(map[string][]TreeEntry)

	for _, entry := range entries {
		// Determine the parent directory
		parentDir := filepath.Dir(entry.Path)
		if parentDir == "." {
			parentDir = ""
		}

		// Check if the path exists
		if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", entry.Path)
		}

		// Add to the appropriate .info file
		infoFiles[parentDir] = append(infoFiles[parentDir], entry)
	}

	// Generate .info files
	for dir, dirEntries := range infoFiles {
		if err := GenerateInfoFile(dir, dirEntries); err != nil {
			return fmt.Errorf("failed to generate .info file for directory %s: %w", dir, err)
		}
	}

	return nil
}

// ParseTreeFile parses a tree-like input file and extracts paths and descriptions
func ParseTreeFile(inputFile string) ([]TreeEntry, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	return parseTreeReader(file)
}

// parseTreeReader parses tree-like content from a reader and extracts paths and descriptions
func parseTreeReader(reader io.Reader) ([]TreeEntry, error) {
	var entries []TreeEntry
	scanner := bufio.NewScanner(reader)
	var pathStack []string // Stack to keep track of current path components

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the tree line
		entry, depth, err := ParseTreeLine(line, pathStack)
		if err != nil {
			continue // Skip malformed lines
		}

		if entry != nil {
			// Handle path building
			// The first entry (usually no tree connectors) is the root
			// Entries with tree connectors (├── └──) at depth 0 are children of root
			// Entries with vertical connectors (│   └──) at depth 1+ are nested

			if len(pathStack) == 0 {
				// First entry is the root
				pathStack = append(pathStack, entry.Path)
				// Don't modify the path for root entry
			} else {
				// Build the full path before updating pathStack
				originalPath := entry.Path

				// Build full path based on current depth
				if depth == 0 {
					// Direct child of root
					entry.Path = filepath.Join(pathStack[0], originalPath)
				} else {
					// Nested entry - use pathStack up to and including current depth
					// For depth 1, we want pathStack[0] and pathStack[1]
					pathComponents := make([]string, 0, depth+2)
					for i := 0; i <= depth && i < len(pathStack); i++ {
						pathComponents = append(pathComponents, pathStack[i])
					}
					pathComponents = append(pathComponents, originalPath)
					entry.Path = filepath.Join(pathComponents...)
				}

				// Now update pathStack for future entries
				if depth >= len(pathStack)-1 {
					// Going deeper - append to stack
					pathStack = append(pathStack, originalPath)
				} else {
					// Going up or same level - adjust stack
					// Keep root (index 0) and update at depth+1
					if depth+1 < len(pathStack) {
						pathStack = pathStack[:depth+2]
						pathStack[depth+1] = originalPath
					} else {
						pathStack = append(pathStack[:depth+1], originalPath)
					}
				}
			}

			entries = append(entries, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return entries, nil
}

// ParseTreeLine parses a single line from tree-like input and returns the entry and its depth
func ParseTreeLine(line string, currentPath []string) (*TreeEntry, int, error) {
	// Skip empty lines
	if strings.TrimSpace(line) == "" {
		return nil, 0, nil
	}

	// Track depth by counting tree connector sets (│   , ├── , └── )
	depth := 0
	cleanLine := line

	// Remove tree connectors and count depth
	for {
		if strings.HasPrefix(cleanLine, "│   ") {
			depth++
			cleanLine = strings.TrimPrefix(cleanLine, "│   ")
		} else if strings.HasPrefix(cleanLine, "│  ") {
			depth++
			cleanLine = strings.TrimPrefix(cleanLine, "│  ")
		} else if strings.HasPrefix(cleanLine, "    ") && !strings.HasPrefix(strings.TrimLeft(cleanLine, " "), "├") && !strings.HasPrefix(strings.TrimLeft(cleanLine, " "), "└") {
			// Some trees use spaces instead of │, but only count if not followed by tree chars
			depth++
			cleanLine = cleanLine[4:]
		} else if strings.HasPrefix(cleanLine, "├──") || strings.HasPrefix(cleanLine, "└──") {
			// Check if there's content after the connector
			if strings.HasPrefix(cleanLine, "├── ") {
				cleanLine = strings.TrimPrefix(cleanLine, "├── ")
				break
			} else if strings.HasPrefix(cleanLine, "└── ") {
				cleanLine = strings.TrimPrefix(cleanLine, "└── ")
				break
			} else if cleanLine == "├──" || cleanLine == "└──" {
				// Just the connector without content
				return nil, depth, nil
			} else {
				// No space after connector, not valid
				break
			}
		} else if strings.HasPrefix(cleanLine, "├─ ") {
			cleanLine = strings.TrimPrefix(cleanLine, "├─ ")
			break
		} else if strings.HasPrefix(cleanLine, "└─ ") {
			cleanLine = strings.TrimPrefix(cleanLine, "└─ ")
			break
		} else {
			// No more tree connectors
			break
		}
	}

	// If the line starts with no tree characters at all, it's the root
	if cleanLine == line && len(currentPath) == 0 {
		// This is the root directory line
		parts := strings.SplitN(cleanLine, " ", 2)
		name := strings.TrimSpace(parts[0])
		if name == "" {
			return nil, 0, fmt.Errorf("empty root name")
		}

		var description string
		if len(parts) > 1 {
			description = strings.TrimSpace(parts[1])
		}

		// Check if it's a directory (ends with /)
		isDir := strings.HasSuffix(name, "/")
		name = strings.TrimSuffix(name, "/")

		return &TreeEntry{
			Path:        name,
			Description: description,
			IsDir:       isDir,
		}, depth, nil
	}

	// If we have an empty cleanLine after removing connectors, skip it
	if strings.TrimSpace(cleanLine) == "" {
		return nil, depth, nil
	}

	// Split by space to separate name and description
	// First, find the file/directory name (first word)
	parts := strings.SplitN(cleanLine, " ", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return nil, depth, fmt.Errorf("no file/directory name found")
	}

	name := strings.TrimSpace(parts[0])
	var description string
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	// Check if it's a directory (ends with /)
	isDir := strings.HasSuffix(name, "/")
	name = strings.TrimSuffix(name, "/")

	return &TreeEntry{
		Path:        name,
		Description: description,
		IsDir:       isDir,
	}, depth, nil
}

// GenerateInfoFile creates or updates a .info file in the specified directory
func GenerateInfoFile(dir string, entries []TreeEntry) error {
	infoPath := filepath.Join(dir, ".info")
	if dir == "" {
		infoPath = ".info"
	}

	// Parse existing .info file to get existing annotations
	parser := info.NewParser()
	existingAnnotations, err := parser.ParseFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to parse existing .info file: %w", err)
	}

	// Merge new entries with existing ones
	annotations := make(map[string]*types.Annotation)
	for path, annotation := range existingAnnotations {
		annotations[path] = annotation
	}

	// Add new entries
	for _, entry := range entries {
		// Get the relative path from the directory
		relPath := entry.Path
		if dir != "" {
			var err error
			relPath, err = filepath.Rel(dir, entry.Path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
		}

		// Add trailing slash for directories
		if entry.IsDir {
			relPath = relPath + "/"
		}

		// Only add if not already present
		if _, exists := annotations[relPath]; !exists && entry.Description != "" {
			annotations[relPath] = &types.Annotation{
				Path:  relPath,
				Notes: entry.Description,
			}
		}
	}

	// Write the updated .info file
	return addinfo.WriteInfoFile(infoPath, annotations)
}
