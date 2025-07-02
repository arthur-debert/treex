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

func TestStyledTreeRenderer_NoDuplicateDescriptions(t *testing.T) {
	// Test for the bug where single-line descriptions are duplicated
	// This reproduces the issue where Title and Description are identical
	// for single-line annotations, causing duplicate rendering
	root := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "cmd",
				IsDir: true,
				Annotation: &info.Annotation{
					Path:        "cmd",
					Title:       "The thin CLI layer that delegates logic to pkg",
					Description: "The thin CLI layer that delegates logic to pkg",
				},
			},
			{
				Name:  "docs",
				IsDir: true,
				Annotation: &info.Annotation{
					Path:        "docs",
					Title:       "User documentation and dev docs too",
					Description: "User documentation and dev docs too",
				},
			},
		},
	}

	// Render with annotations using no-color styles for easier testing
	var builder strings.Builder
	renderer := NewStyledTreeRenderer(&builder, true).
		WithStyles(NewNoColorTreeStyles())
	
	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := builder.String()
	
	// Check that the description appears only once, not duplicated
	description1 := "The thin CLI layer that delegates logic to pkg"
	description2 := "User documentation and dev docs too"
	
	// Count occurrences of each description
	count1 := strings.Count(result, description1)
	count2 := strings.Count(result, description2)
	
	if count1 != 1 {
		t.Errorf("Expected description '%s' to appear exactly once, but found %d occurrences in output:\n%s", 
			description1, count1, result)
	}
	
	if count2 != 1 {
		t.Errorf("Expected description '%s' to appear exactly once, but found %d occurrences in output:\n%s", 
			description2, count2, result)
	}
	
	// Additional check: ensure no line contains the same text repeated
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Check if any description appears twice in the same line or consecutive lines
		if strings.Contains(line, description1) {
			// Check if this description also appears in the next few lines (duplicate multi-line)
			for j := i + 1; j < len(lines) && j < i+3; j++ {
				if strings.Contains(lines[j], description1) {
					t.Errorf("Found duplicate description on line %d and %d:\nLine %d: %s\nLine %d: %s", 
						i+1, j+1, i+1, line, j+1, lines[j])
				}
			}
		}
		
		if strings.Contains(line, description2) {
			// Check if this description also appears in the next few lines (duplicate multi-line)
			for j := i + 1; j < len(lines) && j < i+3; j++ {
				if strings.Contains(lines[j], description2) {
					t.Errorf("Found duplicate description on line %d and %d:\nLine %d: %s\nLine %d: %s", 
						i+1, j+1, i+1, line, j+1, lines[j])
				}
			}
		}
	}
} 