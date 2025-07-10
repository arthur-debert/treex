package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// setupSyncCmd creates a properly initialized test sync command
func setupSyncCmd() *cobra.Command {
	// Reset flags
	forceSync = false
	infoFile = ".info"

	// Create a test root command
	testRootCmd := &cobra.Command{
		Use:   "treex",
		Short: "Test root command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add the sync command
	testRootCmd.AddCommand(syncCmd)

	return testRootCmd
}

// executeSyncCommand is a helper function to execute the sync command
func executeSyncCommand(args ...string) (output string, err error) {
	root := setupSyncCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(append([]string{"sync"}, args...))

	_, err = root.ExecuteC()
	return buf.String(), err
}

func TestSyncCommandNoInfoFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Execute sync command
	output, err := executeSyncCommand()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No .info files found") {
		t.Errorf("expected 'No .info files found', got: %s", output)
	}
}

func TestSyncCommandNoStaleAnnotations(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create some files and directories
	if err := os.Mkdir("src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("src/main.go", []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("docs", 0755); err != nil {
		t.Fatal(err)
	}

	// Create .info file with valid paths
	infoContent := `src Main source directory
src/main.go Main entry point
docs Documentation`

	err = os.WriteFile(".info", []byte(infoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Execute sync command
	output, err := executeSyncCommand()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No stale annotations found") {
		t.Errorf("expected 'No stale annotations found', got: %s", output)
	}
}

func TestSyncCommandWithStaleAnnotations(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create some files but not all
	if err := os.Mkdir("src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("src/main.go", []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	// docs directory doesn't exist
	// tests directory doesn't exist

	// Create .info file with some stale paths
	infoContent := `src Main source directory
src/main.go Main entry point
docs Documentation directory
tests Test files`

	err = os.WriteFile(".info", []byte(infoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Execute sync command with --force
	output, err := executeSyncCommand("--force")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check output
	if !strings.Contains(output, "Found 2 stale annotations") {
		t.Errorf("expected to find 2 stale annotations in output: %s", output)
	}

	if !strings.Contains(output, "for docs annotation") {
		t.Errorf("expected 'for docs annotation' in output: %s", output)
	}

	if !strings.Contains(output, "for tests annotation") {
		t.Errorf("expected 'for tests annotation' in output: %s", output)
	}

	if !strings.Contains(output, "Updated .info") {
		t.Errorf("expected 'Updated .info' in output: %s", output)
	}

	// Verify the .info file was updated correctly
	content, err := os.ReadFile(".info")
	if err != nil {
		t.Fatal(err)
	}

	// Should still have valid annotations
	if !strings.Contains(string(content), "src Main source directory") {
		t.Error("src annotation should remain")
	}
	if !strings.Contains(string(content), "src/main.go Main entry point") {
		t.Error("src/main.go annotation should remain")
	}

	// Should not have stale annotations
	if strings.Contains(string(content), "docs Documentation") {
		t.Error("docs annotation should have been removed")
	}
	if strings.Contains(string(content), "tests Test files") {
		t.Error("tests annotation should have been removed")
	}
}

func TestSyncCommandMultipleInfoFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create directory structure
	if err := os.Mkdir("src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("src/main.go", []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("pkg", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("pkg/core", 0755); err != nil {
		t.Fatal(err)
	}

	// Create root .info file with stale annotation
	rootInfo := `src Source code
docs Documentation
pkg Packages`
	if err := os.WriteFile(".info", []byte(rootInfo), 0644); err != nil {
		t.Fatal(err)
	}

	// Create pkg/.info file with stale annotation
	pkgInfo := `core Core functionality
utils Utility functions`
	if err := os.WriteFile("pkg/.info", []byte(pkgInfo), 0644); err != nil {
		t.Fatal(err)
	}

	// Execute sync command with --force
	output, err := executeSyncCommand("--force")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Check that both .info files were processed
	if !strings.Contains(output, ".info: for docs annotation") {
		t.Errorf("expected root .info file stale annotation in output: %s", output)
	}

	if !strings.Contains(output, filepath.Join("pkg", ".info")+": for utils annotation") {
		t.Errorf("expected pkg/.info file stale annotation in output: %s", output)
	}

	// Verify both files were updated
	if !strings.Contains(output, "Updated .info") {
		t.Errorf("expected root .info to be updated: %s", output)
	}

	if !strings.Contains(output, "Updated "+filepath.Join("pkg", ".info")) {
		t.Errorf("expected pkg/.info to be updated: %s", output)
	}
}
