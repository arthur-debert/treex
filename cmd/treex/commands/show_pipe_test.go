package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"golang.org/x/term"
)

// TestShowCommandPipeDetection tests that treex outputs plain text when piped
func TestShowCommandPipeDetection(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	
	// Create a simple directory structure with .info file
	err := os.Mkdir(tmpDir+"/cmd", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(tmpDir+"/.info", []byte("cmd: Command line tools"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test 1: When output is to a buffer (non-TTY), should use plain text
	t.Run("BufferOutput", func(t *testing.T) {
		var output bytes.Buffer
		cmd := GetRootCommand()
		cmd.SetOut(&output)
		cmd.SetErr(&output)
		cmd.SetArgs([]string{tmpDir})
		
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		result := output.String()
		
		// When output is to buffer, it should NOT contain ANSI codes
		if strings.Contains(result, "\x1b[") {
			t.Errorf("Output contains ANSI escape codes when writing to buffer")
			preview := result
			if len(preview) > 200 {
				preview = preview[:200]
			}
			t.Logf("Output preview: %q", preview)
		}
		
		// Verify content is still there
		if !strings.Contains(result, "cmd") || !strings.Contains(result, "Command line tools") {
			t.Errorf("Output missing expected content")
		}
	})

	// Test 2: Default behavior (would be TTY in real terminal)
	// For now, we'll just verify current behavior
	t.Run("DefaultBehavior", func(t *testing.T) {
		// Reset command for fresh state
		cmd := GetRootCommand()
		cmd.SetArgs([]string{tmpDir})
		
		// Capture via pipe to simulate what happens in shell piping
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		// Run command
		err := cmd.Execute()
		
		// Restore stdout and close pipe
		_ = w.Close()
		os.Stdout = oldStdout
		
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
		
		// Read output
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		result := buf.String()
		
		// Currently, this DOES contain ANSI codes (the bug we're fixing)
		if strings.Contains(result, "\x1b[") {
			t.Logf("Current behavior: Output contains ANSI codes when piped (this is the bug)")
		}
	})
}

// TestIsTTYDetection tests the TTY detection directly
func TestIsTTYDetection(t *testing.T) {
	// Test detection on stdout
	// When running tests, stdout is usually not a TTY
	if term.IsTerminal(int(os.Stdout.Fd())) {
		t.Log("stdout is a terminal (unexpected in test environment)")
	} else {
		t.Log("stdout is not a terminal (expected in test environment)")
	}

	// Test with a pipe (should not be a TTY)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()

	if term.IsTerminal(int(w.Fd())) {
		t.Error("Pipe write end detected as terminal, but should not be")
	}
}