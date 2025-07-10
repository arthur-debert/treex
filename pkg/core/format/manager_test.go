package format

import (
	"errors"
	"strings"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
)

func TestNewRendererManager(t *testing.T) {
	manager := NewRendererManager()

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.registry == nil {
		t.Fatal("Expected manager to have a registry")
	}

	// Should use the default registry
	if manager.registry != GetDefaultRegistry() {
		t.Error("Manager should use the default registry")
	}
}

func TestNewRendererManagerWithRegistry(t *testing.T) {
	customRegistry := NewRendererRegistry()
	manager := NewRendererManagerWithRegistry(customRegistry)

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.registry != customRegistry {
		t.Error("Manager should use the provided custom registry")
	}
}

func TestRenderTree(t *testing.T) {
	// Create test tree
	testTree := &types.Node{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file1.txt",
				Path:  "/root/file1.txt",
				IsDir: false,
				Annotation: &types.Annotation{
					Path:  "/root/file1.txt",
					Notes: "A test file",
				},
			},
			{
				Name:  "dir1",
				Path:  "/root/dir1",
				IsDir: true,
				Children: []*types.Node{
					{
						Name:  "file2.txt",
						Path:  "/root/dir1/file2.txt",
						IsDir: false,
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		setupRegistry  func(*RendererRegistry)
		request        RenderRequest
		wantErr        bool
		errContains    string
		validateResult func(*testing.T, *RenderResponse)
	}{
		{
			name: "successful render with format",
			setupRegistry: func(r *RendererRegistry) {
				if err := r.Register(&mockRenderer{
					format:       "test-format",
					description:  "Test renderer",
					renderOutput: "rendered output",
				}); err != nil {
					t.Fatalf("Failed to register test-format: %v", err)
				}
			},
			request: RenderRequest{
				Tree:          testTree,
				Format:        "test-format",
				Verbose:       true,
				ShowStats:     true,
				SafeMode:      true,
				TerminalWidth: 100,
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *RenderResponse) {
				if resp.Output != "rendered output" {
					t.Errorf("Expected output 'rendered output', got %q", resp.Output)
				}
				if resp.Format != "test-format" {
					t.Errorf("Expected format 'test-format', got %v", resp.Format)
				}
				if resp.RendererUsed != "Test renderer" {
					t.Errorf("Expected renderer description 'Test renderer', got %q", resp.RendererUsed)
				}
				if resp.Stats == nil {
					t.Error("Expected non-nil stats")
				} else {
					if resp.Stats.NodesRendered != 4 { // root + 3 children
						t.Errorf("Expected 4 nodes rendered, got %d", resp.Stats.NodesRendered)
					}
					if resp.Stats.AnnotationsFound != 1 {
						t.Errorf("Expected 1 annotation found, got %d", resp.Stats.AnnotationsFound)
					}
				}
			},
		},
		{
			name: "default format when none specified",
			setupRegistry: func(r *RendererRegistry) {
				if err := r.Register(&mockRenderer{
					format:       FormatColor, // Default format
					description:  "Color renderer",
					renderOutput: "colored output",
				}); err != nil {
					t.Fatalf("Failed to register color format: %v", err)
				}
				// Also register no-color renderer since auto-detection might pick it
				if err := r.Register(&mockRenderer{
					format:       FormatNoColor,
					description:  "No-color renderer",
					renderOutput: "plain output",
				}); err != nil {
					t.Fatalf("Failed to register no-color format: %v", err)
				}
			},
			request: RenderRequest{
				Tree: testTree,
				// Format not specified, should use default
				// Explicitly set IsTTY to true to ensure color format is selected
				IsTTY: func() *bool { b := true; return &b }(),
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *RenderResponse) {
				if resp.Format != FormatColor {
					t.Errorf("Expected default format %v, got %v", FormatColor, resp.Format)
				}
			},
		},
		{
			name: "auto-detect non-TTY uses no-color format",
			setupRegistry: func(r *RendererRegistry) {
				if err := r.Register(&mockRenderer{
					format:       FormatColor,
					description:  "Color renderer",
					renderOutput: "colored output",
				}); err != nil {
					t.Fatalf("Failed to register color format: %v", err)
				}
				if err := r.Register(&mockRenderer{
					format:       FormatNoColor,
					description:  "No-color renderer",
					renderOutput: "plain output",
				}); err != nil {
					t.Fatalf("Failed to register no-color format: %v", err)
				}
			},
			request: RenderRequest{
				Tree: testTree,
				// Explicitly set IsTTY to false to simulate piped output
				IsTTY: func() *bool { b := false; return &b }(),
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *RenderResponse) {
				if resp.Format != FormatNoColor {
					t.Errorf("Expected no-color format for non-TTY, got %v", resp.Format)
				}
			},
		},
		{
			name: "invalid format error",
			setupRegistry: func(r *RendererRegistry) {
				// Don't register any renderers
			},
			request: RenderRequest{
				Tree:   testTree,
				Format: "nonexistent-format",
			},
			wantErr:     true,
			errContains: "invalid format",
		},
		{
			name: "renderer not found error",
			setupRegistry: func(r *RendererRegistry) {
				// Register nothing, default format won't have a renderer
			},
			request: RenderRequest{
				Tree: testTree,
				// Will use default format but no renderer registered
			},
			wantErr:     true,
			errContains: "failed to get renderer",
		},
		{
			name: "render error propagation",
			setupRegistry: func(r *RendererRegistry) {
				if err := r.Register(&mockRenderer{
					format:      "error-format",
					renderError: errors.New("render failed"),
				}); err != nil {
					t.Fatalf("Failed to register error-format: %v", err)
				}
			},
			request: RenderRequest{
				Tree:   testTree,
				Format: "error-format",
			},
			wantErr:     true,
			errContains: "failed to render tree",
		},
		{
			name: "nil tree handling",
			setupRegistry: func(r *RendererRegistry) {
				if err := r.Register(&mockRenderer{
					format:       "test-format",
					renderOutput: "empty",
				}); err != nil {
					t.Fatalf("Failed to register test-format: %v", err)
				}
			},
			request: RenderRequest{
				Tree:   nil,
				Format: "test-format",
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *RenderResponse) {
				if resp.Stats.NodesRendered != 0 {
					t.Errorf("Expected 0 nodes for nil tree, got %d", resp.Stats.NodesRendered)
				}
				if resp.Stats.AnnotationsFound != 0 {
					t.Errorf("Expected 0 annotations for nil tree, got %d", resp.Stats.AnnotationsFound)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRendererRegistry()
			if tt.setupRegistry != nil {
				tt.setupRegistry(registry)
			}

			manager := NewRendererManagerWithRegistry(registry)
			resp, err := manager.RenderTree(tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTree() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("RenderTree() error = %v, want error containing %q", err, tt.errContains)
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, resp)
			}
		})
	}
}

func TestCountNodes(t *testing.T) {
	manager := NewRendererManager()

	tests := []struct {
		name     string
		node     *types.Node
		expected int
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: 0,
		},
		{
			name:     "single node",
			node:     &types.Node{Name: "single"},
			expected: 1,
		},
		{
			name: "node with children",
			node: &types.Node{
				Name: "root",
				Children: []*types.Node{
					{Name: "child1"},
					{Name: "child2"},
					{
						Name: "child3",
						Children: []*types.Node{
							{Name: "grandchild1"},
							{Name: "grandchild2"},
						},
					},
				},
			},
			expected: 6, // root + 3 children + 2 grandchildren
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := manager.countNodes(tt.node)
			if count != tt.expected {
				t.Errorf("countNodes() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestCountAnnotations(t *testing.T) {
	manager := NewRendererManager()

	tests := []struct {
		name     string
		node     *types.Node
		expected int
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: 0,
		},
		{
			name:     "node without annotation",
			node:     &types.Node{Name: "no-annotation"},
			expected: 0,
		},
		{
			name: "node with annotation",
			node: &types.Node{
				Name:       "annotated",
				Annotation: &types.Annotation{Path: "annotated", Notes: "Test"},
			},
			expected: 1,
		},
		{
			name: "mixed tree",
			node: &types.Node{
				Name: "root",
				Children: []*types.Node{
					{
						Name:       "annotated1",
						Annotation: &types.Annotation{Path: "annotated1", Notes: "Test1"},
					},
					{Name: "not-annotated"},
					{
						Name: "dir",
						Children: []*types.Node{
							{
								Name:       "annotated2",
								Annotation: &types.Annotation{Path: "annotated2", Notes: "Test2"},
							},
							{Name: "not-annotated2"},
						},
					},
				},
			},
			expected: 2, // annotated1 + annotated2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := manager.countAnnotations(tt.node)
			if count != tt.expected {
				t.Errorf("countAnnotations() = %d, want %d", count, tt.expected)
			}
		})
	}
}

func TestManagerHelperMethods(t *testing.T) {
	registry := NewRendererRegistry()

	// Register test renderers
	if err := registry.Register(&mockRenderer{
		format:           "terminal-test",
		description:      "Terminal test",
		isTerminalFormat: true,
	}); err != nil {
		t.Fatalf("Failed to register terminal-test: %v", err)
	}
	if err := registry.Register(&mockRenderer{
		format:           "data-test",
		description:      "Data test",
		isTerminalFormat: false,
	}); err != nil {
		t.Fatalf("Failed to register data-test: %v", err)
	}

	manager := NewRendererManagerWithRegistry(registry)

	t.Run("GetAvailableFormats", func(t *testing.T) {
		formats := manager.GetAvailableFormats()

		if len(formats) != 2 {
			t.Errorf("Expected 2 formats, got %d", len(formats))
		}

		if desc, exists := formats["terminal-test"]; !exists || desc != "Terminal test" {
			t.Error("Missing or incorrect terminal-test format")
		}

		if desc, exists := formats["data-test"]; !exists || desc != "Data test" {
			t.Error("Missing or incorrect data-test format")
		}
	})

	t.Run("GetFormatHelp", func(t *testing.T) {
		help := manager.GetFormatHelp()

		if !strings.Contains(help, "Terminal test") {
			t.Error("Help should contain terminal format")
		}

		if !strings.Contains(help, "Data test") {
			t.Error("Help should contain data format")
		}
	})

	t.Run("ValidateFormat", func(t *testing.T) {
		if err := manager.ValidateFormat("terminal-test"); err != nil {
			t.Errorf("ValidateFormat() unexpected error for valid format: %v", err)
		}

		if err := manager.ValidateFormat("invalid"); err == nil {
			t.Error("ValidateFormat() expected error for invalid format")
		}
	})

	t.Run("ParseFormat", func(t *testing.T) {
		format, err := manager.ParseFormat("terminal-test")
		if err != nil {
			t.Errorf("ParseFormat() unexpected error: %v", err)
		}
		if format != "terminal-test" {
			t.Errorf("ParseFormat() = %v, want %v", format, "terminal-test")
		}

		_, err = manager.ParseFormat("invalid")
		if err == nil {
			t.Error("ParseFormat() expected error for invalid format")
		}
	})

	t.Run("GetTerminalFormats", func(t *testing.T) {
		formats := manager.GetTerminalFormats()

		if len(formats) != 1 {
			t.Errorf("Expected 1 terminal format, got %d", len(formats))
		}

		if formats[0] != "terminal-test" {
			t.Errorf("Expected terminal-test format, got %v", formats[0])
		}
	})

	t.Run("GetDataFormats", func(t *testing.T) {
		formats := manager.GetDataFormats()

		if len(formats) != 1 {
			t.Errorf("Expected 1 data format, got %d", len(formats))
		}

		if formats[0] != "data-test" {
			t.Errorf("Expected data-test format, got %v", formats[0])
		}
	})

	t.Run("IsTerminalFormat", func(t *testing.T) {
		if !manager.IsTerminalFormat("terminal-test") {
			t.Error("Expected terminal-test to be a terminal format")
		}

		if manager.IsTerminalFormat("data-test") {
			t.Error("Expected data-test NOT to be a terminal format")
		}

		if manager.IsTerminalFormat("invalid") {
			t.Error("Expected invalid format to return false")
		}
	})
}

func TestGetDefaultManager(t *testing.T) {
	// Test singleton behavior
	manager1 := GetDefaultManager()
	manager2 := GetDefaultManager()

	if manager1 != manager2 {
		t.Error("GetDefaultManager should return the same instance")
	}

	if manager1 == nil {
		t.Fatal("GetDefaultManager returned nil")
	}

	if manager1.registry == nil {
		t.Error("Default manager should have a registry")
	}
}

func TestRenderTreeWithDefaults(t *testing.T) {
	// Register a test renderer in the default registry
	registry := GetDefaultRegistry()
	if err := registry.Register(&mockRenderer{
		format:       "test-defaults",
		renderOutput: "default output",
	}); err != nil {
		t.Fatalf("Failed to register test-defaults: %v", err)
	}

	testTree := &types.Node{
		Name:  "root",
		IsDir: true,
	}

	output, err := RenderTreeWithDefaults(testTree, "test-defaults", true, true)
	if err != nil {
		t.Errorf("RenderTreeWithDefaults() unexpected error: %v", err)
	}

	if output != "default output" {
		t.Errorf("RenderTreeWithDefaults() = %q, want %q", output, "default output")
	}

	// Test error case
	_, err = RenderTreeWithDefaults(testTree, "nonexistent", false, false)
	if err == nil {
		t.Error("RenderTreeWithDefaults() expected error for nonexistent format")
	}
}

func TestSelectFormat(t *testing.T) {
	registry := NewRendererRegistry()

	// Register test renderers
	if err := registry.Register(&mockRenderer{format: "valid-format"}); err != nil {
		t.Fatalf("Failed to register valid-format: %v", err)
	}
	if err := registry.Register(&mockRenderer{format: FormatColor}); err != nil { // Default format
		t.Fatalf("Failed to register color format: %v", err)
	}
	if err := registry.Register(&mockRenderer{format: FormatNoColor}); err != nil {
		t.Fatalf("Failed to register no-color format: %v", err)
	}

	manager := NewRendererManagerWithRegistry(registry)

	tests := []struct {
		name        string
		request     RenderRequest
		wantFormat  OutputFormat
		wantErr     bool
		errContains string
	}{
		{
			name: "explicit format",
			request: RenderRequest{
				Format: "valid-format",
			},
			wantFormat: "valid-format",
			wantErr:    false,
		},
		{
			name: "invalid explicit format",
			request: RenderRequest{
				Format: "invalid-format",
			},
			wantErr:     true,
			errContains: "invalid format",
		},
		{
			name: "default format when empty",
			request: RenderRequest{
				// Explicitly set IsTTY to true for consistent testing
				IsTTY: func() *bool { b := true; return &b }(),
			},
			wantFormat: FormatColor,
			wantErr:    false,
		},
		{
			name: "non-TTY uses no-color format",
			request: RenderRequest{
				// Explicitly set IsTTY to false
				IsTTY: func() *bool { b := false; return &b }(),
			},
			wantFormat: FormatNoColor,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := manager.selectFormat(tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("selectFormat() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("selectFormat() error = %v, want error containing %q", err, tt.errContains)
			}

			if !tt.wantErr && format != tt.wantFormat {
				t.Errorf("selectFormat() = %v, want %v", format, tt.wantFormat)
			}
		})
	}
}
