package info

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// ValidationIssueType represents the type of validation issue found
type ValidationIssueType string

const (
	// IssueInvalidFormat indicates a line with invalid format (no annotation)
	IssueInvalidFormat ValidationIssueType = "invalid_format"
	// IssueDuplicatePath indicates duplicate paths within the same .info file
	IssueDuplicatePath ValidationIssueType = "duplicate_path"
	// IssuePathNotExists indicates the annotated path doesn't exist in filesystem
	IssuePathNotExists ValidationIssueType = "path_not_exists"
	// IssueAncestorPath indicates the path points to an ancestor directory
	IssueAncestorPath ValidationIssueType = "ancestor_path"
	// IssueMultipleFiles indicates multiple .info files annotate the same path
	IssueMultipleFiles ValidationIssueType = "multiple_files"
)

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Type        ValidationIssueType `json:"type"`
	InfoFile    string              `json:"info_file"`
	LineNum     int                 `json:"line_num"`
	Path        string              `json:"path"`
	Message     string              `json:"message"`
	Suggestion  string              `json:"suggestion,omitempty"`
	RelatedFile string              `json:"related_file,omitempty"` // For multiple_files issues
}

// ValidationResult contains the results of validating .info files
type ValidationResult struct {
	Issues       []ValidationIssue `json:"issues"`
	ValidFiles   []string          `json:"valid_files"`
	InvalidFiles []string          `json:"invalid_files"`
	Summary      ValidationSummary `json:"summary"`
}

// ValidationSummary provides counts of different issue types
type ValidationSummary struct {
	TotalFiles       int                         `json:"total_files"`
	TotalIssues      int                         `json:"total_issues"`
	IssuesByType     map[ValidationIssueType]int `json:"issues_by_type"`
	IssuesByFile     map[string]int              `json:"issues_by_file"`
	ValidAnnotations int                         `json:"valid_annotations"`
	TotalAnnotations int                         `json:"total_annotations"`
}

// Validator provides validation functionality for .info files
type Validator struct {
	parser *Parser
	logger Logger
}

// NewInfoValidator creates a new info file validator
func NewInfoValidator() *Validator {
	return &Validator{
		parser: NewParser(),
	}
}

// NewInfoValidatorWithLogger creates a new info file validator with a custom logger
func NewInfoValidatorWithLogger(logger Logger) *Validator {
	return &Validator{
		parser: NewParserWithLogger(logger),
		logger: logger,
	}
}

// ValidateContent validates .info file content without file system access
func (v *Validator) ValidateContent(infoFiles InfoFileMap, pathExists func(string) bool) (*ValidationResult, error) {
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

	// Validate each .info file individually
	allAnnotations := make([]Annotation, 0)
	for infoFile, content := range infoFiles {
		fileIssues, annotations, err := v.validateSingleFileContent(infoFile, content, pathExists)
		if err != nil {
			// If we can't parse the file, add it as invalid and continue
			result.InvalidFiles = append(result.InvalidFiles, infoFile)
			result.Issues = append(result.Issues, ValidationIssue{
				Type:     IssueInvalidFormat,
				InfoFile: infoFile,
				Message:  "Cannot parse .info file: " + err.Error(),
			})
			continue
		}

		// Add issues from this file
		result.Issues = append(result.Issues, fileIssues...)
		allAnnotations = append(allAnnotations, annotations...)

		// Classify file as valid or invalid
		if len(fileIssues) == 0 {
			result.ValidFiles = append(result.ValidFiles, infoFile)
		} else {
			result.InvalidFiles = append(result.InvalidFiles, infoFile)
		}

		// Update summary
		result.Summary.IssuesByFile[infoFile] = len(fileIssues)
		for _, issue := range fileIssues {
			result.Summary.IssuesByType[issue.Type]++
		}
	}

	// Check for cross-file conflicts (multiple .info files annotating same path)
	crossFileIssues := v.findCrossFileConflicts(allAnnotations)
	result.Issues = append(result.Issues, crossFileIssues...)
	for _, issue := range crossFileIssues {
		result.Summary.IssuesByType[issue.Type]++
		result.Summary.IssuesByFile[issue.InfoFile]++

		// Reclassify files that now have cross-file issues as invalid
		for i, validFile := range result.ValidFiles {
			if validFile == issue.InfoFile {
				// Move from valid to invalid
				result.ValidFiles = append(result.ValidFiles[:i], result.ValidFiles[i+1:]...)
				result.InvalidFiles = append(result.InvalidFiles, issue.InfoFile)
				break
			}
		}
	}

	// Calculate summary statistics
	result.Summary.TotalIssues = len(result.Issues)
	result.Summary.TotalAnnotations = len(allAnnotations)

	// Count valid annotations (those without path existence or ancestor issues)
	validCount := 0
	for _, annotation := range allAnnotations {
		hasPathIssue := false
		for _, issue := range result.Issues {
			if issue.InfoFile == annotation.InfoFile &&
				issue.LineNum == annotation.LineNum &&
				(issue.Type == IssuePathNotExists || issue.Type == IssueAncestorPath) {
				hasPathIssue = true
				break
			}
		}
		if !hasPathIssue {
			validCount++
		}
	}
	result.Summary.ValidAnnotations = validCount

	return result, nil
}

// ValidateFileSystem validates all .info files in a directory tree using file system
func (v *Validator) ValidateFileSystem(fs InfoFileSystem, rootPath string) (*ValidationResult, error) {
	// Find all .info files
	infoFiles, err := fs.FindInfoFiles(rootPath)
	if err != nil {
		return nil, err
	}

	// Read all files into map
	infoFileMap := make(InfoFileMap)
	for _, infoFile := range infoFiles {
		reader, err := fs.ReadInfoFile(infoFile)
		if err != nil {
			// We'll handle this in ValidateContent as a parse error
			infoFileMap[infoFile] = ""
			continue
		}

		// Read all content from reader
		content, err := io.ReadAll(reader)
		if err != nil {
			infoFileMap[infoFile] = ""
			continue
		}
		infoFileMap[infoFile] = string(content)
	}

	return v.ValidateContent(infoFileMap, fs.PathExists)
}

// validateSingleFileContent validates a single .info file's content
func (v *Validator) validateSingleFileContent(infoFilePath, content string, pathExists func(string) bool) ([]ValidationIssue, []Annotation, error) {
	issues := make([]ValidationIssue, 0)

	// Parse file and collect all annotations (including invalid ones)
	annotations, parseIssues := v.parseFileWithValidation(content, infoFilePath)
	issues = append(issues, parseIssues...)

	// Validate path existence and scope for valid annotations
	infoDir := filepath.Dir(infoFilePath)
	for _, annotation := range annotations {
		pathIssues := v.validateAnnotationPath(annotation, infoDir, pathExists)
		issues = append(issues, pathIssues...)
	}

	return issues, annotations, nil
}

// parseFileWithValidation parses a .info file and reports format/duplicate issues
func (v *Validator) parseFileWithValidation(content, infoFilePath string) ([]Annotation, []ValidationIssue) {
	var annotations []Annotation
	var issues []ValidationIssue

	lines := strings.Split(content, "\n")
	parsedPaths := make(map[string]int) // path -> first line number

	for lineNum, line := range lines {
		lineNum++ // Convert to 1-based line numbering
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line
		path, annotation, parseOk := v.parser.ParseLine(line)

		if !parseOk {
			// Invalid format
			issues = append(issues, ValidationIssue{
				Type:       IssueInvalidFormat,
				InfoFile:   infoFilePath,
				LineNum:    lineNum,
				Path:       strings.Fields(line)[0], // Best guess at path
				Message:    "Invalid line format: missing annotation or invalid syntax",
				Suggestion: "Format should be: <path> <annotation>",
			})
			continue
		}

		// Check for duplicates
		if firstLine, exists := parsedPaths[path]; exists {
			issues = append(issues, ValidationIssue{
				Type:       IssueDuplicatePath,
				InfoFile:   infoFilePath,
				LineNum:    lineNum,
				Path:       path,
				Message:    fmt.Sprintf("Duplicate path %q (first occurrence at line %d)", path, firstLine),
				Suggestion: "Remove duplicate entry or use different path",
			})
			continue // Skip duplicate
		}

		// Record this path and create annotation
		parsedPaths[path] = lineNum
		annotations = append(annotations, Annotation{
			Path:       path,
			Annotation: annotation,
			InfoFile:   infoFilePath,
			LineNum:    lineNum,
		})
	}

	return annotations, issues
}

// validateAnnotationPath validates a single annotation's path
func (v *Validator) validateAnnotationPath(annotation Annotation, infoDir string, pathExists func(string) bool) []ValidationIssue {
	var issues []ValidationIssue

	// Calculate target path
	targetPath := filepath.Join(infoDir, annotation.Path)
	targetPath = filepath.Clean(targetPath)

	// Check if path exists
	if !pathExists(targetPath) {
		issues = append(issues, ValidationIssue{
			Type:       IssuePathNotExists,
			InfoFile:   annotation.InfoFile,
			LineNum:    annotation.LineNum,
			Path:       annotation.Path,
			Message:    fmt.Sprintf("Path %q does not exist", annotation.Path),
			Suggestion: "Verify the path is correct or create the missing file/directory",
		})
	}

	// Check ancestor relationship
	rel, err := filepath.Rel(targetPath, infoDir)
	if (err == nil && !strings.HasPrefix(rel, "..") && rel != ".") ||
		(err != nil && strings.Contains(err.Error(), "can't make")) {
		issues = append(issues, ValidationIssue{
			Type:       IssueAncestorPath,
			InfoFile:   annotation.InfoFile,
			LineNum:    annotation.LineNum,
			Path:       annotation.Path,
			Message:    fmt.Sprintf("Cannot annotate ancestor path %q", annotation.Path),
			Suggestion: "Move annotation to a .info file at or below the target path",
		})
	}

	return issues
}

// findCrossFileConflicts finds cases where multiple .info files annotate the same path
func (v *Validator) findCrossFileConflicts(annotations []Annotation) []ValidationIssue {
	var issues []ValidationIssue

	// Group annotations by resolved target path
	pathGroups := make(map[string][]Annotation)
	for _, annotation := range annotations {
		infoDir := filepath.Dir(annotation.InfoFile)
		targetPath := filepath.Join(infoDir, annotation.Path)
		targetPath = filepath.Clean(targetPath)
		pathGroups[targetPath] = append(pathGroups[targetPath], annotation)
	}

	// Find paths with multiple annotations
	for targetPath, group := range pathGroups {
		if len(group) > 1 {
			// Sort by depth (deepest first) to identify winner
			sort.Slice(group, func(i, j int) bool {
				dirI := filepath.Dir(group[i].InfoFile)
				dirJ := filepath.Dir(group[j].InfoFile)
				depthI := pathDepth(dirI)
				depthJ := pathDepth(dirJ)
				if depthI != depthJ {
					return depthI > depthJ
				}
				return dirI < dirJ
			})

			winner := group[0]
			for _, annotation := range group[1:] {
				issues = append(issues, ValidationIssue{
					Type:        IssueMultipleFiles,
					InfoFile:    annotation.InfoFile,
					LineNum:     annotation.LineNum,
					Path:        annotation.Path,
					Message:     fmt.Sprintf("Path %q is also annotated in %s (which takes precedence)", targetPath, winner.InfoFile),
					Suggestion:  "Consider removing redundant annotation or moving to distribute files",
					RelatedFile: winner.InfoFile,
				})
			}
		}
	}

	return issues
}
