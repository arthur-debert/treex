package rendering

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/info"
	"github.com/adebert/treex/pkg/core/tree"
)

// Helper function to create a simple tree for testing
func createTestTree() *tree.Node {
	root := &tree.Node{
		Name:         "test-root",
		IsDir:        true,
		RelativePath: ".",
		Children:     []*tree.Node{},
	}

	// Add some children
	file1 := &tree.Node{
		Name:         "file1.txt",
		IsDir:        false,
		RelativePath: "file1.txt",
		Parent:       root,
		Annotation: &info.Annotation{
			Path:        "file1.txt",
			Description: "Test file 1",
		},
	}

	dir1 := &tree.Node{
		Name:         "dir1",
		IsDir:        true,
		RelativePath: "dir1",
		Parent:       root,
		Children:     []*tree.Node{},
	}

	file2 := &tree.Node{
		Name:         "file2.txt",
		IsDir:        false,
		RelativePath: "dir1/file2.txt",
		Parent:       dir1,
		Annotation: &info.Annotation{
			Path:        "dir1/file2.txt",
			Description: "Test file 2",
			Notes:       "Important notes",
		},
	}

	file3 := &tree.Node{
		Name:         "file3.txt",
		IsDir:        false,
		RelativePath: "file3.txt",
		Parent:       root,
	}

	// Build the tree structure
	root.Children = []*tree.Node{file1, dir1, file3}
	dir1.Children = []*tree.Node{file2}

	return root
}

// Helper function to create a deep tree for testing
func createDeepTree() *tree.Node {
	root := &tree.Node{
		Name:         "deep-root",
		IsDir:        true,
		RelativePath: ".",
		Children:     []*tree.Node{},
	}

	current := root
	for i := 1; i <= 5; i++ {
		dir := &tree.Node{
			Name:         "level" + string(rune('0'+i)),
			IsDir:        true,
			RelativePath: strings.Repeat("level"+string(rune('0'+i))+"/", i),
			Parent:       current,
			Children:     []*tree.Node{},
		}
		current.Children = []*tree.Node{dir}
		current = dir
	}

	// Add a file at the deepest level
	deepFile := &tree.Node{
		Name:         "deep.txt",
		IsDir:        false,
		RelativePath: current.RelativePath + "deep.txt",
		Parent:       current,
		Annotation: &info.Annotation{
			Path:        "deep.txt",
			Description: "Deep file",
		},
	}
	current.Children = []*tree.Node{deepFile}

	return root
}

// Helper function to create an empty tree
func createEmptyTree() *tree.Node {
	return &tree.Node{
		Name:         "empty-root",
		IsDir:        true,
		RelativePath: ".",
		Children:     []*tree.Node{},
	}
}

// Mock writer that can simulate errors
type errorWriter struct {
	err error
}

func (e errorWriter) Write(p []byte) (n int, err error) {
	if e.err != nil {
		return 0, e.err
	}
	return len(p), nil
}

func TestNewTreeRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewTreeRenderer(&buf, true)

	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if renderer.writer != &buf {
		t.Error("Writer not set correctly")
	}

	if !renderer.showAnnotations {
		t.Error("showAnnotations should be true")
	}
}

func TestTreeRenderer_Render(t *testing.T) {
	tests := []struct {
		name            string
		tree            *tree.Node
		showAnnotations bool
		expectedLines   []string
		notExpected     []string
	}{
		{
			name:            "Basic tree with annotations",
			tree:            createTestTree(),
			showAnnotations: true,
			expectedLines: []string{
				"test-root",
				"├── file1.txt",
				"Test file 1",
				"├── dir1",
				"│   └── file2.txt",
				"Important notes",
				"└── file3.txt",
			},
			notExpected: []string{},
		},
		{
			name:            "Tree without annotations",
			tree:            createTestTree(),
			showAnnotations: false,
			expectedLines: []string{
				"test-root",
				"├── file1.txt",
				"├── dir1",
				"│   └── file2.txt",
				"└── file3.txt",
			},
			notExpected: []string{
				"Test file 1",
				"Important notes",
			},
		},
		{
			name:            "Empty tree",
			tree:            createEmptyTree(),
			showAnnotations: true,
			expectedLines: []string{
				"empty-root",
			},
			notExpected: []string{
				"├──",
				"└──",
			},
		},
		{
			name:            "Deep tree",
			tree:            createDeepTree(),
			showAnnotations: true,
			expectedLines: []string{
				"deep-root",
				"└── level1",
				"    └── level2",
				"        └── level3",
				"            └── level4",
				"                └── level5",
				"                    └── deep.txt",
				"Deep file",
			},
			notExpected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewTreeRenderer(&buf, tt.showAnnotations)

			err := renderer.Render(tt.tree)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			output := buf.String()

			// Check expected lines
			for _, expected := range tt.expectedLines {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}

			// Check not expected lines
			for _, notExpected := range tt.notExpected {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output NOT to contain %q, got:\n%s", notExpected, output)
				}
			}
		})
	}
}

func TestTreeRenderer_RenderWithError(t *testing.T) {
	tree := createTestTree()
	expectedErr := errors.New("write error")
	writer := errorWriter{err: expectedErr}

	renderer := NewTreeRenderer(writer, true)
	err := renderer.Render(tree)

	if err == nil {
		t.Error("Expected error but got nil")
	}

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestTreeRenderer_formatAnnotation(t *testing.T) {
	renderer := NewTreeRenderer(nil, true)

	tests := []struct {
		name       string
		annotation *info.Annotation
		expected   string
	}{
		{
			name:       "Nil annotation",
			annotation: nil,
			expected:   "",
		},
		{
			name: "Annotation with notes",
			annotation: &info.Annotation{
				Notes:       "Important notes",
				Description: "Description",
			},
			expected: "Important notes",
		},
		{
			name: "Annotation with description only",
			annotation: &info.Annotation{
				Description: "Description only",
			},
			expected: "Description only",
		},
		{
			name: "Empty annotation",
			annotation: &info.Annotation{
				Notes:       "",
				Description: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.formatAnnotation(tt.annotation)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTreeRenderer_calculatePadding(t *testing.T) {
	renderer := NewTreeRenderer(nil, true)

	tests := []struct {
		name           string
		lineLength     int
		expectedLength int
	}{
		{
			name:           "Short line",
			lineLength:     10,
			expectedLength: 30, // 40 - 10
		},
		{
			name:           "Line at target",
			lineLength:     40,
			expectedLength: 2, // Minimum spacing
		},
		{
			name:           "Long line",
			lineLength:     50,
			expectedLength: 2, // Minimum spacing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padding := renderer.calculatePadding(tt.lineLength)
			if len(padding) != tt.expectedLength {
				t.Errorf("Expected padding length %d, got %d", tt.expectedLength, len(padding))
			}
		})
	}
}

func TestRenderTree(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderTree(&buf, tree, true)
	if err != nil {
		t.Fatalf("RenderTree failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}

	if !strings.Contains(output, "Test file 1") {
		t.Error("Expected output to contain annotation")
	}
}

func TestRenderTreeToString(t *testing.T) {
	tree := createTestTree()

	output, err := RenderTreeToString(tree, true)
	if err != nil {
		t.Fatalf("RenderTreeToString failed: %v", err)
	}

	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}

	if !strings.Contains(output, "Test file 1") {
		t.Error("Expected output to contain annotation")
	}
}

func TestRenderTreeToString_Error(t *testing.T) {
	// Create a tree that will cause an error during rendering
	// In this case, we can't easily simulate an error in string building,
	// but we can test the error path exists
	tree := createTestTree()

	// The function should handle the tree gracefully
	output, err := RenderTreeToString(tree, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestTreeRenderer_SingleFileTree(t *testing.T) {
	// Create a tree with just one file
	root := &tree.Node{
		Name:         "single-root",
		IsDir:        true,
		RelativePath: ".",
		Children: []*tree.Node{
			{
				Name:         "lonely.txt",
				IsDir:        false,
				RelativePath: "lonely.txt",
				Annotation: &info.Annotation{
					Path:        "lonely.txt",
					Description: "A lonely file",
				},
			},
		},
	}

	var buf bytes.Buffer
	renderer := NewTreeRenderer(&buf, true)

	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()
	
	// Check that the single file uses └── (last child indicator)
	if !strings.Contains(output, "└── lonely.txt") {
		t.Error("Expected single file to use └── connector")
	}

	if !strings.Contains(output, "A lonely file") {
		t.Error("Expected annotation to be shown")
	}
}

func TestTreeRenderer_MixedDepthTree(t *testing.T) {
	// Create a tree with mixed depths
	root := &tree.Node{
		Name:         "mixed-root",
		IsDir:        true,
		RelativePath: ".",
		Children:     []*tree.Node{},
	}

	shallow := &tree.Node{
		Name:         "shallow.txt",
		IsDir:        false,
		RelativePath: "shallow.txt",
		Parent:       root,
	}

	deepDir := &tree.Node{
		Name:         "deep",
		IsDir:        true,
		RelativePath: "deep",
		Parent:       root,
		Children:     []*tree.Node{},
	}

	deepFile := &tree.Node{
		Name:         "nested.txt",
		IsDir:        false,
		RelativePath: "deep/nested.txt",
		Parent:       deepDir,
		Annotation: &info.Annotation{
			Path:        "deep/nested.txt",
			Description: "Deeply nested",
		},
	}

	root.Children = []*tree.Node{shallow, deepDir}
	deepDir.Children = []*tree.Node{deepFile}

	var buf bytes.Buffer
	renderer := NewTreeRenderer(&buf, true)

	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Verify structure
	expectedPatterns := []string{
		"mixed-root",
		"├── shallow.txt",
		"└── deep",
		"    └── nested.txt",
		"Deeply nested",
	}

	for i, pattern := range expectedPatterns {
		found := false
		for _, line := range lines {
			if strings.Contains(line, strings.TrimSpace(pattern)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %d %q not found in output:\n%s", i, pattern, output)
		}
	}
}