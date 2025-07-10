package formatting

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/format"
	"github.com/adebert/treex/pkg/core/types"
)

// Helper function to create a test tree
func createTestTree() *types.Node {
	root := &types.Node{
		Name:  "test-root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file1.txt",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "file1.txt",
					Notes: "This is a test file",
				},
			},
			{
				Name:  "subdir",
				IsDir: true,
				Annotation: &types.Annotation{
					Path:  "subdir",
					Notes: "A subdirectory for testing",
				},
				Children: []*types.Node{
					{
						Name:  "nested.go",
						IsDir: false,
						Annotation: &types.Annotation{
							Path:  "subdir/nested.go",
							Notes: "Nested Go file",
						},
					},
					{
						Name:  "empty_dir",
						IsDir: true,
					},
				},
			},
			{
				Name:  "no_annotation.md",
				IsDir: false,
			},
		},
	}

	// Set parent relationships
	for _, child := range root.Children {
		child.Parent = root
		if child.IsDir && child.Children != nil {
			for _, grandchild := range child.Children {
				grandchild.Parent = child
			}
		}
	}

	return root
}

func TestColorRenderer(t *testing.T) {
	renderer := &ColorRenderer{}

	tests := []struct {
		name    string
		options format.RenderOptions
		checks  []string
	}{
		{
			name: "basic render",
			options: format.RenderOptions{
				SafeMode: false,
			},
			checks: []string{
				"test-root",
				"file1.txt",
				"subdir",
				"nested.go",
				"empty_dir",
				"no_annotation.md",
			},
		},
		{
			name: "with extra spacing",
			options: format.RenderOptions{
				SafeMode: false,
			},
			checks: []string{
				"test-root",
				"file1.txt",
				"subdir",
			},
		},
		{
			name: "safe mode",
			options: format.RenderOptions{
				SafeMode: true,
			},
			checks: []string{
				"test-root",
				"file1.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := createTestTree()
			output, err := renderer.Render(tree, tt.options)

			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Expected output to contain %q, got:\n%s", check, output)
				}
			}

			// Check format and description
			if renderer.Format() != format.FormatColor {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatColor)
			}

			if !renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = false, want true")
			}

			desc := renderer.Description()
			if desc == "" {
				t.Error("Description() returned empty string")
			}
		})
	}
}

func TestNoColorRenderer(t *testing.T) {
	renderer := &NoColorRenderer{}

	tests := []struct {
		name    string
		options format.RenderOptions
		checks  []string
	}{
		{
			name: "basic render",
			options: format.RenderOptions{
				SafeMode: false,
			},
			checks: []string{
				"test-root",
				"file1.txt",
				"subdir",
				"nested.go",
				"empty_dir",
				"no_annotation.md",
			},
		},
		{
			name: "with safe mode",
			options: format.RenderOptions{
				SafeMode: true,
			},
			checks: []string{
				"test-root",
				"file1.txt",
				"subdir",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := createTestTree()
			output, err := renderer.Render(tree, tt.options)

			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			// No color renderer shouldn't have ANSI escape codes
			if strings.Contains(output, "\x1b[") {
				t.Error("NoColorRenderer output contains ANSI escape codes")
			}

			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("Expected output to contain %q, got:\n%s", check, output)
				}
			}

			// Check format and description
			if renderer.Format() != format.FormatNoColor {
				t.Errorf("Format() = %v, want %v", renderer.Format(), format.FormatNoColor)
			}

			if !renderer.IsTerminalFormat() {
				t.Error("IsTerminalFormat() = false, want true")
			}

			desc := renderer.Description()
			if !strings.Contains(desc, "Plain text") {
				t.Errorf("Description() = %q, expected to contain 'Plain text'", desc)
			}
		})
	}
}

func TestEmptyTree(t *testing.T) {
	emptyRoot := &types.Node{
		Name:  "empty",
		IsDir: true,
	}

	renderers := []struct {
		name     string
		renderer format.Renderer
	}{
		{"ColorRenderer", &ColorRenderer{}},
		{"NoColorRenderer", &NoColorRenderer{}},
	}

	for _, r := range renderers {
		t.Run(r.name, func(t *testing.T) {
			output, err := r.renderer.Render(emptyRoot, format.RenderOptions{})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			if !strings.Contains(output, "empty") {
				t.Errorf("Expected output to contain root name 'empty', got:\n%s", output)
			}
		})
	}
}

func TestDeepNesting(t *testing.T) {
	// Create a deeply nested tree
	root := &types.Node{
		Name:  "root",
		IsDir: true,
	}

	current := root
	for i := 0; i < 5; i++ {
		child := &types.Node{
			Name:   strings.Repeat("level", i+1),
			IsDir:  true,
			Parent: current,
		}
		current.Children = []*types.Node{child}
		current = child
	}

	// Add a file at the deepest level
	current.Children = []*types.Node{
		{
			Name:   "deep.txt",
			IsDir:  false,
			Parent: current,
		},
	}

	renderer := &NoColorRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// Check that all levels are present
	if !strings.Contains(output, "level") {
		t.Error("Expected output to contain nested levels")
	}
	if !strings.Contains(output, "deep.txt") {
		t.Error("Expected output to contain deep.txt file")
	}
}

func TestAnnotations(t *testing.T) {
	root := &types.Node{
		Name:  "root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "annotated.txt",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "annotated.txt",
					Notes: "Full notes about the file",
				},
			},
			{
				Name:  "dir_with_annotation",
				IsDir: true,
				Annotation: &types.Annotation{
					Path:  "dir_with_annotation",
					Notes: "Directory description only",
				},
			},
		},
	}

	renderer := &NoColorRenderer{}
	output, err := renderer.Render(root, format.RenderOptions{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	// The output should contain the annotated items
	if !strings.Contains(output, "annotated.txt") {
		t.Error("Expected output to contain annotated.txt")
	}
	if !strings.Contains(output, "dir_with_annotation") {
		t.Error("Expected output to contain dir_with_annotation")
	}
}
