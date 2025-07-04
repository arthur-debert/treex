package tui

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func TestStyledRenderer_NoDuplicateTitle(t *testing.T) {
	// Create a root node with a child that has annotations
	root := &tree.Node{
		Name:  "scripts",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "build",
				IsDir: false,
				Annotation: &info.Annotation{
					Title:       "Build script for treex",
					Description: "Build script for treex\nCompiles the treex binary and places it in bin/treex",
				},
				Children: []*tree.Node{},
			},
		},
	}

	// Test with styled renderer
	output, err := RenderStyledTreeToString(root, true)
	if err != nil {
		t.Fatalf("Failed to render tree: %v", err)
	}

	// Count occurrences of the title
	titleCount := strings.Count(output, "Build script for treex")
	if titleCount > 1 {
		t.Errorf("Title appears %d times, expected 1.\nOutput:\n%s", titleCount, output)
	}

	// The description line should appear exactly once
	descCount := strings.Count(output, "Compiles the treex binary")
	if descCount != 1 {
		t.Errorf("Description appears %d times, expected 1.\nOutput:\n%s", descCount, output)
	}
}

func TestPlainRenderer_NoDuplicateTitle(t *testing.T) {
	// Create a root node with a child that has annotations
	root := &tree.Node{
		Name:  "scripts",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "build",
				IsDir: false,
				Annotation: &info.Annotation{
					Title:       "Build script for treex",
					Description: "Build script for treex\nCompiles the treex binary and places it in bin/treex",
				},
				Children: []*tree.Node{},
			},
		},
	}

	// Test with plain renderer
	output, err := RenderTreeToString(root, true)
	if err != nil {
		t.Fatalf("Failed to render tree: %v", err)
	}

	// Count occurrences of the title
	titleCount := strings.Count(output, "Build script for treex")
	if titleCount > 1 {
		t.Errorf("Title appears %d times, expected 1.\nOutput:\n%s", titleCount, output)
	}

	// The description line should appear exactly once
	descCount := strings.Count(output, "Compiles the treex binary")
	if descCount != 1 {
		t.Errorf("Description appears %d times, expected 1.\nOutput:\n%s", descCount, output)
	}
}

func TestRenderer_ComplexAnnotations(t *testing.T) {
	// Test various annotation scenarios
	testCases := []struct {
		name        string
		annotation  *info.Annotation
		shouldShow  []string    // Lines that should appear
		shouldNotShow []string  // Lines that should NOT appear (duplicates)
	}{
		{
			name: "Title with multi-line description",
			annotation: &info.Annotation{
				Title:       "Main title",
				Description: "Main title\nSecond line\nThird line",
			},
			shouldShow: []string{"Main title", "Second line", "Third line"},
			shouldNotShow: []string{}, // No duplicates expected
		},
		{
			name: "Title only (same as description)",
			annotation: &info.Annotation{
				Title:       "Single line annotation",
				Description: "Single line annotation",
			},
			shouldShow: []string{"Single line annotation"},
			shouldNotShow: []string{}, // Should appear only once
		},
		{
			name: "No title, multi-line description",
			annotation: &info.Annotation{
				Title:       "",
				Description: "First line is the title\nSecond line is detail",
			},
			shouldShow: []string{"First line is the title", "Second line is detail"},
			shouldNotShow: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a root with the annotated child
			root := &tree.Node{
				Name:  "root",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:       "test",
						IsDir:      false,
						Annotation: tc.annotation,
						Children:   []*tree.Node{},
					},
				},
			}

			// Test styled renderer
			output, err := RenderStyledTreeToString(root, true)
			if err != nil {
				t.Fatalf("Failed to render tree: %v", err)
			}

			// Check that expected lines appear
			for _, line := range tc.shouldShow {
				count := strings.Count(output, line)
				if count != 1 {
					t.Errorf("Expected line '%s' to appear exactly once, found %d times.\nOutput:\n%s", 
						line, count, output)
				}
			}
		})
	}
}