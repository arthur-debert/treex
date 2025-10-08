package treeconstruction_test

import (
	"testing"

	"github.com/jwaldrip/treex/treex/pathcollection"
	"github.com/jwaldrip/treex/treex/treeconstruction"
	"github.com/jwaldrip/treex/treex/types"
)

// Helper function to find a node by path in a tree.
// Returns nil if not found.
func findNodeByPath(root *types.Node, path string) *types.Node {
	if root == nil {
		return nil
	}
	if root.Path == path {
		return root
	}

	for _, child := range root.Children {
		if found := findNodeByPath(child, path); found != nil {
			return found
		}
	}

	return nil
}

func TestBuildTree_EmptyInput(t *testing.T) {
	constructor := treeconstruction.NewConstructor()
	paths := []pathcollection.PathInfo{}

	root := constructor.BuildTree(paths)

	if root != nil {
		t.Error("Expected nil root for empty input, but got a node")
	}
}

func TestBuildTree_RootOnly(t *testing.T) {
	constructor := treeconstruction.NewConstructor()
	paths := []pathcollection.PathInfo{
		{Path: ".", IsDir: true},
	}

	root := constructor.BuildTree(paths)

	if root == nil {
		t.Fatal("Root should not be nil")
	}
	if root.Path != "." {
		t.Errorf("Expected root path '.', got %q", root.Path)
	}
	if len(root.Children) != 0 {
		t.Errorf("Expected root to have 0 children, got %d", len(root.Children))
	}
}

func TestBuildTree_BasicStructure(t *testing.T) {
	constructor := treeconstruction.NewConstructor()
	paths := []pathcollection.PathInfo{
		{Path: ".", IsDir: true},
		{Path: "file1.txt", IsDir: false, Size: 10},
		{Path: "src", IsDir: true},
		{Path: "src/main.go", IsDir: false, Size: 20},
	}

	root := constructor.BuildTree(paths)

	if root == nil {
		t.Fatal("BuildTree returned a nil root")
	}
	if root.Path != "." {
		t.Errorf("Expected root path '.', got %q", root.Path)
	}

	// Check file1.txt
	file1 := findNodeByPath(root, "file1.txt")
	if file1 == nil {
		t.Fatal("Could not find node for 'file1.txt'")
	}
	if file1.Parent != root {
		t.Error("'file1.txt' should be a child of the root")
	}
	if file1.Size != 10 {
		t.Errorf("Expected size 10 for 'file1.txt', got %d", file1.Size)
	}

	// Check src directory
	srcDir := findNodeByPath(root, "src")
	if srcDir == nil {
		t.Fatal("Could not find node for 'src'")
	}
	if !srcDir.IsDir {
		t.Error("'src' should be a directory")
	}
	if srcDir.Parent != root {
		t.Error("'src' should be a child of the root")
	}

	// Check src/main.go
	mainGo := findNodeByPath(root, "src/main.go")
	if mainGo == nil {
		t.Fatal("Could not find node for 'src/main.go'")
	}
	if mainGo.Parent != srcDir {
		t.Error("'src/main.go' should be a child of 'src'")
	}
	if mainGo.Size != 20 {
		t.Errorf("Expected size 20 for 'src/main.go', got %d", mainGo.Size)
	}

	// Check children counts
	if len(root.Children) != 2 {
		t.Errorf("Expected root to have 2 children, got %d", len(root.Children))
	}
	if len(srcDir.Children) != 1 {
		t.Errorf("Expected 'src' to have 1 child, got %d", len(srcDir.Children))
	}
}

func TestBuildTree_UnsortedInput(t *testing.T) {
	constructor := treeconstruction.NewConstructor()
	// Deliberately unsorted input
	paths := []pathcollection.PathInfo{
		{Path: "src/main.go", IsDir: false},
		{Path: "file1.txt", IsDir: false},
		{Path: ".", IsDir: true},
		{Path: "src", IsDir: true},
	}

	root := constructor.BuildTree(paths)

	if root == nil {
		t.Fatal("BuildTree returned a nil root")
	}

	// If the tree is built correctly, the unsorted input was handled.
	srcDir := findNodeByPath(root, "src")
	if srcDir == nil {
		t.Fatal("Could not find 'src' directory")
	}
	mainGo := findNodeByPath(srcDir, "src/main.go")
	if mainGo == nil {
		t.Fatal("Could not find 'src/main.go' as a child of 'src'")
	}
	if mainGo.Parent != srcDir {
		t.Error("Parent-child relationship incorrect for unsorted input")
	}
}

func TestBuildTree_DeeplyNested(t *testing.T) {
	constructor := treeconstruction.NewConstructor()
	paths := []pathcollection.PathInfo{
		{Path: ".", IsDir: true},
		{Path: "a", IsDir: true},
		{Path: "a/b", IsDir: true},
		{Path: "a/b/c", IsDir: true},
		{Path: "a/b/c/d.txt", IsDir: false},
	}

	root := constructor.BuildTree(paths)
	if root == nil {
		t.Fatal("BuildTree returned a nil root")
	}

	d_txt := findNodeByPath(root, "a/b/c/d.txt")
	if d_txt == nil {
		t.Fatal("Could not find 'a/b/c/d.txt'")
	}

	// Check parent chain
	if d_txt.Parent.Path != "a/b/c" {
		t.Errorf("Expected parent of d.txt to be 'a/b/c', got %q", d_txt.Parent.Path)
	}
	if d_txt.Parent.Parent.Path != "a/b" {
		t.Errorf("Expected grandparent of d.txt to be 'a/b', got %q", d_txt.Parent.Parent.Path)
	}
	if d_txt.Parent.Parent.Parent.Path != "a" {
		t.Errorf("Expected great-grandparent of d.txt to be 'a', got %q", d_txt.Parent.Parent.Parent.Path)
	}
	if d_txt.Parent.Parent.Parent.Parent.Path != "." {
		t.Errorf("Expected great-great-grandparent of d.txt to be '.', got %q", d_txt.Parent.Parent.Parent.Parent.Path)
	}
}

func TestBuildTree_FlatStructure(t *testing.T) {
	constructor := treeconstruction.NewConstructor()
	paths := []pathcollection.PathInfo{
		{Path: ".", IsDir: true},
		{Path: "file1.txt", IsDir: false},
		{Path: "file2.txt", IsDir: false},
		{Path: "file3.txt", IsDir: false},
	}

	root := constructor.BuildTree(paths)
	if root == nil {
		t.Fatal("BuildTree returned a nil root")
	}

	if len(root.Children) != 3 {
		t.Fatalf("Expected root to have 3 children, got %d", len(root.Children))
	}

	for _, child := range root.Children {
		if child.Parent != root {
			t.Errorf("Child %q should have root as parent", child.Path)
		}
	}
}
