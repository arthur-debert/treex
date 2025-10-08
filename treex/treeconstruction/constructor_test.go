// see docs/dev/architecture.txt - Phase 3: Tree Construction
package treeconstruction_test

import (
	"testing"

	"treex/treex/pathcollection"
	"treex/treex/treeconstruction"
	"treex/treex/types"
)

func TestBasicTreeConstruction(t *testing.T) {
	// Create sample path information representing a simple project structure
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/project", IsDir: true, Depth: 0},
		{Path: "README.md", AbsolutePath: "/project/README.md", IsDir: false, Depth: 1},
		{Path: "src", AbsolutePath: "/project/src", IsDir: true, Depth: 1},
		{Path: "src/main.go", AbsolutePath: "/project/src/main.go", IsDir: false, Depth: 2},
		{Path: "src/utils.go", AbsolutePath: "/project/src/utils.go", IsDir: false, Depth: 2},
	}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Tree construction failed: %v", err)
	}

	// Verify root node
	if root == nil {
		t.Fatal("Root node is nil")
	}
	if root.Name != "." {
		t.Errorf("Expected root name '.', got %q", root.Name)
	}
	if !root.IsDir {
		t.Error("Root should be a directory")
	}
	if root.Parent != nil {
		t.Error("Root should have no parent")
	}

	// Verify root has correct children
	if len(root.Children) != 2 {
		t.Errorf("Expected root to have 2 children, got %d", len(root.Children))
	}

	// Find README.md and src children
	var readmeNode, srcNode *types.Node
	for _, child := range root.Children {
		if child.Name == "README.md" {
			readmeNode = child
		} else if child.Name == "src" {
			srcNode = child
		}
	}

	// Verify README.md node
	if readmeNode == nil {
		t.Error("README.md node not found in root children")
	} else {
		if readmeNode.IsDir {
			t.Error("README.md should not be a directory")
		}
		if readmeNode.Parent != root {
			t.Error("README.md parent should be root")
		}
		if len(readmeNode.Children) != 0 {
			t.Error("README.md should have no children")
		}
	}

	// Verify src directory node
	if srcNode == nil {
		t.Error("src node not found in root children")
	} else {
		if !srcNode.IsDir {
			t.Error("src should be a directory")
		}
		if srcNode.Parent != root {
			t.Error("src parent should be root")
		}
		if len(srcNode.Children) != 2 {
			t.Errorf("Expected src to have 2 children, got %d", len(srcNode.Children))
		}

		// Verify src's children
		var mainGoNode, utilsGoNode *types.Node
		for _, child := range srcNode.Children {
			if child.Name == "main.go" {
				mainGoNode = child
			} else if child.Name == "utils.go" {
				utilsGoNode = child
			}
		}

		if mainGoNode == nil {
			t.Error("main.go node not found in src children")
		} else {
			if mainGoNode.IsDir {
				t.Error("main.go should not be a directory")
			}
			if mainGoNode.Parent != srcNode {
				t.Error("main.go parent should be src")
			}
		}

		if utilsGoNode == nil {
			t.Error("utils.go node not found in src children")
		} else {
			if utilsGoNode.IsDir {
				t.Error("utils.go should not be a directory")
			}
			if utilsGoNode.Parent != srcNode {
				t.Error("utils.go parent should be src")
			}
		}
	}
}

func TestDeepNestingTreeConstruction(t *testing.T) {
	// Test deep directory nesting to ensure parent-child relationships work correctly
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/deep", IsDir: true, Depth: 0},
		{Path: "level1", AbsolutePath: "/deep/level1", IsDir: true, Depth: 1},
		{Path: "level1/level2", AbsolutePath: "/deep/level1/level2", IsDir: true, Depth: 2},
		{Path: "level1/level2/level3", AbsolutePath: "/deep/level1/level2/level3", IsDir: true, Depth: 3},
		{Path: "level1/level2/level3/deep.txt", AbsolutePath: "/deep/level1/level2/level3/deep.txt", IsDir: false, Depth: 4},
	}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Deep nesting tree construction failed: %v", err)
	}

	// Navigate down the tree to verify structure
	if len(root.Children) != 1 {
		t.Fatalf("Expected root to have 1 child, got %d", len(root.Children))
	}

	level1 := root.Children[0]
	if level1.Name != "level1" {
		t.Errorf("Expected first child to be 'level1', got %q", level1.Name)
	}
	if len(level1.Children) != 1 {
		t.Fatalf("Expected level1 to have 1 child, got %d", len(level1.Children))
	}

	level2 := level1.Children[0]
	if level2.Name != "level2" {
		t.Errorf("Expected level1 child to be 'level2', got %q", level2.Name)
	}
	if level2.Parent != level1 {
		t.Error("level2 parent should be level1")
	}

	level3 := level2.Children[0]
	if level3.Name != "level3" {
		t.Errorf("Expected level2 child to be 'level3', got %q", level3.Name)
	}
	if level3.Parent != level2 {
		t.Error("level3 parent should be level2")
	}

	deepFile := level3.Children[0]
	if deepFile.Name != "deep.txt" {
		t.Errorf("Expected level3 child to be 'deep.txt', got %q", deepFile.Name)
	}
	if deepFile.IsDir {
		t.Error("deep.txt should not be a directory")
	}
	if deepFile.Parent != level3 {
		t.Error("deep.txt parent should be level3")
	}
}

func TestUnsortedPathInput(t *testing.T) {
	// Test that tree construction works even with unsorted input paths
	// The constructor should sort them internally
	pathInfos := []pathcollection.PathInfo{
		{Path: "src/main.go", AbsolutePath: "/project/src/main.go", IsDir: false, Depth: 2},
		{Path: ".", AbsolutePath: "/project", IsDir: true, Depth: 0},
		{Path: "src/lib/helper.go", AbsolutePath: "/project/src/lib/helper.go", IsDir: false, Depth: 3},
		{Path: "src", AbsolutePath: "/project/src", IsDir: true, Depth: 1},
		{Path: "src/lib", AbsolutePath: "/project/src/lib", IsDir: true, Depth: 2},
		{Path: "README.md", AbsolutePath: "/project/README.md", IsDir: false, Depth: 1},
	}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Unsorted path tree construction failed: %v", err)
	}

	// Verify the tree structure is correct despite unsorted input
	if root.Name != "." {
		t.Errorf("Expected root name '.', got %q", root.Name)
	}

	// Find src directory
	var srcNode *types.Node
	for _, child := range root.Children {
		if child.Name == "src" {
			srcNode = child
		}
	}

	if srcNode == nil {
		t.Fatal("src node not found")
	}

	// Verify src has both main.go and lib
	foundMainGo := false
	foundLib := false
	for _, child := range srcNode.Children {
		if child.Name == "main.go" {
			foundMainGo = true
		} else if child.Name == "lib" {
			foundLib = true
			// Verify lib has helper.go
			if len(child.Children) != 1 || child.Children[0].Name != "helper.go" {
				t.Error("lib should have helper.go as its only child")
			}
		}
	}

	if !foundMainGo {
		t.Error("main.go not found in src children")
	}
	if !foundLib {
		t.Error("lib not found in src children")
	}
}

func TestGetNodeByPath(t *testing.T) {
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/test", IsDir: true, Depth: 0},
		{Path: "file.txt", AbsolutePath: "/test/file.txt", IsDir: false, Depth: 1},
		{Path: "dir", AbsolutePath: "/test/dir", IsDir: true, Depth: 1},
		{Path: "dir/nested.txt", AbsolutePath: "/test/dir/nested.txt", IsDir: false, Depth: 2},
	}

	constructor := treeconstruction.NewTreeConstructor()
	_, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Tree construction failed: %v", err)
	}

	// Test GetNodeByPath functionality
	testCases := []struct {
		path     string
		expected string
		isDir    bool
	}{
		{".", ".", true},
		{"file.txt", "file.txt", false},
		{"dir", "dir", true},
		{"dir/nested.txt", "nested.txt", false},
	}

	for _, tc := range testCases {
		node := constructor.GetNodeByPath(tc.path)
		if node == nil {
			t.Errorf("GetNodeByPath(%q) returned nil", tc.path)
			continue
		}
		if node.Name != tc.expected {
			t.Errorf("GetNodeByPath(%q).Name = %q, expected %q", tc.path, node.Name, tc.expected)
		}
		if node.IsDir != tc.isDir {
			t.Errorf("GetNodeByPath(%q).IsDir = %v, expected %v", tc.path, node.IsDir, tc.isDir)
		}
	}

	// Test non-existent path
	nonExistent := constructor.GetNodeByPath("nonexistent")
	if nonExistent != nil {
		t.Error("GetNodeByPath for non-existent path should return nil")
	}
}

func TestWalkTree(t *testing.T) {
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/walk", IsDir: true, Depth: 0},
		{Path: "a.txt", AbsolutePath: "/walk/a.txt", IsDir: false, Depth: 1},
		{Path: "b.txt", AbsolutePath: "/walk/b.txt", IsDir: false, Depth: 1},
		{Path: "subdir", AbsolutePath: "/walk/subdir", IsDir: true, Depth: 1},
		{Path: "subdir/c.txt", AbsolutePath: "/walk/subdir/c.txt", IsDir: false, Depth: 2},
	}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Tree construction failed: %v", err)
	}

	// Collect all visited nodes during tree walk
	var visitedNodes []string
	err = treeconstruction.WalkTree(root, func(node *types.Node) error {
		visitedNodes = append(visitedNodes, node.RelativePath)
		return nil
	})

	if err != nil {
		t.Fatalf("WalkTree failed: %v", err)
	}

	// Expected order: root first, then children in order they were added
	expectedOrder := []string{".", "a.txt", "b.txt", "subdir", "subdir/c.txt"}

	if len(visitedNodes) != len(expectedOrder) {
		t.Errorf("Expected %d visited nodes, got %d", len(expectedOrder), len(visitedNodes))
	}

	// Check that all expected nodes were visited (order may vary due to map iteration)
	for _, expected := range expectedOrder {
		found := false
		for _, visited := range visitedNodes {
			if visited == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to visit node %q but didn't", expected)
		}
	}
}

func TestCalculateTreeStats(t *testing.T) {
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/stats", IsDir: true, Depth: 0},
		{Path: "file1.txt", AbsolutePath: "/stats/file1.txt", IsDir: false, Depth: 1},
		{Path: "file2.txt", AbsolutePath: "/stats/file2.txt", IsDir: false, Depth: 1},
		{Path: "dir1", AbsolutePath: "/stats/dir1", IsDir: true, Depth: 1},
		{Path: "dir1/subfile.txt", AbsolutePath: "/stats/dir1/subfile.txt", IsDir: false, Depth: 2},
		{Path: "dir2", AbsolutePath: "/stats/dir2", IsDir: true, Depth: 1},
		{Path: "dir2/deep", AbsolutePath: "/stats/dir2/deep", IsDir: true, Depth: 2},
		{Path: "dir2/deep/nested.txt", AbsolutePath: "/stats/dir2/deep/nested.txt", IsDir: false, Depth: 3},
	}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Tree construction failed: %v", err)
	}

	stats := treeconstruction.CalculateTreeStats(root)

	// Verify statistics
	if stats.TotalNodes != 8 {
		t.Errorf("Expected 8 total nodes, got %d", stats.TotalNodes)
	}
	if stats.DirectoryNodes != 4 { // ., dir1, dir2, dir2/deep
		t.Errorf("Expected 4 directory nodes, got %d", stats.DirectoryNodes)
	}
	if stats.FileNodes != 4 { // file1.txt, file2.txt, dir1/subfile.txt, dir2/deep/nested.txt
		t.Errorf("Expected 4 file nodes, got %d", stats.FileNodes)
	}
	if stats.MaxDepth != 3 { // dir2/deep/nested.txt is at depth 3
		t.Errorf("Expected max depth 3, got %d", stats.MaxDepth)
	}

	// Average children should be calculated only for directories that have children
	// Root has 4 children (file1.txt, file2.txt, dir1, dir2)
	// dir1 has 1 child (subfile.txt)
	// dir2 has 1 child (deep)
	// dir2/deep has 1 child (nested.txt)
	// Average = (4 + 1 + 1 + 1) / 4 = 1.75
	expectedAverage := 1.75
	if stats.AverageChildren != expectedAverage {
		t.Errorf("Expected average children %.2f, got %.2f", expectedAverage, stats.AverageChildren)
	}
}

func TestEmptyPathList(t *testing.T) {
	// Test edge case of empty path list
	pathInfos := []pathcollection.PathInfo{}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Empty path list construction failed: %v", err)
	}

	if root != nil {
		t.Error("Expected nil root for empty path list")
	}
}

func TestSingleRootPath(t *testing.T) {
	// Test edge case of only root path
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/single", IsDir: true, Depth: 0},
	}

	constructor := treeconstruction.NewTreeConstructor()
	root, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Single root construction failed: %v", err)
	}

	if root == nil {
		t.Fatal("Root should not be nil")
	}
	if root.Name != "." {
		t.Errorf("Expected root name '.', got %q", root.Name)
	}
	if len(root.Children) != 0 {
		t.Errorf("Expected root to have no children, got %d", len(root.Children))
	}
	if root.Parent != nil {
		t.Error("Root should have no parent")
	}
}

func TestGetAllNodes(t *testing.T) {
	pathInfos := []pathcollection.PathInfo{
		{Path: ".", AbsolutePath: "/all", IsDir: true, Depth: 0},
		{Path: "file.txt", AbsolutePath: "/all/file.txt", IsDir: false, Depth: 1},
		{Path: "dir", AbsolutePath: "/all/dir", IsDir: true, Depth: 1},
	}

	constructor := treeconstruction.NewTreeConstructor()
	_, err := constructor.BuildTree(pathInfos)

	if err != nil {
		t.Fatalf("Tree construction failed: %v", err)
	}

	allNodes := constructor.GetAllNodes()

	if len(allNodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(allNodes))
	}

	// Verify all expected paths are present
	expectedPaths := []string{".", "file.txt", "dir"}
	foundPaths := make(map[string]bool)

	for _, node := range allNodes {
		foundPaths[node.RelativePath] = true
	}

	for _, expectedPath := range expectedPaths {
		if !foundPaths[expectedPath] {
			t.Errorf("Expected path %q not found in all nodes", expectedPath)
		}
	}
}