package commands

import (
	"testing"

	"github.com/adebert/treex/pkg/core/types"
)

func TestBuildVirtualTree(t *testing.T) {
	// Test basic functionality
	annotations := map[string]*types.Annotation{
		"Dad": {Path: "Dad", Notes: "Chill, dad"},
		"Mom": {Path: "Mom", Notes: "Listen to your mother"},
		"kids/Sam": {Path: "kids/Sam", Notes: "Little Sam"},
	}

		tree, err := BuildVirtualTree(annotations)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if tree == nil {
		t.Fatal("Expected tree but got nil")
	}

	if !tree.IsDir {
		t.Error("Expected root to be a directory")
	}

	if tree.Name != "root" {
		t.Errorf("Expected root name to be 'root', got '%s'", tree.Name)
	}

	// Should have 3 children: Dad, Mom, kids
	if len(tree.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(tree.Children))
	}

	// Test with empty annotations
	emptyTree, err := BuildVirtualTree(map[string]*types.Annotation{})
	if err == nil {
		t.Error("Expected error for empty annotations but got none")
	}
	if emptyTree != nil {
		t.Error("Expected nil tree for empty annotations")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty format",
			input:    "",
			expected: "",
		},
		{
			name:     "color format",
			input:    "color",
			expected: "color",
		},
		{
			name:     "no-color format",
			input:    "no-color",
			expected: "no-color",
		},
		{
			name:     "markdown format",
			input:    "markdown",
			expected: "markdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFormat(tt.input)
			if string(result) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}

func TestDrawCommandFlags(t *testing.T) {
	// Test that the command has the expected flags
	cmd := drawCmd

	if cmd.Use != "draw [--info-file FILE | -]" {
		t.Errorf("Expected Use to be 'draw [--info-file FILE | -]', got %s", cmd.Use)
	}

	if cmd.Short != "Draw tree diagrams from info files without filesystem validation" {
		t.Errorf("Expected Short to be 'Draw tree diagrams from info files without filesystem validation', got %s", cmd.Short)
	}

	// Check that flags are properly defined
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Error("Expected format flag to be defined")
	}

	infoFileFlag := cmd.Flags().Lookup("info-file")
	if infoFileFlag == nil {
		t.Error("Expected info-file flag to be defined")
	}
}