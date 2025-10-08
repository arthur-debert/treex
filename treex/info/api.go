package info

import (
	"fmt"
	"path/filepath"

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

// Add adds a new annotation to the appropriate .info file
func (api *InfoAPI) Add(targetPath, annotation string) error {
	// Determine the appropriate .info file for this target path
	infoFilePath := api.determineInfoFile(targetPath)

	// Load existing InfoFile or create new one
	var infoFile *InfoFile
	existingInfoFile, err := api.setLoader.LoadSingleInfoFile(infoFilePath)
	if err != nil {
		// File doesn't exist, create new empty InfoFile
		infoFile = NewInfoFile(infoFilePath, "")
	} else {
		infoFile = existingInfoFile
	}

	// Convert target path to relative path for the .info file
	relativePath := api.makeRelativePathForAdd(targetPath, infoFilePath)

	// Add annotation using InfoFile method
	success := infoFile.AddAnnotationForPath(relativePath, annotation)
	if !success {
		return fmt.Errorf("annotation already exists for path %q", relativePath)
	}

	// Write updated InfoFile back to disk
	return api.setWriter.WriteSingleInfoFile(infoFile)
}

// Remove removes an annotation for a specific path
func (api *InfoAPI) Remove(targetPath string) error {
	// Gather all annotations to find the target
	annotations, err := api.Gather(".")
	if err != nil {
		return err
	}

	// Check if annotation exists
	targetAnnotation, exists := annotations[targetPath]
	if !exists {
		return fmt.Errorf("no annotation found for path %q", targetPath)
	}

	// Load the InfoFile containing the annotation
	infoFile, err := api.setLoader.LoadSingleInfoFile(targetAnnotation.InfoFile)
	if err != nil {
		return fmt.Errorf("cannot load .info file %q: %w", targetAnnotation.InfoFile, err)
	}

	// Convert target path to relative path for the .info file
	relativePath := api.makeRelativePathForAdd(targetPath, targetAnnotation.InfoFile)

	// Remove annotation using InfoFile method
	success := infoFile.RemoveAnnotationForPath(relativePath)
	if !success {
		return fmt.Errorf("annotation not found in content for path %q", relativePath)
	}

	// Write updated InfoFile back to disk
	return api.setWriter.WriteSingleInfoFile(infoFile)
}

// Update updates an existing annotation
func (api *InfoAPI) Update(targetPath, newAnnotation string) error {
	// Gather all annotations to find the target
	annotations, err := api.Gather(".")
	if err != nil {
		return err
	}

	// Check if annotation exists
	targetAnnotation, exists := annotations[targetPath]
	if !exists {
		return fmt.Errorf("no annotation found for path %q", targetPath)
	}

	// Load the InfoFile containing the annotation
	infoFile, err := api.setLoader.LoadSingleInfoFile(targetAnnotation.InfoFile)
	if err != nil {
		return fmt.Errorf("cannot load .info file %q: %w", targetAnnotation.InfoFile, err)
	}

	// Convert target path to relative path for the .info file
	relativePath := api.makeRelativePathForAdd(targetPath, targetAnnotation.InfoFile)

	// Update annotation using InfoFile method
	success := infoFile.UpdateAnnotationForPath(relativePath, newAnnotation)
	if !success {
		return fmt.Errorf("annotation not found in content for path %q", relativePath)
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

// determineInfoFile determines the appropriate .info file path for a target path
func (api *InfoAPI) determineInfoFile(targetPath string) string {
	// Simple strategy: use .info file in the same directory as the target
	dir := filepath.Dir(targetPath)
	if dir == "" || dir == "." {
		return ".info"
	}
	return filepath.Join(dir, ".info")
}

// makeRelativePathForAdd converts target path to relative path for the .info file
func (api *InfoAPI) makeRelativePathForAdd(targetPath, infoFilePath string) string {
	infoDir := filepath.Dir(infoFilePath)
	targetDir := filepath.Dir(targetPath)

	// If target is in same directory as info file, use just the filename
	if infoDir == targetDir || (infoDir == "." && targetDir == "") {
		return filepath.Base(targetPath)
	}

	// Calculate relative path from info file directory to target
	rel, err := filepath.Rel(infoDir, targetPath)
	if err != nil {
		// Fallback to the original path if calculation fails
		return targetPath
	}

	return filepath.Clean(rel)
}
