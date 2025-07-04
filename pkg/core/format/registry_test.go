package format

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/tree"
)

// mockRenderer is a test renderer implementation
type mockRenderer struct {
	format           OutputFormat
	description      string
	isTerminalFormat bool
	renderError      error
	renderOutput     string
}

func (m *mockRenderer) Render(root *tree.Node, options RenderOptions) (string, error) {
	if m.renderError != nil {
		return "", m.renderError
	}
	return m.renderOutput, nil
}

func (m *mockRenderer) Format() OutputFormat {
	return m.format
}

func (m *mockRenderer) Description() string {
	return m.description
}

func (m *mockRenderer) IsTerminalFormat() bool {
	return m.isTerminalFormat
}

func TestNewRendererRegistry(t *testing.T) {
	registry := NewRendererRegistry()
	
	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}
	
	if registry.renderers == nil {
		t.Fatal("Expected initialized renderers map")
	}
	
	if registry.aliases == nil {
		t.Fatal("Expected initialized aliases map")
	}
	
	// Check some standard aliases exist
	expectedAliases := []string{
		"color", "minimal", "no-color", "json", "yaml",
		"compact-json", "flat-json", "markdown", "html",
		"colorful", "plain", "text", "simple", "yml",
		"md", "slist",
	}
	
	for _, alias := range expectedAliases {
		if _, exists := registry.aliases[alias]; !exists {
			t.Errorf("Expected alias %q to exist", alias)
		}
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		renderer    Renderer
		wantErr     bool
		errContains string
	}{
		{
			name: "valid renderer",
			renderer: &mockRenderer{
				format:      "test-format",
				description: "Test renderer",
			},
			wantErr: false,
		},
		{
			name:        "nil renderer",
			renderer:    nil,
			wantErr:     true,
			errContains: "renderer cannot be nil",
		},
		{
			name: "empty format",
			renderer: &mockRenderer{
				format:      "",
				description: "Test renderer",
			},
			wantErr:     true,
			errContains: "renderer format cannot be empty",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRendererRegistry()
			err := registry.Register(tt.renderer)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Register() error = %v, want error containing %q", err, tt.errContains)
			}
			
			// If no error expected, verify renderer was registered
			if !tt.wantErr && tt.renderer != nil {
				if registered, exists := registry.renderers[tt.renderer.Format()]; !exists || registered != tt.renderer {
					t.Error("Renderer was not properly registered")
				}
			}
		})
	}
}

func TestGetRenderer(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register a test renderer
	testRenderer := &mockRenderer{
		format:      "test-format",
		description: "Test renderer",
	}
	if err := registry.Register(testRenderer); err != nil {
		t.Fatalf("Failed to register test renderer: %v", err)
	}
	
	tests := []struct {
		name    string
		format  OutputFormat
		wantErr bool
	}{
		{
			name:    "existing renderer",
			format:  "test-format",
			wantErr: false,
		},
		{
			name:    "non-existing renderer",
			format:  "unknown-format",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer, err := registry.GetRenderer(tt.format)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRenderer() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !tt.wantErr && renderer != testRenderer {
				t.Error("GetRenderer() returned wrong renderer")
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register test renderers
	if err := registry.Register(&mockRenderer{format: FormatJSON}); err != nil {
		t.Fatalf("Failed to register JSON format: %v", err)
	}
	if err := registry.Register(&mockRenderer{format: FormatYAML}); err != nil {
		t.Fatalf("Failed to register YAML format: %v", err)
	}
	if err := registry.Register(&mockRenderer{format: FormatColor}); err != nil {
		t.Fatalf("Failed to register Color format: %v", err)
	}
	
	tests := []struct {
		name       string
		formatStr  string
		want       OutputFormat
		wantErr    bool
	}{
		// Direct format matches
		{
			name:      "direct json format",
			formatStr: "json",
			want:      FormatJSON,
			wantErr:   false,
		},
		{
			name:      "direct yaml format",
			formatStr: "yaml",
			want:      FormatYAML,
			wantErr:   false,
		},
		// Alias matches
		{
			name:      "yml alias",
			formatStr: "yml",
			want:      FormatYAML,
			wantErr:   false,
		},
		{
			name:      "colorful alias",
			formatStr: "colorful",
			want:      FormatColor,
			wantErr:   false,
		},
		// Case insensitive
		{
			name:      "uppercase format",
			formatStr: "JSON",
			want:      FormatJSON,
			wantErr:   false,
		},
		{
			name:      "mixed case format",
			formatStr: "YaMl",
			want:      FormatYAML,
			wantErr:   false,
		},
		// Whitespace handling
		{
			name:      "format with spaces",
			formatStr: "  json  ",
			want:      FormatJSON,
			wantErr:   false,
		},
		// Unknown format
		{
			name:      "unknown format",
			formatStr: "unknown",
			want:      "",
			wantErr:   true,
		},
		// Empty format
		{
			name:      "empty format",
			formatStr: "",
			want:      "",
			wantErr:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registry.ParseFormat(tt.formatStr)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListFormats(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register test renderers
	if err := registry.Register(&mockRenderer{
		format:      "format1",
		description: "Description 1",
	}); err != nil {
		t.Fatalf("Failed to register format1: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:      "format2",
		description: "Description 2",
	}); err != nil {
		t.Fatalf("Failed to register format2: %v", err)
	}
	
	formats := registry.ListFormats()
	
	if len(formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(formats))
	}
	
	if desc, exists := formats["format1"]; !exists || desc != "Description 1" {
		t.Errorf("format1 missing or has wrong description: %q", desc)
	}
	
	if desc, exists := formats["format2"]; !exists || desc != "Description 2" {
		t.Errorf("format2 missing or has wrong description: %q", desc)
	}
}

func TestGetTerminalFormats(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register mixed renderers
	if err := registry.Register(&mockRenderer{
		format:           "terminal1",
		isTerminalFormat: true,
	}); err != nil {
		t.Fatalf("Failed to register terminal1: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "terminal2",
		isTerminalFormat: true,
	}); err != nil {
		t.Fatalf("Failed to register terminal2: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "data1",
		isTerminalFormat: false,
	}); err != nil {
		t.Fatalf("Failed to register data1: %v", err)
	}
	
	terminalFormats := registry.GetTerminalFormats()
	
	if len(terminalFormats) != 2 {
		t.Errorf("Expected 2 terminal formats, got %d", len(terminalFormats))
	}
	
	// Check that we got the right formats
	hasTerminal1 := false
	hasTerminal2 := false
	for _, f := range terminalFormats {
		if f == "terminal1" {
			hasTerminal1 = true
		}
		if f == "terminal2" {
			hasTerminal2 = true
		}
	}
	
	if !hasTerminal1 || !hasTerminal2 {
		t.Error("Did not get expected terminal formats")
	}
}

func TestGetDataFormats(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register mixed renderers
	if err := registry.Register(&mockRenderer{
		format:           "terminal1",
		isTerminalFormat: true,
	}); err != nil {
		t.Fatalf("Failed to register terminal1: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "data1",
		isTerminalFormat: false,
	}); err != nil {
		t.Fatalf("Failed to register data1: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "data2",
		isTerminalFormat: false,
	}); err != nil {
		t.Fatalf("Failed to register data2: %v", err)
	}
	
	dataFormats := registry.GetDataFormats()
	
	if len(dataFormats) != 2 {
		t.Errorf("Expected 2 data formats, got %d", len(dataFormats))
	}
	
	// Check that we got the right formats
	hasData1 := false
	hasData2 := false
	for _, f := range dataFormats {
		if f == "data1" {
			hasData1 = true
		}
		if f == "data2" {
			hasData2 = true
		}
	}
	
	if !hasData1 || !hasData2 {
		t.Error("Did not get expected data formats")
	}
}

func TestValidateFormat(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register a renderer
	if err := registry.Register(&mockRenderer{format: "valid-format"}); err != nil {
		t.Fatalf("Failed to register valid-format: %v", err)
	}
	
	tests := []struct {
		name    string
		format  OutputFormat
		wantErr bool
	}{
		{
			name:    "valid format",
			format:  "valid-format",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "invalid-format",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.ValidateFormat(tt.format)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err != nil && !strings.Contains(err.Error(), "not available") {
				t.Errorf("ValidateFormat() error message should mention format not available")
			}
		})
	}
}

func TestDefaultFormat(t *testing.T) {
	registry := NewRendererRegistry()
	
	defaultFormat := registry.DefaultFormat()
	
	if defaultFormat != FormatColor {
		t.Errorf("Expected default format to be %v, got %v", FormatColor, defaultFormat)
	}
}

func TestGetFormatHelp(t *testing.T) {
	registry := NewRendererRegistry()
	
	// Register test renderers
	if err := registry.Register(&mockRenderer{
		format:           "term1",
		description:      "Terminal format 1",
		isTerminalFormat: true,
	}); err != nil {
		t.Fatalf("Failed to register term1: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "data1",
		description:      "Data format 1",
		isTerminalFormat: false,
	}); err != nil {
		t.Fatalf("Failed to register data1: %v", err)
	}
	
	help := registry.GetFormatHelp()
	
	// Check that help contains expected sections
	if !strings.Contains(help, "Available output formats:") {
		t.Error("Help should contain header")
	}
	
	if !strings.Contains(help, "Terminal formats:") {
		t.Error("Help should contain terminal formats section")
	}
	
	if !strings.Contains(help, "Data formats:") {
		t.Error("Help should contain data formats section")
	}
	
	// Check format entries
	if !strings.Contains(help, "term1") || !strings.Contains(help, "Terminal format 1") {
		t.Error("Help should contain terminal format entry")
	}
	
	if !strings.Contains(help, "data1") || !strings.Contains(help, "Data format 1") {
		t.Error("Help should contain data format entry")
	}
}

func TestGetDefaultRegistry(t *testing.T) {
	// Test that we get the same instance
	registry1 := GetDefaultRegistry()
	registry2 := GetDefaultRegistry()
	
	if registry1 != registry2 {
		t.Error("GetDefaultRegistry should return the same instance")
	}
	
	// Test that it's properly initialized
	if registry1 == nil {
		t.Fatal("GetDefaultRegistry returned nil")
	}
	
	if registry1.renderers == nil {
		t.Error("Default registry should have initialized renderers map")
	}
	
	if registry1.aliases == nil {
		t.Error("Default registry should have initialized aliases map")
	}
}

func TestRegistryThreadSafety(t *testing.T) {
	// This test ensures the singleton pattern is thread-safe
	const goroutines = 100
	registries := make([]*RendererRegistry, goroutines)
	done := make(chan bool, goroutines)
	
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			registries[index] = GetDefaultRegistry()
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}
	
	// All should be the same instance
	first := registries[0]
	for i := 1; i < goroutines; i++ {
		if registries[i] != first {
			t.Errorf("Registry instance %d is different from first", i)
		}
	}
}