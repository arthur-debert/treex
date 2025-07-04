package tui

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func TestRenderer_ConnectorAlignment(t *testing.T) {
	// Create a deeply nested tree structure
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "cmd",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "app",
						IsDir: true,
						Children: []*tree.Node{
							{
								Name:  "handler.go",
								IsDir: false,
								Annotation: &info.Annotation{
									Title:       "Request handler",
									Description: "Request handler\nHandles HTTP requests and routes them appropriately",
								},
							},
							{
								Name:  "subdir",
								IsDir: true,
								Children: []*tree.Node{
									{
										Name:  "util.go",
										IsDir: false,
										Annotation: &info.Annotation{
											Title:       "Utility functions",
											Description: "Utility functions\nHelper functions used throughout the application",
										},
									},
								},
							},
						},
					},
					{
						Name:  "root.go",
						IsDir: false,
						Annotation: &info.Annotation{
							Title:       "Root command",
							Description: "Root command\nSets up the CLI root command",
						},
					},
				},
			},
			{
				Name:  "main.go",
				IsDir: false,
				Annotation: &info.Annotation{
					Title:       "Main entry",
					Description: "Main entry\nBootstrap code",
				},
			},
		},
	}

	// Test with plain renderer
	output, err := RenderTreeToString(root, true)
	if err != nil {
		t.Fatalf("Failed to render tree: %v", err)
	}

	// Split output into lines for analysis
	lines := strings.Split(output, "\n")

	// Find the description lines and check their prefixes
	for i, line := range lines {
		// Check for handler.go's description line
		if strings.Contains(line, "Handles HTTP requests") {
			// Count the number of │ characters at the start
			prefix := extractPrefix(line)
			// handler.go is at cmd/app/handler.go
			// The file line shows: │   │   ├── handler.go
			// So the description continuation should be: │   │   │
			expectedConnectors := 3 // Should have 3 │ connectors
			actualConnectors := strings.Count(prefix, "│")
			
			if actualConnectors != expectedConnectors {
				t.Logf("DEBUG: Previous line: %s", lines[i-1])
				t.Errorf("handler.go description: expected %d │ connectors, got %d\nLine: %s\nPrefix: %q", 
					expectedConnectors, actualConnectors, line, prefix)
			}
		}

		// Check for util.go's description line
		if strings.Contains(line, "Helper functions used") {
			// Count the number of │ characters at the start
			prefix := extractPrefix(line)
			// util.go is at cmd/app/subdir/util.go
			// When rendered, subdir passes its continuation prefix to children
			// Since subdir is the last child of app, its continuation uses spaces
			// So util.go's description gets the prefix from subdir's context
			expectedConnectors := 2 // Has 2 │ connectors from the parent context
			actualConnectors := strings.Count(prefix, "│")
			
			if actualConnectors != expectedConnectors {
				t.Logf("DEBUG: Previous line: %s", lines[i-1])
				t.Errorf("util.go description: expected %d │ connectors, got %d\nLine: %s\nPrefix: %q", 
					expectedConnectors, actualConnectors, line, prefix)
			}
		}

		// Check for main.go's description line
		if strings.Contains(line, "Bootstrap code") && i > 0 {
			prefix := extractPrefix(line)
			expectedConnectors := 0 // At root level, no │ connectors
			actualConnectors := strings.Count(prefix, "│")
			
			if actualConnectors != expectedConnectors {
				t.Errorf("main.go description: expected %d │ connectors, got %d\nLine: %s\nPrefix: %q", 
					expectedConnectors, actualConnectors, line, prefix)
			}
		}
	}
}

func TestStyledRenderer_ConnectorAlignment(t *testing.T) {
	// Create a tree with multiple levels
	root := &tree.Node{
		Name:  "project",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "src",
				IsDir: true,
				Children: []*tree.Node{
					{
						Name:  "core",
						IsDir: true,
						Children: []*tree.Node{
							{
								Name:  "engine.go",
								IsDir: false,
								Annotation: &info.Annotation{
									Title:       "Core engine",
									Description: "Core engine\nThe main processing engine\nHandles all core operations",
								},
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
							Title:       "API documentation",
							Description: "API documentation\nComprehensive API reference",
						},
					},
				},
			},
		},
	}

	// Test with styled renderer
	output, err := RenderStyledTreeToString(root, true)
	if err != nil {
		t.Fatalf("Failed to render tree: %v", err)
	}

	// Remove ANSI codes for easier testing
	cleanOutput := stripANSIForTest(output)
	lines := strings.Split(cleanOutput, "\n")

	// Check engine.go's multi-line description
	for _, line := range lines {
		if strings.Contains(line, "The main processing engine") {
			prefix := extractPrefix(line)
			// Should maintain tree structure with proper indentation
			if !strings.Contains(prefix, "│") || len(prefix) < 2 {
				t.Errorf("Multi-line description lost tree structure.\nLine: %s\nPrefix: %q", line, prefix)
			}
		}
		
		// Check the third line of engine.go description
		if strings.Contains(line, "Handles all core operations") {
			prefix := extractPrefix(line)
			// Count │ characters - should match the tree depth
			// engine.go is at project/src/core/engine.go
			// Since core is the last child of src (└──), its continuation only has 1 │ from project level
			connectorCount := strings.Count(prefix, "│")
			expectedCount := 1
			if connectorCount != expectedCount {
				t.Errorf("Third line of description has incorrect connectors: expected %d, got %d\nLine: %s", 
					expectedCount, connectorCount, line)
			}
		}
	}
}

// extractPrefix extracts the prefix (indentation and tree characters) from a line
func extractPrefix(line string) string {
	// Find where the actual content starts (first alphanumeric character)
	for i, ch := range line {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			return line[:i]
		}
	}
	return line
}

// stripANSIForTest removes ANSI escape codes for testing
func stripANSIForTest(text string) string {
	// More comprehensive ANSI stripping that handles various escape sequences
	var result strings.Builder
	i := 0
	
	for i < len(text) {
		// Check for ESC character (0x1B)
		if i < len(text)-1 && text[i] == '\x1b' {
			// Skip the ESC character
			i++
			
			// Handle CSI sequences (ESC[)
			if i < len(text) && text[i] == '[' {
				i++ // Skip [
				// Skip until we find a letter (command character)
				for i < len(text) {
					ch := text[i]
					i++
					// CSI sequences end with a letter
					if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
						break
					}
				}
			} else if i < len(text) && (text[i] == '(' || text[i] == ')') {
				// Handle other escape sequences like ESC( or ESC)
				i++ // Skip the ( or )
				if i < len(text) {
					i++ // Skip the next character
				}
			}
		} else {
			// Regular character, add to result
			result.WriteByte(text[i])
			i++
		}
	}
	
	return result.String()
}