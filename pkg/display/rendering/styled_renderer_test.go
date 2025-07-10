package rendering

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/adebert/treex/pkg/display/styles"
)

func TestNewStyledTreeRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)

	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if renderer.writer != &buf {
		t.Error("Writer not set correctly")
	}

	if !renderer.showAnnotations {
		t.Error("showAnnotations should be true")
	}

	if renderer.styles == nil {
		t.Error("styles should be initialized")
	}

	if renderer.terminalWidth != 80 {
		t.Errorf("Expected default terminal width 80, got %d", renderer.terminalWidth)
	}

	if renderer.tabstop != 0 {
		t.Errorf("Expected initial tabstop 0, got %d", renderer.tabstop)
	}

}

func TestNewStyledTreeRendererWithRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewStyledTreeRendererWithRenderer(&buf, true)

	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if renderer.styleRenderer == nil {
		t.Error("styleRenderer should be initialized")
	}

	if renderer.styles == nil {
		t.Error("styles should be initialized")
	}
}

func TestNewStyledTreeRendererWithAutoTheme(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewStyledTreeRendererWithAutoTheme(&buf, true, false)

	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if renderer.styleRenderer == nil {
		t.Error("styleRenderer should be initialized")
	}

	// Test verbose mode
	renderer2 := NewStyledTreeRendererWithAutoTheme(&buf, true, true)
	if renderer2 == nil {
		t.Fatal("Expected non-nil renderer in verbose mode")
	}
}

func TestNewMinimalStyledTreeRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewMinimalStyledTreeRenderer(&buf, true)

	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if renderer.styleRenderer == nil {
		t.Error("styleRenderer should be initialized")
	}
}

func TestNewNoColorStyledTreeRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewNoColorStyledTreeRenderer(&buf, false)

	if renderer == nil {
		t.Fatal("Expected non-nil renderer")
	}

	if renderer.styleRenderer == nil {
		t.Error("styleRenderer should be initialized")
	}

	if renderer.showAnnotations {
		t.Error("showAnnotations should be false")
	}
}

func TestStyledTreeRenderer_WithMethods(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)

	// Test WithStyles
	customStyles := styles.NewTreeStyles()
	result := renderer.WithStyles(customStyles)
	if result != renderer {
		t.Error("WithStyles should return the same renderer instance")
	}
	if renderer.styles != customStyles {
		t.Error("styles not updated correctly")
	}

	// Test WithTerminalWidth
	result = renderer.WithTerminalWidth(120)
	if result != renderer {
		t.Error("WithTerminalWidth should return the same renderer instance")
	}
	if renderer.terminalWidth != 120 {
		t.Errorf("Expected terminal width 120, got %d", renderer.terminalWidth)
	}

	// Test WithSafeMode
	result = renderer.WithSafeMode(true)
	if result != renderer {
		t.Error("WithSafeMode should return the same renderer instance")
	}
	if !renderer.safeMode {
		t.Error("safeMode should be true")
	}

}

func TestIsProblematicTerminal(t *testing.T) {
	// Save original env vars
	originalTermProgram := os.Getenv("TERM_PROGRAM")
	originalTerm := os.Getenv("TERM")
	originalSafeMode := os.Getenv("TREEX_SAFE_MODE")
	defer func() {
		_ = os.Setenv("TERM_PROGRAM", originalTermProgram)
		_ = os.Setenv("TERM", originalTerm)
		_ = os.Setenv("TREEX_SAFE_MODE", originalSafeMode)
	}()

	tests := []struct {
		name        string
		termProgram string
		term        string
		safeMode    string
		expected    bool
	}{
		{
			name:        "Ghostty terminal program",
			termProgram: "ghostty",
			term:        "xterm-256color",
			safeMode:    "",
			expected:    true,
		},
		{
			name:        "GHOSTTY uppercase",
			termProgram: "GHOSTTY",
			term:        "xterm-256color",
			safeMode:    "",
			expected:    true,
		},
		{
			name:        "Ghostty in TERM",
			termProgram: "",
			term:        "ghostty-256color",
			safeMode:    "",
			expected:    true,
		},
		{
			name:        "Safe mode enabled",
			termProgram: "Terminal.app",
			term:        "xterm-256color",
			safeMode:    "1",
			expected:    true,
		},
		{
			name:        "Safe mode true",
			termProgram: "Terminal.app",
			term:        "xterm-256color",
			safeMode:    "true",
			expected:    true,
		},
		{
			name:        "Normal terminal",
			termProgram: "Terminal.app",
			term:        "xterm-256color",
			safeMode:    "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("TERM_PROGRAM", tt.termProgram)
			_ = os.Setenv("TERM", tt.term)
			if tt.safeMode != "" {
				_ = os.Setenv("TREEX_SAFE_MODE", tt.safeMode)
			} else {
				_ = os.Unsetenv("TREEX_SAFE_MODE")
			}

			result := isProblematicTerminal()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Simple color code",
			input:    "\x1b[31mRed Text\x1b[0m",
			expected: "Red Text",
		},
		{
			name:     "Multiple codes",
			input:    "\x1b[1m\x1b[32mBold Green\x1b[0m Normal",
			expected: "Bold Green Normal",
		},
		{
			name:     "Complex escape sequences",
			input:    "\x1b[2J\x1b[H\x1b[38;5;196mExtended Color\x1b[0m",
			expected: "Extended Color",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStyledTreeRenderer_safeWidth(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)

	tests := []struct {
		name     string
		text     string
		safeMode bool
	}{
		{
			name:     "Plain text normal mode",
			text:     "Hello World",
			safeMode: false,
		},
		{
			name:     "ANSI text normal mode",
			text:     "\x1b[31mRed Text\x1b[0m",
			safeMode: false,
		},
		{
			name:     "Plain text safe mode",
			text:     "Hello World",
			safeMode: true,
		},
		{
			name:     "ANSI text safe mode",
			text:     "\x1b[31mRed Text\x1b[0m",
			safeMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer.safeMode = tt.safeMode
			width := renderer.safeWidth(tt.text)

			// In safe mode, width should equal stripped text length
			if tt.safeMode {
				expectedWidth := len(stripANSI(tt.text))
				if width != expectedWidth {
					t.Errorf("Expected width %d in safe mode, got %d", expectedWidth, width)
				}
			} else {
				// In normal mode, width should be >= 0
				if width < 0 {
					t.Errorf("Width should not be negative, got %d", width)
				}
			}
		})
	}
}

func TestStyledTreeRenderer_Render(t *testing.T) {
	tests := []struct {
		name            string
		tree            *types.Node
		showAnnotations bool
		safeMode        bool
		expectedLines   []string
		notExpected     []string
	}{
		{
			name:            "Basic tree with annotations",
			tree:            createTestTree(),
			showAnnotations: true,
			safeMode:        false,
			expectedLines: []string{
				"test-root",
				"├── file1.txt",
				"Test file", // Glamour may split text
				"├── dir1",
				"│   └── file2.txt",
				"Important", // Glamour may split text
				"└── file3.txt",
			},
			notExpected: []string{},
		},
		{
			name:            "Tree without annotations shown",
			tree:            createTestTree(),
			showAnnotations: false,
			safeMode:        false,
			expectedLines: []string{
				"test-root",
				"├── file1.txt",
				"├── dir1",
				"│   └── file2.txt",
				"└── file3.txt",
			},
			notExpected: []string{
				"Test file", // We're not showing annotations
				"Important", // We're not showing annotations
			},
		},
		{
			name:            "Tree with safe mode",
			tree:            createTestTree(),
			showAnnotations: true,
			safeMode:        true,
			expectedLines: []string{
				"test-root",
				"file1.txt",
				"Test file", // Glamour may split text
			},
			notExpected: []string{},
		},
		{
			name:            "Tree without extra spacing",
			tree:            createTestTree(),
			showAnnotations: true,
			safeMode:        false,
			expectedLines: []string{
				"test-root",
				"├── file1.txt",
				"Test file", // Glamour may split text
				"├── dir1",
			},
			notExpected: []string{},
		},
		{
			name:            "Empty tree",
			tree:            createEmptyTree(),
			showAnnotations: true,
			safeMode:        false,
			expectedLines: []string{
				"empty-root",
			},
			notExpected: []string{
				"├──",
				"└──",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewStyledTreeRenderer(&buf, tt.showAnnotations).
				WithSafeMode(tt.safeMode)

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

func TestStyledTreeRenderer_calculateTabstop(t *testing.T) {
	// Create a tree with varying path lengths
	root := &types.Node{
		Name:         "root",
		IsDir:        true,
		RelativePath: ".",
		Children: []*types.Node{
			{
				Name:         "short.txt",
				IsDir:        false,
				RelativePath: "short.txt",
				Annotation:   &types.Annotation{Path: "short.txt", Notes: "Short"},
			},
			{
				Name:         "very-long-filename-that-exceeds-forty-characters.txt",
				IsDir:        false,
				RelativePath: "very-long-filename-that-exceeds-forty-characters.txt",
				Annotation:   &types.Annotation{Path: "very-long-filename-that-exceeds-forty-characters.txt", Notes: "Long"},
			},
		},
	}

	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)

	// Calculate tabstop
	renderer.calculateTabstop(root)

	// With a long filename, tabstop should be > 40
	if renderer.tabstop <= 40 {
		t.Errorf("Expected tabstop > 40 for long filenames, got %d", renderer.tabstop)
	}

	// Test with only short paths
	shortRoot := &types.Node{
		Name:         "root",
		IsDir:        true,
		RelativePath: ".",
		Children: []*types.Node{
			{
				Name:         "a.txt",
				IsDir:        false,
				RelativePath: "a.txt",
				Annotation:   &types.Annotation{Path: "a.txt", Notes: "A"},
			},
			{
				Name:         "b.txt",
				IsDir:        false,
				RelativePath: "b.txt",
				Annotation:   &types.Annotation{Path: "b.txt", Notes: "B"},
			},
		},
	}

	renderer2 := NewStyledTreeRenderer(&buf, true)
	renderer2.calculateTabstop(shortRoot)

	// With short filenames, tabstop should be 40
	if renderer2.tabstop != 40 {
		t.Errorf("Expected tabstop 40 for short filenames, got %d", renderer2.tabstop)
	}
}

func TestStyledTreeRenderer_formatInlineAnnotation(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)

	tests := []struct {
		name           string
		annotation     *types.Annotation
		expectText     bool
		expectContains []string // Strings that should be in the output
	}{
		{
			name:       "Nil annotation",
			annotation: nil,
			expectText: false,
		},
		{
			name: "Annotation with notes",
			annotation: &types.Annotation{
				Notes: "Important notes",
			},
			expectText:     true,
			expectContains: []string{"Important notes"},
		},
		{
			name: "Annotation with notes only",
			annotation: &types.Annotation{
				Notes: "Description only",
			},
			expectText:     true,
			expectContains: []string{"Description only"},
		},
		{
			name: "Empty annotation",
			annotation: &types.Annotation{
				Notes: "",
			},
			expectText: false,
		},
		{
			name: "Annotation with markdown bold",
			annotation: &types.Annotation{
				Notes: "This has **bold** text",
			},
			expectText:     true,
			expectContains: []string{"bold"}, // Should contain the word "bold" (with styling)
		},
		{
			name: "Annotation with markdown italic",
			annotation: &types.Annotation{
				Notes: "This has *italic* text",
			},
			expectText:     true,
			expectContains: []string{"italic"}, // Should contain the word "italic" (with styling)
		},
		{
			name: "Annotation with markdown code",
			annotation: &types.Annotation{
				Notes: "This has `code` text",
			},
			expectText:     true,
			expectContains: []string{"code"}, // Should contain the word "code" (with styling)
		},
		{
			name: "Annotation with mixed markdown",
			annotation: &types.Annotation{
				Notes: "Command to **add** or *update* entries in `info` files",
			},
			expectText:     true,
			expectContains: []string{"add", "update", "info"}, // All words should be present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.formatInlineAnnotation(tt.annotation)
			if tt.expectText && result == "" {
				t.Error("Expected non-empty result")
			}
			if !tt.expectText && result != "" {
				t.Errorf("Expected empty result, got %q", result)
			}

			// Check for expected content
			// Strip ANSI codes for content checking since Glamour adds styling
			strippedResult := stripANSI(result)
			for _, expected := range tt.expectContains {
				if !strings.Contains(strippedResult, expected) {
					t.Errorf("Expected result to contain %q, got %q (stripped: %q)", expected, result, strippedResult)
				}
			}
		})
	}
}

func TestRenderStyledTree(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderStyledTree(&buf, tree, true)
	if err != nil {
		t.Fatalf("RenderStyledTree failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderStyledTreeToString(t *testing.T) {
	tree := createTestTree()

	output, err := RenderStyledTreeToString(tree, true)
	if err != nil {
		t.Fatalf("RenderStyledTreeToString failed: %v", err)
	}

	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderStyledTreeToStringWithSafeMode(t *testing.T) {
	tree := createTestTree()

	output, err := RenderStyledTreeToStringWithSafeMode(tree, true, true)
	if err != nil {
		t.Fatalf("RenderStyledTreeToStringWithSafeMode failed: %v", err)
	}

	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderStyledTreeWithSafeMode(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderStyledTreeWithSafeMode(&buf, tree, true, true)
	if err != nil {
		t.Fatalf("RenderStyledTreeWithSafeMode failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderStyledTreeWithOptions(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderStyledTreeWithOptions(&buf, tree, true, true)
	if err != nil {
		t.Fatalf("RenderStyledTreeWithOptions failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderMinimalStyledTree(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderMinimalStyledTree(&buf, tree, true)
	if err != nil {
		t.Fatalf("RenderMinimalStyledTree failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderMinimalStyledTreeWithSafeMode(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderMinimalStyledTreeWithSafeMode(&buf, tree, true, true)
	if err != nil {
		t.Fatalf("RenderMinimalStyledTreeWithSafeMode failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderPlainTree(t *testing.T) {
	var buf bytes.Buffer
	tree := createTestTree()

	err := RenderPlainTree(&buf, tree, true)
	if err != nil {
		t.Fatalf("RenderPlainTree failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderPlainTreeToString(t *testing.T) {
	tree := createTestTree()

	output, err := RenderPlainTreeToString(tree, true)
	if err != nil {
		t.Fatalf("RenderPlainTreeToString failed: %v", err)
	}

	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

func TestRenderMinimalStyledTreeToString(t *testing.T) {
	tree := createTestTree()

	output, err := RenderMinimalStyledTreeToString(tree, true, false)
	if err != nil {
		t.Fatalf("RenderMinimalStyledTreeToString failed: %v", err)
	}

	if !strings.Contains(output, "test-root") {
		t.Error("Expected output to contain root name")
	}
}

// Test error handling - using a different name to avoid redeclaration
type styledErrorWriter struct {
	err error
}

func (e styledErrorWriter) Write(p []byte) (n int, err error) {
	if e.err != nil {
		return 0, e.err
	}
	return len(p), nil
}

func TestStyledTreeRenderer_RenderWithError(t *testing.T) {
	tree := createTestTree()
	expectedErr := errors.New("write error")
	writer := styledErrorWriter{err: expectedErr}

	renderer := NewStyledTreeRenderer(writer, true)
	err := renderer.Render(tree)

	if err == nil {
		t.Error("Expected error but got nil")
	}

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// Test complex tree structures
func TestStyledTreeRenderer_ComplexTree(t *testing.T) {
	// Create a complex tree with mixed annotations
	root := &types.Node{
		Name:         "complex-root",
		IsDir:        true,
		RelativePath: ".",
		Children:     []*types.Node{},
	}

	// Directory with annotation
	annotatedDir := &types.Node{
		Name:         "annotated-dir",
		IsDir:        true,
		RelativePath: "annotated-dir",
		Parent:       root,
		Annotation: &types.Annotation{
			Path:  "annotated-dir",
			Notes: "This directory has an annotation",
		},
		Children: []*types.Node{},
	}

	// Files in annotated directory
	for i := 0; i < 3; i++ {
		file := &types.Node{
			Name:         "file" + string(rune('0'+i)) + ".txt",
			IsDir:        false,
			RelativePath: "annotated-dir/file" + string(rune('0'+i)) + ".txt",
			Parent:       annotatedDir,
		}
		if i == 1 {
			file.Annotation = &types.Annotation{
				Path:  file.RelativePath,
				Notes: "Middle file has annotation",
			}
		}
		annotatedDir.Children = append(annotatedDir.Children, file)
	}

	// Regular directory
	regularDir := &types.Node{
		Name:         "regular-dir",
		IsDir:        true,
		RelativePath: "regular-dir",
		Parent:       root,
		Children:     []*types.Node{},
	}

	// Deeply nested structure
	current := regularDir
	for i := 0; i < 3; i++ {
		nested := &types.Node{
			Name:         "nested" + string(rune('0'+i)),
			IsDir:        true,
			RelativePath: current.RelativePath + "/nested" + string(rune('0'+i)),
			Parent:       current,
			Children:     []*types.Node{},
		}
		current.Children = append(current.Children, nested)
		current = nested
	}

	root.Children = []*types.Node{annotatedDir, regularDir}

	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)

	err := renderer.Render(root)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	// Verify structure
	expectedPatterns := []string{
		"complex-root",
		"├── annotated-dir",
		"This directory has", // Glamour may split the text
		"annotation",
		"│   ├── file0.txt",
		"│   ├── file1.txt",
		"Middle file has", // Glamour may split the text
		"annotation",
		"│   └── file2.txt",
		"└── regular-dir",
		"    └── nested0",
		"        └── nested1",
		"            └── nested2",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected pattern %q not found in output:\n%s", pattern, output)
		}
	}
}

// Test safe width timeout simulation
func TestStyledTreeRenderer_safeWidthTimeout(t *testing.T) {
	// This test simulates the timeout behavior by testing with a very short timeout
	// In practice, the 100ms timeout in the real code should be sufficient

	var buf bytes.Buffer
	renderer := NewStyledTreeRenderer(&buf, true)
	renderer.safeMode = false

	// Create a string that might take time to process
	complexString := strings.Repeat("\x1b[38;2;255;0;0m█\x1b[0m", 100)

	// Call safeWidth - it should either complete or timeout and switch to safe mode
	start := time.Now()
	width := renderer.safeWidth(complexString)
	elapsed := time.Since(start)

	// The width should be calculated
	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}

	// If it took more than 100ms, safe mode should be enabled
	if elapsed > 100*time.Millisecond && !renderer.safeMode {
		t.Error("Expected safe mode to be enabled after timeout")
	}
}

// Test terminal width handling
func TestStyledTreeRenderer_TerminalWidth(t *testing.T) {
	// Create a tree with very long annotations
	root := &types.Node{
		Name:         "width-test",
		IsDir:        true,
		RelativePath: ".",
		Children: []*types.Node{
			{
				Name:         "file-with-very-long-annotation.txt",
				IsDir:        false,
				RelativePath: "file-with-very-long-annotation.txt",
				Annotation: &types.Annotation{
					Path:  "file-with-very-long-annotation.txt",
					Notes: strings.Repeat("This is a very long annotation that might wrap. ", 10),
				},
			},
		},
	}

	tests := []struct {
		name          string
		terminalWidth int
	}{
		{
			name:          "Narrow terminal",
			terminalWidth: 40,
		},
		{
			name:          "Standard terminal",
			terminalWidth: 80,
		},
		{
			name:          "Wide terminal",
			terminalWidth: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewStyledTreeRenderer(&buf, true).
				WithTerminalWidth(tt.terminalWidth)

			err := renderer.Render(root)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "width-test") {
				t.Error("Expected output to contain root name")
			}
		})
	}
}
