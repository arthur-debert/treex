package info

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFileFormats(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantPaths   map[string]string // path -> expected annotation
		wantWarnings int
	}{
		{
			name: "whitespace format - simple",
			content: `README.md Project documentation
main.go Main entry point
src/app.go Application logic`,
			wantPaths: map[string]string{
				"README.md":  "Project documentation",
				"main.go":    "Main entry point",
				"src/app.go": "Application logic",
			},
			wantWarnings: 0,
		},
		{
			name: "whitespace format - multiple words in annotation",
			content: `README.md This is the main project documentation file
src/app.go Core application logic with business rules
test.txt A simple test file for testing purposes`,
			wantPaths: map[string]string{
				"README.md":  "This is the main project documentation file",
				"src/app.go": "Core application logic with business rules",
				"test.txt":   "A simple test file for testing purposes",
			},
			wantWarnings: 0,
		},
		{
			name: "colon format - for paths with spaces",
			content: `path with spaces/file.txt: This file has spaces in path
normal/path.go: Regular path with colon
another path/doc.md: Another file with spaces`,
			wantPaths: map[string]string{
				"path with spaces/file.txt": "This file has spaces in path",
				"normal/path.go":            "Regular path with colon",
				"another path/doc.md":       "Another file with spaces",
			},
			wantWarnings: 0,
		},
		{
			name: "mixed formats",
			content: `README.md Simple whitespace format
path with spaces/file.txt: Colon format for spaces
src/main.go Main application entry point
config/app settings.yaml: Configuration file with spaces
test.go Unit tests`,
			wantPaths: map[string]string{
				"README.md":                "Simple whitespace format",
				"path with spaces/file.txt": "Colon format for spaces",
				"src/main.go":              "Main application entry point",
				"config/app settings.yaml": "Configuration file with spaces",
				"test.go":                  "Unit tests",
			},
			wantWarnings: 0,
		},
		{
			name: "whitespace format with tabs",
			content: "file1.txt\tAnnotation with tab separator\nfile2.go\t\tMultiple tabs before annotation",
			wantPaths: map[string]string{
				"file1.txt": "Annotation with tab separator",
				"file2.go":  "Multiple tabs before annotation",
			},
			wantWarnings: 0,
		},
		{
			name: "edge cases - empty lines and missing annotations",
			content: `
valid.txt This is valid

missing_annotation

another.go Valid annotation
`,
			wantPaths: map[string]string{
				"valid.txt":  "This is valid",
				"another.go": "Valid annotation",
			},
			wantWarnings: 1, // missing_annotation line
		},
		{
			name: "colon format handles colons in annotation",
			content: `file.txt: This has: multiple colons in annotation
path/to/file.txt: URL in annotation: https://example.com`,
			wantPaths: map[string]string{
				"file.txt":        "This has: multiple colons in annotation",
				"path/to/file.txt": "URL in annotation: https://example.com",
			},
			wantWarnings: 0,
		},
		{
			name: "whitespace in annotations preserved",
			content: `file1.txt    Annotation   with   extra   spaces
file2.go Annotation with     trailing spaces    `,
			wantPaths: map[string]string{
				"file1.txt": "Annotation   with   extra   spaces",
				"file2.go":  "Annotation with     trailing spaces",
			},
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and .info file
			tempDir := t.TempDir()
			infoPath := filepath.Join(tempDir, ".info")
			
			err := os.WriteFile(infoPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test .info file: %v", err)
			}

			// Parse the file
			parser := NewParser()
			annotations, warnings, err := parser.ParseFileWithWarnings(infoPath)
			if err != nil {
				t.Fatalf("ParseFileWithWarnings failed: %v", err)
			}

			// Check warnings count
			if len(warnings) != tt.wantWarnings {
				t.Errorf("Expected %d warnings, got %d: %v", tt.wantWarnings, len(warnings), warnings)
			}

			// Check annotations
			if len(annotations) != len(tt.wantPaths) {
				t.Errorf("Expected %d annotations, got %d", len(tt.wantPaths), len(annotations))
			}

			for path, wantNotes := range tt.wantPaths {
				ann, exists := annotations[path]
				if !exists {
					t.Errorf("Missing annotation for path %q", path)
					continue
				}
				if ann.Notes != wantNotes {
					t.Errorf("Path %q: expected notes %q, got %q", path, wantNotes, ann.Notes)
				}
			}

			// Check for unexpected annotations
			for path := range annotations {
				if _, expected := tt.wantPaths[path]; !expected {
					t.Errorf("Unexpected annotation for path %q", path)
				}
			}
		})
	}
}

// Test backward compatibility with existing .info files
func TestParseFileBackwardCompatibility(t *testing.T) {
	// Test that old colon-only format still works
	oldFormatContent := `src/main.go: Application entry point
pkg/parser.go: Parser implementation
test/test.go: Unit tests
docs/readme.md: Documentation`

	tempDir := t.TempDir()
	infoPath := filepath.Join(tempDir, ".info")
	
	err := os.WriteFile(infoPath, []byte(oldFormatContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .info file: %v", err)
	}

	parser := NewParser()
	annotations, warnings, err := parser.ParseFileWithWarnings(infoPath)
	if err != nil {
		t.Fatalf("ParseFileWithWarnings failed: %v", err)
	}

	if len(warnings) > 0 {
		t.Errorf("Old format should not produce warnings: %v", warnings)
	}

	expectedCount := 4
	if len(annotations) != expectedCount {
		t.Errorf("Expected %d annotations, got %d", expectedCount, len(annotations))
	}

	// Verify all old format entries are parsed correctly
	expectedAnnotations := map[string]string{
		"src/main.go":   "Application entry point",
		"pkg/parser.go": "Parser implementation",
		"test/test.go":  "Unit tests",
		"docs/readme.md": "Documentation",
	}

	for path, wantNotes := range expectedAnnotations {
		ann, exists := annotations[path]
		if !exists {
			t.Errorf("Missing annotation for path %q", path)
			continue
		}
		if ann.Notes != wantNotes {
			t.Errorf("Path %q: expected notes %q, got %q", path, wantNotes, ann.Notes)
		}
	}
}