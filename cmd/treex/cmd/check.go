package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adebert/treex/pkg/info"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Validate .info files in a directory",
	Long: `Validate .info files in the specified directory (or current directory if not specified).

This command will:
- Parse all .info files in the directory tree
- Check for syntax errors and formatting issues
- Verify that referenced paths exist
- Exit with code 0 if all .info files are valid (prints nothing)
- Exit with code 1 if any .info files have errors (prints error details)

Examples:
  treex check              # Check .info files in current directory
  treex check ./src        # Check .info files in src directory`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheckCmd,
}

func init() {
	// Register the command with root
	rootCmd.AddCommand(checkCmd)
}

// runCheckCmd handles the CLI interface for check command
func runCheckCmd(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}
	
	// Delegate to business logic
	err := validateInfoFiles(targetPath)
	if err != nil {
		// Print the error and exit with code 1
		fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
		os.Exit(1)
	}
	
	// Success - print nothing and exit with code 0
	return nil
}

// validateInfoFiles validates all .info files in the given directory tree
func validateInfoFiles(targetPath string) error {
	// Ensure the target path exists and is a directory
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", targetPath)
	}
	
	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", targetPath)
	}
	
	// Try to parse all .info files in the directory tree
	annotations, err := info.ParseDirectoryTree(absPath)
	if err != nil {
		return fmt.Errorf("failed to parse .info files: %w", err)
	}
	
	// Check that all referenced paths actually exist
	var validationErrors []string
	
	for annotationPath, annotation := range annotations {
		// Convert relative path to absolute path for checking
		fullPath := filepath.Join(absPath, annotationPath)
		
		// Normalize path separators
		fullPath = filepath.Clean(fullPath)
		
		// Check if the path exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			validationErrors = append(validationErrors, 
				fmt.Sprintf("annotation references non-existent path: %s (referenced in annotation: %s)", 
					annotationPath, annotation.Path))
		}
	}
	
	// If there are validation errors, return them
	if len(validationErrors) > 0 {
		errorMsg := "Found validation errors:\n"
		for i, err := range validationErrors {
			errorMsg += fmt.Sprintf("  %d. %s\n", i+1, err)
		}
		return fmt.Errorf("%s", errorMsg)
	}
	
	// All validations passed
	return nil
} 