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
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create .info file with non-existent paths
	infoContent := `nonexistent.txt This file doesn't exist
missing-dir/ This directory doesn't exist
real.txt This is a real file`
	if err := os.WriteFile(".info", []byte(infoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the real file
	if err := os.WriteFile("real.txt", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test that warnings are shown by default
	cmd := newTestShowCommand()
	output, err := executeCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that warnings are present
	if !strings.Contains(output, "⚠️  Warnings found in .info files:") {
		t.Error("expected warnings header")
	}
	if !strings.Contains(output, "Path not found: \"nonexistent.txt\"") {
		t.Error("expected warning for nonexistent.txt")
	}
	if !strings.Contains(output, "Path not found: \"missing-dir\"") {
		t.Error("expected warning for missing-dir")
	}
}

func TestShowCommandIgnoreWarnings(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create .info file with non-existent paths
	infoContent := `nonexistent.txt This file doesn't exist
missing-dir/ This directory doesn't exist
real.txt This is a real file`
	if err := os.WriteFile(".info", []byte(infoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the real file
	if err := os.WriteFile("real.txt", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with --ignore-warnings flag
	cmd := newTestShowCommand()
	cmd.SetArgs([]string{"--ignore-warnings"})
	output, err := executeCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that warnings are NOT present
	if strings.Contains(output, "⚠️  Warnings found in .info files:") {
		t.Error("warnings should not be shown with --ignore-warnings")
	}
	if strings.Contains(output, "Path not found") {
		t.Error("path warnings should not be shown with --ignore-warnings")
	}

	// But the tree should still be shown
	if !strings.Contains(output, "real.txt") || !strings.Contains(output, "This is a real file") {
		t.Error("expected tree output with annotations")
	}
}

func TestShowCommandNoWarnings(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(tempDir)

	// Create .info file with only existing paths
	infoContent := `real1.txt First file
sub Second file`
	if err := os.WriteFile(".info", []byte(infoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the real files
	if err := os.WriteFile("real1.txt", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("sub", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test that no warnings are shown when all paths exist
	cmd := newTestShowCommand()
	output, err := executeCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that warnings are NOT present
	if strings.Contains(output, "⚠️  Warnings found in .info files:") {
		t.Error("should not show warnings when all paths exist")
	}
}

// Helper function to create a new show command for testing
func newTestShowCommand() *cobra.Command {
	// Create a new root command
	root := &cobra.Command{
		Use:   "treex",
		Short: "Test",
	}

	// Create show command
	show := &cobra.Command{
		Use:   "show [path...]",
		Short: "Display annotated file tree",
		RunE:  runShowCmd,
	}

	// Add flags
	show.Flags().BoolVar(&ignoreWarnings, "ignore-warnings", false, "Don't print warnings")
	show.Flags().StringVar(&infoFile, "info-file", ".info", "Info file name")
	show.Flags().StringVar(&ignoreFile, "ignore-file", ".gitignore", "Ignore file")
	show.Flags().BoolVar(&noIgnore, "no-ignore", false, "Don't use ignore file")
	show.Flags().IntVar(&maxDepth, "depth", 10, "Max depth")
	show.Flags().StringVar(&outputFormat, "format", "no-color", "Output format")
	show.Flags().StringVar(&showMode, "show", "mix", "Show mode")

	root.AddCommand(show)
	return root
}

// Helper function to execute a command and capture output
func executeCommand(cmd *cobra.Command) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	err := cmd.Execute()
	return buf.String(), err
}