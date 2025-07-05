package info

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateInfoFiles_ValidFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files and directories
	testFiles := []string{
		"README.md",
		"main.go",
		"src/app.go",
		"docs/guide.md",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Create the file
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create a valid .info file that references existing files
	infoContent := `README.md: Main project documentation

main.go: Application entry point

src/: Source code directory

docs/guide.md: User guide documentation
`

	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Test validation
	err := ValidateInfoFiles(tempDir)
	if err != nil {
		t.Fatalf("ValidateInfoFiles failed for valid files: %v", err)
	}
}

func TestValidateInfoFiles_NonExistentPaths(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files (but not all referenced in .info)
	testFiles := []string{
		"README.md",
		"main.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .info file that references non-existent files
	infoContent := `README.md: Main project documentation

main.go: Application entry point

nonexistent.txt: This file does not exist

missing-dir/: This directory does not exist
`

	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Test validation - should fail
	err := ValidateInfoFiles(tempDir)
	if err == nil {
		t.Error("Expected ValidateInfoFiles to fail for non-existent paths")
	}

	// Check error message contains information about missing files
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "nonexistent.txt") {
		t.Errorf("Expected error message to mention nonexistent.txt, got: %v", err)
	}

	if !strings.Contains(errorMsg, "missing-dir") {
		t.Errorf("Expected error message to mention missing-dir, got: %v", err)
	}

	if !strings.Contains(errorMsg, "validation errors") {
		t.Errorf("Expected error message to indicate validation errors, got: %v", err)
	}
}

func TestValidateInfoFiles_NestedInfoFiles(t *testing.T) {
	// Create a temporary directory structure with nested .info files
	tempDir := t.TempDir()

	// Create files in nested structure
	testFiles := []string{
		"README.md",
		"src/main.go",
		"src/utils.go",
		"src/internal/parser.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create root .info file
	rootInfoContent := `README.md: Main documentation

src/: Source code directory
`

	rootInfoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(rootInfoPath, []byte(rootInfoContent), 0644); err != nil {
		t.Fatalf("Failed to create root .info file: %v", err)
	}

	// Create src/.info file with valid references
	srcInfoContent := `main.go: Main application file

utils.go: Utility functions

internal/: Internal packages directory
`

	srcInfoPath := filepath.Join(tempDir, "src", ".info")
	if err := os.WriteFile(srcInfoPath, []byte(srcInfoContent), 0644); err != nil {
		t.Fatalf("Failed to create src .info file: %v", err)
	}

	// Create src/internal/.info file with invalid reference
	internalInfoContent := `parser.go: Parser implementation

missing.go: This file does not exist
`

	internalInfoPath := filepath.Join(tempDir, "src", "internal", ".info")
	if err := os.WriteFile(internalInfoPath, []byte(internalInfoContent), 0644); err != nil {
		t.Fatalf("Failed to create internal .info file: %v", err)
	}

	// Test validation - should fail due to missing.go
	err := ValidateInfoFiles(tempDir)
	if err == nil {
		t.Error("Expected ValidateInfoFiles to fail for missing.go in nested .info file")
	}

	// Check error mentions the missing file with correct path
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "src/internal/missing.go") {
		t.Errorf("Expected error message to mention src/internal/missing.go, got: %v", err)
	}
}

func TestValidateInfoFiles_NoInfoFiles(t *testing.T) {
	// Create a temporary directory with no .info files
	tempDir := t.TempDir()

	// Create some regular files
	testFiles := []string{
		"README.md",
		"main.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Test validation - should succeed (no .info files to validate)
	err := ValidateInfoFiles(tempDir)
	if err != nil {
		t.Fatalf("ValidateInfoFiles failed for directory with no .info files: %v", err)
	}
}

func TestValidateInfoFiles_InvalidPath(t *testing.T) {
	// Test with non-existent path
	err := ValidateInfoFiles("/nonexistent/path/12345")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	if !strings.Contains(err.Error(), "path does not exist") {
		t.Errorf("Expected error message about non-existent path, got: %v", err)
	}
}

func TestValidateInfoFiles_FileNotDirectory(t *testing.T) {
	// Create a temporary file (not directory)
	tempFile := filepath.Join(t.TempDir(), "notadir.txt")
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with file instead of directory
	err = ValidateInfoFiles(tempFile)
	if err == nil {
		t.Error("Expected error when path is not a directory")
	}

	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Expected error message about not a directory, got: %v", err)
	}
}

func TestValidateInfoFiles_MalformedInfoFile(t *testing.T) {
	// Create a temporary directory with malformed .info file
	tempDir := t.TempDir()

	// Create a file to reference
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a malformed .info file (missing descriptions)
	infoContent := `test.txt
`

	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Test validation - should not panic but might fail validation
	_ = ValidateInfoFiles(tempDir)
	// We don't check for specific errors here because we just want to make sure it doesn't crash
}

func TestValidateInfoFiles_EmptyInfoFile(t *testing.T) {
	// Create a temporary directory with empty .info file
	tempDir := t.TempDir()

	// Create an empty .info file
	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty .info file: %v", err)
	}

	// Test validation - should succeed (empty info file is valid)
	err := ValidateInfoFiles(tempDir)
	if err != nil {
		t.Fatalf("ValidateInfoFiles failed for empty .info file: %v", err)
	}
}
