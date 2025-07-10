package formatting

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/types"
)

// MarkdownRenderer renders trees as Markdown with links
type MarkdownRenderer struct{}

func (r *MarkdownRenderer) Render(root *types.Node, options format.RenderOptions) (string, error) {
	var builder strings.Builder

	// Add header
	builder.WriteString(fmt.Sprintf("# %s\n\n", root.Name))

	// Add tree structure
	r.renderNode(root, "", "", &builder, true)

	return builder.String(), nil
}

func (r *MarkdownRenderer) Format() format.OutputFormat {
	return format.FormatMarkdown
}

func (r *MarkdownRenderer) Description() string {
	return "Markdown format with clickable file links"
}

func (r *MarkdownRenderer) IsTerminalFormat() bool {
	return false
}

// renderNode renders a node and its children in markdown format
func (r *MarkdownRenderer) renderNode(node *types.Node, prefix, currentPath string, builder *strings.Builder, isRoot bool) {
	// Build the current path
	var fullPath string
	if currentPath == "" {
		fullPath = node.Name
	} else {
		fullPath = currentPath + "/" + node.Name
	}

	// Skip the root node display (already shown in header)
	if !isRoot {
		// Create markdown list item with link
		listPrefix := prefix + "* "

		// Create the file/directory link
		var nodeDisplay string
		if node.IsDir {
			// Directory - bold with folder emoji
			nodeDisplay = fmt.Sprintf("**📁 [`%s/`](%s)** ", node.Name, r.createFileLink(fullPath))
		} else {
			// File - code format with file emoji
			nodeDisplay = fmt.Sprintf("📄 [`%s`](%s) ", node.Name, r.createFileLink(fullPath))
		}

		// Add annotation if present
		if node.Annotation != nil {
			notes := strings.TrimSpace(node.Annotation.Notes)
			if notes != "" {
				nodeDisplay += fmt.Sprintf("- %s", notes)
			}
		}

		builder.WriteString(listPrefix + nodeDisplay + "\n")
	}

	// Render children
	for _, child := range node.Children {
		nextPrefix := prefix
		if !isRoot {
			nextPrefix = prefix + "  "
		}
		r.renderNode(child, nextPrefix, fullPath, builder, false)
	}
}

// createFileLink creates a relative file link for the given path
func (r *MarkdownRenderer) createFileLink(path string) string {
	// URL encode the path to handle spaces and special characters
	return url.PathEscape(path)
}
