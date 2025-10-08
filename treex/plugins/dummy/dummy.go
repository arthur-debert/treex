// Package dummy provides a simple test plugin for validating the plugin system
package dummy

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"treex/treex/plugins"
)

// DummyPlugin is a test plugin that demonstrates the plugin interface
// It finds directories containing a ".dummy" file and categorizes files by extension
type DummyPlugin struct{}

// NewDummyPlugin creates a new dummy plugin instance
func NewDummyPlugin() *DummyPlugin {
	return &DummyPlugin{}
}

// Name returns the plugin identifier
func (p *DummyPlugin) Name() string {
	return "dummy"
}

// FindRoots finds directories containing a ".dummy" marker file
// This simulates how real plugins (like git) look for their marker files
func (p *DummyPlugin) FindRoots(fs afero.Fs, searchRoot string) ([]string, error) {
	var roots []string

	// Walk the filesystem looking for .dummy files
	err := afero.Walk(fs, searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors, don't fail the entire search
		}

		// Check if this is a .dummy file
		if !info.IsDir() && info.Name() == ".dummy" {
			// The root is the directory containing the .dummy file
			rootDir := filepath.Dir(path)

			// Convert to relative path from search root
			relativeRoot, err := filepath.Rel(searchRoot, rootDir)
			if err != nil {
				return nil // Skip this root if we can't make it relative
			}

			// Normalize "." for current directory
			if relativeRoot == "." || relativeRoot == "" {
				relativeRoot = "."
			}

			roots = append(roots, relativeRoot)
		}

		return nil
	})

	return roots, err
}

// ProcessRoot analyzes files in a root directory and categorizes them
// This dummy implementation categorizes files by their extensions
func (p *DummyPlugin) ProcessRoot(fs afero.Fs, rootPath string) (*plugins.Result, error) {
	result := &plugins.Result{
		PluginName: p.Name(),
		RootPath:   rootPath,
		Categories: make(map[string][]string),
		Metadata:   make(map[string]interface{}),
	}

	// Count files and categorize by extension
	var totalFiles int

	err := afero.Walk(fs, rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}

		// Only process files, not directories
		if info.IsDir() {
			return nil
		}

		totalFiles++

		// Get relative path from root
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil // Skip files we can't make relative
		}

		// Categorize by file extension
		ext := strings.ToLower(filepath.Ext(relativePath))
		if ext == "" {
			ext = "no-extension"
		} else {
			ext = ext[1:] // Remove the leading dot
		}

		// Add to appropriate category
		if result.Categories[ext] == nil {
			result.Categories[ext] = make([]string, 0)
		}
		result.Categories[ext] = append(result.Categories[ext], relativePath)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Add metadata about the processing
	result.Metadata["total_files"] = totalFiles
	result.Metadata["total_categories"] = len(result.Categories)

	return result, nil
}

// init registers the dummy plugin with the default registry
func init() {
	if err := plugins.RegisterPlugin(NewDummyPlugin()); err != nil {
		panic("failed to register dummy plugin: " + err.Error())
	}
}