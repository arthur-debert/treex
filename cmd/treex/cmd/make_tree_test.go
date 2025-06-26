package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Note: executeCommand and executeCommandC are defined in show_test.go

// setupTestRootCommand creates a test root command with proper command groups
func setupTestRootCommand() *cobra.Command {
	// Reset global variables used by commands
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}

	// Add the same command groups as the real root command
	testRootCmd.AddGroup(&cobra.Group{
		ID:    "main",
		Title: "Edit Annotations",
	})
	testRootCmd.AddGroup(&cobra.Group{
		ID:    "info",
		Title: "Info Files (.info) files:",
	})
	testRootCmd.AddGroup(&cobra.Group{
		ID:    "filesystem",
		Title: "File-system:",
	})
	testRootCmd.AddGroup(&cobra.Group{
		ID:    "help",
		Title: "Help and learning:",
	})

	return testRootCmd
}

// resetMakeTreeCmdFlags resets the flags for makeTreeCmd to their default values.
// This is necessary because cobra commands reuse flag instances across tests.
func resetMakeTreeCmdFlags() {
	makeTreeCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		// Reset to default value
		if err := f.Value.Set(f.DefValue); err != nil {
			// Flag reset error - this shouldn't happen in tests but we handle it
			// to satisfy the linter. In practice, this would indicate a bug in the test setup.
			panic(fmt.Sprintf("failed to reset flag %s: %v", f.Name, err))
		}
	})
}

// TestMakeTreeCmd_ArgumentParsing tests that the command correctly parses different argument combinations
func TestMakeTreeCmd_ArgumentParsing(t *testing.T) {
	// Create temporary directory for the test
	tempDir := t.TempDir()

	// Create a simple input file
	inputFile := filepath.Join(tempDir, "input.txt")
	err := os.WriteFile(inputFile, []byte("test-app\n└── main.go"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Create test root command with make-tree command
	testRootCmd := setupTestRootCommand()
	testRootCmd.AddCommand(makeTreeCmd)

	// Test cases for different argument combinations
	testCases := []struct {
		name          string
		args          []string
		expectSuccess bool
	}{
		{
			name:          "input file and target directory",
			args:          []string{"make-tree", inputFile, tempDir},
			expectSuccess: true,
		},
		{
			name:          "just input file (target defaults to .)",
			args:          []string{"make-tree", inputFile},
			expectSuccess: true,
		},
		{
			name:          "non-existent input file",
			args:          []string{"make-tree", filepath.Join(tempDir, "nonexistent.txt")},
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetMakeTreeCmdFlags()
			resetGlobalFlags()
			output, err := executeCommand(testRootCmd, tc.args...)

			if tc.expectSuccess && err != nil {
				t.Errorf("Expected command to succeed, but got error: %v", err)
			}
			if !tc.expectSuccess && err == nil {
				t.Errorf("Expected command to fail, but it succeeded with output: %s", output)
			}
		})
	}
}

// TestMakeTreeCmd_OutputFormat tests that the command produces the expected output format
func TestMakeTreeCmd_OutputFormat(t *testing.T) {
	// Create temporary directory for the test
	tempDir := t.TempDir()

	// Create a simple input file
	inputFile := filepath.Join(tempDir, "input.txt")
	err := os.WriteFile(inputFile, []byte("test-app\n└── main.go"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Create test root command with make-tree command
	testRootCmd := setupTestRootCommand()
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", inputFile, tempDir)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Check that output follows expected format
	if !strings.Contains(output, "The following files/dirs were created:") {
		t.Errorf("Expected output to contain created files header, got:\n%s", output)
	}

	// Check for bold formatting codes
	if !strings.Contains(output, "\033[1m") && !strings.Contains(output, "\033[0m") {
		t.Errorf("Expected output to use bold formatting, got:\n%s", output)
	}
}

// TestMakeTreeCmd_StdinInput tests that the command correctly reads from stdin
func TestMakeTreeCmd_StdinInput(t *testing.T) {
	// Create temporary directory for the test
	tempDir := t.TempDir()

	// Prepare stdin input
	content := "test-app\n└── main.go"

	// Create test root command with make-tree command
	testRootCmd := setupTestRootCommand()
	testRootCmd.AddCommand(makeTreeCmd)

	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	os.Stdin = r

	// Write content to stdin in a goroutine
	go func() {
		defer func() { _ = w.Close() }()
		if _, err := w.WriteString(content); err != nil {
			t.Errorf("Failed to write to stdin: %v", err)
		}
	}()

	// Test stdin with explicit "-" marker
	output, err := executeCommand(testRootCmd, "make-tree", "-", tempDir)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Check that output indicates files were created
	if !strings.Contains(output, "test-app") {
		t.Errorf("Expected output to contain created files, got:\n%s", output)
	}
}

// TestMakeTreeCmd_InfoFileHeader tests that the command passes the custom header to the info file creation
func TestMakeTreeCmd_InfoFileHeader(t *testing.T) {
	// Create temporary directory for the test
	tempDir := t.TempDir()

	// Create a simple input file
	inputFile := filepath.Join(tempDir, "input.txt")
	err := os.WriteFile(inputFile, []byte("test-app\n└── main.go"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Create test root command with make-tree command
	testRootCmd := setupTestRootCommand()
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	_, err = executeCommand(testRootCmd, "make-tree", inputFile, tempDir)
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Read the created .info file
	infoPath := filepath.Join(tempDir, ".info")
	infoContent, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to read .info file: %v", err)
	}

	// Check that the custom header is present
	infoText := string(infoContent)
	if !strings.Contains(infoText, "Now you can document your project, go crazy!") {
		t.Errorf("Expected .info file to contain the custom header, got:\n%s", infoText)
	}
}
