package format

import (
	"strings"

	"github.com/adebert/treex/pkg/tree"
)

// SimpleListRenderer renders trees as a simple indented list of names.
type SimpleListRenderer struct{}

// NewSimpleListRenderer creates a new SimpleListRenderer.
func NewSimpleListRenderer() *SimpleListRenderer {
	return &SimpleListRenderer{}
}

// Render implements the Renderer interface.
func (r *SimpleListRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	var builder strings.Builder
	r.renderNode(root, &builder, 0)
	return builder.String(), nil
}

func (r *SimpleListRenderer) renderNode(node *tree.Node, builder *strings.Builder, depth int) {
	builder.WriteString(strings.Repeat("  ", depth)) // Indentation
	builder.WriteString(node.Name)
	if node.IsDir {
		builder.WriteString("/")
	}
	builder.WriteString("\n")

	for _, child := range node.Children {
		r.renderNode(child, builder, depth+1)
	}
}

// Format implements the Renderer interface.
func (r *SimpleListRenderer) Format() OutputFormat {
	return FormatSimpleList // We'll define this constant next
}

// Description implements the Renderer interface.
func (r *SimpleListRenderer) Description() string {
	return "Simple indented list of file and directory names"
}

// IsTerminalFormat implements the Renderer interface.
func (r *SimpleListRenderer) IsTerminalFormat() bool {
	return true // Can be used in terminal, though it's plain
}
