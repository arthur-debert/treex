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
	
	content := `README.md
Like the title says, that useful little readme.

LICENSE
MIT, like most things.

.github/workflows/go.yml
CI Unit test workflow
This makes usage of go action, that does pretty much all go setup.
Note that his has no caching just yet.`

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
		expectedDesc := "Like the title says, that useful little readme."
		if readme.Description != expectedDesc {
			t.Errorf("README.md description mismatch.\nExpected: %q\nGot: %q", expectedDesc, readme.Description)
		}
		if readme.Title != "" {
			t.Errorf("README.md should not have a title (single line), got: %q", readme.Title)
		}
	}

	// Test LICENSE annotation
	license, exists := annotations["LICENSE"]
	if !exists {
		t.Error("LICENSE annotation not found")
	} else {
		expectedDesc := "MIT, like most things."
		if license.Description != expectedDesc {
			t.Errorf("LICENSE description mismatch.\nExpected: %q\nGot: %q", expectedDesc, license.Description)
		}
	}

	// Test .github/workflows/go.yml annotation (multi-line)
	workflow, exists := annotations[".github/workflows/go.yml"]
	if !exists {
		t.Error(".github/workflows/go.yml annotation not found")
	} else {
		expectedTitle := "CI Unit test workflow"
		if workflow.Title != expectedTitle {
			t.Errorf("Workflow title mismatch.\nExpected: %q\nGot: %q", expectedTitle, workflow.Title)
		}
		expectedDesc := "CI Unit test workflow\nThis makes usage of go action, that does pretty much all go setup.\nNote that his has no caching just yet."
		if workflow.Description != expectedDesc {
			t.Errorf("Workflow description mismatch.\nExpected: %q\nGot: %q", expectedDesc, workflow.Description)
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
	
	content := `test.txt
A test file.`

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
		expectedDesc := "A test file."
		if testFile.Description != expectedDesc {
			t.Errorf("Description mismatch.\nExpected: %q\nGot: %q", expectedDesc, testFile.Description)
		}
	}
} 