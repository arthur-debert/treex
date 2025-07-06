package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/spf13/cobra"
)

func TestShowCommandWithWarnings(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	
	// Create some files
	err := os.MkdirAll(filepath.Join(tempDir, "src"), 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "src", "main.go"), []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# README"), 0644)
	if err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}
	
	// Create test directory for empty notes warning
	err = os.MkdirAll(filepath.Join(tempDir, "test"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create .info file with various issues
	infoContent := `README.md: Project documentation
src/main.go: Entry point
src/deleted.go: This file doesn't exist
Invalid
: Empty path
test/: Empty notes`
	
	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Run the show command using proper test setup
	output := &bytes.Buffer{}
	
	// Reset global flags to avoid state issues
	resetGlobalFlags()
	
	// Create a new root command for testing
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)
	
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"show", tempDir, "--format=no-color"})
	
	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	outputStr := output.String()
	
	// Check that tree is displayed
	if !strings.Contains(outputStr, "README.md") {
		t.Error("Expected README.md in output")
	}
	if !strings.Contains(outputStr, "src/") || !strings.Contains(outputStr, "src") {
		t.Error("Expected src/ directory in output")
	}
	if !strings.Contains(outputStr, "main.go") {
		t.Error("Expected main.go in output")
	}
	
	// Check that annotations are displayed
	if !strings.Contains(outputStr, "Project documentation") {
		t.Error("Expected README.md annotation in output")
	}
	if !strings.Contains(outputStr, "Entry point") {
		t.Error("Expected main.go annotation in output")
	}
	
	// Check that warnings are displayed
	if !strings.Contains(outputStr, "⚠️  Warnings found in .info files:") {
		t.Error("Expected warning header in output")
	}
	
	// Check for specific warnings
	if !strings.Contains(outputStr, "Invalid format (missing annotation)") {
		t.Error("Expected invalid format warning")
	}
	if !strings.Contains(outputStr, "Path not found") && !strings.Contains(outputStr, "src/deleted.go") {
		t.Error("Expected non-existent path warning")
	}
	if !strings.Contains(outputStr, "Empty path") {
		t.Error("Expected empty path warning")
	}
	if !strings.Contains(outputStr, "Empty notes") {
		t.Errorf("Expected empty notes warning. Output:\n%s", outputStr)
	}
	
	// The command should have succeeded despite warnings
}

func TestShowCommandNoWarnings(t *testing.T) {
	// Create a test directory structure with valid .info
	tempDir := t.TempDir()
	
	// Create some files
	err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# README"), 0644)
	if err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	// Create valid .info file
	infoContent := `README.md: Project documentation`
	
	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Run the show command using proper test setup
	output := &bytes.Buffer{}
	
	// Reset global flags to avoid state issues
	resetGlobalFlags()
	
	// Create a new root command for testing
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)
	
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"show", tempDir, "--format=no-color"})
	
	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	outputStr := output.String()
	
	// Check that tree is displayed
	if !strings.Contains(outputStr, "README.md") {
		t.Error("Expected README.md in output")
	}
	
	// Check that no warnings are displayed
	if strings.Contains(outputStr, "⚠️  Warnings found in .info files:") {
		t.Error("Did not expect warning header when there are no warnings")
	}
}