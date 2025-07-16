package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveAbsolutePath resolves a path to an absolute path with robust error handling
// It handles cases where os.Getwd() might fail by using fallback methods
func ResolveAbsolutePath(path string) (string, error) {
	// If the path is already absolute, return it
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// Try to get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to PWD environment variable
		if pwd := os.Getenv("PWD"); pwd != "" {
			cwd = pwd
		} else {
			// Last resort: try to use the path as-is if it exists
			if _, statErr := os.Stat(path); statErr == nil {
				// The path exists, so we can try to clean it
				return filepath.Clean(path), nil
			}
			return "", fmt.Errorf("cannot determine working directory: %w", err)
		}
	}

	// Join with current working directory and clean
	return filepath.Clean(filepath.Join(cwd, path)), nil
}