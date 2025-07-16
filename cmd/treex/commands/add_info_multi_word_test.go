package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeAddInfoCommand is a helper function to execute the add command
func executeAddInfoCommand(args ...string) (output string, err error) {
	// Create a test root command
	testRootCmd := &cobra.Command{
		Use:   "treex",
		Short: "Test root command",
	}
	
	// Create add command with flags
	testAddCmd := &cobra.Command{
		Use:   "add <path> <description>",
		Short: "Add or update an entry in the current directory's .info file",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runAddInfoCmd,
	}
	testAddCmd.Flags().Bool("replace", false, "Replace existing entry without prompting")
	
	// Add the command
	testRootCmd.AddCommand(testAddCmd)
	
	buf := new(bytes.Buffer)
	testRootCmd.SetOut(buf)
	testRootCmd.SetErr(buf)
	testRootCmd.SetArgs(args)

	_, err = testRootCmd.ExecuteC()
	return buf.String(), err
}

func TestAddInfoMultiWordDescription(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	// Create the directory/file for testing
	err = os.MkdirAll("pkg", 0755)
	require.NoError(t, err)
	
	// Test multi-word description without quotes
	description := "Main package containing core functionality"
	args := append([]string{"add", "pkg"}, strings.Fields(description)...)
	
	// Execute the command with multiple arguments
	_, err = executeAddInfoCommand(args...)
	
	// This should work without error
	require.NoError(t, err)
	
	// Check that the .info file was created with the correct content
	content, err := os.ReadFile(".info")
	require.NoError(t, err)
	
	// The content should contain the full description
	contentStr := string(content)
	assert.Contains(t, contentStr, "pkg")
	assert.Contains(t, contentStr, description)
}

func TestAddInfoMultiWordWithFlags(t *testing.T) {
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
	existingContent := "pkg Old description\n"
	err = os.WriteFile(".info", []byte(existingContent), 0644)
	require.NoError(t, err)

	// Create the directory/file for testing
	err = os.MkdirAll("pkg", 0755)
	require.NoError(t, err)
	
	// Test updating with --replace flag and multi-word description
	description := "Updated package with new functionality"
	args := append([]string{"add", "--replace", "pkg"}, strings.Fields(description)...)
	
	// Execute the command
	_, err = executeAddInfoCommand(args...)
	require.NoError(t, err)
	
	// Check that the description was updated
	content, err := os.ReadFile(".info")
	require.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "pkg")
	assert.Contains(t, contentStr, description)
	assert.NotContains(t, contentStr, "Old description")
}

func TestAddInfoBackwardsCompatibility(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	// Create the directory/file for testing
	err = os.MkdirAll("pkg", 0755)
	require.NoError(t, err)
	
	// Test that quoted strings still work (backwards compatibility)
	description := "Main package containing core functionality"
	args := []string{"add", "pkg", description} // This simulates a quoted string passed as single arg
	
	// Execute the command
	_, err = executeAddInfoCommand(args...)
	require.NoError(t, err)
	
	// Check that the description is correct
	content, err := os.ReadFile(".info")
	require.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "pkg")
	assert.Contains(t, contentStr, description)
}

func TestAddInfoSingleWord(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	// Create the directory/file for testing
	err = os.MkdirAll("pkg", 0755)
	require.NoError(t, err)
	
	// Test single word description (should still work)
	args := []string{"add", "pkg", "Package"}
	
	// Execute the command
	_, err = executeAddInfoCommand(args...)
	require.NoError(t, err)
	
	// Check that the description is correct
	content, err := os.ReadFile(".info")
	require.NoError(t, err)
	
	contentStr := string(content)
	assert.Contains(t, contentStr, "pkg")
	assert.Contains(t, contentStr, "Package")
}