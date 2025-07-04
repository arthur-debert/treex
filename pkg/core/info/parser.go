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

		// This should be a path and title on the same line separated by whitespace
		pathLine := strings.TrimSpace(lines[i])
		i++

		// Split on whitespace to separate path and title
		parts := strings.Fields(pathLine)
		if len(parts) < 2 {
			// Not valid format - must have at least path and title on same line
			// Skip this line and continue
			continue
		}

		// First part is path, rest is title
		path := parts[0]
		titleFromPathLine := strings.Join(parts[1:], " ")

		// Collect description lines until we find a blank line followed by a non-empty line
		// that looks like it could be a new path+title line
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

				// We found a non-empty line. Now we need to decide if it's a new entry or part of description
				nextLine := lines[nextNonEmptyIdx]

				// Check if the next line looks like a new entry (must contain at least two words)
				if isLikelyNewEntry(nextLine) {
					// This looks like a new entry, stop collecting description here
					break
				} else {
					// This empty line is likely part of the description formatting
					if len(descriptionLines) == 0 {
						// No additional description - stop here
						break
					} else {
						// Include the empty line as part of description formatting
						descriptionLines = append(descriptionLines, line)
						i++
						continue
					}
				}
			} else {
				// Non-empty line, add to description
				descriptionLines = append(descriptionLines, line)
				i++
			}
		}

		// Save this annotation
		p.saveAnnotation(path, titleFromPathLine, descriptionLines)
	}

	return p.annotations, nil
}

// isLikelyNewEntry determines if a line looks like the start of a new annotation entry
// In compact format, a new entry must have at least two words (path and title)
func isLikelyNewEntry(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Check if this line contains multiple words (must be "path title" format)
	parts := strings.Fields(trimmed)
	return len(parts) >= 2
}

// saveAnnotation processes and saves an annotation with title from the path line
func (p *Parser) saveAnnotation(path string, title string, descriptionLines []string) {
	// Remove trailing empty lines from description
	for len(descriptionLines) > 0 && strings.TrimSpace(descriptionLines[len(descriptionLines)-1]) == "" {
		descriptionLines = descriptionLines[:len(descriptionLines)-1]
	}

	// Set up the annotation
	var fullDescription string

	if len(descriptionLines) > 0 {
		// Additional description lines after the path+title line
		fullDescription = strings.Join(descriptionLines, "\n")
	}

	// For multi-line descriptions, we need to ensure the title is included in the full description
	// if the title isn't already part of the description
	if len(descriptionLines) == 0 || (len(descriptionLines) > 0 && !strings.HasPrefix(fullDescription, title)) {
		if fullDescription == "" {
			fullDescription = title
		} else {
			// Prepend the title to the description for proper multi-line support
			fullDescription = title + "\n" + fullDescription
		}
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