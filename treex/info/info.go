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
	"io"
	"path/filepath"
	"strings"

	"github.com/jwaldrip/treex/treex/logging"
	"github.com/spf13/afero"
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
	for _, ann := range info.annotations {
		result = append(result, *ann)
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

// GetDistanceToPath calculates distance from this .info file to target path
// Returns 0 if same directory, 1 if child, -1 if parent/outside scope
func (info *InfoFile) GetDistanceToPath(targetPath string) int {
	infoDir := filepath.Dir(info.Path)

	// Handle relative target paths - they are relative to the info file's directory
	var targetDir string
	if filepath.IsAbs(targetPath) {
		targetDir = filepath.Dir(targetPath)
	} else {
		// For relative paths, join with info directory first
		fullTargetPath := filepath.Join(infoDir, targetPath)
		targetDir = filepath.Dir(fullTargetPath)
	}

	// Make paths clean for comparison
	infoDir = filepath.Clean(infoDir)
	targetDir = filepath.Clean(targetDir)

	if infoDir == targetDir {
		return 0 // Same directory
	}

	// Check if target is under info directory
	rel, err := filepath.Rel(infoDir, targetDir)
	if err != nil {
		return -1 // Can't calculate relationship
	}

	if strings.HasPrefix(rel, "..") {
		return -1 // Target is outside or above info directory
	}

	// Count directory levels
	if rel == "." {
		return 0
	}

	return strings.Count(rel, string(filepath.Separator)) + 1
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

// InfoFileLoader handles loading .info files into InfoFile structs.
// This is where disk I/O occurs - all other operations are in-memory.
type InfoFileLoader struct {
	fs InfoFileSystem
}

// NewInfoFileLoader creates a new loader
func NewInfoFileLoader(fs InfoFileSystem) *InfoFileLoader {
	return &InfoFileLoader{fs: fs}
}

// LoadInfoFile loads a single .info file
func (loader *InfoFileLoader) LoadInfoFile(path string) (*InfoFile, error) {
	reader, err := loader.fs.ReadInfoFile(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return NewInfoFile(path, string(data)), nil
}

// LoadInfoFiles loads multiple .info files from a directory tree
func (loader *InfoFileLoader) LoadInfoFiles(rootPath string) ([]*InfoFile, error) {
	infoFilePaths, err := loader.fs.FindInfoFiles(rootPath)
	if err != nil {
		return nil, err
	}

	var infoFiles []*InfoFile
	for _, path := range infoFilePaths {
		infoFile, err := loader.LoadInfoFile(path)
		if err != nil {
			logging.Warn().Str("file", path).Err(err).Msg("cannot load .info file")
			continue
		}
		infoFiles = append(infoFiles, infoFile)
	}

	return infoFiles, nil
}

// InfoFileWriter handles writing InfoFile structs back to disk.
// This is where disk I/O occurs - all other operations are in-memory.
type InfoFileWriter struct {
	fs InfoFileSystem
}

// NewInfoFileWriter creates a new writer
func NewInfoFileWriter(fs InfoFileSystem) *InfoFileWriter {
	return &InfoFileWriter{fs: fs}
}

// WriteInfoFile writes an InfoFile to disk
func (writer *InfoFileWriter) WriteInfoFile(infoFile *InfoFile) error {
	content := infoFile.String()
	return writer.fs.WriteInfoFile(infoFile.Path, content)
}

// WriteInfoFiles writes multiple InfoFiles to disk
// Files with IsEmpty() == true are removed from disk
func (writer *InfoFileWriter) WriteInfoFiles(infoFiles []*InfoFile) error {
	for _, infoFile := range infoFiles {
		if infoFile.IsEmpty() {
			// Remove empty files - this could be enhanced to actually delete
			// For now, write empty content
			if err := writer.fs.WriteInfoFile(infoFile.Path, ""); err != nil {
				return err
			}
		} else {
			if err := writer.WriteInfoFile(infoFile); err != nil {
				return err
			}
		}
	}
	return nil
}

// InfoFileCollection provides operations on collections of InfoFiles.
// All operations are in-memory and work with the parsed tree structures.
type InfoFileCollection struct {
	infoFiles []*InfoFile
}

// NewInfoFileCollection creates a new collection
func NewInfoFileCollection(infoFiles []*InfoFile) *InfoFileCollection {
	return &InfoFileCollection{infoFiles: infoFiles}
}

// Distribute redistributes annotations to their optimal .info files based on path proximity
// Returns a new collection with annotations moved to their closest .info files
func (collection *InfoFileCollection) Distribute() *InfoFileCollection {
	// Create map of all annotations with their current distances
	type annotationWithDistance struct {
		annotation *Annotation
		distance   int
		sourceFile *InfoFile
		targetFile *InfoFile
	}

	var annotationsToMove []annotationWithDistance

	// For each annotation, find the best .info file
	for _, infoFile := range collection.infoFiles {
		for _, annotation := range infoFile.GetAllAnnotations() {
			bestFile := infoFile
			bestDistance := infoFile.GetDistanceToPath(annotation.Path)

			// Find a better file if one exists
			for _, otherFile := range collection.infoFiles {
				distance := otherFile.GetDistanceToPath(annotation.Path)
				if distance >= 0 && (bestDistance < 0 || distance < bestDistance) {
					bestFile = otherFile
					bestDistance = distance
				}
			}

			// If we found a better file, mark for moving
			if bestFile != infoFile {
				annotationsToMove = append(annotationsToMove, annotationWithDistance{
					annotation: &annotation,
					distance:   bestDistance,
					sourceFile: infoFile,
					targetFile: bestFile,
				})
			}
		}
	}

	// Create new collection with moved annotations
	newFiles := make([]*InfoFile, len(collection.infoFiles))
	for i, infoFile := range collection.infoFiles {
		// Start with a copy of the original file structure
		newFiles[i] = &InfoFile{
			Path:        infoFile.Path,
			Lines:       make([]Line, len(infoFile.Lines)),
			annotations: make(map[string]*Annotation),
		}
		copy(newFiles[i].Lines, infoFile.Lines)

		// Copy annotations that aren't being moved
		for path, ann := range infoFile.annotations {
			shouldMove := false
			for _, move := range annotationsToMove {
				if move.annotation.Path == path && move.sourceFile == infoFile {
					shouldMove = true
					break
				}
			}
			if !shouldMove {
				newFiles[i].annotations[path] = ann
			} else {
				// Remove from lines by marking as removed
				newFiles[i].RemoveAnnotationForPath(path)
			}
		}
	}

	// Add moved annotations to their target files
	for _, move := range annotationsToMove {
		for _, newFile := range newFiles {
			if newFile.Path == move.targetFile.Path {
				newFile.AddAnnotationForPath(move.annotation.Path, move.annotation.Annotation)
				break
			}
		}
	}

	return NewInfoFileCollection(newFiles)
}

// Gather merges all annotations into a single .info file at the root
// Returns a new collection with a single .info file containing all annotations
func (collection *InfoFileCollection) Gather(rootPath string) *InfoFileCollection {
	rootInfoPath := filepath.Join(rootPath, ".info")

	// Create a new root info file
	rootInfo := &InfoFile{
		Path:        rootInfoPath,
		Lines:       []Line{},
		annotations: make(map[string]*Annotation),
	}

	// Collect all annotations from all files
	for _, infoFile := range collection.infoFiles {
		for _, annotation := range infoFile.GetAllAnnotations() {
			// Avoid duplicates (first one wins as per spec)
			if !rootInfo.HasAnnotationForPath(annotation.Path) {
				rootInfo.AddAnnotationForPath(annotation.Path, annotation.Annotation)
			}
		}
	}

	// Create empty versions of other info files for removal
	var newFiles []*InfoFile
	newFiles = append(newFiles, rootInfo)

	for _, infoFile := range collection.infoFiles {
		if infoFile.Path != rootInfoPath {
			emptyFile := &InfoFile{
				Path:        infoFile.Path,
				Lines:       []Line{},
				annotations: make(map[string]*Annotation),
			}
			newFiles = append(newFiles, emptyFile)
		}
	}

	return NewInfoFileCollection(newFiles)
}

// GetFilesToRemove returns files that are empty and should be removed
func (collection *InfoFileCollection) GetFilesToRemove() []string {
	var toRemove []string
	for _, infoFile := range collection.infoFiles {
		if infoFile.IsEmpty() {
			toRemove = append(toRemove, infoFile.Path)
		}
	}
	return toRemove
}

// GetFiles returns all InfoFiles in the collection
func (collection *InfoFileCollection) GetFiles() []*InfoFile {
	return collection.infoFiles
}

// Backward compatibility functions - delegate to new modules

// Collector manages the collection and merging of annotations.
// Deprecated: Use InfoAPI or Gatherer directly for new code.
type Collector struct {
	gatherer *Gatherer
}

// NewCollector creates a new annotation collector.
func NewCollector() *Collector {
	return &Collector{
		gatherer: NewGatherer(),
	}
}

// CollectAnnotations walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
func (c *Collector) CollectAnnotations(fsys afero.Fs, root string) (map[string]Annotation, error) {
	infoFS := NewAferoInfoFileSystem(fsys)
	return c.gatherer.GatherFromFileSystem(infoFS, root)
}

// OldValidator provides validation functionality for .info files
// Deprecated: Use InfoAPI.Validate() or Validator directly for new code.
type OldValidator struct {
	validator *Validator
}

// NewValidator creates a new info file validator
// Deprecated: Use NewInfoValidator() from validator.go or InfoAPI for new code.
func NewValidator(fs afero.Fs) *OldValidator {
	return &OldValidator{
		validator: NewInfoValidator(),
	}
}

// NewValidatorWithLogger creates a new info file validator with a custom logger
// Deprecated: Use NewInfoValidator() from validator.go or InfoAPI for new code.
// The logger parameter is ignored - logging now uses the global logging infrastructure.
func NewValidatorWithLogger(fs afero.Fs, logger interface{}) *OldValidator {
	return &OldValidator{
		validator: NewInfoValidator(),
	}
}

// ValidateDirectory validates all .info files in a directory tree
func (v *OldValidator) ValidateDirectory(rootPath string) (*ValidationResult, error) {
	// Note: The original OldValidator had an fs field, but for backward compatibility
	// we'll need the filesystem to be passed in or use the OS filesystem
	// For now, this is a placeholder that would need the filesystem passed through
	panic("OldValidator.ValidateDirectory needs filesystem - use InfoAPI instead")
}
