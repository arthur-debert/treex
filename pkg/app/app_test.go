package app

import (
	"os"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/format" // Import the format package
)

func TestRenderAnnotatedTree_BasicFunctionality(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create a simple .info file
	infoContent := `cmd/
Main binary command with subcommands for different operations.

src/
Source code with core business logic.
`
	
	infoPath := tempDir + "/.info"
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create test .info file: %v", err)
	}
	
	// Create some test directories
	if err := os.MkdirAll(tempDir+"/cmd", 0755); err != nil {
		t.Fatalf("Failed to create cmd directory: %v", err)
	}
	if err := os.MkdirAll(tempDir+"/src", 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}
	
	// Test basic rendering
	options := RenderOptions{
		Verbose:    false,
		NoColor:    true, // Use plain text for predictable testing
		Minimal:    false,
		IgnoreFile: "",
		MaxDepth:   -1,
		SafeMode:   true,
	}
	
	result, err := RenderAnnotatedTree(tempDir, options)
	if err != nil {
		t.Fatalf("RenderAnnotatedTree failed: %v", err)
	}
	
	// Verify the result
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if result.Stats == nil {
		t.Fatal("Expected non-nil stats")
	}
	
	if result.Stats.AnnotationsFound != 2 {
		t.Errorf("Expected 2 annotations, got %d", result.Stats.AnnotationsFound)
	}
	
	if !result.Stats.TreeGenerated {
		t.Error("Expected tree to be generated")
	}
	
	// Check that output contains expected content
	output := result.Output
	if !strings.Contains(output, "cmd") {
		t.Error("Expected output to contain 'cmd'")
	}
	
	if !strings.Contains(output, "src") {
		t.Error("Expected output to contain 'src'")
	}
	
	// Check that annotations are included
	if !strings.Contains(output, "Main binary command") {
		t.Error("Expected output to contain cmd annotation")
	}
	
	if !strings.Contains(output, "Source code with core") {
		t.Error("Expected output to contain src annotation")
	}
}

func TestRenderAnnotatedTree_VerboseMode(t *testing.T) {
	// Use a simple directory without .info files
	tempDir := t.TempDir()
	
	options := RenderOptions{
		Verbose:    true,
		NoColor:    true,
		Minimal:    false,
		IgnoreFile: "",
		MaxDepth:   -1,
		SafeMode:   true,
	}
	
	result, err := RenderAnnotatedTree(tempDir, options)
	if err != nil {
		t.Fatalf("RenderAnnotatedTree failed: %v", err)
	}
	
	// In verbose mode, the VerboseOutput field should be populated
	if result.VerboseOutput == nil {
		t.Fatal("Expected VerboseOutput to be populated in verbose mode")
	}

	if result.VerboseOutput.AnalyzedPath != tempDir {
		t.Errorf("Expected AnalyzedPath to be %s, got %s", tempDir, result.VerboseOutput.AnalyzedPath)
	}

	if result.VerboseOutput.FoundAnnotations != 0 {
		t.Errorf("Expected FoundAnnotations to be 0, got %d", result.VerboseOutput.FoundAnnotations)
	}

	if result.VerboseOutput.ParsedAnnotations == nil {
		t.Error("Expected ParsedAnnotations to be non-nil (even if empty)")
	}

	// The main result.Output should NOT contain verbose strings anymore
	output := result.Output
	if strings.Contains(output, "Analyzing directory:") {
		t.Error("Expected result.Output NOT to contain 'Analyzing directory:'")
	}
	if strings.Contains(output, "=== Parsed Annotations ===") {
		t.Error("Expected result.Output NOT to contain annotations section")
	}
}

func TestRenderAnnotatedTree_VerboseModeWithAnnotations(t *testing.T) {
	tempDir := t.TempDir()
	infoContent := "file.txt\nThis is a file."
	infoPath := tempDir + "/.info"
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create test .info file: %v", err)
	}
	if _, err := os.Create(tempDir + "/file.txt"); err != nil {
		t.Fatalf("Failed to create test file.txt: %v", err)
	}

	options := RenderOptions{
		Verbose:    true,
		NoColor:    true, // Kept for legacy path in RenderAnnotatedTree, though Format takes precedence
		Format:     string(format.FormatNoColor), // Use format.FormatNoColor
		SafeMode:   true,
	}

	result, err := RenderAnnotatedTree(tempDir, options)
	if err != nil {
		t.Fatalf("RenderAnnotatedTree failed: %v", err)
	}

	if result.VerboseOutput == nil {
		t.Fatal("Expected VerboseOutput to be populated")
	}
	if result.VerboseOutput.FoundAnnotations != 1 {
		t.Errorf("Expected FoundAnnotations to be 1, got %d", result.VerboseOutput.FoundAnnotations)
	}
	if _, ok := result.VerboseOutput.ParsedAnnotations["file.txt"]; !ok {
		t.Error("Expected ParsedAnnotations to contain 'file.txt'")
	}
	if !strings.Contains(result.VerboseOutput.TreeStructure, "file.txt") {
		t.Error("Expected VerboseOutput.TreeStructure to contain 'file.txt'")
	}
}

func TestRenderAnnotatedTree_InvalidPath(t *testing.T) {
	options := RenderOptions{
		Verbose:    false,
		NoColor:    true,
		Minimal:    false,
		IgnoreFile: "",
		MaxDepth:   -1,
		SafeMode:   true,
	}
	
	// Test with non-existent path
	_, err := RenderAnnotatedTree("/nonexistent/path/12345", options)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
} 