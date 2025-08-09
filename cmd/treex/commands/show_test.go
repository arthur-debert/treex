package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	// Import pflag
)

// executeCommand is a helper function to execute a cobra command and capture its output.
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

// executeCommandC is a helper function to execute a cobra command and capture its output and error streams.
func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

// resetGlobalFlags resets all global flag variables to their default values
func resetGlobalFlags() {
	verbose = false
	outputFormat = "color"
	infoModeFlag = "mix"
	ignoreFile = ".gitignore"
	noIgnore = false
	infoFile = ".info"
	maxDepth = 10
	infoIgnoreWarnings = false
	overlayPlugins = []string{}
}

// setupShowCmd creates a properly initialized test show command
func setupShowCmd() *cobra.Command {
	// Reset global flag variables
	resetGlobalFlags()

	// Create a clone of the show command to avoid interference
	testShowCmd := &cobra.Command{
		Use:   "show [path...]",
		Short: "Display annotated file tree (default command)",
		Args:  cobra.ArbitraryArgs,
		RunE:  runShowCmd,
	}

	// Add the same flags as the original show command
	testShowCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	testShowCmd.Flags().StringVar(&outputFormat, "format", "color", "Output format")
	testShowCmd.Flags().StringVar(&infoModeFlag, "info-mode", "mix", "View mode: mix, annotated, all")
	testShowCmd.Flags().StringSliceVar(&overlayPlugins, "overlay", []string{}, "Show additional file information")
	testShowCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file")
	testShowCmd.Flags().BoolVar(&noIgnore, "no-ignore", false, "Don't use any ignore file")
	testShowCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	testShowCmd.Flags().BoolVar(&infoIgnoreWarnings, "info-ignore-warnings", false, "Don't print warnings for non-existent paths in .info files")
	testShowCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")

	return testShowCmd
}

func TestShowCmd_VerboseOutput(t *testing.T) {
	tempDir := t.TempDir()
	// Using new format: path:notes
	infoContent := "dummy.txt: Actual Title Line1"
	err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write .info file: %v", err)
	}
	_, err = os.Create(filepath.Join(tempDir, "dummy.txt"))
	if err != nil {
		t.Fatalf("Failed to create dummy.txt: %v", err)
	}

	// Reset global flags for a clean test environment
	resetGlobalFlags()

	// Create test root command and add our test show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	// Execute the show command with verbose flag
	output, err := executeCommand(testRootCmd, "show", tempDir, "-v", "--format=no-color")
	if err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}

	// Check for verbose output strings
	expectedVerboseStrings := []string{
		"Analyzing directory:",
		"Verbose mode enabled",
		"=== Parsed Annotations ===",
		"Path: dummy.txt",
		"  Notes: Actual Title Line1",
		"=== End Annotations ===",
		"=== File Tree Structure ===",
		"dummy.txt (file) [Actual Title Line1]",
		"=== End Tree Structure ===",
		"treex analysis of:",
		"Found 1 annotations",
	}

	for _, s := range expectedVerboseStrings {
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain verbose string: %q\nFull output:\n%s", s, output)
		}
	}

	// Check that the main tree output is also present
	if !strings.Contains(output, "dummy.txt") {
		t.Errorf("Expected output to contain the tree item 'dummy.txt'.\nFull output:\n%s", output)
	}

	// The no-color renderer output for an item with a title.
	// The exact spacing might vary, so check for presence of file and title.
	if !strings.Contains(output, "dummy.txt") || !strings.Contains(output, "Actual Title Line1") {
		t.Errorf("Expected output to contain rendered 'dummy.txt' and 'Actual Title Line1'.\nFull output:\n%s", output)
	}
	// More precise check if needed: "dummy.txt                           Actual Title Line1"
}

func TestShowCmd_NonVerboseOutput(t *testing.T) {
	tempDir := t.TempDir()
	// Using colon format
	infoContent := "file.txt: My File Title Line1\nMy File Description Line2"
	err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write .info file: %v", err)
	}
	_, err = os.Create(filepath.Join(tempDir, "file.txt"))
	if err != nil {
		t.Fatalf("Failed to create file.txt: %v", err)
	}

	// Reset global flags for a clean test environment
	resetGlobalFlags()

	// Create test root command and add our test show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	// Execute the show command
	output, err := executeCommand(testRootCmd, "show", tempDir, "--format=no-color")
	if err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}

	// Check that verbose output strings are NOT present
	unexpectedVerboseStrings := []string{
		"Analyzing directory:",
		"Verbose mode enabled",
		"=== Parsed Annotations ===",
	}

	for _, s := range unexpectedVerboseStrings {
		if strings.Contains(output, s) {
			t.Errorf("Expected output NOT to contain verbose string: %q\nFull output:\n%s", s, output)
		}
	}

	// Check that the main tree output is present and shows the title
	containsFile := strings.Contains(output, "file.txt")
	containsTitle := strings.Contains(output, "My File Title Line1")

	if !containsFile || !containsTitle {
		t.Errorf("Output check failed. Contains 'file.txt': %t, Contains 'My File Title Line1': %t.\nFull output:\n---\n%s\n---", containsFile, containsTitle, output)
	}
}

func TestShowCmd_MultiplePaths(t *testing.T) {
	// Create two test directories with different .info files
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	// Setup first directory
	infoContent1 := "file1.txt: First directory file"
	err := os.WriteFile(filepath.Join(tempDir1, ".info"), []byte(infoContent1), 0644)
	if err != nil {
		t.Fatalf("Failed to write .info file in tempDir1: %v", err)
	}
	_, err = os.Create(filepath.Join(tempDir1, "file1.txt"))
	if err != nil {
		t.Fatalf("Failed to create file1.txt: %v", err)
	}

	// Setup second directory
	infoContent2 := "file2.txt: Second directory file"
	err = os.WriteFile(filepath.Join(tempDir2, ".info"), []byte(infoContent2), 0644)
	if err != nil {
		t.Fatalf("Failed to write .info file in tempDir2: %v", err)
	}
	_, err = os.Create(filepath.Join(tempDir2, "file2.txt"))
	if err != nil {
		t.Fatalf("Failed to create file2.txt: %v", err)
	}

	// Reset global flags for a clean test environment
	resetGlobalFlags()

	// Create test root command and add our test show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	// Execute the show command with multiple paths
	output, err := executeCommand(testRootCmd, "show", tempDir1, tempDir2, "--format=no-color")
	if err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}

	// Check that both directories are present in output
	if !strings.Contains(output, "file1.txt") {
		t.Errorf("Expected output to contain 'file1.txt' from first directory.\nFull output:\n%s", output)
	}

	if !strings.Contains(output, "file2.txt") {
		t.Errorf("Expected output to contain 'file2.txt' from second directory.\nFull output:\n%s", output)
	}

	// Check that both annotations are present
	if !strings.Contains(output, "First directory file") {
		t.Errorf("Expected output to contain 'First directory file' annotation.\nFull output:\n%s", output)
	}

	if !strings.Contains(output, "Second directory file") {
		t.Errorf("Expected output to contain 'Second directory file' annotation.\nFull output:\n%s", output)
	}

	// Check that both directory names appear as root nodes (like Unix tree command)
	dir1Name := filepath.Base(tempDir1)
	dir2Name := filepath.Base(tempDir2)

	if !strings.Contains(output, dir1Name) {
		t.Errorf("Expected output to contain directory name '%s'.\nFull output:\n%s", dir1Name, output)
	}

	if !strings.Contains(output, dir2Name) {
		t.Errorf("Expected output to contain directory name '%s'.\nFull output:\n%s", dir2Name, output)
	}

	// Verify that there's separation between the two trees (multiple trees should have some separation)
	// Count occurrences to ensure both trees are rendered
	file1Count := strings.Count(output, "file1.txt")
	file2Count := strings.Count(output, "file2.txt")

	if file1Count != 1 {
		t.Errorf("Expected exactly 1 occurrence of 'file1.txt', got %d", file1Count)
	}

	if file2Count != 1 {
		t.Errorf("Expected exactly 1 occurrence of 'file2.txt', got %d", file2Count)
	}
}

func TestShowCmd_SinglePath(t *testing.T) {
	// Test single path behavior
	tempDir := t.TempDir()
	infoContent := "test.txt: Test file"
	err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write .info file: %v", err)
	}
	_, err = os.Create(filepath.Join(tempDir, "test.txt"))
	if err != nil {
		t.Fatalf("Failed to create test.txt: %v", err)
	}

	// Reset global flags for a clean test environment
	resetGlobalFlags()

	// Create test root command and add our test show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	// Execute the show command with single path (existing behavior)
	output, err := executeCommand(testRootCmd, "show", tempDir, "--format=no-color")
	if err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}

	// Check that output contains expected content
	if !strings.Contains(output, "test.txt") {
		t.Errorf("Expected output to contain 'test.txt'.\nFull output:\n%s", output)
	}

	if !strings.Contains(output, "Test file") {
		t.Errorf("Expected output to contain 'Test file' annotation.\nFull output:\n%s", output)
	}
}

func TestShowCmd_EmptyArgs(t *testing.T) {
	// Test that no arguments still defaults to current directory
	// Reset global flags for a clean test environment
	resetGlobalFlags()

	// Create test root command and add our test show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	// Execute the show command with no arguments (should default to ".")
	output, err := executeCommand(testRootCmd, "show", "--format=no-color")
	if err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}

	// Should not error and should produce some output
	if len(output) == 0 {
		t.Error("Expected some output when running with no arguments")
	}
}
