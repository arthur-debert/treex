package query

import (
	"fmt"
	"os"
	
	"github.com/adebert/treex/pkg/core/limits"
	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/core/utils"
)

// registerAllAttributes registers all built-in attributes
func registerAllAttributes() error {
	registry := GetGlobalRegistry()
	
	// Register file-name attribute
	if err := registry.RegisterAttribute(&Attribute{
		Name:        "file-name",
		Type:        StringType,
		Description: "The name of the file or directory (without path)",
		Extractor: func(node *types.Node) (interface{}, error) {
			return node.Name, nil
		},
	}); err != nil {
		return err
	}
	
	// Register path attribute
	if err := registry.RegisterAttribute(&Attribute{
		Name:        "path",
		Type:        StringType,
		Description: "The full path from the root",
		Extractor: func(node *types.Node) (interface{}, error) {
			return node.RelativePath, nil
		},
	}); err != nil {
		return err
	}
	
	// Register size attribute
	if err := registry.RegisterAttribute(&Attribute{
		Name:        "size",
		Type:        NumericType,
		Description: "The size of the file or directory in bytes",
		Extractor: func(node *types.Node) (interface{}, error) {
			// For directories, check if size has been pre-calculated (e.g., by overlay)
			if node.IsDir {
				// Check metadata for pre-calculated size
				if sizeVal, exists := node.Metadata["size_bytes"]; exists {
					if size, ok := sizeVal.(int64); ok {
						return size, nil
					}
				}
				// If no pre-calculated size, return 0 for POC
				return int64(0), nil
			}
			
			// For files, get size from filesystem
			info, err := os.Stat(node.Path)
			if err != nil {
				return int64(0), fmt.Errorf("failed to stat file: %w", err)
			}
			
			return info.Size(), nil
		},
	}); err != nil {
		return err
	}
	
	// Register text attribute for searching file contents
	if err := registry.RegisterAttribute(&Attribute{
		Name:        "text",
		Type:        StringType,
		Description: "The text content of the file",
		Extractor: func(node *types.Node) (interface{}, error) {
			// Skip directories
			if node.IsDir {
				return "", nil
			}
			
			// Read file content
			content, err := readFileContent(node.Path)
			if err != nil {
				// Return empty string on error to allow filtering to continue
				return "", nil
			}
			
			return content, nil
		},
	}); err != nil {
		return err
	}
	
	// Register cloc attribute for lines of code
	if err := registry.RegisterAttribute(&Attribute{
		Name:        "cloc",
		Type:        NumericType,
		Description: "The lines of code (excluding blanks and comments)",
		Extractor: func(node *types.Node) (interface{}, error) {
			// Check metadata for cloc value (set by plugin)
			if clocVal, exists := node.Metadata["cloc_cloc"]; exists {
				if cloc, ok := clocVal.(int64); ok {
					return cloc, nil
				}
			}
			// If no cloc data available, return 0
			return int64(0), nil
		},
	}); err != nil {
		return err
	}
	
	return nil
}

// readFileContent reads the content of a file with size and binary checks
func readFileContent(path string) (string, error) {
	return utils.ReadTextFileContent(path, limits.DefaultMaxFileSize)
}

