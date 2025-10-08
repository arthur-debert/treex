// see docs/dev/architecture.txt - Phase 4: Plugin Filtering
package plugins_test

import (
	"testing"

	"github.com/spf13/afero"
	"treex/treex/internal/testutil"
	"treex/treex/plugins"
	"treex/treex/plugins/dummy"
)

func TestRegistry(t *testing.T) {
	registry := plugins.NewRegistry()

	// Test empty registry
	if len(registry.ListPlugins()) != 0 {
		t.Error("New registry should be empty")
	}

	// Test registering a plugin
	dummyPlugin := dummy.NewDummyPlugin()
	err := registry.Register(dummyPlugin)
	if err != nil {
		t.Fatalf("Failed to register dummy plugin: %v", err)
	}

	// Test plugin is registered
	plugins := registry.ListPlugins()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0] != "dummy" {
		t.Errorf("Expected plugin name 'dummy', got %q", plugins[0])
	}

	// Test retrieving plugin
	retrieved := registry.GetPlugin("dummy")
	if retrieved == nil {
		t.Error("Failed to retrieve registered plugin")
	}
	if retrieved.Name() != "dummy" {
		t.Errorf("Retrieved plugin has wrong name: %q", retrieved.Name())
	}

	// Test retrieving non-existent plugin
	nonExistent := registry.GetPlugin("nonexistent")
	if nonExistent != nil {
		t.Error("Should return nil for non-existent plugin")
	}

	// Test duplicate registration
	err = registry.Register(dummyPlugin)
	if err == nil {
		t.Error("Should not allow duplicate plugin registration")
	}
}

func TestRegistryWithEmptyName(t *testing.T) {
	registry := plugins.NewRegistry()

	// Create a mock plugin with empty name
	mockPlugin := &MockPlugin{name: ""}

	err := registry.Register(mockPlugin)
	if err == nil {
		t.Error("Should not allow plugin with empty name")
	}
}

func TestDummyPluginInterface(t *testing.T) {
	plugin := dummy.NewDummyPlugin()

	// Test plugin name
	if plugin.Name() != "dummy" {
		t.Errorf("Expected plugin name 'dummy', got %q", plugin.Name())
	}

	// Test with empty filesystem
	fs := testutil.NewTestFS()
	roots, err := plugin.FindRoots(fs, "/empty")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}
	if len(roots) != 0 {
		t.Errorf("Expected no roots in empty filesystem, got %d", len(roots))
	}
}

func TestDummyPluginFindRoots(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	// Create filesystem with multiple .dummy markers
	fs.MustCreateTree("/test", map[string]interface{}{
		"project1": map[string]interface{}{
			".dummy":   "marker",
			"main.go":  "package main",
			"test.txt": "test content",
		},
		"project2": map[string]interface{}{
			".dummy":    "marker",
			"readme.md": "# Project 2",
		},
		"regular": map[string]interface{}{
			"file.txt": "no marker here",
		},
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"project3": map[string]interface{}{
					".dummy":  "marker",
					"code.py": "print('hello')",
				},
			},
		},
	})

	roots, err := plugin.FindRoots(fs, "/test")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}

	// Should find 3 roots
	if len(roots) != 3 {
		t.Errorf("Expected 3 roots, got %d", len(roots))
	}

	// Verify the found roots
	expectedRoots := map[string]bool{
		"project1":             true,
		"project2":             true,
		"nested/deep/project3": true,
	}

	for _, root := range roots {
		if !expectedRoots[root] {
			t.Errorf("Unexpected root found: %q", root)
		}
		delete(expectedRoots, root)
	}

	if len(expectedRoots) > 0 {
		t.Errorf("Some expected roots were not found: %v", expectedRoots)
	}
}

func TestDummyPluginProcessRoot(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := dummy.NewDummyPlugin()

	// Create a test project
	fs.MustCreateTree("/test/project", map[string]interface{}{
		".dummy":     "marker",
		"main.go":    "package main",
		"utils.go":   "package utils",
		"README.md":  "# Project",
		"config.yml": "key: value",
		"data.json":  "{}",
		"script":     "#!/bin/bash", // no extension
		"subdir": map[string]interface{}{
			"helper.py": "print('hello')",
			"test.js":   "console.log('test')",
		},
	})

	result, err := plugin.ProcessRoot(fs, "/test/project")
	if err != nil {
		t.Fatalf("ProcessRoot failed: %v", err)
	}

	// Verify result structure
	if result.PluginName != "dummy" {
		t.Errorf("Expected plugin name 'dummy', got %q", result.PluginName)
	}
	if result.RootPath != "/test/project" {
		t.Errorf("Expected root path '/test/project', got %q", result.RootPath)
	}

	// Check metadata
	totalFiles, ok := result.Metadata["total_files"].(int)
	if !ok || totalFiles != 9 { // Including .dummy file and all others
		t.Errorf("Expected 9 total files, got %v", totalFiles)
	}

	// Verify categories - now including .dummy file
	expectedCategories := map[string]int{
		"go":           2, // main.go, utils.go
		"md":           1, // README.md
		"yml":          1, // config.yml
		"json":         1, // data.json
		"no-extension": 1, // script
		"py":           1, // helper.py
		"js":           1, // test.js
		"dummy":        1, // .dummy
	}

	for category, expectedCount := range expectedCategories {
		files, exists := result.Categories[category]
		if !exists {
			t.Errorf("Expected category %q not found", category)
			continue
		}
		if len(files) != expectedCount {
			t.Errorf("Category %q: expected %d files, got %d", category, expectedCount, len(files))
		}
	}

	// Verify some specific files are in correct categories
	goFiles := result.Categories["go"]
	foundMainGo := false
	foundUtilsGo := false
	for _, file := range goFiles {
		if file == "main.go" {
			foundMainGo = true
		}
		if file == "utils.go" {
			foundUtilsGo = true
		}
	}
	if !foundMainGo {
		t.Error("main.go not found in 'go' category")
	}
	if !foundUtilsGo {
		t.Error("utils.go not found in 'go' category")
	}
}

func TestEngine(t *testing.T) {
	fs := testutil.NewTestFS()
	registry := plugins.NewRegistry()

	// Register dummy plugin
	dummyPlugin := dummy.NewDummyPlugin()
	err := registry.Register(dummyPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	engine := plugins.NewEngine(registry, fs)

	// Test with empty filesystem
	opts := plugins.ProcessOptions{
		SearchRoot: "/empty",
	}

	results, err := engine.Process(opts)
	if err != nil {
		t.Fatalf("Engine process failed: %v", err)
	}

	if len(results.Results) != 0 {
		t.Error("Expected no results from empty filesystem")
	}
	if len(results.Errors) != 0 {
		t.Errorf("Expected no errors, got: %v", results.Errors)
	}
}

func TestEngineWithContent(t *testing.T) {
	fs := testutil.NewTestFS()
	registry := plugins.NewRegistry()

	// Register dummy plugin
	dummyPlugin := dummy.NewDummyPlugin()
	err := registry.Register(dummyPlugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Create test content
	fs.MustCreateTree("/workspace", map[string]interface{}{
		"project1": map[string]interface{}{
			".dummy":  "marker",
			"main.go": "package main",
		},
		"project2": map[string]interface{}{
			".dummy":    "marker",
			"script.py": "print('hello')",
		},
	})

	engine := plugins.NewEngine(registry, fs)
	opts := plugins.ProcessOptions{
		SearchRoot: "/workspace",
	}

	results, err := engine.Process(opts)
	if err != nil {
		t.Fatalf("Engine process failed: %v", err)
	}

	// Should have results from dummy plugin
	dummyResults, exists := results.Results["dummy"]
	if !exists {
		t.Fatal("Expected results from dummy plugin")
	}

	if len(dummyResults) != 2 {
		t.Errorf("Expected 2 results from dummy plugin, got %d", len(dummyResults))
	}

	// Verify both projects were processed
	processedRoots := make(map[string]bool)
	for _, result := range dummyResults {
		processedRoots[result.RootPath] = true
	}

	expectedRoots := []string{"project1", "project2"}
	for _, expectedRoot := range expectedRoots {
		found := false
		for actualRoot := range processedRoots {
			if actualRoot == expectedRoot || actualRoot == "/workspace/"+expectedRoot {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected root %q was not processed", expectedRoot)
		}
	}
}

func TestEngineWithSpecificPlugins(t *testing.T) {
	fs := testutil.NewTestFS()
	registry := plugins.NewRegistry()

	// Register multiple plugins
	dummyPlugin := dummy.NewDummyPlugin()
	mockPlugin := &MockPlugin{name: "mock"}

	if err := registry.Register(dummyPlugin); err != nil {
		t.Fatalf("Failed to register dummy plugin: %v", err)
	}
	if err := registry.Register(mockPlugin); err != nil {
		t.Fatalf("Failed to register mock plugin: %v", err)
	}

	engine := plugins.NewEngine(registry, fs)

	// Test running only specific plugin
	opts := plugins.ProcessOptions{
		SearchRoot:     "/test",
		EnabledPlugins: []string{"dummy"}, // Only run dummy plugin
	}

	results, err := engine.Process(opts)
	if err != nil {
		t.Fatalf("Engine process failed: %v", err)
	}

	// Should only have dummy results, not mock
	if len(results.Results) > 1 {
		t.Error("Expected only dummy plugin to run")
	}
	if _, exists := results.Results["mock"]; exists {
		t.Error("Mock plugin should not have run")
	}
}

// MockPlugin is a simple test plugin for testing
type MockPlugin struct {
	name string
}

func (m *MockPlugin) Name() string {
	return m.name
}

func (m *MockPlugin) FindRoots(fs afero.Fs, searchRoot string) ([]string, error) {
	return []string{}, nil
}

func (m *MockPlugin) ProcessRoot(fs afero.Fs, rootPath string) (*plugins.Result, error) {
	return &plugins.Result{
		PluginName: m.name,
		RootPath:   rootPath,
		Categories: make(map[string][]string),
		Metadata:   make(map[string]interface{}),
	}, nil
}

func TestDefaultRegistry(t *testing.T) {
	// Test that default registry has plugins (dummy should be auto-registered)
	pluginNames := plugins.DefaultRegistry.ListPlugins()

	found := false
	for _, name := range pluginNames {
		if name == "dummy" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Dummy plugin should be registered in default registry")
	}
}

func TestNewDefaultEngine(t *testing.T) {
	fs := testutil.NewTestFS()
	engine := plugins.NewDefaultEngine(fs)

	if engine == nil {
		t.Error("NewDefaultEngine should not return nil")
	}

	// Should be able to process with default engine
	opts := plugins.ProcessOptions{
		SearchRoot: "/test",
	}

	_, err := engine.Process(opts)
	if err != nil {
		t.Errorf("Default engine should be able to process: %v", err)
	}
}
