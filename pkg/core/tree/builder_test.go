package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/info"
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
	infoContent := `README.md Project readme file

src/main.go Main application entry point`

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
		"file2.txt",            // files in dir1
		"file3.txt",
		"file1.txt", // file in root
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

func TestBuilder_DepthLimit(t *testing.T) {
	// Create a deeply nested directory structure
	tempDir := t.TempDir()

	// Create nested directories: level1/level2/level3/level4/deep.txt
	deepPath := filepath.Join(tempDir, "level1", "level2", "level3", "level4")
	err := os.MkdirAll(deepPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep directory structure: %v", err)
	}

	// Create files at different levels
	testFiles := []string{
		"root.txt",
		"level1/file1.txt",
		"level1/level2/file2.txt",
		"level1/level2/level3/file3.txt",
		"level1/level2/level3/level4/deep.txt",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test with depth limit of 2
	annotations := make(map[string]*info.Annotation) // No annotations for this test
	builder, err := NewBuilderWithOptions(tempDir, annotations, "", 2)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}

	// Verify depth limit is respected
	maxDepthFound := 0
	err = WalkTree(root, func(node *Node, depth int) error {
		if depth > maxDepthFound {
			maxDepthFound = depth
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	// With depth limit 2, we should see: root(0) -> level1(1) -> level2(2)
	// but not file2.txt(3) or deeper
	if maxDepthFound > 2 {
		t.Errorf("Expected max depth 2, but found depth %d", maxDepthFound)
	}

	// Verify that level2 directory exists (at depth 2) but file2.txt doesn't (at depth 3)
	var foundLevel2, foundFile2, foundFile3 bool
	err = WalkTree(root, func(node *Node, depth int) error {
		if node.Name == "level2" {
			foundLevel2 = true
		}
		if node.Name == "file2.txt" {
			foundFile2 = true
		}
		if node.Name == "file3.txt" {
			foundFile3 = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	if !foundLevel2 {
		t.Error("Expected to find level2 directory at depth 2")
	}

	if foundFile2 {
		t.Error("Did not expect to find file2.txt (should be beyond depth limit at depth 3)")
	}

	if foundFile3 {
		t.Error("Did not expect to find file3.txt (should be beyond depth limit)")
	}
}

func TestBuilder_NoDepthLimit(t *testing.T) {
	// Test that -1 means no depth limit
	tempDir := t.TempDir()

	// Create nested directories
	deepPath := filepath.Join(tempDir, "a", "b", "c", "d", "e")
	err := os.MkdirAll(deepPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create deep directory structure: %v", err)
	}

	// Create a file at the deepest level
	deepFile := filepath.Join(deepPath, "deep.txt")
	err = os.WriteFile(deepFile, []byte("deep content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create deep file: %v", err)
	}

	// Build with no depth limit (-1)
	annotations := make(map[string]*info.Annotation)
	builder, err := NewBuilderWithOptions(tempDir, annotations, "", -1)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree: %v", err)
	}

	// Verify the deep file is found
	var foundDeepFile bool
	err = WalkTree(root, func(node *Node, depth int) error {
		if node.Name == "deep.txt" {
			foundDeepFile = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	if !foundDeepFile {
		t.Error("Expected to find deep.txt when no depth limit is set")
	}
}

func TestBuildTreeNestedWithOptions(t *testing.T) {
	// Test the convenience function with depth limit
	tempDir := t.TempDir()

	// Create test structure
	testFiles := []string{
		"file1.txt",
		"dir1/file2.txt",
		"dir1/dir2/file3.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Test with depth limit 1
	root, err := BuildTreeNestedWithOptions(tempDir, "", 1)
	if err != nil {
		t.Fatalf("Failed to build tree with options: %v", err)
	}

	// Should see file1.txt and dir1, but not file2.txt or deeper
	var foundFile1, foundFile2, foundFile3 bool
	err = WalkTree(root, func(node *Node, depth int) error {
		switch node.Name {
		case "file1.txt":
			foundFile1 = true
		case "file2.txt":
			foundFile2 = true
		case "file3.txt":
			foundFile3 = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	if !foundFile1 {
		t.Error("Expected to find file1.txt at depth 1")
	}

	if foundFile2 {
		t.Error("Did not expect to find file2.txt (beyond depth limit)")
	}

	if foundFile3 {
		t.Error("Did not expect to find file3.txt (beyond depth limit)")
	}
}

func TestBuilder_IgnoreWithAnnotationsOverride(t *testing.T) {
	// Test that files with annotations are shown even if they match ignore patterns
	tempDir := t.TempDir()

	// Create test files and directories
	testFiles := []string{
		"README.md",
		"main.go",
		"debug.log",
		"ignored.tmp",
		".venv/pyvenv.cfg",
		".venv/lib/python3.9/site-packages/requests/__init__.py",
		"build/output.bin",
		"src/app.go",
		"src/test.log",
		"node_modules/package/index.js",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Create .gitignore with patterns that would normally exclude some files
	ignoreFile := filepath.Join(tempDir, ".gitignore")
	ignoreContent := `*.log
*.tmp
.venv/
build/
node_modules/
`

	err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create annotations for some files that would normally be ignored
	annotations := map[string]*info.Annotation{
		"debug.log": {
			Path:        "debug.log",
			Title:       "Debug Log File",
			Description: "Debug Log File\nContains application debug information",
		},
		".venv": {
			Path:        ".venv",
			Title:       "Python Virtual Environment",
			Description: "Python Virtual Environment\nContains isolated Python dependencies for this project",
		},
		"ignored.tmp": {
			Path:        "ignored.tmp",
			Title:       "Temporary Work File",
			Description: "Temporary Work File\nUsed for intermediate processing",
		},
		"build/output.bin": {
			Path:        "build/output.bin",
			Title:       "Build Output",
			Description: "Build Output\nCompiled binary from the build process",
		},
	}

	// Build tree with annotations and ignore support
	builder, err := NewBuilderWithIgnore(tempDir, annotations, ignoreFile)
	if err != nil {
		t.Fatalf("Failed to create builder: %v", err)
	}

	root, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build tree with ignore: %v", err)
	}

	// Collect all files found in the tree
	foundFiles := make(map[string]bool)
	foundDirs := make(map[string]bool)
	err = WalkTree(root, func(node *Node, depth int) error {
		if node.IsDir {
			if node.RelativePath != "." {
				foundDirs[node.RelativePath] = true
			}
		} else {
			foundFiles[node.RelativePath] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk tree: %v", err)
	}

	// Files that should be present (either not ignored or have annotations)
	expectedPresent := []string{
		"README.md",        // not ignored
		"main.go",          // not ignored
		"debug.log",        // ignored by *.log BUT has annotation - should be present
		"ignored.tmp",      // ignored by *.tmp BUT has annotation - should be present
		"build/output.bin", // ignored by build/ BUT has annotation - should be present
		"src/app.go",       // not ignored
	}

	// Directories that should be present
	expectedPresentDirs := []string{
		".venv", // ignored by .venv/ BUT has annotation - should be present
		"src",   // not ignored
		"build", // directory itself should be present since build/output.bin has annotation
	}

	// Files that should NOT be present (ignored and no annotations)
	expectedAbsent := []string{
		"src/test.log",     // ignored by *.log and no annotation
		".venv/pyvenv.cfg", // ignored by .venv/ and no annotation
		".venv/lib/python3.9/site-packages/requests/__init__.py", // ignored by .venv/ and no annotation
		"node_modules/package/index.js",                          // ignored by node_modules/ and no annotation
	}

	// Verify expected present files
	for _, file := range expectedPresent {
		if !foundFiles[file] {
			t.Errorf("Expected file %s to be present (should override ignore due to annotation) but it was filtered out", file)
		}
	}

	// Verify expected present directories
	for _, dir := range expectedPresentDirs {
		if !foundDirs[dir] {
			t.Errorf("Expected directory %s to be present (should override ignore due to annotation) but it was filtered out", dir)
		}
	}

	// Verify expected absent files
	for _, file := range expectedAbsent {
		if foundFiles[file] {
			t.Errorf("Expected file %s to be filtered out (ignored and no annotation) but it was present", file)
		}
	}

	// Verify that annotated files have their annotations
	err = WalkTree(root, func(node *Node, depth int) error {
		if node.RelativePath == "debug.log" {
			if node.Annotation == nil {
				t.Error("debug.log should have annotation")
			} else if node.Annotation.Title != "Debug Log File" {
				t.Errorf("debug.log annotation title mismatch: got %q", node.Annotation.Title)
			}
		}
		if node.RelativePath == ".venv" {
			if node.Annotation == nil {
				t.Error(".venv should have annotation")
			} else if node.Annotation.Title != "Python Virtual Environment" {
				t.Errorf(".venv annotation title mismatch: got %q", node.Annotation.Title)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to verify annotations: %v", err)
	}
}
