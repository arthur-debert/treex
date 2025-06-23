package format

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func TestRenderers_SpecialCharacters(t *testing.T) {
	// Create tree with special characters that need encoding/escaping
	specialTree := &tree.Node{
		Name:  "test & <script>",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file with spaces.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "file with spaces.txt",
					Title:       "Title with \"quotes\" & ampersands",
					Description: "Description with <tags> and | pipes",
				},
			},
			{
				Name:  "dir/with/slashes",
				IsDir: true,
				Annotation: &info.Annotation{
					Path:        "dir/with/slashes",
					Description: "Path with unicode: 🌳 and newlines\nSecond line",
				},
			},
		},
	}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewYAMLRenderer(),
		NewMarkdownRenderer(),
		NewHTMLRenderer(),
		NewTableMarkdownRenderer(),
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			options := RenderOptions{SafeMode: true}
			output, err := renderer.Render(specialTree, options)
			if err != nil {
				t.Fatalf("Renderer %s failed with special characters: %v", renderer.Format(), err)
			}

			if output == "" {
				t.Error("Renderer returned empty output with special characters")
			}

			// HTML renderers should escape HTML entities
			if renderer.Format() == FormatHTML || renderer.Format() == FormatCompactHTML || renderer.Format() == FormatTableHTML {
				if strings.Contains(output, "<script>") {
					t.Error("HTML renderer should escape script tags")
				}
				// Should contain escaped version
				if !strings.Contains(output, "&lt;script&gt;") {
					t.Error("HTML renderer should contain escaped script tags")
				}
			}

			// Markdown table should escape pipes
			if renderer.Format() == FormatTableMarkdown {
				if strings.Contains(output, "| pipes") && !strings.Contains(output, "\\| pipes") {
					t.Error("Markdown table should escape pipes in content")
				}
			}

			// All renderers should handle the tree name
			if !strings.Contains(output, "test") {
				t.Error("Output should contain root name")
			}
		})
	}
}

func TestRenderers_NilAnnotations(t *testing.T) {
	// Tree with mix of nil and non-nil annotations
	mixedTree := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:       "annotated.txt",
				IsDir:      false,
				Annotation: &info.Annotation{Path: "annotated.txt", Description: "Has annotation"},
			},
			{
				Name:       "not-annotated.txt",
				IsDir:      false,
				Annotation: nil, // Explicitly nil
			},
			{
				Name:  "empty-annotation.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "empty-annotation.txt",
					Title:       "",
					Description: "",
				},
			},
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
			output, err := renderer.Render(mixedTree, options)
			if err != nil {
				t.Fatalf("Renderer %s failed with nil annotations: %v", renderer.Format(), err)
			}

			// Should handle all files regardless of annotation status
			if !strings.Contains(output, "annotated.txt") {
				t.Error("Should contain annotated file")
			}
			if !strings.Contains(output, "not-annotated.txt") {
				t.Error("Should contain non-annotated file")
			}
			if !strings.Contains(output, "empty-annotation.txt") {
				t.Error("Should contain empty annotation file")
			}
		})
	}
}

func TestRenderers_VeryLongContent(t *testing.T) {
	// Create tree with very long descriptions
	longDescription := strings.Repeat("This is a very long description. ", 100)
	longTree := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "file.txt",
					Title:       "Short title",
					Description: longDescription,
				},
			},
		},
	}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewMarkdownRenderer(),
		// Skip HTMLRenderer for this test as it handles long content differently
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			options := RenderOptions{SafeMode: true}
			output, err := renderer.Render(longTree, options)
			if err != nil {
				t.Fatalf("Renderer %s failed with long content: %v", renderer.Format(), err)
			}

			// Should contain the long description (at least part of it)
			if !strings.Contains(output, "very long description") {
				t.Error("Should contain long description")
			}

			// Output should be reasonable in size (allow for markup overhead)
			expectedMinSize := len(longDescription) / 2  // Allow for possible truncation
			expectedMaxSize := len(longDescription) * 10 // Allow for significant markup overhead
			if len(output) < expectedMinSize {
				t.Error("Output seems too small for long content")
			}
			if len(output) > expectedMaxSize {
				t.Error("Output seems excessively large for long content")
			}
		})
	}
}

func TestRenderers_DeepNesting(t *testing.T) {
	// Create deeply nested tree
	root := &tree.Node{Name: "root", IsDir: true}
	current := root

	// Create 10 levels deep
	for i := 0; i < 10; i++ {
		child := &tree.Node{
			Name:  "level" + string(rune('0'+i)),
			IsDir: true,
			Annotation: &info.Annotation{
				Path:        "level" + string(rune('0'+i)),
				Description: "Level " + string(rune('0'+i)) + " directory",
			},
		}
		current.Children = []*tree.Node{child}
		current = child
	}

	// Add a file at the deepest level
	current.Children = []*tree.Node{
		{
			Name:  "deep-file.txt",
			IsDir: false,
			Annotation: &info.Annotation{
				Path:        "deep-file.txt",
				Description: "File at maximum depth",
			},
		},
	}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewFlatJSONRenderer(),
		NewMarkdownRenderer(),
		NewHTMLRenderer(),
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			options := RenderOptions{SafeMode: true}
			output, err := renderer.Render(root, options)
			if err != nil {
				t.Fatalf("Renderer %s failed with deep nesting: %v", renderer.Format(), err)
			}

			// Should contain root and deep file
			if !strings.Contains(output, "root") {
				t.Error("Should contain root")
			}
			if !strings.Contains(output, "deep-file.txt") {
				t.Error("Should contain deep file")
			}

			// Flat JSON should handle depth in some form
			if renderer.Format() == FormatFlatJSON {
				if !strings.Contains(output, "depth") {
					t.Error("FlatJSON should include depth information")
				}
			}
		})
	}
}

func TestRenderers_EmptyStrings(t *testing.T) {
	// Tree with empty string names and descriptions
	emptyTree := &tree.Node{
		Name:  "", // Empty root name
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "normal.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "normal.txt",
					Title:       "",    // Empty title
					Description: "   ", // Whitespace-only description
				},
			},
		},
	}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewMarkdownRenderer(),
		NewHTMLRenderer(),
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			options := RenderOptions{SafeMode: true}
			output, err := renderer.Render(emptyTree, options)
			if err != nil {
				t.Fatalf("Renderer %s failed with empty strings: %v", renderer.Format(), err)
			}

			// Should not crash and should produce some output
			if len(output) < 10 {
				t.Error("Output is suspiciously short for empty strings test")
			}

			// Should contain the file name
			if !strings.Contains(output, "normal.txt") {
				t.Error("Should contain file name")
			}
		})
	}
}

func TestMarkdownRenderer_URLEncoding(t *testing.T) {
	renderer := NewMarkdownRenderer()

	// Tree with characters that need URL encoding
	tree := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file with spaces & symbols.txt",
				IsDir: false,
			},
			{
				Name:  "path/with/slashes",
				IsDir: true,
			},
		},
	}

	options := RenderOptions{SafeMode: true}
	output, err := renderer.Render(tree, options)
	if err != nil {
		t.Fatalf("MarkdownRenderer failed with URL encoding test: %v", err)
	}

	// Should contain the files (URL encoding format may vary)
	if !strings.Contains(output, "file with spaces") {
		t.Error("Should contain file with spaces in some form")
	}

	if !strings.Contains(output, "path/with/slashes") {
		t.Error("Should contain path with slashes in some form")
	}
}

func TestHTMLRenderer_XSSPrevention(t *testing.T) {
	renderer := NewHTMLRenderer()

	// Tree with potential XSS content
	tree := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "<script>alert('xss')</script>",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "<script>alert('xss')</script>",
					Title:       "<img src=x onerror=alert('xss')>",
					Description: "javascript:alert('xss')",
				},
			},
		},
	}

	options := RenderOptions{SafeMode: true}
	output, err := renderer.Render(tree, options)
	if err != nil {
		t.Fatalf("HTMLRenderer failed with XSS test: %v", err)
	}

	// Should not contain dangerous unescaped content
	if strings.Contains(output, "<script>alert('xss')</script>") {
		t.Error("HTML should not contain exact unescaped script tags")
	}

	// Should contain the filename in some escaped form
	if !strings.Contains(output, "script") {
		t.Error("HTML should contain the filename in some form")
	}
}

func TestFlatJSONRenderer_PathGeneration(t *testing.T) {
	renderer := NewFlatJSONRenderer()

	// Tree with nested structure to test path generation
	tree := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "level1",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "level2",
						IsDir: true,
						Children: []*tree.Node{
							{
								Name:  "file.txt",
								IsDir: false,
							},
						},
					},
				},
			},
		},
	}

	options := RenderOptions{SafeMode: true}
	output, err := renderer.Render(tree, options)
	if err != nil {
		t.Fatalf("FlatJSONRenderer failed: %v", err)
	}

	// Should contain various path elements (format may vary)
	if !strings.Contains(output, "root") {
		t.Error("Should contain root in output")
	}

	if !strings.Contains(output, "level1") {
		t.Error("Should contain level1 in output")
	}

	if !strings.Contains(output, "level2") {
		t.Error("Should contain level2 in output")
	}

	if !strings.Contains(output, "file.txt") {
		t.Error("Should contain file.txt in output")
	}

	// Should have depth information in some form
	if !strings.Contains(output, "depth") {
		t.Error("Should contain depth information")
	}
}

func TestRenderers_ConsistentOutput(t *testing.T) {
	// Test that renderers produce consistent output for the same input
	tree := createTestTree()
	options := RenderOptions{SafeMode: true}

	renderers := []Renderer{
		NewJSONRenderer(),
		NewYAMLRenderer(),
		NewMarkdownRenderer(),
	}

	// Render twice and compare
	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			output1, err1 := renderer.Render(tree, options)
			output2, err2 := renderer.Render(tree, options)

			if err1 != nil || err2 != nil {
				t.Fatalf("Renderer errors: %v, %v", err1, err2)
			}

			if output1 != output2 {
				t.Error("Renderer should produce consistent output for same input")
			}

			if output1 == "" {
				t.Error("Renderer should not produce empty output")
			}
		})
	}
}
