package info

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFileWithWarnings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Test various error cases that should produce warnings
	tests := []struct {
		name            string
		infoContent     string
		expectedWarnings []string
		expectedAnnotations map[string]string // path -> notes
	}{
		{
			name: "malformed lines without colon",
			infoContent: `README.md: Valid annotation
docs/api.md
src/main.go: Another valid annotation
docs/guide.md`,
			expectedWarnings: []string{
				"Line 2: Invalid format (missing annotation): \"docs/api.md\"",
				"Line 4: Invalid format (missing annotation): \"docs/guide.md\"",
			},
			expectedAnnotations: map[string]string{
				"README.md": "Valid annotation",
				"src/main.go": "Another valid annotation",
			},
		},
		{
			name: "empty path or notes",
			infoContent: `: Empty path annotation
test/:
src/utils.go:
: Both empty
  :   Spaces only`,
			expectedWarnings: []string{
				"Line 1: Empty path in annotation",
				"Line 2: Empty notes for path \"test/\"",
				"Line 3: Empty notes for path \"src/utils.go\"",
				"Line 4: Empty path in annotation",
				"Line 5: Empty path in annotation",
			},
			expectedAnnotations: map[string]string{},
		},
		{
			name: "mixed valid and invalid lines",
			infoContent: `src/main.go: Entry point
This
test/test.go: Unit tests
src/deleted.go: File that was removed
Another`,
			expectedWarnings: []string{
				"Line 2: Invalid format (missing annotation): \"This\"",
				"Line 5: Invalid format (missing annotation): \"Another\"",
			},
			expectedAnnotations: map[string]string{
				"src/main.go": "Entry point",
				"test/test.go": "Unit tests",
				"src/deleted.go": "File that was removed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create .info file
			infoFile := filepath.Join(tempDir, tt.name+".info")
			err := os.WriteFile(infoFile, []byte(tt.infoContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test .info file: %v", err)
			}

			// Parse file with warnings collection
			parser := NewParser()
			annotations, warnings, err := parser.ParseFileWithWarnings(infoFile)
			if err != nil {
				t.Fatalf("ParseFileWithWarnings failed: %v", err)
			}

			// Check annotations match expected
			if len(annotations) != len(tt.expectedAnnotations) {
				t.Errorf("Expected %d annotations, got %d", len(tt.expectedAnnotations), len(annotations))
			}
			for path, expectedNotes := range tt.expectedAnnotations {
				if ann, exists := annotations[path]; !exists {
					t.Errorf("Expected annotation for path %q not found", path)
				} else if ann.Notes != expectedNotes {
					t.Errorf("Notes mismatch for %q:\nExpected: %q\nGot: %q", path, expectedNotes, ann.Notes)
				}
			}

			// Check warnings match expected
			if len(warnings) != len(tt.expectedWarnings) {
				t.Errorf("Expected %d warnings, got %d", len(tt.expectedWarnings), len(warnings))
				for i, w := range warnings {
					t.Logf("Warning %d: %s", i+1, w)
				}
			}
			
			// Check warning content (order matters)
			for i, expectedWarning := range tt.expectedWarnings {
				if i >= len(warnings) {
					t.Errorf("Missing warning at index %d: %q", i, expectedWarning)
					continue
				}
				if warnings[i] != expectedWarning {
					t.Errorf("Warning mismatch at index %d:\nExpected: %q\nGot: %q", i, expectedWarning, warnings[i])
				}
			}
		})
	}
}

func TestParseDirectoryTreeWithWarnings(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	
	// Create some actual files
	err := os.MkdirAll(filepath.Join(tempDir, "src"), 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "src", "main.go"), []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# README"), 0644)
	if err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	// Create root .info file with some issues
	rootInfo := `README.md: Project documentation
src/main.go: Entry point
src/deleted.go: This file doesn't exist
Invalid`
	
	err = os.WriteFile(filepath.Join(tempDir, ".info"), []byte(rootInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create root .info file: %v", err)
	}

	// Create src/.info with more issues
	srcInfo := `main.go: Main application
helper.go: Helper functions
: Empty path
../escape.go: Trying to escape directory`
	
	err = os.WriteFile(filepath.Join(tempDir, "src", ".info"), []byte(srcInfo), 0644)
	if err != nil {
		t.Fatalf("Failed to create src .info file: %v", err)
	}

	// Parse directory tree with warnings
	annotations, warnings, err := ParseDirectoryTreeWithWarnings(tempDir)
	if err != nil {
		t.Fatalf("ParseDirectoryTreeWithWarnings failed: %v", err)
	}

	// Should have annotations despite warnings
	expectedAnnotations := map[string]string{
		"README.md": "Project documentation",
		"src/main.go": "Main application", // From src/.info (overrides root)
		"src/helper.go": "Helper functions",
	}

	for path, expectedNotes := range expectedAnnotations {
		if ann, exists := annotations[path]; !exists {
			t.Errorf("Expected annotation for path %q not found", path)
		} else if ann.Notes != expectedNotes {
			t.Errorf("Notes mismatch for %q:\nExpected: %q\nGot: %q", path, expectedNotes, ann.Notes)
		}
	}

	// Check that we have warnings
	if len(warnings) == 0 {
		t.Error("Expected warnings but got none")
	}

	// Check for specific warning types
	hasInvalidFormatWarning := false
	hasNonExistentPathWarning := false
	hasEmptyPathWarning := false
	
	for _, w := range warnings {
		if strings.Contains(w, "Invalid format (missing annotation)") {
			hasInvalidFormatWarning = true
		}
		if strings.Contains(w, "Path not found") && strings.Contains(w, "src/deleted.go") {
			hasNonExistentPathWarning = true
		}
		if strings.Contains(w, "Empty path") {
			hasEmptyPathWarning = true
		}
		t.Logf("Warning: %s", w)
	}

	if !hasInvalidFormatWarning {
		t.Error("Expected invalid format warning")
	}
	if !hasNonExistentPathWarning {
		t.Error("Expected non-existent path warning")
	}
	if !hasEmptyPathWarning {
		t.Error("Expected empty path warning")
	}
}