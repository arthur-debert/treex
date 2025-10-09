package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <json-file> <destination-dir>\n", os.Args[0])
		os.Exit(1)
	}

	jsonFile := os.Args[1]
	destDir := os.Args[2]

	// Read JSON file
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON structure
	var structure map[string]interface{}
	if err := json.Unmarshal(data, &structure); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Create filesystem structure using real filesystem
	fs := afero.NewOsFs()
	if err := createTreeRecursive(fs, destDir, structure); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating filesystem structure: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created filesystem structure in %s\n", destDir)
}

func createTreeRecursive(fs afero.Fs, basePath string, structure map[string]interface{}) error {
	for name, content := range structure {
		fullPath := filepath.Join(basePath, name)

		switch v := content.(type) {
		case string:
			// It's a file with string content
			dir := filepath.Dir(fullPath)
			if err := fs.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			if err := afero.WriteFile(fs, fullPath, []byte(v), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", fullPath, err)
			}
		case map[string]interface{}:
			// It's a directory
			if err := fs.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
			if err := createTreeRecursive(fs, fullPath, v); err != nil {
				return err
			}
		case nil:
			// Empty directory
			if err := fs.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fullPath, err)
			}
		default:
			return fmt.Errorf("unsupported type %T for %s", v, name)
		}
	}
	return nil
}
