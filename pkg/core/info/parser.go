package info

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/core/types"
)


// Parser handles parsing .info files
type Parser struct {
	annotations map[string]*types.Annotation
}

// NewParser creates a new info file parser
func NewParser() *Parser {
	return &Parser{
		annotations: make(map[string]*types.Annotation),
	}
}

// ParseFile parses a .info file and returns a map of path -> annotation
func (p *Parser) ParseFile(infoFilePath string) (map[string]*types.Annotation, error) {
	annotations, _, err := p.ParseFileWithWarnings(infoFilePath)
	return annotations, err
}

// ParseFileWithWarnings parses a .info file and returns annotations plus any warnings
func (p *Parser) ParseFileWithWarnings(infoFilePath string) (map[string]*types.Annotation, []string, error) {
	var warnings []string
	
	file, err := os.Open(infoFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No .info file is not an error, just return empty map
			return make(map[string]*types.Annotation), nil, nil
		}
		return nil, nil, fmt.Errorf("failed to open .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Read all lines first
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading .info file: %w", err)
	}

	// Parse the lines - simple single-line format only
	for lineNum, line := range lines {
		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Find the colon separator
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			// Add warning for lines without colon separator
			warnings = append(warnings, fmt.Sprintf("Line %d: Invalid format (missing colon): %q", lineNum+1, line))
			continue
		}

		// Parse format (path:notes)
		path := strings.TrimSpace(line[:colonIdx])
		notes := strings.TrimSpace(line[colonIdx+1:])
		
		if path == "" {
			// Warn about empty path
			warnings = append(warnings, fmt.Sprintf("Line %d: Empty path in annotation", lineNum+1))
			continue
		}
		
		if notes == "" {
			// Warn about empty notes
			warnings = append(warnings, fmt.Sprintf("Line %d: Empty notes for path %q", lineNum+1, path))
			continue
		}

		// Save this annotation
		p.annotations[path] = &types.Annotation{
			Path:  path,
			Notes: notes,
		}
	}

	return p.annotations, warnings, nil
}




// GetAnnotation returns the annotation for a given path
func (p *Parser) GetAnnotation(path string) (*types.Annotation, bool) {
	annotation, exists := p.annotations[path]
	return annotation, exists
}

// GetAllAnnotations returns all parsed annotations
func (p *Parser) GetAllAnnotations() map[string]*types.Annotation {
	return p.annotations
}

// ParseDirectory looks for a .info file in the given directory and parses it
func ParseDirectory(dirPath string) (map[string]*types.Annotation, error) {
	infoPath := filepath.Join(dirPath, ".info")
	parser := NewParser()
	return parser.ParseFile(infoPath)
}

// ParseDirectoryTree recursively looks for .info files in the entire directory tree
// and merges all annotations with proper path resolution.
// 
// When a file is annotated in multiple .info files (e.g., in both a parent directory
// and a subdirectory), the annotation from the deeper/more specific .info file takes
// precedence. This is achieved through filepath.Walk's lexical ordering, which processes
// parent directories before their subdirectories, allowing later annotations to override
// earlier ones.
func ParseDirectoryTree(rootPath string) (map[string]*types.Annotation, error) {
	annotations, _, err := ParseDirectoryTreeWithWarnings(rootPath)
	return annotations, err
}

// ParseDirectoryTreeWithWarnings recursively looks for .info files and collects warnings
func ParseDirectoryTreeWithWarnings(rootPath string) (map[string]*types.Annotation, []string, error) {
	allAnnotations := make(map[string]*types.Annotation)
	var allWarnings []string

	// Walk the directory tree
	err := filepath.Walk(rootPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't read instead of failing completely
			return nil
		}

		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}

		// Look for .info file in this directory
		infoPath := filepath.Join(currentPath, ".info")
		if _, err := os.Stat(infoPath); os.IsNotExist(err) {
			// No .info file in this directory, continue
			return nil
		}

		// Parse the .info file with proper context
		annotations, warnings, err := parseFileWithContextAndWarnings(infoPath, rootPath, currentPath)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", infoPath, err)
		}
		
		// Collect warnings with file context
		for _, warning := range warnings {
			relPath, _ := filepath.Rel(rootPath, infoPath)
			allWarnings = append(allWarnings, fmt.Sprintf("%s: %s", relPath, warning))
		}

		// Merge annotations (later files override earlier ones if there are conflicts)
		for path, annotation := range annotations {
			allAnnotations[path] = annotation
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}
	
	// Check for non-existent paths
	for annotationPath := range allAnnotations {
		// Convert relative path to absolute path for checking
		fullPath := filepath.Join(rootPath, annotationPath)
		
		// Normalize path separators
		fullPath = filepath.Clean(fullPath)
		
		// Check if the path exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			allWarnings = append(allWarnings, 
				fmt.Sprintf("Path not found: %q", annotationPath))
		}
	}

	return allAnnotations, allWarnings, nil
}

// parseFileWithContext parses a .info file with proper path resolution
// rootPath: the root of the entire tree being analyzed
// contextDir: the directory containing this .info file
func parseFileWithContext(infoFilePath, rootPath, contextDir string) (map[string]*types.Annotation, error) {
	annotations, _, err := parseFileWithContextAndWarnings(infoFilePath, rootPath, contextDir)
	return annotations, err
}

// parseFileWithContextAndWarnings parses a .info file with proper path resolution and collects warnings
func parseFileWithContextAndWarnings(infoFilePath, rootPath, contextDir string) (map[string]*types.Annotation, []string, error) {
	parser := NewParser()

	// Parse the file with warnings
	annotations, warnings, err := parser.ParseFileWithWarnings(infoFilePath)
	if err != nil {
		return nil, nil, err
	}

	// Now resolve paths relative to the context directory
	resolvedAnnotations := make(map[string]*types.Annotation)
	var contextWarnings []string

	for localPath, annotation := range annotations {
		// Validate that the path doesn't try to escape the current directory
		if strings.Contains(localPath, "..") {
			contextWarnings = append(contextWarnings, fmt.Sprintf("Path tries to escape directory: %q", localPath))
			continue // Skip paths that try to go up directories
		}

		// Create absolute path for this annotation
		fullPath := filepath.Join(contextDir, localPath)

		// Convert to path relative to root
		relativePath, err := filepath.Rel(rootPath, fullPath)
		if err != nil {
			contextWarnings = append(contextWarnings, fmt.Sprintf("Cannot resolve path %q: %v", localPath, err))
			continue // Skip if we can't resolve the path
		}

		// Normalize path separators for consistency
		relativePath = filepath.ToSlash(relativePath)

		// Create new annotation with resolved path
		resolvedAnnotation := &types.Annotation{
			Path:  relativePath,
			Notes: annotation.Notes,
		}

		resolvedAnnotations[relativePath] = resolvedAnnotation
	}

	// Combine warnings
	allWarnings := append(warnings, contextWarnings...)

	return resolvedAnnotations, allWarnings, nil
}