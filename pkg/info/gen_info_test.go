package info

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			entry, depth, err := parseTreeLine(tt.line, []string{})
			
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

	entries, err := parseTreeFile(testFile)
	if err != nil {
		t.Fatalf("parseTreeFile failed: %v", err)
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

	// Verify that .info files were created
	expectedInfoFiles := []struct {
		path     string
		contains []string
	}{
		{
			path: ".info",
			contains: []string{"myproject"},
		},
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
			}
		}
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

	err = generateInfoFile("", entries)
	if err != nil {
		t.Fatalf("generateInfoFile failed: %v", err)
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
		"main.go",
	}

	for _, expectedText := range expected {
		if !strings.Contains(contentStr, expectedText) {
			t.Errorf("Expected .info file to contain %q, but it doesn't. Content:\n%s", expectedText, contentStr)
		}
	}
} 