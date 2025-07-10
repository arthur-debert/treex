package formatting

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/types"
)

func TestMarkdownRenderer(t *testing.T) {
	renderer := &MarkdownRenderer{}

	tests := []struct {
		name   string
		tree   *types.Node
		checks []string
	}{
		{
			name: "basic structure",
			tree: createTestTree(),
			checks: []string{
				"# test-root",
				"* 📄 [`file1.txt`]",
				"* **📁 [`subdir/`]",
				"  * 📄 [`nested.go`]",
				"  * **📁 [`empty_dir/`]",
				"* 📄 [`no_annotation.md`]",
			},
		},
		{
			name: "with annotations",
			tree: &types.Node{
				Name:  "root",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "annotated.txt",
						IsDir: false,
						Annotation: &types.Annotation{
							Path:  "file.txt",
							Notes: "This is the note",
						},
					},
					{
						Name:  "dir",
						IsDir: true,
						Annotation: &types.Annotation{
							Notes: "Directory notes",
						},
					},
				},
			},
			checks: []string{
				"- This is the note",
				"- Directory notes",
			},
		},
		{
			name: "URL encoding",
			tree: &types.Node{
				Name:  "root",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "file with spaces.txt",
						IsDir: false,
					},
					{
						Name:  "special#file.txt",
						IsDir: false,
					},
				},
			},
			checks: []string{
				"file%20with%20spaces.txt",
				"special%23file.txt",
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
					t.Errorf("Expected output to contain %q, got:\n%s", check, output)
				}
			}

			// Check format and description
			if renderer.Format() != format.FormatMarkdown {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatMarkdown)
			}

			if renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = true, want false")
			}

			desc := renderer.Description()
			if !strings.Contains(desc, "Markdown") {
				t.Errorf("Description() = %q, expected to contain 'Markdown'", desc)
			}
		})
	}
}

func TestMarkdownSpecialCharacters(t *testing.T) {
	// Test that markdown special characters are handled properly
	root := &types.Node{
		Name:  "root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file_with_underscores.txt",
				IsDir: false,
			},
			{
				Name:  "file*with*asterisks.txt",
				IsDir: false,
			},
			{
				Name:  "[brackets].txt",
				IsDir: false,
			},
		},
	}

	renderers := []format.Renderer{
		&MarkdownRenderer{},
	}

	for _, renderer := range renderers {
		t.Run(string(renderer.Format()), func(t *testing.T) {
			output, err := renderer.Render(root, format.RenderOptions{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			// Files should be present (exact formatting depends on renderer)
			if !strings.Contains(output, "file_with_underscores.txt") {
				t.Error("Expected output to contain file_with_underscores.txt")
			}
			if !strings.Contains(output, "asterisks.txt") {
				t.Error("Expected output to contain asterisks.txt")
			}
			if !strings.Contains(output, "brackets") {
				t.Error("Expected output to contain brackets")
			}
		})
	}
}

func TestMarkdownEmptyAnnotations(t *testing.T) {
	root := &types.Node{
		Name:  "root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file1.txt",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "file1.txt",
					Notes: "", // Empty notes
				},
			},
			{
				Name:  "file2.txt",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "file2.txt",
					Notes: "   \n  \n  ", // Only whitespace
				},
			},
		},
	}

	renderer := &MarkdownRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Debug: print the output to see what's actually rendered
	t.Logf("Markdown output:\n%s", output)

	// Should have files but no annotation text
	if !strings.Contains(output, "file1.txt") {
		t.Error("Expected output to contain file1.txt")
	}
	if !strings.Contains(output, "file2.txt") {
		t.Error("Expected output to contain file2.txt")
	}

	// Should not have dash followed by empty annotation
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "file1.txt") || strings.Contains(line, "file2.txt") {
			trimmed := strings.TrimSpace(line)
			if strings.HasSuffix(trimmed, "-") {
				t.Errorf("File line should not end with dash when annotation is empty: %q", line)
			}
			// Also check there's no "- " with nothing after it
			if strings.Contains(line, "- \n") || strings.HasSuffix(trimmed, "- ") {
				t.Errorf("File line has empty annotation: %q", line)
			}
		}
	}
}
