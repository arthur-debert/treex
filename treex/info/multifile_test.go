package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfoFileSet_Gather tests gathering annotations from multiple InfoFiles
// Extracted from the working infofileset_test.go TestInfoFileSet_Pure tests

func TestInfoFileSet_Gather(t *testing.T) {
	t.Run("Gather consolidates annotations correctly", func(t *testing.T) {
		// Create InfoFiles manually (no filesystem)
		infoFile1 := NewInfoFile(".info", "a.txt  Root annotation\nb.txt  Another root annotation")
		infoFile2 := NewInfoFile("sub/.info", "local.txt  Local annotation")

		// Mock pathExists (pure function)
		pathExists := func(path string) bool {
			validPaths := map[string]bool{
				"a.txt":         true,
				"b.txt":         true,
				"sub/local.txt": true,
			}
			return validPaths[path]
		}

		// Create InfoFileSet and test Gather (pure operation)
		infoFileSet := NewInfoFileSet([]*InfoFile{infoFile1, infoFile2}, pathExists)
		gathered := infoFileSet.Gather()

		// Verify results
		assert.Len(t, gathered, 3)
		assert.Contains(t, gathered, "a.txt")
		assert.Contains(t, gathered, "b.txt")
		assert.Contains(t, gathered, "sub/local.txt")

		// Root .info file annotation for a.txt (only annotation)
		assert.Equal(t, "Root annotation", gathered["a.txt"].Annotation)
		assert.Equal(t, ".info", gathered["a.txt"].InfoFile)
	})

	t.Run("Complex precedence rules", func(t *testing.T) {
		// Test complex precedence scenario with multiple levels
		files := []*InfoFile{
			NewInfoFile(".info", "rootfile.txt  Root level"),
			NewInfoFile("target/.info", "file.txt  Target level"),
			NewInfoFile("target/sub/.info", "subfile.txt  Sub level"),
		}

		pathExists := func(path string) bool { return true }
		infoFileSet := NewInfoFileSet(files, pathExists)

		gathered := infoFileSet.Gather()

		// Each annotation should be present since they don't conflict
		assert.Len(t, gathered, 3)

		// Root level annotation
		rootWinner := gathered["rootfile.txt"]
		assert.Equal(t, "Root level", rootWinner.Annotation)
		assert.Equal(t, ".info", rootWinner.InfoFile)

		// Target level annotation
		targetWinner := gathered["target/file.txt"]
		assert.Equal(t, "Target level", targetWinner.Annotation)
		assert.Equal(t, "target/.info", targetWinner.InfoFile)

		// Sub level annotation
		subWinner := gathered["target/sub/subfile.txt"]
		assert.Equal(t, "Sub level", subWinner.Annotation)
		assert.Equal(t, "target/sub/.info", subWinner.InfoFile)
	})
}

// TestInfoFileSet_Distribute tests distributing annotations to optimal InfoFiles
// Extracted from the working infofileset_test.go TestInfoFileSet_Pure tests

func TestInfoFileSet_Distribute(t *testing.T) {
	t.Run("Distribute moves annotations to optimal files", func(t *testing.T) {
		// Create scenario where annotation in root should move to sub directory
		rootFile := NewInfoFile(".info", "sub/target.txt  Root annotation")
		subFile := NewInfoFile("sub/.info", "local.txt  Local annotation")

		pathExists := func(path string) bool { return true }

		infoFileSet := NewInfoFileSet([]*InfoFile{rootFile, subFile}, pathExists)
		distributed := infoFileSet.Distribute()

		// Verify annotation moved to closer file and empty files are removed
		files := distributed.GetFiles()
		var distributedSubFile *InfoFile
		for _, file := range files {
			if file.Path == "sub/.info" {
				distributedSubFile = file
			}
		}

		// Root file should be removed entirely after distribution (was empty)
		assert.Len(t, files, 1, "Only sub/.info should remain after distribution")
		require.NotNil(t, distributedSubFile, "Sub file should exist")

		subAnnotations := distributedSubFile.GetAllAnnotations()
		assert.Len(t, subAnnotations, 2, "Sub file should have local.txt + target.txt")

		// Verify the specific annotations
		var hasLocal, hasTarget bool
		for _, ann := range subAnnotations {
			if ann.Path == "local.txt" {
				hasLocal = true
			}
			if ann.Path == "target.txt" { // Should be relative to sub/.info
				hasTarget = true
			}
		}
		assert.True(t, hasLocal, "Should have local.txt annotation")
		assert.True(t, hasTarget, "Should have target.txt annotation (moved from root)")
	})
}

// TestInfoFileSet_UtilityOperations tests filtering and utility operations
// Extracted from the working infofileset_test.go TestInfoFileSet_Pure tests

func TestInfoFileSet_UtilityOperations(t *testing.T) {
	t.Run("Filter and RemoveEmpty work correctly", func(t *testing.T) {
		emptyFile := NewInfoFile("empty/.info", "")
		nonEmptyFile := NewInfoFile("full/.info", "file.txt  Annotation")

		pathExists := func(path string) bool { return true }

		infoFileSet := NewInfoFileSet([]*InfoFile{emptyFile, nonEmptyFile}, pathExists)

		// Test Filter
		filtered := infoFileSet.Filter(func(file *InfoFile) bool {
			return file.Path == "full/.info"
		})
		assert.Equal(t, 1, filtered.Count())
		assert.Equal(t, "full/.info", filtered.GetFiles()[0].Path)

		// Test RemoveEmpty
		nonEmpty := infoFileSet.RemoveEmpty()
		assert.Equal(t, 1, nonEmpty.Count())
		assert.Equal(t, "full/.info", nonEmpty.GetFiles()[0].Path)

		// Test GetEmptyFiles
		emptyFiles := infoFileSet.GetEmptyFiles()
		assert.Len(t, emptyFiles, 1)
		assert.Equal(t, "empty/.info", emptyFiles[0].Path)
	})
}

// TestInfoFileSet_MethodChaining tests fluent interface for multiple operations
// Extracted from the working infofileset_test.go TestInfoFileSet_MethodChaining tests

func TestInfoFileSet_MethodChaining(t *testing.T) {
	t.Run("Chain operations fluently", func(t *testing.T) {
		// Create problematic InfoFiles
		emptyFile := NewInfoFile("empty/.info", "")
		invalidFile := NewInfoFile("invalid/.info", "missing.txt  Annotation for missing file")
		validFile := NewInfoFile("valid/.info", "exists.txt  Valid annotation")

		pathExists := func(path string) bool {
			// Only valid/exists.txt exists (resolved paths)
			// invalid/missing.txt does not exist
			return path == "valid/exists.txt"
		}

		infoFileSet := NewInfoFileSet([]*InfoFile{emptyFile, invalidFile, validFile}, pathExists)

		// Chain operations: Clean -> RemoveEmpty -> Filter
		_, cleanedSet := infoFileSet.Clean()
		result := cleanedSet.
			RemoveEmpty().
			Filter(func(file *InfoFile) bool {
				return len(file.GetAllAnnotations()) > 0
			})

		// Should only have the valid file with valid annotations
		assert.Equal(t, 1, result.Count())
		annotations := result.GetAllAnnotations()
		assert.Len(t, annotations, 1)
		assert.Equal(t, "exists.txt", annotations[0].Path)
	})

	t.Run("WithPathValidator changes validation behavior", func(t *testing.T) {
		infoFile := NewInfoFile(".info", "missing.txt  Annotation")

		// Initially, missing.txt doesn't exist
		pathExists1 := func(path string) bool { return false }
		infoFileSet := NewInfoFileSet([]*InfoFile{infoFile}, pathExists1)

		result1 := infoFileSet.Validate()
		assert.True(t, len(result1.Issues) > 0) // Should have issues

		// Change validator to make missing.txt exist
		pathExists2 := func(path string) bool { return true }
		newSet := infoFileSet.WithPathValidator(pathExists2)

		result2 := newSet.Validate()
		assert.True(t, len(result2.Issues) == 0) // Should have no issues now
	})
}

// TestInfoFileSet_EdgeCases tests edge cases and error conditions
// Extracted from the working infofileset_test.go TestInfoFileSet_EdgeCases tests

func TestInfoFileSet_EdgeCases(t *testing.T) {
	t.Run("Empty InfoFileSet operations", func(t *testing.T) {
		pathExists := func(path string) bool { return true }
		emptySet := EmptyInfoFileSet(pathExists)

		assert.Equal(t, 0, emptySet.Count())
		assert.Equal(t, 0, emptySet.GetAnnotationCount())
		assert.False(t, emptySet.HasConflicts())

		gathered := emptySet.Gather()
		assert.Len(t, gathered, 0)

		distributed := emptySet.Distribute()
		assert.Equal(t, 0, distributed.Count())

		result := emptySet.Validate()
		assert.Len(t, result.Issues, 0)
		assert.Equal(t, 0, result.Summary["total_files"])
	})

	t.Run("InfoFileSet with nil pathExists", func(t *testing.T) {
		infoFile := NewInfoFile(".info", "test.txt  Test annotation")
		infoFileSet := NewInfoFileSet([]*InfoFile{infoFile}, nil)

		// Should work without path validation
		gathered := infoFileSet.Gather()
		assert.Len(t, gathered, 1)

		// Validation should skip path existence checks
		result := infoFileSet.Validate()
		hasPathNotExistsIssues := false
		for _, issue := range result.Issues {
			if issue.Type == IssuePathNotExists {
				hasPathNotExistsIssues = true
				break
			}
		}
		assert.False(t, hasPathNotExistsIssues, "Should not check path existence with nil pathExists")
	})
}
