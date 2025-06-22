package info

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// InitOptions contains configuration for initializing .info files
type InitOptions struct {
	Depth int
}

// UserInteraction interface allows the business logic to interact with users
// This enables testing by providing mock implementations
type UserInteraction interface {
	ConfirmOverwrite(targetPath string) (bool, error)
	ShowSuccess(targetPath string, depth int)
}

// DirectoryEntry represents a file or directory entry for init purposes
type DirectoryEntry struct {
	Name  string
	IsDir bool
}

// scanDirectory scans a directory and returns its direct children up to the specified depth
// This is a simplified version that doesn't need the full tree package functionality
func scanDirectory(dirPath string, maxDepth int) ([]DirectoryEntry, error) {
	if maxDepth <= 0 {
		return []DirectoryEntry{}, nil
	}
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}
	
	var result []DirectoryEntry
	
	// Filter and sort entries
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		// Skip hidden files and directories (starting with .)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}
	
	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(filteredEntries, func(i, j int) bool {
		if filteredEntries[i].IsDir() != filteredEntries[j].IsDir() {
			return filteredEntries[i].IsDir() // directories first
		}
		return filteredEntries[i].Name() < filteredEntries[j].Name()
	})
	
	// Convert to our DirectoryEntry format (only direct children)
	for _, entry := range filteredEntries {
		result = append(result, DirectoryEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}
	
	return result, nil
}

// InitializeInfoFile creates a .info file for the given directory
// This is the main business logic function that can be tested independently
func InitializeInfoFile(targetPath string, options InitOptions, userInteraction UserInteraction) error {
	// Ensure the target path exists and is a directory
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", targetPath)
	}
	
	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", targetPath)
	}
	
	// Scan the directory for entries
	entries, err := scanDirectory(absPath, options.Depth)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}
	
	// Check if .info file already exists
	infoPath := filepath.Join(absPath, ".info")
	if _, err := os.Stat(infoPath); err == nil {
		// File exists, ask for confirmation
		shouldOverwrite, err := userInteraction.ConfirmOverwrite(targetPath)
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}
		
		if !shouldOverwrite {
			return nil // User cancelled, not an error
		}
	}
	
	// Create the .info file
	file, err := os.Create(infoPath)
	if err != nil {
		return fmt.Errorf("failed to create .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()
	
	// Write entries to the file
	for _, entry := range entries {
		// Add the file/directory name with trailing slash for directories
		entryName := entry.Name
		if entry.IsDir {
			entryName += "/"
		}
		
		if _, err := fmt.Fprintf(file, "%s\n", entryName); err != nil {
			return fmt.Errorf("failed to write to .info file: %w", err)
		}
		
		// Add an empty description line
		if _, err := fmt.Fprintf(file, "\n\n"); err != nil {
			return fmt.Errorf("failed to write to .info file: %w", err)
		}
	}
	
	// Show success message
	userInteraction.ShowSuccess(targetPath, options.Depth)
	
	return nil
} 