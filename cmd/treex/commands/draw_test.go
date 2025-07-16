package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVirtualTree(t *testing.T) {
	// Test basic functionality
	annotations := map[string]*types.Annotation{
		"Dad": {Path: "Dad", Notes: "Chill, dad"},
		"Mom": {Path: "Mom", Notes: "Listen to your mother"},
		"kids/Sam": {Path: "kids/Sam", Notes: "Little Sam"},
	}

	tree, err := buildVirtualTree(annotations)
	require.NoError(t, err)
	require.NotNil(t, tree)

	assert.Equal(t, "root", tree.Name)
	assert.True(t, tree.IsDir)
	assert.Len(t, tree.Children, 3) // Dad, Mom, kids

	// Test with empty annotations
	emptyTree, err := buildVirtualTree(map[string]*types.Annotation{})
	assert.Error(t, err)
	assert.Nil(t, emptyTree)
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
	
	// Create output buffer
	var output bytes.Buffer
	
	// Reset command state
	drawCmd.SetOut(&output)
	drawCmd.SetErr(&output)
	drawCmd.SetArgs([]string{"--info-file", testFile, "--format", "no-color"})
	
	// Execute the command
	err = drawCmd.Execute()
	require.NoError(t, err)
	
	// Check output contains expected elements
	result := output.String()
	assert.Contains(t, result, "Dad")
	assert.Contains(t, result, "Mom")
	assert.Contains(t, result, "kids")
	assert.Contains(t, result, "Sam")
	assert.Contains(t, result, "Alex")
	assert.Contains(t, result, "Children")
	
	// Should not contain the current directory contents
	assert.NotContains(t, result, "commands")
	assert.NotContains(t, result, "add_info.go")
}

func TestDrawCommandFlags(t *testing.T) {
	// Test that the command has the expected flags
	cmd := drawCmd

	assert.Equal(t, "draw [--info-file FILE | -]", cmd.Use)
	assert.Equal(t, "Draw tree diagrams from info files without filesystem validation", cmd.Short)

	// Check that flags are properly defined
	formatFlag := cmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag)

	infoFileFlag := cmd.Flags().Lookup("info-file")
	assert.NotNil(t, infoFileFlag)
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty format",
			input:    "",
			expected: "",
		},
		{
			name:     "color format",
			input:    "color",
			expected: "color",
		},
		{
			name:     "no-color format",
			input:    "no-color",
			expected: "no-color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFormat(tt.input)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}