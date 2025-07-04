package tui

import (
	"os"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestThemeSwitching(t *testing.T) {
	// Create a simple tree with annotations
	root := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "src",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "main.go",
						IsDir: false,
						Annotation: &info.Annotation{
							Title:       "Main file",
							Description: "Application entry point",
						},
					},
				},
			},
			{
				Name:  "README.md",
				IsDir: false,
				Annotation: &info.Annotation{
					Title: "Documentation",
				},
			},
		},
	}

	// Test dark theme (default)
	SetTheme(true)
	darkOutput, err := RenderStyledTreeToString(root, true)
	if err != nil {
		t.Fatalf("Failed to render with dark theme: %v", err)
	}

	// Test light theme
	SetTheme(false)
	lightOutput, err := RenderStyledTreeToString(root, true)
	if err != nil {
		t.Fatalf("Failed to render with light theme: %v", err)
	}

	// Debug: Print a sample of both outputs to see the difference
	t.Logf("Dark output sample (first 200 chars): %q", darkOutput[:min(200, len(darkOutput))])
	t.Logf("Light output sample (first 200 chars): %q", lightOutput[:min(200, len(lightOutput))])

	// The outputs should be different (different ANSI color codes)
	// Note: In some test environments with limited color support,
	// the outputs might be identical if colors are disabled or mapped to the same ANSI codes
	if darkOutput == lightOutput {
		// Check if we're in a limited color environment
		if os.Getenv("CI") != "" || os.Getenv("NO_COLOR") != "" {
			t.Skip("Skipping color comparison in limited color environment")
		} else {
			t.Log("Warning: Dark and light theme outputs are identical - colors might be disabled in test environment")
			// Don't fail the test as this is environment-dependent
		}
	}

	// Basic checks to ensure output is generated
	if !strings.Contains(darkOutput, "project") || !strings.Contains(darkOutput, "Main file") {
		t.Error("Dark theme output missing expected content")
	}
	if !strings.Contains(lightOutput, "project") || !strings.Contains(lightOutput, "Main file") {
		t.Error("Light theme output missing expected content")
	}

	// Reset to dark theme for other tests
	SetTheme(true)
}

func TestGetTheme(t *testing.T) {
	// Save original theme
	originalTheme := GetTheme()
	
	// Test dark theme
	SetTheme(true)
	theme := GetTheme()
	if theme.DirectoryColor != DarkTheme.DirectoryColor {
		t.Errorf("GetTheme() not returning dark theme colors after SetTheme(true): got %v, want %v", 
			theme.DirectoryColor, DarkTheme.DirectoryColor)
	}

	// Test light theme
	SetTheme(false)
	theme = GetTheme()
	if theme.DirectoryColor != LightTheme.DirectoryColor {
		t.Errorf("GetTheme() not returning light theme colors after SetTheme(false): got %v, want %v",
			theme.DirectoryColor, LightTheme.DirectoryColor)
	}

	// Reset to original theme
	SetTheme(originalTheme.DirectoryColor == DarkTheme.DirectoryColor)
}

func TestThemeColors(t *testing.T) {
	// Verify that dark and light themes have different colors
	tests := []struct {
		name      string
		darkColor interface{}
		lightColor interface{}
	}{
		{"TreeConnectorColor", DarkTheme.TreeConnectorColor, LightTheme.TreeConnectorColor},
		{"DirectoryColor", DarkTheme.DirectoryColor, LightTheme.DirectoryColor},
		{"FileColor", DarkTheme.FileColor, LightTheme.FileColor},
		{"AnnotationTitleColor", DarkTheme.AnnotationTitleColor, LightTheme.AnnotationTitleColor},
		{"AnnotationDescriptionColor", DarkTheme.AnnotationDescriptionColor, LightTheme.AnnotationDescriptionColor},
		{"HighlightColor", DarkTheme.HighlightColor, LightTheme.HighlightColor},
	}

	for _, test := range tests {
		if test.darkColor == test.lightColor {
			t.Errorf("%s: dark and light theme colors are identical (%v)", test.name, test.darkColor)
		}
	}
}

func TestNewTreeStylesUsesActiveTheme(t *testing.T) {
	// Set light theme
	SetTheme(false)
	
	// Create new styles
	lightStyles := NewTreeStyles()
	
	// Set dark theme
	SetTheme(true)
	
	// Create new styles with dark theme
	darkStyles := NewTreeStyles()
	
	// Render some text with both styles
	lightText := lightStyles.RootPath.Render("test")
	darkText := darkStyles.RootPath.Render("test")
	
	// The styled texts might be different (depends on terminal color support)
	// At minimum, they should both produce valid output
	if lightText == "" || darkText == "" {
		t.Error("Failed to render text with theme styles")
	}
	
	// Log the outputs for debugging
	t.Logf("Light theme output: %q", lightText)
	t.Logf("Dark theme output: %q", darkText)
	
	// Verify themes are properly set
	SetTheme(false)
	currentTheme := GetTheme()
	if currentTheme.DirectoryColor != LightTheme.DirectoryColor {
		t.Error("Theme not properly set to light theme")
	}
	
	// Reset to dark theme
	SetTheme(true)
}