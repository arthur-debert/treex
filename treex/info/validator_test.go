package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: Most validator tests have been removed because the Validator struct methods
// are now deprecated and replaced by InfoAPI validation using InfoFile structures.
// See api_test.go TestInfoAPI_Validate for tests of the new validation approach.

func TestValidationIssueTypes(t *testing.T) {
	// Test that validation issue type constants are properly defined
	assert.Equal(t, ValidationIssueType("invalid_format"), IssueInvalidFormat)
	assert.Equal(t, ValidationIssueType("duplicate_path"), IssueDuplicatePath)
	assert.Equal(t, ValidationIssueType("path_not_exists"), IssuePathNotExists)
	assert.Equal(t, ValidationIssueType("ancestor_path"), IssueAncestorPath)
	assert.Equal(t, ValidationIssueType("multiple_files"), IssueMultipleFiles)
}
