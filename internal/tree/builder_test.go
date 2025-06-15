package tree

import (
	"os"
	"path/filepath"
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
	expectedChildren := []string{"docs", "src", "LICENSE", "README.md"}
	if len(root.Children) != len(expectedChildren) {
		t.Errorf("Expected %d children, got %d", len(expectedChildren), len(root.Children))
	}
	
	// Check that children are sorted correctly (directories first, then alphabetically)
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