package info

import (
	"path/filepath"
	"sort"
	"strings"
)

// Editor provides pure content editing functions for .info files
type Editor struct{}

// NewEditor creates a new editor instance
func NewEditor() *Editor {
	return &Editor{}
}

// AddAnnotation adds a new annotation to .info file content
// Returns the updated content with the new annotation appended
func (e *Editor) AddAnnotation(content, targetPath, annotation, infoFilePath string) string {
	// Convert target path to relative path for the .info file
	relativePath := e.makeRelativePath(targetPath, infoFilePath)

	// Escape spaces in path
	escapedPath := e.escapePath(relativePath)

	// Create new line
	newLine := escapedPath + "  " + annotation

	// Add to existing content
	if content == "" {
		return newLine + "\n"
	}

	// Ensure content ends with newline before appending
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	return content + newLine + "\n"
}

// RemoveAnnotation removes an annotation from .info file content
// Returns the updated content with the specified line removed
func (e *Editor) RemoveAnnotation(content string, lineNum int) string {
	lines := strings.Split(content, "\n")

	// Convert to 0-based indexing
	index := lineNum - 1

	// Check bounds
	if index < 0 || index >= len(lines) {
		return content
	}

	// Remove the line
	newLines := append(lines[:index], lines[index+1:]...)

	// If we removed the last non-empty line and it results in just empty content, return empty
	result := strings.Join(newLines, "\n")
	if strings.TrimSpace(result) == "" {
		return ""
	}

	return result
}

// UpdateAnnotation updates an existing annotation in .info file content
// Returns the updated content with the annotation on the specified line changed
func (e *Editor) UpdateAnnotation(content string, lineNum int, targetPath, newAnnotation, infoFilePath string) string {
	lines := strings.Split(content, "\n")

	// Convert to 0-based indexing
	index := lineNum - 1

	// Check bounds
	if index < 0 || index >= len(lines) {
		return content
	}

	// Convert target path to relative path for the .info file
	relativePath := e.makeRelativePath(targetPath, infoFilePath)

	// Escape spaces in path
	escapedPath := e.escapePath(relativePath)

	// Create new line
	newLine := escapedPath + "  " + newAnnotation

	// Replace the line
	lines[index] = newLine

	return strings.Join(lines, "\n")
}

// GenerateContent generates complete .info file content from annotations
// Returns sorted content with all annotations
func (e *Editor) GenerateContent(annotations []Annotation, infoFilePath string) string {
	if len(annotations) == 0 {
		return ""
	}

	// Sort annotations by path for consistent output
	sorted := make([]Annotation, len(annotations))
	copy(sorted, annotations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	var lines []string
	for _, ann := range sorted {
		// Convert to relative path for this .info file
		relativePath := e.makeRelativePath(ann.Path, infoFilePath)

		// Escape spaces in path
		escapedPath := e.escapePath(relativePath)

		// Create line
		line := escapedPath + "  " + ann.Annotation
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}

// makeRelativePath converts a target path to a relative path from the .info file's directory
func (e *Editor) makeRelativePath(targetPath, infoFilePath string) string {
	infoDir := filepath.Dir(infoFilePath)

	// If target path is already relative to the info file directory, use it directly
	if !filepath.IsAbs(targetPath) {
		// Calculate relative path from info dir
		rel, err := filepath.Rel(infoDir, targetPath)
		if err != nil {
			// Fallback to original path
			return targetPath
		}

		// If the relative path goes up and then comes back to the same level,
		// we might want to use the basename
		if strings.HasPrefix(rel, "../") {
			// For paths like "sub/target.txt" with info file "sub/.info",
			// we want just "target.txt"
			if filepath.Dir(targetPath) == infoDir {
				return filepath.Base(targetPath)
			}
		}

		return rel
	}

	// For absolute paths, make them relative to info directory
	rel, err := filepath.Rel(infoDir, targetPath)
	if err != nil {
		return filepath.Base(targetPath)
	}

	return rel
}

// escapePath escapes spaces in file paths for .info file format
func (e *Editor) escapePath(path string) string {
	return strings.ReplaceAll(path, " ", "\\ ")
}
