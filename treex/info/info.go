package info

import (
	"bufio"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/afero"
)

// Logger interface for warning reporting during info processing
type Logger interface {
	Printf(format string, v ...interface{})
}

// Annotation represents a single entry in an .info file.
type Annotation struct {
	Path       string
	Annotation string
	InfoFile   string // The path to the .info file this annotation came from.
	LineNum    int
}

// Parse reads an .info file from an io.Reader and returns a list of annotations.
func Parse(reader io.Reader, infoFilePath string) ([]Annotation, error) {
	return ParseWithLogger(reader, infoFilePath, nil)
}

// ParseWithLogger reads an .info file from an io.Reader and returns a list of annotations,
// using the provided logger for warnings.
func ParseWithLogger(reader io.Reader, infoFilePath string, logger Logger) ([]Annotation, error) {
	logf := func(format string, v ...interface{}) {
		if logger != nil {
			logger.Printf(format, v...)
		}
		// If no logger, silently ignore warnings during parsing
	}

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
			logf("info: ignoring line %d in %q: no annotation found (missing space separator)", lineNum, infoFilePath)
			continue // Line has no space separator, so no annotation.
		}

		path = line[:pathEnd]
		annotation = strings.TrimSpace(line[pathEnd+1:])

		if annotation == "" {
			logf("info: ignoring line %d in %q: empty annotation for path %q", lineNum, infoFilePath, path)
			continue // No annotation content.
		}

		// Unescape spaces in the path.
		path = strings.ReplaceAll(path, "\\ ", " ")

		// Per spec, first entry for a path in a file wins.
		if parsedPaths[path] {
			logf("info: ignoring duplicate path %q at line %d in %q (first occurrence wins)", path, lineNum, infoFilePath)
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
type Collector struct {
	logger Logger
}

// NewCollector creates a new annotation collector.
func NewCollector() *Collector {
	return &Collector{}
}

// NewCollectorWithLogger creates a new annotation collector with a custom logger.
func NewCollectorWithLogger(logger Logger) *Collector {
	return &Collector{
		logger: logger,
	}
}

// logf logs a warning message using the configured logger, or log.Printf if no logger is set
func (c *Collector) logf(format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
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
				c.logf("info: cannot open .info file %q: %v", path, err)
				return nil
			}
			defer func() {
				_ = file.Close() // Error is intentionally ignored
			}()

			annotations, err := ParseWithLogger(file, path, c.logger)
			if err != nil {
				c.logf("info: cannot parse .info file %q: %v", path, err)
				return nil
			}
			allAnnotations = append(allAnnotations, annotations...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return c.merge(fsys, allAnnotations), nil
}

func (c *Collector) merge(fsys afero.Fs, annotations []Annotation) map[string]Annotation {
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
			c.logf("info: invalid annotation in %q: cannot annotate ancestor path %q", ann.InfoFile, ann.Path)
			continue
		}

		// Validate that the target path exists in the filesystem
		if _, err := fsys.Stat(targetPath); err != nil {
			c.logf("info: invalid annotation in %q: path %q does not exist", ann.InfoFile, ann.Path)
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
