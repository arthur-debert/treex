// Package info provides InfoFileSet - a unified abstraction for InfoFile collection operations.
//
// # Filesystem Boundaries
//
// InfoFileSet maintains strict separation between I/O and business logic:
//
//   - InfoFileSetLoader: Handles reading from filesystem (I/O boundary)
//   - InfoFileSet: Pure in-memory operations with no filesystem access
//   - InfoFileSetWriter: Handles writing to filesystem (I/O boundary)
//
// This design enables:
//   - Pure function testing without filesystem dependencies
//   - Clear separation of concerns between I/O and business logic
//   - Easy mocking and dependency injection for testing
//
// # Usage Pattern
//
//	// Load from filesystem
//	loader := NewInfoFileSetLoader(fs)
//	infoFileSet, _ := loader.LoadFromPath(rootPath)
//
//	// Pure operations (no filesystem access)
//	gathered := infoFileSet.Gather()
//	cleanResult, cleanedSet := infoFileSet.Clean()
//	distributedSet := cleanedSet.Distribute()
//
//	// Write back to filesystem
//	writer := NewInfoFileSetWriter(fs)
//	writer.WriteInfoFileSet(distributedSet)
//
// # Empty File Handling
//
// InfoFiles with no annotations are considered "empty" and will be deleted
// from the filesystem during serialization via InfoFileSetWriter.
package info

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// InfoFileSet represents a collection of InfoFiles with unified operations.
// All operations are pure functions that work in-memory without filesystem access.
type InfoFileSet struct {
	files      []*InfoFile
	pathExists func(string) bool // Injected for path validation
}

// NewInfoFileSet creates a new InfoFileSet from a collection of InfoFiles
func NewInfoFileSet(files []*InfoFile, pathExists func(string) bool) *InfoFileSet {
	return &InfoFileSet{
		files:      files,
		pathExists: pathExists,
	}
}

// EmptyInfoFileSet creates an empty InfoFileSet
func EmptyInfoFileSet(pathExists func(string) bool) *InfoFileSet {
	return NewInfoFileSet([]*InfoFile{}, pathExists)
}

// ============================================================================
// CORE COLLECTION OPERATIONS
// ============================================================================

// Gather extracts and merges all annotations from the InfoFiles with conflict resolution.
// Returns a map of target path to winning annotation.
// Pure function - no filesystem access.
func (set *InfoFileSet) Gather() map[string]Annotation {
	allAnnotations := set.getAllAnnotations()
	return set.mergeAnnotations(allAnnotations)
}

// Distribute redistributes annotations to their optimal InfoFiles based on path proximity.
// Returns a new InfoFileSet with annotations moved to their closest InfoFiles.
// Pure function - no filesystem access.
func (set *InfoFileSet) Distribute() *InfoFileSet {
	// Create map of all annotations with their current distances
	type annotationWithDistance struct {
		annotation   *Annotation
		distance     int
		sourceFile   *InfoFile
		targetFile   *InfoFile
		originalPath string // Original path before transformation
	}

	var annotationsToMove []annotationWithDistance

	// For each annotation, find the best .info file based on resolved target path
	for _, infoFile := range set.files {
		for _, annotation := range infoFile.GetAllAnnotations() {
			// Calculate the resolved target path
			infoDir := filepath.Dir(infoFile.Path)
			resolvedTargetPath := filepath.Join(infoDir, annotation.Path)
			resolvedTargetPath = filepath.Clean(resolvedTargetPath)

			bestFile := infoFile
			bestDistance := 1000 // Large initial distance
			bestRelativePath := annotation.Path

			// Find the best .info file for this resolved target
			for _, candidateFile := range set.files {
				candidateDir := filepath.Dir(candidateFile.Path)

				// Calculate what the relative path would be from candidate to target
				relativePath, err := filepath.Rel(candidateDir, resolvedTargetPath)
				if err != nil {
					continue // Can't calculate relative path
				}

				// Calculate distance (levels between candidate dir and target)
				targetDir := filepath.Dir(resolvedTargetPath)
				distance := calculatePathDistance(candidateDir, targetDir)

				// Prefer shorter distances and paths that don't go outside the candidate's scope
				// In case of tie, prefer lexicographically first directory name
				if distance >= 0 && (distance < bestDistance ||
					(distance == bestDistance && candidateDir < filepath.Dir(bestFile.Path))) {
					bestFile = candidateFile
					bestDistance = distance
					bestRelativePath = relativePath
				}
			}

			// If we found a better file, mark for moving
			if bestFile != infoFile {
				// Create a new annotation with the correct relative path for the target file
				newAnnotation := Annotation{
					Path:       bestRelativePath,
					Annotation: annotation.Annotation,
					InfoFile:   bestFile.Path,
					LineNum:    annotation.LineNum, // Will be updated when added
				}

				annotationsToMove = append(annotationsToMove, annotationWithDistance{
					annotation:   &newAnnotation,
					distance:     bestDistance,
					sourceFile:   infoFile,
					targetFile:   bestFile,
					originalPath: annotation.Path, // Store original path for comparison
				})
			}
		}
	}

	// Create new collection with moved annotations
	newFiles := make([]*InfoFile, len(set.files))
	for i, infoFile := range set.files {
		// Start with a copy of the original file structure
		newFiles[i] = &InfoFile{
			Path:        infoFile.Path,
			Lines:       make([]Line, len(infoFile.Lines)),
			annotations: make(map[string]*Annotation),
		}
		copy(newFiles[i].Lines, infoFile.Lines)

		// Copy annotations that aren't being moved
		for path, ann := range infoFile.annotations {
			shouldMove := false
			for _, move := range annotationsToMove {
				// Compare original annotation path before transformation
				if ann.Path == move.originalPath && move.sourceFile == infoFile {
					shouldMove = true
					break
				}
			}
			if !shouldMove {
				newFiles[i].annotations[path] = ann
			} else {
				// Remove from lines by marking as removed
				newFiles[i].RemoveAnnotationForPath(path)
			}
		}
	}

	// Add moved annotations to their target files
	for _, move := range annotationsToMove {
		for _, newFile := range newFiles {
			if newFile.Path == move.targetFile.Path {
				newFile.AddAnnotationForPath(move.annotation.Path, move.annotation.Annotation)
				break
			}
		}
	}

	return NewInfoFileSet(newFiles, set.pathExists).RemoveEmpty()
}

// Validate performs comprehensive validation across all InfoFiles.
// Returns validation results including cross-file conflicts.
// Pure function - only uses injected pathExists function for filesystem checks.
func (set *InfoFileSet) Validate() *ValidationResult {
	result := &ValidationResult{
		Issues:       make([]ValidationIssue, 0),
		ValidFiles:   make([]string, 0),
		InvalidFiles: make([]string, 0),
		Summary:      make(map[string]interface{}),
	}

	issuesByType := make(map[ValidationIssueType]int)
	issuesByFile := make(map[string]int)

	result.Summary["total_files"] = len(set.files)
	allAnnotations := make([]Annotation, 0)

	// Process each InfoFile
	for _, infoFile := range set.files {
		fileIssues := make([]ValidationIssue, 0)

		// Check malformed lines from InfoFile
		for _, line := range infoFile.Lines {
			if line.Type == LineTypeMalformed {
				issueType := IssueInvalidFormat
				if line.ParseError == "duplicate path (first occurrence wins)" {
					issueType = IssueDuplicatePath
				}

				fileIssues = append(fileIssues, ValidationIssue{
					Type:     issueType,
					InfoFile: infoFile.Path,
					LineNum:  line.LineNum,
					Path:     "", // No valid path for malformed lines
					Message:  line.ParseError,
				})
			}
		}

		// Check valid annotations for path existence and ancestor issues
		for _, annotation := range infoFile.GetAllAnnotations() {
			infoDir := filepath.Dir(infoFile.Path)
			targetPath := filepath.Join(infoDir, annotation.Path)
			targetPath = filepath.Clean(targetPath)

			// Check if path exists (only filesystem interaction via injected function)
			if set.pathExists != nil && !set.pathExists(targetPath) {
				fileIssues = append(fileIssues, ValidationIssue{
					Type:     IssuePathNotExists,
					InfoFile: infoFile.Path,
					LineNum:  annotation.LineNum,
					Path:     annotation.Path,
					Message:  fmt.Sprintf("path does not exist: %s", targetPath),
				})
			}

			// Check for ancestor path annotations
			rel, err := filepath.Rel(infoDir, targetPath)
			if err == nil && strings.HasPrefix(rel, "..") {
				fileIssues = append(fileIssues, ValidationIssue{
					Type:     IssueAncestorPath,
					InfoFile: infoFile.Path,
					LineNum:  annotation.LineNum,
					Path:     annotation.Path,
					Message:  "cannot annotate ancestor path",
				})
			}

			allAnnotations = append(allAnnotations, annotation)
		}

		// Classify file and update summary
		if len(fileIssues) == 0 {
			result.ValidFiles = append(result.ValidFiles, infoFile.Path)
		} else {
			result.InvalidFiles = append(result.InvalidFiles, infoFile.Path)
		}

		result.Issues = append(result.Issues, fileIssues...)
		issuesByFile[infoFile.Path] = len(fileIssues)
		for _, issue := range fileIssues {
			issuesByType[issue.Type]++
		}
	}

	// Check for cross-file conflicts
	crossFileIssues := set.findCrossFileConflicts(allAnnotations)
	result.Issues = append(result.Issues, crossFileIssues...)
	for _, issue := range crossFileIssues {
		issuesByType[issue.Type]++
		issuesByFile[issue.InfoFile]++

		// Reclassify files with cross-file issues as invalid
		for i, validFile := range result.ValidFiles {
			if validFile == issue.InfoFile {
				result.ValidFiles = append(result.ValidFiles[:i], result.ValidFiles[i+1:]...)
				result.InvalidFiles = append(result.InvalidFiles, issue.InfoFile)
				break
			}
		}
	}

	result.Summary["total_issues"] = len(result.Issues)
	result.Summary["issues_by_type"] = issuesByType
	result.Summary["issues_by_file"] = issuesByFile

	return result
}

// Clean removes problematic annotations based on validation results.
// Returns CleanResult and a new InfoFileSet with issues resolved.
// Pure function - no filesystem access.
func (set *InfoFileSet) Clean() (*CleanResult, *InfoFileSet) {
	validationResult := set.Validate()

	result := &CleanResult{
		RemovedAnnotations: make([]Annotation, 0),
		UpdatedFiles:       make([]string, 0),
		Summary:            CleanSummary{},
	}

	// Group issues by file
	fileIssues := make(map[string][]ValidationIssue)
	for _, issue := range validationResult.Issues {
		fileIssues[issue.InfoFile] = append(fileIssues[issue.InfoFile], issue)
	}

	// Process each file with issues
	newFiles := make([]*InfoFile, 0, len(set.files))
	for _, infoFile := range set.files {
		issues, hasIssues := fileIssues[infoFile.Path]
		if hasIssues {
			cleanedFile := set.cleanSingleFile(infoFile, issues, result)
			newFiles = append(newFiles, cleanedFile)
			result.UpdatedFiles = append(result.UpdatedFiles, infoFile.Path)
			result.Summary.FilesModified++
		} else {
			// No issues, keep file as-is
			newFiles = append(newFiles, infoFile)
		}
	}

	return result, NewInfoFileSet(newFiles, set.pathExists)
}

// ============================================================================
// FLUENT INTERFACE FOR METHOD CHAINING
// ============================================================================

// Filter returns a new InfoFileSet containing only InfoFiles matching the predicate.
// Pure function - no filesystem access.
func (set *InfoFileSet) Filter(predicate func(*InfoFile) bool) *InfoFileSet {
	filtered := make([]*InfoFile, 0)
	for _, file := range set.files {
		if predicate(file) {
			filtered = append(filtered, file)
		}
	}
	return NewInfoFileSet(filtered, set.pathExists)
}

// RemoveEmpty returns a new InfoFileSet with empty InfoFiles removed.
// Empty InfoFiles will be deleted during serialization.
// Pure function - no filesystem access.
func (set *InfoFileSet) RemoveEmpty() *InfoFileSet {
	return set.Filter(func(file *InfoFile) bool {
		return !file.IsEmpty()
	})
}

// WithPathValidator returns a new InfoFileSet with a different path existence function.
// Useful for testing or changing validation behavior.
func (set *InfoFileSet) WithPathValidator(pathExists func(string) bool) *InfoFileSet {
	return NewInfoFileSet(set.files, pathExists)
}

// ============================================================================
// QUERY AND AGGREGATION OPERATIONS
// ============================================================================

// GetFiles returns the underlying InfoFile slice.
// Returns a copy to prevent external modification.
func (set *InfoFileSet) GetFiles() []*InfoFile {
	result := make([]*InfoFile, len(set.files))
	copy(result, set.files)
	return result
}

// Count returns the number of InfoFiles in the set.
func (set *InfoFileSet) Count() int {
	return len(set.files)
}

// GetAllAnnotations extracts all annotations from all InfoFiles.
// Pure function - no filesystem access.
func (set *InfoFileSet) GetAllAnnotations() []Annotation {
	return set.getAllAnnotations()
}

// GetAnnotationCount returns the total number of annotations across all InfoFiles.
func (set *InfoFileSet) GetAnnotationCount() int {
	count := 0
	for _, file := range set.files {
		count += len(file.GetAllAnnotations())
	}
	return count
}

// HasConflicts checks if there are any cross-file annotation conflicts.
// Pure function - no filesystem access.
func (set *InfoFileSet) HasConflicts() bool {
	allAnnotations := set.getAllAnnotations()
	conflicts := set.findCrossFileConflicts(allAnnotations)
	return len(conflicts) > 0
}

// GetFilePaths returns all InfoFile paths in the set.
func (set *InfoFileSet) GetFilePaths() []string {
	paths := make([]string, len(set.files))
	for i, file := range set.files {
		paths[i] = file.Path
	}
	return paths
}

// GetEmptyFiles returns InfoFiles that have no annotations.
// These files will be deleted during serialization.
func (set *InfoFileSet) GetEmptyFiles() []*InfoFile {
	var empty []*InfoFile
	for _, file := range set.files {
		if file.IsEmpty() {
			empty = append(empty, file)
		}
	}
	return empty
}

// ============================================================================
// INTERNAL HELPER METHODS - Pure functions, no filesystem access
// ============================================================================

func (set *InfoFileSet) getAllAnnotations() []Annotation {
	var allAnnotations []Annotation
	for _, infoFile := range set.files {
		annotations := infoFile.GetAllAnnotations()
		allAnnotations = append(allAnnotations, annotations...)
	}
	return allAnnotations
}

func (set *InfoFileSet) mergeAnnotations(annotations []Annotation) map[string]Annotation {
	result := make(map[string]Annotation)

	// Group annotations by resolved target path, filtering out invalid ones
	pathGroups := make(map[string][]Annotation)
	for _, annotation := range annotations {
		infoDir := filepath.Dir(annotation.InfoFile)
		targetPath := filepath.Join(infoDir, annotation.Path)
		targetPath = filepath.Clean(targetPath)

		// Rule: .info files can't annotate their ancestors.
		// Check if targetPath is an ancestor of infoDir
		rel, err := filepath.Rel(targetPath, infoDir)

		// Two cases indicate ancestor relationship:
		// 1. Rel succeeds and infoDir is contained within targetPath (rel doesn't start with "..")
		// 2. Rel fails because targetPath is above infoDir in the hierarchy
		if (err == nil && !strings.HasPrefix(rel, "..") && rel != ".") ||
			(err != nil && strings.Contains(err.Error(), "can't make")) {
			// Skip invalid annotation that tries to annotate ancestor path
			continue
		}

		// Validate that the target path exists (if pathExists function is available)
		if set.pathExists != nil && !set.pathExists(targetPath) {
			// Skip annotation for non-existent path
			continue
		}

		pathGroups[targetPath] = append(pathGroups[targetPath], annotation)
	}

	// Apply precedence rules for each path
	for targetPath, group := range pathGroups {
		winner := set.selectWinnerAnnotation(group)
		result[targetPath] = winner
	}

	return result
}

func (set *InfoFileSet) selectWinnerAnnotation(annotations []Annotation) Annotation {
	if len(annotations) == 1 {
		return annotations[0]
	}

	// Sort by precedence rules (matching documentation):
	// 1. Closest to target wins (smallest distance from .info dir to target)
	// 2. Lexicographical order of directory paths
	// 3. Line number within file
	sort.Slice(annotations, func(i, j int) bool {
		dirI := filepath.Dir(annotations[i].InfoFile)
		dirJ := filepath.Dir(annotations[j].InfoFile)

		// Calculate distances from .info directory to target
		// Distance 0: same directory, Distance 1: one level down, etc.
		targetI := filepath.Join(dirI, annotations[i].Path)
		targetJ := filepath.Join(dirJ, annotations[j].Path)
		targetDirI := filepath.Dir(filepath.Clean(targetI))
		targetDirJ := filepath.Dir(filepath.Clean(targetJ))

		distanceI := calculatePathDistance(dirI, targetDirI)
		distanceJ := calculatePathDistance(dirJ, targetDirJ)

		if distanceI != distanceJ {
			return distanceI < distanceJ // Closer wins (smaller distance)
		}

		// Rule: if distance is same, lexicographical order of .info file dir wins.
		if dirI != dirJ {
			return dirI < dirJ
		}

		// Rule: if same .info file, lower line number wins.
		return annotations[i].LineNum < annotations[j].LineNum
	})

	return annotations[0]
}

// calculatePathDistance calculates distance between info dir and target dir
func calculatePathDistance(infoDir, targetDir string) int {
	if infoDir == targetDir {
		return 0
	}

	rel, err := filepath.Rel(infoDir, targetDir)
	if err != nil {
		return 1000 // Large distance if can't calculate
	}

	if rel == "." {
		return 0
	}

	if strings.HasPrefix(rel, "..") {
		// Target is above info dir - count levels up
		return strings.Count(rel, "..")
	}

	// Target is below info dir - count levels down
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func (set *InfoFileSet) findCrossFileConflicts(annotations []Annotation) []ValidationIssue {
	pathToFiles := make(map[string][]Annotation)

	// Group annotations by resolved target path
	for _, ann := range annotations {
		infoDir := filepath.Dir(ann.InfoFile)
		targetPath := filepath.Join(infoDir, ann.Path)
		targetPath = filepath.Clean(targetPath)

		pathToFiles[targetPath] = append(pathToFiles[targetPath], ann)
	}

	var issues []ValidationIssue
	for _, anns := range pathToFiles {
		if len(anns) > 1 {
			// Multiple files annotate the same path - create issues for all but the first
			for i := 1; i < len(anns); i++ {
				issues = append(issues, ValidationIssue{
					Type:        IssueMultipleFiles,
					InfoFile:    anns[i].InfoFile,
					LineNum:     anns[i].LineNum,
					Path:        anns[i].Path,
					Message:     fmt.Sprintf("path already annotated in %s", anns[0].InfoFile),
					RelatedFile: anns[0].InfoFile,
				})
			}
		}
	}

	return issues
}

func (set *InfoFileSet) cleanSingleFile(infoFile *InfoFile, issues []ValidationIssue, result *CleanResult) *InfoFile {
	// Create a copy of the InfoFile to avoid modifying the original
	cleanedFile := infoFile.Clone()

	// Process issues by type
	for _, issue := range issues {
		switch issue.Type {
		case IssuePathNotExists, IssueInvalidFormat, IssueDuplicatePath, IssueAncestorPath:
			// For valid annotations with path issues, remove them
			if issue.Path != "" && cleanedFile.HasAnnotationForPath(issue.Path) {
				if cleanedFile.RemoveAnnotationForPath(issue.Path) {
					// Track what we're removing
					switch issue.Type {
					case IssuePathNotExists:
						result.Summary.InvalidPathsRemoved++
					case IssueDuplicatePath:
						result.Summary.DuplicatesRemoved++
					}

					// Create annotation for removed item
					result.RemovedAnnotations = append(result.RemovedAnnotations, Annotation{
						Path:       issue.Path,
						InfoFile:   issue.InfoFile,
						LineNum:    issue.LineNum,
						Annotation: fmt.Sprintf("(removed: %s)", issue.Type),
					})
				}
			}

			// For malformed lines, mark them as removed
			if issue.Type == IssueInvalidFormat || issue.Type == IssueDuplicatePath {
				for i := range cleanedFile.Lines {
					if cleanedFile.Lines[i].LineNum == issue.LineNum && cleanedFile.Lines[i].Type == LineTypeMalformed {
						cleanedFile.Lines[i].ParseError = "removed"
						break
					}
				}
			}
		case IssueMultipleFiles:
			// For cross-file conflicts, remove the annotation from the later file
			if cleanedFile.HasAnnotationForPath(issue.Path) {
				if cleanedFile.RemoveAnnotationForPath(issue.Path) {
					result.RemovedAnnotations = append(result.RemovedAnnotations, Annotation{
						Path:       issue.Path,
						InfoFile:   issue.InfoFile,
						LineNum:    issue.LineNum,
						Annotation: fmt.Sprintf("(removed: %s)", issue.Type),
					})
				}
			}
		}
	}

	return cleanedFile
}
