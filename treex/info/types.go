package info

type ValidationIssueType string

const (
	IssueInvalidFormat ValidationIssueType = "invalid_format"
	IssueDuplicatePath ValidationIssueType = "duplicate_path"
	IssuePathNotExists ValidationIssueType = "path_not_exists"
	IssueAncestorPath  ValidationIssueType = "ancestor_path"
	IssueMultipleFiles ValidationIssueType = "multiple_files"
)

type ValidationIssue struct {
	Type        ValidationIssueType `json:"type"`
	InfoFile    string              `json:"info_file"`
	LineNum     int                 `json:"line_num"`
	Path        string              `json:"path"`
	Message     string              `json:"message"`
	Suggestion  string              `json:"suggestion,omitempty"`
	RelatedFile string              `json:"related_file,omitempty"` // For multiple_files issues
}

type ValidationResult struct {
	Issues       []ValidationIssue      `json:"issues"`
	ValidFiles   []string               `json:"valid_files"`
	InvalidFiles []string               `json:"invalid_files"`
	Summary      map[string]interface{} `json:"summary"`
}

func (r *ValidationResult) GetIssues(issueType ValidationIssueType) []ValidationIssue {
	var issues []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Type == issueType {
			issues = append(issues, issue)
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
