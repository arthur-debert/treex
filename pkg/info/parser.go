package info

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Annotation represents a single file/directory annotation
type Annotation struct {
	Path        string
	Title       string // First line of description (if it ends with newline)
	Description string // Full description
}

// Parser handles parsing .info files
type Parser struct {
	annotations map[string]*Annotation
}

// NewParser creates a new info file parser
func NewParser() *Parser {
	return &Parser{
		annotations: make(map[string]*Annotation),
	}
}

// ParseFile parses a .info file and returns a map of path -> annotation
func (p *Parser) ParseFile(infoFilePath string) (map[string]*Annotation, error) {
	file, err := os.Open(infoFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No .info file is not an error, just return empty map
			return make(map[string]*Annotation), nil
		}
		return nil, fmt.Errorf("failed to open .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	scanner := bufio.NewScanner(file)
	var lines []string

	// Read all lines first
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .info file: %w", err)
	}

	// Parse the lines
	i := 0
	for i < len(lines) {
		// Skip empty lines at the beginning
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}

		if i >= len(lines) {
			break
		}

		// This should be a path and title on the same line separated by whitespace
		pathLine := strings.TrimSpace(lines[i])
		i++

		// Split on whitespace to separate path and title
		parts := strings.Fields(pathLine)
		if len(parts) < 2 {
			// Not valid format - must have at least path and title on same line
			// Skip this line and continue
			continue
		}

		// First part is path, rest is title
		path := parts[0]
		titleFromPathLine := strings.Join(parts[1:], " ")

		// Collect description lines until we find a blank line followed by a non-empty line
		// that looks like it could be a new path+title line
		var descriptionLines []string

		for i < len(lines) {
			line := lines[i]

			// If this is an empty line, we need to look ahead more carefully
			if strings.TrimSpace(line) == "" {
				// Look ahead to find the next non-empty line
				nextNonEmptyIdx := i + 1
				for nextNonEmptyIdx < len(lines) && strings.TrimSpace(lines[nextNonEmptyIdx]) == "" {
					nextNonEmptyIdx++
				}

				if nextNonEmptyIdx >= len(lines) {
					// No more non-empty lines, include this empty line and we're done
					descriptionLines = append(descriptionLines, line)
					i++
					break
				}

				// We found a non-empty line. Now we need to decide if it's a new entry or part of description
				nextLine := lines[nextNonEmptyIdx]

				// Check if the next line looks like a new entry (must contain at least two words)
				if isLikelyNewEntry(nextLine) {
					// This looks like a new entry, stop collecting description here
					break
				} else {
					// This empty line is likely part of the description formatting
					if len(descriptionLines) == 0 {
						// No additional description - stop here
						break
					} else {
						// Include the empty line as part of description formatting
						descriptionLines = append(descriptionLines, line)
						i++
						continue
					}
				}
			} else {
				// Non-empty line, add to description
				descriptionLines = append(descriptionLines, line)
				i++
			}
		}

		// Save this annotation
		p.saveAnnotation(path, titleFromPathLine, descriptionLines)
	}

	return p.annotations, nil
}

// isLikelyNewEntry determines if a line looks like the start of a new annotation entry
// In compact format, a new entry must have at least two words (path and title)
func isLikelyNewEntry(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	// Check if this line contains multiple words (must be "path title" format)
	parts := strings.Fields(trimmed)
	return len(parts) >= 2
}

// saveAnnotation processes and saves an annotation with title from the path line
func (p *Parser) saveAnnotation(path string, title string, descriptionLines []string) {
	// Remove trailing empty lines from description
	for len(descriptionLines) > 0 && strings.TrimSpace(descriptionLines[len(descriptionLines)-1]) == "" {
		descriptionLines = descriptionLines[:len(descriptionLines)-1]
	}

	// Set up the annotation
	var fullDescription string

	if len(descriptionLines) > 0 {
		// Additional description lines after the path+title line
		fullDescription = strings.Join(descriptionLines, "\n")
	}

	// For multi-line descriptions, we need to ensure the title is included in the full description
	// if the title isn't already part of the description
	if len(descriptionLines) == 0 || (len(descriptionLines) > 0 && !strings.HasPrefix(fullDescription, title)) {
		if fullDescription == "" {
			fullDescription = title
		} else {
			// Prepend the title to the description for proper multi-line support
			fullDescription = title + "\n" + fullDescription
		}
	}

	annotation := &Annotation{
		Path:        path,
		Title:       title,
		Description: fullDescription,
	}

	p.annotations[path] = annotation
}

// GetAnnotation returns the annotation for a given path
func (p *Parser) GetAnnotation(path string) (*Annotation, bool) {
	annotation, exists := p.annotations[path]
	return annotation, exists
}

// GetAllAnnotations returns all parsed annotations
func (p *Parser) GetAllAnnotations() map[string]*Annotation {
	return p.annotations
}

// ParseDirectory looks for a .info file in the given directory and parses it
func ParseDirectory(dirPath string) (map[string]*Annotation, error) {
	infoPath := filepath.Join(dirPath, ".info")
	parser := NewParser()
	return parser.ParseFile(infoPath)
}

// ParseDirectoryTree recursively looks for .info files in the entire directory tree
// and merges all annotations with proper path resolution
func ParseDirectoryTree(rootPath string) (map[string]*Annotation, error) {
	allAnnotations := make(map[string]*Annotation)

	// Walk the directory tree
	err := filepath.Walk(rootPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't read instead of failing completely
			return nil
		}

		// Skip if not a directory
		if !info.IsDir() {
			return nil
		}

		// Look for .info file in this directory
		infoPath := filepath.Join(currentPath, ".info")
		if _, err := os.Stat(infoPath); os.IsNotExist(err) {
			// No .info file in this directory, continue
			return nil
		}

		// Parse the .info file with proper context
		annotations, err := parseFileWithContext(infoPath, rootPath, currentPath)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", infoPath, err)
		}

		// Merge annotations (later files override earlier ones if there are conflicts)
		for path, annotation := range annotations {
			allAnnotations[path] = annotation
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return allAnnotations, nil
}

// parseFileWithContext parses a .info file with proper path resolution
// rootPath: the root of the entire tree being analyzed
// contextDir: the directory containing this .info file
func parseFileWithContext(infoFilePath, rootPath, contextDir string) (map[string]*Annotation, error) {
	parser := NewParser()

	// Parse the file normally first
	annotations, err := parser.ParseFile(infoFilePath)
	if err != nil {
		return nil, err
	}

	// Now resolve paths relative to the context directory
	resolvedAnnotations := make(map[string]*Annotation)

	for localPath, annotation := range annotations {
		// Validate that the path doesn't try to escape the current directory
		if strings.Contains(localPath, "..") {
			continue // Skip paths that try to go up directories
		}

		// Create absolute path for this annotation
		fullPath := filepath.Join(contextDir, localPath)

		// Convert to path relative to root
		relativePath, err := filepath.Rel(rootPath, fullPath)
		if err != nil {
			continue // Skip if we can't resolve the path
		}

		// Normalize path separators for consistency
		relativePath = filepath.ToSlash(relativePath)

		// Create new annotation with resolved path
		resolvedAnnotation := &Annotation{
			Path:        relativePath,
			Title:       annotation.Title,
			Description: annotation.Description,
		}

		resolvedAnnotations[relativePath] = resolvedAnnotation
	}

	return resolvedAnnotations, nil
}

// TreeEntry represents a parsed entry from a tree-like input file
type TreeEntry struct {
	Path        string
	Description string
	IsDir       bool
}

// GenerateInfoFromTree parses a tree-like input file and generates .info files
func GenerateInfoFromTree(inputFile string) error {
	entries, err := parseTreeFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse tree file: %w", err)
	}

	return generateInfoFromEntries(entries)
}

// GenerateInfoFromReader parses tree-like content from a reader and generates .info files
func GenerateInfoFromReader(reader io.Reader) error {
	entries, err := parseTreeReader(reader)
	if err != nil {
		return fmt.Errorf("failed to parse tree content: %w", err)
	}

	return generateInfoFromEntries(entries)
}

// generateInfoFromEntries is the common logic for generating .info files from entries
func generateInfoFromEntries(entries []TreeEntry) error {
	// Group entries by their parent directories
	infoFiles := make(map[string][]TreeEntry)

	for _, entry := range entries {
		// Determine the parent directory
		parentDir := filepath.Dir(entry.Path)
		if parentDir == "." {
			parentDir = ""
		}

		// Check if the path exists
		if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", entry.Path)
		}

		// Add to the appropriate .info file
		infoFiles[parentDir] = append(infoFiles[parentDir], entry)
	}

	// Generate .info files
	for dir, dirEntries := range infoFiles {
		if err := generateInfoFile(dir, dirEntries); err != nil {
			return fmt.Errorf("failed to generate .info file for directory %s: %w", dir, err)
		}
	}

	return nil
}

// parseTreeFile parses a tree-like input file and extracts paths and descriptions
func parseTreeFile(inputFile string) ([]TreeEntry, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	return parseTreeReader(file)
}

// parseTreeReader parses tree-like content from a reader and extracts paths and descriptions
func parseTreeReader(reader io.Reader) ([]TreeEntry, error) {
	var entries []TreeEntry
	scanner := bufio.NewScanner(reader)
	var pathStack []string // Stack to keep track of current path components

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the tree line
		entry, depth, err := parseTreeLine(line, pathStack)
		if err != nil {
			continue // Skip malformed lines
		}

		if entry != nil {
			// Handle path building
			// The first entry (usually no tree connectors) is the root
			// Entries with tree connectors (├── └──) at depth 0 are children of root
			// Entries with vertical connectors (│   └──) at depth 1+ are nested

			if len(pathStack) == 0 {
				// First entry is the root
				pathStack = append(pathStack, entry.Path)
			} else {
				// Adjust pathStack based on depth
				// depth 0 means direct child of root, depth 1 means nested, etc.
				targetLength := depth + 1 // +1 because root is at index 0

				if targetLength < len(pathStack) {
					// Going back up, truncate
					pathStack = pathStack[:targetLength]
				}

				// Add this entry at the appropriate level
				if targetLength == len(pathStack) {
					// Adding at same level or going deeper
					pathStack = append(pathStack, entry.Path)
				} else {
					// Replace the last entry at this level
					pathStack[targetLength-1] = entry.Path
					pathStack = pathStack[:targetLength]
				}
			}

			// Create the full path
			fullPath := strings.Join(pathStack, "/")
			entry.Path = fullPath

			entries = append(entries, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return entries, nil
}

// parseTreeLine parses a single line from the tree format
func parseTreeLine(line string, currentPath []string) (*TreeEntry, int, error) {
	depth := 0

	// Calculate depth based on indentation and tree characters
	runes := []rune(line)
	i := 0
	depth = 0

	// Count vertical connectors (│) to determine depth
	for i < len(runes) {
		char := runes[i]
		if char == ' ' || char == '\t' {
			i++
			continue
		} else if char == '│' {
			// Vertical connector - increment depth and skip
			depth++
			i++
			// Skip any following spaces
			for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
				i++
			}
		} else if char == '├' || char == '└' {
			// Horizontal connectors - this is the actual entry line
			i++
			// Skip connector characters (─)
			for i < len(runes) && runes[i] == '─' {
				i++
			}
			// Skip any following spaces
			for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
				i++
			}
			break
		} else {
			// No connectors, this is a root-level entry
			break
		}
	}

	// Extract the remaining content (path and description)
	if i >= len(runes) {
		return nil, depth, nil // No content, just connectors
	}

	content := strings.TrimSpace(string(runes[i:]))
	if content == "" {
		return nil, depth, nil
	}

	// Split path and description
	parts := strings.SplitN(content, " ", 2)
	if len(parts) < 1 {
		return nil, depth, fmt.Errorf("invalid line format")
	}

	path := strings.TrimSpace(parts[0])
	description := ""
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	// Remove trailing slash to normalize directory names
	isDir := strings.HasSuffix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		return nil, depth, fmt.Errorf("empty path")
	}

	return &TreeEntry{
		Path:        path,
		Description: description,
		IsDir:       isDir,
	}, depth, nil
}

// generateInfoFile creates a .info file in the specified directory
func generateInfoFile(dir string, entries []TreeEntry) error {
	// Determine the path for the .info file
	infoPath := ".info"
	if dir != "" {
		infoPath = filepath.Join(dir, ".info")

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create the .info file
	file, err := os.Create(infoPath)
	if err != nil {
		return fmt.Errorf("failed to create .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	// Write entries to the file
	for _, entry := range entries {
		// Get the relative path (just the filename/dirname)
		relativePath := filepath.Base(entry.Path)
		if entry.IsDir {
			relativePath += "/"
		}

		// Write the path and description in compact format
		if entry.Description != "" {
			if _, err := fmt.Fprintf(file, "%s %s\n", relativePath, entry.Description); err != nil {
				return fmt.Errorf("failed to write to .info file: %w", err)
			}
		} else {
			// If no description, write at least a placeholder
			if _, err := fmt.Fprintf(file, "%s No description\n", relativePath); err != nil {
				return fmt.Errorf("failed to write to .info file: %w", err)
			}
		}

		// Add a blank line between entries
		if _, err := fmt.Fprintf(file, "\n"); err != nil {
			return fmt.Errorf("failed to write to .info file: %w", err)
		}
	}

	return nil
}

// WriteInfoFile writes annotations to a .info file
func WriteInfoFile(filePath string, annotations map[string]*Annotation) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	// Write each annotation
	for path, annotation := range annotations {
		// Get the title to use on the same line as the path
		title := annotation.Title
		if title == "" {
			// If no title, extract from description or use a placeholder
			if annotation.Description != "" {
				// Use first line of description as title
				lines := strings.Split(annotation.Description, "\n")
				title = lines[0]
			} else {
				title = "No description"
			}
		}

		// Write the path and title
		if _, err := fmt.Fprintf(file, "%s %s\n", path, title); err != nil {
			return fmt.Errorf("failed to write path and title: %w", err)
		}

		// Check if there's additional description content beyond the title
		description := annotation.Description
		// If description has multiple lines, write the rest after the first line
		if strings.Contains(description, "\n") {
			lines := strings.Split(description, "\n")
			// Skip the first line if it's the same as the title
			startIdx := 0
			if len(lines) > 0 && lines[0] == title {
				startIdx = 1
			}

			// Write remaining description lines
			for i := startIdx; i < len(lines); i++ {
				if _, err := fmt.Fprintf(file, "%s\n", lines[i]); err != nil {
					return fmt.Errorf("failed to write description line: %w", err)
				}
			}
		}

		// Add blank line between entries
		if _, err := fmt.Fprintf(file, "\n"); err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}
	}

	return nil
}

// UpdateAction represents the action to take when updating an existing entry
type UpdateAction int

const (
	UpdateActionReplace UpdateAction = iota
	UpdateActionAppend
	UpdateActionAbort
)

// AddOrUpdateEntry adds or updates an entry in a .info file
func AddOrUpdateEntry(dirPath, entryPath, description string, action UpdateAction) error {
	infoFilePath := filepath.Join(dirPath, ".info")

	// Parse existing .info file if it exists
	annotations, err := ParseDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to parse existing .info file: %w", err)
	}

	// Check if entry already exists
	existingAnnotation, exists := annotations[entryPath]

	if exists {
		switch action {
		case UpdateActionReplace:
			// Create a new annotation with the new description
			title := description
			// If multi-line, use first line as title
			if strings.Contains(description, "\n") {
				lines := strings.Split(description, "\n")
				if len(lines) > 0 {
					title = lines[0]
				}
			}

			annotations[entryPath] = &Annotation{
				Path:        entryPath,
				Title:       title,
				Description: description,
			}
		case UpdateActionAppend:
			// Append to existing description
			var newDescription string
			if existingAnnotation.Description != "" {
				newDescription = existingAnnotation.Description + "\n" + description
			} else {
				newDescription = description
			}

			// Determine title - use first line of combined description
			title := newDescription
			if strings.Contains(newDescription, "\n") {
				lines := strings.Split(newDescription, "\n")
				if len(lines) > 0 {
					title = lines[0]
				}
			}

			annotations[entryPath] = &Annotation{
				Path:        entryPath,
				Title:       title,
				Description: newDescription,
			}
		case UpdateActionAbort:
			// Don't make any changes
			return nil
		}
	} else {
		// Create new annotation
		title := description
		// If multi-line, use first line as title
		if strings.Contains(description, "\n") {
			lines := strings.Split(description, "\n")
			if len(lines) > 0 {
				title = lines[0]
			}
		}

		// Add the new annotation
		annotations[entryPath] = &Annotation{
			Path:        entryPath,
			Title:       title,
			Description: description,
		}
	}

	// Write the updated .info file
	if err := WriteInfoFile(infoFilePath, annotations); err != nil {
		return fmt.Errorf("failed to write .info file: %w", err)
	}

	return nil
}

// EntryExists checks if an entry exists in a .info file
func EntryExists(dirPath, entryPath string) (bool, *Annotation, error) {
	annotations, err := ParseDirectory(dirPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse .info file: %w", err)
	}

	if annotation, exists := annotations[entryPath]; exists {
		return true, annotation, nil
	}

	return false, nil, nil
}

// UserChoice represents the user's choice when an entry already exists
type UserChoice int

const (
	UserChoiceReplace UserChoice = iota
	UserChoiceAppend
	UserChoiceQuit
)

// ActionType represents the type of action performed
type ActionType int

const (
	ActionAdded ActionType = iota
	ActionUpdated
	ActionCancelled
)

// ActionResult represents the result of an add-info operation
type ActionResult struct {
	Action ActionType
	Path   string
}

// UserPromptFunc is a function type for prompting the user
type UserPromptFunc func(path, currentDesc, newDesc string) (UserChoice, error)

// AddInfoEntry handles the business logic for adding/updating .info entries
func AddInfoEntry(dirPath, entryPath, description string, forceReplace bool, promptFunc UserPromptFunc) (*ActionResult, error) {
	// Check if entry already exists
	exists, existingAnnotation, err := EntryExists(dirPath, entryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing entry: %w", err)
	}

	action := UpdateActionReplace
	actionType := ActionAdded

	if exists {
		actionType = ActionUpdated

		if !forceReplace {
			// Entry exists and we haven't been told to force replace - ask user
			choice, err := promptFunc(entryPath, existingAnnotation.Description, description)
			if err != nil {
				return nil, fmt.Errorf("failed to get user choice: %w", err)
			}

			switch choice {
			case UserChoiceReplace:
				action = UpdateActionReplace
			case UserChoiceAppend:
				action = UpdateActionAppend
			case UserChoiceQuit:
				return &ActionResult{
					Action: ActionCancelled,
					Path:   entryPath,
				}, nil
			}
		}
	}

	// Add or update the entry
	if err := AddOrUpdateEntry(dirPath, entryPath, description, action); err != nil {
		return nil, fmt.Errorf("failed to update .info file: %w", err)
	}

	return &ActionResult{
		Action: actionType,
		Path:   entryPath,
	}, nil
}
