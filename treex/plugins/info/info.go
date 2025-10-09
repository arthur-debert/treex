// Package info provides a plugin for annotation-based file filtering using .info files
package info

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jwaldrip/treex/info"
	"github.com/jwaldrip/treex/treex/plugins"
	"github.com/jwaldrip/treex/treex/types"
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
		Cache:      make(map[string]interface{}),
	}

	// Initialize categories
	result.Categories["annotated"] = make([]string, 0)

	// Create an InfoAPI to parse .info files in this root
	api := info.NewInfoAPI(fs)
	annotations, err := api.Gather(rootPath)
	if err != nil {
		// If we can't collect annotations, return empty result (not an error)
		// This handles cases where .info files exist but are unreadable/invalid
		return result, nil
	}

	// Store the raw annotations in cache for efficient data enrichment
	result.Cache["annotations"] = annotations

	// Add only annotated files to the result
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
		result.Categories["annotated"] = append(result.Categories["annotated"], relativePath)
	}

	return result, nil
}

// GetAnnotationDetails extracts detailed annotation information for a root
// This is a helper method for getting more detailed annotation metadata
func (p *InfoPlugin) GetAnnotationDetails(fs afero.Fs, rootPath string) (map[string]interface{}, error) {
	details := make(map[string]interface{})

	api := info.NewInfoAPI(fs)
	annotations, err := api.Gather(rootPath)
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

// GetCategories returns the filter categories provided by the info plugin
// Implements FilterPlugin interface
func (p *InfoPlugin) GetCategories() []plugins.FilterPluginCategory {
	return []plugins.FilterPluginCategory{
		{
			Name:        "annotated",
			Description: "Files with annotations in .info files",
		},
	}
}

// EnrichNode attaches annotation data to nodes that have annotations
// Implements DataPlugin interface
func (p *InfoPlugin) EnrichNode(fs afero.Fs, node *types.Node) error {
	// Check if this node has an annotation by looking for .info files
	// in the current directory or parent directories

	// Get the directory containing this file
	nodeDir := filepath.Dir(node.Path)
	if node.IsDir {
		nodeDir = node.Path
	}

	// Use the InfoAPI to find annotation for this specific path
	api := info.NewInfoAPI(fs)

	// Try to find annotation starting from the node's directory
	searchPath := "."
	if nodeDir != "." && nodeDir != "" {
		searchPath = nodeDir
	}

	annotations, err := api.Gather(searchPath)
	if err != nil {
		// If we can't gather annotations, skip enrichment (not an error)
		return nil
	}

	// Look for annotation for this specific file
	for filePath, annotation := range annotations {
		// Normalize paths for comparison
		normalizedFilePath := filepath.ToSlash(filePath)
		normalizedNodePath := filepath.ToSlash(node.Path)

		if normalizedFilePath == normalizedNodePath {
			// Found annotation for this node - convert to types.Annotation and store
			nodeAnnotation := &types.Annotation{
				Path:  annotation.Path,
				Notes: annotation.Annotation,
			}
			node.SetPluginData("info", nodeAnnotation)
			break
		}
	}

	return nil
}

// EnrichNodeWithCache attaches annotation data using cached results from filtering phase
// Implements CachedDataPlugin interface for efficient data enrichment
func (p *InfoPlugin) EnrichNodeWithCache(fs afero.Fs, node *types.Node, pluginResults []*plugins.Result) error {
	// Look through all plugin results to find cached annotations
	for _, result := range pluginResults {
		if result.PluginName != p.Name() {
			continue
		}

		// Check if we have cached annotations for this result
		cachedAnnotations, exists := result.Cache["annotations"]
		if !exists {
			continue
		}

		// Type assert to get the annotations map
		annotations, ok := cachedAnnotations.(map[string]info.Annotation)
		if !ok {
			continue
		}

		// Look for annotation for this specific file
		for filePath, annotation := range annotations {
			// Handle both absolute and relative paths in cache
			// Make filePath relative to match node.Path which is always relative
			var normalizedFilePath string
			if filepath.IsAbs(filePath) {
				// Try to make absolute path relative to result root
				if rel, err := filepath.Rel(result.RootPath, filePath); err == nil && !strings.HasPrefix(rel, "..") {
					normalizedFilePath = filepath.ToSlash(rel)
				} else {
					// If we can't make it relative, use basename for comparison
					normalizedFilePath = filepath.ToSlash(filepath.Base(filePath))
				}
			} else {
				normalizedFilePath = filepath.ToSlash(filePath)
			}

			normalizedNodePath := filepath.ToSlash(node.Path)

			if normalizedFilePath == normalizedNodePath {
				// Found annotation for this node - convert to types.Annotation and store
				nodeAnnotation := &types.Annotation{
					Path:  annotation.Path,
					Notes: annotation.Annotation,
				}
				node.SetPluginData("info", nodeAnnotation)
				return nil
			}
		}
	}

	// No cached annotation found - this is normal for non-annotated files
	return nil
}

// init registers the info plugin with the default registry
func init() {
	if err := plugins.RegisterPlugin(NewInfoPlugin()); err != nil {
		panic("failed to register info plugin: " + err.Error())
	}
}
