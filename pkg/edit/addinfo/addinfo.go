package addinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/types"
)

// UpdateAction represents the action taken when updating an entry
type UpdateAction int

const (
	// UpdateActionReplace replaces existing entry
	UpdateActionReplace UpdateAction = iota
	// UpdateActionAppend appends to existing entry
	UpdateActionAppend
	// UpdateActionSkip skips if entry exists
	UpdateActionSkip
)

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

// ActionResult represents the result of an add/update operation
type ActionResult struct {
	Action      ActionType
	Path        string
	Description string
	InfoFile    string
}

// UserPromptFunc is a function type for user prompts
type UserPromptFunc func(path string, existingDesc string, newDesc string) (UserChoice, error)

// WriteInfoFile writes annotations to a .info file
func WriteInfoFile(filePath string, annotations map[string]*types.Annotation) error {
	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Open file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create .info file: %w", err)
	}
	defer func() {
		_ = file.Close() // Ignore error in defer
	}()

	// Sort paths for consistent output
	paths := make([]string, 0, len(annotations))
	for path := range annotations {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// Write annotations
	for i, path := range paths {
		annotation := annotations[path]
		if annotation.Notes == "" {
			continue // Skip empty annotations
		}

		// Ensure the path uses forward slashes
		normalizedPath := filepath.ToSlash(path)

		// Write in the new format: path:notes
		notes := annotation.Notes

		// Only use the first line if there are multiple lines
		if idx := strings.Index(notes, "\n"); idx != -1 {
			notes = notes[:idx]
		}

		// Write path:notes format (single line only)
		if _, err := fmt.Fprintf(file, "%s: %s\n", normalizedPath, notes); err != nil {
			return fmt.Errorf("failed to write annotation: %w", err)
		}

		// Add blank line between entries (except after the last one)
		if i < len(paths)-1 {
			if _, err := fmt.Fprintln(file); err != nil {
				return fmt.Errorf("failed to write blank line: %w", err)
			}
		}
	}

	return nil
}

// AddOrUpdateEntry adds or updates an entry in the appropriate .info file
func AddOrUpdateEntry(dirPath, entryPath, description string, action UpdateAction) error {
	// Normalize paths
	dirPath = filepath.Clean(dirPath)
	entryPath = filepath.Clean(entryPath)

	// Check if entry exists
	entryFullPath := filepath.Join(dirPath, entryPath)
	if _, err := os.Stat(entryFullPath); os.IsNotExist(err) {
		return fmt.Errorf("entry does not exist: %s", entryFullPath)
	}

	// Determine which .info file to update
	// The .info file should be in the parent directory of the entry
	entryDir := filepath.Dir(entryFullPath)
	infoPath := filepath.Join(entryDir, ".info")

	// Parse existing .info file
	parser := info.NewParser()
	annotations, err := parser.ParseFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to parse .info file: %w", err)
	}

	// Get the relative path from the .info file's directory
	relPath, err := filepath.Rel(entryDir, entryFullPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Handle the entry
	if existing, exists := annotations[relPath]; exists {
		switch action {
		case UpdateActionSkip:
			return nil // Do nothing
		case UpdateActionAppend:
			// Append new description to existing
			if existing.Notes != "" && description != "" {
				existing.Notes = existing.Notes + "\n" + description
			} else if description != "" {
				existing.Notes = description
			}
		case UpdateActionReplace:
			// Replace with new description
			existing.Notes = description
		}
	} else {
		// Add new entry
		annotations[relPath] = &types.Annotation{
			Path:  relPath,
			Notes: description,
		}
	}

	// Write updated .info file
	return WriteInfoFile(infoPath, annotations)
}

// EntryExists checks if an entry exists in the info files
func EntryExists(dirPath, entryPath string) (bool, *types.Annotation, error) {
	// Normalize paths
	dirPath = filepath.Clean(dirPath)
	entryPath = filepath.Clean(entryPath)

	// Get full path
	entryFullPath := filepath.Join(dirPath, entryPath)

	// Check if the file/directory exists
	if _, err := os.Stat(entryFullPath); os.IsNotExist(err) {
		return false, nil, nil
	}

	// Determine which .info file to check
	entryDir := filepath.Dir(entryFullPath)
	infoPath := filepath.Join(entryDir, ".info")

	// Parse .info file
	parser := info.NewParser()
	annotations, err := parser.ParseFile(infoPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse .info file: %w", err)
	}

	// Get the relative path from the .info file's directory
	relPath, err := filepath.Rel(entryDir, entryFullPath)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get relative path: %w", err)
	}

	// Check if entry exists
	if annotation, exists := annotations[relPath]; exists {
		return true, annotation, nil
	}

	return false, nil, nil
}

// AddInfoEntry adds a new info entry with user interaction
func AddInfoEntry(dirPath, entryPath, description string, forceReplace bool, promptFunc UserPromptFunc) (*ActionResult, error) {
	// Check if entry already exists
	exists, existingAnnotation, err := EntryExists(dirPath, entryPath)
	if err != nil {
		return nil, err
	}

	result := &ActionResult{
		Path:        filepath.Join(dirPath, entryPath),
		Description: description,
	}

	// Determine which .info file will be updated
	entryFullPath := filepath.Join(dirPath, entryPath)
	entryDir := filepath.Dir(entryFullPath)
	result.InfoFile = filepath.Join(entryDir, ".info")

	if exists && existingAnnotation != nil {
		if forceReplace {
			// Force replace
			err = AddOrUpdateEntry(dirPath, entryPath, description, UpdateActionReplace)
			result.Action = ActionUpdated
		} else if promptFunc != nil {
			// Ask user what to do
			choice, err := promptFunc(entryPath, existingAnnotation.Notes, description)
			if err != nil {
				return nil, err
			}

			switch choice {
			case UserChoiceReplace:
				if err = AddOrUpdateEntry(dirPath, entryPath, description, UpdateActionReplace); err != nil {
					return nil, err
				}
				result.Action = ActionUpdated
			case UserChoiceAppend:
				if err = AddOrUpdateEntry(dirPath, entryPath, description, UpdateActionAppend); err != nil {
					return nil, err
				}
				result.Action = ActionUpdated
			case UserChoiceQuit:
				result.Action = ActionCancelled
			}
		} else {
			// Default to skip if no prompt function
			result.Action = ActionCancelled
		}
	} else {
		// New entry
		err = AddOrUpdateEntry(dirPath, entryPath, description, UpdateActionReplace)
		result.Action = ActionAdded
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}


// AddOrUpdateEntryInNamedFile adds or updates an entry in a custom-named info file
func AddOrUpdateEntryInNamedFile(dirPath, entryPath, description, infoFileName string, action UpdateAction) error {
	// Normalize paths
	dirPath = filepath.Clean(dirPath)
	entryPath = filepath.Clean(entryPath)

	// Check if entry exists
	entryFullPath := filepath.Join(dirPath, entryPath)
	if _, err := os.Stat(entryFullPath); os.IsNotExist(err) {
		return fmt.Errorf("entry does not exist: %s", entryFullPath)
	}

	// Determine which info file to update
	entryDir := filepath.Dir(entryFullPath)
	infoPath := filepath.Join(entryDir, infoFileName)

	// Parse existing info file
	parser := info.NewParser()
	annotations, err := parser.ParseFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to parse info file: %w", err)
	}

	// Get the relative path from the info file's directory
	relPath, err := filepath.Rel(entryDir, entryFullPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Handle the entry
	if existing, exists := annotations[relPath]; exists {
		switch action {
		case UpdateActionSkip:
			return nil // Do nothing
		case UpdateActionAppend:
			// Append new description to existing
			if existing.Notes != "" && description != "" {
				existing.Notes = existing.Notes + "\n" + description
			} else if description != "" {
				existing.Notes = description
			}
		case UpdateActionReplace:
			// Replace with new description
			existing.Notes = description
		}
	} else {
		// Add new entry
		annotations[relPath] = &types.Annotation{
			Path:  relPath,
			Notes: description,
		}
	}

	// Write updated info file
	return WriteInfoFile(infoPath, annotations)
}
