package info

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CollectOptions contains options for collecting .info files
type CollectOptions struct {
	// InfoFileName is the name of info files to look for (default: ".info")
	InfoFileName string
	// DryRun if true, doesn't modify any files
	DryRun bool
	// PreserveOrder if true, maintains the order of entries from each file
	PreserveOrder bool
}

// CollectResult contains the result of collecting .info files
type CollectResult struct {
	// RootPath is the directory where collection was performed
	RootPath string
	// CollectedFiles is the list of .info files that were collected
	CollectedFiles []string
	// TotalEntries is the total number of entries collected
	TotalEntries int
	// ConflictResolutions contains paths that had conflicts and which file won
	ConflictResolutions map[string]string
	// MergedContent is the content for the root .info file
	MergedContent string
	// Errors contains any non-fatal errors encountered
	Errors []error
}

// CollectInfoFiles collects all .info files from subdirectories into the root
func CollectInfoFiles(rootPath string, options CollectOptions) (*CollectResult, error) {
	// Set defaults
	if options.InfoFileName == "" {
		options.InfoFileName = DefaultInfoFileName
	}

	// Resolve absolute path
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	result := &CollectResult{
		RootPath:            absRoot,
		CollectedFiles:      []string{},
		ConflictResolutions: make(map[string]string),
		Errors:              []error{},
	}

	// Find all .info files in the tree
	infoFiles, err := findAllInfoFiles(absRoot, options.InfoFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to find info files: %w", err)
	}

	// Collect all entries from all files
	allEntries := make(map[string]*collectedEntry)
	
	for _, infoPath := range infoFiles {
		entries, err := collectEntriesFromFile(infoPath, absRoot)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error reading %s: %w", infoPath, err))
			continue
		}

		// Merge entries, resolving conflicts
		for path, entry := range entries {
			if existing, exists := allEntries[path]; exists {
				// Conflict: keep the entry from the file closer to the target
				if isCloserToPath(entry.sourceFile, existing.sourceFile, path, absRoot) {
					allEntries[path] = entry
					result.ConflictResolutions[path] = entry.sourceFile
				}
			} else {
				allEntries[path] = entry
			}
		}

		result.CollectedFiles = append(result.CollectedFiles, infoPath)
	}

	// Generate merged content
	result.MergedContent = generateMergedContent(allEntries, options.PreserveOrder)
	result.TotalEntries = len(allEntries)

	// If not dry run, write the file and delete children
	if !options.DryRun {
		rootInfoPath := filepath.Join(absRoot, options.InfoFileName)
		
		// Write merged content
		if err := os.WriteFile(rootInfoPath, []byte(result.MergedContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to write root info file: %w", err)
		}

		// Delete child .info files
		for _, infoPath := range result.CollectedFiles {
			// Don't delete the root .info file
			if infoPath == rootInfoPath {
				continue
			}
			
			if err := os.Remove(infoPath); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to remove %s: %w", infoPath, err))
			}
		}
	}

	return result, nil
}

// collectedEntry represents an entry collected from an info file
type collectedEntry struct {
	path        string
	annotation  string
	sourceFile  string
	originalLine string
}

// findAllInfoFiles recursively finds all info files in a directory tree
func findAllInfoFiles(rootPath string, infoFileName string) ([]string, error) {
	var infoFiles []string
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}
		
		if !info.IsDir() && info.Name() == infoFileName {
			infoFiles = append(infoFiles, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort for consistent ordering
	sort.Strings(infoFiles)
	return infoFiles, nil
}

// collectEntriesFromFile reads entries from a single info file
func collectEntriesFromFile(infoPath string, rootPath string) (map[string]*collectedEntry, error) {
	content, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]*collectedEntry)
	infoDir := filepath.Dir(infoPath)
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the entry - support both colon and whitespace format
		var entryPath, annotation string
		
		colonIdx := strings.Index(line, ":")
		if colonIdx != -1 {
			// Colon format: path:annotation
			entryPath = strings.TrimSpace(line[:colonIdx])
			annotation = strings.TrimSpace(line[colonIdx+1:])
		} else {
			// Whitespace format: path annotation
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			entryPath = fields[0]
			// Find where the path ends to preserve spacing in annotation
			pathEnd := strings.Index(line, entryPath) + len(entryPath)
			annotation = strings.TrimSpace(line[pathEnd:])
		}

		// Convert path to be relative to root
		var targetPath string
		if filepath.IsAbs(entryPath) {
			targetPath = entryPath
		} else {
			// Path is relative to the info file's directory
			targetPath = filepath.Join(infoDir, entryPath)
		}

		// Make path relative to root
		relPath, err := filepath.Rel(rootPath, targetPath)
		if err != nil {
			continue
		}

		// Normalize path separators
		relPath = filepath.ToSlash(relPath)
		
		// Preserve trailing slashes from original entry
		if strings.HasSuffix(entryPath, "/") && !strings.HasSuffix(relPath, "/") {
			relPath = relPath + "/"
		}
		
		entries[relPath] = &collectedEntry{
			path:         relPath,
			annotation:   annotation,
			sourceFile:   infoPath,
			originalLine: line,
		}
	}

	return entries, nil
}

// isCloserToPath determines if file1 is closer to targetPath than file2
func isCloserToPath(file1, file2, targetPath, rootPath string) bool {
	// Convert target path to absolute
	absTarget := filepath.Join(rootPath, targetPath)
	
	// Calculate distances
	dist1 := calculatePathDistance(file1, absTarget)
	dist2 := calculatePathDistance(file2, absTarget)
	
	// Closer means fewer directories between them
	return dist1 < dist2
}

// calculatePathDistance calculates the number of directories between two paths
func calculatePathDistance(from, to string) int {
	fromDir := filepath.Dir(from)
	toDir := filepath.Dir(to)
	
	// Find common ancestor
	fromParts := strings.Split(filepath.ToSlash(fromDir), "/")
	toParts := strings.Split(filepath.ToSlash(toDir), "/")
	
	commonLen := 0
	for i := 0; i < len(fromParts) && i < len(toParts); i++ {
		if fromParts[i] == toParts[i] {
			commonLen++
		} else {
			break
		}
	}
	
	// Distance is the sum of steps up from 'from' and down to 'to'
	distance := (len(fromParts) - commonLen) + (len(toParts) - commonLen)
	return distance
}

// generateMergedContent creates the content for the merged .info file
func generateMergedContent(entries map[string]*collectedEntry, preserveOrder bool) string {
	// Get all paths
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	
	// Sort paths for consistent output
	sort.Strings(paths)
	
	// Build content
	var lines []string
	for _, path := range paths {
		entry := entries[path]
		line := fmt.Sprintf("%s: %s", path, entry.annotation)
		lines = append(lines, line)
	}
	
	return strings.Join(lines, "\n") + "\n"
}