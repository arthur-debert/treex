package info

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/adebert/treex/pkg/core/types"
)

// ParseInfoFile parses a file in .info format without path validation.
func ParseInfoFile(filePath string) (map[string]*types.Annotation, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return parseInfoFromReader(file)
}

// ParseInfoFromReader parses info format from an io.Reader, useful for stdin
func ParseInfoFromReader(reader io.Reader) (map[string]*types.Annotation, []string, error) {
	return parseInfoFromReader(reader)
}

// parseInfoFromReader is the common implementation for parsing from a reader
func parseInfoFromReader(reader io.Reader) (map[string]*types.Annotation, []string, error) {
	annotations := make(map[string]*types.Annotation)
	var warnings []string
	scanner := bufio.NewScanner(reader)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		colonIdx := strings.Index(line, ":")
		var path, notes string

		if colonIdx != -1 {
			path = strings.TrimSpace(line[:colonIdx])
			notes = strings.TrimSpace(line[colonIdx+1:])
		} else {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				warnings = append(warnings, fmt.Sprintf("Line %d: Invalid format (missing annotation): %q", lineNum, line))
				continue
			}
			path = fields[0]
			pathEnd := strings.Index(line, path) + len(path)
			notes = strings.TrimSpace(line[pathEnd:])
		}

		if path == "" {
			warnings = append(warnings, fmt.Sprintf("Line %d: Empty path in annotation", lineNum))
			continue
		}

		if notes == "" {
			warnings = append(warnings, fmt.Sprintf("Line %d: Empty notes for path %q", lineNum, path))
			continue
		}

		annotations[path] = &types.Annotation{
			Path:  path,
			Notes: notes,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading input: %w", err)
	}

	return annotations, warnings, nil
}
