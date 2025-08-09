package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// IsTextFile checks if a file contains text (non-binary) content
// It reads the first 512 bytes and checks for binary indicators
func IsTextFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = file.Close()
	}()

	// Read first 512 bytes to check if binary
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	// Empty files are considered text
	if n == 0 {
		return true, nil
	}

	return !IsBinaryContent(buffer[:n]), nil
}

// IsBinaryContent checks if a byte slice contains binary data
// It uses multiple heuristics:
// - Presence of null bytes
// - Ratio of printable characters
// - Common binary file signatures
func IsBinaryContent(data []byte) bool {
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

// ReadTextFileContent reads a text file's content with size and binary checks
// Returns an error if the file is binary or too large
func ReadTextFileContent(path string, maxSize int64) (string, error) {
	// Check file size first
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	// Skip files larger than the limit for performance
	if info.Size() > maxSize {
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
	if IsBinaryContent(buffer[:n]) {
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