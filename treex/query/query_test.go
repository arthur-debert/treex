// see docs/dev/architecture.txt - Phase 5: User Queries
package query_test

import (
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/jwaldrip/treex/treex/query"
	"github.com/jwaldrip/treex/treex/types"
)

func TestProcessor_EmptyQueries(t *testing.T) {
	processor := query.NewProcessor()

	// Create a simple test tree
	root := &types.Node{
		Name:  "root",
		Path:  ".",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file.txt",
				Path:  "file.txt",
				IsDir: false,
			},
		},
	}

	result := processor.Process(root)

	// With no queries, should return the original tree
	if result == nil {
		t.Fatal("Expected non-nil result with empty queries")
	}
	if result.Name != "root" {
		t.Errorf("Expected root name 'root', got %q", result.Name)
	}
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(result.Children))
	}
}

func TestProcessor_WithQueries(t *testing.T) {
	processor := query.NewProcessor()

	// Add a path query that matches .go files anywhere
	processor.AddQuery(query.NewPathQuery("**/*.go"))

	// Create test tree with mixed file types
	root := &types.Node{
		Name:  "project",
		Path:  "project",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "main.go",
				Path:  "project/main.go",
				IsDir: false,
			},
			{
				Name:  "README.md",
				Path:  "project/README.md",
				IsDir: false,
			},
			{
				Name:  "utils.go",
				Path:  "project/utils.go",
				IsDir: false,
			},
		},
	}

	result := processor.Process(root)

	// Should filter to only .go files
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Name != "project" {
		t.Errorf("Expected root name 'project', got %q", result.Name)
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 children (.go files), got %d", len(result.Children))
	}

	// Verify only .go files remain
	for _, child := range result.Children {
		if child.Name != "main.go" && child.Name != "utils.go" {
			t.Errorf("Unexpected child: %q", child.Name)
		}
	}
}

func TestProcessor_NilTree(t *testing.T) {
	processor := query.NewProcessor()
	processor.AddQuery(query.NewPathQuery("*.go"))

	result := processor.Process(nil)
	if result != nil {
		t.Error("Expected nil result for nil input")
	}
}

func TestBuilder_FluentInterface(t *testing.T) {
	processor := query.NewBuilder().
		WithPathPattern("**/src/**/*.go").
		WithNamePattern("test*").
		Build()

	// Create test tree
	root := &types.Node{
		Name:  "project",
		Path:  "project",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "src",
				Path:  "project/src",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "test_main.go",
						Path:  "project/src/test_main.go",
						IsDir: false,
					},
					{
						Name:  "main.go",
						Path:  "project/src/main.go",
						IsDir: false,
					},
				},
			},
		},
	}

	// Set parent relationships
	root.Children[0].Parent = root
	for _, child := range root.Children[0].Children {
		child.Parent = root.Children[0]
	}

	result := processor.Process(root)

	// Should only include test_main.go (matches both path and name patterns)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Navigate to src directory
	if len(result.Children) != 1 || result.Children[0].Name != "src" {
		t.Fatal("Expected src directory")
	}

	srcDir := result.Children[0]
	if len(srcDir.Children) != 1 {
		t.Errorf("Expected 1 file in src, got %d", len(srcDir.Children))
	}

	if srcDir.Children[0].Name != "test_main.go" {
		t.Errorf("Expected test_main.go, got %q", srcDir.Children[0].Name)
	}
}

func TestBuilder_EmptyPatterns(t *testing.T) {
	processor := query.NewBuilder().
		WithPathPattern(""). // Empty pattern should be ignored
		WithNamePattern(""). // Empty pattern should be ignored
		Build()

	root := &types.Node{
		Name:  "test",
		Path:  "test",
		IsDir: true,
		Children: []*types.Node{
			{Name: "file.txt", Path: "test/file.txt", IsDir: false},
		},
	}

	result := processor.Process(root)

	// With no actual queries (empty patterns ignored), should return original tree
	if result == nil || len(result.Children) != 1 {
		t.Error("Expected original tree with empty patterns")
	}
}

func TestQueryProcessor_ParentChildRelationships(t *testing.T) {
	processor := query.NewProcessor()
	processor.AddQuery(query.NewNameQuery("*.txt"))

	// Create a deeper tree structure
	root := &types.Node{
		Name:  "root",
		Path:  "root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "subdir",
				Path:  "root/subdir",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "file.txt",
						Path:  "root/subdir/file.txt",
						IsDir: false,
					},
					{
						Name:  "other.md",
						Path:  "root/subdir/other.md",
						IsDir: false,
					},
				},
			},
		},
	}

	// Set original parent relationships
	root.Children[0].Parent = root
	for _, child := range root.Children[0].Children {
		child.Parent = root.Children[0]
	}

	result := processor.Process(root)

	// Verify parent-child relationships are maintained
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Should have subdir (because it contains matching file)
	if len(result.Children) != 1 || result.Children[0].Name != "subdir" {
		t.Fatal("Expected subdir in result")
	}

	subdir := result.Children[0]
	if subdir.Parent != result {
		t.Error("Parent relationship not properly set for subdir")
	}

	// Should only have file.txt
	if len(subdir.Children) != 1 || subdir.Children[0].Name != "file.txt" {
		t.Fatal("Expected only file.txt in subdir")
	}

	file := subdir.Children[0]
	if file.Parent != subdir {
		t.Error("Parent relationship not properly set for file")
	}
}

func TestProcessor_ComplexTreeFiltering(t *testing.T) {
	// Test with a more complex tree structure from testutil
	fs := testutil.NewTestFS()

	// Create a complex project structure
	fs.MustCreateTree("/project", map[string]interface{}{
		"src": map[string]interface{}{
			"main.go":      "package main",
			"utils.go":     "package utils",
			"test_main.go": "package main // test",
		},
		"docs": map[string]interface{}{
			"README.md":   "# Project",
			"guide.txt":   "Guide content",
			"test_doc.md": "# Test documentation",
		},
		"config": map[string]interface{}{
			"app.yaml":  "config: value",
			"test.yaml": "test: config",
		},
	})

	// Build tree using testutil
	tree := testutil.MustBuildFileTree(fs, map[string]interface{}{
		"root": "/project",
	})

	// Test path pattern query
	processor := query.NewBuilder().
		WithPathPattern("**/test*").
		Build()

	result := processor.Process(tree)

	// Should include directories that contain matching files
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Find matching files using helper
	matchingFiles := findMatchingFiles(result, "test")
	expectedFiles := []string{"test_main.go", "test_doc.md", "test.yaml"}

	if len(matchingFiles) != len(expectedFiles) {
		t.Errorf("Expected %d matching files, got %d: %v",
			len(expectedFiles), len(matchingFiles), matchingFiles)
	}

	for _, expected := range expectedFiles {
		found := false
		for _, actual := range matchingFiles {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %q not found in results", expected)
		}
	}
}

// Helper function to find all files with a substring in their name
func findMatchingFiles(node *types.Node, substring string) []string {
	var files []string

	var walk func(*types.Node)
	walk = func(n *types.Node) {
		if n == nil {
			return
		}

		if !n.IsDir && len(n.Name) > 0 {
			// Check if the filename contains the substring
			for i := 0; i <= len(n.Name)-len(substring); i++ {
				if n.Name[i:i+len(substring)] == substring {
					files = append(files, n.Name)
					break
				}
			}
		}

		for _, child := range n.Children {
			walk(child)
		}
	}

	walk(node)
	return files
}
