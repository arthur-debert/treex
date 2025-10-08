package treebuilder

import "github.com/spf13/afero"

type FileSystem interface {
	afero.Fs
}

type Builder struct {
	fs FileSystem
}

func NewBuilder(fs FileSystem) *Builder {
	return &Builder{fs: fs}
}
