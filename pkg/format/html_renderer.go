package format

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/adebert/treex/pkg/tree"
)

// HTMLRenderer renders trees as interactive HTML with collapsible sections
type HTMLRenderer struct{}


func (r *HTMLRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	var builder strings.Builder

	// Add HTML document structure
	builder.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Tree: ` + html.EscapeString(root.Name) + `</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; line-height: 1.6; }
        h1 { color: #333; border-bottom: 2px solid #eee; padding-bottom: 10px; }
        details { margin-left: 20px; }
        summary { cursor: pointer; padding: 5px; border-radius: 3px; transition: background-color 0.2s; }
        summary:hover { background-color: #f5f5f5; }
        .file { margin-left: 20px; padding: 3px 0; }
        .annotation { color: #666; font-style: italic; margin-left: 5px; }
        .directory-icon { color: #4a90e2; }
        .file-icon { color: #7ed321; }
        a { text-decoration: none; color: inherit; }
        a:hover { text-decoration: underline; }
        code { background-color: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
`)

	_, _ = fmt.Fprintf(&builder, "    <h1>🌳 %s</h1>\n", html.EscapeString(root.Name))

	// Render the tree structure
	r.renderNode(root, &builder, "", true)

	builder.WriteString(`
</body>
</html>`)

	return builder.String(), nil
}

func (r *HTMLRenderer) Format() OutputFormat {
	return FormatHTML
}

func (r *HTMLRenderer) Description() string {
	return "Interactive HTML with collapsible directory trees"
}

func (r *HTMLRenderer) IsTerminalFormat() bool {
	return false
}

// renderNode renders a single node as HTML
func (r *HTMLRenderer) renderNode(node *tree.Node, builder *strings.Builder, currentPath string, isRoot bool) {
	var fullPath string
	if currentPath == "" {
		fullPath = node.Name
	} else {
		fullPath = currentPath + "/" + node.Name
	}

	// Skip root node display (already in title)
	if !isRoot {
		if node.IsDir && len(node.Children) > 0 {
			// Directory with children - use collapsible details
			builder.WriteString("    <details>\n")
			_, _ = fmt.Fprintf(builder, "        <summary><span class=\"directory-icon\">📁</span> <a href=\"%s\"><code>%s/</code></a>",
				url.PathEscape(fullPath), html.EscapeString(node.Name))

			// Add annotation
			if node.Annotation != nil {
				if node.Annotation.Title != "" {
					_, _ = fmt.Fprintf(builder, " <span class=\"annotation\">- %s</span>", html.EscapeString(node.Annotation.Title))
				}
			}
			builder.WriteString("</summary>\n")

			// Add detailed annotation if present
			if node.Annotation != nil && node.Annotation.Description != "" {
				lines := strings.Split(strings.TrimSpace(node.Annotation.Description), "\n")
				startIdx := 0
				if node.Annotation.Title == "" && len(lines) > 0 {
					startIdx = 1
				}

				for i := startIdx; i < len(lines); i++ {
					line := strings.TrimSpace(lines[i])
					if line != "" {
						_, _ = fmt.Fprintf(builder, "        <div class=\"annotation\">%s</div>\n", html.EscapeString(line))
					}
				}
			}
		} else if node.IsDir {
			// Empty directory
			builder.WriteString("    <div class=\"file\">")
			_, _ = fmt.Fprintf(builder, "<span class=\"directory-icon\">📁</span> <a href=\"%s\"><code>%s/</code></a>",
				url.PathEscape(fullPath), html.EscapeString(node.Name))

			if node.Annotation != nil && node.Annotation.Title != "" {
				_, _ = fmt.Fprintf(builder, " <span class=\"annotation\">- %s</span>", html.EscapeString(node.Annotation.Title))
			}
			builder.WriteString("</div>\n")
		} else {
			// File
			builder.WriteString("    <div class=\"file\">")
			_, _ = fmt.Fprintf(builder, "<span class=\"file-icon\">📄</span> <a href=\"%s\"><code>%s</code></a>",
				url.PathEscape(fullPath), html.EscapeString(node.Name))

			if node.Annotation != nil {
				if node.Annotation.Title != "" {
					_, _ = fmt.Fprintf(builder, " <span class=\"annotation\">- %s</span>", html.EscapeString(node.Annotation.Title))
				} else {
					lines := strings.Split(strings.TrimSpace(node.Annotation.Description), "\n")
					if len(lines) > 0 && lines[0] != "" {
						_, _ = fmt.Fprintf(builder, " <span class=\"annotation\">- %s</span>", html.EscapeString(lines[0]))
					}
				}
			}
			builder.WriteString("</div>\n")
		}
	}

	// Render children
	for _, child := range node.Children {
		r.renderNode(child, builder, fullPath, false)
	}

	// Close details tag for directories with children
	if !isRoot && node.IsDir && len(node.Children) > 0 {
		builder.WriteString("    </details>\n")
	}
}

// CompactHTMLRenderer renders a more compact HTML version
type CompactHTMLRenderer struct{}


func (r *CompactHTMLRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	var builder strings.Builder

	// Start HTML structure (more compact)
	_, _ = fmt.Fprintf(&builder, "<div class=\"tree-container\">\n<h3>📁 %s</h3>\n", html.EscapeString(root.Name))

	// Render tree
	r.renderCompactNode(root, &builder, "", true, 0)

	builder.WriteString("</div>")

	return builder.String(), nil
}

func (r *CompactHTMLRenderer) Format() OutputFormat {
	return FormatCompactHTML
}

func (r *CompactHTMLRenderer) Description() string {
	return "Compact HTML format without full document structure"
}

func (r *CompactHTMLRenderer) IsTerminalFormat() bool {
	return false
}

func (r *CompactHTMLRenderer) renderCompactNode(node *tree.Node, builder *strings.Builder, currentPath string, isRoot bool, depth int) {
	var fullPath string
	if currentPath == "" {
		fullPath = node.Name
	} else {
		fullPath = currentPath + "/" + node.Name
	}

	if !isRoot {
		indent := strings.Repeat("  ", depth)

		if node.IsDir && len(node.Children) > 0 {
			_, _ = fmt.Fprintf(builder, "%s<details><summary>📁 <a href=\"%s\">%s/</a>",
				indent, url.PathEscape(fullPath), html.EscapeString(node.Name))

			if node.Annotation != nil && node.Annotation.Title != "" {
				_, _ = fmt.Fprintf(builder, " <em>%s</em>", html.EscapeString(node.Annotation.Title))
			}
			builder.WriteString("</summary>\n")
		} else {
			icon := "📄"
			if node.IsDir {
				icon = "📁"
			}
			_, _ = fmt.Fprintf(builder, "%s<div>%s <a href=\"%s\">%s</a>",
				indent, icon, url.PathEscape(fullPath), html.EscapeString(node.Name))

			if node.Annotation != nil && node.Annotation.Title != "" {
				_, _ = fmt.Fprintf(builder, " <em>%s</em>", html.EscapeString(node.Annotation.Title))
			}
			builder.WriteString("</div>\n")
		}
	}

	// Render children
	for _, child := range node.Children {
		r.renderCompactNode(child, builder, fullPath, false, depth+1)
	}

	// Close details if directory with children
	if !isRoot && node.IsDir && len(node.Children) > 0 {
		indent := strings.Repeat("  ", depth)
		_, _ = fmt.Fprintf(builder, "%s</details>\n", indent)
	}
}

// TableHTMLRenderer renders as an HTML table
type TableHTMLRenderer struct{}


func (r *TableHTMLRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	var builder strings.Builder

	// Start HTML document
	builder.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Structure: ` + html.EscapeString(root.Name) + `</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f8f9fa; font-weight: 600; }
        tr:nth-child(even) { background-color: #f8f9fa; }
        tr:hover { background-color: #e9ecef; }
        .type-icon { font-size: 16px; }
        .path-cell { font-family: 'Monaco', 'Menlo', monospace; }
        .depth-0 { font-weight: bold; }
        .depth-1 { padding-left: 20px; }
        .depth-2 { padding-left: 40px; }
        .depth-3 { padding-left: 60px; }
        .depth-4 { padding-left: 80px; }
        a { text-decoration: none; color: #007bff; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
`)

	_, _ = fmt.Fprintf(&builder, "    <h1>📊 %s - File Structure</h1>\n", html.EscapeString(root.Name))

	// Create table
	builder.WriteString(`    <table>
        <thead>
            <tr>
                <th>Type</th>
                <th>Path</th>
                <th>Description</th>
            </tr>
        </thead>
        <tbody>
`)

	// Collect paths for table
	var paths []struct {
		fullPath   string
		name       string
		isDir      bool
		annotation string
		depth      int
	}

	r.collectPaths(root, "", 0, &paths)

	// Render table rows
	for _, path := range paths {
		typeIcon := "📄"
		typeText := "File"
		if path.isDir {
			typeIcon = "📁"
			typeText = "Directory"
		}

		depthClass := fmt.Sprintf("depth-%d", path.depth)
		if path.depth > 4 {
			depthClass = "depth-4"
		}

		annotation := path.annotation
		if annotation == "" {
			annotation = "-"
		}

		_, _ = fmt.Fprintf(&builder, "            <tr>\n")
		_, _ = fmt.Fprintf(&builder, "                <td><span class=\"type-icon\">%s</span> %s</td>\n", typeIcon, typeText)
		_, _ = fmt.Fprintf(&builder, "                <td class=\"path-cell %s\"><a href=\"%s\">%s</a></td>\n",
			depthClass, url.PathEscape(path.fullPath), html.EscapeString(path.name))
		_, _ = fmt.Fprintf(&builder, "                <td>%s</td>\n", html.EscapeString(annotation))
		_, _ = fmt.Fprintf(&builder, "            </tr>\n")
	}

	builder.WriteString(`        </tbody>
    </table>
</body>
</html>`)

	return builder.String(), nil
}

func (r *TableHTMLRenderer) collectPaths(node *tree.Node, parentPath string, depth int, paths *[]struct {
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
		if node.Annotation.Title != "" {
			annotation = node.Annotation.Title
		} else {
			lines := strings.Split(strings.TrimSpace(node.Annotation.Description), "\n")
			if len(lines) > 0 && lines[0] != "" {
				annotation = lines[0]
			}
		}
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

	for _, child := range node.Children {
		r.collectPaths(child, fullPath, depth+1, paths)
	}
}

func (r *TableHTMLRenderer) Format() OutputFormat {
	return FormatTableHTML
}

func (r *TableHTMLRenderer) Description() string {
	return "HTML table format with sortable columns"
}

func (r *TableHTMLRenderer) IsTerminalFormat() bool {
	return false
}
