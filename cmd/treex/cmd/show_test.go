package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag" // Import pflag
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

	// Reset global flags for a clean test environment for showCmd
	// This is important because cobra flags can persist across test runs.
	resetShowCmdFlags()

	// Execute the show command with verbose flag
	// Note: We are testing the `showCmd` directly.
	// If `rootCmd` has persistent flags that `showCmd` relies on,
	// they might need to be set up here or `rootCmd` should be executed.
	// For this test, we assume `showCmd` can be tested in isolation or `rootCmd` is simple.
	// To be absolutely sure, one might need to reinitialize rootCmd or use rootCmd.ExecuteC()

	// Re-initialize root command and add show command to it for each test run
	// to ensure flag states are clean.
	testRootCmd := setupTestRootCommand()
	testRootCmd.AddCommand(showCmd) // Add the actual showCmd

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

// resetShowCmdFlags resets the flags for showCmd to their default values.
// This is necessary because cobra commands reuse flag instances.
func resetShowCmdFlags() {
	verbose = false // Assuming verbose is the flag variable for -v
	path = ""
	outputFormat = "color"    // Default value
	ignoreFile = ".gitignore" // Default value
	maxDepth = 10             // Default value
	safeMode = false

	// If flags are defined on showCmd directly, reset them:
	// showCmd.Flags().VisitAll(func(f *pflag.Flag) {
	// 	f.Value.Set(f.DefValue)
	// })
	// However, these are global vars in the current codebase.
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

	resetShowCmdFlags()
	// Attempt to more thoroughly reset flag states on the shared showCmd object
	showCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		// For flags that might not have a DefValue correctly set by default after parsing,
		// or if their value isn't reset by just setting the bound variable,
		// explicitly set them to their default string value.
		// This is more of a diagnostic step.
		// f.Value.Set(f.DefValue) // This can be problematic if DefValue isn't what we expect now
	})

	testRootCmd := setupTestRootCommand()
	testRootCmd.AddCommand(showCmd) // ensure showCmd is part of this specific test's root

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
	// A more precise check if the renderer's behavior is stable for no-color:
	// e.g. if !strings.Contains(output, "file.txt                            My File Title Line1")
}
