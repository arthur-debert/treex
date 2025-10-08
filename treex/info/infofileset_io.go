package info

import (
	"github.com/jwaldrip/treex/treex/logging"
)

// ============================================================================
// FILESYSTEM I/O BOUNDARIES for InfoFileSet
// ============================================================================

// InfoFileSetLoader handles loading InfoFileSet from filesystem.
// This is where disk I/O occurs - InfoFileSet operations are pure functions.
type InfoFileSetLoader struct {
	fs InfoFileSystem
}

// NewInfoFileSetLoader creates a new InfoFileSet loader
func NewInfoFileSetLoader(fs InfoFileSystem) *InfoFileSetLoader {
	return &InfoFileSetLoader{fs: fs}
}

// LoadFromPath loads all .info files from a directory tree into an InfoFileSet.
// This is the primary entry point for reading InfoFiles from filesystem.
func (loader *InfoFileSetLoader) LoadFromPath(rootPath string) (*InfoFileSet, error) {
	// Find all .info files
	infoFilePaths, err := loader.fs.FindInfoFiles(rootPath)
	if err != nil {
		return nil, err
	}

	// Load each InfoFile
	infoFiles := make([]*InfoFile, 0, len(infoFilePaths))
	for _, path := range infoFilePaths {
		infoFile, err := loader.loadSingleInfoFile(path)
		if err != nil {
			logging.Warn().Str("file", path).Err(err).Msg("cannot load .info file")
			continue
		}
		infoFiles = append(infoFiles, infoFile)
	}

	// Create InfoFileSet with filesystem-aware path validation
	pathExists := loader.fs.PathExists
	return NewInfoFileSet(infoFiles, pathExists), nil
}

// LoadFromPaths loads specific .info file paths into an InfoFileSet.
// Useful for testing or partial loading scenarios.
func (loader *InfoFileSetLoader) LoadFromPaths(infoPaths []string) (*InfoFileSet, error) {
	infoFiles := make([]*InfoFile, 0, len(infoPaths))
	for _, path := range infoPaths {
		infoFile, err := loader.loadSingleInfoFile(path)
		if err != nil {
			logging.Warn().Str("file", path).Err(err).Msg("cannot load .info file")
			continue
		}
		infoFiles = append(infoFiles, infoFile)
	}

	pathExists := loader.fs.PathExists
	return NewInfoFileSet(infoFiles, pathExists), nil
}

func (loader *InfoFileSetLoader) loadSingleInfoFile(path string) (*InfoFile, error) {
	reader, err := loader.fs.ReadInfoFile(path)
	if err != nil {
		return nil, err
	}

	// Read all content manually (since io.ReadAll might not be available)
	var buf []byte
	chunk := make([]byte, 1024)
	for {
		n, readErr := reader.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	return NewInfoFile(path, string(buf)), nil
}

// ============================================================================

// InfoFileSetWriter handles writing InfoFileSet to filesystem.
// This is where disk I/O occurs - InfoFileSet operations are pure functions.
type InfoFileSetWriter struct {
	fs InfoFileSystem
}

// NewInfoFileSetWriter creates a new InfoFileSet writer
func NewInfoFileSetWriter(fs InfoFileSystem) *InfoFileSetWriter {
	return &InfoFileSetWriter{fs: fs}
}

// WriteInfoFileSet writes an InfoFileSet to the filesystem.
// Empty InfoFiles are deleted from the filesystem.
func (writer *InfoFileSetWriter) WriteInfoFileSet(infoFileSet *InfoFileSet) error {
	for _, infoFile := range infoFileSet.GetFiles() {
		if infoFile.IsEmpty() {
			// Delete empty files from filesystem
			err := writer.deleteInfoFile(infoFile.Path)
			if err != nil {
				logging.Warn().Str("file", infoFile.Path).Err(err).Msg("cannot delete empty .info file")
				// Continue with other files - don't fail the entire operation
			}
		} else {
			// Write non-empty files
			content := infoFile.String()
			err := writer.fs.WriteInfoFile(infoFile.Path, content)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteInfoFiles writes multiple InfoFiles to disk (backward compatibility).
// Empty InfoFiles are deleted from the filesystem.
func (writer *InfoFileSetWriter) WriteInfoFiles(infoFiles []*InfoFile) error {
	for _, infoFile := range infoFiles {
		if infoFile.IsEmpty() {
			// Delete empty files from filesystem
			err := writer.deleteInfoFile(infoFile.Path)
			if err != nil {
				logging.Warn().Str("file", infoFile.Path).Err(err).Msg("cannot delete empty .info file")
				// Continue with other files
			}
		} else {
			// Write non-empty files
			content := infoFile.String()
			err := writer.fs.WriteInfoFile(infoFile.Path, content)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (writer *InfoFileSetWriter) deleteInfoFile(path string) error {
	// For now, write empty content to "delete" the file
	// This could be enhanced to actually delete the file if InfoFileSystem supports it
	return writer.fs.WriteInfoFile(path, "")
}

// ============================================================================
// CONVENIENCE FUNCTIONS
// ============================================================================

// LoadInfoFileSet is a convenience function that loads an InfoFileSet from a path
// using the provided InfoFileSystem.
func LoadInfoFileSet(rootPath string, fs InfoFileSystem) (*InfoFileSet, error) {
	loader := NewInfoFileSetLoader(fs)
	return loader.LoadFromPath(rootPath)
}

// LoadInfoFileSetFromPaths is a convenience function that loads an InfoFileSet from
// specific paths using the provided InfoFileSystem.
func LoadInfoFileSetFromPaths(infoPaths []string, fs InfoFileSystem) (*InfoFileSet, error) {
	loader := NewInfoFileSetLoader(fs)
	return loader.LoadFromPaths(infoPaths)
}

// WriteInfoFileSet is a convenience function that writes an InfoFileSet to filesystem
// using the provided InfoFileSystem.
func WriteInfoFileSet(infoFileSet *InfoFileSet, fs InfoFileSystem) error {
	writer := NewInfoFileSetWriter(fs)
	return writer.WriteInfoFileSet(infoFileSet)
}
