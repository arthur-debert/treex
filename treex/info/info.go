// Package info provides in-memory processing of .info annotation files.
//
// # Architecture Overview
//
// This package implements a high-performance, in-memory approach to processing .info files
// that eliminates the string conversion bottlenecks found in traditional line-by-line parsing.
//
// # File Boundaries and Disk Access
//
// The design strictly separates in-memory operations from disk I/O:
//
//   - InfoFileLoader: Handles reading from disk (at package boundaries)
//   - InfoFile: Pure in-memory operations with structured data
//   - InfoFileWriter: Handles writing to disk (at package boundaries)
//   - InfoFileCollection: In-memory operations on collections of InfoFiles
//
// The only time InfoFile touches disk is during path validation using the injectable
// PathExists function. All other operations work purely on the parsed tree structure.
//
// # Performance Benefits
//
// Traditional approach: File → []byte → string → []string (lines) → []Annotation
// New approach: File → InfoFile struct → direct operations on parsed tree
//
// This eliminates:
//   - Repeated string conversions (byte[] ↔ string)
//   - Line-by-line string splitting and trimming
//   - Redundant path cleaning operations
//   - Multiple parsing passes over the same content
//
// # Usage Patterns
//
//	// Load files from disk once
//	loader := NewInfoFileLoader(fs)
//	infoFiles, _ := loader.LoadInfoFiles(rootPath)
//
//	// Work with parsed structures (no more string operations)
//	collection := NewInfoFileCollection(infoFiles)
//	distributed := collection.Distribute()  // Smart redistribution
//	gathered := collection.Gather(rootPath) // Merge to root
//
//	// Write back to disk once
//	writer := NewInfoFileWriter(fs)
//	writer.WriteInfoFiles(distributed.GetFiles())
//
// # Round-trip Fidelity
//
// InfoFile preserves:
//   - Comment lines (# prefix)
//   - Blank lines
//   - Original formatting and spacing
//   - Malformed lines (for error reporting)
//
// This ensures that files can be read, modified, and written back with minimal changes
// to non-semantic content.
package info

import (
	"strings"
)

// Annotation represents a single entry in an .info file.
type Annotation struct {
	Path       string
	Annotation string
	InfoFile   string // The path to the .info file this annotation came from.
	LineNum    int
}

// Line represents a single line in an .info file (parsed or raw)
type Line struct {
	LineNum    int
	Raw        string // Original line content for preserving formatting
	Type       LineType
	Annotation *Annotation // nil for non-annotation lines
	ParseError string      // error message if parsing failed
}

type LineType int

const (
	LineTypeComment LineType = iota
	LineTypeBlank
	LineTypeAnnotation
	LineTypeMalformed
)

// InfoFile represents a parsed .info file with methods for manipulation.
// All operations are in-memory only - disk access occurs only at package boundaries
// via InfoFileLoader and InfoFileWriter.
type InfoFile struct {
	Path        string
	Lines       []Line
	annotations map[string]*Annotation // path -> annotation for fast lookup
}

// NewInfoFile creates a new InfoFile from content
func NewInfoFile(path, content string) *InfoFile {
	parser := NewParser()
	lines := strings.Split(content, "\n")

	infoFile := &InfoFile{
		Path:        path,
		Lines:       make([]Line, 0, len(lines)),
		annotations: make(map[string]*Annotation),
	}

	parsedPaths := make(map[string]bool)

	for lineNum, rawLine := range lines {
		lineNum++ // Convert to 1-based
		trimmed := strings.TrimSpace(rawLine)

		line := Line{
			LineNum: lineNum,
			Raw:     rawLine,
		}

		if trimmed == "" {
			line.Type = LineTypeBlank
		} else if strings.HasPrefix(trimmed, "#") {
			line.Type = LineTypeComment
		} else {
			// Try to parse as annotation
			path, annotation, ok := parser.ParseLine(trimmed)
			if !ok {
				line.Type = LineTypeMalformed
				line.ParseError = "no annotation found (missing space separator)"
			} else if parsedPaths[path] {
				line.Type = LineTypeMalformed
				line.ParseError = "duplicate path (first occurrence wins)"
			} else {
				line.Type = LineTypeAnnotation
				line.Annotation = &Annotation{
					Path:       path,
					Annotation: annotation,
					InfoFile:   infoFile.Path,
					LineNum:    lineNum,
				}
				infoFile.annotations[path] = line.Annotation
				parsedPaths[path] = true
			}
		}

		infoFile.Lines = append(infoFile.Lines, line)
	}

	return infoFile
}

// HasAnnotationForPath checks if an annotation exists for the given path
func (info *InfoFile) HasAnnotationForPath(path string) bool {
	_, exists := info.annotations[path]
	return exists
}

// GetAnnotationForPath returns the annotation for a path, or nil if not found
func (info *InfoFile) GetAnnotationForPath(path string) *Annotation {
	return info.annotations[path]
}

// GetAllAnnotations returns all valid annotations in this file
func (info *InfoFile) GetAllAnnotations() []Annotation {
	result := make([]Annotation, 0, len(info.annotations))

	// Collect annotations from Lines in order (to maintain original file order)
	for _, line := range info.Lines {
		if line.Type == LineTypeAnnotation && line.Annotation != nil {
			result = append(result, *line.Annotation)
		}
	}

	return result
}

// UpdateAnnotationForPath updates an existing annotation
func (info *InfoFile) UpdateAnnotationForPath(path, newAnnotation string) bool {
	if ann := info.annotations[path]; ann != nil {
		ann.Annotation = newAnnotation
		// Update the line as well
		for i := range info.Lines {
			if info.Lines[i].Annotation == ann {
				info.Lines[i].Raw = path + " " + newAnnotation
				break
			}
		}
		return true
	}
	return false
}

// RemoveAnnotationForPath removes an annotation for a path
func (info *InfoFile) RemoveAnnotationForPath(path string) bool {
	if ann := info.annotations[path]; ann != nil {
		delete(info.annotations, path)
		// Mark the line as removed (could be filtered out during serialization)
		for i := range info.Lines {
			if info.Lines[i].Annotation == ann {
				info.Lines[i].Type = LineTypeMalformed
				info.Lines[i].ParseError = "removed"
				info.Lines[i].Annotation = nil
				break
			}
		}
		return true
	}
	return false
}

// AddAnnotationForPath adds a new annotation (appends to end)
func (info *InfoFile) AddAnnotationForPath(path, annotation string) bool {
	if info.HasAnnotationForPath(path) {
		return false // Already exists
	}

	newAnnotation := &Annotation{
		Path:       path,
		Annotation: annotation,
		InfoFile:   info.Path,
		LineNum:    len(info.Lines) + 1,
	}

	newLine := Line{
		LineNum:    len(info.Lines) + 1,
		Raw:        path + " " + annotation,
		Type:       LineTypeAnnotation,
		Annotation: newAnnotation,
	}

	info.Lines = append(info.Lines, newLine)
	info.annotations[path] = newAnnotation
	return true
}

// IsEmpty returns true if the file has no valid annotations
func (info *InfoFile) IsEmpty() bool {
	return len(info.annotations) == 0
}

// String serializes the InfoFile back to .info format
func (info *InfoFile) String() string {
	var result []string
	for _, line := range info.Lines {
		if line.Type == LineTypeMalformed && line.ParseError == "removed" {
			continue // Skip removed lines
		}
		result = append(result, line.Raw)
	}
	return strings.Join(result, "\n")
}
