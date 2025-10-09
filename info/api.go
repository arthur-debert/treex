package info

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
)

// InfoAPI provides the main interface for info file operations
type InfoAPI struct {
	fs        InfoFileSystem
	setLoader *InfoFileSetLoader
	setWriter *InfoFileSetWriter
}

// NewInfoAPI creates a new info API instance using afero filesystem
func NewInfoAPI(fs afero.Fs) *InfoAPI {
	afs := NewAferoInfoFileSystem(fs)
	return &InfoAPI{
		fs:        afs,
		setLoader: NewInfoFileSetLoader(afs),
		setWriter: NewInfoFileSetWriter(afs),
	}
}

// NewInfoAPIWithFileSystem creates a new info API instance with custom filesystem
func NewInfoAPIWithFileSystem(fs InfoFileSystem) *InfoAPI {
	return &InfoAPI{
		fs:        fs,
		setLoader: NewInfoFileSetLoader(fs),
		setWriter: NewInfoFileSetWriter(fs),
	}
}

// Gather collects and merges all annotations from .info files in a directory tree
// Uses the new InfoFileSet approach for optimal performance
func (api *InfoAPI) Gather(rootPath string) (map[string]Annotation, error) {
	infoFileSet, err := api.setLoader.LoadFromPath(rootPath)
	if err != nil {
		return nil, err
	}

	return infoFileSet.Gather(), nil
}

// Validate validates all .info files in a directory tree
func (api *InfoAPI) Validate(rootPath string) (*ValidationResult, error) {
	infoFileSet, err := api.setLoader.LoadFromPath(rootPath)
	if err != nil {
		return nil, err
	}

	return infoFileSet.Validate(), nil
}

// Add adds a new annotation directly to a specific .info file
func (api *InfoAPI) Add(infoFilePath, targetPath, annotation string) error {
	// Validate the annotation path is reasonable (not empty, not just whitespace)
	if strings.TrimSpace(targetPath) == "" {
		return fmt.Errorf("target path cannot be empty")
	}
	if strings.TrimSpace(annotation) == "" {
		return fmt.Errorf("annotation cannot be empty")
	}

	// Load existing InfoFile or create new one
	var infoFile *InfoFile
	existingInfoFile, err := api.setLoader.LoadSingleInfoFile(infoFilePath)
	if err != nil {
		// File doesn't exist, create new empty InfoFile
		infoFile = NewInfoFile(infoFilePath, "")
	} else {
		infoFile = existingInfoFile
	}

	// Add annotation using InfoFile method (targetPath is used as-is)
	success := infoFile.AddAnnotationForPath(targetPath, annotation)
	if !success {
		return fmt.Errorf("annotation already exists for path %q", targetPath)
	}

	// Write updated InfoFile back to disk
	return api.setWriter.WriteSingleInfoFile(infoFile)
}

// Remove removes an annotation from a specific .info file
func (api *InfoAPI) Remove(infoFilePath, targetPath string) error {
	// Validate the target path
	if strings.TrimSpace(targetPath) == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	// Load the InfoFile
	infoFile, err := api.setLoader.LoadSingleInfoFile(infoFilePath)
	if err != nil {
		return fmt.Errorf("cannot load .info file %q: %w", infoFilePath, err)
	}

	// Remove annotation using InfoFile method
	success := infoFile.RemoveAnnotationForPath(targetPath)
	if !success {
		return fmt.Errorf("annotation not found for path %q", targetPath)
	}

	// Write updated InfoFile back to disk
	return api.setWriter.WriteSingleInfoFile(infoFile)
}

// Update updates an existing annotation in a specific .info file
func (api *InfoAPI) Update(infoFilePath, targetPath, newAnnotation string) error {
	// Validate inputs
	if strings.TrimSpace(targetPath) == "" {
		return fmt.Errorf("target path cannot be empty")
	}
	if strings.TrimSpace(newAnnotation) == "" {
		return fmt.Errorf("annotation cannot be empty")
	}

	// Load the InfoFile
	infoFile, err := api.setLoader.LoadSingleInfoFile(infoFilePath)
	if err != nil {
		return fmt.Errorf("cannot load .info file %q: %w", infoFilePath, err)
	}

	// Update annotation using InfoFile method
	success := infoFile.UpdateAnnotationForPath(targetPath, newAnnotation)
	if !success {
		return fmt.Errorf("annotation not found for path %q", targetPath)
	}

	// Write updated InfoFile back to disk
	return api.setWriter.WriteSingleInfoFile(infoFile)
}

// List lists all current annotations in a directory tree
func (api *InfoAPI) List(rootPath string) ([]Annotation, error) {
	annotations, err := api.Gather(rootPath)
	if err != nil {
		return nil, err
	}

	var result []Annotation
	for _, ann := range annotations {
		result = append(result, ann)
	}

	return result, nil
}

// GetAnnotation retrieves the effective annotation for a specific path
func (api *InfoAPI) GetAnnotation(targetPath string) (*Annotation, error) {
	annotations, err := api.Gather(".")
	if err != nil {
		return nil, err
	}

	if ann, exists := annotations[targetPath]; exists {
		return &ann, nil
	}

	return nil, fmt.Errorf("no annotation found for path %q", targetPath)
}

// Clean removes invalid or redundant annotations
func (api *InfoAPI) Clean(rootPath string) (*CleanResult, error) {
	infoFileSet, err := api.setLoader.LoadFromPath(rootPath)
	if err != nil {
		return nil, err
	}

	cleanResult, cleanedSet := infoFileSet.Clean()

	// Write the cleaned InfoFileSet back to disk
	err = api.setWriter.WriteInfoFileSet(cleanedSet)
	if err != nil {
		return nil, err
	}

	// Return the CleanResult from InfoFileSet.Clean()
	return cleanResult, nil
}

// Distribute redistributes annotations to their optimal .info files based on path proximity
func (api *InfoAPI) Distribute(rootPath string) error {
	infoFileSet, err := api.setLoader.LoadFromPath(rootPath)
	if err != nil {
		return err
	}

	distributedSet := infoFileSet.Distribute()

	// Write the distributed InfoFileSet back to disk
	return api.setWriter.WriteInfoFileSet(distributedSet)
}
