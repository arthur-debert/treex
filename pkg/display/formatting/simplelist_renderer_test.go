package formatting

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/tree"
)

func TestSimpleListRenderer(t *testing.T) {
	renderer := &SimpleListRenderer{}

	tests := []struct {
		name   string
		tree   *tree.Node
		checks []string
	}{
		{
			name: "basic structure",
			tree: createTestTree(),
			checks: []string{
				"test-root/",
				"  file1.txt",
				"  subdir/",
				"    nested.go",
				"    empty_dir/",
				"  no_annotation.md",
			},
		},
		{
			name: "single file",
			tree: &tree.Node{
				Name:  "single.txt",
				IsDir: false,
			},
			checks: []string{
				"single.txt\n",
			},
		},
		{
			name: "directory with slash",
			tree: &tree.Node{
				Name:  "mydir",
				IsDir: true,
			},
			checks: []string{
				"mydir/\n",
			},
		},
		{
			name: "deeply nested",
			tree: &tree.Node{
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
										Name:  "level3",
										IsDir: true,
										Children: []*tree.Node{
											{
												Name:  "deep.txt",
												IsDir: false,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			checks: []string{
				"root/",
				"  level1/",
				"    level2/",
				"      level3/",
				"        deep.txt",
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
			if renderer.Format() != format.FormatSimpleList {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatSimpleList)
			}

			if !renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = false, want true")
			}

			desc := renderer.Description()
			if !strings.Contains(desc, "Simple indented list") {
				t.Errorf("Description() = %q, expected to contain 'Simple indented list'", desc)
			}
		})
	}
}

func TestSimpleListRendererIgnoresAnnotations(t *testing.T) {
	// SimpleListRenderer should ignore annotations
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "annotated.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Title:       "This should be ignored",
					Description: "This should also be ignored",
					Notes:       "And this too",
				},
			},
			{
				Name:  "annotated_dir",
				IsDir: true,
				Annotation: &info.Annotation{
					Description: "Directory annotation - ignored",
				},
			},
		},
	}

	renderer := &SimpleListRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check that files/dirs are present
	if !strings.Contains(output, "annotated.txt") {
		t.Error("Expected output to contain annotated.txt")
	}
	if !strings.Contains(output, "annotated_dir/") {
		t.Error("Expected output to contain annotated_dir/")
	}

	// Check that annotations are NOT present
	if strings.Contains(output, "should be ignored") {
		t.Error("Output should not contain annotation text")
	}
	if strings.Contains(output, "Directory annotation") {
		t.Error("Output should not contain directory annotation")
	}
}

func TestSimpleListRendererConsistentIndentation(t *testing.T) {
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file1.txt",
				IsDir: false,
			},
			{
				Name:  "dir1",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "file2.txt",
						IsDir: false,
					},
					{
						Name:  "dir2",
						IsDir: true,
						Children: []*tree.Node{
							{
								Name:  "file3.txt",
								IsDir: false,
							},
						},
					},
				},
			},
		},
	}

	renderer := &SimpleListRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	
	// Check each line has correct indentation
	expectedIndents := map[string]int{
		"root/":      0,
		"file1.txt":  2,
		"dir1/":      2,
		"file2.txt":  4,
		"dir2/":      4,
		"file3.txt":  6,
	}

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		
		if expected, ok := expectedIndents[trimmed]; ok {
			if indent != expected {
				t.Errorf("Line %q has indent %d, expected %d", trimmed, indent, expected)
			}
		}
	}
}

func TestSimpleListRendererEmptyTree(t *testing.T) {
	root := &tree.Node{
		Name:  "empty",
		IsDir: true,
	}

	renderer := &SimpleListRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Should just have the root directory
	expected := "empty/\n"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestSimpleListRendererMixedContent(t *testing.T) {
	// Test with a mix of files and directories at various levels
	root := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "README.md", IsDir: false},
			{Name: ".gitignore", IsDir: false},
			{
				Name:  "src",
				IsDir: true,
				Children: []*tree.Node{
					{Name: "main.go", IsDir: false},
					{Name: "utils.go", IsDir: false},
					{
						Name:  "internal",
						IsDir: true,
						Children: []*tree.Node{
							{Name: "config.go", IsDir: false},
						},
					},
				},
			},
			{
				Name:  "docs",
				IsDir: true,
				Children: []*tree.Node{
					{Name: "api.md", IsDir: false},
					{Name: "guide.md", IsDir: false},
				},
			},
			{Name: "Makefile", IsDir: false},
		},
	}

	renderer := &SimpleListRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Verify the structure
	expected := `project/
  README.md
  .gitignore
  src/
    main.go
    utils.go
    internal/
      config.go
  docs/
    api.md
    guide.md
  Makefile
`

	if output != expected {
		t.Errorf("Output does not match expected structure.\nExpected:\n%s\nGot:\n%s", expected, output)
	}
}

func TestSimpleListRendererSpecialCharacters(t *testing.T) {
	// Test that special characters are preserved
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "file with spaces.txt", IsDir: false},
			{Name: "file-with-dashes.txt", IsDir: false},
			{Name: "file_with_underscores.txt", IsDir: false},
			{Name: "file@special#chars$.txt", IsDir: false},
			{Name: "日本語.txt", IsDir: false},
			{Name: "emoji😀.txt", IsDir: false},
		},
	}

	renderer := &SimpleListRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// All special characters should be preserved exactly
	checks := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file@special#chars$.txt",
		"日本語.txt",
		"emoji😀.txt",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected output to contain %q exactly", check)
		}
	}
}