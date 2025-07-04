package tree

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/info"
)

func TestViewMode_All(t *testing.T) {
	// Create test annotations
	annotations := map[string]*info.Annotation{
		"file1.go": {
			Title:       "File 1",
			Description: "Annotated file 1",
		},
		"dir1": {
			Title:       "Directory 1", 
			Description: "Annotated directory",
		},
	}

	// Create a mock tree structure
	root := &Node{
		Name:         "root",
		Path:         "/test/root",
		RelativePath: ".",
		IsDir:        true,
		Children: []*Node{
			{
				Name:         "file1.go",
				Path:         "/test/root/file1.go",
				RelativePath: "file1.go",
				IsDir:        false,
				Annotation:   annotations["file1.go"],
			},
			{
				Name:         "file2.go",
				Path:         "/test/root/file2.go", 
				RelativePath: "file2.go",
				IsDir:        false,
			},
			{
				Name:         "dir1",
				Path:         "/test/root/dir1",
				RelativePath: "dir1",
				IsDir:        true,
				Annotation:   annotations["dir1"],
				Children: []*Node{
					{
						Name:         "nested.go",
						Path:         "/test/root/dir1/nested.go",
						RelativePath: "dir1/nested.go",
						IsDir:        false,
					},
				},
			},
		},
	}

	// Apply view mode "all"
	vb := &ViewBuilder{
		viewOptions: ViewOptions{Mode: ViewModeAll},
	}
	vb.applyViewMode(root)

	// Verify all nodes are still present
	if len(root.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(root.Children))
	}
	if root.Children[0].Name != "file1.go" {
		t.Errorf("Expected first child to be file1.go, got %s", root.Children[0].Name)
	}
	if root.Children[1].Name != "file2.go" {
		t.Errorf("Expected second child to be file2.go, got %s", root.Children[1].Name)
	}
	if root.Children[2].Name != "dir1" {
		t.Errorf("Expected third child to be dir1, got %s", root.Children[2].Name)
	}
	if len(root.Children[2].Children) != 1 {
		t.Errorf("Expected dir1 to have 1 child, got %d", len(root.Children[2].Children))
	}
}

func TestViewMode_Annotated(t *testing.T) {
	// Create test annotations
	annotations := map[string]*info.Annotation{
		"file1.go": {
			Title:       "File 1",
			Description: "Annotated file 1",
		},
		"dir1/nested_annotated.go": {
			Title:       "Nested annotated",
			Description: "Annotated nested file",
		},
	}

	// Create a mock tree structure
	root := &Node{
		Name:         "root",
		Path:         "/test/root",
		RelativePath: ".",
		IsDir:        true,
		Children: []*Node{
			{
				Name:         "file1.go",
				Path:         "/test/root/file1.go",
				RelativePath: "file1.go",
				IsDir:        false,
				Annotation:   annotations["file1.go"],
			},
			{
				Name:         "file2.go",
				Path:         "/test/root/file2.go",
				RelativePath: "file2.go",
				IsDir:        false,
			},
			{
				Name:         "dir1",
				Path:         "/test/root/dir1",
				RelativePath: "dir1",
				IsDir:        true,
				Children: []*Node{
					{
						Name:         "nested.go",
						Path:         "/test/root/dir1/nested.go",
						RelativePath: "dir1/nested.go",
						IsDir:        false,
					},
					{
						Name:         "nested_annotated.go",
						Path:         "/test/root/dir1/nested_annotated.go",
						RelativePath: "dir1/nested_annotated.go",
						IsDir:        false,
						Annotation:   annotations["dir1/nested_annotated.go"],
					},
				},
			},
			{
				Name:         "empty_dir",
				Path:         "/test/root/empty_dir",
				RelativePath: "empty_dir",
				IsDir:        true,
				Children:     []*Node{},
			},
		},
	}

	// Set up parent references
	for _, child := range root.Children {
		child.Parent = root
		for _, nested := range child.Children {
			nested.Parent = child
		}
	}

	// Apply view mode "annotated"
	vb := &ViewBuilder{
		viewOptions: ViewOptions{Mode: ViewModeAnnotated},
	}
	vb.applyViewMode(root)

	// Verify only annotated nodes and their parent directories are present
	if len(root.Children) != 3 {
		t.Errorf("Expected 3 children (file1.go, dir1, and message), got %d", len(root.Children))
	}
	if root.Children[0].Name != "file1.go" {
		t.Errorf("Expected first child to be file1.go, got %s", root.Children[0].Name)
	}
	if root.Children[1].Name != "dir1" {
		t.Errorf("Expected second child to be dir1, got %s", root.Children[1].Name)
	}
	
	// Check the message node
	lastNode := root.Children[2]
	if lastNode.Name != "" {
		t.Errorf("Expected message node to have empty name, got %s", lastNode.Name)
	}
	if lastNode.Annotation == nil {
		t.Error("Expected message node to have annotation")
	}
	if lastNode.Annotation != nil && lastNode.Annotation.Description != "treex --show all to see all paths" {
		t.Errorf("Expected message description, got %s", lastNode.Annotation.Description)
	}
	
	// Verify dir1 only contains annotated files
	dir1 := root.Children[1]
	if len(dir1.Children) != 1 {
		t.Errorf("Expected dir1 to have 1 child, got %d", len(dir1.Children))
	}
	if dir1.Children[0].Name != "nested_annotated.go" {
		t.Errorf("Expected dir1's child to be nested_annotated.go, got %s", dir1.Children[0].Name)
	}
}

func TestViewMode_Mix_TopLevel(t *testing.T) {
	// Create test annotations
	annotations := map[string]*info.Annotation{
		"annotated1.go": {
			Title: "Annotated 1",
		},
		"annotated2.go": {
			Title: "Annotated 2",
		},
	}

	// Create a mock tree structure with many unannotated files
	root := &Node{
		Name:         "root",
		Path:         "/test/root",
		RelativePath: ".",
		IsDir:        true,
		Children: []*Node{
			{Name: "annotated1.go", RelativePath: "annotated1.go", Annotation: annotations["annotated1.go"]},
			{Name: "annotated2.go", RelativePath: "annotated2.go", Annotation: annotations["annotated2.go"]},
			{Name: "unannotated1.go", RelativePath: "unannotated1.go"},
			{Name: "unannotated2.go", RelativePath: "unannotated2.go"},
			{Name: "unannotated3.go", RelativePath: "unannotated3.go"},
			{Name: "unannotated4.go", RelativePath: "unannotated4.go"},
			{Name: "unannotated5.go", RelativePath: "unannotated5.go"},
			{Name: "unannotated6.go", RelativePath: "unannotated6.go"},
			{Name: "unannotated7.go", RelativePath: "unannotated7.go"},
			{Name: "unannotated8.go", RelativePath: "unannotated8.go"},
		},
	}

	// Set up parent references
	for _, child := range root.Children {
		child.Parent = root
	}

	// Apply view mode "mix"
	vb := &ViewBuilder{
		viewOptions: ViewOptions{Mode: ViewModeMix},
	}
	vb.applyViewMode(root)

	// Verify we have: 2 annotated + 6 context + 1 "more items" = 9 total
	if len(root.Children) != 9 {
		t.Errorf("Expected 9 children (2 annotated + 6 context + 1 more items), got %d", len(root.Children))
	}
	
	// Count annotated files
	annotatedCount := 0
	moreItemsFound := false
	for _, child := range root.Children {
		if child.Annotation != nil && child.Name != "" {
			annotatedCount++
		}
		if strings.HasPrefix(child.Name, "... ") && strings.HasSuffix(child.Name, " more items") {
			moreItemsFound = true
		}
	}
	
	if annotatedCount != 2 {
		t.Errorf("Expected 2 annotated files, got %d", annotatedCount)
	}
	if !moreItemsFound {
		t.Error("Should have 'more items' indicator")
	}
}

func TestViewMode_Mix_SubDirectory(t *testing.T) {
	// Create test annotations
	annotations := map[string]*info.Annotation{
		"dir1/annotated1.go": {Title: "Annotated 1"},
		"dir1/annotated2.go": {Title: "Annotated 2"},
	}

	// Create a directory with 2 annotated files
	dir1 := &Node{
		Name:         "dir1",
		Path:         "/test/root/dir1",
		RelativePath: "dir1",
		IsDir:        true,
		Children: []*Node{
			{Name: "annotated1.go", RelativePath: "dir1/annotated1.go", Annotation: annotations["dir1/annotated1.go"]},
			{Name: "annotated2.go", RelativePath: "dir1/annotated2.go", Annotation: annotations["dir1/annotated2.go"]},
			{Name: "unannotated1.go", RelativePath: "dir1/unannotated1.go"},
			{Name: "unannotated2.go", RelativePath: "dir1/unannotated2.go"},
			{Name: "unannotated3.go", RelativePath: "dir1/unannotated3.go"},
			{Name: "unannotated4.go", RelativePath: "dir1/unannotated4.go"},
		},
	}

	// Set up parent references
	for _, child := range dir1.Children {
		child.Parent = dir1
	}

	// Apply view mode "mix" at depth 1
	vb := &ViewBuilder{
		viewOptions: ViewOptions{Mode: ViewModeMix},
	}
	vb.applyMixMode(dir1, 1)

	// Verify we have: 2 annotated + 2 context + 1 "more items" = 5 total
	if len(dir1.Children) != 5 {
		t.Errorf("Expected 5 children (2 annotated + 2 context + 1 more items), got %d", len(dir1.Children))
	}
	
	// Count types
	annotatedCount := 0
	unannotatedCount := 0
	moreItemsFound := false
	
	for _, child := range dir1.Children {
		if child.Annotation != nil {
			annotatedCount++
		} else if strings.HasPrefix(child.Name, "... ") && strings.HasSuffix(child.Name, " more items") {
			moreItemsFound = true
		} else {
			unannotatedCount++
		}
	}
	
	if annotatedCount != 2 {
		t.Errorf("Expected 2 annotated files, got %d", annotatedCount)
	}
	if unannotatedCount != 2 {
		t.Errorf("Expected 2 context paths, got %d", unannotatedCount)
	}
	if !moreItemsFound {
		t.Error("Should have 'more items' indicator")
	}
}

func TestViewMode_Mix_FewUnannotatedFiles(t *testing.T) {
	// Create test annotations
	annotations := map[string]*info.Annotation{
		"annotated.go": {Title: "Annotated"},
	}

	// Create root with few unannotated files
	root := &Node{
		Name:         "root",
		Path:         "/test/root",
		RelativePath: ".",
		IsDir:        true,
		Children: []*Node{
			{Name: "annotated.go", RelativePath: "annotated.go", Annotation: annotations["annotated.go"]},
			{Name: "file1.go", RelativePath: "file1.go"},
			{Name: "file2.go", RelativePath: "file2.go"},
			{Name: "file3.go", RelativePath: "file3.go"},
		},
	}

	// Set up parent references
	for _, child := range root.Children {
		child.Parent = root
	}

	// Apply view mode "mix"
	vb := &ViewBuilder{
		viewOptions: ViewOptions{Mode: ViewModeMix},
	}
	vb.applyViewMode(root)

	// Should show all files (1 annotated + 3 unannotated = 4 total)
	if len(root.Children) != 4 {
		t.Errorf("Expected 4 children (1 annotated + 3 unannotated), got %d", len(root.Children))
	}
	
	// Verify no "more items" indicator
	moreItemsFound := false
	for _, child := range root.Children {
		if strings.HasPrefix(child.Name, "... ") {
			moreItemsFound = true
		}
	}
	if moreItemsFound {
		t.Error("Should not have 'more items' when showing all files")
	}
}