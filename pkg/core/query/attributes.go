package query

import (
	"fmt"
	"io"
	"os"
	"strings"
	
	"github.com/adebert/treex/pkg/core/limits"
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
	
	return nil
}

// readFileContent reads the content of a file with size and binary checks
func readFileContent(path string) (string, error) {
	// Check file size first
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	
	// Skip files larger than the limit for performance
	if info.Size() > limits.DefaultMaxFileSize {
		return "", fmt.Errorf("file too large")
	}
	
	// Open and read file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()
	
	// Read first 512 bytes to check if binary
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	
	// Check if file appears to be binary
	if isBinary(buffer[:n]) {
		return "", fmt.Errorf("binary file")
	}
	
	// Reset to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	
	// Read entire file
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	
	return string(content), nil
}

// isBinary checks if a byte slice contains binary data
func isBinary(data []byte) bool {
	// Check for null bytes which indicate binary
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	
	// Check if most bytes are printable
	printable := 0
	for _, b := range data {
		if b >= 32 && b <= 126 || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}
	
	// If less than 80% printable, consider it binary
	if len(data) > 0 && float64(printable)/float64(len(data)) < 0.8 {
		return true
	}
	
	// Check for common binary file signatures
	if len(data) >= 4 {
		// ELF
		if data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
			return true
		}
		// PNG
		if data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' {
			return true
		}
		// PDF
		if strings.HasPrefix(string(data), "%PDF") {
			return true
		}
		// JPEG
		if data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
			return true
		}
	}
	
	return false
}