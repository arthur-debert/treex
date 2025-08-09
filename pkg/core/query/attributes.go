package query

import (
	"fmt"
	"os"
	
	"github.com/adebert/treex/pkg/core/types"
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
	
	return nil
}