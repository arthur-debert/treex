package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/info"
)

func TestViewBuilder_Integration_AllMode(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"annotated.go",
		"unannotated1.go", 
		"unannotated2.go",
		"dir1/nested_annotated.go",
		"dir1/nested_unannotated.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .info file with annotations
	infoContent := `annotated.go Main annotated file
This is the main file with annotation

dir1/nested_annotated.go Nested annotated file`

	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Parse annotations (use ParseDirectoryTree to handle nested annotations)
	annotations, err := info.ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	// Build tree with ViewModeAll
	viewOptions := ViewOptions{Mode: ViewModeAll}
	builder := NewViewBuilder(tempDir, annotations, viewOptions)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}

	// Verify all files are present
	fileCount := countNodes(root, false)
	if fileCount != 5 { // 3 files in root + 2 in dir1
		t.Errorf("Expected 5 files, got %d", fileCount)
	}

	// Verify directories
	dirCount := countNodes(root, true)
	if dirCount != 2 { // root + dir1
		t.Errorf("Expected 2 directories, got %d", dirCount)
	}
}

func TestViewBuilder_Integration_AnnotatedMode(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"annotated1.go",
		"annotated2.go",
		"unannotated1.go", 
		"unannotated2.go",
		"unannotated3.go",
		"dir1/nested_annotated.go",
		"dir1/nested_unannotated1.go",
		"dir1/nested_unannotated2.go",
		"dir2/all_unannotated.go",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .info file with annotations
	infoContent := `annotated1.go First annotated file

annotated2.go Second annotated file

dir1/nested_annotated.go Nested annotated file`

	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Parse annotations (use ParseDirectoryTree to handle nested annotations)
	annotations, err := info.ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	// Build tree with ViewModeAnnotated
	viewOptions := ViewOptions{Mode: ViewModeAnnotated}
	builder := NewViewBuilder(tempDir, annotations, viewOptions)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}

	// Count annotated files (excluding the message node)
	annotatedCount := 0
	hasMessageNode := false
	err = WalkTree(root, func(node *Node, depth int) error {
		if node.Annotation != nil && node.Name != "" {
			annotatedCount++
			// Found annotated node
		}
		if node.Name == "" && node.Annotation != nil && 
			node.Annotation.Description == "treex --show all to see all paths" {
			hasMessageNode = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	if annotatedCount != 3 {
		t.Errorf("Expected 3 annotated files, got %d", annotatedCount)
	}

	if !hasMessageNode {
		t.Error("Expected message node with 'treex --show all to see all paths'")
	}

	// Verify dir2 is not present (has no annotations)
	dir2Found := false
	for _, child := range root.Children {
		if child.Name == "dir2" {
			dir2Found = true
			break
		}
	}
	if dir2Found {
		t.Error("dir2 should not be present in annotated mode")
	}
}

func TestViewBuilder_Integration_MixMode(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create many files to test context limits
	var testFiles []string
	
	// Add annotated files
	testFiles = append(testFiles, "annotated1.go", "annotated2.go")
	
	// Add many unannotated files (more than 6)
	for i := 1; i <= 10; i++ {
		testFiles = append(testFiles, fmt.Sprintf("unannotated%d.go", i))
	}

	// Add directory with multiple annotated files
	testFiles = append(testFiles, 
		"multi_annotated/ann1.go",
		"multi_annotated/ann2.go", 
		"multi_annotated/ann3.go",
		"multi_annotated/unann1.go",
		"multi_annotated/unann2.go",
		"multi_annotated/unann3.go",
		"multi_annotated/unann4.go",
	)

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .info file with annotations
	infoContent := `annotated1.go First annotated

annotated2.go Second annotated

multi_annotated/ann1.go Multi ann 1

multi_annotated/ann2.go Multi ann 2

multi_annotated/ann3.go Multi ann 3`

	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}

	// Parse annotations (use ParseDirectoryTree to handle nested annotations)
	annotations, err := info.ParseDirectoryTree(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse annotations: %v", err)
	}

	// Build tree with ViewModeMix
	viewOptions := ViewOptions{Mode: ViewModeMix}
	builder := NewViewBuilder(tempDir, annotations, viewOptions)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}

	// Check root level: should have 2 annotated + 6 context + 1 dir + 1 "more items"
	rootFileCount := 0
	rootDirCount := 0
	rootMoreItems := false
	
	for _, child := range root.Children {
		if child.IsDir {
			rootDirCount++
		} else if strings.HasPrefix(child.Name, "... ") && strings.HasSuffix(child.Name, " more items") {
			rootMoreItems = true
		} else {
			rootFileCount++
		}
	}

	if rootFileCount != 8 { // 2 annotated + 6 context
		t.Errorf("Expected 8 files at root, got %d", rootFileCount)
	}
	
	if !rootMoreItems {
		t.Error("Expected 'more items' indicator at root level")
	}

	// Check multi_annotated directory
	var multiDir *Node
	for _, child := range root.Children {
		if child.Name == "multi_annotated" {
			multiDir = child
			break
		}
	}

	if multiDir == nil {
		t.Fatal("Could not find multi_annotated directory")
	}

	// Should have 3 annotated + 2 context + 1 "more items"
	multiAnnotated := 0
	multiContext := 0  
	multiMoreItems := false

	for _, child := range multiDir.Children {
		if child.Annotation != nil {
			multiAnnotated++
		} else if strings.HasPrefix(child.Name, "... ") && strings.HasSuffix(child.Name, " more items") {
			multiMoreItems = true
		} else {
			multiContext++
		}
	}

	if multiAnnotated != 3 {
		t.Errorf("Expected 3 annotated files in multi_annotated, got %d", multiAnnotated)
	}
	
	if multiContext != 2 {
		t.Errorf("Expected 2 context files in multi_annotated, got %d", multiContext)
	}
	
	if !multiMoreItems {
		t.Error("Expected 'more items' indicator in multi_annotated directory")
	}
}

// Helper function to count nodes
func countNodes(root *Node, countDirs bool) int {
	count := 0
	err := WalkTree(root, func(node *Node, depth int) error {
		if countDirs && node.IsDir {
			count++
		} else if !countDirs && !node.IsDir && !strings.HasPrefix(node.Name, "... ") {
			count++
		}
		return nil
	})
	if err != nil {
		return -1
	}
	return count
}