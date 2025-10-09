package infofile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfoFile_Creation tests InfoFile creation and basic parsing
// Core InfoFile functionality tests that don't duplicate edit operations

func TestInfoFile_NewInfoFile(t *testing.T) {
	content := `# This is a comment
file1.txt This is annotation 1
file2.txt This is annotation 2

# Another comment
nested/file3.txt This is annotation 3`

	infoFile := NewInfoFile("/test/.info", content)

	assert.Equal(t, "/test/.info", infoFile.Path)
	assert.Len(t, infoFile.Lines, 6)
	assert.Len(t, infoFile.annotations, 3)

	// Check line types
	assert.Equal(t, LineTypeComment, infoFile.Lines[0].Type)
	assert.Equal(t, LineTypeAnnotation, infoFile.Lines[1].Type)
	assert.Equal(t, LineTypeAnnotation, infoFile.Lines[2].Type)
	assert.Equal(t, LineTypeBlank, infoFile.Lines[3].Type)
	assert.Equal(t, LineTypeComment, infoFile.Lines[4].Type)
	assert.Equal(t, LineTypeAnnotation, infoFile.Lines[5].Type)

	// Check annotations
	assert.True(t, infoFile.HasAnnotationForPath("file1.txt"))
	assert.True(t, infoFile.HasAnnotationForPath("file2.txt"))
	assert.True(t, infoFile.HasAnnotationForPath("nested/file3.txt"))
	assert.False(t, infoFile.HasAnnotationForPath("nonexistent.txt"))
}

func TestInfoFile_HasAnnotationForPath(t *testing.T) {
	content := `file1.txt Annotation 1
file2.txt Annotation 2`

	infoFile := NewInfoFile("/test/.info", content)

	assert.True(t, infoFile.HasAnnotationForPath("file1.txt"))
	assert.True(t, infoFile.HasAnnotationForPath("file2.txt"))
	assert.False(t, infoFile.HasAnnotationForPath("file3.txt"))
}

func TestInfoFile_GetAnnotationForPath(t *testing.T) {
	content := `file1.txt Annotation 1
file2.txt Annotation 2`

	infoFile := NewInfoFile("/test/.info", content)

	ann1 := infoFile.GetAnnotationForPath("file1.txt")
	require.NotNil(t, ann1)
	assert.Equal(t, "file1.txt", ann1.Path)
	assert.Equal(t, "Annotation 1", ann1.Annotation)
	assert.Equal(t, "/test/.info", ann1.InfoFile)
	assert.Equal(t, 1, ann1.LineNum)

	ann2 := infoFile.GetAnnotationForPath("file2.txt")
	require.NotNil(t, ann2)
	assert.Equal(t, "file2.txt", ann2.Path)
	assert.Equal(t, "Annotation 2", ann2.Annotation)

	ann3 := infoFile.GetAnnotationForPath("nonexistent.txt")
	assert.Nil(t, ann3)
}

func TestInfoFile_GetAllAnnotations(t *testing.T) {
	content := `file1.txt Annotation 1
file2.txt Annotation 2
file3.txt Annotation 3`

	infoFile := NewInfoFile("/test/.info", content)

	annotations := infoFile.GetAllAnnotations()
	assert.Len(t, annotations, 3)

	// Check that all paths are present
	paths := make(map[string]bool)
	for _, ann := range annotations {
		paths[ann.Path] = true
	}
	assert.True(t, paths["file1.txt"])
	assert.True(t, paths["file2.txt"])
	assert.True(t, paths["file3.txt"])
}

func TestInfoFile_IsEmpty(t *testing.T) {
	// Empty file
	emptyFile := NewInfoFile("/test/.info", "")
	assert.True(t, emptyFile.IsEmpty())

	// File with only comments and blank lines
	commentOnlyFile := NewInfoFile("/test/.info", `# Comment
	
# Another comment`)
	assert.True(t, commentOnlyFile.IsEmpty())

	// File with annotations
	fileWithAnnotations := NewInfoFile("/test/.info", `file.txt Annotation`)
	assert.False(t, fileWithAnnotations.IsEmpty())
}

func TestInfoFile_String(t *testing.T) {
	content := `# Comment
file1.txt Annotation 1
file2.txt Annotation 2

# Another comment`

	infoFile := NewInfoFile("/test/.info", content)

	// Remove one annotation
	infoFile.RemoveAnnotationForPath("file2.txt")

	result := infoFile.String()
	expected := `# Comment
file1.txt Annotation 1

# Another comment`

	assert.Equal(t, expected, result)
}
