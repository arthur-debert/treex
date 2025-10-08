package info

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfoFileSet_Pure tests InfoFileSet operations as pure functions without filesystem
func TestInfoFileSet_Pure(t *testing.T) {
	t.Run("Gather consolidates annotations correctly", func(t *testing.T) {
		// Create InfoFiles manually (no filesystem)
		infoFile1 := NewInfoFile(".info", "a.txt  Root annotation\nb.txt  Another root annotation")
		infoFile2 := NewInfoFile("sub/.info", "local.txt  Local annotation\n../a.txt  Conflicting annotation")

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

		// Root .info file should win for a.txt (precedence rules)
		assert.Equal(t, "Root annotation", gathered["a.txt"].Annotation)
		assert.Equal(t, ".info", gathered["a.txt"].InfoFile)
	})

	t.Run("Validate detects cross-file conflicts", func(t *testing.T) {
		infoFile1 := NewInfoFile(".info", "target.txt  Root annotation")
		infoFile2 := NewInfoFile("sub/.info", "../target.txt  Sub annotation")

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

	t.Run("Distribute moves annotations to optimal files", func(t *testing.T) {
		// Create scenario where annotation in root should move to sub directory
		rootFile := NewInfoFile(".info", "sub/target.txt  Root annotation")
		subFile := NewInfoFile("sub/.info", "local.txt  Local annotation")

		pathExists := func(path string) bool { return true }

		infoFileSet := NewInfoFileSet([]*InfoFile{rootFile, subFile}, pathExists)
		distributed := infoFileSet.Distribute()

		// Verify annotation moved to closer file
		files := distributed.GetFiles()
		var distributedRootFile, distributedSubFile *InfoFile
		for _, file := range files {
			switch file.Path {
			case ".info":
				distributedRootFile = file
			case "sub/.info":
				distributedSubFile = file
			}
		}

		require.NotNil(t, distributedRootFile, "Root file should exist")
		require.NotNil(t, distributedSubFile, "Sub file should exist")

		rootAnnotations := distributedRootFile.GetAllAnnotations()
		subAnnotations := distributedSubFile.GetAllAnnotations()

		// Root file should be empty after distribution (sub/target.txt moved to sub/.info)
		assert.Len(t, rootAnnotations, 0, "Root file should be empty after distribution")
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

// TestInfoFileSet_WithFilesystem tests filesystem I/O boundaries
func TestInfoFileSet_WithFilesystem(t *testing.T) {
	t.Run("LoadFromPath reads all .info files", func(t *testing.T) {
		// Create in-memory filesystem
		fs := afero.NewMemMapFs()

		// Create test .info files
		err := afero.WriteFile(fs, ".info", []byte("a.txt  Root annotation"), 0644)
		require.NoError(t, err)

		err = afero.WriteFile(fs, "sub/.info", []byte("local.txt  Sub annotation"), 0644)
		require.NoError(t, err)

		// Load via InfoFileSetLoader
		afs := NewAferoInfoFileSystem(fs)
		loader := NewInfoFileSetLoader(afs)
		infoFileSet, err := loader.LoadFromPath(".")

		require.NoError(t, err)
		assert.Equal(t, 2, infoFileSet.Count())

		paths := infoFileSet.GetFilePaths()
		assert.Contains(t, paths, ".info")
		assert.Contains(t, paths, "sub/.info")
	})

	t.Run("WriteInfoFileSet saves and deletes correctly", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		afs := NewAferoInfoFileSystem(fs)

		// Create InfoFileSet with one empty and one non-empty file
		emptyFile := NewInfoFile("empty/.info", "")
		nonEmptyFile := NewInfoFile("full/.info", "file.txt  Annotation")

		pathExists := func(path string) bool { return true }
		infoFileSet := NewInfoFileSet([]*InfoFile{emptyFile, nonEmptyFile}, pathExists)

		// Write to filesystem
		writer := NewInfoFileSetWriter(afs)
		err := writer.WriteInfoFileSet(infoFileSet)
		require.NoError(t, err)

		// Verify non-empty file was written
		content, err := afero.ReadFile(fs, "full/.info")
		require.NoError(t, err)
		assert.Contains(t, string(content), "file.txt  Annotation")

		// Verify empty file was "deleted" (written as empty)
		content, err = afero.ReadFile(fs, "empty/.info")
		require.NoError(t, err)
		assert.Empty(t, string(content))
	})

	t.Run("Round-trip preserves data correctly", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		afs := NewAferoInfoFileSystem(fs)

		// Create original .info file
		originalContent := `# Comment line
a.txt  First annotation
b.txt  Second annotation
# Another comment`

		err := afero.WriteFile(fs, "test/.info", []byte(originalContent), 0644)
		require.NoError(t, err)

		// Load -> Process -> Save
		loader := NewInfoFileSetLoader(afs)
		infoFileSet, err := loader.LoadFromPath("test")
		require.NoError(t, err)

		// Perform some operation (gather and filter)
		gathered := infoFileSet.Gather()
		assert.Len(t, gathered, 2)

		filtered := infoFileSet.RemoveEmpty()

		// Write back
		writer := NewInfoFileSetWriter(afs)
		err = writer.WriteInfoFileSet(filtered)
		require.NoError(t, err)

		// Verify content preserved
		savedContent, err := afero.ReadFile(fs, "test/.info")
		require.NoError(t, err)

		savedStr := string(savedContent)
		assert.Contains(t, savedStr, "a.txt  First annotation")
		assert.Contains(t, savedStr, "b.txt  Second annotation")
		assert.Contains(t, savedStr, "# Comment line")
	})
}

// TestInfoFileSet_MethodChaining tests fluent interface
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
		assert.Equal(t, 0, result.Summary.TotalFiles)
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

	t.Run("Complex precedence rules", func(t *testing.T) {
		// Test complex precedence scenario with multiple levels
		files := []*InfoFile{
			NewInfoFile(".info", "target/file.txt  Root level"),
			NewInfoFile("target/.info", "file.txt  Target level"),
			NewInfoFile("target/sub/.info", "../file.txt  Sub level"),
		}

		pathExists := func(path string) bool { return true }
		infoFileSet := NewInfoFileSet(files, pathExists)

		gathered := infoFileSet.Gather()

		// target/.info should win (deepest directory for the target path)
		winner := gathered["target/file.txt"]
		assert.Equal(t, "Target level", winner.Annotation)
		assert.Equal(t, "target/.info", winner.InfoFile)
	})
}
