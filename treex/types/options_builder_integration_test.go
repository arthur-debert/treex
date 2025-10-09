package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsBuilderIntegration(t *testing.T) {
	t.Run("basic options building", func(t *testing.T) {
		options := NewOptionsBuilder().
			WithRoot("/test").
			WithMaxDepth(5).
			WithHidden().
			WithDirsOnly().
			WithExcludes("*.tmp", "node_modules").
			WithBuiltinIgnores().
			Build()

		assert.Equal(t, "/test", options.Root)
		assert.Equal(t, 5, options.Tree.MaxDepth)
		assert.True(t, options.Tree.ShowHidden)
		assert.True(t, options.Tree.DirsOnly)
		assert.Equal(t, []string{"*.tmp", "node_modules"}, options.Patterns.Excludes)
		assert.True(t, options.Patterns.UseBuiltinIgnores)
	})

	t.Run("plugin filter building", func(t *testing.T) {
		options := NewOptionsBuilder().
			WithRoot(".").
			WithPluginFilter("git", "staged").
			WithPluginFilter("git", "unstaged").
			WithPluginFilter("info", "annotated").
			Build()

		require.NotNil(t, options.Plugins.Filters)
		assert.True(t, options.Plugins.Filters["git"]["staged"])
		assert.True(t, options.Plugins.Filters["git"]["unstaged"])
		assert.True(t, options.Plugins.Filters["info"]["annotated"])
	})

	t.Run("plugin filter bulk setting", func(t *testing.T) {
		filters := map[string]map[string]bool{
			"git": {
				"staged":   true,
				"unstaged": false,
			},
			"info": {
				"annotated": true,
			},
		}

		options := NewOptionsBuilder().
			WithPluginFilters(filters).
			Build()

		assert.Equal(t, filters, options.Plugins.Filters)
	})

	t.Run("plugin filter disabling", func(t *testing.T) {
		options := NewOptionsBuilder().
			WithPluginFilter("git", "staged").
			WithoutPluginFilter("git", "staged").
			Build()

		assert.False(t, options.Plugins.Filters["git"]["staged"])
	})

	t.Run("default options", func(t *testing.T) {
		options := DefaultTreeOptions()

		assert.Equal(t, ".", options.Root)
		assert.Equal(t, 3, options.Tree.MaxDepth)
		assert.False(t, options.Tree.ShowHidden)
		assert.False(t, options.Tree.DirsOnly)
		assert.Empty(t, options.Patterns.Excludes)
		assert.Equal(t, ".gitignore", options.Patterns.IgnoreFilePath)
		assert.False(t, options.Patterns.NoIgnoreFile)
		assert.True(t, options.Patterns.UseBuiltinIgnores)
		assert.NotNil(t, options.Plugins.Filters)
		assert.Empty(t, options.Search)
	})

	t.Run("fluent interface chaining", func(t *testing.T) {
		// Test that all methods return builder for chaining
		options := NewOptionsBuilder().
			WithRoot("/custom").
			WithMaxDepth(10).
			WithHidden().
			WithDirsOnly().
			WithExclude("*.log").
			WithExcludes("*.tmp", "*.bak").
			WithIgnoreFile(".customignore").
			WithoutIgnoreFile().
			WithBuiltinIgnores().
			WithoutBuiltinIgnores().
			WithPluginFilter("test", "category").
			WithoutPluginFilter("test", "category").
			WithSearch("term1", "term2").
			Build()

		// Verify the final result
		assert.Equal(t, "/custom", options.Root)
		assert.Equal(t, 10, options.Tree.MaxDepth)
		assert.True(t, options.Tree.ShowHidden)
		assert.True(t, options.Tree.DirsOnly)
		assert.Contains(t, options.Patterns.Excludes, "*.log")
		assert.Contains(t, options.Patterns.Excludes, "*.tmp")
		assert.Contains(t, options.Patterns.Excludes, "*.bak")
		assert.Equal(t, ".customignore", options.Patterns.IgnoreFilePath)
		assert.True(t, options.Patterns.NoIgnoreFile)
		assert.False(t, options.Patterns.UseBuiltinIgnores)
		assert.False(t, options.Plugins.Filters["test"]["category"])
		assert.Equal(t, []string{"term1", "term2"}, options.Search)
	})
}

func TestOptionsValidation(t *testing.T) {
	t.Run("empty root gets defaulted", func(t *testing.T) {
		options := TreeOptions{Root: ""}
		err := options.Validate()
		require.NoError(t, err)
		assert.Equal(t, ".", options.Root)
	})

	t.Run("zero max depth gets defaulted", func(t *testing.T) {
		options := TreeOptions{Tree: TreeDisplayOptions{MaxDepth: 0}}
		err := options.Validate()
		require.NoError(t, err)
		assert.Equal(t, 3, options.Tree.MaxDepth)
	})

	t.Run("negative max depth gets defaulted", func(t *testing.T) {
		options := TreeOptions{Tree: TreeDisplayOptions{MaxDepth: -1}}
		err := options.Validate()
		require.NoError(t, err)
		assert.Equal(t, 3, options.Tree.MaxDepth)
	})
}
