package info

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteInfoFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create test annotations
	annotations := map[string]*Annotation{
		"README.md": {
			Path:        "README.md",
			Title:       "Main project documentation",
			Description: "Main project documentation",
		},
		"src/main.go": {
			Path:        "src/main.go",
			Title:       "Application Entry Point",
			Description: "Application Entry Point\nHandles command line arguments",
		},
		"config/": {
			Path:        "config/",
			Title:       "Configuration files",
			Description: "Configuration files",
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

	// Parse the file back and verify contents
	parser := NewParser()
	parsedAnnotations, err := parser.ParseFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to parse written .info file: %v", err)
	}

	// Check that all annotations were written correctly
	if len(parsedAnnotations) != len(annotations) {
		t.Errorf("Expected %d annotations, got %d", len(annotations), len(parsedAnnotations))
	}

	// Verify each annotation
	for path, expected := range annotations {
		parsed, exists := parsedAnnotations[path]
		if !exists {
			t.Errorf("Annotation for '%s' not found in parsed file", path)
			continue
		}

		if parsed.Description != expected.Description {
			t.Errorf("Description mismatch for '%s'.\nExpected: %q\nGot: %q", path, expected.Description, parsed.Description)
		}

		if parsed.Title != expected.Title {
			t.Errorf("Title mismatch for '%s'.\nExpected: %q\nGot: %q", path, expected.Title, parsed.Title)
		}
	}
}

func TestAddOrUpdateEntry_NewEntry(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Add a new entry
	err := AddOrUpdateEntry(tempDir, "test.txt", "A test file", UpdateActionReplace)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Verify the .info file was created
	infoPath := filepath.Join(tempDir, ".info")
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Fatal(".info file was not created")
	}

	// Parse and verify the content
	annotations, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse created .info file: %v", err)
	}

	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(annotations))
	}

	testFile, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else if testFile.Description != "A test file" {
		t.Errorf("Wrong description. Expected: 'A test file', Got: '%s'", testFile.Description)
	}
}

func TestAddOrUpdateEntry_UpdateReplace(t *testing.T) {
	// Create a temporary directory with existing .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create initial .info file
	initialContent := `test.txt Original description

other.txt Another file`

	err := os.WriteFile(infoPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial .info file: %v", err)
	}

	// Update existing entry with replace action
	err = AddOrUpdateEntry(tempDir, "test.txt", "New description", UpdateActionReplace)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Parse and verify the content
	annotations, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse updated .info file: %v", err)
	}

	// Should have 2 entries (original other.txt and updated test.txt)
	if len(annotations) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(annotations))
	}

	testFile, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else if testFile.Description != "New description" {
		t.Errorf("Wrong description. Expected: 'New description', Got: '%s'", testFile.Description)
	}

	// Verify other.txt is unchanged
	otherFile, exists := annotations["other.txt"]
	if !exists {
		t.Error("other.txt annotation not found")
	} else if otherFile.Description != "Another file" {
		t.Errorf("other.txt description changed unexpectedly: '%s'", otherFile.Description)
	}
}

func TestAddOrUpdateEntry_UpdateAppend(t *testing.T) {
	// Create a temporary directory with existing .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create initial .info file
	initialContent := `test.txt Original description`

	err := os.WriteFile(infoPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial .info file: %v", err)
	}

	// Update existing entry with append action
	err = AddOrUpdateEntry(tempDir, "test.txt", "Additional info", UpdateActionAppend)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Parse and verify the content
	annotations, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse updated .info file: %v", err)
	}

	testFile, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else {
		expected := "Original description\nAdditional info"
		if testFile.Description != expected {
			t.Errorf("Wrong description. Expected: %q, Got: %q", expected, testFile.Description)
		}
	}
}

func TestAddOrUpdateEntry_UpdateAbort(t *testing.T) {
	// Create a temporary directory with existing .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create initial .info file
	initialContent := `test.txt Original description`

	err := os.WriteFile(infoPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial .info file: %v", err)
	}

	// Update existing entry with abort action (should not change anything)
	err = AddOrUpdateEntry(tempDir, "test.txt", "This should not be used", UpdateActionAbort)
	if err != nil {
		t.Fatalf("AddOrUpdateEntry failed: %v", err)
	}

	// Parse and verify the content is unchanged
	annotations, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse .info file: %v", err)
	}

	testFile, exists := annotations["test.txt"]
	if !exists {
		t.Error("test.txt annotation not found")
	} else if testFile.Description != "Original description" {
		t.Errorf("Description should not have changed. Expected: 'Original description', Got: '%s'", testFile.Description)
	}
}

func TestEntryExists_Exists(t *testing.T) {
	// Create a temporary directory with .info file
	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")

	// Create .info file
	content := `test.txt A test file

another.txt Another test file`

	err := os.WriteFile(infoPath, []byte(content), 0644)
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
		t.Error("Annotation should not be nil for existing entry")
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
		t.Error("Annotation should be nil for non-existent entry")
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
		t.Error("Annotation should be nil when .info file doesn't exist")
	}
}

func TestAddOrUpdateEntry_ComplexWorkflow(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

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
	err = AddOrUpdateEntry(tempDir, "src/main.go", "This should not be added", UpdateActionAbort)
	if err != nil {
		t.Fatalf("Failed to abort update: %v", err)
	}

	// Verify the final state
	annotations, err := ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse final .info file: %v", err)
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
		expected := "Project documentation\nUpdated with examples"
		if readme.Description != expected {
			t.Errorf("README.md description mismatch.\nExpected: %q\nGot: %q", expected, readme.Description)
		}
	}

	// Check src/main.go was not changed by the aborted update
	main, exists := annotations["src/main.go"]
	if !exists {
		t.Error("src/main.go annotation not found")
	} else if main.Description != "Main application file" {
		t.Errorf("src/main.go should not have been updated. Expected: 'Main application file', Got: '%s'", main.Description)
	}

	// Check config.json was added
	config, exists := annotations["config.json"]
	if !exists {
		t.Error("config.json annotation not found")
	} else if config.Description != "Configuration settings" {
		t.Errorf("config.json has wrong description. Expected: 'Configuration settings', Got: '%s'", config.Description)
	}
}
