package info

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// InfoAPI provides the main interface for info file operations
type InfoAPI struct {
	fs        InfoFileSystem
	gatherer  *Gatherer
	validator *Validator
	editor    *Editor
	loader    *InfoFileLoader
	writer    *InfoFileWriter
}

// NewInfoAPI creates a new info API instance using afero filesystem
func NewInfoAPI(fs afero.Fs) *InfoAPI {
	afs := NewAferoInfoFileSystem(fs)
	return &InfoAPI{
		fs:        afs,
		gatherer:  NewGatherer(),
		validator: NewInfoValidator(),
		editor:    NewEditor(),
		loader:    NewInfoFileLoader(afs),
		writer:    NewInfoFileWriter(afs),
	}
}

// NewInfoAPIWithFileSystem creates a new info API instance with custom filesystem
func NewInfoAPIWithFileSystem(fs InfoFileSystem) *InfoAPI {
	return &InfoAPI{
		fs:        fs,
		gatherer:  NewGatherer(),
		validator: NewInfoValidator(),
		editor:    NewEditor(),
		loader:    NewInfoFileLoader(fs),
		writer:    NewInfoFileWriter(fs),
	}
}

// Gather collects and merges all annotations from .info files in a directory tree
// Uses the new InfoFile-based approach for better performance
func (api *InfoAPI) Gather(rootPath string) (map[string]Annotation, error) {
	infoFiles, err := api.loader.LoadInfoFiles(rootPath)
	if err != nil {
		return nil, err
	}

	return api.gatherer.GatherFromInfoFiles(infoFiles, api.fs.PathExists), nil
}

// GatherLegacy collects and merges all annotations using the old string-based approach
// DEPRECATED: Use Gather() for better performance
func (api *InfoAPI) GatherLegacy(rootPath string) (map[string]Annotation, error) {
	return api.gatherer.GatherFromFileSystem(api.fs, rootPath)
}

// Validate validates all .info files in a directory tree
func (api *InfoAPI) Validate(rootPath string) (*ValidationResult, error) {
	return api.validator.ValidateFileSystem(api.fs, rootPath)
}

// Add adds a new annotation to the appropriate .info file
func (api *InfoAPI) Add(targetPath, annotation string) error {
	// Determine the appropriate .info file for this target path
	infoFilePath := api.determineInfoFile(targetPath)

	// Read existing content
	content := api.readFileContent(infoFilePath)

	// Use editor to handle business logic
	newContent := api.editor.AddAnnotationToFile(targetPath, annotation, infoFilePath, content)

	// Write the new content
	return api.fs.WriteInfoFile(infoFilePath, newContent)
}

// Remove removes an annotation for a specific path
func (api *InfoAPI) Remove(targetPath string) error {
	// Gather all annotations to find the target
	annotations, err := api.Gather(".")
	if err != nil {
		return err
	}

	// Check if annotation exists
	targetAnnotation, exists := annotations[targetPath]
	if !exists {
		return fmt.Errorf("no annotation found for path %q", targetPath)
	}

	// Read the .info file content
	content := api.readFileContent(targetAnnotation.InfoFile)

	// Use editor to handle business logic
	newContent, found := api.editor.RemoveAnnotationFromContent(targetPath, content, annotations)
	if !found {
		return fmt.Errorf("annotation not found in content for path %q", targetPath)
	}

	// If content would be empty, write empty file
	if strings.TrimSpace(newContent) == "" {
		newContent = ""
	}

	// Write the new content
	return api.fs.WriteInfoFile(targetAnnotation.InfoFile, newContent)
}

// Update updates an existing annotation
func (api *InfoAPI) Update(targetPath, newAnnotation string) error {
	// Gather all annotations to find the target
	annotations, err := api.Gather(".")
	if err != nil {
		return err
	}

	// Check if annotation exists
	targetAnnotation, exists := annotations[targetPath]
	if !exists {
		return fmt.Errorf("no annotation found for path %q", targetPath)
	}

	// Read the .info file content
	content := api.readFileContent(targetAnnotation.InfoFile)

	// Use editor to handle business logic
	newContent, found := api.editor.UpdateAnnotationInContent(targetPath, newAnnotation, content, annotations)
	if !found {
		return fmt.Errorf("annotation not found in content for path %q", targetPath)
	}

	// Write the new content
	return api.fs.WriteInfoFile(targetAnnotation.InfoFile, newContent)
}

// List lists all current annotations in a directory tree
func (api *InfoAPI) List(rootPath string) ([]Annotation, error) {
	annotations, err := api.Gather(rootPath)
	if err != nil {
		return nil, err
	}

	var result []Annotation
	for _, ann := range annotations {
		result = append(result, ann)
	}

	return result, nil
}

// GetAnnotation retrieves the effective annotation for a specific path
func (api *InfoAPI) GetAnnotation(targetPath string) (*Annotation, error) {
	annotations, err := api.Gather(".")
	if err != nil {
		return nil, err
	}

	if ann, exists := annotations[targetPath]; exists {
		return &ann, nil
	}

	return nil, fmt.Errorf("no annotation found for path %q", targetPath)
}

// Clean removes invalid or redundant annotations
func (api *InfoAPI) Clean(rootPath string) (*CleanResult, error) {
	// First validate to find issues
	validationResult, err := api.Validate(rootPath)
	if err != nil {
		return nil, err
	}

	result := &CleanResult{
		RemovedAnnotations: make([]Annotation, 0),
		UpdatedFiles:       make([]string, 0),
		Summary:            CleanSummary{},
	}

	// Group issues by file
	fileIssues := make(map[string][]ValidationIssue)
	for _, issue := range validationResult.Issues {
		fileIssues[issue.InfoFile] = append(fileIssues[issue.InfoFile], issue)
	}

	// Process each file with issues
	for infoFile, issues := range fileIssues {
		if api.cleanFile(infoFile, issues, result) {
			result.UpdatedFiles = append(result.UpdatedFiles, infoFile)
			result.Summary.FilesModified++
		}
	}

	return result, nil
}

// cleanFile removes problematic annotations from a single .info file
func (api *InfoAPI) cleanFile(infoFilePath string, issues []ValidationIssue, result *CleanResult) bool {
	reader, err := api.fs.ReadInfoFile(infoFilePath)
	if err != nil {
		return false
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return false
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Track which lines to remove
	linesToRemove := make(map[int]bool)
	for _, issue := range issues {
		switch issue.Type {
		case IssuePathNotExists, IssueInvalidFormat, IssueDuplicatePath, IssueAncestorPath:
			linesToRemove[issue.LineNum] = true

			// Track what we're removing
			switch issue.Type {
			case IssuePathNotExists:
				result.Summary.InvalidPathsRemoved++
			case IssueDuplicatePath:
				result.Summary.DuplicatesRemoved++
			}

			// Create annotation for removed item
			result.RemovedAnnotations = append(result.RemovedAnnotations, Annotation{
				Path:       issue.Path,
				InfoFile:   issue.InfoFile,
				LineNum:    issue.LineNum,
				Annotation: fmt.Sprintf("(removed: %s)", issue.Type),
			})
		}
	}

	if len(linesToRemove) == 0 {
		return false
	}

	// Build new content without problematic lines
	var newLines []string
	for i, line := range lines {
		if !linesToRemove[i+1] { // Convert to 1-based line numbering
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	err = api.fs.WriteInfoFile(infoFilePath, newContent)
	return err == nil
}

// determineInfoFile determines the appropriate .info file path for a target path
func (api *InfoAPI) determineInfoFile(targetPath string) string {
	// Simple strategy: use .info file in the same directory as the target
	dir := filepath.Dir(targetPath)
	if dir == "" || dir == "." {
		return ".info"
	}
	return filepath.Join(dir, ".info")
}

// readFileContent reads content from a file, returning empty string if file doesn't exist
func (api *InfoAPI) readFileContent(filePath string) string {
	reader, err := api.fs.ReadInfoFile(filePath)
	if err != nil {
		return "" // File doesn't exist, return empty content
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "" // Error reading, return empty content
	}

	return string(data)
}

// CleanResult contains the results of a clean operation
type CleanResult struct {
	RemovedAnnotations []Annotation `json:"removed_annotations"`
	UpdatedFiles       []string     `json:"updated_files"`
	Summary            CleanSummary `json:"summary"`
}

// CleanSummary provides counts of clean operations
type CleanSummary struct {
	InvalidPathsRemoved int `json:"invalid_paths_removed"`
	DuplicatesRemoved   int `json:"duplicates_removed"`
	FilesModified       int `json:"files_modified"`
}
