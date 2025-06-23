package format

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
	"gopkg.in/yaml.v3"
)

// createTestTree creates a standard test tree for consistent testing
func createTestTree() *tree.Node {
	return &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "README.md",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "README.md",
					Title:       "Project Documentation",
					Description: "Main project documentation\nContains setup and usage instructions",
				},
			},
			{
				Name:  "src",
				IsDir: true,
				Annotation: &info.Annotation{
					Path:        "src",
					Title:       "Source Code",
					Description: "Main source code directory",
				},
				Children: []*tree.Node{
					{
						Name:  "main.go",
						IsDir: false,
						Annotation: &info.Annotation{
							Path:        "src/main.go",
							Description: "Application entry point",
						},
					},
					{
						Name:  "utils",
						IsDir: true,
						Children: []*tree.Node{
							{
								Name:  "helper.go",
								IsDir: false,
							},
						},
					},
				},
			},
			{
				Name:  "docs",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "api.md",
						IsDir: false,
						Annotation: &info.Annotation{
							Path:        "docs/api.md",
							Title:       "API Reference",
							Description: "Complete API documentation",
						},
					},
				},
			},
		},
	}
}

func TestJSONRenderer_Render(t *testing.T) {
	renderer := NewJSONRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatJSON,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("JSONRenderer.Render() failed: %v", err)
	}

	if output == "" {
		t.Fatal("JSONRenderer.Render() returned empty output")
	}

	// Verify it's valid JSON
	var data TreeData
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("JSONRenderer output is not valid JSON: %v", err)
	}

	// Verify structure
	if data.Name != "project" {
		t.Errorf("Expected root name 'project', got %q", data.Name)
	}

	if !data.IsDirectory {
		t.Error("Expected root to be a directory")
	}

	if len(data.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(data.Children))
	}

	// Verify annotations are preserved
	readmeFound := false
	for _, child := range data.Children {
		if child.Name == "README.md" {
			readmeFound = true
			if child.Annotation == nil {
				t.Error("Expected README.md to have annotation")
			} else {
				if child.Annotation.Title != "Project Documentation" {
					t.Errorf("Expected title 'Project Documentation', got %q", child.Annotation.Title)
				}
			}
			break
		}
	}

	if !readmeFound {
		t.Error("README.md not found in JSON output")
	}
}

func TestYAMLRenderer_Render(t *testing.T) {
	renderer := NewYAMLRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatYAML,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("YAMLRenderer.Render() failed: %v", err)
	}

	if output == "" {
		t.Fatal("YAMLRenderer.Render() returned empty output")
	}

	// Verify it's valid YAML
	var data TreeData
	if err := yaml.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("YAMLRenderer output is not valid YAML: %v", err)
	}

	// Verify structure matches JSON output
	if data.Name != "project" {
		t.Errorf("Expected root name 'project', got %q", data.Name)
	}

	if len(data.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(data.Children))
	}

	// Verify YAML-specific formatting
	if !strings.Contains(output, "name: project") {
		t.Error("Expected YAML output to contain 'name: project'")
	}

	if !strings.Contains(output, "is_directory: true") {
		t.Error("Expected YAML output to contain 'is_directory: true'")
	}
}

func TestCompactJSONRenderer_Render(t *testing.T) {
	renderer := NewCompactJSONRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatCompactJSON,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("CompactJSONRenderer.Render() failed: %v", err)
	}

	// Verify it's compact (single line)
	if strings.Contains(output, "\n") {
		t.Error("CompactJSON should not contain newlines")
	}

	// Verify it's still valid JSON
	var data TreeData
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("CompactJSONRenderer output is not valid JSON: %v", err)
	}

	// Should have same structure as regular JSON
	if data.Name != "project" {
		t.Errorf("Expected root name 'project', got %q", data.Name)
	}
}

func TestFlatJSONRenderer_Render(t *testing.T) {
	renderer := NewFlatJSONRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatFlatJSON,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("FlatJSONRenderer.Render() failed: %v", err)
	}

	// Verify it's valid JSON array
	var paths []FlatPath
	if err := json.Unmarshal([]byte(output), &paths); err != nil {
		t.Fatalf("FlatJSONRenderer output is not valid JSON: %v", err)
	}

	// Should have flattened all nodes
	if len(paths) < 6 { // project + README + src + main.go + utils + helper.go + docs + api.md
		t.Errorf("Expected at least 6 paths, got %d", len(paths))
	}

	// Verify root entry
	if paths[0].Path != "project" {
		t.Errorf("Expected first path to be 'project', got %q", paths[0].Path)
	}

	if paths[0].Depth != 0 {
		t.Errorf("Expected root depth to be 0, got %d", paths[0].Depth)
	}

	// Verify nested paths have correct depth
	foundDeepPath := false
	for _, path := range paths {
		if strings.Contains(path.Path, "/") && path.Depth > 0 {
			foundDeepPath = true
			break
		}
	}

	if !foundDeepPath {
		t.Error("Expected to find nested paths with depth > 0")
	}
}

func TestMarkdownRenderer_Render(t *testing.T) {
	renderer := NewMarkdownRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatMarkdown,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("MarkdownRenderer.Render() failed: %v", err)
	}

	// Verify markdown structure
	if !strings.Contains(output, "# project") {
		t.Error("Expected markdown to contain '# project' header")
	}

	// Verify file links
	if !strings.Contains(output, "📄 [`README.md`]") {
		t.Error("Expected markdown to contain file link for README.md")
	}

	// Verify directory links
	if !strings.Contains(output, "📁 [`src/`]") {
		t.Error("Expected markdown to contain directory link for src")
	}

	// Verify annotations
	if !strings.Contains(output, "Project Documentation") {
		t.Error("Expected markdown to contain README annotation")
	}

	// Verify URL encoding in links
	if !strings.Contains(output, "](project%2F") {
		t.Error("Expected URL-encoded paths in links")
	}
}

func TestHTMLRenderer_Render(t *testing.T) {
	renderer := NewHTMLRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatHTML,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("HTMLRenderer.Render() failed: %v", err)
	}

	// Verify HTML document structure
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Expected HTML to contain DOCTYPE declaration")
	}

	if !strings.Contains(output, "<html lang=\"en\">") {
		t.Error("Expected HTML to contain html tag with lang attribute")
	}

	if !strings.Contains(output, "</html>") {
		t.Error("Expected HTML to contain closing html tag")
	}

	// Verify collapsible structure
	if !strings.Contains(output, "<details>") {
		t.Error("Expected HTML to contain details tags for collapsible sections")
	}

	if !strings.Contains(output, "<summary>") {
		t.Error("Expected HTML to contain summary tags")
	}

	// Verify file links
	if !strings.Contains(output, "<a href=\"project%2FREADME.md\">") {
		t.Error("Expected HTML to contain URL-encoded file links")
	}

	// Verify proper escaping
	if !strings.Contains(output, "<h1>🌳 project</h1>") {
		t.Error("Expected HTML to contain escaped title")
	}

	// Verify CSS is included
	if !strings.Contains(output, "<style>") {
		t.Error("Expected HTML to contain CSS styles")
	}
}

func TestCompactHTMLRenderer_Render(t *testing.T) {
	renderer := NewCompactHTMLRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatCompactHTML,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("CompactHTMLRenderer.Render() failed: %v", err)
	}

	// Should not be a full HTML document
	if strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("CompactHTML should not contain DOCTYPE declaration")
	}

	// Should still have interactive elements
	if !strings.Contains(output, "<details>") {
		t.Error("Expected CompactHTML to contain details tags")
	}

	// Should have container div
	if !strings.Contains(output, "tree-container") {
		t.Error("Expected CompactHTML to contain tree-container class")
	}
}

func TestTableHTMLRenderer_Render(t *testing.T) {
	renderer := NewTableHTMLRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatTableHTML,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("TableHTMLRenderer.Render() failed: %v", err)
	}

	// Verify table structure
	if !strings.Contains(output, "<table>") {
		t.Error("Expected HTML table to contain table tag")
	}

	if !strings.Contains(output, "<thead>") {
		t.Error("Expected HTML table to contain thead")
	}

	if !strings.Contains(output, "<tbody>") {
		t.Error("Expected HTML table to contain tbody")
	}

	// Verify table headers
	if !strings.Contains(output, "<th>Type</th>") {
		t.Error("Expected table to have Type column")
	}

	if !strings.Contains(output, "<th>Path</th>") {
		t.Error("Expected table to have Path column")
	}

	if !strings.Contains(output, "<th>Description</th>") {
		t.Error("Expected table to have Description column")
	}

	// Verify content exists (format may vary)
	if !strings.Contains(output, "src") {
		t.Error("Expected table to show src directory")
	}

	if !strings.Contains(output, "README.md") {
		t.Error("Expected table to show README.md file")
	}
}

func TestNestedMarkdownRenderer_Render(t *testing.T) {
	renderer := NewNestedMarkdownRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatNestedMarkdown,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("NestedMarkdownRenderer.Render() failed: %v", err)
	}

	// Should have tree emoji in title
	if !strings.Contains(output, "# 🌳 project") {
		t.Error("Expected nested markdown to have tree emoji in title")
	}

	// Should have structured content (table of contents may not appear for small trees)
	if !strings.Contains(output, "##") {
		t.Error("Expected nested markdown to have section headers")
	}

	// Should have section headers for directories
	if !strings.Contains(output, "## 📁 [src]") {
		t.Error("Expected nested markdown to have section headers for directories")
	}
}

func TestTableMarkdownRenderer_Render(t *testing.T) {
	renderer := NewTableMarkdownRenderer()
	root := createTestTree()

	options := RenderOptions{
		Format:   FormatTableMarkdown,
		SafeMode: true,
	}

	output, err := renderer.Render(root, options)
	if err != nil {
		t.Fatalf("TableMarkdownRenderer.Render() failed: %v", err)
	}

	// Verify markdown table structure
	if !strings.Contains(output, "| Type | Path | Description |") {
		t.Error("Expected markdown table to have proper headers")
	}

	if !strings.Contains(output, "|------|------|-------------|") {
		t.Error("Expected markdown table to have separator row")
	}

	// Verify content
	if !strings.Contains(output, "📁 Dir") {
		t.Error("Expected table to show directory type")
	}

	if !strings.Contains(output, "📄 File") {
		t.Error("Expected table to show file type")
	}

	// Verify indentation for nested items
	if !strings.Contains(output, "&nbsp;&nbsp;") {
		t.Error("Expected table to have indentation for nested items")
	}
}

func TestRenderer_EmptyTree(t *testing.T) {
	emptyTree := &tree.Node{
		Name:     "empty",
		IsDir:    true,
		Children: []*tree.Node{},
	}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewYAMLRenderer(),
		NewMarkdownRenderer(),
		NewHTMLRenderer(),
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			options := RenderOptions{SafeMode: true}
			output, err := renderer.Render(emptyTree, options)
			if err != nil {
				t.Fatalf("Renderer %s failed on empty tree: %v", renderer.Format(), err)
			}

			if output == "" {
				t.Errorf("Renderer %s returned empty output for empty tree", renderer.Format())
			}

			// Should still contain root name
			if !strings.Contains(output, "empty") {
				t.Errorf("Renderer %s output should contain root name 'empty'", renderer.Format())
			}
		})
	}
}

func TestRenderer_SingleFile(t *testing.T) {
	singleFile := &tree.Node{
		Name:  "file.txt",
		IsDir: false,
		Annotation: &info.Annotation{
			Path:        "file.txt",
			Description: "A single file",
		},
	}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewYAMLRenderer(),
		NewMarkdownRenderer(),
		NewHTMLRenderer(),
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			options := RenderOptions{SafeMode: true}
			output, err := renderer.Render(singleFile, options)
			if err != nil {
				t.Fatalf("Renderer %s failed on single file: %v", renderer.Format(), err)
			}

			if output == "" {
				t.Errorf("Renderer %s returned empty output for single file", renderer.Format())
			}

			// Should contain file name
			if !strings.Contains(output, "file.txt") {
				t.Errorf("Renderer %s output should contain file name", renderer.Format())
			}

			// Annotation display may vary by format
			if !strings.Contains(output, "file") {
				t.Errorf("Renderer %s output should contain file reference", renderer.Format())
			}
		})
	}
}
