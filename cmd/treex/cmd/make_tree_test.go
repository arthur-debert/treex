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

func TestMakeTreeCmd_DryRun(t *testing.T) {
	resetMakeTreeCmdFlags()
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

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", inputFile, tempDir, "--dry-run")
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

func TestMakeTreeCmd_ActualCreation(t *testing.T) {
	resetMakeTreeCmdFlags()
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

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", inputFile, targetDir)
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

func TestMakeTreeCmd_WithExistingFiles(t *testing.T) {
	resetMakeTreeCmdFlags()
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

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", inputFile, tempDir)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Check output mentions skipped files
	if !strings.Contains(output, "Skipped") {
		t.Errorf("expected output to mention skipped files, got:\n%s", output)
	}
}

func TestMakeTreeCmd_Force(t *testing.T) {
	resetMakeTreeCmdFlags()
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
	if err = file.Close(); err != nil {
		t.Fatalf("failed to close existing file: %v", err)
	}

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `my-app
└── README.md New documentation`

	err = os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", inputFile, tempDir, "--force")
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

func TestMakeTreeCmd_NoInfo(t *testing.T) {
	resetMakeTreeCmdFlags()
	tempDir := t.TempDir()

	// Create input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `simple-app
└── main.go Entry point`

	err := os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", inputFile, tempDir, "--no-info")
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

func TestMakeTreeCmd_InvalidInputFile(t *testing.T) {
	resetMakeTreeCmdFlags()
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command - should fail
	_, err := executeCommand(testRootCmd, "make-tree", nonExistentFile, tempDir)
	if err == nil {
		t.Fatal("expected command to fail with non-existent input file")
	}

	// Check error message
	if !strings.Contains(err.Error(), "failed to parse input file") {
		t.Errorf("expected error about parsing input file, got: %v", err)
	}
}

func TestMakeTreeCmd_InfoFileInput(t *testing.T) {
	resetMakeTreeCmdFlags()
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

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Execute command
	output, err := executeCommand(testRootCmd, "make-tree", infoFile, targetDir)
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

func TestMakeTreeCmd_StdinInput(t *testing.T) {
	resetMakeTreeCmdFlags()
	tempDir := t.TempDir()

	content := `stdin-app
├── api/ REST API endpoints
├── web/ Frontend assets
└── config.yaml Configuration`

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	os.Stdin = r

	// Write content to stdin in a goroutine
	go func() {
		defer func() {
			_ = w.Close()
		}()
		if _, err := w.WriteString(content); err != nil {
			t.Errorf("failed to write to stdin: %v", err)
		}
	}()

	// Execute command with no input file (should read from stdin)
	output, err := executeCommand(testRootCmd, "make-tree", tempDir)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Verify structure was created
	expectedPaths := []string{
		filepath.Join(tempDir, "stdin-app"),
		filepath.Join(tempDir, "stdin-app", "api"),
		filepath.Join(tempDir, "stdin-app", "web"),
		filepath.Join(tempDir, "stdin-app", "config.yaml"),
		filepath.Join(tempDir, ".info"),
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected path %s to exist", path)
		}
	}

	// Check output contains expected information
	expectedStrings := []string{
		"Created file structure",
		"stdin-app",
		"api",
		"web",
		"config.yaml",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestMakeTreeCmd_ExplicitStdinDash(t *testing.T) {
	resetMakeTreeCmdFlags()
	tempDir := t.TempDir()

	content := `dash-app
└── main.py Entry point`

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	os.Stdin = r

	// Write content to stdin in a goroutine
	go func() {
		defer func() {
			_ = w.Close()
		}()
		if _, err := w.WriteString(content); err != nil {
			t.Errorf("failed to write to stdin: %v", err)
		}
	}()

	targetDir := filepath.Join(tempDir, "target")

	// Execute command with explicit stdin marker "-"
	output, err := executeCommand(testRootCmd, "make-tree", "-", targetDir)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Verify structure was created
	expectedPaths := []string{
		filepath.Join(targetDir, "dash-app"),
		filepath.Join(targetDir, "dash-app", "main.py"),
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
		"dash-app",
		"main.py",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestMakeTreeCmd_StdinDryRun(t *testing.T) {
	resetMakeTreeCmdFlags()
	tempDir := t.TempDir()

	content := `dry-app
└── test.txt Test file`

	// Create test root command with make-tree command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(makeTreeCmd)

	// Create a pipe to simulate stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	os.Stdin = r

	// Write content to stdin in a goroutine
	go func() {
		defer func() {
			_ = w.Close()
		}()
		if _, err := w.WriteString(content); err != nil {
			t.Errorf("failed to write to stdin: %v", err)
		}
	}()

	// Execute command with dry-run flag
	output, err := executeCommand(testRootCmd, "make-tree", "-", tempDir, "--dry-run")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Check output contains dry run information
	expectedStrings := []string{
		"DRY RUN",
		"dry-app",
		"test.txt",
		"Would create",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}

	// Verify nothing was actually created
	dryAppPath := filepath.Join(tempDir, "dry-app")
	if _, err := os.Stat(dryAppPath); !os.IsNotExist(err) {
		t.Error("expected dry-app directory to not exist in dry run mode")
	}
}
