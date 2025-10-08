package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInfoFileSet_Validation tests validation operations on InfoFileSet instances
// Extracted from the working infofileset_test.go TestInfoFileSet_Pure tests

func TestInfoFileSet_Validation(t *testing.T) {
	t.Run("Validate detects cross-file conflicts", func(t *testing.T) {
		infoFile1 := NewInfoFile(".info", "sub/target.txt  Root annotation")
		infoFile2 := NewInfoFile("sub/.info", "target.txt  Sub annotation")

		pathExists := func(path string) bool { return true }

		infoFileSet := NewInfoFileSet([]*InfoFile{infoFile1, infoFile2}, pathExists)
		result := infoFileSet.Validate()

		// Should detect cross-file conflict
		assert.True(t, len(result.Issues) > 0)
		assert.True(t, infoFileSet.HasConflicts())

		// Find the cross-file conflict issue
		var foundConflict bool
		for _, issue := range result.Issues {
			if issue.Type == IssueMultipleFiles {
				foundConflict = true
				assert.Equal(t, "sub/.info", issue.InfoFile) // Later file should be flagged
				assert.Equal(t, ".info", issue.RelatedFile)  // Earlier file should be related
			}
		}
		assert.True(t, foundConflict, "Should find cross-file conflict")
	})
}

// TestInfoFileSet_Clean tests cleaning operations on InfoFileSet instances
// Extracted from the working infofileset_test.go TestInfoFileSet_Pure tests

func TestInfoFileSet_Clean(t *testing.T) {
	t.Run("Clean removes problematic annotations", func(t *testing.T) {
		// Create InfoFile with problematic annotations
		infoFile := NewInfoFile(".info", `
valid.txt     Valid annotation
missing.txt   Annotation for missing file
invalid_line
duplicate.txt First annotation
duplicate.txt Duplicate annotation
`)

		pathExists := func(path string) bool {
			return path == "valid.txt" // Only valid.txt exists
		}

		infoFileSet := NewInfoFileSet([]*InfoFile{infoFile}, pathExists)
		cleanResult, cleanedSet := infoFileSet.Clean()

		// Verify clean results
		assert.True(t, cleanResult.Summary.FilesModified > 0)
		assert.True(t, cleanResult.Summary.InvalidPathsRemoved > 0)
		assert.Len(t, cleanResult.RemovedAnnotations, 2) // missing.txt + duplicate

		// Verify cleaned set only has valid annotation
		cleanedAnnotations := cleanedSet.GetAllAnnotations()
		assert.Len(t, cleanedAnnotations, 1)
		assert.Equal(t, "valid.txt", cleanedAnnotations[0].Path)
	})
}
