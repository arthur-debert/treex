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
				"parent": {Notes: "Parent annotation"},
				"child":  {Notes: "Child annotation"},
			},
			expectError: false,
			expectNodes: []string{"parent", "child"},
		},
		{
			name: "directory tree",
			annotations: map[string]*types.Annotation{
				"dir/":     {Notes: "Directory"},
				"dir/file": {Notes: "File in dir"},
			},
			expectError: false,
			expectNodes: []string{"dir", "file"},
		},
		{
			name: "nested directories",
			annotations: map[string]*types.Annotation{
				"a/":    {Notes: "Dir A"},
				"a/b/":  {Notes: "Dir B"},
				"a/b/c": {Notes: "File C"},
			},
			expectError: false,
			expectNodes: []string{"a", "b", "c"},
		},
		{
			name: "complex paths",
			annotations: map[string]*types.Annotation{
				"a/b/../c": {Notes: "Path C"},
				"d//e/":    {Notes: "Path E"},
			},
			expectError: false,
			expectNodes: []string{"a", "c", "d", "e"},
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
			root, err := BuildVirtualTree(tt.annotations)

			if tt.expectError {
				assert.Error(t, err, "Expected error for test: %s", tt.name)
				assert.Nil(t, root, "Expected nil root on error")
			} else {
				assert.NoError(t, err, "Unexpected error for test: %s", tt.name)
				assert.NotNil(t, root, "Expected non-nil root")
				// Root can be either "root" or the single top-level directory
				switch tt.name {
				case "directory tree":
					assert.Equal(t, "dir", root.Name, "Root should be 'dir' when it's the only top-level directory")
				case "nested directories":
					assert.Equal(t, "a", root.Name, "Root should be 'a' when it's the only top-level directory")
				default:
					assert.Equal(t, "root", root.Name, "Root should be named 'root'")
				}
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
	// Create a test info file with unsorted content
	content := `family/ The family
family/mom Mom annotation
family/kids/ The children
family/kids/bob Bob annotation
family/dad Dad annotation
family/kids/alice Alice annotation`

	tmpFile := createTempInfoFile(t, content)
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	// Test the command by calling runDrawCmd directly
	var output bytes.Buffer
	
	// Create a fresh command instance for testing
	cmd := &cobra.Command{
		Use:     "draw",
		RunE:    runDrawCmd,
	}
	cmd.Flags().StringP("format", "f", "color", "Output format")
	cmd.Flags().String("info-file", "", "Info file to read from")
	
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"--info-file", tmpFile, "--format=no-color"})
	err := cmd.Execute()

	// Check results
	require.NoError(t, err)

	// The output should contain the tree structure (formatting may vary)
	outputStr := output.String()
	assert.Contains(t, outputStr, "family")
	assert.Contains(t, outputStr, "dad")
	assert.Contains(t, outputStr, "Dad annotation")
	assert.Contains(t, outputStr, "mom")
	assert.Contains(t, outputStr, "Mom annotation")
	assert.Contains(t, outputStr, "kids")
	assert.Contains(t, outputStr, "The children")
	assert.Contains(t, outputStr, "alice")
	assert.Contains(t, outputStr, "Alice annotation")
	assert.Contains(t, outputStr, "bob")
	assert.Contains(t, outputStr, "Bob annotation")
	
	// Check the tree structure is correct
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	assert.Equal(t, "family", lines[0])
	assert.Contains(t, lines[1], "├─ dad")
	assert.Contains(t, lines[2], "├─ kids")
	assert.Contains(t, lines[3], "│  ├─ alice")
	assert.Contains(t, lines[4], "│  └─ bob")
	assert.Contains(t, lines[5], "└─ mom")
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

func TestDrawCommandStdin(t *testing.T) {
	// Prepare stdin with test data
	content := "from/stdin\nfrom/stdin/child child annotation"
	r, w, err := os.Pipe()
	require.NoError(t, err)

	_, err = w.WriteString(content)
	require.NoError(t, err)
	_ = w.Close()

	// Redirect stdin
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Create a fresh command instance for testing
	cmd := &cobra.Command{
		Use:     "draw",
		RunE:    runDrawCmd,
	}
	cmd.Flags().StringP("format", "f", "color", "Output format")
	cmd.Flags().String("info-file", "", "Info file to read from")
	
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--format=no-color"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Check output
	output := out.String()
	assert.Contains(t, output, "stdin", "Output should contain 'stdin'")
	assert.Contains(t, output, "child annotation", "Output should contain 'child annotation'")
}

func TestDrawCommandNoInput(t *testing.T) {
	// Create a fresh command instance for testing
	cmd := &cobra.Command{
		Use:     "draw",
		RunE:    runDrawCmd,
	}
	cmd.Flags().StringP("format", "f", "color", "Output format")
	cmd.Flags().String("info-file", "", "Info file to read from")
	
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}

func TestDrawCommandEmptyInfoFile(t *testing.T) {
	// Create an empty temp file
	tmpFile, err := os.CreateTemp("", "empty-*.info")
	require.NoError(t, err)
	t.Logf("Created empty file: %s", tmpFile.Name())
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	// Create a fresh command instance for testing
	cmd := &cobra.Command{
		Use:     "draw",
		RunE:    runDrawCmd,
	}
	cmd.Flags().StringP("format", "f", "color", "Output format")
	cmd.Flags().String("info-file", "", "Info file to read from")
	
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--info-file", tmpFile.Name()})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no annotations found")
}
