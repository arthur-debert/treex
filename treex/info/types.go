package info

import (
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jwaldrip/treex/treex/logging"
)

// ============================================================================
// VALIDATION TYPES
// ============================================================================

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

// ============================================================================
// GATHERER TYPE AND IMPLEMENTATION
// ============================================================================

// Gatherer coordinates the collection and merging of annotations
type Gatherer struct {
	parser *Parser
}

// NewGatherer creates a new gatherer instance
func NewGatherer() *Gatherer {
	return &Gatherer{
		parser: NewParser(),
	}
}

// GatherFromInfoFiles takes a collection of InfoFile structs and returns merged annotations.
// This method works purely in memory without file system access and eliminates string parsing.
func (g *Gatherer) GatherFromInfoFiles(infoFiles []*InfoFile, pathExists func(string) bool) map[string]Annotation {
	var allAnnotations []Annotation

	for _, infoFile := range infoFiles {
		annotations := infoFile.GetAllAnnotations()
		allAnnotations = append(allAnnotations, annotations...)
	}

	return g.mergeAnnotations(allAnnotations, pathExists)
}

// GatherFromFileSystem walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
// DEPRECATED: Use InfoAPI.Gather() which uses InfoFile-based loading for better performance
func (g *Gatherer) GatherFromFileSystem(fs InfoFileSystem, root string) (map[string]Annotation, error) {
	infoFilePaths, err := fs.FindInfoFiles(root)
	if err != nil {
		return nil, err
	}

	var allAnnotations []Annotation

	for _, infoFilePath := range infoFilePaths {
		reader, err := fs.ReadInfoFile(infoFilePath)
		if err != nil {
			logging.Warn().Str("file", infoFilePath).Err(err).Msg("cannot open .info file")
			continue
		}

		// Read content from reader and parse
		data, err := io.ReadAll(reader)
		if err != nil {
			logging.Warn().Str("file", infoFilePath).Err(err).Msg("cannot read .info file")
			continue
		}

		annotations := g.parser.Parse(string(data), infoFilePath)
		allAnnotations = append(allAnnotations, annotations...)
	}

	return g.mergeAnnotations(allAnnotations, fs.PathExists), nil
}

// mergeAnnotations implements the annotation merging logic
func (g *Gatherer) mergeAnnotations(annotations []Annotation, pathExists func(string) bool) map[string]Annotation {
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
			logging.Warn().Str("info_file", ann.InfoFile).Str("path", ann.Path).Msg("invalid annotation: cannot annotate ancestor path")
			continue
		}

		// Validate that the target path exists in the filesystem
		if pathExists != nil && !pathExists(targetPath) {
			logging.Warn().Str("info_file", ann.InfoFile).Str("path", ann.Path).Msg("invalid annotation: path does not exist")
			continue
		}

		contenders[targetPath] = append(contenders[targetPath], ann)
	}

	result := make(map[string]Annotation)

	for targetPath, anns := range contenders {
		if len(anns) == 1 {
			result[targetPath] = anns[0]
			continue
		}

		// Multiple annotations for the same target path - apply precedence rules
		winner := g.selectWinner(anns)
		result[targetPath] = winner
	}

	return result
}

// selectWinner chooses the winning annotation based on precedence rules
func (g *Gatherer) selectWinner(annotations []Annotation) Annotation {
	if len(annotations) == 1 {
		return annotations[0]
	}

	// Sort by precedence rules (matching original merger.go logic):
	// 1. Deepest .info file wins (highest directory depth)
	// 2. Lexicographical order of directory paths
	// 3. Line number within file
	sort.Slice(annotations, func(i, j int) bool {
		dirI := filepath.Dir(annotations[i].InfoFile)
		dirJ := filepath.Dir(annotations[j].InfoFile)

		// Rule: closest (deepest) .info file wins.
		// Calculate depth correctly, handling "." as root (depth 0)
		depthI := g.pathDepth(dirI)
		depthJ := g.pathDepth(dirJ)

		if depthI != depthJ {
			return depthI > depthJ // Deeper path wins
		}

		// Rule: if distance is same, lexicographical order of .info file dir wins.
		if dirI != dirJ {
			return dirI < dirJ
		}

		// Rule: if same .info file, lower line number wins.
		return annotations[i].LineNum < annotations[j].LineNum
	})

	return annotations[0]
}

// pathDepth calculates the depth of a directory path, with "." being depth 0
func (g *Gatherer) pathDepth(dir string) int {
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

// ============================================================================
// VALIDATOR TYPES (for backward compatibility)
// ============================================================================

// Validator provides validation functionality for .info files
// Deprecated: Use InfoAPI.Validate() for new code which uses InfoFile-based validation
type Validator struct {
	parser *Parser
}

// NewInfoValidator creates a new info file validator
// Deprecated: Use InfoAPI.Validate() for new code which uses InfoFile-based validation
func NewInfoValidator() *Validator {
	return &Validator{
		parser: NewParser(),
	}
}
