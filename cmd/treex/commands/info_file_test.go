package commands

import (
	"os"
	"strings"
	"testing"
)

func TestInfoFileFlagCommands(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create a custom info file
	customInfoContent := `src Source code
docs Documentation`
	err = os.WriteFile(".project-info", []byte(customInfoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create some directories
	if err := os.Mkdir("src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("docs", 0755); err != nil {
		t.Fatal(err)
	}

	// Test 1: Using del command with custom info file
	infoFile = ".info" // Reset before command
	_, err = executeDelCommand("--info-file", ".project-info", "src")
	if err != nil {
		t.Errorf("unexpected error with del command: %v", err)
	}

	// Verify src was deleted from custom file
	content, _ := os.ReadFile(".project-info")
	if strings.Contains(string(content), "src Source code") {
		t.Error("src annotation should have been deleted")
	}

	// Test 2: Using search command with custom info file
	// Re-create the file
	if err := os.WriteFile(".project-info", []byte(customInfoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset info file before search
	infoFile = ".info"
	output, err := executeSearchCommand("--info-file", ".project-info", "source")
	if err != nil {
		t.Errorf("unexpected error with search command: %v", err)
	}

	if !strings.Contains(output, "Found 1 matches for 'source'") {
		t.Errorf("expected to find match for 'source', got: %s", output)
	}

	// Test 3: Using sync command with custom info file
	// Remove src directory to make annotation stale
	if err := os.RemoveAll("src"); err != nil {
		t.Fatal(err)
	}

	// Reset info file before sync
	infoFile = ".info"
	output, err = executeSyncCommand("--info-file", ".project-info", "--force")
	if err != nil {
		t.Errorf("unexpected error with sync command: %v", err)
	}

	if !strings.Contains(output, "Found 1 stale annotations") {
		t.Errorf("expected to find stale annotation, got: %s", output)
	}
}

// Also test that default .info file still works
func TestDefaultInfoFile(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory
	oldDir, _ := os.Getwd()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	// Create default .info file
	defaultInfoContent := `src Main source
tests Testing code`
	err = os.WriteFile(".info", []byte(defaultInfoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create directories
	if err := os.Mkdir("src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("tests", 0755); err != nil {
		t.Fatal(err)
	}

	// Test search with default info file
	output, err := executeSearchCommand("main")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Found 1 matches for 'main'") {
		t.Errorf("expected to find match for 'main' in default .info file, got: %s", output)
	}
}
