package info

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Annotation represents a single file/directory annotation
type Annotation struct {
	Path  string
	Notes string // Complete notes for the file/directory
	
	// Deprecated fields for backwards compatibility
	Title       string // Deprecated: Use Notes instead
	Description string // Deprecated: Use Notes instead
}

// Parser handles parsing .info files
type Parser struct {
	annotations map[string]*Annotation
}

// NewParser creates a new info file parser
func NewParser() *Parser {
	return &Parser{
		annotations: make(map[string]*Annotation),
	}
}

// ParseFile parses a .info file and returns a map of path -> annotation
func (p *Parser) ParseFile(infoFilePath string) (map[string]*Annotation, error) {
	file, err := os.Open(infoFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No .info file is not an error, just return empty map
			return make(map[string]*Annotation), nil
		}
		return nil, fmt.Errorf("failed to open .info file: %w", err)
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
		return nil, fmt.Errorf("error reading .info file: %w", err)
	}

	// Parse the lines - simple single-line format only
	for _, line := range lines {
		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Find the colon separator
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			// Old format (path notes) - backwards compatibility
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			
			path := parts[0]
			notes := strings.Join(parts[1:], " ")
			
			p.annotations[path] = &Annotation{
				Path:        path,
				Notes:       notes,
				Title:       notes,       // For backwards compatibility
				Description: notes,       // For backwards compatibility
			}
			continue
		}

		// New format (path:notes)
		path := strings.TrimSpace(line[:colonIdx])
		notes := strings.TrimSpace(line[colonIdx+1:])
		
		if path == "" || notes == "" {
			// Skip invalid entries
			continue
		}

		// Save this annotation
		p.annotations[path] = &Annotation{
			Path:        path,
			Notes:       notes,
			Title:       notes,       // For backwards compatibility
			Description: notes,       // For backwards compatibility
		}
	}

	return p.annotations, nil
}




// GetAnnotation returns the annotation for a given path
func (p *Parser) GetAnnotation(path string) (*Annotation, bool) {
	annotation, exists := p.annotations[path]
	return annotation, exists
}

// GetAllAnnotations returns all parsed annotations
func (p *Parser) GetAllAnnotations() map[string]*Annotation {
	return p.annotations
}

// ParseDirectory looks for a .info file in the given directory and parses it
func ParseDirectory(dirPath string) (map[string]*Annotation, error) {
	infoPath := filepath.Join(dirPath, ".info")
	parser := NewParser()
	return parser.ParseFile(infoPath)
}

// ParseDirectoryTree recursively looks for .info files in the entire directory tree
// and merges all annotations with proper path resolution
func ParseDirectoryTree(rootPath string) (map[string]*Annotation, error) {
	allAnnotations := make(map[string]*Annotation)

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
		annotations, err := parseFileWithContext(infoPath, rootPath, currentPath)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", infoPath, err)
		}

		// Merge annotations (later files override earlier ones if there are conflicts)
		for path, annotation := range annotations {
			allAnnotations[path] = annotation
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return allAnnotations, nil
}

// parseFileWithContext parses a .info file with proper path resolution
// rootPath: the root of the entire tree being analyzed
// contextDir: the directory containing this .info file
func parseFileWithContext(infoFilePath, rootPath, contextDir string) (map[string]*Annotation, error) {
	parser := NewParser()

	// Parse the file normally first
	annotations, err := parser.ParseFile(infoFilePath)
	if err != nil {
		return nil, err
	}

	// Now resolve paths relative to the context directory
	resolvedAnnotations := make(map[string]*Annotation)

	for localPath, annotation := range annotations {
		// Validate that the path doesn't try to escape the current directory
		if strings.Contains(localPath, "..") {
			continue // Skip paths that try to go up directories
		}

		// Create absolute path for this annotation
		fullPath := filepath.Join(contextDir, localPath)

		// Convert to path relative to root
		relativePath, err := filepath.Rel(rootPath, fullPath)
		if err != nil {
			continue // Skip if we can't resolve the path
		}

		// Normalize path separators for consistency
		relativePath = filepath.ToSlash(relativePath)

		// Create new annotation with resolved path
		resolvedAnnotation := &Annotation{
			Path:        relativePath,
			Notes:       annotation.Notes,
			Title:       annotation.Title,       // For backwards compatibility
			Description: annotation.Description, // For backwards compatibility
		}

		resolvedAnnotations[relativePath] = resolvedAnnotation
	}

	return resolvedAnnotations, nil
}