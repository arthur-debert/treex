package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupInitCmd creates a properly initialized test init command
func setupInitCmd() *cobra.Command {
	// Reset forceInit to default
	forceInit = false

	// Create a test root command
	testRootCmd := &cobra.Command{
		Use:   "treex",
		Short: "Test root command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Create init command with flags
	testInitCmd := &cobra.Command{
		Use:   "init [path...]",
		Short: "Initialize a .info file for a directory or specific paths",
		Args:  cobra.ArbitraryArgs,
		RunE:  runInitCmd,
	}
	testInitCmd.Flags().IntP("depth", "d", 3, "Maximum depth to scan")
	testInitCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing .info file without confirmation")

	// Add the init command
	testRootCmd.AddCommand(testInitCmd)

	return testRootCmd
}

// executeInitCommand is a helper function to execute the init command
func executeInitCommand(args ...string) (output string, err error) {
	root := setupInitCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(append([]string{"init"}, args...))

	_, err = root.ExecuteC()
	return buf.String(), err
}

func TestInitCommandForceFlag(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	// Create an existing .info file
	existingContent := "src Source code\ntest Test files\n"
	err = os.WriteFile(".info", []byte(existingContent), 0644)
	require.NoError(t, err)

	// Create some directories to be scanned
	err = os.MkdirAll("src/main", 0755)
	require.NoError(t, err)
	err = os.MkdirAll("test", 0755)
	require.NoError(t, err)

	// Test with --force flag
	_, err = executeInitCommand("--force", "--depth=1")
	require.NoError(t, err)

	// Check that the file was overwritten
	content, err := os.ReadFile(".info")
	require.NoError(t, err)

	// The new content should not have annotations (just paths)
	contentStr := string(content)
	assert.Contains(t, contentStr, "src")
	assert.Contains(t, contentStr, "test")
	assert.NotContains(t, contentStr, "Source code") // Old annotation should be gone

	// The success message goes to stdout, not the captured output
}

func TestInitCommandForceFlagWithPaths(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	// Create an existing .info file
	existingContent := "oldpath Old annotation\n"
	err = os.WriteFile(".info", []byte(existingContent), 0644)
	require.NoError(t, err)

	// Create some files
	err = os.MkdirAll("src", 0755)
	require.NoError(t, err)
	err = os.WriteFile("src/main.go", []byte("package main"), 0644)
	require.NoError(t, err)
	err = os.WriteFile("README.md", []byte("# Test"), 0644)
	require.NoError(t, err)

	// Test with --force flag and specific paths
	_, err = executeInitCommand("--force", "src/main.go", "README.md")
	require.NoError(t, err)

	// Check that the file was overwritten
	content, err := os.ReadFile(".info")
	require.NoError(t, err)

	// The new content should only have the specified paths
	contentStr := string(content)
	assert.Contains(t, contentStr, "src/main.go")
	assert.Contains(t, contentStr, "README.md")
	assert.NotContains(t, contentStr, "oldpath") // Old content should be gone

	// The success message goes to stdout, not the captured output
}

func TestInitCommandNoForceFlag(t *testing.T) {
	// This test verifies that without --force, the command would prompt for confirmation
	// Since we can't easily test interactive prompts in unit tests, we'll just verify
	// that the forceInit flag is properly set to false by default
	
	root := setupInitCmd()
	initCommand, _, err := root.Find([]string{"init"})
	require.NoError(t, err)
	
	// Get the force flag
	forceFlag := initCommand.Flags().Lookup("force")
	require.NotNil(t, forceFlag)
	
	// Verify default value
	assert.Equal(t, "false", forceFlag.DefValue)
	assert.Equal(t, "f", forceFlag.Shorthand)
}