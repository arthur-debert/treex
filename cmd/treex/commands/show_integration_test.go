package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestShowCmd_Integration_BasicTree(t *testing.T) {
	// Create a test directory structure
	tempDir := t.TempDir()

	// Create directory structure
	dirs := []string{
		"src",
		"src/utils",
		"docs",
		"tests",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tempDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create files
	files := []string{
		"README.md",
		"main.go",
		"src/app.go",
		"src/utils/helper.go",
		"docs/guide.md",
		"tests/app_test.go",
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Create .info file with annotations
	infoContent := `README.md: Project documentation
The main readme file for the project

main.go: Application entry point
The main function that starts the application

src/app.go: Core application logic

docs/guide.md: User guide
Comprehensive guide for users`

	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Run the show command
	resetGlobalFlags()
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	output, err := executeCommand(testRootCmd, "show", tempDir, "--format=no-color")
	if err != nil {
		t.Fatalf("Show command failed: %v", err)
	}

	// Verify the output contains expected elements
	expectedElements := []string{
		"README.md",
		"Project documentation",
		"main.go",
		"Application entry point",
		"src",
		"app.go",
		"Core application logic",
		"docs",
		"guide.md",
		"User guide",
		"tests", // Directory shows but contents might not in mix mode
	}

	// The directory name in output will be the temp dir name, not our choice
	// So we check for file names and structure instead

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected output to contain '%s'.\nFull output:\n%s", elem, output)
		}
	}

	// Verify tree structure connectors
	if !strings.Contains(output, "├──") || !strings.Contains(output, "└──") {
		t.Errorf("Expected output to contain tree connectors (├──, └──).\nFull output:\n%s", output)
	}
}

func TestShowCmd_Integration_ViewModes(t *testing.T) {
	// Create a test directory structure with many files
	tempDir := t.TempDir()

	// Create many files to test view modes
	annotatedFiles := map[string]string{
		"important1.go": "Critical file 1",
		"important2.go": "Critical file 2",
		"src/core.go":   "Core functionality",
	}

	// Create annotated files
	for file := range annotatedFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Create many unannotated files
	for i := 1; i <= 10; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Create .info file
	var infoLines []string
	for file, desc := range annotatedFiles {
		infoLines = append(infoLines, fmt.Sprintf("%s: %s", file, desc))
	}
	infoContent := strings.Join(infoLines, "\n\n")

	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	testCases := []struct {
		name             string
		viewMode         string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:     "annotated mode",
			viewMode: "annotated",
			shouldContain: []string{
				"important1.go",
				"important2.go",
				"core.go",
				"Critical file 1",
				"treex --show all to see all paths",
			},
			shouldNotContain: []string{
				"file1.txt",
				"file5.txt",
			},
		},
		{
			name:     "all mode",
			viewMode: "all",
			shouldContain: []string{
				"important1.go",
				"file1.txt",
				"file10.txt",
				"Critical file 1",
			},
			shouldNotContain: []string{
				"more items",
				"treex --show all",
			},
		},
		{
			name:     "mix mode",
			viewMode: "mix",
			shouldContain: []string{
				"important1.go",
				"Critical file 1",
				"file1.txt",  // Some context files
				"more items", // Should have "more items" indicator
			},
			shouldNotContain: []string{
				"file7.txt", // Shouldn't show files beyond context limit (6 unannotated + 2 annotated)
				"file8.txt", // Shouldn't show files beyond context limit
				"treex --show all",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resetGlobalFlags()
			testRootCmd := &cobra.Command{Use: "treex"}
			testShowCmd := setupShowCmd()
			testRootCmd.AddCommand(testShowCmd)

			output, err := executeCommand(testRootCmd, "show", tempDir, "--format=no-color", "--show="+tc.viewMode)
			if err != nil {
				t.Fatalf("Show command failed: %v", err)
			}

			// Check elements that should be present
			for _, elem := range tc.shouldContain {
				if !strings.Contains(output, elem) {
					t.Errorf("Expected output to contain '%s' in %s mode.\nFull output:\n%s",
						elem, tc.viewMode, output)
				}
			}

			// Check elements that should NOT be present
			for _, elem := range tc.shouldNotContain {
				if strings.Contains(output, elem) {
					t.Errorf("Expected output NOT to contain '%s' in %s mode.\nFull output:\n%s",
						elem, tc.viewMode, output)
				}
			}
		})
	}
}

func TestShowCmd_Integration_SingleLineAnnotations(t *testing.T) {
	// Test that single-line annotations are displayed correctly
	tempDir := t.TempDir()

	// Create a file
	if err := os.WriteFile(filepath.Join(tempDir, "complex.go"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create .info file with new format (single-line only)
	infoContent := `complex.go: Complex component that handles authentication and session management`

	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	resetGlobalFlags()
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	output, err := executeCommand(testRootCmd, "show", tempDir, "--format=no-color")
	if err != nil {
		t.Fatalf("Show command failed: %v", err)
	}

	// Check that the annotation appears
	if !strings.Contains(output, "Complex component that handles authentication and session management") {
		t.Errorf("Expected output to contain annotation.\nFull output:\n%s", output)
	}
}

func TestShowCmd_Integration_DeepNesting(t *testing.T) {
	// Test deeply nested directory structures
	tempDir := t.TempDir()

	// Create deep directory structure
	deepPath := "src/internal/core/handlers/api/v2"
	if err := os.MkdirAll(filepath.Join(tempDir, deepPath), 0755); err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Create file at various levels
	files := []string{
		"README.md",
		"src/main.go",
		"src/internal/utils.go",
		"src/internal/core/engine.go",
		filepath.Join(deepPath, "handler.go"),
	}

	for _, file := range files {
		if err := os.WriteFile(filepath.Join(tempDir, file), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Create .info with annotations at different levels
	infoContent := `README.md: Project root documentation

src/main.go: Main entry point

src/internal/core/engine.go: Core processing engine
The heart of the application`

	if err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Also create a nested .info file
	nestedInfoContent := `handler.go: API v2 handler
Handles all v2 API requests`

	if err := os.WriteFile(filepath.Join(tempDir, deepPath, ".info"), []byte(nestedInfoContent), 0644); err != nil {
		t.Fatalf("Failed to create nested .info file: %v", err)
	}

	resetGlobalFlags()
	testRootCmd := &cobra.Command{Use: "treex"}
	testShowCmd := setupShowCmd()
	testRootCmd.AddCommand(testShowCmd)

	output, err := executeCommand(testRootCmd, "show", tempDir, "--format=no-color")
	if err != nil {
		t.Fatalf("Show command failed: %v", err)
	}

	// Verify deep nesting is shown correctly
	expectedElements := []string{
		"src",
		"internal",
		"core",
		"handlers",
		"api",
		"v2",
		"handler.go",
		"API v2 handler",
		"engine.go",
		"Core processing engine",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected output to contain '%s'.\nFull output:\n%s", elem, output)
		}
	}

	// Verify proper indentation (more │ characters = deeper nesting)
	lines := strings.Split(output, "\n")
	var v2Line string
	for _, line := range lines {
		if strings.Contains(line, "handler.go") && strings.Contains(line, "API v2 handler") {
			v2Line = line
			break
		}
	}

	// Count tree characters to verify deep nesting
	treeChars := strings.Count(v2Line, "│")
	if treeChars < 4 { // Should have at least 4 levels of nesting
		t.Errorf("Expected deep nesting (at least 4 │ chars), found %d in line: %s", treeChars, v2Line)
	}
}
