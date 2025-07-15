package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestDrawCommand tests the basic functionality of the draw command
func TestDrawCommand(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a test info file
	familyContent := `Dad Chill, dad
Mom Listen to your mother
kids/Sam Little Sam
kids/Jane Big sister Jane`

	familyFile := filepath.Join(tempDir, "family.txt")
	err := os.WriteFile(familyFile, []byte(familyContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test draw command with file
	output := &bytes.Buffer{}
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}
	testDrawCmd := &cobra.Command{
		Use:  "draw",
		Args: cobra.NoArgs,
		RunE: runDrawCmd,
	}
	testDrawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file")
	testDrawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	testRootCmd.AddCommand(testDrawCmd)

	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"draw", "--info-file", familyFile, "--format=no-color"})

	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Draw command failed: %v", err)
	}

	outputStr := output.String()

	// Check that all entries are displayed
	if !strings.Contains(outputStr, "Dad") {
		t.Errorf("Expected 'Dad' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Chill, dad") {
		t.Errorf("Expected Dad's annotation in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Mom") {
		t.Errorf("Expected 'Mom' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Listen to your mother") {
		t.Errorf("Expected Mom's annotation in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "kids") {
		t.Errorf("Expected 'kids' directory in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Sam") {
		t.Errorf("Expected 'Sam' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Little Sam") {
		t.Errorf("Expected Sam's annotation in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Jane") {
		t.Errorf("Expected 'Jane' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Big sister Jane") {
		t.Errorf("Expected Jane's annotation in output, got: %s", outputStr)
	}
}

// TestDrawCommandWithDirectories tests draw command with directory paths
func TestDrawCommandWithDirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test info file with directories
	projectContent := `src/ Source code directory
src/main.go Entry point
src/util/ Utility functions
src/util/helper.go Helper functions
docs/ Documentation
docs/README.md Main documentation`

	projectFile := filepath.Join(tempDir, "project.txt")
	err := os.WriteFile(projectFile, []byte(projectContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test draw command
	output := &bytes.Buffer{}
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}
	testDrawCmd := &cobra.Command{
		Use:  "draw",
		Args: cobra.NoArgs,
		RunE: runDrawCmd,
	}
	testDrawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file")
	testDrawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	testRootCmd.AddCommand(testDrawCmd)

	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"draw", "--info-file", projectFile, "--format=no-color"})

	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Draw command failed: %v", err)
	}

	outputStr := output.String()

	// Check directory structure
	if !strings.Contains(outputStr, "src/") || !strings.Contains(outputStr, "src") {
		t.Errorf("Expected 'src/' directory in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Source code directory") {
		t.Errorf("Expected src annotation in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "main.go") {
		t.Errorf("Expected 'main.go' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "util/") || !strings.Contains(outputStr, "util") {
		t.Errorf("Expected 'util/' subdirectory in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "helper.go") {
		t.Errorf("Expected 'helper.go' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "docs/") || !strings.Contains(outputStr, "docs") {
		t.Errorf("Expected 'docs/' directory in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "README.md") {
		t.Errorf("Expected 'README.md' in output, got: %s", outputStr)
	}
}

// TestDrawCommandFromStdin tests reading from stdin
func TestDrawCommandFromStdin(t *testing.T) {
	// Create test content
	treeContent := `root The root node
branch1 First branch
branch2 Second branch`

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	// Write content to the pipe
	go func() {
		defer w.Close()
		_, _ = w.Write([]byte(treeContent))
	}()

	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Replace stdin with our pipe
	os.Stdin = r

	// Test draw command reading from stdin
	output := &bytes.Buffer{}
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}
	testDrawCmd := &cobra.Command{
		Use:  "draw",
		Args: cobra.NoArgs,
		RunE: runDrawCmd,
	}
	testDrawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file")
	testDrawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	testRootCmd.AddCommand(testDrawCmd)

	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"draw", "--format=no-color"})

	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Draw command failed: %v", err)
	}

	outputStr := output.String()

	// Check output
	if !strings.Contains(outputStr, "root") {
		t.Errorf("Expected 'root' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "The root node") {
		t.Errorf("Expected root annotation in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "branch1") {
		t.Errorf("Expected 'branch1' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "branch2") {
		t.Errorf("Expected 'branch2' in output, got: %s", outputStr)
	}
}

// TestDrawCommandMarkdownFormat tests markdown output format
func TestDrawCommandMarkdownFormat(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple test file
	content := `docs/ Documentation
docs/api.md API docs
src/ Source code`

	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test with markdown format
	output := &bytes.Buffer{}
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}
	testDrawCmd := &cobra.Command{
		Use:  "draw",
		Args: cobra.NoArgs,
		RunE: runDrawCmd,
	}
	testDrawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file")
	testDrawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	testRootCmd.AddCommand(testDrawCmd)

	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"draw", "--info-file", testFile, "--format=markdown"})

	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Draw command failed: %v", err)
	}

	outputStr := output.String()

	// Check for markdown formatting
	if !strings.Contains(outputStr, "```") {
		t.Errorf("Expected markdown code block in output, got: %s", outputStr)
	}

	// Content should still be there
	if !strings.Contains(outputStr, "docs/") {
		t.Errorf("Expected 'docs/' in markdown output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Documentation") {
		t.Errorf("Expected annotation in markdown output, got: %s", outputStr)
	}
}

// TestDrawCommandWithWarnings tests warning handling in draw command
func TestDrawCommandWithWarnings(t *testing.T) {
	tempDir := t.TempDir()

	// Create content with invalid lines
	content := `valid/path Valid annotation
invalid line without annotation
: Empty path
another/path Another valid one`

	testFile := filepath.Join(tempDir, "warnings.txt")
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test without --ignore-warnings
	output := &bytes.Buffer{}
	resetGlobalFlags()

	testRootCmd := &cobra.Command{Use: "treex"}
	testDrawCmd := &cobra.Command{
		Use:  "draw",
		Args: cobra.NoArgs,
		RunE: runDrawCmd,
	}
	testDrawCmd.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file")
	testDrawCmd.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	testDrawCmd.Flags().BoolVar(&ignoreWarnings, "ignore-warnings", false, "Don't print warnings")
	testRootCmd.AddCommand(testDrawCmd)

	testRootCmd.SetOut(output)
	testRootCmd.SetErr(output)
	testRootCmd.SetArgs([]string{"draw", "--info-file", testFile, "--format=no-color"})

	err = testRootCmd.Execute()
	if err != nil {
		t.Fatalf("Draw command failed: %v", err)
	}

	outputStr := output.String()

	// Should show warnings
	if !strings.Contains(outputStr, "⚠️  Warnings found in tree data:") {
		t.Errorf("Expected warnings header, got: %s", outputStr)
	}

	// Test with --ignore-warnings
	output.Reset()
	resetGlobalFlags()

	testRootCmd2 := &cobra.Command{Use: "treex"}
	testDrawCmd2 := &cobra.Command{
		Use:  "draw",
		Args: cobra.NoArgs,
		RunE: runDrawCmd,
	}
	testDrawCmd2.Flags().StringVar(&infoFile, "info-file", "", "Read tree data from specified file")
	testDrawCmd2.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	testDrawCmd2.Flags().BoolVar(&ignoreWarnings, "ignore-warnings", false, "Don't print warnings")
	testRootCmd2.AddCommand(testDrawCmd2)

	testRootCmd2.SetOut(output)
	testRootCmd2.SetErr(output)
	testRootCmd2.SetArgs([]string{"draw", "--info-file", testFile, "--format=no-color", "--ignore-warnings"})

	err = testRootCmd2.Execute()
	if err != nil {
		t.Fatalf("Draw command failed: %v", err)
	}

	outputStr = output.String()

	// Should not show warnings
	if strings.Contains(outputStr, "⚠️  Warnings") {
		t.Errorf("Warnings should be suppressed with --ignore-warnings, got: %s", outputStr)
	}

	// But valid entries should still be shown
	if !strings.Contains(outputStr, "valid/path") || !strings.Contains(outputStr, "valid") {
		t.Errorf("Expected valid entries in output, got: %s", outputStr)
	}
}