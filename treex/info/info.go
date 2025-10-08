package info

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/afero"
)

// Logger interface for warning reporting during info processing
type Logger interface {
	Printf(format string, v ...interface{})
}

// Annotation represents a single entry in an .info file.
type Annotation struct {
	Path       string
	Annotation string
	InfoFile   string // The path to the .info file this annotation came from.
	LineNum    int
}

// Parse reads an .info file from an io.Reader and returns a list of annotations.
func Parse(reader io.Reader, infoFilePath string) ([]Annotation, error) {
	return ParseWithLogger(reader, infoFilePath, nil)
}

// ParseWithLogger reads an .info file from an io.Reader and returns a list of annotations,
// using the provided logger for warnings.
func ParseWithLogger(reader io.Reader, infoFilePath string, logger Logger) ([]Annotation, error) {
	logf := func(format string, v ...interface{}) {
		if logger != nil {
			logger.Printf(format, v...)
		}
		// If no logger, silently ignore warnings during parsing
	}

	var annotations []Annotation
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	parsedPaths := make(map[string]bool)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var path, annotation string
		var pathEnd = -1
		var isEscaped = false

		// Find the first unescaped whitespace to split path and annotation.
		for i, r := range line {
			if unicode.IsSpace(r) && !isEscaped {
				pathEnd = i
				break
			}
			if r == '\\' && !isEscaped {
				isEscaped = true
			} else {
				isEscaped = false
			}
		}

		if pathEnd == -1 {
			logf("info: ignoring line %d in %q: no annotation found (missing space separator)", lineNum, infoFilePath)
			continue // Line has no space separator, so no annotation.
		}

		path = line[:pathEnd]
		annotation = strings.TrimSpace(line[pathEnd+1:])

		if annotation == "" {
			logf("info: ignoring line %d in %q: empty annotation for path %q", lineNum, infoFilePath, path)
			continue // No annotation content.
		}

		// Unescape spaces in the path.
		path = strings.ReplaceAll(path, "\\ ", " ")

		// Per spec, first entry for a path in a file wins.
		if parsedPaths[path] {
			logf("info: ignoring duplicate path %q at line %d in %q (first occurrence wins)", path, lineNum, infoFilePath)
			continue
		}
		parsedPaths[path] = true

		annotations = append(annotations, Annotation{
			Path:       path,
			Annotation: annotation,
			InfoFile:   infoFilePath,
			LineNum:    lineNum,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return annotations, nil
}

// Collector manages the collection and merging of annotations.
type Collector struct {
	logger Logger
}

// NewCollector creates a new annotation collector.
func NewCollector() *Collector {
	return &Collector{}
}

// NewCollectorWithLogger creates a new annotation collector with a custom logger.
func NewCollectorWithLogger(logger Logger) *Collector {
	return &Collector{
		logger: logger,
	}
}

// logf logs a warning message using the configured logger, or log.Printf if no logger is set
func (c *Collector) logf(format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}

// CollectAnnotations walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
func (c *Collector) CollectAnnotations(fsys afero.Fs, root string) (map[string]Annotation, error) {
	var allAnnotations []Annotation

	err := afero.Walk(fsys, root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == ".info" {
			file, err := fsys.Open(path)
			if err != nil {
				c.logf("info: cannot open .info file %q: %v", path, err)
				return nil
			}
			defer func() {
				_ = file.Close() // Error is intentionally ignored
			}()

			annotations, err := ParseWithLogger(file, path, c.logger)
			if err != nil {
				c.logf("info: cannot parse .info file %q: %v", path, err)
				return nil
			}
			allAnnotations = append(allAnnotations, annotations...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return c.merge(fsys, allAnnotations), nil
}

func (c *Collector) merge(fsys afero.Fs, annotations []Annotation) map[string]Annotation {
	contenders := make(map[string][]Annotation)
	for _, ann := range annotations {
		infoDir := filepath.Dir(ann.InfoFile)
		targetPath := filepath.Join(infoDir, ann.Path)

		// Normalize paths to handle "." and other relative parts.
		targetPath = filepath.Clean(targetPath)

		// Rule: .info files can't annotate their ancestors.
		// Check if targetPath is an ancestor of infoDir
		rel, err := filepath.Rel(targetPath, infoDir)

		// Two cases indicate ancestor relationship:
		// 1. Rel succeeds and infoDir is contained within targetPath (rel doesn't start with "..")
		// 2. Rel fails because targetPath is above infoDir in the hierarchy
		if (err == nil && !strings.HasPrefix(rel, "..") && rel != ".") ||
			(err != nil && strings.Contains(err.Error(), "can't make")) {
			c.logf("info: invalid annotation in %q: cannot annotate ancestor path %q", ann.InfoFile, ann.Path)
			continue
		}

		// Validate that the target path exists in the filesystem
		if _, err := fsys.Stat(targetPath); err != nil {
			c.logf("info: invalid annotation in %q: path %q does not exist", ann.InfoFile, ann.Path)
			continue
		}

		contenders[targetPath] = append(contenders[targetPath], ann)
	}

	winner := make(map[string]Annotation)
	for path, anns := range contenders {
		sort.Slice(anns, func(i, j int) bool {
			dirI := filepath.Dir(anns[i].InfoFile)
			dirJ := filepath.Dir(anns[j].InfoFile)

			// Rule: closest (deepest) .info file wins.
			// Calculate depth correctly, handling "." as root (depth 0)
			depthI := pathDepth(dirI)
			depthJ := pathDepth(dirJ)

			if depthI != depthJ {
				return depthI > depthJ // Deeper path wins
			}

			// Rule: if distance is same, lexicographical order of .info file dir wins.
			if dirI != dirJ {
				return dirI < dirJ
			}

			// Rule: if same .info file, lower line number wins.
			return anns[i].LineNum < anns[j].LineNum
		})
		winner[path] = anns[0]
	}

	return winner
}

// pathDepth calculates the depth of a directory path, with "." being depth 0
func pathDepth(dir string) int {
	if dir == "." {
		return 0
	}
	// Clean the path to handle any redundant separators
	clean := filepath.Clean(dir)
	if clean == "." {
		return 0
	}
	// Count the separators
	return strings.Count(clean, string(filepath.Separator)) + 1
}

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
	fs     afero.Fs
	logger Logger
}

// NewValidator creates a new info file validator
func NewValidator(fs afero.Fs) *Validator {
	return &Validator{
		fs: fs,
	}
}

// NewValidatorWithLogger creates a new info file validator with a custom logger
func NewValidatorWithLogger(fs afero.Fs, logger Logger) *Validator {
	return &Validator{
		fs:     fs,
		logger: logger,
	}
}

// ValidateDirectory validates all .info files in a directory tree
func (v *Validator) ValidateDirectory(rootPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Issues:       make([]ValidationIssue, 0),
		ValidFiles:   make([]string, 0),
		InvalidFiles: make([]string, 0),
		Summary: ValidationSummary{
			IssuesByType: make(map[ValidationIssueType]int),
			IssuesByFile: make(map[string]int),
		},
	}

	// Find all .info files
	infoFiles := make([]string, 0)
	err := afero.Walk(v.fs, rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}
		if !info.IsDir() && filepath.Base(path) == ".info" {
			infoFiles = append(infoFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result.Summary.TotalFiles = len(infoFiles)

	// Validate each .info file individually
	allAnnotations := make([]Annotation, 0)
	for _, infoFile := range infoFiles {
		fileIssues, annotations, err := v.validateSingleFile(infoFile, rootPath)
		if err != nil {
			// If we can't read the file, add it as invalid and continue
			result.InvalidFiles = append(result.InvalidFiles, infoFile)
			result.Issues = append(result.Issues, ValidationIssue{
				Type:     IssueInvalidFormat,
				InfoFile: infoFile,
				Message:  "Cannot read .info file: " + err.Error(),
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

// validateSingleFile validates a single .info file
func (v *Validator) validateSingleFile(infoFilePath, rootPath string) ([]ValidationIssue, []Annotation, error) {
	issues := make([]ValidationIssue, 0)

	// Read and parse the file
	file, err := v.fs.Open(infoFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()

	// Parse file and collect all annotations (including invalid ones)
	annotations, parseIssues := v.parseFileWithValidation(file, infoFilePath)
	issues = append(issues, parseIssues...)

	// Validate path existence and scope for valid annotations
	infoDir := filepath.Dir(infoFilePath)
	for _, annotation := range annotations {
		pathIssues := v.validateAnnotationPath(annotation, infoDir, rootPath)
		issues = append(issues, pathIssues...)
	}

	return issues, annotations, nil
}

// parseFileWithValidation parses a .info file and reports format/duplicate issues
func (v *Validator) parseFileWithValidation(reader io.Reader, infoFilePath string) ([]Annotation, []ValidationIssue) {
	var annotations []Annotation
	var issues []ValidationIssue

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	parsedPaths := make(map[string]int) // path -> first line number

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line
		path, annotation, parseOk := v.parseLine(line)

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

// parseLine parses a single line and returns path, annotation, and success flag
func (v *Validator) parseLine(line string) (string, string, bool) {
	var path, annotation string
	var pathEnd = -1
	var isEscaped = false

	// Find the first unescaped whitespace to split path and annotation
	for i, r := range line {
		if unicode.IsSpace(r) && !isEscaped {
			pathEnd = i
			break
		}
		if r == '\\' && !isEscaped {
			isEscaped = true
		} else {
			isEscaped = false
		}
	}

	if pathEnd == -1 {
		return "", "", false // No space separator found
	}

	path = line[:pathEnd]
	annotation = strings.TrimSpace(line[pathEnd+1:])

	if annotation == "" {
		return path, "", false // Empty annotation
	}

	// Unescape spaces in the path
	path = strings.ReplaceAll(path, "\\ ", " ")

	return path, annotation, true
}

// validateAnnotationPath validates a single annotation's path
func (v *Validator) validateAnnotationPath(annotation Annotation, infoDir, rootPath string) []ValidationIssue {
	var issues []ValidationIssue

	// Calculate target path
	targetPath := filepath.Join(infoDir, annotation.Path)
	targetPath = filepath.Clean(targetPath)

	// Check if path exists
	if _, err := v.fs.Stat(targetPath); err != nil {
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
