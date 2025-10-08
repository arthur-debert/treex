package info

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestInfoFileLoader_LoadInfoFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)
	loader := NewInfoFileLoader(infoFS)

	// Create a test .info file
	content := `file1.txt Annotation 1
file2.txt Annotation 2`

	err := afero.WriteFile(fs, "/test/.info", []byte(content), 0644)
	require.NoError(t, err)

	// Load the file
	infoFile, err := loader.LoadInfoFile("/test/.info")
	require.NoError(t, err)

	assert.Equal(t, "/test/.info", infoFile.Path)
	assert.Len(t, infoFile.annotations, 2)
	assert.True(t, infoFile.HasAnnotationForPath("file1.txt"))
	assert.True(t, infoFile.HasAnnotationForPath("file2.txt"))
}

func TestInfoFileLoader_LoadInfoFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)
	loader := NewInfoFileLoader(infoFS)

	// Create test .info files
	err := afero.WriteFile(fs, "/project/.info", []byte("root.txt Root annotation"), 0644)
	require.NoError(t, err)

	err = fs.MkdirAll("/project/sub", 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "/project/sub/.info", []byte("sub.txt Sub annotation"), 0644)
	require.NoError(t, err)

	// Load all files
	infoFiles, err := loader.LoadInfoFiles("/project")
	require.NoError(t, err)

	assert.Len(t, infoFiles, 2)

	// Check that both files were loaded
	paths := make(map[string]bool)
	for _, infoFile := range infoFiles {
		paths[infoFile.Path] = true
	}
	assert.True(t, paths["/project/.info"])
	assert.True(t, paths["/project/sub/.info"])
}

func TestInfoFileWriter_WriteInfoFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)
	writer := NewInfoFileWriter(infoFS)

	// Create directory first
	err := fs.MkdirAll("/test", 0755)
	require.NoError(t, err)

	// Create an InfoFile with some content to avoid empty string issue
	infoFile := NewInfoFile("/test/.info", "# Comment")
	infoFile.AddAnnotationForPath("file1.txt", "Test annotation")

	// Write to disk
	err = writer.WriteInfoFile(infoFile)
	require.NoError(t, err)

	// Verify content was written
	content, err := afero.ReadFile(fs, "/test/.info")
	require.NoError(t, err)
	expected := "# Comment\nfile1.txt Test annotation"
	assert.Equal(t, expected, string(content))
}
