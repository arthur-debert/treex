package maketree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseTreeLine has been removed as tree format is no longer supported

// TestParseTreeText has been removed as tree format is no longer supported

func TestMakeTreeFromText_DryRun(t *testing.T) {
	tempDir := t.TempDir()

	content := `my-app/: Application root
cmd/: Command line utilities
pkg/: Core application code
README.md: Main documentation`

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
		filepath.Join(tempDir, "my-app") + " (dry run)",
		filepath.Join(tempDir, "cmd") + " (dry run)",
		filepath.Join(tempDir, "pkg") + " (dry run)",
	}

	if len(result.CreatedDirs) != len(expectedDirs) {
		t.Errorf("expected %d directories, got %d", len(expectedDirs), len(result.CreatedDirs))
	}

	// Check that files are reported as would-be-created
	expectedFiles := []string{
		filepath.Join(tempDir, "README.md") + " (dry run)",
	}

	if len(result.CreatedFiles) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d", len(expectedFiles), len(result.CreatedFiles))
	}

	// Check that .info file would be created
	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}

	// Verify nothing was actually created
	cmdPath := filepath.Join(tempDir, "cmd")
	if _, err := os.Stat(cmdPath); !os.IsNotExist(err) {
		t.Error("expected cmd directory to not exist in dry run mode")
	}
}

func TestMakeTreeFromText_ActualCreation(t *testing.T) {
	tempDir := t.TempDir()

	content := `cmd/: Command line utilities
pkg/: Core application code
README.md: Main documentation`

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
		filepath.Join(tempDir, "cmd"),
		filepath.Join(tempDir, "pkg"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Verify files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "README.md"),
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
				"cmd/:",
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
	existingDir := filepath.Join(tempDir, "cmd")
	err := os.MkdirAll(existingDir, 0755)
	if err != nil {
		t.Fatalf("failed to create existing directory: %v", err)
	}

	existingFile := filepath.Join(tempDir, "README.md")
	file, err := os.Create(existingFile)
	if err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}
	if err = file.Close(); err != nil {
		t.Fatalf("failed to close existing file: %v", err)
	}

	content := `cmd/: Command line utilities
pkg/: Core application code
README.md: Main documentation`

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
	existingFile := filepath.Join(tempDir, "README.md")
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

	content := `README.md: Main documentation`

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
	inputFile := filepath.Join(tempDir, "input.info")
	content := `src/: Source code
docs/: Documentation
Makefile: Build configuration`

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
		filepath.Join(targetDir, "src"),
		filepath.Join(targetDir, "docs"),
		filepath.Join(targetDir, "Makefile"),
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

func TestParseInfoContent(t *testing.T) {
	// Test parsing .info format content
	infoContent := `cmd/: Command line utilities
pkg/: Core application code
README.md: Main project documentation`

	entries, err := parseInfoContent(infoContent, "test-root")
	if err != nil {
		t.Fatalf("parseInfoContent failed: %v", err)
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

	if len(entries) != len(expectedEntries) {
		t.Errorf("expected %d entries, got %d", len(expectedEntries), len(entries))
	}

	for _, entry := range entries {
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

	content := `main.go: Entry point`

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
	mainGoPath := filepath.Join(tempDir, "main.go")
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

	content := `src/: Source code
docs/: Documentation
README.md: Main documentation`

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
		filepath.Join(tempDir, "src"),
		filepath.Join(tempDir, "docs"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Verify files were created
	expectedFiles := []string{
		filepath.Join(tempDir, "README.md"),
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

	content := `config.json: Configuration file`

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
		filepath.Join(tempDir, "config.json") + " (dry run)",
	}

	if len(result.CreatedFiles) != len(expectedFiles) {
		t.Errorf("expected %d files, got %d", len(expectedFiles), len(result.CreatedFiles))
	}

	// Check that .info file would be created
	if !result.InfoFileCreated {
		t.Error("expected InfoFileCreated to be true")
	}

	// Verify nothing was actually created
	testPath := filepath.Join(tempDir, "config.json")
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("expected config.json to not exist in dry run mode")
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

	_, err := MakeTreeFromReader(reader, tempDir, options)
	if err == nil {
		t.Error("expected error for empty input")
	}
	if !strings.Contains(err.Error(), "no valid entries") {
		t.Errorf("expected 'no valid entries' error, got: %v", err)
	}
}

// Test that we only accept .info format and reject tree format
func TestOnlyInfoFormatSupported(t *testing.T) {
	tempDir := t.TempDir()

	// Tree format should be rejected
	treeContent := `myproject
├── cmd/ Command line utilities
├── docs/ All documentation
└── README.md Main documentation`

	reader := strings.NewReader(treeContent)
	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: false,
	}

	_, err := MakeTreeFromReader(reader, tempDir, options)
	if err == nil {
		t.Error("expected error for tree format input, but got none")
	}
	if !strings.Contains(err.Error(), "no valid entries") {
		t.Errorf("expected format error, got: %v", err)
	}
}

// Test directory detection by trailing slash
func TestDirectoryDetectionByTrailingSlash(t *testing.T) {
	tempDir := t.TempDir()

	// Both with and without trailing slash for directories
	infoContent := `src/: Source code directory
build: Build directory (no slash)
README.md: Documentation file`

	reader := strings.NewReader(infoContent)
	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: false,
	}

	_, err := MakeTreeFromReader(reader, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromReader failed: %v", err)
	}

	// src/ should be created as directory due to trailing slash
	srcPath := filepath.Join(tempDir, "src")
	info, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("src path should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("src/ should be a directory due to trailing slash")
	}

	// build should be created as a file (no trailing slash)
	buildPath := filepath.Join(tempDir, "build")
	info, err = os.Stat(buildPath)
	if err != nil {
		t.Fatalf("build path should exist: %v", err)
	}
	if info.IsDir() {
		t.Error("build should be a file (no trailing slash)")
	}

	// README.md should be a file
	readmePath := filepath.Join(tempDir, "README.md")
	info, err = os.Stat(readmePath)
	if err != nil {
		t.Fatalf("README.md should exist: %v", err)
	}
	if info.IsDir() {
		t.Error("README.md should be a file")
	}
}

// Test directory detection by path prefix
func TestDirectoryDetectionByPathPrefix(t *testing.T) {
	tempDir := t.TempDir()

	// config is implicitly a directory because config/app.conf exists
	infoContent := `config: Configuration directory
config/app.conf: Application settings
src: Source directory
src/main.go: Main entry point`

	reader := strings.NewReader(infoContent)
	options := MakeTreeOptions{
		Force:      false,
		DryRun:     false,
		CreateInfo: false,
	}

	_, err := MakeTreeFromReader(reader, tempDir, options)
	if err != nil {
		t.Fatalf("MakeTreeFromReader failed: %v", err)
	}

	// config should be created as directory (has child path)
	configPath := filepath.Join(tempDir, "config")
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("config path should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("config should be a directory (has child paths)")
	}

	// src should be created as directory (has child path)
	srcPath := filepath.Join(tempDir, "src")
	info, err = os.Stat(srcPath)
	if err != nil {
		t.Fatalf("src path should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("src should be a directory (has child paths)")
	}

	// Verify files exist
	appConfPath := filepath.Join(tempDir, "config", "app.conf")
	if _, err := os.Stat(appConfPath); os.IsNotExist(err) {
		t.Error("config/app.conf should exist")
	}

	mainGoPath := filepath.Join(tempDir, "src", "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		t.Error("src/main.go should exist")
	}
}
