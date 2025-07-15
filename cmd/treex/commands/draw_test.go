package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVirtualTree(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]*types.Annotation
		expectError bool
		expected    func(t *testing.T, root *types.Node)
	}{
		{
			name: "simple flat structure",
			annotations: map[string]*types.Annotation{
				"file1.txt": {Path: "file1.txt", Notes: "First file"},
				"file2.txt": {Path: "file2.txt", Notes: "Second file"},
			},
			expectError: false,
			expected: func(t *testing.T, root *types.Node) {
				assert.Equal(t, "root", root.Name)
				assert.True(t, root.IsDir)
				assert.Len(t, root.Children, 2)
				
				// Check files exist
				var file1, file2 *types.Node
				for _, child := range root.Children {
					if child.Name == "file1.txt" {
						file1 = child
					} else if child.Name == "file2.txt" {
						file2 = child
					}
				}
				
				require.NotNil(t, file1)
				require.NotNil(t, file2)
				assert.Equal(t, "First file", file1.Annotation.Notes)
				assert.Equal(t, "Second file", file2.Annotation.Notes)
				assert.False(t, file1.IsDir)
				assert.False(t, file2.IsDir)
			},
		},
		{
			name: "nested directory structure",
			annotations: map[string]*types.Annotation{
				"src/": {Path: "src/", Notes: "Source code"},
				"src/main.go": {Path: "src/main.go", Notes: "Main file"},
				"docs/": {Path: "docs/", Notes: "Documentation"},
				"docs/README.md": {Path: "docs/README.md", Notes: "Read me"},
			},
			expectError: false,
			expected: func(t *testing.T, root *types.Node) {
				assert.Equal(t, "root", root.Name)
				assert.Len(t, root.Children, 2)
				
				// Find src and docs directories
				var src, docs *types.Node
				for _, child := range root.Children {
					if child.Name == "src" {
						src = child
					} else if child.Name == "docs" {
						docs = child
					}
				}
				
				require.NotNil(t, src)
				require.NotNil(t, docs)
				assert.True(t, src.IsDir)
				assert.True(t, docs.IsDir)
				assert.Equal(t, "Source code", src.Annotation.Notes)
				assert.Equal(t, "Documentation", docs.Annotation.Notes)
				
				// Check src contents
				assert.Len(t, src.Children, 1)
				mainGo := src.Children[0]
				assert.Equal(t, "main.go", mainGo.Name)
				assert.Equal(t, "Main file", mainGo.Annotation.Notes)
				assert.False(t, mainGo.IsDir)
				
				// Check docs contents
				assert.Len(t, docs.Children, 1)
				readme := docs.Children[0]
				assert.Equal(t, "README.md", readme.Name)
				assert.Equal(t, "Read me", readme.Annotation.Notes)
				assert.False(t, readme.IsDir)
			},
		},
		{
			name:        "empty annotations",
			annotations: map[string]*types.Annotation{},
			expectError: true,
		},
		{
			name:        "nil annotations",
			annotations: nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := BuildVirtualTree(tt.annotations)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, root)
			
			if tt.expected != nil {
				tt.expected(t, root)
			}
		})
	}
}

func TestDrawCommand(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.info")
	
	testContent := `Dad Chill, dad
Mom Listen to your mother
kids/ Children
kids/Sam Little Sam
kids/Alex The smart one`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)
	
	// Create a new draw command to avoid interference with global state
	testCmd := &cobra.Command{
		Use:   "draw --info-file FILE",
		Short: "Draw tree diagrams from info files",
		RunE:  runDrawCmd,
	}
	
	// Set up flags
	var testOutputFormat string
	var testInfoFile string
	testCmd.Flags().StringVarP(&testOutputFormat, "format", "f", "color", "Output format")
	testCmd.Flags().StringVar(&testInfoFile, "info-file", "", "Info file to draw from")
	
	// Set args
	testCmd.SetArgs([]string{"--info-file", testFile})
	
	// Capture output
	var output strings.Builder
	testCmd.SetOut(&output)
	testCmd.SetErr(&output)
	
	// Run the command
	err = testCmd.Execute()
	require.NoError(t, err)
	
	// Check output contains expected elements
	result := output.String()
	assert.Contains(t, result, "root")
	assert.Contains(t, result, "Dad")
	assert.Contains(t, result, "Mom")
	assert.Contains(t, result, "kids")
	assert.Contains(t, result, "Sam")
	assert.Contains(t, result, "Alex")
	assert.Contains(t, result, "Chill, dad")
	assert.Contains(t, result, "Listen to your mother")
	assert.Contains(t, result, "Children")
	assert.Contains(t, result, "Little Sam")
	assert.Contains(t, result, "The smart one")
}

func TestDrawCommandWithNonExistentFile(t *testing.T) {
	// Create a new draw command to avoid interference with global state
	testCmd := &cobra.Command{
		Use:   "draw --info-file FILE",
		Short: "Draw tree diagrams from info files",
		RunE:  runDrawCmd,
	}
	
	// Set up flags
	var testOutputFormat string
	var testInfoFile string
	testCmd.Flags().StringVarP(&testOutputFormat, "format", "f", "color", "Output format")
	testCmd.Flags().StringVar(&testInfoFile, "info-file", "", "Info file to draw from")
	
	testCmd.SetArgs([]string{"--info-file", "/nonexistent/file.info"})
	
	// Capture error output
	var errorOutput strings.Builder
	testCmd.SetErr(&errorOutput)
	
	// Run the command - should fail
	err := testCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse info file")
}

func TestDrawCommandWithEmptyFile(t *testing.T) {
	// Create empty test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.info")
	
	err := os.WriteFile(testFile, []byte(""), 0644)
	require.NoError(t, err)
	
	// Create a new draw command to avoid interference with global state
	testCmd := &cobra.Command{
		Use:   "draw --info-file FILE",
		Short: "Draw tree diagrams from info files",
		RunE:  runDrawCmd,
	}
	
	// Set up flags
	var testOutputFormat string
	var testInfoFile string
	testCmd.Flags().StringVarP(&testOutputFormat, "format", "f", "color", "Output format")
	testCmd.Flags().StringVar(&testInfoFile, "info-file", "", "Info file to draw from")
	
	testCmd.SetArgs([]string{"--info-file", testFile})
	
	// Run the command - should fail
	err = testCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no annotations found")
}

func TestDrawCommandWithInvalidFormat(t *testing.T) {
	// Create test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.info")
	
	err := os.WriteFile(testFile, []byte("file.txt Some annotation"), 0644)
	require.NoError(t, err)
	
	// Create a new draw command to avoid interference with global state
	testCmd := &cobra.Command{
		Use:   "draw --info-file FILE",
		Short: "Draw tree diagrams from info files",
		RunE:  runDrawCmd,
	}
	
	// Set up flags
	var testOutputFormat string
	var testInfoFile string
	testCmd.Flags().StringVarP(&testOutputFormat, "format", "f", "color", "Output format")
	testCmd.Flags().StringVar(&testInfoFile, "info-file", "", "Info file to draw from")
	
	testCmd.SetArgs([]string{"--info-file", testFile, "--format", "invalid"})
	
	// Run the command - should fail
	err = testCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

func TestEnsureParentDirectories(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		expected    func(t *testing.T, nodeMap map[string]*types.Node)
	}{
		{
			name:        "simple path",
			path:        "a/b/c",
			expectError: false,
			expected: func(t *testing.T, nodeMap map[string]*types.Node) {
				assert.Contains(t, nodeMap, "a")
				assert.Contains(t, nodeMap, "a/b")
				assert.True(t, nodeMap["a"].IsDir)
				assert.True(t, nodeMap["a/b"].IsDir)
				assert.Equal(t, "a", nodeMap["a"].Name)
				assert.Equal(t, "b", nodeMap["a/b"].Name)
			},
		},
		{
			name:        "single level",
			path:        "file.txt",
			expectError: false,
			expected: func(t *testing.T, nodeMap map[string]*types.Node) {
				// Should not create any parent directories
				assert.Len(t, nodeMap, 1) // Only root should exist
			},
		},
		{
			name:        "empty path",
			path:        "",
			expectError: false,
			expected: func(t *testing.T, nodeMap map[string]*types.Node) {
				assert.Len(t, nodeMap, 1) // Only root should exist
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root node
			root := &types.Node{
				Name:         "root",
				Path:         "",
				RelativePath: "",
				IsDir:        true,
				Children:     make([]*types.Node, 0),
				Parent:       nil,
			}
			
			nodeMap := make(map[string]*types.Node)
			nodeMap[""] = root
			
			err := EnsureParentDirectories(tt.path, nodeMap, root)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			
			if tt.expected != nil {
				tt.expected(t, nodeMap)
			}
		})
	}
}