package infofile

import (
	"strings"
	"unicode"
)

// Parser handles parsing of .info file content into annotations
type Parser struct{}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{}
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
