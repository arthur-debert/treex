package format

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/tree"
)

// RegistryMockRenderer for testing registry functionality
type RegistryMockRenderer struct {
	format      OutputFormat
	description string
	isTerminal  bool
}

func (m *RegistryMockRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	return "mock output for " + string(m.format), nil
}

func (m *RegistryMockRenderer) Format() OutputFormat {
	return m.format
}

func (m *RegistryMockRenderer) Description() string {
	return m.description
}

func (m *RegistryMockRenderer) IsTerminalFormat() bool {
	return m.isTerminal
}

func TestNewRendererRegistry(t *testing.T) {
	registry := NewRendererRegistry()

	if registry == nil {
		t.Fatal("NewRendererRegistry() returned nil")
	}

	if registry.renderers == nil {
		t.Fatal("Registry renderers map is nil")
	}

	if registry.aliases == nil {
		t.Fatal("Registry aliases map is nil")
	}

	// Check default aliases are present
	expectedAliases := []string{"color", "minimal", "no-color", "json", "yaml"}
	for _, alias := range expectedAliases {
		if _, exists := registry.aliases[alias]; !exists {
			t.Errorf("Expected alias %q not found", alias)
		}
	}
}

func TestRendererRegistry_Register(t *testing.T) {
	registry := NewRendererRegistry()

	// Test successful registration
	registryMockRenderer := &RegistryMockRenderer{
		format:      "test-format",
		description: "Test format",
		isTerminal:  true,
	}

	err := registry.Register(registryMockRenderer)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Test retrieval
	retrieved, err := registry.GetRenderer("test-format")
	if err != nil {
		t.Fatalf("GetRenderer() failed: %v", err)
	}

	if retrieved != registryMockRenderer {
		t.Error("Retrieved renderer is not the same as registered")
	}

	// Test registration with nil renderer
	err = registry.Register(nil)
	if err == nil {
		t.Error("Expected error when registering nil renderer")
	}

	// Test registration with empty format
	emptyFormatRenderer := &RegistryMockRenderer{format: "", description: "Empty"}
	err = registry.Register(emptyFormatRenderer)
	if err == nil {
		t.Error("Expected error when registering renderer with empty format")
	}
}

func TestRendererRegistry_ParseFormat(t *testing.T) {
	registry := NewRendererRegistry()

	// Register renderers for the formats we want to test aliases for
	colorRenderer := &RegistryMockRenderer{format: FormatColor, description: "Color format", isTerminal: true}
	minimalRenderer := &RegistryMockRenderer{format: FormatMinimal, description: "Minimal format", isTerminal: true}
	testRenderer := &RegistryMockRenderer{format: "test", description: "Test format", isTerminal: true}

	_ = registry.Register(colorRenderer)
	_ = registry.Register(minimalRenderer)
	_ = registry.Register(testRenderer)

	// Test direct format parsing
	format, err := registry.ParseFormat("test")
	if err != nil {
		t.Fatalf("ParseFormat() failed for direct format: %v", err)
	}
	if format != "test" {
		t.Errorf("Expected 'test', got %q", format)
	}

	// Test alias parsing
	format, err = registry.ParseFormat("color")
	if err != nil {
		t.Fatalf("ParseFormat() failed for alias: %v", err)
	}
	if format != FormatColor {
		t.Errorf("Expected %q, got %q", FormatColor, format)
	}

	// Test case insensitive parsing
	format, err = registry.ParseFormat("COLOR")
	if err != nil {
		t.Fatalf("ParseFormat() failed for uppercase: %v", err)
	}
	if format != FormatColor {
		t.Errorf("Expected %q, got %q", FormatColor, format)
	}

	// Test whitespace handling
	format, err = registry.ParseFormat("  minimal  ")
	if err != nil {
		t.Fatalf("ParseFormat() failed for whitespace: %v", err)
	}
	if format != FormatMinimal {
		t.Errorf("Expected %q, got %q", FormatMinimal, format)
	}

	// Test alternative aliases
	format, err = registry.ParseFormat("colorful")
	if err != nil {
		t.Fatalf("ParseFormat() failed for colorful alias: %v", err)
	}
	if format != FormatColor {
		t.Errorf("Expected %q for colorful alias, got %q", FormatColor, format)
	}

	// Test unknown format
	_, err = registry.ParseFormat("unknown")
	if err == nil {
		t.Error("Expected error for unknown format")
	}
}

func TestRendererRegistry_ListFormats(t *testing.T) {
	registry := NewRendererRegistry()

	// Register test renderers
	testRenderers := []*RegistryMockRenderer{
		{format: "format1", description: "Description 1", isTerminal: true},
		{format: "format2", description: "Description 2", isTerminal: false},
	}

	for _, renderer := range testRenderers {
		_ = registry.Register(renderer)
	}

	formats := registry.ListFormats()

	if len(formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(formats))
	}

	if formats["format1"] != "Description 1" {
		t.Errorf("Expected 'Description 1', got %q", formats["format1"])
	}

	if formats["format2"] != "Description 2" {
		t.Errorf("Expected 'Description 2', got %q", formats["format2"])
	}
}

func TestRendererRegistry_GetTerminalAndDataFormats(t *testing.T) {
	registry := NewRendererRegistry()

	// Register mixed renderers
	testRenderers := []*RegistryMockRenderer{
		{format: "terminal1", description: "Terminal 1", isTerminal: true},
		{format: "terminal2", description: "Terminal 2", isTerminal: true},
		{format: "data1", description: "Data 1", isTerminal: false},
		{format: "data2", description: "Data 2", isTerminal: false},
	}

	for _, renderer := range testRenderers {
		_ = registry.Register(renderer)
	}

	terminalFormats := registry.GetTerminalFormats()
	dataFormats := registry.GetDataFormats()

	if len(terminalFormats) != 2 {
		t.Errorf("Expected 2 terminal formats, got %d", len(terminalFormats))
	}

	if len(dataFormats) != 2 {
		t.Errorf("Expected 2 data formats, got %d", len(dataFormats))
	}

	// Check terminal formats contain expected formats
	terminalMap := make(map[OutputFormat]bool)
	for _, format := range terminalFormats {
		terminalMap[format] = true
	}

	if !terminalMap["terminal1"] || !terminalMap["terminal2"] {
		t.Error("Terminal formats missing expected formats")
	}

	// Check data formats contain expected formats
	dataMap := make(map[OutputFormat]bool)
	for _, format := range dataFormats {
		dataMap[format] = true
	}

	if !dataMap["data1"] || !dataMap["data2"] {
		t.Error("Data formats missing expected formats")
	}
}

func TestRendererRegistry_ValidateFormat(t *testing.T) {
	registry := NewRendererRegistry()

	// Register a test renderer
	registryTestRenderer := &RegistryMockRenderer{format: "test", description: "Test", isTerminal: true}
	_ = registry.Register(registryTestRenderer)

	// Test valid format
	err := registry.ValidateFormat("test")
	if err != nil {
		t.Errorf("ValidateFormat() failed for valid format: %v", err)
	}

	// Test invalid format
	err = registry.ValidateFormat("invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}

	// Check error message contains available formats
	if !strings.Contains(err.Error(), "test") {
		t.Error("Error message should contain available formats")
	}
}

func TestRendererRegistry_DefaultFormat(t *testing.T) {
	registry := NewRendererRegistry()

	defaultFormat := registry.DefaultFormat()
	if defaultFormat != FormatColor {
		t.Errorf("Expected default format %q, got %q", FormatColor, defaultFormat)
	}
}

func TestRendererRegistry_GetFormatHelp(t *testing.T) {
	registry := NewRendererRegistry()

	// Register mixed renderers
	_ = registry.Register(&RegistryMockRenderer{format: "terminal", description: "Terminal format", isTerminal: true})
	_ = registry.Register(&RegistryMockRenderer{format: "data", description: "Data format", isTerminal: false})

	help := registry.GetFormatHelp()

	if help == "" {
		t.Error("GetFormatHelp() returned empty string")
	}

	// Check help contains format information
	if !strings.Contains(help, "terminal") {
		t.Error("Help should contain terminal format")
	}

	if !strings.Contains(help, "data") {
		t.Error("Help should contain data format")
	}

	if !strings.Contains(help, "Terminal formats:") {
		t.Error("Help should contain terminal section")
	}

	if !strings.Contains(help, "Data formats:") {
		t.Error("Help should contain data section")
	}
}

func TestGetDefaultRegistry(t *testing.T) {
	registry := GetDefaultRegistry()

	if registry == nil {
		t.Fatal("GetDefaultRegistry() returned nil")
	}

	// Test that built-in renderers are registered
	expectedFormats := []OutputFormat{FormatColor, FormatMinimal, FormatNoColor}

	for _, format := range expectedFormats {
		renderer, err := registry.GetRenderer(format)
		if err != nil {
			t.Errorf("Expected format %q not registered: %v", format, err)
		}

		if renderer == nil {
			t.Errorf("Renderer for format %q is nil", format)
		}
	}

	// Test singleton behavior
	registry2 := GetDefaultRegistry()
	if registry != registry2 {
		t.Error("GetDefaultRegistry() should return the same instance")
	}
}

func TestRenderConvenienceFunction(t *testing.T) {
	// Create a simple test tree
	root := &tree.Node{
		Name:  "test",
		IsDir: true,
	}

	options := RenderOptions{
		Format:   FormatColor,
		SafeMode: true,
	}

	output, err := Render(root, options)
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	if output == "" {
		t.Error("Render() returned empty output")
	}

	// Test with empty format (should use default)
	options.Format = ""
	output, err = Render(root, options)
	if err != nil {
		t.Fatalf("Render() with empty format failed: %v", err)
	}

	if output == "" {
		t.Error("Render() with empty format returned empty output")
	}
}
