package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInfoFileFlagBehavior(t *testing.T) {
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
	if err := os.MkdirAll("docs", 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create both .info and other.txt files in root
	err = os.WriteFile(".info", []byte(`README.md From .info file`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.WriteFile("other.txt", []byte(`README.md From other.txt file`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create both .info and other.txt files in docs
	err = os.WriteFile("docs/.info", []byte(`guide.md From docs/.info file`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.WriteFile("docs/other.txt", []byte(`guide.md From docs/other.txt file`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create actual files
	if err := os.WriteFile("README.md", []byte("# README"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/guide.md", []byte("# Guide"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Default behavior should use .info files
	output, err := executeShowCommand()
	if err != nil {
		t.Errorf("unexpected error with default show command: %v", err)
	}
	
	if !strings.Contains(output, "From .info file") {
		t.Error("Default behavior should show annotations from .info files")
	}
	if strings.Contains(output, "From other.txt file") {
		t.Error("Default behavior should NOT show annotations from other.txt files")
	}
	
	// Test 2: Using --info-file other.txt should ONLY use other.txt files
	output, err = executeShowCommand("--info-file", "other.txt")
	if err != nil {
		t.Errorf("unexpected error with --info-file other.txt: %v", err)
	}
	
	// Check if it shows other.txt annotations
	if !strings.Contains(output, "From other.txt file") {
		t.Error("--info-file other.txt should show annotations from other.txt files")
	}
	
	// Check if it still shows .info annotations (this would be the bug)
	if strings.Contains(output, "From .info file") {
		t.Error("BUG: --info-file other.txt should NOT show annotations from .info files")
	}
}

// executeShowCommand runs the show command with given arguments
func executeShowCommand(args ...string) (string, error) {
	// Reset flags before running command
	outputFormat = "no-color"
	showMode = "mix"
	ignoreFile = ".gitignore"
	noIgnore = false
	infoFile = ".info"
	maxDepth = 10
	verbose = false

	// Create command
	cmd := rootCmd
	cmd.SetArgs(args)

	// Capture output
	var output strings.Builder
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute command
	err := cmd.Execute()
	return output.String(), err
}