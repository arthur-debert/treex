package geninfo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/info"
)

func TestParseTreeLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectPath  string
		expectDesc  string
		expectIsDir bool
		expectDepth int
		expectError bool
	}{
		{
			name:        "simple file with description",
			line:        "├── main.go Application entry point",
			expectPath:  "main.go",
			expectDesc:  "Application entry point",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "directory with trailing slash",
			line:        "├── cmd/ Command line interface",
			expectPath:  "cmd",
			expectDesc:  "Command line interface",
			expectIsDir: true,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "nested directory",
			line:        "│   └── internal/ Internal packages",
			expectPath:  "internal",
			expectDesc:  "Internal packages",
			expectIsDir: true,
			expectDepth: 1,
			expectError: false,
		},
		{
			name:        "file without description",
			line:        "└── README.md",
			expectPath:  "README.md",
			expectDesc:  "",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "empty line should be skipped",
			line:        "",
			expectPath:  "",
			expectDesc:  "",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "line with only connectors",
			line:        "├──",
			expectPath:  "",
			expectDesc:  "",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, depth, err := ParseTreeLine(tt.line, []string{})
			
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectPath == "" && entry != nil {
				t.Errorf("expected nil entry for empty path, got %+v", entry)
				return
			}

			if tt.expectPath != "" && entry == nil {
				t.Errorf("expected entry but got nil")
				return
			}

			if entry != nil {
				if entry.Path != tt.expectPath {
					t.Errorf("expected path %q, got %q", tt.expectPath, entry.Path)
				}
				if entry.Description != tt.expectDesc {
					t.Errorf("expected description %q, got %q", tt.expectDesc, entry.Description)
				}
				if entry.IsDir != tt.expectIsDir {
					t.Errorf("expected IsDir %v, got %v", tt.expectIsDir, entry.IsDir)
				}
			}

			if depth != tt.expectDepth {
				t.Errorf("expected depth %d, got %d", tt.expectDepth, depth)
			}
		})
	}
}

func TestParseTreeFile(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-tree.txt")
	
	content := `myproject
├── cmd/ The go code for the cli utility
├── docs/ All documentation
│   └── dev/ Dev related, including technical topics
├── pkg/ The main parser code
├── scripts/ Various utilities
└── README.md Main documentation file`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	entries, err := ParseTreeFile(testFile)
	if err != nil {
		t.Fatalf("ParseTreeFile failed: %v", err)
	}

	expectedEntries := []struct {
		path        string
		description string
		isDir       bool
	}{
		{"myproject", "", false},
		{"myproject/cmd", "The go code for the cli utility", true},
		{"myproject/docs", "All documentation", true},
		{"myproject/docs/dev", "Dev related, including technical topics", true},
		{"myproject/pkg", "The main parser code", true},
		{"myproject/scripts", "Various utilities", true},
		{"myproject/README.md", "Main documentation file", false},
	}

	if len(entries) != len(expectedEntries) {
		t.Errorf("expected %d entries, got %d", len(expectedEntries), len(entries))
		for i, entry := range entries {
			t.Logf("Entry %d: %s -> %s (isDir: %v)", i, entry.Path, entry.Description, entry.IsDir)
		}
		return
	}

	for i, expected := range expectedEntries {
		if i >= len(entries) {
			t.Errorf("missing entry %d: %s", i, expected.path)
			continue
		}
		
		entry := entries[i]
		if entry.Path != expected.path {
			t.Errorf("entry %d: expected path %q, got %q", i, expected.path, entry.Path)
		}
		if entry.Description != expected.description {
			t.Errorf("entry %d: expected description %q, got %q", i, expected.description, entry.Description)
		}
		if entry.IsDir != expected.isDir {
			t.Errorf("entry %d: expected IsDir %v, got %v", i, expected.isDir, entry.IsDir)
		}
	}
}

func TestGenerateInfoFromTree(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	
	// Create the directory structure that matches our test input
	dirs := []string{
		"myproject",
		"myproject/cmd",
		"myproject/docs",
		"myproject/docs/dev",
		"myproject/pkg",
		"myproject/scripts",
	}
	
	for _, dir := range dirs {
		fullPath := filepath.Join(tempDir, dir)
		err := os.MkdirAll(fullPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", fullPath, err)
		}
	}
	
	// Create the README.md file
	readmePath := filepath.Join(tempDir, "myproject", "README.md")
	err := os.WriteFile(readmePath, []byte("# My Project"), 0644)
	if err != nil {
		t.Fatalf("Failed to create README.md: %v", err)
	}

	// Create the test input file
	testFile := filepath.Join(tempDir, "test-input.txt")
	content := `myproject
├── cmd/ The go code for the cli utility
├── docs/ All documentation
│   └── dev/ Dev related, including technical topics
├── pkg/ The main parser code
├── scripts/ Various utilities
└── README.md Main documentation file`

	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Change to temp directory to make paths work
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run the generator
	err = GenerateInfoFromTree(testFile)
	if err != nil {
		t.Fatalf("GenerateInfoFromTree failed: %v", err)
	}
	
	// Debug: List all .info files created
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == ".info" {
			t.Logf("Found .info file at: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Logf("Error walking directory: %v", err)
	}

	// Verify that .info files were created
	expectedInfoFiles := []struct {
		path     string
		contains []string
	}{
		{
			path: "myproject/.info",
			contains: []string{"cmd/", "The go code for the cli utility", "docs/", "All documentation", "pkg/", "The main parser code", "scripts/", "Various utilities", "README.md", "Main documentation file"},
		},
		{
			path: "myproject/docs/.info",
			contains: []string{"dev/", "Dev related, including technical topics"},
		},
	}

	for _, expected := range expectedInfoFiles {
		content, err := os.ReadFile(expected.path)
		if err != nil {
			t.Errorf("Failed to read .info file %s: %v", expected.path, err)
			continue
		}

		contentStr := string(content)
		for _, expectedText := range expected.contains {
			if !strings.Contains(contentStr, expectedText) {
				t.Errorf("Expected .info file %s to contain %q, but it doesn't. Content:\n%s", expected.path, expectedText, contentStr)
				if expected.path == ".info" && len(content) == 0 {
					t.Logf("Root .info file is empty")
				}
			}
		}
	}
	
	// Check that root .info exists but is empty (since myproject has no description)
	rootInfo, err := os.ReadFile(".info")
	if err != nil {
		t.Errorf("Root .info file should exist: %v", err)
	} else if len(rootInfo) > 0 {
		t.Errorf("Root .info file should be empty but contains: %s", string(rootInfo))
	}
}

func TestGenerateInfoFromTree_NonExistentPath(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	
	// Create the test input file with a non-existent path
	testFile := filepath.Join(tempDir, "test-input.txt")
	content := `myproject
├── nonexistent/ This directory does not exist
└── also-missing.txt This file does not exist`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test input file: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run the generator - it should fail with a descriptive error
	err = GenerateInfoFromTree(testFile)
	if err == nil {
		t.Error("Expected error for non-existent paths, but got none")
		return
	}

	if !strings.Contains(err.Error(), "path does not exist") {
		t.Errorf("Expected error message to contain 'path does not exist', got: %v", err)
	}
}

func TestGenerateInfoFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	entries := []TreeEntry{
		{Path: "cmd", Description: "Command line utilities", IsDir: true},
		{Path: "README.md", Description: "Main documentation", IsDir: false},
		{Path: "main.go", Description: "", IsDir: false},
	}

	err = GenerateInfoFile("", entries)
	if err != nil {
		t.Fatalf("GenerateInfoFile failed: %v", err)
	}

	// Check that .info file was created
	content, err := os.ReadFile(".info")
	if err != nil {
		t.Fatalf("Failed to read generated .info file: %v", err)
	}

	contentStr := string(content)
	expected := []string{
		"cmd/",
		"Command line utilities",
		"README.md",
		"Main documentation",
	}

	for _, expectedText := range expected {
		if !strings.Contains(contentStr, expectedText) {
			t.Errorf("Expected .info file to contain %q, but it doesn't. Content:\n%s", expectedText, contentStr)
		}
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

	// Verify root .info file exists but is empty (test-project has no description)
	rootInfoContent, err := os.ReadFile(".info")
	if err != nil {
		t.Fatalf("failed to read root .info file: %v", err)
	}

	if len(rootInfoContent) > 0 {
		t.Errorf("expected root .info to be empty since test-project has no description, got: %s", string(rootInfoContent))
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

func TestParseFileWithColonFormat(t *testing.T) {
	// Test the single-line format where path and description are separated by colon
	tempDir := t.TempDir()
	infoFile := filepath.Join(tempDir, ".info")

	// Updated to use colon format
	content := `README.md: Like the title says, that useful little readme.
LICENSE: MIT license file
.github/workflows/go.yml: CI Unit test workflow - makes usage of go action, that does pretty much all go setup. Note that his has no caching just yet.
config.json: Configuration file - Contains database settings and API keys.
single.txt: Just a title with no description`

	err := os.WriteFile(infoFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .info file: %v", err)
	}

	parser := info.NewParser()
	annotations, err := parser.ParseFile(infoFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check that we parsed the expected number of annotations
	expectedCount := 5
	if len(annotations) != expectedCount {
		t.Errorf("Expected %d annotations, got %d", expectedCount, len(annotations))
		for path, ann := range annotations {
			t.Logf("Found: %s -> Notes: %q", path, ann.Notes)
		}
	}

	// Test README.md annotation (single line with space format)
	readme, exists := annotations["README.md"]
	if !exists {
		t.Error("README.md annotation not found")
	} else {
		expectedNotes := "Like the title says, that useful little readme."
		if readme.Notes != expectedNotes {
			t.Errorf("README.md notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, readme.Notes)
		}
	}

	// Test LICENSE annotation (single line with space format)
	license, exists := annotations["LICENSE"]
	if !exists {
		t.Error("LICENSE annotation not found")
	} else {
		expectedNotes := "MIT license file"
		if license.Notes != expectedNotes {
			t.Errorf("LICENSE notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, license.Notes)
		}
	}

	// Test .github/workflows/go.yml annotation (single line format)
	workflow, exists := annotations[".github/workflows/go.yml"]
	if !exists {
		t.Error(".github/workflows/go.yml annotation not found")
	} else {
		expectedNotes := "CI Unit test workflow - makes usage of go action, that does pretty much all go setup. Note that his has no caching just yet."
		if workflow.Notes != expectedNotes {
			t.Errorf("Workflow notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, workflow.Notes)
		}
	}

	// Test config.json annotation (single line format)
	config, exists := annotations["config.json"]
	if !exists {
		t.Error("config.json annotation not found")
	} else {
		expectedNotes := "Configuration file - Contains database settings and API keys."
		if config.Notes != expectedNotes {
			t.Errorf("Config notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, config.Notes)
		}
	}

	// Test single.txt annotation (single line format)
	single, exists := annotations["single.txt"]
	if !exists {
		t.Error("single.txt annotation not found")
	} else {
		expectedNotes := "Just a title with no description"
		if single.Notes != expectedNotes {
			t.Errorf("Single notes mismatch.\nExpected: %q\nGot: %q", expectedNotes, single.Notes)
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

	// Test empty input
	reader := strings.NewReader("")

	err = GenerateInfoFromReader(reader)
	if err != nil {
		t.Fatalf("GenerateInfoFromReader should not fail on empty input: %v", err)
	}

	// Verify no .info files were created (since no valid entries)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".info") {
			t.Errorf("unexpected .info file created: %s", entry.Name())
		}
	}
}
