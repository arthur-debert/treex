package maketree

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/info"
)

// MakeTreeOptions contains configuration options for creating file trees
type MakeTreeOptions struct {
	Force      bool // Overwrite existing files/directories
	DryRun     bool // Don't actually create files, just show what would be created
	CreateInfo bool // Create a master .info file (default: true)
}

// MakeResult contains information about what was created
type MakeResult struct {
	CreatedDirs     []string
	CreatedFiles    []string
	SkippedPaths    []string
	InfoFileCreated bool
	DryRun          bool
}

// InputSource represents the type of input being used
type InputSource int

const (
	SourceTreeText InputSource = iota // Tree-like text input
	SourceInfoFile                    // .info file input
)

// TreeStructure represents a parsed tree structure
type TreeStructure struct {
	RootPath string
	Entries  []TreeEntry
	Source   InputSource
}

// TreeEntry represents a single entry in the tree structure
type TreeEntry struct {
	Path         string
	Description  string
	IsDir        bool
	RelativePath string // Path relative to root
}

// MakeTreeFromFile creates a file tree structure from either a tree-like text file or .info file
func MakeTreeFromFile(inputFile string, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	// Determine input type and parse accordingly
	treeStructure, err := parseInputFile(inputFile, rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input file: %w", err)
	}

	return makeTreeStructure(treeStructure, options)
}

// MakeTreeFromReader creates a file tree structure from tree-like text content from a reader
func MakeTreeFromReader(reader io.Reader, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	// Read all content from reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	return MakeTreeFromText(string(content), rootDir, options)
}

// MakeTreeFromText creates a file tree structure from tree-like text content
func MakeTreeFromText(content string, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	treeStructure, err := parseTreeText(content, rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tree text: %w", err)
	}

	return makeTreeStructure(treeStructure, options)
}

// parseInputFile determines the input type and parses accordingly
func parseInputFile(inputFile, rootDir string) (*TreeStructure, error) {
	// Check if it's a .info file
	if filepath.Base(inputFile) == ".info" {
		return parseInfoFile(inputFile, rootDir)
	}

	// Otherwise, treat as tree-like text
	return parseTreeFile(inputFile, rootDir)
}

// parseInfoFile parses a .info file and converts it to a tree structure
func parseInfoFile(infoFile, rootDir string) (*TreeStructure, error) {
	annotations, err := info.ParseDirectory(filepath.Dir(infoFile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse .info file: %w", err)
	}

	var entries []TreeEntry
	for path, annotation := range annotations {
		// Determine if it's a directory (has trailing slash or exists as directory)
		isDir := strings.HasSuffix(path, "/")
		cleanPath := strings.TrimSuffix(path, "/")

		entry := TreeEntry{
			Path:         filepath.Join(rootDir, cleanPath),
			Description:  annotation.Description,
			IsDir:        isDir,
			RelativePath: cleanPath,
		}
		entries = append(entries, entry)
	}

	return &TreeStructure{
		RootPath: rootDir,
		Entries:  entries,
		Source:   SourceInfoFile,
	}, nil
}

// parseTreeFile parses a tree-like text file
func parseTreeFile(inputFile, rootDir string) (*TreeStructure, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() {
		// Ignore close errors in defer to avoid masking the main error
		_ = file.Close()
	}()

	content, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	return parseTreeText(string(content), rootDir)
}

// parseTreeText parses tree-like text content (reuses logic from info package)
func parseTreeText(content, rootDir string) (*TreeStructure, error) {
	lines := strings.Split(content, "\n")
	var entries []TreeEntry
	var pathStack []string

	for _, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the tree line (reuse logic from info package)
		entry, depth, err := parseTreeLine(line, pathStack)
		if err != nil {
			continue // Skip malformed lines
		}

		if entry != nil {
			// Handle path building (similar to info.parseTreeFile)
			if len(pathStack) == 0 {
				// First entry is the root - it's always a directory if it has children
				pathStack = append(pathStack, entry.Name)
				entry.IsDir = true // Root is always a directory
			} else {
				// Adjust pathStack based on depth
				targetLength := depth + 1

				if targetLength < len(pathStack) {
					pathStack = pathStack[:targetLength]
				}

				if targetLength == len(pathStack) {
					pathStack = append(pathStack, entry.Name)
				} else {
					pathStack[targetLength-1] = entry.Name
					pathStack = pathStack[:targetLength]
				}
			}

			// Create the full path
			relativePath := strings.Join(pathStack, "/")
			fullPath := filepath.Join(rootDir, relativePath)

			entries = append(entries, TreeEntry{
				Path:         fullPath,
				Description:  entry.Description,
				IsDir:        entry.IsDir,
				RelativePath: relativePath,
			})
		}
	}

	return &TreeStructure{
		RootPath: rootDir,
		Entries:  entries,
		Source:   SourceTreeText,
	}, nil
}

// parseTreeLine parses a single line from the tree format
func parseTreeLine(line string, currentPath []string) (*treeLineEntry, int, error) {
	depth := 0

	// Calculate depth based on indentation and tree characters
	runes := []rune(line)
	i := 0

	// Count vertical connectors (│) to determine depth
	for i < len(runes) {
		char := runes[i]
		if char == ' ' || char == '\t' {
			i++
			continue
		} else if char == '│' {
			// Vertical connector - increment depth and skip
			depth++
			i++
			// Skip any following spaces
			for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
				i++
			}
		} else if char == '├' || char == '└' {
			// Horizontal connectors - this is the actual entry line
			i++
			// Skip connector characters (─)
			for i < len(runes) && runes[i] == '─' {
				i++
			}
			// Skip any following spaces
			for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
				i++
			}
			break
		} else {
			// No connectors, this is a root-level entry
			break
		}
	}

	// Extract the remaining content (path and description)
	if i >= len(runes) {
		return nil, depth, nil // No content, just connectors
	}

	content := strings.TrimSpace(string(runes[i:]))
	if content == "" {
		return nil, depth, nil
	}

	// Split path and description
	parts := strings.SplitN(content, " ", 2)
	if len(parts) < 1 {
		return nil, depth, fmt.Errorf("invalid line format")
	}

	name := strings.TrimSpace(parts[0])
	description := ""
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	// Remove trailing slash to normalize directory names
	isDir := strings.HasSuffix(name, "/")
	name = strings.TrimSuffix(name, "/")

	if name == "" {
		return nil, depth, fmt.Errorf("empty path")
	}

	return &treeLineEntry{
		Name:        name,
		Description: description,
		IsDir:       isDir,
	}, depth, nil
}

// treeLineEntry represents a parsed line from tree text
type treeLineEntry struct {
	Name        string
	Description string
	IsDir       bool
}

// makeTreeStructure creates the actual file/directory structure
func makeTreeStructure(treeStructure *TreeStructure, options MakeTreeOptions) (*MakeResult, error) {
	result := &MakeResult{
		DryRun: options.DryRun,
	}

	// Create directories first, then files
	for _, entry := range treeStructure.Entries {
		if entry.IsDir {
			if err := createDirectory(entry.Path, options, result); err != nil {
				return result, fmt.Errorf("failed to create directory %s: %w", entry.Path, err)
			}
		}
	}

	for _, entry := range treeStructure.Entries {
		if !entry.IsDir {
			if err := createFile(entry.Path, options, result); err != nil {
				return result, fmt.Errorf("failed to create file %s: %w", entry.Path, err)
			}
		}
	}

	// Create master .info file if requested and there are entries
	if options.CreateInfo && len(treeStructure.Entries) > 0 {
		if err := createMasterInfoFile(treeStructure, options, result); err != nil {
			return result, fmt.Errorf("failed to create master .info file: %w", err)
		}
	}

	return result, nil
}

// createDirectory creates a directory
func createDirectory(dirPath string, options MakeTreeOptions, result *MakeResult) error {
	// Check if directory already exists
	if _, err := os.Stat(dirPath); err == nil {
		if !options.Force {
			result.SkippedPaths = append(result.SkippedPaths, dirPath+" (already exists)")
			return nil
		}
	}

	if options.DryRun {
		result.CreatedDirs = append(result.CreatedDirs, dirPath+" (dry run)")
		return nil
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	result.CreatedDirs = append(result.CreatedDirs, dirPath)
	return nil
}

// createFile creates a file
func createFile(filePath string, options MakeTreeOptions, result *MakeResult) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		if !options.Force {
			result.SkippedPaths = append(result.SkippedPaths, filePath+" (already exists)")
			return nil
		}
	}

	if options.DryRun {
		result.CreatedFiles = append(result.CreatedFiles, filePath+" (dry run)")
		return nil
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create empty file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		// Ignore close errors in defer to avoid masking the main error
		_ = file.Close()
	}()

	result.CreatedFiles = append(result.CreatedFiles, filePath)
	return nil
}

// createMasterInfoFile creates a master .info file with all entries
func createMasterInfoFile(treeStructure *TreeStructure, options MakeTreeOptions, result *MakeResult) error {
	infoPath := filepath.Join(treeStructure.RootPath, ".info")

	// Check if .info file already exists
	if _, err := os.Stat(infoPath); err == nil {
		if !options.Force {
			result.SkippedPaths = append(result.SkippedPaths, infoPath+" (already exists)")
			return nil
		}
	}

	if options.DryRun {
		result.InfoFileCreated = true
		return nil
	}

	// Create .info file content
	file, err := os.Create(infoPath)
	if err != nil {
		return err
	}
	defer func() {
		// Ignore close errors in defer to avoid masking the main error
		_ = file.Close()
	}()

	writer := bufio.NewWriter(file)
	defer func() {
		// Ignore flush errors in defer to avoid masking the main error
		_ = writer.Flush()
	}()

	// Write header comment
	if _, err := writer.WriteString("# File structure created by treex make-tree\n\n"); err != nil {
		return err
	}

	// Group entries by parent directory and write them
	entriesByParent := make(map[string][]TreeEntry)
	for _, entry := range treeStructure.Entries {
		parent := filepath.Dir(entry.RelativePath)
		if parent == "." {
			parent = ""
		}
		entriesByParent[parent] = append(entriesByParent[parent], entry)
	}

	// Write root level entries first
	if rootEntries, exists := entriesByParent[""]; exists {
		for _, entry := range rootEntries {
			if err := writeInfoEntry(writer, entry); err != nil {
				return err
			}
		}
	}

	result.InfoFileCreated = true
	return nil
}

// writeInfoEntry writes a single entry to the .info file
func writeInfoEntry(writer *bufio.Writer, entry TreeEntry) error {
	// Write path (with trailing slash for directories)
	path := filepath.Base(entry.RelativePath)
	if entry.IsDir {
		path += "/"
	}

	if _, err := writer.WriteString(path + "\n"); err != nil {
		return err
	}

	// Write description if it exists
	if entry.Description != "" {
		if _, err := writer.WriteString(entry.Description + "\n"); err != nil {
			return err
		}
	}

	// Add blank line between entries
	if _, err := writer.WriteString("\n"); err != nil {
		return err
	}

	return nil
}
