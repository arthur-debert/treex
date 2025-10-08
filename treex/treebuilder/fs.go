package treebuilder

import (
	"github.com/spf13/afero"
	"treex/treex/types"
)

type FileSystem interface {
	afero.Fs
}

type Builder struct {
	fs FileSystem
}

func NewBuilder(fs FileSystem) *Builder {
	return &Builder{fs: fs}
}

// Build creates a tree from the filesystem according to options
// This is a placeholder implementation for testing the helpers
func (b *Builder) Build(opts types.TreeOptions) (*types.Node, error) {
	// TODO: Implement the actual tree building logic
	// For now, return a simple stub to test the helper infrastructure
	return &types.Node{
		Name:     "stub",
		Path:     opts.Root,
		IsDir:    true,
		Children: []*types.Node{},
	}, nil
}
