package plugins

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginOptionService(t *testing.T) {
	// Create a test registry with known plugins
	registry := NewRegistry()

	// Register a mock plugin for testing
	mockPlugin := &MockFilterPlugin{
		name: "test",
		categories: []FilterPluginCategory{
			{Name: "category1", Description: "Test category 1"},
			{Name: "category2", Description: "Test category 2"},
		},
	}
	err := registry.Register(mockPlugin)
	require.NoError(t, err)

	service := NewPluginOptionService(registry)

	t.Run("GetAvailableOptions", func(t *testing.T) {
		options := service.GetAvailableOptions()

		// Should have options for our mock plugin
		assert.Len(t, options, 2)

		// Check the options
		expectedOptions := []PluginOptionDefinition{
			{PluginName: "test", CategoryName: "category1", Description: "Test category 1"},
			{PluginName: "test", CategoryName: "category2", Description: "Test category 2"},
		}

		for _, expected := range expectedOptions {
			found := false
			for _, actual := range options {
				if actual.PluginName == expected.PluginName &&
					actual.CategoryName == expected.CategoryName &&
					actual.Description == expected.Description {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected option not found: %+v", expected)
		}
	})

	t.Run("ValidatePluginFilters valid", func(t *testing.T) {
		filters := map[string]map[string]bool{
			"test": {
				"category1": true,
				"category2": false,
			},
		}

		err := service.ValidatePluginFilters(filters)
		assert.NoError(t, err)
	})

	t.Run("ValidatePluginFilters invalid plugin", func(t *testing.T) {
		filters := map[string]map[string]bool{
			"nonexistent": {
				"category": true,
			},
		}

		err := service.ValidatePluginFilters(filters)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown plugin: nonexistent")
	})

	t.Run("ValidatePluginFilters invalid category", func(t *testing.T) {
		filters := map[string]map[string]bool{
			"test": {
				"invalid_category": true,
			},
		}

		err := service.ValidatePluginFilters(filters)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown category invalid_category for plugin test")
	})

	t.Run("ValidatePluginFilters disabled filters are not validated", func(t *testing.T) {
		// Disabled filters (false) should not be validated
		filters := map[string]map[string]bool{
			"test": {
				"invalid_category": false, // This should not cause an error
			},
		}

		err := service.ValidatePluginFilters(filters)
		assert.NoError(t, err)
	})

	t.Run("GetPluginNames", func(t *testing.T) {
		names := service.GetPluginNames()
		assert.Contains(t, names, "test")
	})

	t.Run("GetPluginCategories valid plugin", func(t *testing.T) {
		categories, err := service.GetPluginCategories("test")
		require.NoError(t, err)
		assert.Len(t, categories, 2)
		assert.Equal(t, "category1", categories[0].Name)
		assert.Equal(t, "category2", categories[1].Name)
	})

	t.Run("GetPluginCategories invalid plugin", func(t *testing.T) {
		_, err := service.GetPluginCategories("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugin not found: nonexistent")
	})
}

func TestDefaultPluginOptionService(t *testing.T) {
	// Test that the default service works with the default registry
	service := GetDefaultPluginOptionService()

	// The default service should be functional even if no plugins are registered
	// (plugins are registered via init() functions when their packages are imported)
	options := service.GetAvailableOptions()
	pluginNames := service.GetPluginNames()

	// Basic functionality should work regardless of what's registered
	assert.NotNil(t, options)
	assert.NotNil(t, pluginNames)

	// Test validation with empty filters should not error
	err := service.ValidatePluginFilters(map[string]map[string]bool{})
	assert.NoError(t, err)

	// Test that service methods don't panic with no registered plugins
	_, err = service.GetPluginCategories("nonexistent")
	assert.Error(t, err) // Should error for nonexistent plugin
}

// MockFilterPlugin is a test implementation of FilterPlugin
type MockFilterPlugin struct {
	name       string
	categories []FilterPluginCategory
}

func (m *MockFilterPlugin) Name() string {
	return m.name
}

func (m *MockFilterPlugin) FindRoots(fs afero.Fs, searchRoot string) ([]string, error) {
	return []string{"."}, nil
}

func (m *MockFilterPlugin) ProcessRoot(fs afero.Fs, rootPath string) (*Result, error) {
	return &Result{
		PluginName: m.name,
		RootPath:   rootPath,
		Categories: make(map[string][]string),
		Metadata:   make(map[string]interface{}),
		Cache:      make(map[string]interface{}),
	}, nil
}

func (m *MockFilterPlugin) GetCategories() []FilterPluginCategory {
	return m.categories
}
