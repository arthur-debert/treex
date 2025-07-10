package format

import (
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
)

func TestRender(t *testing.T) {
	// Setup registry with test renderer
	registry := GetDefaultRegistry()
	testRenderer := &mockRenderer{
		format:       "test-render",
		renderOutput: "test output",
		description:  "Test renderer",
	}
	if err := registry.Register(testRenderer); err != nil {
		t.Fatalf("Failed to register test renderer: %v", err)
	}

	// Also register default format renderer
	defaultRenderer := &mockRenderer{
		format:       registry.DefaultFormat(),
		renderOutput: "default output",
		description:  "Default renderer",
	}
	if err := registry.Register(defaultRenderer); err != nil {
		t.Fatalf("Failed to register default renderer: %v", err)
	}

	testTree := &types.Node{
		Name:  "test",
		IsDir: false,
	}

	tests := []struct {
		name        string
		options     RenderOptions
		wantOutput  string
		wantErr     bool
		errContains string
	}{
		{
			name: "render with specified format",
			options: RenderOptions{
				Format: "test-render",
			},
			wantOutput: "test output",
			wantErr:    false,
		},
		{
			name:       "render with default format",
			options:    RenderOptions{}, // Empty format
			wantOutput: "default output",
			wantErr:    false,
		},
		{
			name: "render with non-existent format",
			options: RenderOptions{
				Format: "non-existent",
			},
			wantErr:     true,
			errContains: "no renderer registered",
		},
		{
			name: "render with all options",
			options: RenderOptions{
				Format:        "test-render",
				Verbose:       true,
				ShowStats:     true,
				IgnoreFile:    ".gitignore",
				MaxDepth:      5,
				SafeMode:      true,
				TerminalWidth: 120,
			},
			wantOutput: "test output",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Render(testTree, tt.options)

			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Render() error = %v, want error containing %q", err, tt.errContains)
			}

			if !tt.wantErr && output != tt.wantOutput {
				t.Errorf("Render() = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}

func TestParseFormatString(t *testing.T) {
	// This uses the default registry's ParseFormat method
	registry := GetDefaultRegistry()

	// Register some test formats
	if err := registry.Register(&mockRenderer{format: "test-format"}); err != nil {
		t.Fatalf("Failed to register test-format: %v", err)
	}
	if err := registry.Register(&mockRenderer{format: FormatColor}); err != nil {
		t.Fatalf("Failed to register Color format: %v", err)
	}

	tests := []struct {
		name      string
		formatStr string
		want      OutputFormat
		wantErr   bool
	}{
		{
			name:      "valid format",
			formatStr: "test-format",
			want:      "test-format",
			wantErr:   false,
		},
		{
			name:      "valid Color format",
			formatStr: "color",
			want:      FormatColor,
			wantErr:   false,
		},
		{
			name:      "invalid format",
			formatStr: "invalid",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "empty format",
			formatStr: "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "format with whitespace",
			formatStr: "  test-format  ",
			want:      "test-format",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormatString(tt.formatStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormatString() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("ParseFormatString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListAvailableFormats(t *testing.T) {
	// Clear and setup registry for predictable test
	registry := GetDefaultRegistry()

	// Clear existing renderers for clean test
	registry.renderers = make(map[OutputFormat]Renderer)

	// Register test formats
	if err := registry.Register(&mockRenderer{
		format:      "format-a",
		description: "Format A Description",
	}); err != nil {
		t.Fatalf("Failed to register format-a: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:      "format-b",
		description: "Format B Description",
	}); err != nil {
		t.Fatalf("Failed to register format-b: %v", err)
	}

	formats := ListAvailableFormats()

	if len(formats) != 2 {
		t.Errorf("Expected 2 formats, got %d", len(formats))
	}

	if desc, exists := formats["format-a"]; !exists {
		t.Error("format-a not found in available formats")
	} else if desc != "Format A Description" {
		t.Errorf("format-a description = %q, want %q", desc, "Format A Description")
	}

	if desc, exists := formats["format-b"]; !exists {
		t.Error("format-b not found in available formats")
	} else if desc != "Format B Description" {
		t.Errorf("format-b description = %q, want %q", desc, "Format B Description")
	}
}

func TestGetFormatHelp_Default(t *testing.T) {
	// Clear and setup registry for predictable test
	registry := GetDefaultRegistry()

	// Clear existing renderers for clean test
	registry.renderers = make(map[OutputFormat]Renderer)

	// Register mixed format types
	if err := registry.Register(&mockRenderer{
		format:           "terminal-fmt",
		description:      "A terminal format",
		isTerminalFormat: true,
	}); err != nil {
		t.Fatalf("Failed to register terminal-fmt: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "data-fmt",
		description:      "A data format",
		isTerminalFormat: false,
	}); err != nil {
		t.Fatalf("Failed to register data-fmt: %v", err)
	}

	help := GetFormatHelp()

	// Verify help structure
	if !strings.Contains(help, "Available output formats:") {
		t.Error("Help should contain header")
	}

	if !strings.Contains(help, "Terminal formats:") {
		t.Error("Help should contain terminal formats section")
	}

	if !strings.Contains(help, "Data formats:") {
		t.Error("Help should contain data formats section")
	}

	// Verify format entries
	if !strings.Contains(help, "terminal-fmt") {
		t.Error("Help should list terminal-fmt")
	}

	if !strings.Contains(help, "A terminal format") {
		t.Error("Help should contain terminal format description")
	}

	if !strings.Contains(help, "data-fmt") {
		t.Error("Help should list data-fmt")
	}

	if !strings.Contains(help, "A data format") {
		t.Error("Help should contain data format description")
	}
}

func TestDefaultFunctionsIntegration(t *testing.T) {
	// This test verifies that all the default.go functions work together properly

	// Clear registry for clean test
	registry := GetDefaultRegistry()
	registry.renderers = make(map[OutputFormat]Renderer)

	// Register a test renderer
	testRenderer := &mockRenderer{
		format:       "integration-test",
		description:  "Integration test format",
		renderOutput: "integration output",
	}
	if err := registry.Register(testRenderer); err != nil {
		t.Fatalf("Failed to register integration-test: %v", err)
	}

	// Test ParseFormatString
	format, err := ParseFormatString("integration-test")
	if err != nil {
		t.Errorf("ParseFormatString() unexpected error: %v", err)
	}
	if format != "integration-test" {
		t.Errorf("ParseFormatString() = %v, want %v", format, "integration-test")
	}

	// Test ListAvailableFormats
	formats := ListAvailableFormats()
	if _, exists := formats["integration-test"]; !exists {
		t.Error("integration-test not found in available formats")
	}

	// Test GetFormatHelp
	help := GetFormatHelp()
	if !strings.Contains(help, "integration-test") {
		t.Error("Help should contain integration-test format")
	}

	// Test Render
	testTree := &types.Node{Name: "test"}
	output, err := Render(testTree, RenderOptions{Format: format})
	if err != nil {
		t.Errorf("Render() unexpected error: %v", err)
	}
	if output != "integration output" {
		t.Errorf("Render() = %q, want %q", output, "integration output")
	}
}
