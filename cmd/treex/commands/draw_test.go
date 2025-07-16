package commands

import (
	"bytes"
	"os"
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
		expectNodes []string
	}{
		{
			name: "simple tree",
			annotations: map[string]*types.Annotation{
				"parent": {Text: "Parent annotation"},
				"child":  {Text: "Child annotation"},
			},
			expectError: false,
			expectNodes: []string{"parent", "child"},
		},
		{
			name: "directory tree",
			annotations: map[string]*types.Annotation{
				"dir/":     {Text: "Directory"},
				"dir/file": {Text: "File in dir"},
			},
			expectError: false,
			expectNodes: []string{"dir", "file"},
		},
		{
			name: "nested directories",
			annotations: map[string]*types.Annotation{
				"a/":    {Text: "Dir A"},
				"a/b/":  {Text: "Dir B"},
				"a/b/c": {Text: "File C"},
			},
			expectError: false,
			expectNodes: []string{"a", "b", "c"},
		},
		{
			name:        "empty annotations",
			annotations: map[string]*types.Annotation{},
			expectError: true,
			expectNodes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := buildVirtualTree(tt.annotations)

			if tt.expectError {
				assert.Error(t, err, "Expected error for test: %s", tt.name)
				assert.Nil(t, root, "Expected nil root on error")
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s", tt.name)
				assert.NotNil(t, root, "Expected non-nil root")
				assert.Equal(t, "root", root.Name, "Root should be named 'root'")
				assert.True(t, root.IsDir, "Root should be a directory")

				// Check that all expected nodes are present
				allNodes := collectAllNodes(root)
				for _, expectedNode := range tt.expectNodes {
					found := false
					for _, node := range allNodes {
						if node.Name == expectedNode {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected node %s not found", expectedNode)
				}
			}
		})
	}
}

func TestDrawCommandIntegration(t *testing.T) {
	// Create a test info file
	content := `family/ The family
family/dad Dad annotation
family/mom Mom annotation
family/kids/ The children
family/kids/alice Alice annotation
family/kids/bob Bob annotation`

	tmpFile := createTempInfoFile(t, content)
	defer os.Remove(tmpFile)

	// Save original values
	originalOutputFormat := outputFormat
	originalDrawInfoFile := drawInfoFile
	defer func() {
		outputFormat = originalOutputFormat
		drawInfoFile = originalDrawInfoFile
	}()

	// Test the command
	cmd := &cobra.Command{}
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Set flags
	outputFormat = "no-color"
	drawInfoFile = tmpFile

	// Run command
	err := runDrawCmd(cmd, []string{})

	// Check results
	assert.NoError(t, err)

	result := output.String()
	assert.Contains(t, result, "family")
	assert.Contains(t, result, "dad")
	assert.Contains(t, result, "mom")
	assert.Contains(t, result, "kids")
	assert.Contains(t, result, "alice")
	assert.Contains(t, result, "bob")

	// Check that it's formatted as a tree
	assert.Contains(t, result, "├─")
	assert.Contains(t, result, "└─")
}

func TestDrawCommandFlags(t *testing.T) {
	// Test that the draw command has the correct flags
	cmd := drawCmd

	// Check that info-file flag exists
	infoFileFlag := cmd.Flags().Lookup("info-file")
	assert.NotNil(t, infoFileFlag, "info-file flag should exist")

	// Check that format flag exists
	formatFlag := cmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag, "format flag should exist")
	assert.Equal(t, "color", formatFlag.DefValue, "format flag should default to color")
}

func TestDrawCommandHelp(t *testing.T) {
	// Test that the draw command has help text
	cmd := drawCmd
	assert.NotEmpty(t, cmd.Long, "Draw command should have help text")
	assert.Contains(t, cmd.Long, "tree diagrams", "Help should mention tree diagrams")
	assert.Contains(t, cmd.Long, "info files", "Help should mention info files")
}

func TestSortNodeChildren(t *testing.T) {
	// Create a test tree with unsorted children
	root := &types.Node{
		Name:  "root",
		IsDir: true,
		Children: []*types.Node{
			{Name: "zzz", IsDir: false},
			{Name: "aaa", IsDir: false},
			{Name: "mmm", IsDir: false},
		},
	}

	// Sort the children
	sortNodeChildren(root)

	// Check that children are sorted
	require.Len(t, root.Children, 3)
	assert.Equal(t, "aaa", root.Children[0].Name)
	assert.Equal(t, "mmm", root.Children[1].Name)
	assert.Equal(t, "zzz", root.Children[2].Name)
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"color", "color"},
		{"no-color", "no-color"},
		{"markdown", "markdown"},
		{"invalid", "invalid"}, // Should pass through for validation by manager
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFormat(tt.input)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

// Helper functions

func createTempInfoFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "treex-test-*.info")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func collectAllNodes(root *types.Node) []*types.Node {
	var nodes []*types.Node

	var collect func(*types.Node)
	collect = func(node *types.Node) {
		nodes = append(nodes, node)
		for _, child := range node.Children {
			collect(child)
		}
	}

	collect(root)
	return nodes
}