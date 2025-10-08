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
	logger    Logger
}

// NewInfoAPI creates a new info API instance using afero filesystem
func NewInfoAPI(fs afero.Fs) *InfoAPI {
	afs := NewAferoInfoFileSystem(fs)
	return &InfoAPI{
		fs:        afs,
		gatherer:  NewGatherer(),
		validator: NewInfoValidator(),
	}
}

// NewInfoAPIWithLogger creates a new info API instance with custom logger
func NewInfoAPIWithLogger(fs afero.Fs, logger Logger) *InfoAPI {
	afs := NewAferoInfoFileSystem(fs)
	return &InfoAPI{
		fs:        afs,
		gatherer:  NewGathererWithLogger(logger),
		validator: NewInfoValidatorWithLogger(logger),
		logger:    logger,
	}
}

// NewInfoAPIWithFileSystem creates a new info API instance with custom filesystem
func NewInfoAPIWithFileSystem(fs InfoFileSystem) *InfoAPI {
	return &InfoAPI{
		fs:        fs,
		gatherer:  NewGatherer(),
		validator: NewInfoValidator(),
	}
}

// Gather collects and merges all annotations from .info files in a directory tree
func (api *InfoAPI) Gather(rootPath string) (map[string]Annotation, error) {
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
	infoDir := filepath.Dir(infoFilePath)

	// Calculate the relative path from the .info file to the target
	relativePath, err := filepath.Rel(infoDir, targetPath)
	if err != nil {
		// If we can't make it relative, use the original path
		relativePath = targetPath
	}

	// Read existing content
	content := ""
	if reader, err := api.fs.ReadInfoFile(infoFilePath); err == nil {
		if data, err := io.ReadAll(reader); err == nil {
			content = string(data)
		}
	}

	// Escape spaces in target path for .info format
	escapedPath := strings.ReplaceAll(relativePath, " ", "\\ ")

	// Add new annotation
	newLine := fmt.Sprintf("%s  %s", escapedPath, annotation)
	if content == "" {
		content = newLine + "\n"
	} else {
		content = strings.TrimSpace(content) + "\n" + newLine + "\n"
	}

	return api.fs.WriteInfoFile(infoFilePath, content)
}

// Remove removes an annotation for a specific path
func (api *InfoAPI) Remove(targetPath string) error {
	// Find which .info file contains this annotation
	annotations, err := api.Gather(".")
	if err != nil {
		return err
	}

	var targetAnnotation *Annotation
	for path, ann := range annotations {
		if path == targetPath {
			targetAnnotation = &ann
			break
		}
	}

	if targetAnnotation == nil {
		return fmt.Errorf("no annotation found for path %q", targetPath)
	}

	// Read the .info file
	reader, err := api.fs.ReadInfoFile(targetAnnotation.InfoFile)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Remove the specific line
	var newLines []string
	for i, line := range lines {
		if i+1 == targetAnnotation.LineNum {
			continue // Skip the target line
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	return api.fs.WriteInfoFile(targetAnnotation.InfoFile, newContent)
}

// Update updates an existing annotation
func (api *InfoAPI) Update(targetPath, newAnnotation string) error {
	// Find which .info file contains this annotation
	annotations, err := api.Gather(".")
	if err != nil {
		return err
	}

	var targetAnnotation *Annotation
	for path, ann := range annotations {
		if path == targetPath {
			targetAnnotation = &ann
			break
		}
	}

	if targetAnnotation == nil {
		return fmt.Errorf("no annotation found for path %q", targetPath)
	}

	// Read the .info file
	reader, err := api.fs.ReadInfoFile(targetAnnotation.InfoFile)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Update the specific line
	for i := range lines {
		if i+1 == targetAnnotation.LineNum {
			// Escape spaces in target path for .info format
			escapedPath := strings.ReplaceAll(targetAnnotation.Path, " ", "\\ ")
			lines[i] = fmt.Sprintf("%s  %s", escapedPath, newAnnotation)
			break
		}
	}

	newContent := strings.Join(lines, "\n")
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
