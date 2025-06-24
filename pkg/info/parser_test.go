package info

import (
	"os"
	"path/filepath"
	"strings"
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

func TestParseDirectoryTree(t *testing.T) {
	// Create a temporary directory structure with multiple .info files
	tempDir := t.TempDir()

	// Create root .info file
	rootInfo := `README.md
Root level readme file

main.go
Main application entry point`

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

	subInfo := `parser.go
Handles parsing of .info files

builder.go
Constructs file trees from filesystem`

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

	nestedInfo := `config.json
Deep configuration file`

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
	} else if readme.Description != "Root level readme file" {
		t.Errorf("Root README.md annotation incorrect: %q", readme.Description)
	}

	// Check internal level annotations (should have "internal/" prefix)
	if parser, exists := annotations["internal/parser.go"]; !exists {
		t.Error("internal/parser.go annotation not found")
	} else if parser.Description != "Handles parsing of .info files" {
		t.Errorf("internal/parser.go annotation incorrect: %q", parser.Description)
	}

	// Check nested level annotations (should have "internal/deep/" prefix)
	if config, exists := annotations["internal/deep/config.json"]; !exists {
		t.Error("internal/deep/config.json annotation not found")
	} else if config.Description != "Deep configuration file" {
		t.Errorf("internal/deep/config.json annotation incorrect: %q", config.Description)
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

	infoContent := `file.txt
A test file

nested/deep.txt
A deeply nested file`

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
	}

	if _, exists := annotations["sub/nested/deep.txt"]; !exists {
		t.Error("sub/nested/deep.txt annotation not found")
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
	infoContent := `../../../etc/passwd
Dangerous path attempt

file.txt
Safe file

../parent.txt
Another dangerous path`

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

func TestGenerateInfoFromReader(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create the directory structure first
	dirs := []string{
		"test-project",
		"test-project/src",
		"test-project/docs",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(tempDir, dir), 0755)
		if err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create files
	files := []string{
		"test-project/README.md",
		"test-project/src/main.go",
		"test-project/docs/guide.md",
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("failed to create file %s: %v", file, err)
		}
		_ = f.Close()
	}

	// Change to temp directory so relative paths work
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Test content with tree structure
	content := `test-project
├── src/ Source code directory
│   └── main.go Main application file
├── docs/ Documentation directory  
│   └── guide.md User guide
└── README.md Project documentation`

	reader := strings.NewReader(content)

	// Test GenerateInfoFromReader
	err = GenerateInfoFromReader(reader)
	if err != nil {
		t.Fatalf("GenerateInfoFromReader failed: %v", err)
	}

	// Verify .info files were created
	expectedInfoFiles := []string{
		".info",
		"test-project/.info",
		"test-project/src/.info",
		"test-project/docs/.info",
	}

	for _, infoFile := range expectedInfoFiles {
		if _, err := os.Stat(infoFile); os.IsNotExist(err) {
			t.Errorf("expected .info file %s to be created", infoFile)
		}
	}

	// Verify content of root .info file
	rootInfoContent, err := os.ReadFile(".info")
	if err != nil {
		t.Fatalf("failed to read root .info file: %v", err)
	}

	rootContent := string(rootInfoContent)
	if !strings.Contains(rootContent, "test-project") {
		t.Errorf("expected root .info to contain 'test-project', got: %s", rootContent)
	}

	// Verify content of test-project/.info file
	projectInfoContent, err := os.ReadFile("test-project/.info")
	if err != nil {
		t.Fatalf("failed to read test-project/.info file: %v", err)
	}

	projectContent := string(projectInfoContent)
	expectedEntries := []string{"src/", "docs/", "README.md"}
	for _, entry := range expectedEntries {
		if !strings.Contains(projectContent, entry) {
			t.Errorf("expected test-project/.info to contain '%s', got: %s", entry, projectContent)
		}
	}
}

func TestParseFileWithSpaceFormat(t *testing.T) {
	// Test the new format where path and title are on the same line separated by space
	tempDir := t.TempDir()
	infoFile := filepath.Join(tempDir, ".info")

	content := `README.md Like the title says, that useful little readme.

LICENSE MIT license file

.github/workflows/go.yml CI Unit test workflow
This makes usage of go action, that does pretty much all go setup.
Note that his has no caching just yet.

config.json Configuration file
Contains database settings
and API keys.

single.txt Just a title with no description`

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
	expectedCount := 5
	if len(annotations) != expectedCount {
		t.Errorf("Expected %d annotations, got %d", expectedCount, len(annotations))
		for path, ann := range annotations {
			t.Logf("Found: %s -> Title: %q, Desc: %q", path, ann.Title, ann.Description)
		}
	}

	// Test README.md annotation (single line with space format)
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

	// Test LICENSE annotation (single line with space format)
	license, exists := annotations["LICENSE"]
	if !exists {
		t.Error("LICENSE annotation not found")
	} else {
		expectedTitle := "MIT license file"
		expectedDesc := "MIT license file"
		if license.Title != expectedTitle {
			t.Errorf("LICENSE title mismatch.\nExpected: %q\nGot: %q", expectedTitle, license.Title)
		}
		if license.Description != expectedDesc {
			t.Errorf("LICENSE description mismatch.\nExpected: %q\nGot: %q", expectedDesc, license.Description)
		}
	}

	// Test .github/workflows/go.yml annotation (space format with additional description)
	workflow, exists := annotations[".github/workflows/go.yml"]
	if !exists {
		t.Error(".github/workflows/go.yml annotation not found")
	} else {
		expectedTitle := "CI Unit test workflow"
		expectedDesc := "This makes usage of go action, that does pretty much all go setup.\nNote that his has no caching just yet."
		if workflow.Title != expectedTitle {
			t.Errorf("Workflow title mismatch.\nExpected: %q\nGot: %q", expectedTitle, workflow.Title)
		}
		if workflow.Description != expectedDesc {
			t.Errorf("Workflow description mismatch.\nExpected: %q\nGot: %q", expectedDesc, workflow.Description)
		}
	}

	// Test config.json annotation (space format with multi-line description)
	config, exists := annotations["config.json"]
	if !exists {
		t.Error("config.json annotation not found")
	} else {
		expectedTitle := "Configuration file"
		expectedDesc := "Contains database settings\nand API keys."
		if config.Title != expectedTitle {
			t.Errorf("Config title mismatch.\nExpected: %q\nGot: %q", expectedTitle, config.Title)
		}
		if config.Description != expectedDesc {
			t.Errorf("Config description mismatch.\nExpected: %q\nGot: %q", expectedDesc, config.Description)
		}
	}

	// Test single.txt annotation (space format with only title)
	single, exists := annotations["single.txt"]
	if !exists {
		t.Error("single.txt annotation not found")
	} else {
		expectedTitle := "Just a title with no description"
		expectedDesc := "Just a title with no description"
		if single.Title != expectedTitle {
			t.Errorf("Single title mismatch.\nExpected: %q\nGot: %q", expectedTitle, single.Title)
		}
		if single.Description != expectedDesc {
			t.Errorf("Single description mismatch.\nExpected: %q\nGot: %q", expectedDesc, single.Description)
		}
	}
}

func TestParseFileMixedFormats(t *testing.T) {
	// Test mixed traditional and space formats in the same file
	tempDir := t.TempDir()
	infoFile := filepath.Join(tempDir, ".info")

	content := `README.md
Traditional format description

LICENSE MIT license
More description here

src/main.go Main application file
With additional description

docs/
Traditional directory format
Contains all documentation files

bin/ Binary output directory`

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
	expectedCount := 5
	if len(annotations) != expectedCount {
		t.Errorf("Expected %d annotations, got %d", expectedCount, len(annotations))
		for path, ann := range annotations {
			t.Logf("Found: %s -> Title: %q, Desc: %q", path, ann.Title, ann.Description)
		}
	}

	// Test README.md (traditional format)
	readme, exists := annotations["README.md"]
	if !exists {
		t.Error("README.md annotation not found")
	} else {
		expectedDesc := "Traditional format description"
		if readme.Description != expectedDesc {
			t.Errorf("README.md description mismatch.\nExpected: %q\nGot: %q", expectedDesc, readme.Description)
		}
		if readme.Title != "" {
			t.Errorf("README.md should not have a title (single line traditional), got: %q", readme.Title)
		}
	}

	// Test LICENSE (space format)
	license, exists := annotations["LICENSE"]
	if !exists {
		t.Error("LICENSE annotation not found")
	} else {
		expectedTitle := "MIT license"
		expectedDesc := "More description here"
		if license.Title != expectedTitle {
			t.Errorf("LICENSE title mismatch.\nExpected: %q\nGot: %q", expectedTitle, license.Title)
		}
		if license.Description != expectedDesc {
			t.Errorf("LICENSE description mismatch.\nExpected: %q\nGot: %q", expectedDesc, license.Description)
		}
	}

	// Test src/main.go (space format with additional description)
	main, exists := annotations["src/main.go"]
	if !exists {
		t.Error("src/main.go annotation not found")
	} else {
		expectedTitle := "Main application file"
		expectedDesc := "With additional description"
		if main.Title != expectedTitle {
			t.Errorf("Main title mismatch.\nExpected: %q\nGot: %q", expectedTitle, main.Title)
		}
		if main.Description != expectedDesc {
			t.Errorf("Main description mismatch.\nExpected: %q\nGot: %q", expectedDesc, main.Description)
		}
	}

	// Test docs/ (traditional format)
	docs, exists := annotations["docs/"]
	if !exists {
		t.Error("docs/ annotation not found")
	} else {
		expectedTitle := "Traditional directory format"
		expectedDesc := "Traditional directory format\nContains all documentation files"
		if docs.Title != expectedTitle {
			t.Errorf("Docs title mismatch.\nExpected: %q\nGot: %q", expectedTitle, docs.Title)
		}
		if docs.Description != expectedDesc {
			t.Errorf("Docs description mismatch.\nExpected: %q\nGot: %q", expectedDesc, docs.Description)
		}
	}

	// Test bin/ (space format only title)
	bin, exists := annotations["bin/"]
	if !exists {
		t.Error("bin/ annotation not found")
	} else {
		expectedTitle := "Binary output directory"
		expectedDesc := "Binary output directory"
		if bin.Title != expectedTitle {
			t.Errorf("Bin title mismatch.\nExpected: %q\nGot: %q", expectedTitle, bin.Title)
		}
		if bin.Description != expectedDesc {
			t.Errorf("Bin description mismatch.\nExpected: %q\nGot: %q", expectedDesc, bin.Description)
		}
	}
}

func TestGenerateInfoFromReader_EmptyInput(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Test with empty content
	reader := strings.NewReader("")

	// This should not fail but also shouldn't create any files
	err = GenerateInfoFromReader(reader)
	if err != nil {
		t.Fatalf("GenerateInfoFromReader with empty input failed: %v", err)
	}

	// Verify no .info files were created
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if entry.Name() == ".info" {
			t.Error("expected no .info file to be created for empty input")
		}
	}
}
