package info

import (
	"path/filepath"
	"sort"
	"strings"
)

// FileOperation represents an operation to be performed on a file
type FileOperation struct {
	Type     FileOperationType
	FilePath string
	Content  string
}

// FileOperationType represents the type of file operation
type FileOperationType string

const (
	// OpCreate indicates creating a new file
	OpCreate FileOperationType = "create"
	// OpUpdate indicates updating an existing file
	OpUpdate FileOperationType = "update"
	// OpDelete indicates deleting a file
	OpDelete FileOperationType = "delete"
)

// Distributor handles distributing annotations to .info files
type Distributor struct {
	editor *Editor
}

// NewDistributor creates a new distributor instance
func NewDistributor() *Distributor {
	return &Distributor{
		editor: NewEditor(),
	}
}

// DistributeAnnotations takes a map of annotations and returns file operations
// needed to write them to .info files
func (d *Distributor) DistributeAnnotations(annotations map[string]Annotation) []FileOperation {
	// Group annotations by .info file
	fileGroups := make(map[string][]Annotation)
	for _, ann := range annotations {
		fileGroups[ann.InfoFile] = append(fileGroups[ann.InfoFile], ann)
	}

	var operations []FileOperation
	for infoFile, anns := range fileGroups {
		content := d.editor.GenerateContent(anns, infoFile)
		operations = append(operations, FileOperation{
			Type:     OpUpdate,
			FilePath: infoFile,
			Content:  content,
		})
	}

	return operations
}

// DistributeToFiles takes annotations and distributes them to appropriate .info files
// using a specific distribution strategy
func (d *Distributor) DistributeToFiles(annotations []Annotation, strategy DistributionStrategy) []FileOperation {
	// Group by target path to handle conflicts
	pathGroups := make(map[string][]Annotation)
	for _, ann := range annotations {
		infoDir := filepath.Dir(ann.InfoFile)
		targetPath := filepath.Join(infoDir, ann.Path)
		targetPath = filepath.Clean(targetPath)
		pathGroups[targetPath] = append(pathGroups[targetPath], ann)
	}

	// Resolve conflicts and distribute according to strategy
	finalAnnotations := make(map[string]Annotation)
	for targetPath, group := range pathGroups {
		winner := d.resolveConflict(group, strategy)

		// Determine the appropriate .info file for this annotation
		infoFile := d.determineInfoFile(targetPath, strategy)

		// Calculate relative path from .info file to target
		relativePath, err := filepath.Rel(filepath.Dir(infoFile), targetPath)
		if err != nil {
			relativePath = targetPath
		}

		// Create new annotation with corrected paths
		finalAnnotations[infoFile+":"+relativePath] = Annotation{
			Path:       relativePath,
			Annotation: winner.Annotation,
			InfoFile:   infoFile,
			LineNum:    1, // Will be reassigned when generating content
		}
	}

	return d.DistributeAnnotations(finalAnnotations)
}

// AddAnnotation returns file operations to add a single annotation
func (d *Distributor) AddAnnotation(targetPath, annotation, infoFile string, existingContent string) FileOperation {
	content := d.editor.AddAnnotation(existingContent, targetPath, annotation, infoFile)

	return FileOperation{
		Type:     OpUpdate,
		FilePath: infoFile,
		Content:  content,
	}
}

// RemoveAnnotation returns file operations to remove an annotation
func (d *Distributor) RemoveAnnotation(annotation Annotation, existingContent string) FileOperation {
	content := d.editor.RemoveAnnotation(existingContent, annotation.LineNum)

	if strings.TrimSpace(content) == "" {
		// If file would be empty, delete it
		return FileOperation{
			Type:     OpDelete,
			FilePath: annotation.InfoFile,
		}
	}

	return FileOperation{
		Type:     OpUpdate,
		FilePath: annotation.InfoFile,
		Content:  content,
	}
}

// UpdateAnnotation returns file operations to update an annotation
func (d *Distributor) UpdateAnnotation(annotation Annotation, newAnnotationText, existingContent string) FileOperation {
	// Reconstruct target path from annotation
	infoDir := filepath.Dir(annotation.InfoFile)
	targetPath := filepath.Join(infoDir, annotation.Path)

	content := d.editor.UpdateAnnotation(existingContent, annotation.LineNum, targetPath, newAnnotationText, annotation.InfoFile)

	return FileOperation{
		Type:     OpUpdate,
		FilePath: annotation.InfoFile,
		Content:  content,
	}
}

// resolveConflict determines the winning annotation when multiple conflict
func (d *Distributor) resolveConflict(annotations []Annotation, strategy DistributionStrategy) Annotation {
	if len(annotations) == 1 {
		return annotations[0]
	}

	// Use the same conflict resolution as merger
	sort.Slice(annotations, func(i, j int) bool {
		dirI := filepath.Dir(annotations[i].InfoFile)
		dirJ := filepath.Dir(annotations[j].InfoFile)

		depthI := pathDepth(dirI)
		depthJ := pathDepth(dirJ)

		if depthI != depthJ {
			return depthI > depthJ // Deeper path wins
		}

		if dirI != dirJ {
			return dirI < dirJ
		}

		return annotations[i].LineNum < annotations[j].LineNum
	})

	return annotations[0]
}

// determineInfoFile determines the appropriate .info file for a target path
func (d *Distributor) determineInfoFile(targetPath string, strategy DistributionStrategy) string {
	switch strategy {
	case DistributeByDepth:
		// Place in the same directory as the target
		dir := filepath.Dir(targetPath)
		if dir == "" || dir == "." {
			return ".info"
		}
		return filepath.Join(dir, ".info")
	case DistributeByProximity:
		// Same as DistributeByDepth for now
		dir := filepath.Dir(targetPath)
		if dir == "" || dir == "." {
			return ".info"
		}
		return filepath.Join(dir, ".info")
	case DistributeConsolidate:
		// Place everything in root .info
		return ".info"
	default:
		// Default to same directory
		dir := filepath.Dir(targetPath)
		if dir == "" || dir == "." {
			return ".info"
		}
		return filepath.Join(dir, ".info")
	}
}

// DistributionStrategy defines how annotations should be distributed
type DistributionStrategy int

const (
	// DistributeByDepth places annotations in deepest valid .info file
	DistributeByDepth DistributionStrategy = iota
	// DistributeByProximity places annotations in nearest .info file to target
	DistributeByProximity
	// DistributeConsolidate places all annotations in root .info file
	DistributeConsolidate
)
