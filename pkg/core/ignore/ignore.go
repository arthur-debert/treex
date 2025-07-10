package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// IgnoreMatcher handles .gitignore-style pattern matching
type IgnoreMatcher struct {
	patterns []ignorePattern
}

type ignorePattern struct {
	pattern   string
	regex     *regexp.Regexp
	isNegated bool
	isDir     bool
}

// NewIgnoreMatcher creates a new ignore matcher from a file
func NewIgnoreMatcher(ignoreFilePath string) (*IgnoreMatcher, error) {
	matcher := &IgnoreMatcher{}

	// If ignore file doesn't exist, return empty matcher (no filtering)
	if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
		return matcher, nil
	}

	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern, err := parseIgnorePattern(line)
		if err != nil {
			continue // Skip invalid patterns
		}

		matcher.patterns = append(matcher.patterns, pattern)
	}

	return matcher, scanner.Err()
}

// ShouldIgnore returns true if the given path should be ignored
func (m *IgnoreMatcher) ShouldIgnore(relativePath string, isDir bool) bool {
	// Normalize path separators
	relativePath = filepath.ToSlash(relativePath)

	ignored := false

	// Check each pattern
	for _, pattern := range m.patterns {
		matched := false
		if pattern.regex != nil {
			matched = pattern.regex.MatchString(relativePath)
		}

		if matched {
			if pattern.isNegated {
				ignored = false // Negated pattern un-ignores the path
			} else {
				ignored = true // Normal pattern ignores the path
			}
		}
	}

	return ignored
}

// parseIgnorePattern converts a gitignore pattern to a regex
func parseIgnorePattern(pattern string) (ignorePattern, error) {
	original := pattern
	isNegated := false
	isDir := false

	// Handle negation (!)
	if strings.HasPrefix(pattern, "!") {
		isNegated = true
		pattern = pattern[1:]
	}

	// Handle directory-only patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		isDir = true
		pattern = strings.TrimSuffix(pattern, "/")
	}

	// Handle leading slash (absolute path from root)
	var regexPattern string
	if strings.HasPrefix(pattern, "/") {
		regexPattern = "^" + regexp.QuoteMeta(pattern[1:])
	} else {
		// Pattern can match anywhere in the path
		regexPattern = regexp.QuoteMeta(pattern)
	}

	// Convert gitignore wildcards to regex
	regexPattern = strings.ReplaceAll(regexPattern, "\\*\\*", ".*") // ** matches any path
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", "[^/]*") // * matches anything except /
	regexPattern = strings.ReplaceAll(regexPattern, "\\?", "[^/]")  // ? matches single char except /

	// If pattern doesn't start with ^, it can match at any directory level
	if !strings.HasPrefix(regexPattern, "^") {
		regexPattern = "(^|.*/)(" + regexPattern + ")"
	}

	// Add appropriate endings based on pattern type
	if isDir {
		// Directory patterns should match the directory and everything inside it
		regexPattern += "(|/.*)$"
	} else {
		// File patterns should match the file exactly or as part of a path
		regexPattern += "(/.*)?$"
	}

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return ignorePattern{}, err
	}

	return ignorePattern{
		pattern:   original,
		regex:     regex,
		isNegated: isNegated,
		isDir:     isDir,
	}, nil
}
