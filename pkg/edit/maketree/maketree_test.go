package maketree

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
		expectName  string
		expectDesc  string
		expectIsDir bool
		expectDepth int
		expectError bool
	}{
		{
			name:        "simple file with description",
			line:        "├── main.go Application entry point",
			expectName:  "main.go",
			expectDesc:  "Application entry point",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "directory with trailing slash",
			line:        "├── cmd/ Command line interface",
			expectName:  "cmd",
			expectDesc:  "Command line interface",
			expectIsDir: true,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "nested directory",
			line:        "│   └── internal/ Internal packages",
			expectName:  "internal",
			expectDesc:  "Internal packages",
			expectIsDir: true,
			expectDepth: 1,
			expectError: false,
		},
		{
			name:        "file without description",
			line:        "└── README.md",
			expectName:  "README.md",
			expectDesc:  "",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "empty line should be skipped",
			line:        "",
			expectName:  "",
			expectDesc:  "",
			expectIsDir: false,
			expectDepth: 0,
			expectError: false,
		},
		{
			name:        "simple root entry without connectors",
			line:        "myproject Main application directory",
			expectName:  "myproject",
			expectDesc:  "Main application directory",
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

			if tt.expectName == "" && entry != nil {
				t.Errorf("expected nil entry for empty name, got %+v", entry)
				return
			}

			if tt.expectName != "" && entry == nil {
				t.Errorf("expected entry but got nil")
				return
			}

			if entry != nil {
				if entry.Name != tt.expectName {
					t.Errorf("expected name %q, got %q", tt.expectName, entry.Name)
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

func TestParseTreeText(t *testing.T) {
	content := `myproject
├── cmd/ Command line utilities
├── docs/ All documentation
│   └── guides/ User guides and tutorials
├── pkg/ Core application code
├── scripts/ Build and deployment scripts
└── README.md Main project documentation`

	treeStructure, err := parseTreeText(content, "test-root")
	if err != nil {
		t.Fatalf("parseTreeText failed: %v", err)
	}

	expectedEntries := []struct {
		relativePath string
		description  string
		isDir        bool
	}{
		{"myproject", "", true}, // Root is now correctly treated as a directory
		{"myproject/cmd", "Command line utilities", true},
		{"myproject/docs", "All documentation", true},
		{"myproject/docs/guides", "User guides and tutorials", true},
		{"myproject/pkg", "Core application code", true},
		{"myproject/scripts", "Build and deployment scripts", true},
		{"myproject/README.md", "Main project documentation", false},
	}

	if len(treeStructure.Entries) != len(expectedEntries) {
		t.Errorf("expected %d entries, got %d", len(expectedEntries), len(treeStructure.Entries))
		for i, entry := range treeStructure.Entries {
			t.Logf("Entry %d: %s -> %s (isDir: %v)", i, entry.RelativePath, entry.Description, entry.IsDir)
		}
		return
	}

	for i, expected := range expectedEntries {
		if i >= len(treeStructure.Entries) {
			t.Errorf("missing entry %d: %s", i, expected.relativePath)
			continue
		}

		entry := treeStructure.Entries[i]
		if entry.RelativePath != expected.relativePath {
			t.Errorf("entry %d: expected relative path %q, got %q", i, expected.relativePath, entry.RelativePath)
		}
		if entry.Description != expected.description {
			t.Errorf("entry %d: expected description %q, got %q", i, expected.description, entry.Description)
		}
		if entry.IsDir != expected.isDir {
			t.Errorf("entry %d: expected IsDir %v, got %v", i, expected.isDir, entry.IsDir)
		}
	}

	// Check that paths are properly constructed
	if treeStructure.RootPath != "test-root" {
		t.Errorf("expected root path %q, got %q", "test-root", treeStructure.RootPath)
	}

	if treeStructure.Source != SourceTreeText {
		t.Errorf("expected source %v, got %v", SourceTreeText, treeStructure.Source)
	}
}

func TestMakeTreeFromText_DryRun(t *testing.T) {
	tempDir := t.TempDir()

	content := `my-app
├── cmd/ Command line utilities
├── pkg/ Core application code
└── README.md Main documentation`

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     true,
		CreateInfo: true,
	}

	result, err := MakeTreeFromText(content, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromText failed: %v", err)
	}

	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}

	// Check that directories are reported as would-be-created
	expectedDirs := []string{
		filepath.Join(tempDir, "my-app") + " (dry run)", // Root is now a directory
		filepath.Join(tempDir, "my-app", "cmd") + " (dry run)",
		filepath.Join(tempDir, "my-app", "pkg") + " (dry run)",
	}

	if len(result.CreatedDirs) != len(expectedDirs) {
		t.Errorf("expected %d directories, got %d", len(expectedDirs), len(result.CreatedDirs))
	}

	// Check that files are reported as would-be-created
	expectedFiles := []string{
		filepath.Join(tempDir, "my-app", "README.md") + " (dry run)",
	}

	if len(result.CreatedFiles) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d", len(expectedFiles), len(result.CreatedFiles))
	}

	// Check that .info file would be created
	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}

	// Verify nothing was actually created
	myAppPath := filepath.Join(tempDir, "my-app")
	if _, err := os.Stat(myAppPath); !os.IsNotExist(err) {
		t.Error("expected my-app directory to not exist in dry run mode")
	}
}

func TestMakeTreeFromText_ActualCreation(t *testing.T) {
	tempDir := t.TempDir()

	content := `my-app
├── cmd/ Command line utilities
├── pkg/ Core application code
└── README.md Main documentation`

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: true,
	}

	result, err := MakeTreeFromText(content, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromText failed: %v", err)
	}

	if result.DryRun {
		t.Error("expected DryRun to be false")
	}

	// Verify directories were created
	expectedDirs := []string{
		filepath.Join(tempDir, "my-app", "cmd"),
		filepath.Join(tempDir, "my-app", "pkg"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Verify files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "my-app"),
		filepath.Join(tempDir, "my-app", "README.md"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}

	// Verify .info file was created
	infoPath := filepath.Join(tempDir, ".info")
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Error("expected .info file to be created")
	} else {
		// Check .info file content
		content, err := os.ReadFile(infoPath)
		if err != nil {
			t.Errorf("failed to read .info file: %v", err)
		} else {
			contentStr := string(content)
			expectedStrings := []string{
				"my-app",
				"created by treex make-tree",
			}
			for _, expected := range expectedStrings {
				if !strings.Contains(contentStr, expected) {
					t.Errorf("expected .info file to contain %q, got:\n%s", expected, contentStr)
				}
			}
		}
	}

	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}
}

func TestMakeTreeFromText_WithExistingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create some existing files/directories
	existingDir := filepath.Join(tempDir, "my-app", "cmd")
	err := os.MkdirAll(existingDir, 0755)
	if err != nil {
		t.Fatalf("failed to create existing directory: %v", err)
	}

	existingFile := filepath.Join(tempDir, "my-app", "README.md")
	err = os.MkdirAll(filepath.Dir(existingFile), 0755)
	if err != nil {
		t.Fatalf("failed to create parent directory: %v", err)
	}
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("failed to close existing file: %v", err)
	}

	content := `my-app
├── cmd/ Command line utilities
├── pkg/ Core application code
└── README.md Main documentation`

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: true,
	}

	result, err := MakeTreeFromText(content, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromText failed: %v", err)
	}

	// Check that existing paths were skipped
	expectedSkipped := []string{
		existingDir + " (already exists)",
		existingFile + " (already exists)",
	}

	if len(result.SkippedPaths) < 2 {
		t.Errorf("expected at least 2 skipped paths, got %d: %v", len(result.SkippedPaths), result.SkippedPaths)
	}

	for _, expected := range expectedSkipped {
		found := false
		for _, skipped := range result.SkippedPaths {
			if skipped == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find skipped path %q in %v", expected, result.SkippedPaths)
		}
	}
}

func TestMakeTreeFromText_WithForce(t *testing.T) {
	tempDir := t.TempDir()

	// Create existing file
	existingFile := filepath.Join(tempDir, "my-app", "README.md")
	err := os.MkdirAll(filepath.Dir(existingFile), 0755)
	if err != nil {
		t.Fatalf("failed to create parent directory: %v", err)
	}
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}
	_, err = file.WriteString("existing content")
	if err != nil {
		t.Fatalf("failed to write to existing file: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("failed to close existing file: %v", err)
	}

	content := `my-app
└── README.md Main documentation`

	options := MakeTreeOptions{
		Force:      true, // Force overwrite
		DryRun:     false,
		CreateInfo: true,
	}

	result, err := MakeTreeFromText(content, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromText failed: %v", err)
	}

	// With force=true, existing files should be overwritten, not skipped
	if len(result.SkippedPaths) > 0 {
		t.Errorf("expected no skipped paths with force=true, got: %v", result.SkippedPaths)
	}

	// Verify file was recreated (should be empty now)
	fileContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Errorf("failed to read recreated file: %v", err)
	} else if len(fileContent) != 0 {
		t.Errorf("expected recreated file to be empty, got content: %q", string(fileContent))
	}
}

func TestMakeTreeFromFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test input file
	inputFile := filepath.Join(tempDir, "input.txt")
	content := `my-project
├── src/ Source code
├── docs/ Documentation
└── Makefile Build configuration`

	err := os.WriteFile(inputFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	targetDir := filepath.Join(tempDir, "target")
	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: true,
	}

	result, err := MakeTreeFromFile(inputFile, targetDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromFile failed: %v", err)
	}

	// Verify structure was created
	expectedPaths := []string{
		filepath.Join(targetDir, "my-project"),
		filepath.Join(targetDir, "my-project", "src"),
		filepath.Join(targetDir, "my-project", "docs"),
		filepath.Join(targetDir, "my-project", "Makefile"),
		filepath.Join(targetDir, ".info"),
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected path %s to exist", path)
		}
	}

	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}
}

func TestParseInfoFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a .info file with compact format
	infoContent := `cmd/ Command line utilities

pkg/ Core application code

README.md Main project documentation`

	infoPath := filepath.Join(tempDir, ".info")
	err := os.WriteFile(infoPath, []byte(infoContent), 0644)
	if err != nil {
		t.Fatalf("failed to create .info file: %v", err)
	}

	treeStructure, err := parseInfoFile(infoPath, "test-root")
	if err != nil {
		t.Fatalf("parseInfoFile failed: %v", err)
	}

	if treeStructure.Source != SourceInfoFile {
		t.Errorf("expected source %v, got %v", SourceInfoFile, treeStructure.Source)
	}

	if treeStructure.RootPath != "test-root" {
		t.Errorf("expected root path %q, got %q", "test-root", treeStructure.RootPath)
	}

	// Check that entries were parsed correctly
	expectedEntries := map[string]struct {
		description string
		isDir       bool
	}{
		"cmd":       {"Command line utilities", true},
		"pkg":       {"Core application code", true},
		"README.md": {"Main project documentation", false},
	}

	if len(treeStructure.Entries) != len(expectedEntries) {
		t.Errorf("expected %d entries, got %d", len(expectedEntries), len(treeStructure.Entries))
	}

	for _, entry := range treeStructure.Entries {
		expected, exists := expectedEntries[entry.RelativePath]
		if !exists {
			t.Errorf("unexpected entry: %s", entry.RelativePath)
			continue
		}

		if entry.Description != expected.description {
			t.Errorf("entry %s: expected description %q, got %q", entry.RelativePath, expected.description, entry.Description)
		}

		if entry.IsDir != expected.isDir {
			t.Errorf("entry %s: expected IsDir %v, got %v", entry.RelativePath, expected.isDir, entry.IsDir)
		}
	}
}

func TestMakeTreeFromText_NoInfoFile(t *testing.T) {
	tempDir := t.TempDir()

	content := `simple-project
└── main.go Entry point`

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: false, // Don't create .info file
	}

	result, err := MakeTreeFromText(content, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromText failed: %v", err)
	}

	// Verify structure was created
	mainGoPath := filepath.Join(tempDir, "simple-project", "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("expected main.go to exist")
	}

	// Verify .info file was NOT created
	infoPath := filepath.Join(tempDir, ".info")
	if _, err := os.Stat(infoPath); !os.IsNotExist(err) {
		t.Error("expected .info file to NOT exist")
	}

	if result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be false")
	}
}

func TestMakeTreeFromReader(t *testing.T) {
	tempDir := t.TempDir()

	content := `reader-app
├── src/ Source code  
├── docs/ Documentation
└── README.md Main documentation`

	reader := strings.NewReader(content)

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: true,
	}

	result, err := MakeTreeFromReader(reader, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromReader failed: %v", err)
	}

	// Verify directories were created
	expectedDirs := []string{
		filepath.Join(tempDir, "reader-app"),
		filepath.Join(tempDir, "reader-app", "src"),
		filepath.Join(tempDir, "reader-app", "docs"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Verify files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "reader-app", "README.md"),
		filepath.Join(tempDir, ".info"),
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}

	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}
}

func TestMakeTreeFromReader_DryRun(t *testing.T) {
	tempDir := t.TempDir()

	content := `stdin-test
└── config.json Configuration file`

	reader := strings.NewReader(content)

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     true,
		CreateInfo: true,
	}

	result, err := MakeTreeFromReader(reader, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromReader failed: %v", err)
	}

	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}

	// Check that files are reported as would-be-created
	expectedFiles := []string{
		filepath.Join(tempDir, "stdin-test", "config.json") + " (dry run)",
	}

	if len(result.CreatedFiles) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d", len(expectedFiles), len(result.CreatedFiles))
	}

	// Check that .info file would be created
	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}

	// Verify nothing was actually created
	testPath := filepath.Join(tempDir, "stdin-test")
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("expected stdin-test directory to not exist in dry run mode")
	}
}

func TestMakeTreeFromReader_EmptyInput(t *testing.T) {
	tempDir := t.TempDir()

	reader := strings.NewReader("")

	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: true,
	}

	result, err := MakeTreeFromReader(reader, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromReader failed: %v", err)
	}

	// Empty input should result in no created files or directories
	if len(result.CreatedDirs) != 0 {
		t.Errorf("expected 0 directories, got %d", len(result.CreatedDirs))
	}

	if len(result.CreatedFiles) != 0 {
		t.Errorf("expected 0 files, got %d", len(result.CreatedFiles))
	}

	// .info file should not be created for empty input
	if result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be false for empty input")
	}
}
