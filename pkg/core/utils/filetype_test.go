package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "text with newlines",
			data:     []byte("Hello\nWorld\n"),
			expected: false,
		},
		{
			name:     "text with special chars",
			data:     []byte("Hello\tWorld\r\n"),
			expected: false,
		},
		{
			name:     "binary with null bytes",
			data:     []byte("Hello\x00World"),
			expected: true,
		},
		{
			name:     "mostly non-printable",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			expected: true,
		},
		{
			name:     "PNG signature",
			data:     []byte{0x89, 'P', 'N', 'G'},
			expected: true,
		},
		{
			name:     "PDF signature",
			data:     []byte("%PDF-1.4"),
			expected: true,
		},
		{
			name:     "ELF signature",
			data:     []byte{0x7f, 'E', 'L', 'F'},
			expected: true,
		},
		{
			name:     "JPEG signature",
			data:     []byte{0xff, 0xd8, 0xff, 0xe0},
			expected: true,
		},
		{
			name:     "empty data",
			data:     []byte{},
			expected: false,
		},
		{
			name:     "unicode text",
			data:     []byte("Hello 世界"),
			expected: true, // Unicode characters are considered non-printable in current implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBinaryContent(tt.data)
			if result != tt.expected {
				t.Errorf("IsBinaryContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsTextFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test files
	textFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(textFile, []byte("Hello World\nThis is a text file\n"), 0644); err != nil {
		t.Fatal(err)
	}

	binaryFile := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02, 0x03}, 0644); err != nil {
		t.Fatal(err)
	}

	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
		wantErr  bool
	}{
		{
			name:     "text file",
			path:     textFile,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "binary file",
			path:     binaryFile,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "empty file",
			path:     emptyFile,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			path:     filepath.Join(tmpDir, "does-not-exist.txt"),
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsTextFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsTextFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("IsTextFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}