// Package info provides a plugin for annotation-based file filtering using .info files
package info

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jwaldrip/treex/treex/info"
	"github.com/jwaldrip/treex/treex/plugins"
	"github.com/spf13/afero"
)

// InfoPlugin categorizes files based on whether they have annotations in .info files
// It finds directories containing .info files and categorizes files as annotated or non-annotated
type InfoPlugin struct{}

// NewInfoPlugin creates a new info plugin instance
func NewInfoPlugin() *InfoPlugin {
	return &InfoPlugin{}
}

// Name returns the plugin identifier
func (p *InfoPlugin) Name() string {
	return "info"
}

// FindRoots discovers directories containing .info files
// Returns the parent directories of .info files as annotation roots
func (p *InfoPlugin) FindRoots(fs afero.Fs, searchRoot string) ([]string, error) {
	var roots []string
	rootsMap := make(map[string]bool) // Deduplicate roots

	// Walk the filesystem looking for .info files
	err := afero.Walk(fs, searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors, don't fail the entire search
		}

		// Check if this is a .info file
		if !info.IsDir() && info.Name() == ".info" {
			// The root is the directory containing the .info file
			infoDir := filepath.Dir(path)

			// Convert to relative path from search root
			relativeRoot, err := filepath.Rel(searchRoot, infoDir)
			if err != nil {
				return nil // Skip this root if we can't make it relative
			}

			// Normalize "." for current directory
			if relativeRoot == "." || relativeRoot == "" {
				relativeRoot = "."
			}

			// Add to roots if not already present (deduplicate)
			if !rootsMap[relativeRoot] {
				roots = append(roots, relativeRoot)
				rootsMap[relativeRoot] = true
			}
		}

		return nil
	})

	return roots, err
}

// ProcessRoot analyzes .info files in a directory root and categorizes files
// Uses the info collector to parse annotations and determine which files are annotated
func (p *InfoPlugin) ProcessRoot(fs afero.Fs, rootPath string) (*plugins.Result, error) {
	result := &plugins.Result{
		PluginName: p.Name(),
		RootPath:   rootPath,
		Categories: make(map[string][]string),
		Metadata:   make(map[string]interface{}),
	}

	// Initialize categories
	result.Categories["annotated"] = make([]string, 0)
	result.Categories["non-annotated"] = make([]string, 0)

	// Create a collector to parse .info files in this root
	collector := info.NewCollector()
	annotations, err := collector.CollectAnnotations(fs, rootPath)
	if err != nil {
		// If we can't collect annotations, return empty result (not an error)
		// This handles cases where .info files exist but are unreadable/invalid
		result.Metadata["error"] = "failed to collect annotations: " + err.Error()
		result.Metadata["total_files"] = 0
		result.Metadata["annotated_count"] = 0
		result.Metadata["non_annotated_count"] = 0
		result.Metadata["total_annotations"] = 0
		result.Metadata["info_file_count"] = 0
		result.Metadata["info_files"] = []string{}
		result.Metadata["annotation_sources"] = map[string]int{}
		return result, nil
	}

	// Track which files have annotations
	// Need to adjust annotation paths to be relative to rootPath
	annotatedFiles := make(map[string]bool)
	for annotationPath := range annotations {
		// If rootPath is not ".", we need to make annotation paths relative to rootPath
		relativePath := annotationPath
		if rootPath != "." {
			// Try to make the annotation path relative to rootPath
			if rel, err := filepath.Rel(rootPath, annotationPath); err == nil && !strings.HasPrefix(rel, "..") {
				relativePath = rel
			} else {
				// If the annotation is outside rootPath scope, keep original path
				relativePath = annotationPath
			}
		}

		// Normalize path separators
		relativePath = filepath.ToSlash(relativePath)
		annotatedFiles[relativePath] = true
	}

	// Walk through all files in the root to categorize them
	var allFiles []string
	err = afero.Walk(fs, rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}

		// Skip .info files themselves
		if !info.IsDir() && info.Name() == ".info" {
			return nil
		}

		// Calculate relative path from root
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil // Skip if we can't calculate relative path
		}

		// Normalize "." for current directory
		if relativePath == "." || relativePath == "" {
			relativePath = "."
		}

		// Normalize path separators for consistency
		normalizedPath := filepath.ToSlash(relativePath)
		allFiles = append(allFiles, normalizedPath)

		// Categorize based on whether the file has an annotation
		if annotatedFiles[normalizedPath] {
			result.Categories["annotated"] = append(result.Categories["annotated"], normalizedPath)
		} else {
			result.Categories["non-annotated"] = append(result.Categories["non-annotated"], normalizedPath)
		}

		return nil
	})

	if err != nil {
		result.Metadata["error"] = "failed to walk files: " + err.Error()
		return result, nil
	}

	// Add metadata about the annotation processing
	result.Metadata["total_files"] = len(allFiles)
	result.Metadata["annotated_count"] = len(result.Categories["annotated"])
	result.Metadata["non_annotated_count"] = len(result.Categories["non-annotated"])
	result.Metadata["total_annotations"] = len(annotations)

	// Find all .info files in the filesystem (not just those with winning annotations)
	allInfoFiles := make([]string, 0)
	annotationSources := make(map[string]int) // Map info file to annotation count

	// Walk filesystem to find all .info files
	_ = afero.Walk(fs, rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}

		if !info.IsDir() && info.Name() == ".info" {
			// Calculate relative path from root
			relativePath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return nil // Skip if we can't calculate relative path
			}

			// Normalize path separators
			normalizedPath := filepath.ToSlash(relativePath)
			allInfoFiles = append(allInfoFiles, normalizedPath)
		}

		return nil
	})

	// Count annotations per .info file that contributed to final results
	for _, annotation := range annotations {
		infoFile := annotation.InfoFile
		annotationSources[infoFile]++
	}

	result.Metadata["info_files"] = allInfoFiles
	result.Metadata["info_file_count"] = len(allInfoFiles)
	result.Metadata["annotation_sources"] = annotationSources

	// Add sample annotations for debugging/inspection
	if len(annotations) > 0 {
		sampleAnnotations := make(map[string]string)
		count := 0
		for path, annotation := range annotations {
			if count >= 5 { // Limit to first 5 annotations
				break
			}
			sampleAnnotations[path] = annotation.Annotation
			count++
		}
		result.Metadata["sample_annotations"] = sampleAnnotations
	}

	return result, nil
}

// GetAnnotationDetails extracts detailed annotation information for a root
// This is a helper method for getting more detailed annotation metadata
func (p *InfoPlugin) GetAnnotationDetails(fs afero.Fs, rootPath string) (map[string]interface{}, error) {
	details := make(map[string]interface{})

	collector := info.NewCollector()
	annotations, err := collector.CollectAnnotations(fs, rootPath)
	if err != nil {
		return details, err
	}

	// Group annotations by .info file (only winning annotations are included)
	byInfoFile := make(map[string][]info.Annotation)
	for _, annotation := range annotations {
		infoFile := annotation.InfoFile
		byInfoFile[infoFile] = append(byInfoFile[infoFile], annotation)
	}

	details["annotations_by_file"] = byInfoFile
	details["total_info_files"] = len(byInfoFile)
	details["total_annotations"] = len(annotations)

	// Calculate depth distribution of .info files that have winning annotations
	depthCounts := make(map[int]int)
	for infoFile := range byInfoFile {
		depth := strings.Count(infoFile, string(filepath.Separator))
		if infoFile == ".info" {
			depth = 0 // Root .info file
		}
		depthCounts[depth]++
	}
	details["info_file_depths"] = depthCounts

	return details, nil
}

// init registers the info plugin with the default registry
func init() {
	if err := plugins.RegisterPlugin(NewInfoPlugin()); err != nil {
		panic("failed to register info plugin: " + err.Error())
	}
}
