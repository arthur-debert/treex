package cmd

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
	path = ""
	outputFormat = "color"
	ignoreFile = ".gitignore"
	maxDepth = 10
	safeMode = false
}

// setupShowCmd creates a properly initialized test show command
func setupShowCmd() *cobra.Command {
	// Reset global flag variables
	resetGlobalFlags()

	// Create a clone of the show command to avoid interference
	testShowCmd := &cobra.Command{
		Use:   "show [path]",
		Short: "Display annotated file tree (default command)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runShowCmd,
	}

	// Add the same flags as the original show command
	testShowCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	testShowCmd.Flags().StringVarP(&path, "path", "p", "", "Path to analyze")
	testShowCmd.Flags().StringVar(&outputFormat, "format", "color", "Output format")
	testShowCmd.Flags().StringVar(&ignoreFile, "use-ignore-file", ".gitignore", "Use specified ignore file")
	testShowCmd.Flags().IntVarP(&maxDepth, "depth", "d", 10, "Maximum depth to traverse")
	testShowCmd.Flags().BoolVar(&safeMode, "safe-mode", false, "Force safe terminal rendering mode")

	return testShowCmd
}

func TestShowCmd_VerboseOutput(t *testing.T) {
	tempDir := t.TempDir()
	// Using compact format: path and title on the same line
	infoContent := "dummy.txt Actual Title Line1\nActual Description Line2"
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
		"  Title: Actual Title Line1",
		"  Description: Actual Title Line1\nActual Description Line2", // Full description includes title line
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
	// Using compact format: path and title on the same line
	infoContent := "file.txt My File Title Line1\nMy File Description Line2"
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
