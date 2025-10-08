package info

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// InfoAPI provides the main interface for info file operations
type InfoAPI struct {
	fs       InfoFileSystem
	gatherer *Gatherer
	loader   *InfoFileLoader
	writer   *InfoFileWriter
}

// NewInfoAPI creates a new info API instance using afero filesystem
func NewInfoAPI(fs afero.Fs) *InfoAPI {
	afs := NewAferoInfoFileSystem(fs)
	return &InfoAPI{
		fs:       afs,
		gatherer: NewGatherer(),
		loader:   NewInfoFileLoader(afs),
		writer:   NewInfoFileWriter(afs),
	}
}

// NewInfoAPIWithFileSystem creates a new info API instance with custom filesystem
func NewInfoAPIWithFileSystem(fs InfoFileSystem) *InfoAPI {
	return &InfoAPI{
		fs:       fs,
		gatherer: NewGatherer(),
		loader:   NewInfoFileLoader(fs),
		writer:   NewInfoFileWriter(fs),
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
	// Load all InfoFiles
	infoFiles, err := api.loader.LoadInfoFiles(rootPath)
	if err != nil {
		return nil, err
	}

	return api.validateInfoFiles(infoFiles), nil
}

// Add adds a new annotation to the appropriate .info file
func (api *InfoAPI) Add(targetPath, annotation string) error {
	// Determine the appropriate .info file for this target path
	infoFilePath := api.determineInfoFile(targetPath)

	// Load existing InfoFile or create new one
	var infoFile *InfoFile
	existingInfoFile, err := api.loader.LoadInfoFile(infoFilePath)
	if err != nil {
		// File doesn't exist, create new empty InfoFile
		infoFile = NewInfoFile(infoFilePath, "")
	} else {
		infoFile = existingInfoFile
	}

	// Convert target path to relative path for the .info file
	relativePath := api.makeRelativePathForAdd(targetPath, infoFilePath)

	// Add annotation using InfoFile method
	success := infoFile.AddAnnotationForPath(relativePath, annotation)
	if !success {
		return fmt.Errorf("annotation already exists for path %q", relativePath)
	}

	// Write updated InfoFile back to disk
	return api.writer.WriteInfoFile(infoFile)
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

	// Load the InfoFile containing the annotation
	infoFile, err := api.loader.LoadInfoFile(targetAnnotation.InfoFile)
	if err != nil {
		return fmt.Errorf("cannot load .info file %q: %w", targetAnnotation.InfoFile, err)
	}

	// Convert target path to relative path for the .info file
	relativePath := api.makeRelativePathForAdd(targetPath, targetAnnotation.InfoFile)

	// Remove annotation using InfoFile method
	success := infoFile.RemoveAnnotationForPath(relativePath)
	if !success {
		return fmt.Errorf("annotation not found in content for path %q", relativePath)
	}

	// Write updated InfoFile back to disk
	return api.writer.WriteInfoFile(infoFile)
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

	// Load the InfoFile containing the annotation
	infoFile, err := api.loader.LoadInfoFile(targetAnnotation.InfoFile)
	if err != nil {
		return fmt.Errorf("cannot load .info file %q: %w", targetAnnotation.InfoFile, err)
	}

	// Convert target path to relative path for the .info file
	relativePath := api.makeRelativePathForAdd(targetPath, targetAnnotation.InfoFile)

	// Update annotation using InfoFile method
	success := infoFile.UpdateAnnotationForPath(relativePath, newAnnotation)
	if !success {
		return fmt.Errorf("annotation not found in content for path %q", relativePath)
	}

	// Write updated InfoFile back to disk
	return api.writer.WriteInfoFile(infoFile)
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

	// Process each file with issues using InfoFile
	for infoFilePath, issues := range fileIssues {
		if api.cleanInfoFile(infoFilePath, issues, result) {
			result.UpdatedFiles = append(result.UpdatedFiles, infoFilePath)
			result.Summary.FilesModified++
		}
	}

	return result, nil
}

// cleanInfoFile removes problematic annotations from a single .info file using InfoFile
func (api *InfoAPI) cleanInfoFile(infoFilePath string, issues []ValidationIssue, result *CleanResult) bool {
	// Load the InfoFile
	infoFile, err := api.loader.LoadInfoFile(infoFilePath)
	if err != nil {
		return false
	}

	modified := false

	// Process issues by type
	for _, issue := range issues {
		switch issue.Type {
		case IssuePathNotExists, IssueInvalidFormat, IssueDuplicatePath, IssueAncestorPath:
			// For valid annotations with path issues, remove them
			if issue.Path != "" && infoFile.HasAnnotationForPath(issue.Path) {
				if infoFile.RemoveAnnotationForPath(issue.Path) {
					modified = true

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

			// For malformed lines, mark them as removed (they're already marked in InfoFile)
			if issue.Type == IssueInvalidFormat || issue.Type == IssueDuplicatePath {
				for i := range infoFile.Lines {
					if infoFile.Lines[i].LineNum == issue.LineNum && infoFile.Lines[i].Type == LineTypeMalformed {
						infoFile.Lines[i].ParseError = "removed"
						modified = true
						break
					}
				}
			}
		case IssueMultipleFiles:
			// For cross-file conflicts, remove the annotation from the later file
			if infoFile.HasAnnotationForPath(issue.Path) {
				if infoFile.RemoveAnnotationForPath(issue.Path) {
					modified = true
					result.RemovedAnnotations = append(result.RemovedAnnotations, Annotation{
						Path:       issue.Path,
						InfoFile:   issue.InfoFile,
						LineNum:    issue.LineNum,
						Annotation: fmt.Sprintf("(removed: %s)", issue.Type),
					})
				}
			}
		}
	}

	if !modified {
		return false
	}

	// Write the cleaned InfoFile back to disk
	err = api.writer.WriteInfoFile(infoFile)
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

// makeRelativePathForAdd converts target path to relative path for the .info file
func (api *InfoAPI) makeRelativePathForAdd(targetPath, infoFilePath string) string {
	infoDir := filepath.Dir(infoFilePath)
	targetDir := filepath.Dir(targetPath)

	// If target is in same directory as info file, use just the filename
	if infoDir == targetDir || (infoDir == "." && targetDir == "") {
		return filepath.Base(targetPath)
	}

	// Calculate relative path from info file directory to target
	rel, err := filepath.Rel(infoDir, targetPath)
	if err != nil {
		// Fallback to the original path if calculation fails
		return targetPath
	}

	return filepath.Clean(rel)
}

// validateInfoFiles performs validation using InfoFile parsed data
func (api *InfoAPI) validateInfoFiles(infoFiles []*InfoFile) *ValidationResult {
	result := &ValidationResult{
		Issues:       make([]ValidationIssue, 0),
		ValidFiles:   make([]string, 0),
		InvalidFiles: make([]string, 0),
		Summary: ValidationSummary{
			IssuesByType: make(map[ValidationIssueType]int),
			IssuesByFile: make(map[string]int),
		},
	}

	result.Summary.TotalFiles = len(infoFiles)
	allAnnotations := make([]Annotation, 0)

	// Process each InfoFile
	for _, infoFile := range infoFiles {
		fileIssues := make([]ValidationIssue, 0)

		// Check malformed lines from InfoFile
		for _, line := range infoFile.Lines {
			if line.Type == LineTypeMalformed {
				issueType := IssueInvalidFormat
				if line.ParseError == "duplicate path (first occurrence wins)" {
					issueType = IssueDuplicatePath
				}

				fileIssues = append(fileIssues, ValidationIssue{
					Type:     issueType,
					InfoFile: infoFile.Path,
					LineNum:  line.LineNum,
					Path:     "", // No valid path for malformed lines
					Message:  line.ParseError,
				})
			}
		}

		// Check valid annotations for path existence and ancestor issues
		for _, annotation := range infoFile.GetAllAnnotations() {
			infoDir := filepath.Dir(infoFile.Path)
			targetPath := filepath.Join(infoDir, annotation.Path)
			targetPath = filepath.Clean(targetPath)

			// Check if path exists
			if !api.fs.PathExists(targetPath) {
				fileIssues = append(fileIssues, ValidationIssue{
					Type:     IssuePathNotExists,
					InfoFile: infoFile.Path,
					LineNum:  annotation.LineNum,
					Path:     annotation.Path,
					Message:  fmt.Sprintf("path does not exist: %s", targetPath),
				})
			}

			// Check for ancestor path annotations
			rel, err := filepath.Rel(infoDir, targetPath)
			if err == nil && strings.HasPrefix(rel, "..") {
				fileIssues = append(fileIssues, ValidationIssue{
					Type:     IssueAncestorPath,
					InfoFile: infoFile.Path,
					LineNum:  annotation.LineNum,
					Path:     annotation.Path,
					Message:  "cannot annotate ancestor path",
				})
			}

			allAnnotations = append(allAnnotations, annotation)
		}

		// Classify file and update summary
		if len(fileIssues) == 0 {
			result.ValidFiles = append(result.ValidFiles, infoFile.Path)
		} else {
			result.InvalidFiles = append(result.InvalidFiles, infoFile.Path)
		}

		result.Issues = append(result.Issues, fileIssues...)
		result.Summary.IssuesByFile[infoFile.Path] = len(fileIssues)
		for _, issue := range fileIssues {
			result.Summary.IssuesByType[issue.Type]++
		}
	}

	// Check for cross-file conflicts
	crossFileIssues := api.findCrossFileConflicts(allAnnotations)
	result.Issues = append(result.Issues, crossFileIssues...)
	for _, issue := range crossFileIssues {
		result.Summary.IssuesByType[issue.Type]++
		result.Summary.IssuesByFile[issue.InfoFile]++

		// Reclassify files with cross-file issues as invalid
		for i, validFile := range result.ValidFiles {
			if validFile == issue.InfoFile {
				result.ValidFiles = append(result.ValidFiles[:i], result.ValidFiles[i+1:]...)
				result.InvalidFiles = append(result.InvalidFiles, issue.InfoFile)
				break
			}
		}
	}

	result.Summary.TotalIssues = len(result.Issues)
	return result
}

// findCrossFileConflicts checks for multiple .info files annotating the same path
func (api *InfoAPI) findCrossFileConflicts(annotations []Annotation) []ValidationIssue {
	pathToFiles := make(map[string][]Annotation)

	// Group annotations by resolved target path
	for _, ann := range annotations {
		infoDir := filepath.Dir(ann.InfoFile)
		targetPath := filepath.Join(infoDir, ann.Path)
		targetPath = filepath.Clean(targetPath)

		pathToFiles[targetPath] = append(pathToFiles[targetPath], ann)
	}

	var issues []ValidationIssue
	for _, anns := range pathToFiles {
		if len(anns) > 1 {
			// Multiple files annotate the same path - create issues for all but the first
			for i := 1; i < len(anns); i++ {
				issues = append(issues, ValidationIssue{
					Type:        IssueMultipleFiles,
					InfoFile:    anns[i].InfoFile,
					LineNum:     anns[i].LineNum,
					Path:        anns[i].Path,
					Message:     fmt.Sprintf("path already annotated in %s", anns[0].InfoFile),
					RelatedFile: anns[0].InfoFile,
				})
			}
		}
	}

	return issues
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
