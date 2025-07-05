package formatting

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/tree"
)

func TestHTMLRenderer(t *testing.T) {
	renderer := &HTMLRenderer{}

	tests := []struct {
		name   string
		tree   *tree.Node
		checks []string
	}{
		{
			name: "basic HTML structure",
			tree: createTestTree(),
			checks: []string{
				"<!DOCTYPE html>",
				"<html lang=\"en\">",
				"<meta charset=\"UTF-8\">",
				"<title>File Tree: test-root</title>",
				"<h1>🌳 test-root</h1>",
				"</body>",
				"</html>",
			},
		},
		{
			name: "files and directories",
			tree: createTestTree(),
			checks: []string{
				"file1.txt",
				"subdir/",
				"nested.go",
				"empty_dir/",
				"no_annotation.md",
			},
		},
		{
			name: "annotations",
			tree: &tree.Node{
				Name:  "root",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "annotated.txt",
						IsDir: false,
						Annotation: &info.Annotation{
							Path:  "annotated.txt",
							Notes: "This is a very important file\nWith multiple lines",
						},
					},
				},
			},
			checks: []string{
				"This is a very important file",
				"<span class=\"annotation\">",
			},
		},
		{
			name: "collapsible directories",
			tree: createTestTree(),
			checks: []string{
				"<details>",
				"<summary>",
				"</details>",
				"📁",
			},
		},
		{
			name: "file and directory icons",
			tree: createTestTree(),
			checks: []string{
				"📁", // directory icon
				"📄", // file icon
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := renderer.Render(tt.tree, format.RenderOptions{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Expected output to contain %q", check)
				}
			}

			// Check format and description
			if renderer.Format() != format.FormatHTML {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatHTML)
			}

			if renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = true, want false")
			}

			desc := renderer.Description()
			if !strings.Contains(desc, "Interactive HTML") {
				t.Errorf("Description() = %q, expected to contain 'Interactive HTML'", desc)
			}
		})
	}
}

func TestCompactHTMLRenderer(t *testing.T) {
	renderer := &CompactHTMLRenderer{}

	tests := []struct {
		name   string
		tree   *tree.Node
		checks []string
	}{
		{
			name: "compact structure",
			tree: createTestTree(),
			checks: []string{
				"<div class=\"tree-container\">",
				"<h3>📁 test-root</h3>",
				"</div>",
			},
		},
		{
			name: "no full HTML document",
			tree: createTestTree(),
			checks: []string{
				"file1.txt",
				"subdir",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := renderer.Render(tt.tree, format.RenderOptions{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			// Should NOT have full HTML document structure
			if strings.Contains(output, "<!DOCTYPE html>") {
				t.Error("Compact HTML should not contain DOCTYPE")
			}
			if strings.Contains(output, "<html") {
				t.Error("Compact HTML should not contain <html> tag")
			}

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Expected output to contain %q", check)
				}
			}

			// Check format
			if renderer.Format() != format.FormatCompactHTML {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatCompactHTML)
			}

			if renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = true, want false")
			}
		})
	}
}

func TestTableHTMLRenderer(t *testing.T) {
	renderer := &TableHTMLRenderer{}

	tests := []struct {
		name   string
		tree   *tree.Node
		checks []string
	}{
		{
			name: "table structure",
			tree: createTestTree(),
			checks: []string{
				"<!DOCTYPE html>",
				"<table>",
				"</table>",
				"<thead>",
				"<tbody>",
				"<th>Type</th>",
				"<th>Path</th>",
				"<th>Description</th>",
			},
		},
		{
			name: "table rows",
			tree: createTestTree(),
			checks: []string{
				"<tr>",
				"</tr>",
				"<td>",
				"file1.txt",
				"subdir",
			},
		},
		{
			name: "file types",
			tree: createTestTree(),
			checks: []string{
				"📁", // Directory icon
				"📄", // File icon
				"Directory",
				"File",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := renderer.Render(tt.tree, format.RenderOptions{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Expected output to contain %q", check)
				}
			}

			// Check format
			if renderer.Format() != format.FormatTableHTML {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatTableHTML)
			}

			if renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = true, want false")
			}

			desc := renderer.Description()
			if !strings.Contains(desc, "table") {
				t.Errorf("Description() = %q, expected to contain 'table'", desc)
			}
		})
	}
}

func TestHTMLEscaping(t *testing.T) {
	// Create a tree with special characters that need escaping
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "<script>alert('xss')</script>.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:  "<script>alert('xss')</script>.txt",
					Notes: "File with <html> tags & special chars",
				},
			},
			{
				Name:  "dir&name",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name: "file\"with'quotes.txt",
						IsDir: false,
					},
				},
			},
		},
	}

	renderers := []format.Renderer{
		&HTMLRenderer{},
		&CompactHTMLRenderer{},
		&TableHTMLRenderer{},
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			output, err := renderer.Render(root, format.RenderOptions{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			// Check that dangerous content is escaped
			if strings.Contains(output, "<script>alert") {
				t.Error("Output contains unescaped <script> tag")
			}
			if strings.Contains(output, "</script>") {
				t.Error("Output contains unescaped </script> tag")
			}

			// Check that escaped versions are present
			if !strings.Contains(output, "&lt;script&gt;") && !strings.Contains(output, "%3Cscript%3E") {
				t.Error("Expected output to contain escaped script tag")
			}
		})
	}
}

func TestHTMLDepthHandling(t *testing.T) {
	// Create a deeply nested structure
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
	}

	current := root
	for i := 0; i < 6; i++ {
		child := &tree.Node{
			Name:   "level" + string(rune('0'+i)),
			IsDir:  true,
			Parent: current,
		}
		current.Children = []*tree.Node{child}
		current = child
	}

	renderer := &TableHTMLRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check depth classes
	for i := 0; i < 5; i++ {
		depthClass := "depth-" + string(rune('0'+i))
		if !strings.Contains(output, depthClass) {
			t.Errorf("Expected output to contain depth class %q", depthClass)
		}
	}
}

func TestHTMLURLEncoding(t *testing.T) {
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file with spaces.txt",
				IsDir: false,
			},
			{
				Name:  "special#chars&in=name.txt",
				IsDir: false,
			},
		},
	}

	renderer := &HTMLRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check URL encoding - the files should be present with proper HTML escaping
	if !strings.Contains(output, "file with spaces.txt") && !strings.Contains(output, "file%20with%20spaces.txt") {
		t.Errorf("Expected spaces to be handled in output:\n%s", output)
	}
	if !strings.Contains(output, "special#chars&amp;in=name.txt") && !strings.Contains(output, "special%23chars%26in%3Dname.txt") {
		t.Errorf("Expected special characters to be handled in output:\n%s", output)
	}
}

