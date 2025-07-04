package formatting

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/tree"
)

// MarkdownRenderer renders trees as Markdown with links
type MarkdownRenderer struct{}


func (r *MarkdownRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
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
func (r *MarkdownRenderer) renderNode(node *tree.Node, prefix, currentPath string, builder *strings.Builder, isRoot bool) {
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
			notes := node.Annotation.Notes
			if notes == "" {
				notes = node.Annotation.Description
			}
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

// NestedMarkdownRenderer renders trees as nested markdown with better organization
type NestedMarkdownRenderer struct{}


func (r *NestedMarkdownRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
	var builder strings.Builder

	// Add header with tree emoji
	builder.WriteString(fmt.Sprintf("# 🌳 %s\n\n", root.Name))

	// Add a table of contents if there are many top-level items
	if len(root.Children) > 5 {
		builder.WriteString("## 📋 Contents\n\n")
		for _, child := range root.Children {
			if child.IsDir {
				builder.WriteString(fmt.Sprintf("- [📁 %s](#%s)\n", child.Name, r.createAnchor(child.Name)))
			}
		}
		builder.WriteString("\n")
	}

	// Render structure
	r.renderNestedNode(root, "", "", &builder, true, 2)

	return builder.String(), nil
}

func (r *NestedMarkdownRenderer) Format() format.OutputFormat {
	return format.FormatNestedMarkdown
}

func (r *NestedMarkdownRenderer) Description() string {
	return "Nested Markdown with sections and table of contents"
}

func (r *NestedMarkdownRenderer) IsTerminalFormat() bool {
	return false
}

func (r *NestedMarkdownRenderer) renderNestedNode(node *tree.Node, prefix, currentPath string, builder *strings.Builder, isRoot bool, headerLevel int) {
	// Build the current path
	var fullPath string
	if currentPath == "" {
		fullPath = node.Name
	} else {
		fullPath = currentPath + "/" + node.Name
	}

	// Skip the root node display (already shown in main header)
	if !isRoot {
		// Create section header for directories, list item for files
		if node.IsDir && headerLevel <= 6 {
			// Directory as section header
			_, _ = fmt.Fprintf(builder, "%s 📁 [%s](%s)\n\n",
				strings.Repeat("#", headerLevel), node.Name, r.createFileLink(fullPath))

			// Add directory annotation
			if node.Annotation != nil && node.Annotation.Description != "" {
				_, _ = fmt.Fprintf(builder, "%s\n\n", node.Annotation.Description)
			}
		} else {
			// File as list item or regular directory if too deep
			var icon string
			if node.IsDir {
				icon = "📁"
			} else {
				icon = "📄"
			}

			_, _ = fmt.Fprintf(builder, "- %s [`%s`](%s)", icon, node.Name, r.createFileLink(fullPath))

			// Add annotation
			if node.Annotation != nil && node.Annotation.Description != "" {
				// Use first line of description
				lines := strings.Split(strings.TrimSpace(node.Annotation.Description), "\n")
				if len(lines) > 0 && lines[0] != "" {
					_, _ = fmt.Fprintf(builder, " - **%s**", lines[0])
					// Add remaining lines if any
					if len(lines) > 1 {
						remainingText := strings.Join(lines[1:], " ")
						remainingText = strings.TrimSpace(remainingText)
						if remainingText != "" {
							_, _ = fmt.Fprintf(builder, ": %s", remainingText)
						}
					}
				}
			}
			builder.WriteString("\n")
		}
	}

	// Render children
	for _, child := range node.Children {
		nextHeaderLevel := headerLevel + 1
		if node.IsDir && !isRoot && headerLevel <= 6 {
			// If current node is a directory section, don't increment for first level of children
			r.renderNestedNode(child, prefix, fullPath, builder, false, nextHeaderLevel)
		} else {
			r.renderNestedNode(child, prefix, fullPath, builder, false, headerLevel)
		}
	}

	// Add spacing after sections
	if !isRoot && node.IsDir && headerLevel <= 6 && len(node.Children) > 0 {
		builder.WriteString("\n")
	}
}

func (r *NestedMarkdownRenderer) createAnchor(text string) string {
	// Create GitHub-style anchor links
	anchor := strings.ToLower(text)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = strings.ReplaceAll(anchor, "/", "")
	return anchor
}

func (r *NestedMarkdownRenderer) createFileLink(path string) string {
	// URL encode the path to handle spaces and special characters
	return url.PathEscape(path)
}

// TableMarkdownRenderer renders trees as a markdown table
type TableMarkdownRenderer struct{}


func (r *TableMarkdownRenderer) Render(root *tree.Node, options format.RenderOptions) (string, error) {
	var builder strings.Builder

	// Add header
	builder.WriteString(fmt.Sprintf("# 📊 %s - File Structure\n\n", root.Name))

	// Create table header
	builder.WriteString("| Type | Path | Description |\n")
	builder.WriteString("|------|------|-------------|\n")

	// Collect all paths in a flat structure
	var paths []struct {
		fullPath   string
		name       string
		isDir      bool
		annotation string
		depth      int
	}

	r.collectTablePaths(root, "", 0, &paths)

	// Render table rows
	for _, path := range paths {
		var typeIcon string
		if path.isDir {
			typeIcon = "📁 Dir"
		} else {
			typeIcon = "📄 File"
		}

		// Create indented path display
		indent := strings.Repeat("&nbsp;&nbsp;", path.depth)
		pathDisplay := fmt.Sprintf("%s[`%s`](%s)", indent, path.name, url.PathEscape(path.fullPath))

		annotation := path.annotation
		if annotation == "" {
			annotation = "-"
		}

		builder.WriteString(fmt.Sprintf("| %s | %s | %s |\n", typeIcon, pathDisplay, annotation))
	}

	return builder.String(), nil
}

func (r *TableMarkdownRenderer) collectTablePaths(node *tree.Node, parentPath string, depth int, paths *[]struct {
	fullPath   string
	name       string
	isDir      bool
	annotation string
	depth      int
}) {
	var fullPath string
	if parentPath == "" {
		fullPath = node.Name
	} else {
		fullPath = parentPath + "/" + node.Name
	}

	var annotation string
	if node.Annotation != nil {
		// Use the first line of the description
		lines := strings.Split(strings.TrimSpace(node.Annotation.Description), "\n")
		if len(lines) > 0 && lines[0] != "" {
			annotation = lines[0]
		}
		// Escape markdown in annotation
		annotation = strings.ReplaceAll(annotation, "|", "\\|")
	}

	*paths = append(*paths, struct {
		fullPath   string
		name       string
		isDir      bool
		annotation string
		depth      int
	}{
		fullPath:   fullPath,
		name:       node.Name,
		isDir:      node.IsDir,
		annotation: annotation,
		depth:      depth,
	})

	// Process children
	for _, child := range node.Children {
		r.collectTablePaths(child, fullPath, depth+1, paths)
	}
}

func (r *TableMarkdownRenderer) Format() format.OutputFormat {
	return format.FormatTableMarkdown
}

func (r *TableMarkdownRenderer) Description() string {
	return "Markdown table format with file details"
}

func (r *TableMarkdownRenderer) IsTerminalFormat() bool {
	return false
}
