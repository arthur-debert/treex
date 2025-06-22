package tui

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func TestTreeRenderer_BasicTree(t *testing.T) {
	// Create a simple test tree
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
						Name:  "file3.txt",
						IsDir: false,
					},
				},
			},
			{
				Name:  "file4.txt",
				IsDir: false,
			},
		},
	}

	// Render without annotations
	result, err := RenderTreeToString(root, false)
	if err != nil {
		t.Fatalf("RenderTreeToString failed: %v", err)
	}

	expected := `root
├── file1.txt
├── dir1
│   ├── file2.txt
│   └── file3.txt
└── file4.txt
`

	if result != expected {
		t.Errorf("Tree rendering mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestTreeRenderer_WithAnnotations(t *testing.T) {
	// Create a test tree with annotations
	root := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "README.md",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "README.md",
					Description: "Project documentation",
				},
			},
			{
				Name:  "src",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "main.go",
						IsDir: false,
						Annotation: &info.Annotation{
							Path:        "src/main.go",
							Title:       "Main Entry Point",
							Description: "Main Entry Point\nApplication startup logic",
						},
					},
				},
			},
			{
				Name:  "LICENSE",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "LICENSE",
					Description: "MIT License",
				},
			},
		},
	}

	// Render with annotations
	result, err := RenderTreeToString(root, true)
	if err != nil {
		t.Fatalf("RenderTreeToString failed: %v", err)
	}

	// Check that annotations are included
	if !strings.Contains(result, "Project documentation") {
		t.Error("Expected to find 'Project documentation' annotation")
	}

	if !strings.Contains(result, "Main Entry Point") {
		t.Error("Expected to find 'Main Entry Point' annotation")
	}

	if !strings.Contains(result, "MIT License") {
		t.Error("Expected to find 'MIT License' annotation")
	}

	// Check that tree structure is preserved
	if !strings.Contains(result, "├── README.md") {
		t.Error("Expected to find tree connector '├── README.md'")
	}

	if !strings.Contains(result, "└── LICENSE") {
		t.Error("Expected to find tree connector '└── LICENSE'")
	}
}

func TestTreeRenderer_EmptyTree(t *testing.T) {
	// Create an empty root
	root := &tree.Node{
		Name:     "empty",
		IsDir:    true,
		Children: []*tree.Node{},
	}

	result, err := RenderTreeToString(root, false)
	if err != nil {
		t.Fatalf("RenderTreeToString failed: %v", err)
	}

	expected := "empty\n"
	if result != expected {
		t.Errorf("Empty tree rendering mismatch.\nExpected: %q\nGot: %q", expected, result)
	}
}

func TestTreeRenderer_SingleFile(t *testing.T) {
	// Create a root with a single file
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "single.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "single.txt",
					Description: "A single file",
				},
			},
		},
	}

	result, err := RenderTreeToString(root, true)
	if err != nil {
		t.Fatalf("RenderTreeToString failed: %v", err)
	}

	// Should use └── for the last (and only) child
	if !strings.Contains(result, "└── single.txt") {
		t.Error("Expected to find '└── single.txt' for single child")
	}

	if !strings.Contains(result, "A single file") {
		t.Error("Expected to find annotation 'A single file'")
	}
}

func TestTreeRenderer_DeepNesting(t *testing.T) {
	// Create a deeply nested structure
	root := &tree.Node{
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
								Name:  "deep.txt",
								IsDir: false,
							},
						},
					},
				},
			},
		},
	}

	result, err := RenderTreeToString(root, false)
	if err != nil {
		t.Fatalf("RenderTreeToString failed: %v", err)
	}

	expected := `root
└── level1
    └── level2
        └── deep.txt
`

	if result != expected {
		t.Errorf("Deep nesting rendering mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestFormatAnnotation(t *testing.T) {
	renderer := NewTreeRenderer(&strings.Builder{}, true)

	// Test with title
	annotationWithTitle := &info.Annotation{
		Path:        "test.txt",
		Title:       "Test Title",
		Description: "Test Title\nDetailed description",
	}

	result := renderer.formatAnnotation(annotationWithTitle)
	if result != "Test Title" {
		t.Errorf("Expected 'Test Title', got %q", result)
	}

	// Test without title
	annotationWithoutTitle := &info.Annotation{
		Path:        "test.txt",
		Description: "First line\nSecond line",
	}

	result = renderer.formatAnnotation(annotationWithoutTitle)
	if result != "First line" {
		t.Errorf("Expected 'First line', got %q", result)
	}

	// Test with nil annotation
	result = renderer.formatAnnotation(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil annotation, got %q", result)
	}
} 