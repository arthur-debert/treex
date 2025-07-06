package maketree

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MakeTreeOptions contains configuration options for creating file trees
type MakeTreeOptions struct {
	Force      bool   // Overwrite existing files/directories
	DryRun     bool   // Don't actually create files, just show what would be created
	CreateInfo bool   // Create a master .info file (default: true)
	InfoHeader string // Optional header text to add to the .info file
}

// MakeResult contains information about what was created
type MakeResult struct {
	CreatedDirs     []string
	CreatedFiles    []string
	SkippedPaths    []string
	InfoFileCreated bool
	DryRun          bool
}

// Removed InputSource enum - only .info format is supported

// TreeStructure represents a parsed tree structure
type TreeStructure struct {
	RootPath string
	Entries  []TreeEntry
}

// TreeEntry represents a single entry in the tree structure
type TreeEntry struct {
	Path         string
	Description  string
	IsDir        bool
	RelativePath string // Path relative to root
}

// MakeTreeFromFile creates a file tree structure from a .info file
func MakeTreeFromFile(inputFile string, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	// Read and parse the .info file
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	return makeTreeFromInfoContent(string(content), rootDir, options)
}

// MakeTreeFromReader creates a file tree structure from .info format content from a reader
func MakeTreeFromReader(reader io.Reader, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	// Read all content from reader
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	return makeTreeFromInfoContent(string(content), rootDir, options)
}

// MakeTreeFromText creates a file tree structure from .info format text content
func MakeTreeFromText(content string, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	return makeTreeFromInfoContent(content, rootDir, options)
}

// makeTreeFromInfoContent processes .info format content and creates the tree structure
func makeTreeFromInfoContent(content string, rootDir string, options MakeTreeOptions) (*MakeResult, error) {
	entries, err := parseInfoContent(content, rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse info content: %w", err)
	}

	treeStructure := &TreeStructure{
		RootPath: rootDir,
		Entries:  entries,
	}

	return makeTreeStructure(treeStructure, options)
}

// parseInfoContent parses .info format content (path: description)
func parseInfoContent(content string, rootDir string) ([]TreeEntry, error) {
	lines := strings.Split(content, "\n")
	pathEntries := make(map[string]*TreeEntry)
	var hasValidEntry bool

	// First pass: parse all entries
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Find the colon separator
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			// Skip lines without colon separator
			continue
		}

		// Parse format (path: description)
		path := strings.TrimSpace(line[:colonIdx])
		description := strings.TrimSpace(line[colonIdx+1:])

		if path == "" {
			// Skip empty paths
			continue
		}

		hasValidEntry = true

		// Check if path explicitly ends with / (indicating directory)
		isDir := strings.HasSuffix(path, "/")
		path = strings.TrimSuffix(path, "/")

		// Store the entry
		pathEntries[path] = &TreeEntry{
			Path:         filepath.Join(rootDir, path),
			Description:  description,
			IsDir:        isDir,
			RelativePath: path,
		}
	}

	if !hasValidEntry {
		return nil, fmt.Errorf("no valid entries found (expected format: 'path: description')")
	}

	// Second pass: determine directories by path prefix
	// If a path is a prefix of another path, it must be a directory
	for path1, entry1 := range pathEntries {
		if entry1.IsDir {
			continue // Already marked as directory
		}

		for path2 := range pathEntries {
			if path1 != path2 && strings.HasPrefix(path2, path1+"/") {
				// path1 is a prefix of path2, so path1 must be a directory
				entry1.IsDir = true
				break
			}
		}
	}

	// Convert map to slice
	var entries []TreeEntry
	for _, entry := range pathEntries {
		entries = append(entries, *entry)
	}

	return entries, nil
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

	// Write custom header if provided
	if options.InfoHeader != "" {
		if _, err := writer.WriteString(options.InfoHeader + "\n\n"); err != nil {
			return err
		}
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

	// Write in format: path: description
	line := path
	if entry.Description != "" {
		line += ": " + entry.Description
	} else {
		line += ":"
	}

	if _, err := writer.WriteString(line + "\n"); err != nil {
		return err
	}

	return nil
}
