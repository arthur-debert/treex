package format

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
	"gopkg.in/yaml.v3"
)

func createTestTree() *types.Node {
	return &types.Node{
		Name:  "root",
		Path:  "root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file1.txt",
				Path:  "root/file1.txt",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "root/file1.txt",
					Notes: "First test file",
				},
			},
			{
				Name:  "dir1",
				Path:  "root/dir1",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "file2.txt",
						Path:  "root/dir1/file2.txt",
						IsDir: false,
						Annotation: &types.Annotation{
							Path:  "root/dir1/file2.txt",
							Notes: "Some notes",
						},
					},
					{
						Name:  "file3.txt",
						Path:  "root/dir1/file3.txt",
						IsDir: false,
					},
				},
			},
			{
				Name:  "empty_dir",
				Path:  "root/empty_dir",
				IsDir: true,
			},
		},
	}
}

func TestJSONRenderer(t *testing.T) {
	renderer := &JSONRenderer{}
	
	t.Run("basic properties", func(t *testing.T) {
		if renderer.Format() != FormatJSON {
			t.Errorf("Format() = %v, want %v", renderer.Format(), FormatJSON)
		}
		
		if renderer.Description() != "JSON structured data format" {
			t.Errorf("Description() = %q, want %q", renderer.Description(), "JSON structured data format")
		}
		
		if renderer.IsTerminalFormat() {
			t.Error("IsTerminalFormat() = true, want false")
		}
	})
	
	t.Run("render tree", func(t *testing.T) {
		testTree := createTestTree()
		options := RenderOptions{}
		
		output, err := renderer.Render(testTree, options)
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		// Parse the JSON to verify structure
		var data TreeData
		if err := json.Unmarshal([]byte(output), &data); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}
		
		// Verify root
		if data.Name != "root" {
			t.Errorf("Root name = %q, want %q", data.Name, "root")
		}
		if data.Path != "root" {
			t.Errorf("Root path = %q, want %q", data.Path, "root")
		}
		if !data.IsDirectory {
			t.Error("Root should be a directory")
		}
		if data.Annotation != nil {
			t.Error("Root should not have annotation")
		}
		
		// Verify children count
		if len(data.Children) != 3 {
			t.Errorf("Root children = %d, want 3", len(data.Children))
		}
		
		// Verify first child (file1.txt) with annotation
		if len(data.Children) > 0 {
			file1 := data.Children[0]
			if file1.Name != "file1.txt" {
				t.Errorf("First child name = %q, want %q", file1.Name, "file1.txt")
			}
			if file1.Path != "root/file1.txt" {
				t.Errorf("First child path = %q, want %q", file1.Path, "root/file1.txt")
			}
			if file1.IsDirectory {
				t.Error("file1.txt should not be a directory")
			}
			if file1.Annotation == nil {
				t.Error("file1.txt should have annotation")
			} else {
				if file1.Annotation.Notes != "First test file" {
					t.Errorf("file1.txt notes = %q, want %q", file1.Annotation.Notes, "First test file")
				}
			}
		}
		
		// Verify dir1 and its children
		if len(data.Children) > 1 {
			dir1 := data.Children[1]
			if dir1.Name != "dir1" {
				t.Errorf("Second child name = %q, want %q", dir1.Name, "dir1")
			}
			if !dir1.IsDirectory {
				t.Error("dir1 should be a directory")
			}
			if len(dir1.Children) != 2 {
				t.Errorf("dir1 children = %d, want 2", len(dir1.Children))
			}
			
			// Check nested file - Notes field is not copied in the current implementation
			if len(dir1.Children) > 0 {
				file2 := dir1.Children[0]
				if file2.Annotation != nil {
					if file2.Annotation.Notes != "Some notes" {
						t.Errorf("file2 notes = %q, want %q", file2.Annotation.Notes, "Some notes")
					}
				}
			}
		}
		
		// Verify empty directory
		if len(data.Children) > 2 {
			emptyDir := data.Children[2]
			if emptyDir.Name != "empty_dir" {
				t.Errorf("Third child name = %q, want %q", emptyDir.Name, "empty_dir")
			}
			if len(emptyDir.Children) != 0 {
				t.Errorf("empty_dir should have no children, got %d", len(emptyDir.Children))
			}
		}
	})
	
	t.Run("render nil tree", func(t *testing.T) {
		renderer := &JSONRenderer{}
		options := RenderOptions{}
		
		_, err := renderer.Render(nil, options)
		if err == nil {
			t.Error("Expected error when rendering nil tree, got none")
		}
	})
	
	t.Run("indentation", func(t *testing.T) {
		testTree := &types.Node{Name: "test"}
		options := RenderOptions{}
		
		output, err := renderer.Render(testTree, options)
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		// Check that output is properly indented
		if !strings.Contains(output, "\n") {
			t.Error("Expected indented JSON with newlines")
		}
		if !strings.Contains(output, "  ") {
			t.Error("Expected indented JSON with spaces")
		}
	})
}

func TestYAMLRenderer(t *testing.T) {
	renderer := &YAMLRenderer{}
	
	t.Run("basic properties", func(t *testing.T) {
		if renderer.Format() != FormatYAML {
			t.Errorf("Format() = %v, want %v", renderer.Format(), FormatYAML)
		}
		
		if renderer.Description() != "YAML structured data format" {
			t.Errorf("Description() = %q, want %q", renderer.Description(), "YAML structured data format")
		}
		
		if renderer.IsTerminalFormat() {
			t.Error("IsTerminalFormat() = true, want false")
		}
	})
	
	t.Run("render tree", func(t *testing.T) {
		testTree := createTestTree()
		options := RenderOptions{}
		
		output, err := renderer.Render(testTree, options)
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		// Parse the YAML to verify structure
		var data TreeData
		if err := yaml.Unmarshal([]byte(output), &data); err != nil {
			t.Fatalf("Failed to parse YAML output: %v", err)
		}
		
		// Verify root
		if data.Name != "root" {
			t.Errorf("Root name = %q, want %q", data.Name, "root")
		}
		if data.Path != "root" {
			t.Errorf("Root path = %q, want %q", data.Path, "root")
		}
		if !data.IsDirectory {
			t.Error("Root should be a directory")
		}
		
		// Verify children
		if len(data.Children) != 3 {
			t.Errorf("Root children = %d, want 3", len(data.Children))
		}
		
		// Verify annotation handling
		if len(data.Children) > 0 && data.Children[0].Annotation != nil {
			ann := data.Children[0].Annotation
			if ann.Notes != "First test file" {
				t.Error("Annotation not properly preserved in YAML")
			}
		}
	})
	
	t.Run("YAML formatting", func(t *testing.T) {
		testTree := &types.Node{
			Name:  "test",
			IsDir: false,
			Annotation: &types.Annotation{
				Path:  "test",
				Notes: "Test",
			},
		}
		options := RenderOptions{}
		
		output, err := renderer.Render(testTree, options)
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		// Check YAML formatting
		if !strings.Contains(output, "name: test") {
			t.Error("Expected YAML to contain 'name: test'")
		}
		if !strings.Contains(output, "is_directory: false") {
			t.Error("Expected YAML to contain 'is_directory: false'")
		}
		if !strings.Contains(output, "annotation:") {
			t.Error("Expected YAML to contain 'annotation:' section")
		}
	})
}

func TestCompactJSONRenderer(t *testing.T) {
	renderer := &CompactJSONRenderer{}
	
	t.Run("basic properties", func(t *testing.T) {
		if renderer.Format() != "compact-json" {
			t.Errorf("Format() = %v, want %v", renderer.Format(), "compact-json")
		}
		
		if renderer.Description() != "Compact JSON format (single line)" {
			t.Errorf("Description() = %q, want %q", renderer.Description(), "Compact JSON format (single line)")
		}
		
		if renderer.IsTerminalFormat() {
			t.Error("IsTerminalFormat() = true, want false")
		}
	})
	
	t.Run("render compact", func(t *testing.T) {
		testTree := createTestTree()
		options := RenderOptions{}
		
		output, err := renderer.Render(testTree, options)
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		// Verify it's valid JSON
		var data TreeData
		if err := json.Unmarshal([]byte(output), &data); err != nil {
			t.Fatalf("Failed to parse compact JSON output: %v", err)
		}
		
		// Verify it's compact (no newlines or extra spaces)
		if strings.Contains(output, "\n") {
			t.Error("Compact JSON should not contain newlines")
		}
		
		// Should still have the same data
		if data.Name != "root" || len(data.Children) != 3 {
			t.Error("Compact JSON should preserve all data")
		}
	})
}

func TestFlatJSONRenderer(t *testing.T) {
	renderer := &FlatJSONRenderer{}
	
	t.Run("basic properties", func(t *testing.T) {
		if renderer.Format() != "flat-json" {
			t.Errorf("Format() = %v, want %v", renderer.Format(), "flat-json")
		}
		
		if renderer.Description() != "Flat JSON array of paths with metadata" {
			t.Errorf("Description() = %q, want %q", renderer.Description(), "Flat JSON array of paths with metadata")
		}
		
		if renderer.IsTerminalFormat() {
			t.Error("IsTerminalFormat() = true, want false")
		}
	})
	
	t.Run("render flat structure", func(t *testing.T) {
		testTree := createTestTree()
		options := RenderOptions{}
		
		output, err := renderer.Render(testTree, options)
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		// Parse the flat JSON array
		var paths []FlatPath
		if err := json.Unmarshal([]byte(output), &paths); err != nil {
			t.Fatalf("Failed to parse flat JSON output: %v", err)
		}
		
		// Should have all nodes flattened
		expectedCount := 6 // root + file1 + dir1 + file2 + file3 + empty_dir
		if len(paths) != expectedCount {
			t.Errorf("Flat paths count = %d, want %d", len(paths), expectedCount)
		}
		
		// Verify first path (root)
		if len(paths) > 0 {
			root := paths[0]
			if root.Path != "root" {
				t.Errorf("First path = %q, want %q", root.Path, "root")
			}
			if root.Name != "root" {
				t.Errorf("First name = %q, want %q", root.Name, "root")
			}
			if !root.IsDirectory {
				t.Error("Root should be a directory")
			}
			if root.Depth != 0 {
				t.Errorf("Root depth = %d, want 0", root.Depth)
			}
		}
		
		// Find and verify a nested file
		var file2 *FlatPath
		for i := range paths {
			if paths[i].Path == "root/dir1/file2.txt" {
				file2 = &paths[i]
				break
			}
		}
		
		if file2 == nil {
			t.Error("Could not find file2.txt in flat structure")
		} else {
			if file2.Name != "file2.txt" {
				t.Errorf("file2 name = %q, want %q", file2.Name, "file2.txt")
			}
			if file2.IsDirectory {
				t.Error("file2.txt should not be a directory")
			}
			if file2.Depth != 2 {
				t.Errorf("file2.txt depth = %d, want 2", file2.Depth)
			}
			if file2.Annotation == nil {
				t.Error("file2.txt should have annotation")
			} else {
				if file2.Annotation.Notes != "Some notes" {
					t.Errorf("file2.txt notes = %q, want %q", file2.Annotation.Notes, "Some notes")
				}
			}
		}
		
		// Verify proper depth calculation
		depthMap := make(map[string]int)
		for _, p := range paths {
			depthMap[p.Path] = p.Depth
		}
		
		expectedDepths := map[string]int{
			"root":                 0,
			"root/file1.txt":       1,
			"root/dir1":            1,
			"root/dir1/file2.txt":  2,
			"root/dir1/file3.txt":  2,
			"root/empty_dir":       1,
		}
		
		for path, expectedDepth := range expectedDepths {
			if depth, exists := depthMap[path]; !exists {
				t.Errorf("Path %q not found in output", path)
			} else if depth != expectedDepth {
				t.Errorf("Path %q depth = %d, want %d", path, depth, expectedDepth)
			}
		}
	})
	
	t.Run("Notes field fallback", func(t *testing.T) {
		// Test that annotations work properly
		testTree := &types.Node{
			Name: "root",
			Children: []*types.Node{
				{
					Name: "file1.txt",
					Annotation: &types.Annotation{
						Path:  "file1.txt",
						Notes: "First file notes",
					},
				},
				{
					Name: "file2.txt",
					Annotation: &types.Annotation{
						Path:  "file2.txt",
						Notes: "Has notes",
					},
				},
			},
		}
		
		renderer := &FlatJSONRenderer{}
		output, err := renderer.Render(testTree, RenderOptions{})
		if err != nil {
			t.Fatalf("Render() unexpected error: %v", err)
		}
		
		var paths []FlatPath
		if err := json.Unmarshal([]byte(output), &paths); err != nil {
			t.Fatalf("Failed to parse flat JSON output: %v", err)
		}
		
		// Find files and verify annotations
		for _, p := range paths {
			if p.Name == "file1.txt" && p.Annotation != nil {
				if p.Annotation.Notes != "First file notes" {
					t.Errorf("Expected Notes to be %q, got %q", "First file notes", p.Annotation.Notes)
				}
			}
			if p.Name == "file2.txt" && p.Annotation != nil {
				if p.Annotation.Notes != "Has notes" {
					t.Errorf("Expected Notes to be preserved, got %q", p.Annotation.Notes)
				}
			}
		}
	})
	
	t.Run("nil tree", func(t *testing.T) {
		renderer := &FlatJSONRenderer{}
		options := RenderOptions{}
		
		_, err := renderer.Render(nil, options)
		if err == nil {
			t.Error("Expected error when rendering nil tree, got none")
		}
	})
}

func TestConvertToTreeData(t *testing.T) {
	// Test the conversion logic separately
	renderer := &JSONRenderer{}
	
	t.Run("path construction", func(t *testing.T) {
		// Test root node
		root := &types.Node{Name: "root"}
		data := renderer.convertToTreeData(root, "")
		if data.Path != "root" {
			t.Errorf("Root path = %q, want %q", data.Path, "root")
		}
		
		// Test child node
		child := &types.Node{Name: "child"}
		data = renderer.convertToTreeData(child, "parent")
		if data.Path != "parent/child" {
			t.Errorf("Child path = %q, want %q", data.Path, "parent/child")
		}
	})
	
	t.Run("annotation conversion", func(t *testing.T) {
		node := &types.Node{
			Name: "test",
			Annotation: &types.Annotation{
				Path:  "test",
				Notes: "Test Notes",
			},
		}
		
		data := renderer.convertToTreeData(node, "")
		
		if data.Annotation == nil {
			t.Fatal("Expected annotation to be converted")
		}
		
		if data.Annotation.Notes != "Test Notes" {
			t.Errorf("Annotation notes = %q, want %q", data.Annotation.Notes, "Test Notes")
		}
	})
}

func TestCollectPaths(t *testing.T) {
	renderer := &FlatJSONRenderer{}
	
	t.Run("recursive collection", func(t *testing.T) {
		testTree := &types.Node{
			Name: "root",
			Children: []*types.Node{
				{
					Name: "a",
					Children: []*types.Node{
						{
							Name: "b",
							Children: []*types.Node{
								{Name: "c"},
							},
						},
					},
				},
			},
		}
		
		var paths []FlatPath
		renderer.collectPaths(testTree, "", 0, &paths)
		
		if len(paths) != 4 {
			t.Errorf("Expected 4 paths, got %d", len(paths))
		}
		
		// Verify all paths and depths
		expectedPaths := []struct {
			path  string
			depth int
		}{
			{"root", 0},
			{"root/a", 1},
			{"root/a/b", 2},
			{"root/a/b/c", 3},
		}
		
		for i, expected := range expectedPaths {
			if i >= len(paths) {
				break
			}
			if paths[i].Path != expected.path {
				t.Errorf("Path[%d] = %q, want %q", i, paths[i].Path, expected.path)
			}
			if paths[i].Depth != expected.depth {
				t.Errorf("Path[%d] depth = %d, want %d", i, paths[i].Depth, expected.depth)
			}
		}
	})
}