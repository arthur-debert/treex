package info

import (
	"strings"
	"unicode"

	"github.com/jwaldrip/treex/treex/logging"
)

// Parser handles parsing of .info file content into annotations
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses .info file content and returns a list of annotations.
func Parse(content, infoFilePath string) []Annotation {
	parser := NewParser()
	return parser.Parse(content, infoFilePath)
}

// Parse parses .info file content and returns a list of annotations,
// logging warnings for malformed content.
func (p *Parser) Parse(content, infoFilePath string) []Annotation {

	var annotations []Annotation
	lines := strings.Split(content, "\n")
	parsedPaths := make(map[string]bool)

	for lineNum, line := range lines {
		lineNum++ // Convert to 1-based line numbering
		line = strings.TrimSpace(line)

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
			logging.Warn().Int("line", lineNum).Str("file", infoFilePath).Msg("ignoring line: no annotation found (missing space separator)")
			continue // Line has no space separator, so no annotation.
		}

		path = line[:pathEnd]
		annotation = strings.TrimSpace(line[pathEnd+1:])

		if annotation == "" {
			logging.Warn().Int("line", lineNum).Str("file", infoFilePath).Str("path", path).Msg("ignoring line: empty annotation for path")
			continue // No annotation content.
		}

		// Unescape spaces in the path.
		path = strings.ReplaceAll(path, "\\ ", " ")

		// Per spec, first entry for a path in a file wins.
		if parsedPaths[path] {
			logging.Warn().Str("path", path).Int("line", lineNum).Str("file", infoFilePath).Msg("ignoring duplicate path (first occurrence wins)")
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

	return annotations
}

// ParseLine parses a single line and returns path, annotation, and success flag
func (p *Parser) ParseLine(line string) (string, string, bool) {
	var path, annotation string
	var pathEnd = -1
	var isEscaped = false

	// Find the first unescaped whitespace to split path and annotation
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
		return "", "", false // No space separator found
	}

	path = line[:pathEnd]
	annotation = strings.TrimSpace(line[pathEnd+1:])

	if annotation == "" {
		return path, "", false // Empty annotation
	}

	// Unescape spaces in the path
	path = strings.ReplaceAll(path, "\\ ", " ")

	return path, annotation, true
}
