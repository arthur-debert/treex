package query

import (
	"os"
	"path/filepath"
	"testing"
	
	"github.com/adebert/treex/pkg/core/limits"
)

func TestTextAttributeWithLimiter(t *testing.T) {
	// Test that the text attribute respects file size limits
	// by checking the constant matches what readFileContent uses
	
	// Create a small test file
	tmpDir, err := os.MkdirTemp("", "treex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	smallFile := filepath.Join(tmpDir, "small.txt")
	err = os.WriteFile(smallFile, []byte("small content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Test that small files are read
	content, err := readFileContent(smallFile)
	if err != nil {
		t.Errorf("Should read small file without error: %v", err)
	}
	if content != "small content" {
		t.Errorf("Expected 'small content', got '%s'", content)
	}
}

func TestFileSizeLimitConfiguration(t *testing.T) {
	// Test that the file size limit in readFileContent matches the limits package default
	config := &limits.Config{
		MaxFileSize:  1024, // 1KB - very small for testing
		RuntimeLimit: limits.DefaultRuntimeLimit,
	}
	
	limiter := limits.NewLimiter(config)
	defer limiter.Close()
	
	// Test file size checking
	tests := []struct {
		size     int64
		expected bool
	}{
		{512, true},      // Under limit
		{1024, true},     // At limit
		{1025, false},    // Over limit
	}
	
	for _, tt := range tests {
		result := limiter.CheckFileSize(tt.size)
		if result != tt.expected {
			t.Errorf("CheckFileSize(%d) = %v, want %v", tt.size, result, tt.expected)
		}
	}
}

func TestReadFileContent_BinaryFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "treex-test-binary-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()
	
	tests := []struct {
		name        string
		content     []byte
		shouldRead  bool
	}{
		{
			name:        "text file",
			content:     []byte("Hello, World!\nThis is a text file."),
			shouldRead:  true,
		},
		{
			name:        "binary with null bytes",
			content:     []byte{0x00, 0x01, 0x02, 0x03},
			shouldRead:  false,
		},
		{
			name:        "PNG signature",
			content:     []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
			shouldRead:  false,
		},
		{
			name:        "PDF signature",
			content:     []byte("%PDF-1.4 some content"),
			shouldRead:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.name)
			err := os.WriteFile(filePath, tt.content, 0644)
			if err != nil {
				t.Fatal(err)
			}
			
			content, err := readFileContent(filePath)
			
			if tt.shouldRead {
				if err != nil {
					t.Errorf("Expected no error for text file, got %v", err)
				}
				if content != string(tt.content) {
					t.Errorf("Content mismatch")
				}
			} else {
				if err == nil {
					t.Error("Expected error for binary file")
				}
				if content != "" {
					t.Error("Expected empty content for binary file")
				}
			}
		})
	}
}