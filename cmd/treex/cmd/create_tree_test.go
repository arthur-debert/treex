package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Note: executeCommand and executeCommandC are defined in show_test.go

// resetCreateTreeCmdFlags resets the flags for createTreeCmd to their default values.
// This is necessary because cobra commands reuse flag instances across tests.
func resetCreateTreeCmdFlags() {
	createTreeCmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		// Reset to default value
		f.Value.Set(f.DefValue)
	})
}

func TestCreateTreeCmd_DryRun(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `my-app
├── cmd/ Command line utilities
├── pkg/ Core application code
└── README.md Main documentation`

	err := os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "create-tree", inputFile, tempDir, "--dry-run")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Check output contains expected information
	expectedStrings := []string{
		"DRY RUN",
		"my-app",
		"cmd",
		"pkg",
		"README.md",
		"Would create",
		"directories",
		"files",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}

	// Verify nothing was actually created
	myAppPath := filepath.Join(tempDir, "my-app")
	if _, err := os.Stat(myAppPath); !os.IsNotExist(err) {
		t.Error("expected my-app directory to not exist in dry run mode")
	}
}

func TestCreateTreeCmd_ActualCreation(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `my-app
├── src/ Source code
└── README.md Documentation`

	err := os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	targetDir := filepath.Join(tempDir, "target")

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "create-tree", inputFile, targetDir)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Check output
	expectedStrings := []string{
		"Created file structure",
		"my-app",
		"src",
		"README.md",
		"successfully", // lowercase to match "File structure created successfully!"
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}

	// Verify files were actually created
	expectedPaths := []string{
		filepath.Join(targetDir, "my-app"),
		filepath.Join(targetDir, "my-app", "src"),
		filepath.Join(targetDir, "my-app", "README.md"),
		filepath.Join(targetDir, ".info"),
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected path %s to exist", path)
		}
	}
}

func TestCreateTreeCmd_WithExistingFiles(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()

	// Create existing structure
	existingDir := filepath.Join(tempDir, "my-app", "src")
	err := os.MkdirAll(existingDir, 0755)
	if err != nil {
		t.Fatalf("failed to create existing directory: %v", err)
	}

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `my-app
├── src/ Source code
└── docs/ Documentation`

	err = os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "create-tree", inputFile, tempDir)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Check output mentions skipped files
	if !strings.Contains(output, "Skipped") {
		t.Errorf("expected output to mention skipped files, got:\n%s", output)
	}
}

func TestCreateTreeCmd_Force(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()

	// Create existing file
	existingFile := filepath.Join(tempDir, "my-app", "README.md")
	err := os.MkdirAll(filepath.Dir(existingFile), 0755)
	if err != nil {
		t.Fatalf("failed to create parent directory: %v", err)
	}
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}
	_, err = file.WriteString("existing content")
	if err != nil {
		t.Fatalf("failed to write to existing file: %v", err)
	}
	file.Close()

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `my-app
└── README.md New documentation`

	err = os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "create-tree", inputFile, tempDir, "--force")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Check that file was overwritten (should be empty now)
	fileContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Errorf("failed to read recreated file: %v", err)
	} else if len(fileContent) != 0 {
		t.Errorf("expected recreated file to be empty, got content: %q", string(fileContent))
	}

	// Output should not mention skipped files with --force
	if strings.Contains(output, "Skipped") {
		t.Errorf("expected no skipped files with --force, got:\n%s", output)
	}
}

func TestCreateTreeCmd_NoInfo(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `simple-app
└── main.go Entry point`

	err := os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "create-tree", inputFile, tempDir, "--no-info")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Verify structure was created
	mainGoPath := filepath.Join(tempDir, "simple-app", "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("expected main.go to exist")
	}

	// Verify .info file was NOT created
	infoPath := filepath.Join(tempDir, ".info")
	if _, err := os.Stat(infoPath); !os.IsNotExist(err) {
		t.Error("expected .info file to NOT exist with --no-info flag")
	}

	// Output should not mention .info file creation
	if strings.Contains(output, "Created master .info file") {
		t.Errorf("expected no mention of .info file creation with --no-info, got:\n%s", output)
	}
}

func TestCreateTreeCmd_InvalidInputFile(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command - should fail
	_, err := executeCommand(testRootCmd, "create-tree", nonExistentFile, tempDir)
	if err == nil {
		t.Fatal("expected command to fail with non-existent input file")
	}

	// Check error message
	if !strings.Contains(err.Error(), "failed to parse input file") {
		t.Errorf("expected error about parsing input file, got: %v", err)
	}
}

func TestCreateTreeCmd_InfoFileInput(t *testing.T) {
	resetCreateTreeCmdFlags()
	tempDir := t.TempDir()

	// Create a .info file
	infoContent := `cmd/
Command line utilities

pkg/
Core application code

README.md
Main project documentation`

	infoFile := filepath.Join(tempDir, ".info")
	err := os.WriteFile(infoFile, []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("failed to create .info file: %v", err)
	}

	targetDir := filepath.Join(tempDir, "target")

	// Create test root command with create-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(createTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "create-tree", infoFile, targetDir)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Verify structure was created
	expectedPaths := []string{
		filepath.Join(targetDir, "cmd"),
		filepath.Join(targetDir, "pkg"),
		filepath.Join(targetDir, "README.md"),
		filepath.Join(targetDir, ".info"),
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected path %s to exist", path)
		}
	}

	// Check output
	expectedStrings := []string{
		"Created file structure",
		"cmd",
		"pkg",
		"README.md",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

// Helper function to reset the cobra command between tests
func resetCreateTreeCmd() {
	createTreeCmd.SetArgs([]string{})
	createTreeCmd.SetOut(nil)
	createTreeCmd.SetErr(nil)
}
