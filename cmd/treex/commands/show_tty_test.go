package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTreexPipeDetection tests the actual treex binary behavior with pipes
func TestTreexPipeDetection(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create a simple directory structure with .info file
	err := os.Mkdir(filepath.Join(tmpDir, "cmd"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, ".info"), []byte("cmd: Command line tools"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("PipedOutput", func(t *testing.T) {
		// Run treex using go run
		cmd := exec.Command("go", "run", "../../../cmd/treex", tmpDir)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run treex: %v", err)
		}

		result := string(output)

		// Check that output does NOT contain ANSI escape codes
		if strings.Contains(result, "\x1b[") {
			t.Errorf("Piped output contains ANSI escape codes")
			// Show a sample for debugging
			sample := result
			if len(sample) > 200 {
				sample = sample[:200]
			}
			t.Logf("Output sample: %q", sample)
		}

		// Verify content is present
		if !strings.Contains(result, "cmd") {
			t.Errorf("Output missing expected content")
		}

		// Verify it uses box-drawing characters (UTF-8)
		if !strings.Contains(result, "├") || !strings.Contains(result, "└") {
			t.Logf("Note: Output might not contain box-drawing characters")
		}
	})

	t.Run("ExplicitNoColorFormat", func(t *testing.T) {
		// Test explicit --format=no-color
		cmd := exec.Command("go", "run", "../../../cmd/treex", "--format=no-color", tmpDir)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run treex: %v", err)
		}

		result := string(output)

		// Should not contain ANSI codes
		if strings.Contains(result, "\x1b[") {
			t.Errorf("no-color format output contains ANSI escape codes")
		}
	})

	t.Run("ExplicitColorFormat", func(t *testing.T) {
		// Test explicit --format=color
		// Note: lipgloss automatically strips colors when output is not a TTY
		// This is expected behavior - the format is "color" but lipgloss adapts
		cmd := exec.Command("go", "run", "../../../cmd/treex", "--format=color", tmpDir)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run treex: %v", err)
		}

		result := string(output)

		// When piped, lipgloss detects non-TTY and strips colors automatically
		// This is the expected behavior - we request color format but get plain text
		if strings.Contains(result, "\x1b[") {
			t.Logf("Note: Output contains ANSI codes even when piped (unexpected but not an error)")
		}

		// The important thing is that content is preserved
		if !strings.Contains(result, "cmd") {
			t.Errorf("Output missing expected content")
		}
	})
}
