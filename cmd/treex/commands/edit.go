package commands

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/spf13/cobra"
)

//go:embed edit.help.txt
var editHelp string

var (
	waitForEditor bool
	editAllFiles  bool
)

var editCmd = &cobra.Command{
	Use:   "info-edit [path]",
	Short: "Open .info file(s) in editor at specific annotation line",
	Long:  editHelp,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEditCmd,
}

func init() {
	editCmd.Flags().BoolVar(&waitForEditor, "wait", false, "Wait for editor to close (useful for GUI editors)")
	editCmd.Flags().BoolVar(&editAllFiles, "all", false, "Edit all .info files when path is found in multiple files")
	editCmd.Flags().StringVar(&infoFile, "info-file", ".info", "Use specified info file name instead of .info")
	rootCmd.AddCommand(editCmd)
}

// AnnotationLocation represents the location of an annotation in a file
type AnnotationLocation struct {
	File string
	Line int
}

// runEditCmd handles the CLI interface for the edit command
func runEditCmd(cmd *cobra.Command, args []string) error {
	// If no path provided, just open the .info file
	if len(args) == 0 {
		return openEditor(infoFile, 0)
	}

	targetPath := args[0]

	// Find all .info files in current directory tree
	infoFiles, err := findInfoFiles(".", infoFile)
	if err != nil {
		return fmt.Errorf("failed to find .info files: %w", err)
	}

	if len(infoFiles) == 0 {
		return fmt.Errorf("no %s files found", infoFile)
	}

	// Find the annotation in .info files
	locations, err := findAnnotationLocations(infoFiles, targetPath)
	if err != nil {
		return fmt.Errorf("failed to search for annotation: %w", err)
	}

	if len(locations) == 0 {
		return fmt.Errorf("no annotation found for path: %s", targetPath)
	}

	// Handle multiple matches
	if len(locations) > 1 && !editAllFiles {
		cmd.Printf("Found annotation in multiple files:\n")
		for i, loc := range locations {
			cmd.Printf("  %d. %s:%d\n", i+1, loc.File, loc.Line)
		}
		cmd.Printf("\nUse --all to edit all files, or specify --info-file to choose one.\n")
		return fmt.Errorf("annotation found in %d files", len(locations))
	}

	// Open editor(s)
	for _, loc := range locations {
		if err := openEditor(loc.File, loc.Line); err != nil {
			return fmt.Errorf("failed to open editor for %s: %w", loc.File, err)
		}
		
		if !editAllFiles {
			break // Only open the first one if not --all
		}
	}

	return nil
}

// findAnnotationLocations searches for an annotation path in multiple .info files
func findAnnotationLocations(infoFiles []string, targetPath string) ([]AnnotationLocation, error) {
	var locations []AnnotationLocation
	parser := info.NewParser()

	for _, infoFile := range infoFiles {
		annotations, _, err := parser.ParseFileWithWarnings(infoFile)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Check if the path exists in this file
		if _, exists := annotations[targetPath]; exists {
			// Find the line number by re-reading the file
			lineNum, err := findLineNumber(infoFile, targetPath)
			if err != nil {
				continue // Skip if we can't find the line
			}
			
			locations = append(locations, AnnotationLocation{
				File: infoFile,
				Line: lineNum,
			})
		}
	}

	return locations, nil
}

// findLineNumber finds the line number of a path in an .info file
func findLineNumber(infoFile, targetPath string) (int, error) {
	content, err := os.ReadFile(infoFile)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line to extract the path
		var path string
		if strings.Contains(line, ":") {
			// Format: "path: annotation"
			parts := strings.SplitN(line, ":", 2)
			path = strings.TrimSpace(parts[0])
		} else {
			// Format: "path annotation"
			parts := strings.Fields(line)
			if len(parts) > 0 {
				path = parts[0]
			}
		}

		if path == targetPath {
			return i + 1, nil // Line numbers are 1-based
		}
	}

	return 0, fmt.Errorf("path not found in file")
}

// openEditor opens the specified file in the user's editor
func openEditor(filename string, lineNumber int) error {
	editor := getEditor()
	if editor == "" {
		return fmt.Errorf("no editor found. Set EDITOR or VISUAL environment variable")
	}

	cmd, err := buildEditorCommand(editor, filename, lineNumber)
	if err != nil {
		return fmt.Errorf("failed to build editor command: %w", err)
	}

	// Set up the command
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if waitForEditor {
		return cmd.Run()
	} else {
		return cmd.Start()
	}
}

// getEditor returns the user's preferred editor
func getEditor() string {
	// Check EDITOR first, then VISUAL
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}
	
	// Try some common editors
	commonEditors := []string{"vim", "nano", "emacs", "code", "subl"}
	for _, editor := range commonEditors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}
	
	return ""
}

// buildEditorCommand builds the appropriate command for different editors
func buildEditorCommand(editor, filename string, lineNumber int) (*exec.Cmd, error) {
	// Get the base editor name (remove path and arguments)
	editorName := filepath.Base(strings.Fields(editor)[0])
	
	// Handle different editor syntaxes for line numbers
	switch editorName {
	case "vim", "vi", "nvim":
		if lineNumber > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNumber), filename), nil
		}
		return exec.Command(editor, filename), nil
		
	case "nano":
		if lineNumber > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNumber), filename), nil
		}
		return exec.Command(editor, filename), nil
		
	case "emacs":
		if lineNumber > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNumber), filename), nil
		}
		return exec.Command(editor, filename), nil
		
	case "code": // VS Code
		if lineNumber > 0 {
			return exec.Command(editor, "--goto", fmt.Sprintf("%s:%d", filename, lineNumber)), nil
		}
		return exec.Command(editor, filename), nil
		
	case "subl": // Sublime Text
		if lineNumber > 0 {
			return exec.Command(editor, fmt.Sprintf("%s:%d", filename, lineNumber)), nil
		}
		return exec.Command(editor, filename), nil
		
	default:
		// For unknown editors, try the vim-style syntax first
		if lineNumber > 0 {
			return exec.Command(editor, fmt.Sprintf("+%d", lineNumber), filename), nil
		}
		return exec.Command(editor, filename), nil
	}
}