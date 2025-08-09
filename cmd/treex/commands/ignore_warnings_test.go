package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestIgnoreWarningsFlag tests that --ignore-warnings suppresses warning output
func TestIgnoreWarningsFlag(t *testing.T) {
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

	// Create .info file with non-existent path
	infoContent := `src/main.go Entry point
src/deleted.go This file doesn't exist
nonexistent/path.txt Another missing file`

	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Test 1: Without --ignore-warnings, warnings should be displayed
	output := &bytes.Buffer{}
	resetGlobalFlags()

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

	// Check that warnings are displayed
	if !strings.Contains(outputStr, "⚠️  Warnings found in .info files:") {
		t.Error("Expected warning header in output without --ignore-warnings")
	}

	if !strings.Contains(outputStr, "Path not found") {
		t.Error("Expected path not found warning without --ignore-warnings")
	}

	// Test 2: With --ignore-warnings, warnings should NOT be displayed
	output.Reset()
	resetGlobalFlags()

	testRootCmd2 := &cobra.Command{Use: "treex"}
	testShowCmd2 := setupShowCmd()
	testRootCmd2.AddCommand(testShowCmd2)

	testRootCmd2.SetOut(output)
	testRootCmd2.SetErr(output)
	testRootCmd2.SetArgs([]string{"show", tempDir, "--format=no-color", "--info-ignore-warnings"})

	err = testRootCmd2.Execute()
	if err != nil {
		t.Fatalf("Command failed with --info-ignore-warnings: %v", err)
	}

	outputStr = output.String()

	// Check that warnings are NOT displayed
	if strings.Contains(outputStr, "⚠️  Warnings found in .info files:") {
		t.Error("Warning header should not be displayed with --info-ignore-warnings")
	}

	if strings.Contains(outputStr, "Path not found") {
		t.Error("Path not found warning should not be displayed with --info-ignore-warnings")
	}

	// But the tree should still be displayed normally
	if !strings.Contains(outputStr, "src/") && !strings.Contains(outputStr, "src") {
		t.Errorf("Expected src/ directory in output even with --info-ignore-warnings, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "main.go") {
		t.Errorf("Expected main.go in output even with --info-ignore-warnings, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Entry point") {
		t.Errorf("Expected annotation in output even with --info-ignore-warnings, got: %s", outputStr)
	}
}

// TestIgnoreWarningsWithMultipleInfoFiles tests --ignore-warnings with nested .info files
func TestIgnoreWarningsWithMultipleInfoFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested structure
	err := os.MkdirAll(filepath.Join(tempDir, "pkg", "util"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Root .info with warning
	rootInfo := `pkg/ Package directory
missing.txt This file doesn't exist`
	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfo), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Nested .info with warning
	pkgInfo := `util/ Utilities
nonexistent.go Missing file`
	err = os.WriteFile(filepath.Join(tempDir, "pkg", ".info"), []byte(pkgInfo), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with --ignore-warnings
	output := &bytes.Buffer{}
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"show", tempDir, "--format=no-color", "--info-ignore-warnings"})

	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	outputStr := output.String()

	// No warnings should be displayed
	if strings.Contains(outputStr, "⚠️  Warnings") {
		t.Error("Warnings should be suppressed with --info-ignore-warnings")
	}

	// But annotations should still work
	if !strings.Contains(outputStr, "Package directory") {
		t.Error("Expected root annotation to be displayed")
	}

	if !strings.Contains(outputStr, "Utilities") {
		t.Error("Expected nested annotation to be displayed")
	}
}

// TestIgnoreWarningsHelp tests that the flag appears in help
func TestIgnoreWarningsHelp(t *testing.T) {
	output := &bytes.Buffer{}
	
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)
	
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"show", "--help"})
	
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}
	
	helpStr := output.String()
	if !strings.Contains(helpStr, "--info-ignore-warnings") {
		t.Error("Expected --info-ignore-warnings flag in help output")
	}
	
	if !strings.Contains(helpStr, "Don't print warnings") {
		t.Error("Expected flag description in help output")
	}
}