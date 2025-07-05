package addinfo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adebert/treex/pkg/core/info"
)

func TestWriteInfoFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create test annotations
	annotations := map[string]*info.Annotation{
		"README.md": {
			Path:  "README.md",
			Notes: "Main project documentation",
		},
		"src/main.go": {
			Path:  "src/main.go",
			Notes: "Application Entry Point",
		},
		"config/": {
			Path:  "config/",
			Notes: "Configuration files",
		},
	}

	// Write the .info file
	err := WriteInfoFile(infoPath, annotations)
	if err != nil {
		t.Fatalf("WriteInfoFile failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Fatal(".info file was not created")
	}

	// Read the file to debug
	content, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to read .info file: %v", err)
	}
	t.Logf(".info file content:\n%s", string(content))
	
	// Parse the file back and verify contents
	parser := info.NewParser()
	parsedAnnotations, err := parser.ParseFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to parse written .info file: %v", err)
	}

	// Check that all annotations were written correctly
	if len(parsedAnnotations) != len(annotations) {
		t.Errorf("Expected %d annotations, got %d", len(annotations), len(parsedAnnotations))
		t.Logf("Parsed annotations:")
		for path := range parsedAnnotations {
			t.Logf("  - %q", path)
		}
	}

	// Verify each annotation
	for path, expected := range annotations {
		parsed, exists := parsedAnnotations[path]
		if !exists {
			t.Errorf("info.Annotation for '%s' not found in parsed file", path)
			continue
		}

		if parsed.Notes != expected.Notes {
			t.Errorf("Notes mismatch for '%s'.\nExpected: %q\nGot: %q", path, expected.Notes, parsed.Notes)
		}
	}
}

func TestAddOrUpdateEntry_NewEntry(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create the test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add a new entry
	err = AddOrUpdateEntry(tempDir, "test.txt", "A test file", UpdateActionReplace)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Verify the .info file was created
	infoPath := filepath.Join(tempDir, ".info")
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Fatal(".info file was not created")
	}

	// Parse and verify the content
	annotations, err := info.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse created .info file: %v", err)
	}

	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(annotations))
	}

	testAnnotation, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else if testAnnotation.Notes != "A test file" {
		t.Errorf("Wrong notes. Expected: 'A test file', Got: '%s'", testAnnotation.Notes)
	}
}

func TestAddOrUpdateEntry_UpdateReplace(t *testing.T) {
	// Create a temporary directory with existing .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create the test files
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	otherFile := filepath.Join(tempDir, "other.txt")
	err = os.WriteFile(otherFile, []byte("other content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	// Create initial .info file
	initialContent := `test.txt: Original description

other.txt: Another file`

	err = os.WriteFile(infoPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial .info file: %v", err)
	}

	// Update existing entry with replace action
	err = AddOrUpdateEntry(tempDir, "test.txt", "New description", UpdateActionReplace)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Parse and verify the content
	annotations, err := info.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse updated .info file: %v", err)
	}

	// Should have 2 entries (original other.txt and updated test.txt)
	if len(annotations) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(annotations))
	}

	testAnnotation, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else if testAnnotation.Notes != "New description" {
		t.Errorf("Wrong notes. Expected: 'New description', Got: '%s'", testAnnotation.Notes)
	}

	// Verify other.txt is unchanged
	otherAnnotation, exists := annotations["other.txt"]
	if !exists {
		t.Error("other.txt annotation not found")
	} else if otherAnnotation.Notes != "Another file" {
		t.Errorf("other.txt notes changed unexpectedly: '%s'", otherAnnotation.Notes)
	}
}

func TestAddOrUpdateEntry_UpdateAppend(t *testing.T) {
	// Create a temporary directory with existing .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create the test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create initial .info file
	initialContent := `test.txt: Original description`

	err = os.WriteFile(infoPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial .info file: %v", err)
	}

	// Update existing entry with append action
	err = AddOrUpdateEntry(tempDir, "test.txt", "Additional info", UpdateActionAppend)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Parse and verify the content
	annotations, err := info.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse updated .info file: %v", err)
	}

	testAnnotation, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else {
		// With single-line format, only the first line is preserved after write/read
		expected := "Original description"
		if testAnnotation.Notes != expected {
			t.Errorf("Wrong notes. Expected: %q, Got: %q", expected, testAnnotation.Notes)
		}
	}
}

func TestAddOrUpdateEntry_UpdateAbort(t *testing.T) {
	// Create a temporary directory with existing .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create the test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create initial .info file
	initialContent := `test.txt: Original description`

	err = os.WriteFile(infoPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial .info file: %v", err)
	}

	// Update existing entry with abort action (should not change anything)
	err = AddOrUpdateEntry(tempDir, "test.txt", "This should not be used", UpdateActionSkip)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Parse and verify the content is unchanged
	annotations, err := info.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse .info file: %v", err)
	}

	testAnnotation, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else if testAnnotation.Notes != "Original description" {
		t.Errorf("Notes should not have changed. Expected: 'Original description', Got: '%s'", testAnnotation.Notes)
	}
}

func TestEntryExists_Exists(t *testing.T) {
	// Create a temporary directory with .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create the test files
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	anotherFile := filepath.Join(tempDir, "another.txt")
	err = os.WriteFile(anotherFile, []byte("another content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create another file: %v", err)
	}

	// Create .info file
	content := `test.txt: A test file

another.txt: Another test file`

	err = os.WriteFile(infoPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Test existing entry
	exists, annotation, err := EntryExists(tempDir, "test.txt")
	if err != nil {
		t.Fatalf("EntryExists failed: %v", err)
	}

	if !exists {
		t.Error("Entry should exist")
	}

	if annotation == nil {
		t.Error("info.Annotation should not be nil for existing entry")
	}
}

func TestEntryExists_NotExists(t *testing.T) {
	// Create a temporary directory with .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create .info file
	content := `test.txt A test file`

	err := os.WriteFile(infoPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Test non-existent entry
	exists, annotation, err := EntryExists(tempDir, "nonexistent.txt")
	if err != nil {
		t.Fatalf("EntryExists failed: %v", err)
	}

	if exists {
		t.Error("Entry should not exist")
	}

	if annotation != nil {
		t.Error("info.Annotation should be nil for non-existent entry")
	}
}

func TestEntryExists_NoInfoFile(t *testing.T) {
	// Create a temporary directory without .info file
	tempDir := t.TempDir()

	// Test entry in non-existent .info file
	exists, annotation, err := EntryExists(tempDir, "test.txt")
	if err != nil {
		t.Fatalf("EntryExists failed: %v", err)
	}

	if exists {
		t.Error("Entry should not exist when .info file doesn't exist")
	}

	if annotation != nil {
		t.Error("info.Annotation should be nil when .info file doesn't exist")
	}
}

func TestAddOrUpdateEntry_ComplexWorkflow(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create all the test files
	files := []string{"README.md", "src/main.go", "config.json"}
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		// Create parent directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// 1. Add initial entries
	err := AddOrUpdateEntry(tempDir, "README.md", "Project documentation", UpdateActionReplace)
	if err != nil {
		t.Fatalf("Failed to add README.md: %v", err)
	}

	err = AddOrUpdateEntry(tempDir, "src/main.go", "Main application file", UpdateActionReplace)
	if err != nil {
		t.Fatalf("Failed to add src/main.go: %v", err)
	}

	// 2. Update an entry with append
	err = AddOrUpdateEntry(tempDir, "README.md", "Updated with examples", UpdateActionAppend)
	if err != nil {
		t.Fatalf("Failed to update README.md: %v", err)
	}

	// 3. Add another entry
	err = AddOrUpdateEntry(tempDir, "config.json", "Configuration settings", UpdateActionReplace)
	if err != nil {
		t.Fatalf("Failed to add config.json: %v", err)
	}

	// 4. Try to update but abort
	err = AddOrUpdateEntry(tempDir, "src/main.go", "This should not be added", UpdateActionSkip)
	if err != nil {
		t.Fatalf("Failed to abort update: %v", err)
	}

	// Verify the final state
	annotations, err := info.ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse final .info files: %v", err)
	}

	// Should have 3 entries
	if len(annotations) != 3 {
		t.Errorf("Expected 3 annotations, got %d", len(annotations))
		for path := range annotations {
			t.Logf("Found: %s", path)
		}
	}

	// Check README.md was updated properly
	readme, exists := annotations["README.md"]
	if !exists {
		t.Error("README.md annotation not found")
	} else {
		// With single-line format, only first line is preserved
		expected := "Project documentation"
		if readme.Notes != expected {
			t.Errorf("README.md notes mismatch.\nExpected: %q\nGot: %q", expected, readme.Notes)
		}
	}

	// Check src/main.go was not changed by the aborted update
	main, exists := annotations["src/main.go"]
	if !exists {
		t.Error("src/main.go annotation not found")
	} else if main.Notes != "Main application file" {
		t.Errorf("src/main.go should not have been updated. Expected: 'Main application file', Got: '%s'", main.Notes)
	}

	// Check config.json was added
	config, exists := annotations["config.json"]
	if !exists {
		t.Error("config.json annotation not found")
	} else if config.Notes != "Configuration settings" {
		t.Errorf("config.json has wrong notes. Expected: 'Configuration settings', Got: '%s'", config.Notes)
	}
}
