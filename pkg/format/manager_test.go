package format

import (
	"testing"

	"github.com/adebert/treex/pkg/info"
	"github.com/adebert/treex/pkg/tree"
)

func TestNewRendererManager(t *testing.T) {
	manager := NewRendererManager()

	if manager == nil {
		t.Fatal("NewRendererManager() returned nil")
	}

	if manager.registry == nil {
		t.Fatal("RendererManager registry is nil")
	}
}

func TestNewRendererManagerWithRegistry(t *testing.T) {
	customRegistry := NewRendererRegistry()
	mockRenderer := &MockRenderer{format: "custom", description: "Custom", isTerminal: true}
	_ = customRegistry.Register(mockRenderer)

	manager := NewRendererManagerWithRegistry(customRegistry)

	if manager == nil {
		t.Fatal("NewRendererManagerWithRegistry() returned nil")
	}

	if manager.registry != customRegistry {
		t.Error("RendererManager should use provided registry")
	}
}

func TestRendererManager_RenderTree(t *testing.T) {
	manager := NewRendererManager()

	// Create a simple test tree
	root := &tree.Node{
		Name:  "test",
		IsDir: true,
		Children: []*tree.Node{
			{
				Name:  "file.txt",
				IsDir: false,
				Annotation: &info.Annotation{
					Path:        "file.txt",
					Title:       "Test File",
					Description: "A test file",
				},
			},
		},
	}

	// Test basic rendering with default format
	request := RenderRequest{
		Tree:          root,
		Format:        "", // Should use default
		Verbose:       false,
		SafeMode:      true,
		TerminalWidth: 80,
	}

	response, err := manager.RenderTree(request)
	if err != nil {
		t.Fatalf("RenderTree() failed: %v", err)
	}

	if response == nil {
		t.Fatal("RenderTree() returned nil response")
	}

	if response.Output == "" {
		t.Error("RenderTree() returned empty output")
	}

	if response.Format == "" {
		t.Error("RenderTree() response should include format")
	}

	if response.Stats == nil {
		t.Error("RenderTree() response should include stats")
	}

	if response.Stats.NodesRendered != 2 {
		t.Errorf("Expected 2 nodes rendered, got %d", response.Stats.NodesRendered)
	}

	if response.Stats.AnnotationsFound != 1 {
		t.Errorf("Expected 1 annotation found, got %d", response.Stats.AnnotationsFound)
	}
}

func TestRendererManager_RenderTreeWithSpecificFormat(t *testing.T) {
	manager := NewRendererManager()

	root := &tree.Node{
		Name:  "test",
		IsDir: true,
	}

	// Test with specific format
	request := RenderRequest{
		Tree:          root,
		Format:        FormatNoColor,
		SafeMode:      true,
		TerminalWidth: 80,
	}

	response, err := manager.RenderTree(request)
	if err != nil {
		t.Fatalf("RenderTree() with specific format failed: %v", err)
	}

	if response.Format != FormatNoColor {
		t.Errorf("Expected format %q, got %q", FormatNoColor, response.Format)
	}
}

func TestRendererManager_RenderTreeWithLegacyFlags(t *testing.T) {
	manager := NewRendererManager()

	root := &tree.Node{
		Name:  "test",
		IsDir: true,
	}

	// Test legacy no-color flag
	request := RenderRequest{
		Tree:          root,
		LegacyNoColor: true,
		SafeMode:      true,
		TerminalWidth: 80,
	}

	response, err := manager.RenderTree(request)
	if err != nil {
		t.Fatalf("RenderTree() with legacy no-color failed: %v", err)
	}

	if response.Format != FormatNoColor {
		t.Errorf("Expected format %q from legacy flag, got %q", FormatNoColor, response.Format)
	}

	// Test legacy minimal flag
	request.LegacyNoColor = false
	request.LegacyMinimal = true

	response, err = manager.RenderTree(request)
	if err != nil {
		t.Fatalf("RenderTree() with legacy minimal failed: %v", err)
	}

	if response.Format != FormatMinimal {
		t.Errorf("Expected format %q from legacy flag, got %q", FormatMinimal, response.Format)
	}
}

func TestRendererManager_RenderTreeWithInvalidFormat(t *testing.T) {
	manager := NewRendererManager()

	root := &tree.Node{
		Name:  "test",
		IsDir: true,
	}

	request := RenderRequest{
		Tree:   root,
		Format: "invalid-format",
	}

	_, err := manager.RenderTree(request)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestRendererManager_SelectFormat(t *testing.T) {
	manager := NewRendererManager()

	// Test explicit format selection
	request := RenderRequest{Format: FormatColor}
	format, err := manager.selectFormat(request)
	if err != nil {
		t.Fatalf("selectFormat() with explicit format failed: %v", err)
	}
	if format != FormatColor {
		t.Errorf("Expected %q, got %q", FormatColor, format)
	}

	// Test legacy no-color
	request = RenderRequest{LegacyNoColor: true}
	format, err = manager.selectFormat(request)
	if err != nil {
		t.Fatalf("selectFormat() with legacy no-color failed: %v", err)
	}
	if format != FormatNoColor {
		t.Errorf("Expected %q, got %q", FormatNoColor, format)
	}

	// Test legacy minimal
	request = RenderRequest{LegacyMinimal: true}
	format, err = manager.selectFormat(request)
	if err != nil {
		t.Fatalf("selectFormat() with legacy minimal failed: %v", err)
	}
	if format != FormatMinimal {
		t.Errorf("Expected %q, got %q", FormatMinimal, format)
	}

	// Test default format selection
	request = RenderRequest{}
	format, err = manager.selectFormat(request)
	if err != nil {
		t.Fatalf("selectFormat() with defaults failed: %v", err)
	}
	if format != FormatColor {
		t.Errorf("Expected default format %q, got %q", FormatColor, format)
	}

	// Test invalid format
	request = RenderRequest{Format: "invalid"}
	_, err = manager.selectFormat(request)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestRendererManager_CountNodes(t *testing.T) {
	manager := NewRendererManager()

	// Test nil node
	count := manager.countNodes(nil)
	if count != 0 {
		t.Errorf("Expected 0 for nil node, got %d", count)
	}

	// Test single node
	single := &tree.Node{Name: "single", IsDir: false}
	count = manager.countNodes(single)
	if count != 1 {
		t.Errorf("Expected 1 for single node, got %d", count)
	}

	// Test tree with children
	root := &tree.Node{
		Name:  "root",
		IsDir: true,
		Children: []*tree.Node{
			{Name: "child1", IsDir: false},
			{Name: "child2", IsDir: false},
			{
				Name:  "subdir",
				IsDir: true,
				Children: []*tree.Node{
					{Name: "grandchild", IsDir: false},
				},
			},
		},
	}

	count = manager.countNodes(root)
	if count != 5 { // root + 2 children + subdir + grandchild
		t.Errorf("Expected 5 nodes in tree, got %d", count)
	}
}

func TestRendererManager_CountAnnotations(t *testing.T) {
	manager := NewRendererManager()

	// Test nil node
	count := manager.countAnnotations(nil)
	if count != 0 {
		t.Errorf("Expected 0 annotations for nil node, got %d", count)
	}

	// Test node without annotation
	unannotated := &tree.Node{Name: "test", IsDir: false}
	count = manager.countAnnotations(unannotated)
	if count != 0 {
		t.Errorf("Expected 0 annotations for unannotated node, got %d", count)
	}

	// Test node with annotation
	annotated := &tree.Node{
		Name:       "test",
		IsDir:      false,
		Annotation: &info.Annotation{Path: "test", Description: "Test"},
	}
	count = manager.countAnnotations(annotated)
	if count != 1 {
		t.Errorf("Expected 1 annotation for annotated node, got %d", count)
	}

	// Test tree with mixed annotations
	root := &tree.Node{
		Name:       "root",
		IsDir:      true,
		Annotation: &info.Annotation{Path: "root", Description: "Root"},
		Children: []*tree.Node{
			{Name: "unannotated", IsDir: false},
			{
				Name:       "annotated",
				IsDir:      false,
				Annotation: &info.Annotation{Path: "annotated", Description: "Annotated"},
			},
		},
	}

	count = manager.countAnnotations(root)
	if count != 2 { // root + annotated child
		t.Errorf("Expected 2 annotations in tree, got %d", count)
	}
}

func TestRendererManager_DelegatedMethods(t *testing.T) {
	manager := NewRendererManager()

	// Test GetAvailableFormats
	formats := manager.GetAvailableFormats()
	if len(formats) == 0 {
		t.Error("GetAvailableFormats() returned empty map")
	}

	// Test GetFormatHelp
	help := manager.GetFormatHelp()
	if help == "" {
		t.Error("GetFormatHelp() returned empty string")
	}

	// Test ValidateFormat
	err := manager.ValidateFormat(FormatColor)
	if err != nil {
		t.Errorf("ValidateFormat() failed for valid format: %v", err)
	}

	err = manager.ValidateFormat("invalid")
	if err == nil {
		t.Error("ValidateFormat() should fail for invalid format")
	}

	// Test ParseFormat
	format, err := manager.ParseFormat("color")
	if err != nil {
		t.Errorf("ParseFormat() failed: %v", err)
	}
	if format != FormatColor {
		t.Errorf("Expected %q, got %q", FormatColor, format)
	}

	// Test GetTerminalFormats
	terminalFormats := manager.GetTerminalFormats()
	if len(terminalFormats) == 0 {
		t.Error("GetTerminalFormats() returned empty slice")
	}

	// Test IsTerminalFormat
	if !manager.IsTerminalFormat(FormatColor) {
		t.Error("Expected FormatColor to be a terminal format")
	}

	if manager.IsTerminalFormat("invalid") {
		t.Error("Expected invalid format to not be terminal format")
	}
}

func TestGetDefaultManager(t *testing.T) {
	manager1 := GetDefaultManager()
	manager2 := GetDefaultManager()

	if manager1 == nil {
		t.Fatal("GetDefaultManager() returned nil")
	}

	if manager1 != manager2 {
		t.Error("GetDefaultManager() should return the same instance")
	}
}

func TestRenderTreeWithDefaults(t *testing.T) {
	root := &tree.Node{
		Name:  "test",
		IsDir: true,
	}

	output, err := RenderTreeWithDefaults(root, FormatNoColor, false, true)
	if err != nil {
		t.Fatalf("RenderTreeWithDefaults() failed: %v", err)
	}

	if output == "" {
		t.Error("RenderTreeWithDefaults() returned empty output")
	}
}
