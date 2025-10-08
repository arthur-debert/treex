package info

import (
	"bufio"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/afero"
)

// Annotation represents a single entry in an .info file.
type Annotation struct {
	Path       string
	Annotation string
	InfoFile   string // The path to the .info file this annotation came from.
	LineNum    int
}

// Parse reads an .info file from an io.Reader and returns a list of annotations.
func Parse(reader io.Reader, infoFilePath string) ([]Annotation, error) {
	var annotations []Annotation
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	parsedPaths := make(map[string]bool)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var path, annotation string
		var pathEnd = -1
		var isEscaped = false

		// Find the first unescaped whitespace to split path and annotation.
		for i, r := range line {
			if unicode.IsSpace(r) && !isEscaped {
				pathEnd = i
				break
			}
			if r == '\\' && !isEscaped {
				isEscaped = true
			} else {
				isEscaped = false
			}
		}

		if pathEnd == -1 {
			continue // Line has no space separator, so no annotation.
		}

		path = line[:pathEnd]
		annotation = strings.TrimSpace(line[pathEnd+1:])

		if annotation == "" {
			continue // No annotation content.
		}

		// Unescape spaces in the path.
		path = strings.ReplaceAll(path, "\\ ", " ")

		// Per spec, first entry for a path in a file wins.
		if parsedPaths[path] {
			continue
		}
		parsedPaths[path] = true

		annotations = append(annotations, Annotation{
			Path:       path,
			Annotation: annotation,
			InfoFile:   infoFilePath,
			LineNum:    lineNum,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return annotations, nil
}

// Collector manages the collection and merging of annotations.
type Collector struct{}

// NewCollector creates a new annotation collector.
func NewCollector() *Collector {
	return &Collector{}
}

// CollectAnnotations walks the filesystem from root, finds all .info files,
// parses them, and returns a map of path to the winning annotation.
func (c *Collector) CollectAnnotations(fsys afero.Fs, root string) (map[string]Annotation, error) {
	var allAnnotations []Annotation

	err := afero.Walk(fsys, root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == ".info" {
			file, err := fsys.Open(path)
			if err != nil {
				// TODO: log warning and continue
				return nil
			}
			defer func() {
				_ = file.Close() // Error is intentionally ignored
			}()

			annotations, err := Parse(file, path)
			if err != nil {
				// TODO: log warning and continue
				return nil
			}
			allAnnotations = append(allAnnotations, annotations...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return c.merge(allAnnotations), nil
}

func (c *Collector) merge(annotations []Annotation) map[string]Annotation {
	contenders := make(map[string][]Annotation)
	for _, ann := range annotations {
		infoDir := filepath.Dir(ann.InfoFile)
		targetPath := filepath.Join(infoDir, ann.Path)

		// Normalize paths to handle "." and other relative parts.
		targetPath = filepath.Clean(targetPath)

		// Rule: .info files can't annotate their ancestors.
		rel, err := filepath.Rel(targetPath, infoDir)
		if err == nil && !strings.HasPrefix(rel, "..") && rel != "." {
			// targetPath is an ancestor of infoDir. Invalid.
			// TODO: log warning
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
			depthI := len(strings.Split(dirI, string(filepath.Separator)))
			depthJ := len(strings.Split(dirJ, string(filepath.Separator)))

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
