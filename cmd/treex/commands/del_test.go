package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// setupDelCmd creates a properly initialized test info del command
func setupDelCmd() *cobra.Command {
	// Reset infoFile to default
	infoFile = ".info"

	// Create a test root command
	testRootCmd := &cobra.Command{
		Use:   "treex",
		Short: "Test root command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Create and add info command with del subcommand
	testInfoCmd := &cobra.Command{Use: "info", Short: "Info commands"}
	testInfoCmd.AddCommand(delCmd)
	testRootCmd.AddCommand(testInfoCmd)

	return testRootCmd
}

// executeDelCommand is a helper function to execute the info del command
func executeDelCommand(args ...string) (output string, err error) {
	root := setupDelCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(append([]string{"info", "del"}, args...))

	_, err = root.ExecuteC()
	return buf.String(), err
}

func TestDelCommand(t *testing.T) {
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

	// Create a test .info file
	infoContent := `src Main source directory
tests Test files
docs: Documentation files
build/output: Build output directory`

	err = os.WriteFile(".info", []byte(infoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkResult func(t *testing.T)
	}{
		{
			name:        "delete existing annotation",
			args:        []string{"src"},
			expectError: false,
			checkResult: func(t *testing.T) {
				content, err := os.ReadFile(".info")
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(content), "src Main source directory") {
					t.Error("annotation for 'src' should have been deleted")
				}
				if !strings.Contains(string(content), "tests Test files") {
					t.Error("other annotations should remain")
				}
			},
		},
		{
			name:        "delete path with spaces",
			args:        []string{"build/output"},
			expectError: false,
			checkResult: func(t *testing.T) {
				content, err := os.ReadFile(".info")
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(content), "build/output") {
					t.Error("annotation for 'build/output' should have been deleted")
				}
			},
		},
		{
			name:        "delete non-existent path",
			args:        []string{"nonexistent"},
			expectError: true,
			checkResult: nil,
		},
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset .info file for each test
			if err := os.WriteFile(".info", []byte(infoContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Execute command
			_, err := executeDelCommand(tt.args...)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t)
			}
		})
	}
}

func TestDelCommandNoInfoFile(t *testing.T) {
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

	// Execute command - should fail
	_, err = executeDelCommand("src")
	if err == nil {
		t.Error("expected error when no .info file exists")
	}
	if !strings.Contains(err.Error(), "no .info file found") {
		t.Errorf("unexpected error message: %v", err)
	}
}
