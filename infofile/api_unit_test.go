package infofile

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// TestInfoAPI_InputValidation tests API input validation logic without filesystem
// These unit tests cover core validation logic that was only tested via integration tests

func TestInfoAPI_InputValidation(t *testing.T) {
	// Create API with minimal filesystem for validation testing
	fs := afero.NewMemMapFs()
	api := NewInfoAPI(fs)

	t.Run("Add validates empty path", func(t *testing.T) {
		err := api.Add(".info", "", "Test annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path cannot be empty")
	})

	t.Run("Add validates whitespace-only path", func(t *testing.T) {
		err := api.Add(".info", "   ", "Test annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path cannot be empty")
	})

	t.Run("Add validates empty annotation", func(t *testing.T) {
		err := api.Add(".info", "target.txt", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation cannot be empty")
	})

	t.Run("Add validates whitespace-only annotation", func(t *testing.T) {
		err := api.Add(".info", "target.txt", "   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation cannot be empty")
	})

	t.Run("Remove validates empty path", func(t *testing.T) {
		err := api.Remove(".info", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path cannot be empty")
	})

	t.Run("Remove validates whitespace-only path", func(t *testing.T) {
		err := api.Remove(".info", "   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path cannot be empty")
	})

	t.Run("Update validates empty path", func(t *testing.T) {
		err := api.Update(".info", "", "New annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path cannot be empty")
	})

	t.Run("Update validates whitespace-only path", func(t *testing.T) {
		err := api.Update(".info", "   ", "New annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path cannot be empty")
	})

	t.Run("Update validates empty annotation", func(t *testing.T) {
		err := api.Update(".info", "target.txt", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation cannot be empty")
	})

	t.Run("Update validates whitespace-only annotation", func(t *testing.T) {
		err := api.Update(".info", "target.txt", "   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation cannot be empty")
	})
}

// TestInfoAPI_ErrorHandling tests API error handling logic
// These unit tests cover error handling paths that were only tested via integration

func TestInfoAPI_ErrorHandling(t *testing.T) {
	fs := afero.NewMemMapFs()
	api := NewInfoAPI(fs)

	t.Run("GetAnnotation handles missing annotation", func(t *testing.T) {
		// Test that GetAnnotation properly returns error for missing annotations
		_, err := api.GetAnnotation("nonexistent.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no annotation found for path")
	})

	t.Run("Remove handles missing info file", func(t *testing.T) {
		// Test that Remove properly handles missing .info files
		err := api.Remove("nonexistent/.info", "target.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot load .info file")
	})

	t.Run("Update handles missing info file", func(t *testing.T) {
		// Test that Update properly handles missing .info files
		err := api.Update("nonexistent/.info", "target.txt", "New annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot load .info file")
	})

	t.Run("Remove handles missing annotation", func(t *testing.T) {
		// Create .info file with one annotation
		err := afero.WriteFile(fs, "test.info", []byte("existing.txt  Existing annotation"), 0644)
		assert.NoError(t, err)

		// Try to remove non-existent annotation
		err = api.Remove("test.info", "nonexistent.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation not found for path")
	})

	t.Run("Update handles missing annotation", func(t *testing.T) {
		// Create .info file with one annotation
		err := afero.WriteFile(fs, "test2.info", []byte("existing.txt  Existing annotation"), 0644)
		assert.NoError(t, err)

		// Try to update non-existent annotation
		err = api.Update("test2.info", "nonexistent.txt", "New annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation not found for path")
	})

	t.Run("Add handles duplicate annotation", func(t *testing.T) {
		// Create .info file with one annotation
		err := afero.WriteFile(fs, "test3.info", []byte("existing.txt  Existing annotation"), 0644)
		assert.NoError(t, err)

		// Try to add duplicate annotation
		err = api.Add("test3.info", "existing.txt", "Duplicate annotation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation already exists for path")
	})
}
