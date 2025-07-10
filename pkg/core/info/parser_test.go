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

	content := `README.md: Like the title says, that useful little readme.
LICENSE: MIT, like most things.
.github/workflows/go.yml: CI Unit test workflow - makes usage of go action, that does pretty much all go setup. Note that his has no caching just yet.`

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
		expectedNotes := "Like the title says, that useful little readme."
		if readme.Notes != expectedNotes {
			t.Errorf("README.md notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, readme.Notes)
		}
	}

	// Test LICENSE annotation
	license, exists := annotations["LICENSE"]
	if !exists {
		t.Error("LICENSE annotation not found")
	} else {
		expectedNotes := "MIT, like most things."
		if license.Notes != expectedNotes {
			t.Errorf("LICENSE notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, license.Notes)
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

	content := `test.txt: A test file.`

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
		expectedNotes := "A test file."
		if testFile.Notes != expectedNotes {
			t.Errorf("Notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, testFile.Notes)
		}
	}
}

func TestParseDirectoryTree(t *testing.T) {
	// Create a temporary directory structure with multiple .info files
	tempDir := t.TempDir()

	// Create root .info file
	rootInfo := `README.md: Root level readme file

main.go: Main application entry point`

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

	subInfo := `parser.go: Handles parsing of .info files

builder.go: Constructs file trees from filesystem`

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

	nestedInfo := `config.json: Deep configuration file`

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
	} else if readme.Notes != "Root level readme file" {
		t.Errorf("Root README.md notes incorrect: %q", readme.Notes)
	}

	// Check internal level annotations (should have "internal/" prefix)
	if parser, exists := annotations["internal/parser.go"]; !exists {
		t.Error("internal/parser.go annotation not found")
	} else if parser.Notes != "Handles parsing of .info files" {
		t.Errorf("internal/parser.go notes incorrect: %q", parser.Notes)
	}

	// Check nested level annotations (should have "internal/deep/" prefix)
	if config, exists := annotations["internal/deep/config.json"]; !exists {
		t.Error("internal/deep/config.json annotation not found")
	} else if config.Notes != "Deep configuration file" {
		t.Errorf("internal/deep/config.json notes incorrect: %q", config.Notes)
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

	infoContent := `file.txt: A test file

nested/deep.txt: A deeply nested file`

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
		if fileAnnotation.Notes != "A test file" {
			t.Errorf("file.txt notes incorrect: %q", fileAnnotation.Notes)
		}
	}

	// Check nested path resolution
	if _, exists := annotations["sub/nested/deep.txt"]; !exists {
		t.Error("sub/nested/deep.txt annotation not found")
	} else {
		// Check that content is preserved
		deepAnnotation := annotations["sub/nested/deep.txt"]
		if deepAnnotation.Notes != "A deeply nested file" {
			t.Errorf("deep.txt notes incorrect: %q", deepAnnotation.Notes)
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
	infoContent := `../../../etc/passwd: Dangerous path attempt

file.txt: Safe file

../parent.txt: Another dangerous path`

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

func TestNestedInfoFilePrecedence(t *testing.T) {
	// Test that deeper .info files take precedence over parent .info files
	// when the same file is annotated in both
	tempDir := t.TempDir()

	// Create files that will be annotated
	srcDir := filepath.Join(tempDir, "src")
	err := os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	// Create some actual files to annotate
	err = os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test.txt: %v", err)
	}

	// Create root .info file with annotations for both files
	rootInfo := `src/main.go: Project entry point
test.txt: Root test file`

	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create root .info file: %v", err)
	}

	// Create src/.info file with different annotation for main.go
	srcInfo := `main.go: Main function`

	err = os.WriteFile(filepath.Join(srcDir, ".info"), []byte(srcInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create src .info file: %v", err)
	}

	// Parse the directory tree
	annotations, err := ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("ParseDirectoryTree failed: %v", err)
	}

	// Verify we have 2 annotations
	if len(annotations) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(annotations))
		for path, ann := range annotations {
			t.Logf("Found: %s: %s", path, ann.Notes)
		}
	}

	// Check that src/main.go has the annotation from the deeper .info file
	if mainAnnotation, exists := annotations["src/main.go"]; !exists {
		t.Error("src/main.go annotation not found")
	} else if mainAnnotation.Notes != "Main function" {
		t.Errorf("src/main.go should have annotation from deeper .info file.\nExpected: %q\nGot: %q",
			"Main function", mainAnnotation.Notes)
	}

	// Check that test.txt still has its root annotation
	if testAnnotation, exists := annotations["test.txt"]; !exists {
		t.Error("test.txt annotation not found")
	} else if testAnnotation.Notes != "Root test file" {
		t.Errorf("test.txt notes incorrect: %q", testAnnotation.Notes)
	}
}

func TestMultiLevelNestedInfoFilePrecedence(t *testing.T) {
	// Test precedence with multiple levels of nested .info files
	tempDir := t.TempDir()

	// Create a deeper directory structure
	deepDir := filepath.Join(tempDir, "src", "core", "handlers")
	err := os.MkdirAll(deepDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Create test files at various levels
	files := map[string]string{
		filepath.Join(tempDir, "README.md"):                         "# Project",
		filepath.Join(tempDir, "src", "main.go"):                    "package main",
		filepath.Join(tempDir, "src", "core", "core.go"):            "package core",
		filepath.Join(tempDir, "src", "core", "handlers", "api.go"): "package handlers",
	}

	for path, content := range files {
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Create .info files at various levels with overlapping annotations
	// Root level - annotates everything
	rootInfo := `README.md: Project documentation
src/main.go: Application entry point (root annotation)
src/core/core.go: Core package (root annotation)
src/core/handlers/api.go: API handlers (root annotation)`

	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create root .info: %v", err)
	}

	// src level - overrides main.go and adds annotations for deeper files
	srcInfo := `main.go: Main function implementation
core/core.go: Core business logic (src annotation)
core/handlers/api.go: HTTP handlers (src annotation)`

	err = os.WriteFile(filepath.Join(tempDir, "src", ".info"), []byte(srcInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create src .info: %v", err)
	}

	// core level - overrides core.go and api.go
	coreInfo := `core.go: Core domain models
handlers/api.go: REST API endpoints (core annotation)`

	err = os.WriteFile(filepath.Join(tempDir, "src", "core", ".info"), []byte(coreInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create core .info: %v", err)
	}

	// handlers level - final override for api.go
	handlersInfo := `api.go: RESTful API handler functions`

	err = os.WriteFile(filepath.Join(tempDir, "src", "core", "handlers", ".info"), []byte(handlersInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create handlers .info: %v", err)
	}

	// Parse the directory tree
	annotations, err := ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("ParseDirectoryTree failed: %v", err)
	}

	// Define expected results - deepest .info should win
	expected := map[string]string{
		"README.md":                "Project documentation",         // Only in root
		"src/main.go":              "Main function implementation",  // Overridden by src/.info
		"src/core/core.go":         "Core domain models",            // Overridden by core/.info
		"src/core/handlers/api.go": "RESTful API handler functions", // Overridden by handlers/.info
	}

	// Verify all expected annotations
	for path, expectedNotes := range expected {
		if ann, exists := annotations[path]; !exists {
			t.Errorf("Annotation for %s not found", path)
		} else if ann.Notes != expectedNotes {
			t.Errorf("Wrong annotation for %s.\nExpected: %q\nGot: %q",
				path, expectedNotes, ann.Notes)
		}
	}

	// Verify we have exactly the expected number of annotations
	if len(annotations) != len(expected) {
		t.Errorf("Expected %d annotations, got %d", len(expected), len(annotations))
		for path, ann := range annotations {
			t.Logf("Found: %s: %s", path, ann.Notes)
		}
	}
}
