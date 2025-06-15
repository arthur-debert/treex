package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adebert/treex/internal/info"
)

func TestBuildTree(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	
	// Create test files and directories
	testFiles := []string{
		"README.md",
		"LICENSE",
		"src/main.go",
		"src/utils.go",
		"docs/guide.md",
	}
	
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		// Create the file
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}
	
	// Create a .info file with annotations
	infoContent := `README.md
Project readme file

src/main.go
Main application entry point`
	
	infoPath := filepath.Join(tempDir, ".info")
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		t.Fatalf("Failed to create .info file: %v", err)
	}
	
	// Build the tree
	root, err := BuildTree(tempDir)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	
	// Verify root node
	if !root.IsDir {
		t.Error("Root should be a directory")
	}
	
	if root.Parent != nil {
		t.Error("Root should have no parent")
	}
	
	// Verify we have the expected children
	expectedChildren := []string{"README.md", "docs", "src", "LICENSE"} // README.md first (annotated), then others
	if len(root.Children) != len(expectedChildren) {
		t.Errorf("Expected %d children, got %d", len(expectedChildren), len(root.Children))
	}
	
	// Check that children are sorted correctly (annotated files first, then others)
	childNames := make([]string, len(root.Children))
	for i, child := range root.Children {
		childNames[i] = child.Name
	}
	
	for i, expected := range expectedChildren {
		if i >= len(childNames) || childNames[i] != expected {
			t.Errorf("Expected child %d to be %s, got %s", i, expected, childNames[i])
		}
	}
	
	// Verify annotations are attached correctly
	var readmeNode *Node
	for _, child := range root.Children {
		if child.Name == "README.md" {
			readmeNode = child
			break
		}
	}
	
	if readmeNode == nil {
		t.Fatal("README.md node not found")
	}
	
	if readmeNode.Annotation == nil {
		t.Error("README.md should have an annotation")
	} else {
		expectedDesc := "Project readme file"
		if readmeNode.Annotation.Description != expectedDesc {
			t.Errorf("README.md annotation mismatch. Expected: %q, Got: %q", 
				expectedDesc, readmeNode.Annotation.Description)
		}
	}
	
	// Verify nested structure
	var srcNode *Node
	for _, child := range root.Children {
		if child.Name == "src" && child.IsDir {
			srcNode = child
			break
		}
	}
	
	if srcNode == nil {
		t.Fatal("src directory not found")
	}
	
	if len(srcNode.Children) != 2 {
		t.Errorf("src directory should have 2 children, got %d", len(srcNode.Children))
	}
	
	// Check for main.go annotation
	var mainGoNode *Node
	for _, child := range srcNode.Children {
		if child.Name == "main.go" {
			mainGoNode = child
			break
		}
	}
	
	if mainGoNode == nil {
		t.Fatal("main.go node not found")
	}
	
	if mainGoNode.Annotation == nil {
		t.Error("main.go should have an annotation")
	} else {
		expectedDesc := "Main application entry point"
		if mainGoNode.Annotation.Description != expectedDesc {
			t.Errorf("main.go annotation mismatch. Expected: %q, Got: %q", 
				expectedDesc, mainGoNode.Annotation.Description)
		}
	}
}

func TestWalkTree(t *testing.T) {
	// Create a simple test tree
	tempDir := t.TempDir()
	
	// Create test structure
	testFiles := []string{
		"file1.txt",
		"dir1/file2.txt",
		"dir1/file3.txt",
	}
	
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}
	
	// Build tree
	root, err := BuildTree(tempDir)
	if err != nil {
		t.Fatalf("BuildTree failed: %v", err)
	}
	
	// Walk the tree and collect visited nodes
	var visitedNodes []string
	var visitedDepths []int
	
	err = WalkTree(root, func(node *Node, depth int) error {
		visitedNodes = append(visitedNodes, node.Name)
		visitedDepths = append(visitedDepths, depth)
		return nil
	})
	
	if err != nil {
		t.Fatalf("WalkTree failed: %v", err)
	}
	
	// Verify we visited all nodes in the correct order
	expectedNodes := []string{
		filepath.Base(tempDir), // root
		"dir1",                 // directory first
		"file2.txt",           // files in dir1
		"file3.txt",
		"file1.txt",           // file in root
	}
	
	if len(visitedNodes) != len(expectedNodes) {
		t.Errorf("Expected to visit %d nodes, visited %d", len(expectedNodes), len(visitedNodes))
	}
	
	for i, expected := range expectedNodes {
		if i >= len(visitedNodes) || visitedNodes[i] != expected {
			t.Errorf("Expected to visit node %d as %s, got %s", i, expected, visitedNodes[i])
		}
	}
	
	// Verify depths are correct
	expectedDepths := []int{0, 1, 2, 2, 1}
	for i, expected := range expectedDepths {
		if i >= len(visitedDepths) || visitedDepths[i] != expected {
			t.Errorf("Expected depth %d for node %d, got %d", expected, i, visitedDepths[i])
		}
	}
}

func TestBuilderWithAnnotations(t *testing.T) {
	// Create test annotations
	annotations := map[string]*info.Annotation{
		"test.txt": {
			Path:        "test.txt",
			Description: "A test file",
		},
		"subdir/nested.txt": {
			Path:        "subdir/nested.txt",
			Title:       "Nested File",
			Description: "Nested File\nA file in a subdirectory",
		},
	}
	
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create the files
	if err := os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test.txt: %v", err)
	}
	
	subdir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	
	if err := os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create nested.txt: %v", err)
	}
	
	// Build tree with annotations
	builder := NewBuilder(tempDir, annotations)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	
	// Find and verify annotated nodes
	var testTxtNode, nestedTxtNode *Node
	
	err = WalkTree(root, func(node *Node, depth int) error {
		if node.Name == "test.txt" {
			testTxtNode = node
		}
		if node.Name == "nested.txt" {
			nestedTxtNode = node
		}
		return nil
	})
	
	if err != nil {
		t.Fatalf("WalkTree failed: %v", err)
	}
	
	// Verify test.txt annotation
	if testTxtNode == nil {
		t.Fatal("test.txt node not found")
	}
	
	if testTxtNode.Annotation == nil {
		t.Error("test.txt should have annotation")
	} else {
		if testTxtNode.Annotation.Description != "A test file" {
			t.Errorf("test.txt annotation mismatch: %q", testTxtNode.Annotation.Description)
		}
	}
	
	// Verify nested.txt annotation
	if nestedTxtNode == nil {
		t.Fatal("nested.txt node not found")
	}
	
	if nestedTxtNode.Annotation == nil {
		t.Error("nested.txt should have annotation")
	} else {
		if nestedTxtNode.Annotation.Title != "Nested File" {
			t.Errorf("nested.txt title mismatch: %q", nestedTxtNode.Annotation.Title)
		}
	}
}

func TestBuilder_MaxFilesProtection(t *testing.T) {
	// Create a test directory with many files
	tempDir := t.TempDir()
	
	// Create annotations for only a few files
	annotations := map[string]*info.Annotation{
		"important1.txt": {
			Path:        "important1.txt",
			Description: "Important file 1",
		},
		"important2.txt": {
			Path:        "important2.txt", 
			Description: "Important file 2",
		},
	}
	
	// Create test files (more than MAX_FILES_PER_DIR)
	testFiles := make([]string, 15) // More than MAX_FILES_PER_DIR (10)
	for i := 0; i < 15; i++ {
		if i < 2 {
			testFiles[i] = fmt.Sprintf("important%d.txt", i+1)
		} else {
			testFiles[i] = fmt.Sprintf("file%02d.txt", i+1)
		}
	}
	
	// Create the test files
	for _, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}
	
	// Build tree
	builder := NewBuilder(tempDir, annotations)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}
	
	// Count children and check for "more files..." indicator
	childCount := len(root.Children)
	var hasMoreFilesIndicator bool
	var annotatedCount int
	var regularFileCount int
	
	for _, child := range root.Children {
		if strings.Contains(child.Name, "more files not shown") {
			hasMoreFilesIndicator = true
		} else if child.Annotation != nil {
			annotatedCount++
		} else {
			regularFileCount++
		}
	}
	
	// Verify behavior
	if annotatedCount != 2 {
		t.Errorf("Expected 2 annotated files, got %d", annotatedCount)
	}
	
	if regularFileCount > MAX_FILES_PER_DIR {
		t.Errorf("Expected at most %d regular files, got %d", MAX_FILES_PER_DIR, regularFileCount)
	}
	
	if !hasMoreFilesIndicator {
		t.Error("Expected 'more files not shown' indicator but didn't find it")
	}
	
	// Total should be: annotated files + limited regular files + indicator
	expectedMaxChildren := 2 + MAX_FILES_PER_DIR + 1 // +1 for the indicator
	if childCount > expectedMaxChildren {
		t.Errorf("Too many children: expected at most %d, got %d", expectedMaxChildren, childCount)
	}
}

func TestBuilder_MaxFilesProtection_UnderLimit(t *testing.T) {
	// Test case where total files is under the limit
	tempDir := t.TempDir()
	
	annotations := map[string]*info.Annotation{
		"annotated.txt": {
			Path:        "annotated.txt",
			Description: "Annotated file",
		},
	}
	
	// Create fewer files than the limit
	testFiles := []string{"annotated.txt", "file1.txt", "file2.txt", "file3.txt"}
	
	for _, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}
	
	// Build tree
	builder := NewBuilder(tempDir, annotations)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}
	
	// Check that all files are shown and no indicator is present
	childCount := len(root.Children)
	var hasMoreFilesIndicator bool
	
	for _, child := range root.Children {
		if strings.Contains(child.Name, "more files not shown") {
			hasMoreFilesIndicator = true
		}
	}
	
	if hasMoreFilesIndicator {
		t.Error("Didn't expect 'more files not shown' indicator when under limit")
	}
	
	if childCount != len(testFiles) {
		t.Errorf("Expected %d children, got %d", len(testFiles), childCount)
	}
}

func TestBuilder_MaxFilesProtection_OnlyAnnotated(t *testing.T) {
	// Test case where all files are annotated (should show all regardless of limit)
	tempDir := t.TempDir()
	
	// Create more annotated files than the limit
	annotations := make(map[string]*info.Annotation)
	testFiles := make([]string, 15) // More than MAX_FILES_PER_DIR
	
	for i := 0; i < 15; i++ {
		fileName := fmt.Sprintf("annotated%02d.txt", i+1)
		testFiles[i] = fileName
		annotations[fileName] = &info.Annotation{
			Path:        fileName,
			Description: fmt.Sprintf("Annotated file %d", i+1),
		}
	}
	
	// Create the test files
	for _, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}
	
	// Build tree
	builder := NewBuilder(tempDir, annotations)
	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}
	
	// All annotated files should be shown, no indicator should be present
	childCount := len(root.Children)
	var hasMoreFilesIndicator bool
	var annotatedCount int
	
	for _, child := range root.Children {
		if strings.Contains(child.Name, "more files not shown") {
			hasMoreFilesIndicator = true
		} else if child.Annotation != nil {
			annotatedCount++
		}
	}
	
	if hasMoreFilesIndicator {
		t.Error("Didn't expect 'more files not shown' indicator when all files are annotated")
	}
	
	if annotatedCount != 15 {
		t.Errorf("Expected all 15 annotated files to be shown, got %d", annotatedCount)
	}
	
	if childCount != 15 {
		t.Errorf("Expected 15 children, got %d", childCount)
	}
} 