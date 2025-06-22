package info

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateInfoFiles validates all .info files in the given directory tree
// This is the main business logic function that can be tested independently
func ValidateInfoFiles(targetPath string) error {
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
	annotations, err := ParseDirectoryTree(absPath)
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