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
	Path        string
	Title       string // First line of description (if it ends with newline)
	Description string // Full description
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
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	
	// Read all lines first
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .info file: %w", err)
	}

	// Parse the lines
	i := 0
	for i < len(lines) {
		// Skip empty lines at the beginning
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
		
		if i >= len(lines) {
			break
		}
		
		// This should be a path
		path := strings.TrimSpace(lines[i])
		i++
		
		// Collect description lines until we find a blank line followed by a non-empty line
		// that looks like it could be a new path
		var descriptionLines []string
		
		for i < len(lines) {
			line := lines[i]
			
			// If this is an empty line, we need to look ahead more carefully
			if strings.TrimSpace(line) == "" {
				// Look ahead to find the next non-empty line
				nextNonEmptyIdx := i + 1
				for nextNonEmptyIdx < len(lines) && strings.TrimSpace(lines[nextNonEmptyIdx]) == "" {
					nextNonEmptyIdx++
				}
				
				if nextNonEmptyIdx >= len(lines) {
					// No more non-empty lines, include this empty line and we're done
					descriptionLines = append(descriptionLines, line)
					i++
					break
				}
				
				// We found a non-empty line. Now we need to decide if it's a new path or part of description
				// If we haven't collected any description yet, this empty line is likely formatting
				// If we have description, this might be the separator
				if len(descriptionLines) == 0 {
					// This is likely just formatting after the path, skip it and continue
					i++
					continue
				} else {
					// We have some description already. This empty line + next non-empty line
					// likely indicates a new annotation
					break
				}
			} else {
				// Non-empty line, add to description
				descriptionLines = append(descriptionLines, line)
				i++
			}
		}
		
		// Save this annotation
		p.saveAnnotation(path, descriptionLines)
	}

	return p.annotations, nil
}

// saveAnnotation processes and saves an annotation
func (p *Parser) saveAnnotation(path string, descriptionLines []string) {
	if len(descriptionLines) == 0 {
		return
	}

	// Remove trailing empty lines
	for len(descriptionLines) > 0 && strings.TrimSpace(descriptionLines[len(descriptionLines)-1]) == "" {
		descriptionLines = descriptionLines[:len(descriptionLines)-1]
	}

	if len(descriptionLines) == 0 {
		return
	}

	// Join all lines for full description
	fullDescription := strings.Join(descriptionLines, "\n")
	
	// Determine title - first line if it's followed by more content
	var title string
	if len(descriptionLines) > 1 {
		title = strings.TrimSpace(descriptionLines[0])
	}

	annotation := &Annotation{
		Path:        path,
		Title:       title,
		Description: fullDescription,
	}

	p.annotations[path] = annotation
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
			Title:       annotation.Title,
			Description: annotation.Description,
		}
		
		resolvedAnnotations[relativePath] = resolvedAnnotation
	}

	return resolvedAnnotations, nil
} 