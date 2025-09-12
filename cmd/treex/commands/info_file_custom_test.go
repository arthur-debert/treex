package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestCustomInfoFileName tests that --info-file flag correctly changes the info file name
func TestCustomInfoFileName(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create directory structure
	dirs := []string{"docs", "src"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create .info files (should be ignored when using custom name)
	defaultInfoContent := `docs Default info for docs
src Default info for src`
	if err := os.WriteFile(".info", []byte(defaultInfoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create custom info files
	customInfoContent := `docs Custom documentation
src Custom source code`
	if err := os.WriteFile("other.txt", []byte(customInfoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Also create nested custom info files
	nestedCustomInfo := `README.md Important readme file`
	if err := os.WriteFile("docs/other.txt", []byte(nestedCustomInfo), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/README.md", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Default behavior should use .info files
	output := &bytes.Buffer{}
	
	// Reset global flags
	resetGlobalFlags()
	
	// Create test command
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)
	
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"show", ".", "--format=no-color"})
	
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error with default info files: %v", err)
	}
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "Default info for docs") {
		t.Errorf("expected to see default .info annotation, got: %s", outputStr)
	}

	// Test 2: Using --info-file should use custom files and NOT .info files
	output.Reset()
	resetGlobalFlags()
	
	testRootCmd2 := &cobra.Command{Use: "treex"}
	testShowCmd2 := setupShowCmd()
	testRootCmd2.AddCommand(testShowCmd2)
	
	testRootCmd2.SetOut(output)
	testRootCmd2.SetErr(output)
	testRootCmd2.SetArgs([]string{"show", ".", "--format=no-color", "--info-file", "other.txt"})
	
	err = testRootCmd2.Execute()
	if err != nil {
		t.Fatalf("unexpected error with custom info files: %v", err)
	}
	
	outputStr = output.String()
	if strings.Contains(outputStr, "Default info") {
		t.Errorf("should not see default .info annotations when using --info-file, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Custom documentation") {
		t.Errorf("expected to see custom info annotation for docs, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Custom source code") {
		t.Errorf("expected to see custom info annotation for src, got: %s", outputStr)
	}

	// Should also see nested custom info
	if !strings.Contains(outputStr, "Important readme file") {
		t.Errorf("expected to see nested custom info annotation, got: %s", outputStr)
	}
}

// TestCustomInfoFileWithMultiplePaths tests custom info files work with multiple paths
func TestCustomInfoFileWithMultiplePaths(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()
	
	// Create two separate directory trees
	if err := os.MkdirAll("project1/src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("project2/lib", 0755); err != nil {
		t.Fatal(err)
	}

	// Create custom info files for each project
	project1Info := `src Project 1 source code`
	if err := os.WriteFile("project1/my.info", []byte(project1Info), 0644); err != nil {
		t.Fatal(err)
	}

	project2Info := `lib Project 2 library`
	if err := os.WriteFile("project2/my.info", []byte(project2Info), 0644); err != nil {
		t.Fatal(err)
	}

	// Test showing multiple paths with custom info file
	output := &bytes.Buffer{}
	resetGlobalFlags()
	
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)
	
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"show", "project1", "project2", "--format=no-color", "--info-file", "my.info"})
	
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error with multiple paths: %v", err)
	}
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "Project 1 source code") {
		t.Errorf("expected to see project1 annotation, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Project 2 library") {
		t.Errorf("expected to see project2 annotation, got: %s", outputStr)
	}
}

// TestCustomInfoFileCommands tests that all commands respect custom info file names
func TestCustomInfoFileCommands(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create structure and custom info file
	if err := os.Mkdir("docs", 0755); err != nil {
		t.Fatal(err)
	}

	customInfo := `docs Documentation folder`
	if err := os.WriteFile("project.info", []byte(customInfo), 0644); err != nil {
		t.Fatal(err)
	}

	// Test info-search command
	output := &bytes.Buffer{}
	resetGlobalFlags()
	
	// Create info-search command
	testRootCmd := &cobra.Command{Use: "treex"}
	testInfoCmd := &cobra.Command{Use: "info", Short: "Info commands"}
	testSearchCmd := &cobra.Command{
		Use:  "search <term>",
		Args: cobra.ExactArgs(1),
		RunE: runSearch,
	}
	testSearchCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name")
	testInfoCmd.AddCommand(testSearchCmd)
	testRootCmd.AddCommand(testInfoCmd)
	
	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"info", "search", "Documentation", "--info-file", "project.info"})
	
	err := testRootCmd.Execute()
	if err != nil {
		t.Fatalf("info-search failed: %v", err)
	}
	
	outputStr := output.String()
	if !strings.Contains(outputStr, "Found 1 matches") {
		t.Errorf("info search should find match in custom info file, got: %s", outputStr)
	}

	// Test info-sync command (remove docs directory to make annotation stale)
	if err := os.RemoveAll("docs"); err != nil {
		t.Fatal(err)
	}

	output.Reset()
	resetGlobalFlags()
	
	// Create info-sync command
	testRootCmd2 := &cobra.Command{Use: "treex"}
	testInfoCmd2 := &cobra.Command{Use: "info", Short: "Info commands"}
	testSyncCmd := &cobra.Command{
		Use:  "sync",
		RunE: runSync,
	}
	testSyncCmd.Flags().BoolVar(&forceSync, "force", false, "Remove stale annotations without confirmation")
	testSyncCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name")
	testInfoCmd2.AddCommand(testSyncCmd)
	testRootCmd2.AddCommand(testInfoCmd2)
	
	testRootCmd2.SetOut(output)
	testRootCmd2.SetErr(output)
	testRootCmd2.SetArgs([]string{"info", "sync", "--info-file", "project.info", "--force"})
	
	err = testRootCmd2.Execute()
	if err != nil {
		t.Fatalf("info sync failed: %v", err)
	}

	// Check that the custom info file was updated
	content, _ := os.ReadFile("project.info")
	if strings.Contains(string(content), "docs Documentation") {
		t.Error("info sync should have removed stale annotation from custom info file")
	}
}