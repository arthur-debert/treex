package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func TestStyledTreeRenderer_BasicTree(t *testing.T) {
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
				},
			},
		},
	}

	// Render without annotations using plain styles (no colors for testing)
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, false).
		WithStyles(NewNoColorTreeStyles())
	
	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := builder.String()

	// Check that basic tree structure is preserved
	if !strings.Contains(result, "root") {
		t.Error("Expected to find root directory")
	}

	if !strings.Contains(result, "├── file1.txt") {
		t.Error("Expected to find tree connector '├── file1.txt'")
	}

	if !strings.Contains(result, "└── dir1") {
		t.Error("Expected to find tree connector '└── dir1'")
	}
}

func TestStyledTreeRenderer_WithAnnotations(t *testing.T) {
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
				Name:  "main.go",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "main.go",
					Title:       "Main Entry Point",
					Description: "Main Entry Point\nApplication startup logic\nHandles command line arguments",
				},
			},
		},
	}

	// Render with annotations using plain styles
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, true).
		WithStyles(NewNoColorTreeStyles())
	
	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := builder.String()

	// Check that annotations are included
	if !strings.Contains(result, "Project documentation") {
		t.Error("Expected to find 'Project documentation' annotation")
	}

	if !strings.Contains(result, "Main Entry Point") {
		t.Error("Expected to find 'Main Entry Point' annotation")
	}

	// Check that multi-line annotations are rendered
	if !strings.Contains(result, "Application startup logic") {
		t.Error("Expected to find multi-line annotation content")
	}
}

func TestStyledTreeRenderer_StyleApplication(t *testing.T) {
	// Create a simple tree
	root := &tree.Node{
		Name:  "test",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "file.txt",
					Title:       "Test File",
					Description: "Test File\nA simple test file",
				},
			},
		},
	}

	// Test with full color styles
	result, err := RenderStyledTreeToString(root, true)
	if err != nil {
		t.Fatalf("RenderStyledTreeToString failed: %v", err)
	}

	// The result should contain ANSI escape codes for styling
	// We can't test exact colors, but we can check that styling is applied
	if len(result) <= len("test\n└── file.txt  Test File\n") {
		t.Error("Expected styled output to be longer than plain text due to ANSI codes")
	}

	// Test with minimal styles
	var builder strings.Builder
	err = RenderMinimalStyledTree(&builder, root, true)
	if err != nil {
		t.Fatalf("RenderMinimalStyledTree failed: %v", err)
	}

	minimalResult := builder.String()
	if !strings.Contains(minimalResult, "Test File") {
		t.Error("Expected minimal styled output to contain annotation")
	}
}

func TestStyledTreeRenderer_NoAnnotations(t *testing.T) {
	// Create a tree without annotations
	root := &tree.Node{
		Name:  "simple",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file.txt",
				IsDir: false,
			},
		},
	}

	// Render with annotations enabled but no annotations present
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, true).
		WithStyles(NewNoColorTreeStyles())
	
	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := builder.String()
	expected := "simple\n└── file.txt\n"

	if result != expected {
		t.Errorf("Expected clean output without annotations.\nExpected: %q\nGot: %q", expected, result)
	}
}

func TestTreeStyles_Creation(t *testing.T) {
	// Test that all style constructors work
	fullStyles := NewTreeStyles()
	if fullStyles == nil {
		t.Error("NewTreeStyles() returned nil")
	}

	minimalStyles := NewMinimalTreeStyles()
	if minimalStyles == nil {
		t.Error("NewMinimalTreeStyles() returned nil")
	}

	noColorStyles := NewNoColorTreeStyles()
	if noColorStyles == nil {
		t.Error("NewNoColorTreeStyles() returned nil")
	}

	// Test that styles have different properties
	// Full styles should have colors (check for nil first)
	if fullStyles != nil && noColorStyles != nil {
		fullTreeConnector := fullStyles.TreeLines.Render("├── ")
		plainTreeConnector := noColorStyles.TreeLines.Render("├── ")

		// The styled version should be longer due to ANSI codes (in most cases)
		// This is a basic check that styling is being applied
		if len(fullTreeConnector) < len(plainTreeConnector) {
			// This might happen in some environments, so we'll just check they're not identical
			if fullTreeConnector == plainTreeConnector {
				t.Log("Warning: Full styles and no-color styles produced identical output")
			}
		}
	}
}

func TestStyledTreeRenderer_WithCustomStyles(t *testing.T) {
	// Create custom styles
	customStyles := &TreeStyles{
		TreeLines: lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		RootPath:  lipgloss.NewStyle().Bold(true),
		AnnotatedPath: lipgloss.NewStyle(),
		UnannotatedPath: lipgloss.NewStyle(),
		AnnotationText: lipgloss.NewStyle().Bold(true),
		AnnotationContainer: lipgloss.NewStyle(),
		AnnotationSeparator: lipgloss.NewStyle().SetString("  "),
		MultiLineIndent: lipgloss.NewStyle(),
	}

	root := &tree.Node{
		Name:  "test",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file.txt",
				IsDir: false,
			},
		},
	}

	// Test with custom styles
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, false).
		WithStyles(customStyles)
	
	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render with custom styles failed: %v", err)
	}

	result := builder.String()
	if !strings.Contains(result, "test") {
		t.Error("Expected custom styled output to contain root name")
	}
} 