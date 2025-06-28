package info

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockUserInteraction is a test implementation of UserInteraction
type MockUserInteraction struct {
	OverwriteResponse      bool
	OverwriteError         error
	SuccessCalled          bool
	SuccessPath            string
	SuccessDepth           int
	SuccessWithPathsCalled bool
	SuccessPathCount       int
}

func (m *MockUserInteraction) ConfirmOverwrite(targetPath string) (bool, error) {
	return m.OverwriteResponse, m.OverwriteError
}

func (m *MockUserInteraction) ShowSuccess(targetPath string, depth int) {
	m.SuccessCalled = true
	m.SuccessPath = targetPath
	m.SuccessDepth = depth
}

func (m *MockUserInteraction) ShowSuccessWithPaths(targetPath string, pathCount int) {
	m.SuccessWithPathsCalled = true
	m.SuccessPath = targetPath
	m.SuccessPathCount = pathCount
}

func TestInitializeInfoFile_BasicFunctionality(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create some test files and directories
	testFiles := []string{
		"README.md",
		"main.go",
		"src/app.go",
		"src/utils.go",
		"docs/guide.md",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Create the file
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Test options
	options := InitOptions{
		Depth: 2,
	}

	// Mock user interaction
	mockUI := &MockUserInteraction{
		OverwriteResponse: true,
	}

	// Run the initialization
	err := InitializeInfoFile(tempDir, options, mockUI)
	if err != nil {
		t.Fatalf("InitializeInfoFile failed: %v", err)
	}

	// Verify success was called
	if !mockUI.SuccessCalled {
		t.Error("Expected ShowSuccess to be called")
	}

	if mockUI.SuccessPath != tempDir {
		t.Errorf("Expected success path %s, got %s", tempDir, mockUI.SuccessPath)
	}

	if mockUI.SuccessDepth != 2 {
		t.Errorf("Expected success depth %d, got %d", 2, mockUI.SuccessDepth)
	}

	// Verify .info file was created
	infoPath := filepath.Join(tempDir, ".info")
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Fatal(".info file was not created")
	}

	// Read and verify the .info file content
	content, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to read .info file: %v", err)
	}

	contentStr := string(content)

	// Should contain direct children
	expectedEntries := []string{"docs/", "src/", "README.md", "main.go"}
	for _, entry := range expectedEntries {
		if !strings.Contains(contentStr, entry) {
			t.Errorf("Expected .info file to contain '%s', content:\n%s", entry, contentStr)
		}
	}

	// Should NOT contain nested files (depth limit)
	unexpectedEntries := []string{"app.go", "utils.go", "guide.md"}
	for _, entry := range unexpectedEntries {
		if strings.Contains(contentStr, entry) {
			t.Errorf("Expected .info file NOT to contain nested file '%s' due to depth limit, content:\n%s", entry, contentStr)
		}
	}
}

func TestInitializeInfoFile_OverwriteConfirmation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create existing .info file
	infoPath := filepath.Join(tempDir, ".info")
	existingContent := "existing content"
	err := os.WriteFile(infoPath, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing .info file: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test refusing to overwrite
	t.Run("RefuseOverwrite", func(t *testing.T) {
		options := InitOptions{Depth: 1}
		mockUI := &MockUserInteraction{
			OverwriteResponse: false, // User says no
		}

		err := InitializeInfoFile(tempDir, options, mockUI)
		if err != nil {
			t.Fatalf("InitializeInfoFile failed: %v", err)
		}

		// Verify file was not overwritten
		content, err := os.ReadFile(infoPath)
		if err != nil {
			t.Fatalf("Failed to read .info file: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Expected .info file to remain unchanged when user refuses overwrite")
		}

		// Success should not be called when user cancels
		if mockUI.SuccessCalled {
			t.Error("Expected ShowSuccess NOT to be called when user cancels")
		}
	})

	// Test accepting overwrite
	t.Run("AcceptOverwrite", func(t *testing.T) {
		options := InitOptions{Depth: 1}
		mockUI := &MockUserInteraction{
			OverwriteResponse: true, // User says yes
		}

		err := InitializeInfoFile(tempDir, options, mockUI)
		if err != nil {
			t.Fatalf("InitializeInfoFile failed: %v", err)
		}

		// Verify file was overwritten
		content, err := os.ReadFile(infoPath)
		if err != nil {
			t.Fatalf("Failed to read .info file: %v", err)
		}

		if string(content) == existingContent {
			t.Error("Expected .info file to be overwritten when user accepts")
		}

		// Should contain the test file
		if !strings.Contains(string(content), "test.txt") {
			t.Error("Expected new .info file to contain test.txt")
		}

		// Success should be called
		if !mockUI.SuccessCalled {
			t.Error("Expected ShowSuccess to be called when operation succeeds")
		}
	})
}

func TestInitializeInfoFile_InvalidPath(t *testing.T) {
	options := InitOptions{Depth: 3}
	mockUI := &MockUserInteraction{}

	// Test with non-existent path
	err := InitializeInfoFile("/nonexistent/path/12345", options, mockUI)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	if !strings.Contains(err.Error(), "path does not exist") {
		t.Errorf("Expected error message about non-existent path, got: %v", err)
	}
}

func TestInitializeInfoFile_FileNotDirectory(t *testing.T) {
	// Create a temporary file (not directory)
	tempFile := filepath.Join(t.TempDir(), "notadir.txt")
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	options := InitOptions{Depth: 3}
	mockUI := &MockUserInteraction{}

	// Test with file instead of directory
	err = InitializeInfoFile(tempFile, options, mockUI)
	if err == nil {
		t.Error("Expected error when path is not a directory")
	}

	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("Expected error message about not being a directory, got: %v", err)
	}
}

func TestInitializeInfoFile_DepthLimiting(t *testing.T) {
	// Create a deep directory structure
	tempDir := t.TempDir()

	// Create nested structure: level0/level1/level2/level3/
	deepPath := filepath.Join(tempDir, "level1", "level2", "level3")
	err := os.MkdirAll(deepPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Create files at different levels
	files := []string{
		"root.txt",                       // depth 0 (root)
		"level1/first.txt",               // depth 1
		"level1/level2/second.txt",       // depth 2
		"level1/level2/level3/third.txt", // depth 3
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Test with depth limit of 1
	options := InitOptions{Depth: 1}
	mockUI := &MockUserInteraction{}

	err = InitializeInfoFile(tempDir, options, mockUI)
	if err != nil {
		t.Fatalf("InitializeInfoFile failed: %v", err)
	}

	// Read the generated .info file
	infoPath := filepath.Join(tempDir, ".info")
	content, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to read .info file: %v", err)
	}

	contentStr := string(content)

	// Should contain direct children only
	if !strings.Contains(contentStr, "level1/") {
		t.Error("Expected .info file to contain level1/ directory")
	}

	if !strings.Contains(contentStr, "root.txt") {
		t.Error("Expected .info file to contain root.txt file")
	}

	// Should NOT contain deeper files (they should be limited by depth)
	deepFiles := []string{"first.txt", "second.txt", "third.txt"}
	for _, file := range deepFiles {
		if strings.Contains(contentStr, file) {
			t.Errorf("Expected .info file NOT to contain deep file '%s' due to depth limit, content:\n%s", file, contentStr)
		}
	}
}

func TestInitializeInfoFileWithPaths_BasicFunctionality(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files and directories
	testPaths := []string{
		"src/main.go",
		"docs/README.md",
		"config",
		"bin/app",
	}

	// Create the actual files and directories
	for _, path := range testPaths {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Create the file/directory
		if strings.HasSuffix(path, "/") || path == "config" {
			// It's a directory
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", fullPath, err)
			}
		} else {
			// It's a file
			if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", fullPath, err)
			}
		}
	}

	// Change to the temp directory for testing relative paths
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Mock user interaction
	mockUI := &MockUserInteraction{
		OverwriteResponse: true,
	}

	// Run the initialization with specific paths
	err = InitializeInfoFileWithPaths(".", testPaths, mockUI)
	if err != nil {
		t.Fatalf("InitializeInfoFileWithPaths failed: %v", err)
	}

	// Verify success was called
	if !mockUI.SuccessWithPathsCalled {
		t.Error("Expected ShowSuccessWithPaths to be called")
	}

	if mockUI.SuccessPath != "." {
		t.Errorf("Expected success path '.', got %s", mockUI.SuccessPath)
	}

	if mockUI.SuccessPathCount != len(testPaths) {
		t.Errorf("Expected success path count %d, got %d", len(testPaths), mockUI.SuccessPathCount)
	}

	// Verify .info file was created
	infoPath := ".info"
	if _, err := os.Stat(infoPath); os.IsNotExist(err) {
		t.Fatal(".info file was not created")
	}

	// Read and verify the .info file content
	content, err := os.ReadFile(infoPath)
	if err != nil {
		t.Fatalf("Failed to read .info file: %v", err)
	}

	contentStr := string(content)

	// Should contain all specified paths (with proper directory suffix)
	expectedEntries := []string{
		"bin/app",        // file
		"config/",        // directory with suffix added
		"docs/README.md", // file
		"src/main.go",    // file
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(contentStr, entry) {
			t.Errorf("Expected .info file to contain '%s', content:\n%s", entry, contentStr)
		}
	}

	// Verify sorting: directories first, then files, both alphabetically
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	nonEmptyLines := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, strings.TrimSpace(line))
		}
	}

	// Should be sorted with directories first
	if len(nonEmptyLines) < 4 {
		t.Fatalf("Expected at least 4 entries, got %d: %v", len(nonEmptyLines), nonEmptyLines)
	}

	// First entry should be config/ (directory)
	if nonEmptyLines[0] != "config/" {
		t.Errorf("Expected first entry to be 'config/', got '%s'", nonEmptyLines[0])
	}
}

func TestInitializeInfoFileWithPaths_NonexistentPath(t *testing.T) {
	tempDir := t.TempDir()

	// Test with a path that doesn't exist
	testPaths := []string{
		"src/main.go",
		"nonexistent/file.txt", // This doesn't exist
	}

	// Create only one of the files
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
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

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	mockUI := &MockUserInteraction{
		OverwriteResponse: true,
	}

	// This should fail because one path doesn't exist
	err = InitializeInfoFileWithPaths(".", testPaths, mockUI)
	if err == nil {
		t.Fatal("Expected InitializeInfoFileWithPaths to fail with nonexistent path")
	}

	if !strings.Contains(err.Error(), "nonexistent/file.txt") {
		t.Errorf("Expected error to mention nonexistent path, got: %v", err)
	}
}

func TestInitializeInfoFileWithPaths_OverwriteConfirmation(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create existing .info file
	infoPath := filepath.Join(tempDir, ".info")
	existingContent := "existing content"
	if err := os.WriteFile(infoPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to create existing .info file: %v", err)
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

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test refusing to overwrite
	t.Run("RefuseOverwrite", func(t *testing.T) {
		mockUI := &MockUserInteraction{
			OverwriteResponse: false, // User says no
		}

		err := InitializeInfoFileWithPaths(".", []string{"test.txt"}, mockUI)
		if err != nil {
			t.Fatalf("InitializeInfoFileWithPaths failed: %v", err)
		}

		// Verify file was not overwritten
		content, err := os.ReadFile(infoPath)
		if err != nil {
			t.Fatalf("Failed to read .info file: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Expected .info file to remain unchanged when user refuses overwrite")
		}

		// Success should not be called when user cancels
		if mockUI.SuccessWithPathsCalled {
			t.Error("Expected ShowSuccessWithPaths NOT to be called when user cancels")
		}
	})

	// Test accepting overwrite
	t.Run("AcceptOverwrite", func(t *testing.T) {
		mockUI := &MockUserInteraction{
			OverwriteResponse: true, // User says yes
		}

		err := InitializeInfoFileWithPaths(".", []string{"test.txt"}, mockUI)
		if err != nil {
			t.Fatalf("InitializeInfoFileWithPaths failed: %v", err)
		}

		// Verify file was overwritten
		content, err := os.ReadFile(infoPath)
		if err != nil {
			t.Fatalf("Failed to read .info file: %v", err)
		}

		if string(content) == existingContent {
			t.Error("Expected .info file to be overwritten when user accepts")
		}

		// Should contain the test file
		if !strings.Contains(string(content), "test.txt") {
			t.Error("Expected new .info file to contain test.txt")
		}

		// Success should be called
		if !mockUI.SuccessWithPathsCalled {
			t.Error("Expected ShowSuccessWithPaths to be called when operation succeeds")
		}
	})
}
