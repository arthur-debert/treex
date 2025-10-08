package info

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// Parser handles parsing of .info file content into annotations
type Parser struct {
	logger Logger
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{}
}

// NewParserWithLogger creates a new parser instance with a custom logger
func NewParserWithLogger(logger Logger) *Parser {
	return &Parser{logger: logger}
}

// Parse reads an .info file from an io.Reader and returns a list of annotations.
func Parse(reader io.Reader, infoFilePath string) ([]Annotation, error) {
	parser := NewParser()
	return parser.ParseWithLogger(reader, infoFilePath, nil)
}

// ParseWithLogger reads an .info file from an io.Reader and returns a list of annotations,
// using the provided logger for warnings.
func ParseWithLogger(reader io.Reader, infoFilePath string, logger Logger) ([]Annotation, error) {
	parser := NewParserWithLogger(logger)
	return parser.ParseWithLogger(reader, infoFilePath, logger)
}

// ParseWithLogger reads an .info file from an io.Reader and returns a list of annotations,
// using the parser's configured logger for warnings.
func (p *Parser) ParseWithLogger(reader io.Reader, infoFilePath string, logger Logger) ([]Annotation, error) {
	// Use provided logger or fall back to parser's logger
	logf := func(format string, v ...interface{}) {
		if logger != nil {
			logger.Printf(format, v...)
		} else if p.logger != nil {
			p.logger.Printf(format, v...)
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
