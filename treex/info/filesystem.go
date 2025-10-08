package info

import (
	"io"
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
)

// InfoFileSystem provides an abstraction for file system operations needed by the info system
type InfoFileSystem interface {
	// ReadInfoFile opens and returns a reader for an .info file
	ReadInfoFile(path string) (io.Reader, error)

	// WriteInfoFile writes content to an .info file
	WriteInfoFile(path string, content string) error

	// PathExists checks if a path exists in the file system
	PathExists(path string) bool

	// FindInfoFiles discovers all .info files under a root directory
	FindInfoFiles(root string) ([]string, error)
}

// AferoInfoFileSystem implements InfoFileSystem using afero.Fs
type AferoInfoFileSystem struct {
	fs afero.Fs
}

// NewAferoInfoFileSystem creates a new AferoInfoFileSystem
func NewAferoInfoFileSystem(fs afero.Fs) *AferoInfoFileSystem {
	return &AferoInfoFileSystem{fs: fs}
}

// ReadInfoFile opens and returns a reader for an .info file
func (afs *AferoInfoFileSystem) ReadInfoFile(path string) (io.Reader, error) {
	return afs.fs.Open(path)
}

// WriteInfoFile writes content to an .info file
func (afs *AferoInfoFileSystem) WriteInfoFile(path string, content string) error {
	return afero.WriteFile(afs.fs, path, []byte(content), 0644)
}

// PathExists checks if a path exists in the file system
func (afs *AferoInfoFileSystem) PathExists(path string) bool {
	_, err := afs.fs.Stat(path)
	return err == nil
}

// FindInfoFiles discovers all .info files under a root directory
func (afs *AferoInfoFileSystem) FindInfoFiles(root string) ([]string, error) {
	var infoFiles []string

	err := afero.Walk(afs.fs, root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}
		if !info.IsDir() && filepath.Base(path) == ".info" {
			infoFiles = append(infoFiles, path)
		}
		return nil
	})

	return infoFiles, err
}
