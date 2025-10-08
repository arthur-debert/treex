package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInfoFile_EditOperations tests Add/Remove/Update operations on InfoFile instances
// These tests are moved from info_test.go and focus on pure InfoFile editing operations

func TestInfoFile_AddAnnotationForPath(t *testing.T) {
	content := `file1.txt Existing annotation`

	infoFile := NewInfoFile("/test/.info", content)

	// Add new annotation
	result := infoFile.AddAnnotationForPath("file2.txt", "New annotation")
	assert.True(t, result)

	assert.True(t, infoFile.HasAnnotationForPath("file2.txt"))
	ann := infoFile.GetAnnotationForPath("file2.txt")
	assert.Equal(t, "New annotation", ann.Annotation)

	// Check that a new line was added
	assert.Len(t, infoFile.Lines, 2)
	assert.Equal(t, "file2.txt New annotation", infoFile.Lines[1].Raw)

	// Try to add duplicate annotation
	result = infoFile.AddAnnotationForPath("file1.txt", "Duplicate annotation")
	assert.False(t, result)
}

func TestInfoFile_UpdateAnnotationForPath(t *testing.T) {
	content := `file1.txt Original annotation
file2.txt Another annotation`

	infoFile := NewInfoFile("/test/.info", content)

	// Update existing annotation
	result := infoFile.UpdateAnnotationForPath("file1.txt", "Updated annotation")
	assert.True(t, result)

	ann := infoFile.GetAnnotationForPath("file1.txt")
	assert.Equal(t, "Updated annotation", ann.Annotation)

	// Check that the line was updated
	assert.Equal(t, "file1.txt Updated annotation", infoFile.Lines[0].Raw)

	// Try to update non-existent annotation
	result = infoFile.UpdateAnnotationForPath("nonexistent.txt", "New annotation")
	assert.False(t, result)
}

func TestInfoFile_RemoveAnnotationForPath(t *testing.T) {
	content := `file1.txt Annotation 1
file2.txt Annotation 2
file3.txt Annotation 3`

	infoFile := NewInfoFile("/test/.info", content)

	// Remove existing annotation
	result := infoFile.RemoveAnnotationForPath("file2.txt")
	assert.True(t, result)

	assert.False(t, infoFile.HasAnnotationForPath("file2.txt"))
	assert.Len(t, infoFile.annotations, 2)

	// Check that the line was marked as removed
	assert.Equal(t, LineTypeMalformed, infoFile.Lines[1].Type)
	assert.Equal(t, "removed", infoFile.Lines[1].ParseError)

	// Try to remove non-existent annotation
	result = infoFile.RemoveAnnotationForPath("nonexistent.txt")
	assert.False(t, result)
}

func TestInfoFile_DuplicatePaths(t *testing.T) {
	content := `file1.txt First annotation
file1.txt Second annotation
file2.txt Valid annotation`

	infoFile := NewInfoFile("/test/.info", content)

	// Should only have 2 annotations (first occurrence wins)
	assert.Len(t, infoFile.annotations, 2)
	assert.True(t, infoFile.HasAnnotationForPath("file1.txt"))
	assert.True(t, infoFile.HasAnnotationForPath("file2.txt"))

	// First annotation should win
	ann := infoFile.GetAnnotationForPath("file1.txt")
	assert.Equal(t, "First annotation", ann.Annotation)

	// Second line should be marked as malformed
	assert.Equal(t, LineTypeMalformed, infoFile.Lines[1].Type)
	assert.Equal(t, "duplicate path (first occurrence wins)", infoFile.Lines[1].ParseError)
}
