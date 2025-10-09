package info

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfoFileSetLoader tests single file loading functionality
func TestInfoFileSetLoader_LoadSingleInfoFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)
	loader := NewInfoFileSetLoader(infoFS)

	// Create a test .info file
	content := `file1.txt Annotation 1
file2.txt Annotation 2`

	err := afero.WriteFile(fs, "/test/.info", []byte(content), 0644)
	require.NoError(t, err)

	// Load the file
	infoFile, err := loader.LoadSingleInfoFile("/test/.info")
	require.NoError(t, err)

	assert.Equal(t, "/test/.info", infoFile.Path)
	assert.Len(t, infoFile.annotations, 2)
	assert.True(t, infoFile.HasAnnotationForPath("file1.txt"))
	assert.True(t, infoFile.HasAnnotationForPath("file2.txt"))
}

// TestInfoFileSetLoader_LoadFromPath tests directory traversal and loading
func TestInfoFileSetLoader_LoadFromPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)
	loader := NewInfoFileSetLoader(infoFS)

	// Create test .info files
	err := afero.WriteFile(fs, "/project/.info", []byte("root.txt Root annotation"), 0644)
	require.NoError(t, err)

	err = fs.MkdirAll("/project/sub", 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, "/project/sub/.info", []byte("sub.txt Sub annotation"), 0644)
	require.NoError(t, err)

	// Load all files
	infoFileSet, err := loader.LoadFromPath("/project")
	require.NoError(t, err)

	infoFiles := infoFileSet.GetFiles()
	assert.Len(t, infoFiles, 2)

	// Check that both files were loaded
	paths := make(map[string]bool)
	for _, infoFile := range infoFiles {
		paths[infoFile.Path] = true
	}
	assert.True(t, paths["/project/.info"])
	assert.True(t, paths["/project/sub/.info"])
}

// TestInfoFileSetWriter_WriteSingleInfoFile tests single file writing
func TestInfoFileSetWriter_WriteSingleInfoFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)
	writer := NewInfoFileSetWriter(infoFS)

	// Create directory first
	err := fs.MkdirAll("/test", 0755)
	require.NoError(t, err)

	// Create an InfoFile with some content to avoid empty string issue
	infoFile := NewInfoFile("/test/.info", "# Comment")
	infoFile.AddAnnotationForPath("file1.txt", "Test annotation")

	// Write to disk
	err = writer.WriteSingleInfoFile(infoFile)
	require.NoError(t, err)

	// Verify content was written
	content, err := afero.ReadFile(fs, "/test/.info")
	require.NoError(t, err)
	expected := "# Comment\nfile1.txt Test annotation"
	assert.Equal(t, expected, string(content))
}

// TestInfoFileSet_WithFilesystem tests filesystem I/O boundaries
// These tests are extracted from the working filesystem integration tests
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

		// Create the directory and target files
		err := fs.MkdirAll("test", 0755)
		require.NoError(t, err)

		err = afero.WriteFile(fs, "test/a.txt", []byte("content a"), 0644)
		require.NoError(t, err)

		err = afero.WriteFile(fs, "test/b.txt", []byte("content b"), 0644)
		require.NoError(t, err)

		// Create original .info file
		originalContent := `# Comment line
a.txt  First annotation
b.txt  Second annotation
# Another comment`

		err = afero.WriteFile(fs, "test/.info", []byte(originalContent), 0644)
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

// TestAferoInfoFileSystem tests the filesystem abstraction
func TestAferoInfoFileSystem(t *testing.T) {
	fs := afero.NewMemMapFs()
	infoFS := NewAferoInfoFileSystem(fs)

	t.Run("ReadInfoFile", func(t *testing.T) {
		content := "test.txt Test annotation"
		err := afero.WriteFile(fs, ".info", []byte(content), 0644)
		require.NoError(t, err)

		reader, err := infoFS.ReadInfoFile(".info")
		require.NoError(t, err)

		data, err := afero.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("WriteInfoFile", func(t *testing.T) {
		content := "file.txt File annotation"
		err := infoFS.WriteInfoFile("test/.info", content)
		require.NoError(t, err)

		data, err := afero.ReadFile(fs, "test/.info")
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("PathExists", func(t *testing.T) {
		err := afero.WriteFile(fs, "existing.txt", []byte("content"), 0644)
		require.NoError(t, err)

		assert.True(t, infoFS.PathExists("existing.txt"))
		assert.False(t, infoFS.PathExists("nonexistent.txt"))
	})

	t.Run("FindInfoFiles", func(t *testing.T) {
		// Use a clean filesystem for this test
		testFS := afero.NewMemMapFs()
		testInfoFS := NewAferoInfoFileSystem(testFS)

		// Create .info files in different directories
		err := afero.WriteFile(testFS, "test/.info", []byte("root"), 0644)
		require.NoError(t, err)

		err = testFS.MkdirAll("test/sub/deep", 0755)
		require.NoError(t, err)

		err = afero.WriteFile(testFS, "test/sub/.info", []byte("sub"), 0644)
		require.NoError(t, err)

		err = afero.WriteFile(testFS, "test/sub/deep/.info", []byte("deep"), 0644)
		require.NoError(t, err)

		infoFiles, err := testInfoFS.FindInfoFiles("test")
		require.NoError(t, err)

		assert.Len(t, infoFiles, 3)
		assert.Contains(t, infoFiles, "test/.info")
		assert.Contains(t, infoFiles, "test/sub/.info")
		assert.Contains(t, infoFiles, "test/sub/deep/.info")
	})
}
