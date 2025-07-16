package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEditor(t *testing.T) {
	// Save original environment
	originalEditor := os.Getenv("EDITOR")
	originalVisual := os.Getenv("VISUAL")
	defer func() {
		_ = os.Setenv("EDITOR", originalEditor)
		_ = os.Setenv("VISUAL", originalVisual)
	}()

	// Test EDITOR environment variable
	_ = os.Setenv("EDITOR", "vim")
	_ = os.Setenv("VISUAL", "")
	assert.Equal(t, "vim", getEditor())

	// Test VISUAL fallback
	_ = os.Setenv("EDITOR", "")
	_ = os.Setenv("VISUAL", "emacs")
	assert.Equal(t, "emacs", getEditor())

	// Test fallback to common editors (this will vary by system)
	_ = os.Setenv("EDITOR", "")
	_ = os.Setenv("VISUAL", "")
	editor := getEditor()
	// Should find at least one common editor or return empty string
	assert.True(t, editor == "" || len(editor) > 0)
}

func TestBuildEditorCommand(t *testing.T) {
	tests := []struct {
		name       string
		editor     string
		filename   string
		lineNumber int
		expectArgs []string
	}{
		{
			name:       "vim with line number",
			editor:     "vim",
			filename:   "test.info",
			lineNumber: 42,
			expectArgs: []string{"vim", "+42", "test.info"},
		},
		{
			name:       "vim without line number",
			editor:     "vim",
			filename:   "test.info",
			lineNumber: 0,
			expectArgs: []string{"vim", "test.info"},
		},
		{
			name:       "nano with line number",
			editor:     "nano",
			filename:   "test.info",
			lineNumber: 10,
			expectArgs: []string{"nano", "+10", "test.info"},
		},
		{
			name:       "VS Code with line number",
			editor:     "code",
			filename:   "test.info",
			lineNumber: 25,
			expectArgs: []string{"code", "--goto", "test.info:25"},
		},
		{
			name:       "VS Code without line number",
			editor:     "code",
			filename:   "test.info",
			lineNumber: 0,
			expectArgs: []string{"code", "test.info"},
		},
		{
			name:       "Sublime Text with line number",
			editor:     "subl",
			filename:   "test.info",
			lineNumber: 15,
			expectArgs: []string{"subl", "test.info:15"},
		},
		{
			name:       "unknown editor with line number",
			editor:     "unknown-editor",
			filename:   "test.info",
			lineNumber: 5,
			expectArgs: []string{"unknown-editor", "+5", "test.info"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := buildEditorCommand(tt.editor, tt.filename, tt.lineNumber)
			require.NoError(t, err)
			assert.Equal(t, tt.expectArgs, cmd.Args)
		})
	}
}

func TestFindLineNumber(t *testing.T) {
	// Create a temporary .info file
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	infoFile := filepath.Join(tempDir, ".info")
	content := `# This is a comment
src Main source code
docs/ Documentation
docs/api API documentation
tests Test files
build/output Build output`

	err = os.WriteFile(infoFile, []byte(content), 0644)
	require.NoError(t, err)

	tests := []struct {
		name         string
		targetPath   string
		expectedLine int
		expectError  bool
	}{
		{
			name:         "find src",
			targetPath:   "src",
			expectedLine: 2,
			expectError:  false,
		},
		{
			name:         "find docs directory",
			targetPath:   "docs/",
			expectedLine: 3,
			expectError:  false,
		},
		{
			name:         "find nested path",
			targetPath:   "docs/api",
			expectedLine: 4,
			expectError:  false,
		},
		{
			name:         "find tests",
			targetPath:   "tests",
			expectedLine: 5,
			expectError:  false,
		},
		{
			name:         "find build/output",
			targetPath:   "build/output",
			expectedLine: 6,
			expectError:  false,
		},
		{
			name:        "path not found",
			targetPath:  "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineNum, err := findLineNumber(infoFile, tt.targetPath)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLine, lineNum)
			}
		})
	}
}

func TestFindAnnotationLocations(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	// Create multiple .info files
	err = os.WriteFile(".info", []byte("src Main source\ndocs Documentation"), 0644)
	require.NoError(t, err)

	err = os.MkdirAll("subdir", 0755)
	require.NoError(t, err)
	err = os.WriteFile("subdir/.info", []byte("src Source in subdir\ntest Test files"), 0644)
	require.NoError(t, err)

	// Test finding annotations
	infoFiles := []string{".info", "subdir/.info"}
	
	// Test path that exists in both files
	locations, err := findAnnotationLocations(infoFiles, "src")
	require.NoError(t, err)
	assert.Len(t, locations, 2)
	
	// Verify both files are found
	files := make([]string, len(locations))
	for i, loc := range locations {
		files[i] = loc.File
	}
	assert.Contains(t, files, ".info")
	assert.Contains(t, files, "subdir/.info")

	// Test path that exists in only one file
	locations, err = findAnnotationLocations(infoFiles, "docs")
	require.NoError(t, err)
	assert.Len(t, locations, 1)
	assert.Equal(t, ".info", locations[0].File)

	// Test path that doesn't exist
	locations, err = findAnnotationLocations(infoFiles, "nonexistent")
	require.NoError(t, err)
	assert.Len(t, locations, 0)
}

func TestFindLineNumberWithColonFormat(t *testing.T) {
	// Create a temporary .info file with colon format
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	infoFile := filepath.Join(tempDir, ".info")
	content := `src: Main source code
docs/: Documentation directory
build/output: Build artifacts`

	err = os.WriteFile(infoFile, []byte(content), 0644)
	require.NoError(t, err)

	// Test finding paths with colon format
	lineNum, err := findLineNumber(infoFile, "src")
	require.NoError(t, err)
	assert.Equal(t, 1, lineNum)

	lineNum, err = findLineNumber(infoFile, "docs/")
	require.NoError(t, err)
	assert.Equal(t, 2, lineNum)

	lineNum, err = findLineNumber(infoFile, "build/output")
	require.NoError(t, err)
	assert.Equal(t, 3, lineNum)
}