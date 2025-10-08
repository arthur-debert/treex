package info

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/jwaldrip/treex/treex/logging"
)

// Merger handles merging of annotations according to precedence rules
type Merger struct{}

// NewMerger creates a new merger instance
func NewMerger() *Merger {
	return &Merger{}
}

// MergeAnnotations takes a list of annotations and merges them according to precedence rules.
// Returns a map of resolved target path to the winning annotation.
func (m *Merger) MergeAnnotations(annotations []Annotation, pathExists func(string) bool) map[string]Annotation {
	contenders := make(map[string][]Annotation)

	for _, ann := range annotations {
		infoDir := filepath.Dir(ann.InfoFile)
		targetPath := filepath.Join(infoDir, ann.Path)

		// Normalize paths to handle "." and other relative parts.
		targetPath = filepath.Clean(targetPath)

		// Rule: .info files can't annotate their ancestors.
		// Check if targetPath is an ancestor of infoDir
		rel, err := filepath.Rel(targetPath, infoDir)

		// Two cases indicate ancestor relationship:
		// 1. Rel succeeds and infoDir is contained within targetPath (rel doesn't start with "..")
		// 2. Rel fails because targetPath is above infoDir in the hierarchy
		if (err == nil && !strings.HasPrefix(rel, "..") && rel != ".") ||
			(err != nil && strings.Contains(err.Error(), "can't make")) {
			logging.Warn().Str("info_file", ann.InfoFile).Str("path", ann.Path).Msg("invalid annotation: cannot annotate ancestor path")
			continue
		}

		// Validate that the target path exists in the filesystem
		if !pathExists(targetPath) {
			logging.Warn().Str("info_file", ann.InfoFile).Str("path", ann.Path).Msg("invalid annotation: path does not exist")
			continue
		}

		contenders[targetPath] = append(contenders[targetPath], ann)
	}

	winner := make(map[string]Annotation)
	for path, anns := range contenders {
		sort.Slice(anns, func(i, j int) bool {
			dirI := filepath.Dir(anns[i].InfoFile)
			dirJ := filepath.Dir(anns[j].InfoFile)

			// Rule: closest (deepest) .info file wins.
			// Calculate depth correctly, handling "." as root (depth 0)
			depthI := pathDepth(dirI)
			depthJ := pathDepth(dirJ)

			if depthI != depthJ {
				return depthI > depthJ // Deeper path wins
			}

			// Rule: if distance is same, lexicographical order of .info file dir wins.
			if dirI != dirJ {
				return dirI < dirJ
			}

			// Rule: if same .info file, lower line number wins.
			return anns[i].LineNum < anns[j].LineNum
		})
		winner[path] = anns[0]
	}

	return winner
}

// pathDepth calculates the depth of a directory path, with "." being depth 0
func pathDepth(dir string) int {
	if dir == "." {
		return 0
	}
	// Clean the path to handle any redundant separators
	clean := filepath.Clean(dir)
	if clean == "." {
		return 0
	}
	// Count the separators
	return strings.Count(clean, string(filepath.Separator)) + 1
}
