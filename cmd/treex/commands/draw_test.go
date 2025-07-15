package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDrawCommand(t *testing.T) {
	tests := []struct {
		name           string
		infoContent    string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "simple family tree",
			infoContent: `Dad Chill, dad
Mom Listen to your mother
kids/Sam Little Sam
kids/Emma Little Emma`,
			expectedOutput: []string{
				"Dad",
				"Chill, dad",
				"Mom", 
				"Listen to your mother",
				"kids",
				"Sam",
				"Little Sam",
				"Emma",
				"Little Emma",
			},
			expectError: false,
		},
		{
			name: "nested directories",
			infoContent: `company/ The company
company/engineering/ Engineering department
company/engineering/backend/ Backend team
company/engineering/frontend/ Frontend team
company/sales/ Sales department`,
			expectedOutput: []string{
				"company",
				"The company",
				"engineering",
				"Engineering department",
				"backend",
				"Backend team",
				"frontend",
				"Frontend team",
				"sales",
				"Sales department",
			},
			expectError: false,
		},
		{
			name: "mixed files and directories",
			infoContent: `README.md Project documentation
src/ Source code
src/main.go Main file
src/utils/ Utilities
src/utils/helper.go Helper functions`,
			expectedOutput: []string{
				"README.md",
				"Project documentation",
				"src",
				"Source code",
				"main.go",
				"Main file",
				"utils",
				"Utilities",
				"helper.go",
				"Helper functions",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with info content
			tmpFile, err := os.CreateTemp("", "test-draw-*.txt")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.infoContent); err != nil {
				t.Fatal(err)
			}
			tmpFile.Close()

			// Execute draw command
			cmd := newTestDrawCommand()
			cmd.SetArgs([]string{"--info-file", tmpFile.Name(), "--format", "no-color"})
			
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			
			err = cmd.Execute()
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if !tt.expectError {
				output := buf.String()
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(output, expected) {
						t.Errorf("expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
					}
				}
			}
		})
	}
}

func TestDrawCommandFromStdin(t *testing.T) {
	// Save original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r

	// Write test data to pipe
	testData := `project/ My project
project/docs/ Documentation
project/src/ Source code`
	
	go func() {
		defer w.Close()
		w.WriteString(testData)
	}()

	// Execute draw command
	cmd := newTestDrawCommand()
	cmd.SetArgs([]string{"--format", "no-color"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	expectedStrings := []string{"project", "My project", "docs", "Documentation", "src", "Source code"}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
		}
	}
}

func TestDrawCommandNoInput(t *testing.T) {
	cmd := newTestDrawCommand()
	cmd.SetArgs([]string{})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no input provided")
	}
	
	if !strings.Contains(err.Error(), "no input provided") {
		t.Errorf("expected 'no input provided' error, got: %v", err)
	}
}

func TestDrawCommandInvalidFormat(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test-draw-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.WriteString("test/ Test")
	tmpFile.Close()

	cmd := newTestDrawCommand()
	cmd.SetArgs([]string{"--info-file", tmpFile.Name(), "--format", "invalid-format"})
	
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

// Helper to create test draw command
func newTestDrawCommand() *cobra.Command {
	// Reset global variables
	infoFile = ""
	outputFormat = "color"
	
	// Create root command with draw subcommand
	root := &cobra.Command{Use: "treex"}
	
	draw := &cobra.Command{
		Use:   "draw",
		Short: "Draw tree diagrams",
		Args:  cobra.NoArgs,
		RunE:  runDrawCmd,
	}
	
	draw.Flags().StringVar(&infoFile, "info-file", "", "Info file")
	draw.Flags().StringVarP(&outputFormat, "format", "f", "color", "Output format")
	
	root.AddCommand(draw)
	
	// Need to set the subcommand
	root.SetArgs(append([]string{"draw"}, root.Flags().Args()...))
	
	return root
}