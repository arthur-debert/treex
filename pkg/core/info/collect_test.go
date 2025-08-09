package info

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectInfoFiles(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     map[string]string // path -> content
		expectedResult func(t *testing.T, result *CollectResult)
		expectError    bool
	}{
		{
			name: "collects from single subdirectory",
			setupFiles: map[string]string{
				"root/.info":        "README.md: Main docs\n",
				"root/src/.info":    "main.go: Entry point\nutils.go: Utilities\n",
			},
			expectedResult: func(t *testing.T, result *CollectResult) {
				assert.Equal(t, 2, len(result.CollectedFiles))
				assert.Equal(t, 3, result.TotalEntries)
				assert.Contains(t, result.MergedContent, "README.md: Main docs")
				assert.Contains(t, result.MergedContent, "src/main.go: Entry point")
				assert.Contains(t, result.MergedContent, "src/utils.go: Utilities")
			},
		},
		{
			name: "collects from multiple nested directories",
			setupFiles: map[string]string{
				"root/docs/.info":           "guide.md: User guide\n",
				"root/src/.info":            "app.go: Application\n",
				"root/src/handlers/.info":   "api.go: API handlers\n",
			},
			expectedResult: func(t *testing.T, result *CollectResult) {
				assert.Equal(t, 3, len(result.CollectedFiles))
				assert.Equal(t, 3, result.TotalEntries)
				assert.Contains(t, result.MergedContent, "docs/guide.md: User guide")
				assert.Contains(t, result.MergedContent, "src/app.go: Application")
				assert.Contains(t, result.MergedContent, "src/handlers/api.go: API handlers")
			},
		},
		{
			name: "handles conflicts with closest file winning",
			setupFiles: map[string]string{
				"root/.info":        "src/main.go: Root annotation\n",
				"root/src/.info":    "main.go: Src annotation\n",
			},
			expectedResult: func(t *testing.T, result *CollectResult) {
				assert.Equal(t, 1, len(result.ConflictResolutions))
				assert.Contains(t, result.MergedContent, "src/main.go: Src annotation")
				assert.NotContains(t, result.MergedContent, "Root annotation")
			},
		},
		{
			name: "creates root info file if it doesn't exist",
			setupFiles: map[string]string{
				"root/sub/.info": "file.txt: A file\n",
			},
			expectedResult: func(t *testing.T, result *CollectResult) {
				assert.Equal(t, 1, len(result.CollectedFiles))
				assert.Equal(t, 1, result.TotalEntries)
				assert.Contains(t, result.MergedContent, "sub/file.txt: A file")
			},
		},
		{
			name: "handles empty info files",
			setupFiles: map[string]string{
				"root/.info":     "",
				"root/sub/.info": "",
			},
			expectedResult: func(t *testing.T, result *CollectResult) {
				assert.Equal(t, 2, len(result.CollectedFiles))
				assert.Equal(t, 0, result.TotalEntries)
				assert.Equal(t, "\n", result.MergedContent)
			},
		},
		{
			name: "ignores comments and invalid lines",
			setupFiles: map[string]string{
				"root/sub/.info": "# This is a comment\nfile.txt: Valid entry\nInvalid line without colon\n  \nother.txt: Another valid\n",
			},
			expectedResult: func(t *testing.T, result *CollectResult) {
				assert.Equal(t, 1, len(result.CollectedFiles))
				assert.Equal(t, 2, result.TotalEntries)
				assert.Contains(t, result.MergedContent, "sub/file.txt: Valid entry")
				assert.Contains(t, result.MergedContent, "sub/other.txt: Another valid")
				assert.NotContains(t, result.MergedContent, "comment")
				assert.NotContains(t, result.MergedContent, "Invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()
			
			// Setup test files
			for path, content := range tt.setupFiles {
				fullPath := filepath.Join(tempDir, path)
				dir := filepath.Dir(fullPath)
				require.NoError(t, os.MkdirAll(dir, 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			// Run collection
			rootPath := filepath.Join(tempDir, "root")
			result, err := CollectInfoFiles(rootPath, CollectOptions{})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				tt.expectedResult(t, result)
				
				// Verify root .info file was created/updated
				rootInfoPath := filepath.Join(rootPath, ".info")
				content, err := os.ReadFile(rootInfoPath)
				assert.NoError(t, err)
				assert.Equal(t, result.MergedContent, string(content))
				
				// Verify child .info files were deleted
				for _, infoPath := range result.CollectedFiles {
					if infoPath != rootInfoPath {
						_, err := os.Stat(infoPath)
						assert.True(t, os.IsNotExist(err), "Child info file should be deleted: %s", infoPath)
					}
				}
			}
		})
	}
}

func TestDryRun(t *testing.T) {
	tempDir := t.TempDir()
	
	// Setup test files
	setupFiles := map[string]string{
		"root/.info":     "README.md: Docs\n",
		"root/src/.info": "main.go: Main\n",
	}
	
	for path, content := range setupFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}
	
	// Run with dry run
	rootPath := filepath.Join(tempDir, "root")
	result, err := CollectInfoFiles(rootPath, CollectOptions{DryRun: true})
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalEntries)
	
	// Verify files were NOT modified
	for path, expectedContent := range setupFiles {
		fullPath := filepath.Join(tempDir, path)
		content, err := os.ReadFile(fullPath)
		assert.NoError(t, err)
		assert.Equal(t, expectedContent, string(content), "File should not be modified in dry run: %s", path)
	}
}

func TestCalculatePathDistance(t *testing.T) {
	tests := []struct {
		from     string
		to       string
		expected int
	}{
		{
			from:     "/root/.info",
			to:       "/root/file.txt",
			expected: 0, // Same directory
		},
		{
			from:     "/root/src/.info",
			to:       "/root/src/main.go",
			expected: 0, // Same directory
		},
		{
			from:     "/root/.info",
			to:       "/root/src/main.go",
			expected: 1, // One level down
		},
		{
			from:     "/root/src/.info",
			to:       "/root/file.txt",
			expected: 1, // One level up
		},
		{
			from:     "/root/a/b/.info",
			to:       "/root/x/y/z.txt",
			expected: 4, // 2 up + 2 down
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			distance := calculatePathDistance(tt.from, tt.to)
			assert.Equal(t, tt.expected, distance)
		})
	}
}

func TestIsCloserToPath(t *testing.T) {
	tests := []struct {
		name       string
		file1      string
		file2      string
		targetPath string
		rootPath   string
		expect1Closer bool
	}{
		{
			name:       "subdirectory info is closer to its files",
			file1:      "/root/src/.info",
			file2:      "/root/.info",
			targetPath: "src/main.go",
			rootPath:   "/root",
			expect1Closer: true,
		},
		{
			name:       "root info is farther from subdirectory files",
			file1:      "/root/.info",
			file2:      "/root/src/.info",
			targetPath: "src/main.go",
			rootPath:   "/root",
			expect1Closer: false,
		},
		{
			name:       "same distance ties go to first",
			file1:      "/root/a/.info",
			file2:      "/root/b/.info",
			targetPath: "c/file.txt",
			rootPath:   "/root",
			expect1Closer: false, // Same distance, so not closer
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCloserToPath(tt.file1, tt.file2, tt.targetPath, tt.rootPath)
			assert.Equal(t, tt.expect1Closer, result)
		})
	}
}

func TestGenerateMergedContent(t *testing.T) {
	entries := map[string]*collectedEntry{
		"b/file2.txt": {path: "b/file2.txt", annotation: "Second file"},
		"a/file1.txt": {path: "a/file1.txt", annotation: "First file"},
		"c/file3.txt": {path: "c/file3.txt", annotation: "Third file"},
	}
	
	content := generateMergedContent(entries, false)
	
	// Should be sorted alphabetically
	lines := strings.Split(strings.TrimSpace(content), "\n")
	assert.Equal(t, 3, len(lines))
	assert.Equal(t, "a/file1.txt: First file", lines[0])
	assert.Equal(t, "b/file2.txt: Second file", lines[1])
	assert.Equal(t, "c/file3.txt: Third file", lines[2])
}