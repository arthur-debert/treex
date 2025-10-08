package info

// This file only contains validation types now - no imports needed

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
