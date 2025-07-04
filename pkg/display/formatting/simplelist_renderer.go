package formatting

import (
	"strings"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/tree"
)

// SimpleListRenderer renders trees as a simple indented list of names.
type SimpleListRenderer struct{}

// Render implements the Renderer interface.
func (r *SimpleListRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
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
func (r *SimpleListRenderer) Format() format.OutputFormat {
	return format.FormatSimpleList // We'll define this constant next
}

// Description implements the Renderer interface.
func (r *SimpleListRenderer) Description() string {
	return "Simple indented list of file and directory names"
}

// IsTerminalFormat implements the Renderer interface.
func (r *SimpleListRenderer) IsTerminalFormat() bool {
	return true // Can be used in terminal, though it's plain
}
