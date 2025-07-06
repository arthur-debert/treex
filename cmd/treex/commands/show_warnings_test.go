package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowCommandWithWarnings(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	t.Logf("Test directory: %s", tempDir)
	
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
Invalid line without colon
: Empty path
test/: Empty notes`
	
	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Run the show command
	output := &bytes.Buffer{}
	cmd := rootCmd
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"show", tempDir})
	
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	
	outputStr := output.String()
	
	// Check that tree is displayed
	if !strings.Contains(outputStr, "README.md") {
		t.Error("Expected README.md in output")
	}
	if !strings.Contains(outputStr, "src/") {
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
	if !strings.Contains(outputStr, "Invalid format (missing colon)") {
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
	
	// Ensure exit code is 0 (success)
	if cmd.Execute() != nil {
		t.Error("Expected exit code 0 despite warnings")
	}
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

	// Run the show command
	output := &bytes.Buffer{}
	cmd := rootCmd
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"show", tempDir})
	
	err = cmd.Execute()
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