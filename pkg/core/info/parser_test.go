package info

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temporary .info file for testing
	tempDir := t.TempDir()
	infoFile := filepath.Join(tempDir, ".info")

	content := `README.md Like the title says, that useful little readme.
LICENSE MIT, like most things.
.github/workflows/go.yml CI Unit test workflow - makes usage of go action, that does pretty much all go setup. Note that his has no caching just yet.`

	err := os.WriteFile(infoFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .info file: %v", err)
	}

	parser := NewParser()
	annotations, err := parser.ParseFile(infoFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check that we parsed the expected number of annotations
	expectedCount := 3
	if len(annotations) != expectedCount {
		t.Errorf("Expected %d annotations, got %d", expectedCount, len(annotations))
	}

	// Test README.md annotation
	readme, exists := annotations["README.md"]
	if !exists {
		t.Error("README.md annotation not found")
	} else {
		expectedTitle := "Like the title says, that useful little readme."
		expectedDesc := "Like the title says, that useful little readme."
		if readme.Title != expectedTitle {
			t.Errorf("README.md title mismatch.\nExpected: %q\nGot: %q", expectedTitle, readme.Title)
		}
		if readme.Description != expectedDesc {
			t.Errorf("README.md description mismatch.\nExpected: %q\nGot: %q", expectedDesc, readme.Description)
		}
	}

	// Test LICENSE annotation
	license, exists := annotations["LICENSE"]
	if !exists {
		t.Error("LICENSE annotation not found")
	} else {
		expectedTitle := "MIT, like most things."
		expectedDesc := "MIT, like most things."
		if license.Title != expectedTitle {
			t.Errorf("LICENSE title mismatch.\nExpected: %q\nGot: %q", expectedTitle, license.Title)
		}
		if license.Description != expectedDesc {
			t.Errorf("LICENSE description mismatch.\nExpected: %q\nGot: %q", expectedDesc, license.Description)
		}
	}

	// Test .github/workflows/go.yml annotation (single-line)
	workflow, exists := annotations[".github/workflows/go.yml"]
	if !exists {
		t.Error(".github/workflows/go.yml annotation not found")
	} else {
		expectedNotes := "CI Unit test workflow - makes usage of go action, that does pretty much all go setup. Note that his has no caching just yet."
		if workflow.Notes != expectedNotes {
			t.Errorf("Workflow notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, workflow.Notes)
		}
	}
}

func TestParseFileNotExists(t *testing.T) {
	parser := NewParser()
	annotations, err := parser.ParseFile("nonexistent.info")

	if err != nil {
		t.Errorf("ParseFile should not error on non-existent file, got: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("Expected empty annotations map for non-existent file, got %d annotations", len(annotations))
	}
}

func TestParseDirectory(t *testing.T) {
	// Create a temporary directory with .info file
	tempDir := t.TempDir()
	infoFile := filepath.Join(tempDir, ".info")

	content := `test.txt A test file.`

	err := os.WriteFile(infoFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .info file: %v", err)
	}

	annotations, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(annotations))
	}

	testFile, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else {
		expectedTitle := "A test file."
		expectedDesc := "A test file."
		if testFile.Title != expectedTitle {
			t.Errorf("Title mismatch.\nExpected: %q\nGot: %q", expectedTitle, testFile.Title)
		}
		if testFile.Description != expectedDesc {
			t.Errorf("Description mismatch.\nExpected: %q\nGot: %q", expectedDesc, testFile.Description)
		}
	}
}

func TestParseDirectoryTree(t *testing.T) {
	// Create a temporary directory structure with multiple .info files
	tempDir := t.TempDir()

	// Create root .info file
	rootInfo := `README.md Root level readme file

main.go Main application entry point`

	err := os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create root .info file: %v", err)
	}

	// Create subdirectory with its own .info file
	subDir := filepath.Join(tempDir, "internal")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	subInfo := `parser.go Handles parsing of .info files

builder.go Constructs file trees from filesystem`

	err = os.WriteFile(filepath.Join(subDir, ".info"), []byte(subInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub .info file: %v", err)
	}

	// Create nested subdirectory with .info file
	nestedDir := filepath.Join(subDir, "deep")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	nestedInfo := `config.json Deep configuration file`

	err = os.WriteFile(filepath.Join(nestedDir, ".info"), []byte(nestedInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create nested .info file: %v", err)
	}

	// Parse the entire directory tree
	annotations, err := ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("ParseDirectoryTree failed: %v", err)
	}

	// Verify we got annotations from all .info files
	expectedCount := 5 // 2 from root + 2 from internal + 1 from deep
	if len(annotations) != expectedCount {
		t.Errorf("Expected %d annotations, got %d", expectedCount, len(annotations))
		for path := range annotations {
			t.Logf("Found annotation for: %s", path)
		}
	}

	// Check root level annotations
	if readme, exists := annotations["README.md"]; !exists {
		t.Error("Root README.md annotation not found")
	} else if readme.Title != "Root level readme file" {
		t.Errorf("Root README.md title incorrect: %q", readme.Title)
	}

	// Check internal level annotations (should have "internal/" prefix)
	if parser, exists := annotations["internal/parser.go"]; !exists {
		t.Error("internal/parser.go annotation not found")
	} else if parser.Title != "Handles parsing of .info files" {
		t.Errorf("internal/parser.go title incorrect: %q", parser.Title)
	}

	// Check nested level annotations (should have "internal/deep/" prefix)
	if config, exists := annotations["internal/deep/config.json"]; !exists {
		t.Error("internal/deep/config.json annotation not found")
	} else if config.Title != "Deep configuration file" {
		t.Errorf("internal/deep/config.json title incorrect: %q", config.Title)
	}
}

func TestParseFileWithContext(t *testing.T) {
	// Create a temporary .info file
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	infoContent := `file.txt A test file

nested/deep.txt A deeply nested file`

	infoPath := filepath.Join(subDir, ".info")
	err = os.WriteFile(infoPath, []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Parse with context
	annotations, err := parseFileWithContext(infoPath, tempDir, subDir)
	if err != nil {
		t.Fatalf("parseFileWithContext failed: %v", err)
	}

	// Check that paths are resolved correctly
	if len(annotations) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(annotations))
	}

	// Check resolved paths
	if _, exists := annotations["sub/file.txt"]; !exists {
		t.Error("sub/file.txt annotation not found")
	} else {
		// Check that content is preserved
		fileAnnotation := annotations["sub/file.txt"]
		if fileAnnotation.Title != "A test file" {
			t.Errorf("file.txt title incorrect: %q", fileAnnotation.Title)
		}
	}

	// Check nested path resolution
	if _, exists := annotations["sub/nested/deep.txt"]; !exists {
		t.Error("sub/nested/deep.txt annotation not found")
	} else {
		// Check that content is preserved
		deepAnnotation := annotations["sub/nested/deep.txt"]
		if deepAnnotation.Title != "A deeply nested file" {
			t.Errorf("deep.txt title incorrect: %q", deepAnnotation.Title)
		}
	}
}

func TestParseFileWithContextSecurityCheck(t *testing.T) {
	// Test that .. paths are rejected
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create .info file with dangerous paths
	infoContent := `../../../etc/passwd Dangerous path attempt

file.txt Safe file

../parent.txt Another dangerous path`

	infoPath := filepath.Join(subDir, ".info")
	err = os.WriteFile(infoPath, []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Parse with context
	annotations, err := parseFileWithContext(infoPath, tempDir, subDir)
	if err != nil {
		t.Fatalf("parseFileWithContext failed: %v", err)
	}

	// Should only have the safe file, dangerous paths should be filtered out
	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation (safe file only), got %d", len(annotations))
		for path := range annotations {
			t.Logf("Found annotation for: %s", path)
		}
	}

	// Check that only safe file remains
	if _, exists := annotations["sub/file.txt"]; !exists {
		t.Error("sub/file.txt annotation not found")
	}
}

