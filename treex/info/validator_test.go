package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateContent(t *testing.T) {
	tests := []struct {
		name               string
		infoFiles          InfoFileMap
		existingPaths      map[string]bool
		expectedIssueCount int
		expectedIssueTypes []ValidationIssueType
		expectedValidFiles int
	}{
		{
			name: "valid .info files",
			infoFiles: InfoFileMap{
				".info":     "a.txt  Valid annotation for a",
				"sub/.info": "local.txt  Valid annotation for local",
			},
			existingPaths: map[string]bool{
				"a.txt":         true,
				"sub/local.txt": true,
			},
			expectedIssueCount: 0,
			expectedIssueTypes: []ValidationIssueType{},
			expectedValidFiles: 2,
		},
		{
			name: "invalid format and duplicates",
			infoFiles: InfoFileMap{
				".info": `
a.txt  Valid annotation
invalid_line_no_space
b.txt  Another valid annotation
a.txt  Duplicate annotation
c.txt
`,
			},
			existingPaths: map[string]bool{
				"a.txt": true,
				"b.txt": true,
			},
			expectedIssueCount: 3, // invalid format, duplicate, missing annotation
			expectedIssueTypes: []ValidationIssueType{
				IssueInvalidFormat,
				IssueDuplicatePath,
				IssueInvalidFormat,
			},
			expectedValidFiles: 0,
		},
		{
			name: "path validation issues",
			infoFiles: InfoFileMap{
				".info": `
existing.txt  Valid annotation
missing.txt   Annotation for missing file
`,
				"sub/.info": `
../existing.txt  Valid parent annotation
../..            Invalid ancestor annotation
`,
			},
			existingPaths: map[string]bool{
				"existing.txt": true,
			},
			expectedIssueCount: 4, // missing file + missing ancestor file + ancestor path + cross-file conflict
			expectedIssueTypes: []ValidationIssueType{
				IssuePathNotExists,
				IssuePathNotExists, // Both missing.txt and ../.. don't exist
				IssueAncestorPath,
				IssueMultipleFiles,
			},
			expectedValidFiles: 0,
		},
		{
			name: "cross-file conflicts",
			infoFiles: InfoFileMap{
				".info":     "target.txt  Root annotation",
				"sub/.info": "../target.txt  Sub annotation (should win)",
			},
			existingPaths: map[string]bool{
				"target.txt": true,
			},
			expectedIssueCount: 1, // multiple files conflict
			expectedIssueTypes: []ValidationIssueType{
				IssueMultipleFiles,
			},
			expectedValidFiles: 1, // sub/.info remains valid (wins), .info becomes invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewInfoValidator()

			// Create pathExists function from test data
			pathExists := func(path string) bool {
				return tt.existingPaths[path]
			}

			result, err := validator.ValidateContent(tt.infoFiles, pathExists)

			require.NoError(t, err)
			require.NotNil(t, result)

			// Check issue count
			assert.Len(t, result.Issues, tt.expectedIssueCount, "Issue count mismatch")

			// Check issue types
			actualTypes := make([]ValidationIssueType, len(result.Issues))
			for i, issue := range result.Issues {
				actualTypes[i] = issue.Type
			}
			for _, expectedType := range tt.expectedIssueTypes {
				assert.Contains(t, actualTypes, expectedType, "Expected issue type %s not found", expectedType)
			}

			// Check valid files count
			assert.Len(t, result.ValidFiles, tt.expectedValidFiles, "Valid files count mismatch")

			// Verify summary consistency
			assert.Equal(t, len(result.Issues), result.Summary.TotalIssues)
			assert.Equal(t, len(result.ValidFiles)+len(result.InvalidFiles), result.Summary.TotalFiles)

			// Check that issue counts by type match
			for issueType, count := range result.Summary.IssuesByType {
				actualCount := 0
				for _, issue := range result.Issues {
					if issue.Type == issueType {
						actualCount++
					}
				}
				assert.Equal(t, actualCount, count, "Issue count mismatch for type %s", issueType)
			}
		})
	}
}

func TestValidator_parseFileWithValidation(t *testing.T) {
	tests := []struct {
		name                string
		content             string
		expectedAnnotations int
		expectedIssues      int
		expectedTypes       []ValidationIssueType
	}{
		{
			name: "valid content",
			content: `
# This is a comment
a.txt  Annotation for a
b.txt  Annotation for b
`,
			expectedAnnotations: 2,
			expectedIssues:      0,
			expectedTypes:       []ValidationIssueType{},
		},
		{
			name: "invalid format",
			content: `
valid.txt  Valid annotation
invalid_line
another_invalid_line
path_without_annotation
`,
			expectedAnnotations: 1,
			expectedIssues:      3,
			expectedTypes: []ValidationIssueType{
				IssueInvalidFormat,
				IssueInvalidFormat,
				IssueInvalidFormat,
			},
		},
		{
			name: "duplicates",
			content: `
file.txt  First annotation
other.txt  Other annotation
file.txt  Duplicate annotation
file.txt  Another duplicate
`,
			expectedAnnotations: 2, // Only first occurrence of file.txt + other.txt
			expectedIssues:      2, // Two duplicates
			expectedTypes: []ValidationIssueType{
				IssueDuplicatePath,
				IssueDuplicatePath,
			},
		},
		{
			name: "escaped spaces",
			content: `
path\ with\ spaces.txt  Annotation for spaced path
normal.txt              Normal annotation
`,
			expectedAnnotations: 2,
			expectedIssues:      0,
			expectedTypes:       []ValidationIssueType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewInfoValidator()

			annotations, issues := validator.parseFileWithValidation(tt.content, "test.info")

			assert.Len(t, annotations, tt.expectedAnnotations, "Annotation count mismatch")
			assert.Len(t, issues, tt.expectedIssues, "Issue count mismatch")

			// Check issue types
			actualTypes := make([]ValidationIssueType, len(issues))
			for i, issue := range issues {
				actualTypes[i] = issue.Type
			}
			for _, expectedType := range tt.expectedTypes {
				assert.Contains(t, actualTypes, expectedType, "Expected issue type %s not found", expectedType)
			}

			// Verify annotations have proper structure
			for _, annotation := range annotations {
				assert.NotEmpty(t, annotation.Path, "Annotation path should not be empty")
				assert.NotEmpty(t, annotation.Annotation, "Annotation text should not be empty")
				assert.Equal(t, "test.info", annotation.InfoFile)
				assert.Greater(t, annotation.LineNum, 0, "Line number should be positive")
			}
		})
	}
}

func TestValidator_findCrossFileConflicts(t *testing.T) {
	// Create annotations from different .info files that conflict
	annotations := []Annotation{
		{
			Path:       "target.txt",
			Annotation: "Root annotation",
			InfoFile:   ".info",
			LineNum:    1,
		},
		{
			Path:       "../target.txt",
			Annotation: "Sub annotation",
			InfoFile:   "sub/.info",
			LineNum:    1,
		},
		{
			Path:       "local.txt",
			Annotation: "Local annotation",
			InfoFile:   "sub/.info",
			LineNum:    2,
		},
		{
			Path:       "../target.txt",
			Annotation: "Deep annotation",
			InfoFile:   "sub/deep/.info",
			LineNum:    1,
		},
	}

	validator := NewInfoValidator()
	issues := validator.findCrossFileConflicts(annotations)

	// Should find 1 conflict for target.txt:
	// - .info loses to sub/.info
	// Note: sub/.info vs sub/deep/.info conflict is not happening because both point to different resolved paths
	assert.Len(t, issues, 1, "Should find 1 cross-file conflict")

	// Check that all conflicts are of the right type
	for _, issue := range issues {
		assert.Equal(t, IssueMultipleFiles, issue.Type)
		assert.NotEmpty(t, issue.RelatedFile, "Related file should be specified")
		assert.NotEmpty(t, issue.Message, "Message should not be empty")
	}

	// Check that the deeper files are identified as winners
	conflictFiles := make([]string, 0)
	for _, issue := range issues {
		conflictFiles = append(conflictFiles, issue.InfoFile)
	}

	// .info should lose to deeper files
	assert.Contains(t, conflictFiles, ".info")
}

func TestValidator_ValidationSummary(t *testing.T) {
	infoFiles := InfoFileMap{
		".info": `
valid.txt    Valid annotation
invalid_line
missing.txt  Missing file annotation
duplicate.txt First annotation
duplicate.txt Duplicate annotation
`,
		"sub/.info": `
../valid.txt  Conflicting annotation
local.txt     Local annotation
`,
	}

	existingPaths := map[string]bool{
		"valid.txt":     true,
		"sub/local.txt": true,
	}
	pathExists := func(path string) bool {
		return existingPaths[path]
	}

	validator := NewInfoValidator()
	result, err := validator.ValidateContent(infoFiles, pathExists)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify summary statistics
	summary := result.Summary
	assert.Equal(t, 2, summary.TotalFiles)
	assert.Greater(t, summary.TotalIssues, 0)
	assert.Greater(t, summary.TotalAnnotations, 0)
	assert.GreaterOrEqual(t, summary.ValidAnnotations, 0)
	assert.LessOrEqual(t, summary.ValidAnnotations, summary.TotalAnnotations)

	// Verify issue type counts
	totalTypeCount := 0
	for _, count := range summary.IssuesByType {
		totalTypeCount += count
		assert.Greater(t, count, 0, "Issue type count should be positive")
	}
	assert.Equal(t, summary.TotalIssues, totalTypeCount, "Issue type counts should sum to total")

	// Verify file issue counts
	totalFileCount := 0
	for filename, count := range summary.IssuesByFile {
		totalFileCount += count
		if count > 0 { // Only check files that actually have issues
			assert.Greater(t, count, 0, "File issue count should be positive")
			assert.Contains(t, []string{".info", "sub/.info"}, filename, "Issue file should be a known .info file")
		}
	}
	assert.Equal(t, summary.TotalIssues, totalFileCount, "File issue counts should sum to total")
}
