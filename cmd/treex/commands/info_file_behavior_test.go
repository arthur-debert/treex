package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestInfoFileBehaviorIsolation verifies that --info-file completely replaces .info file lookup
func TestInfoFileBehaviorIsolation(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-info-test-*")
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
	if err := os.MkdirAll("docs", 0755); err != nil {
		t.Fatal(err)
	}

	// Create .info files with one set of annotations
	rootInfoContent := `main.go Main entry point
README.md Project documentation`
	
	docsInfoContent := `api.md API documentation
guide.md User guide`
	
	if err := os.WriteFile(".info", []byte(rootInfoContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/.info", []byte(docsInfoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create other.txt files with different annotations
	rootOtherContent := `app.go Application logic
config.go Configuration`
	
	docsOtherContent := `tutorial.md Tutorial
examples.md Examples`
	
	if err := os.WriteFile("other.txt", []byte(rootOtherContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/other.txt", []byte(docsOtherContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Search with default .info files
	infoFile = ".info" // Reset to default
	output, err := executeSearchCommand("main")
	if err != nil {
		t.Errorf("search with default .info failed: %v", err)
	}
	
	if !strings.Contains(output, "Found 1 matches for 'main'") {
		t.Errorf("expected to find 'main' in .info files, got: %s", output)
	}

	// Test 2: Search with --info-file other.txt should NOT find .info content
	infoFile = ".info" // Reset before command
	output, err = executeSearchCommand("--info-file", "other.txt", "main")
	if err != nil {
		t.Errorf("search with other.txt failed: %v", err)
	}
	
	if strings.Contains(output, "Found 1 matches for 'main'") {
		t.Errorf("BUG: found 'main' when using other.txt, should only search other.txt files")
	}

	// Test 3: Search with --info-file other.txt should find other.txt content
	infoFile = ".info" // Reset before command
	output, err = executeSearchCommand("--info-file", "other.txt", "application")
	if err != nil {
		t.Errorf("search with other.txt failed: %v", err)
	}
	
	if !strings.Contains(output, "Found 1 matches for 'application'") {
		t.Errorf("expected to find 'application' in other.txt files, got: %s", output)
	}

	// Test 4: Verify with show command - create actual files to avoid warnings
	if err := os.WriteFile("main.go", []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("app.go", []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	// Show with default .info should show main.go annotation
	output, err = executeShowCommand("--format", "no-color", "--show", "annotated")
	if err != nil {
		t.Errorf("show with default .info failed: %v", err)
	}
	
	if !strings.Contains(output, "main.go") || !strings.Contains(output, "Main entry point") {
		t.Errorf("expected to see main.go annotation from .info file, got: %s", output)
	}

	// Show with --info-file other.txt should show app.go annotation
	infoFile = ".info" // Reset before command
	output, err = executeShowCommand("--format", "no-color", "--show", "annotated", "--info-file", "other.txt")
	if err != nil {
		t.Errorf("show with other.txt failed: %v", err)
	}
	
	if !strings.Contains(output, "app.go") || !strings.Contains(output, "Application logic") {
		t.Errorf("expected to see app.go annotation from other.txt file, got: %s", output)
	}
	
	if strings.Contains(output, "main.go") && strings.Contains(output, "Main entry point") {
		t.Errorf("BUG: should not see main.go annotation when using other.txt, got: %s", output)
	}
}

// executeShowCommandForBehaviorTest helper function for testing show command
func executeShowCommandForBehaviorTest(args ...string) (string, error) {
	// Create a fresh show command for testing
	testShowCmd := setupShowCmd()
	
	// Create a root command and add show command
	testRootCmd := &cobra.Command{Use: "treex"}
	testRootCmd.AddCommand(testShowCmd)
	
	// Capture output
	output := &bytes.Buffer{}
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	
	// Set arguments (prepend "show" to the args)
	testRootCmd.SetArgs(append([]string{"show"}, args...))
	
	// Execute command
	err := testRootCmd.Execute()
	return output.String(), err
}