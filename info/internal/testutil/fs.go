package testutil

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

type TestFS struct {
	afero.Fs
}

func NewTestFS() *TestFS {
	return &TestFS{Fs: afero.NewMemMapFs()}
}

func (fs *TestFS) CreateTree(root string, structure map[string]interface{}) error {
	return fs.createTreeRecursive(root, structure)
}

func (fs *TestFS) MustCreateTree(root string, structure map[string]interface{}) {
	if err := fs.CreateTree(root, structure); err != nil {
		panic(fmt.Sprintf("failed to create test tree: %v", err))
	}
}

func (fs *TestFS) createTreeRecursive(basePath string, structure map[string]interface{}) error {
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
			if err := fs.createTreeRecursive(fullPath, v); err != nil {
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

// Helper to create a simple test tree structure
func SimpleTestTree() map[string]interface{} {
	return map[string]interface{}{
		"src": map[string]interface{}{
			"main.go": "package main\n\nfunc main() {}\n",
			"lib": map[string]interface{}{
				"utils.go": "package lib\n",
				"test.go":  "package lib\n",
			},
		},
		"docs": map[string]interface{}{
			"README.txt": "Documentation\n",
		},
		".gitignore": "*.tmp\n",
	}
}

// SetFileTime sets the modification time of a file (useful for testing)
func (fs *TestFS) SetFileTime(path string, modTime time.Time) error {
	return fs.Chtimes(path, modTime, modTime)
}
