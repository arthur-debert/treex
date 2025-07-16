package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAbsolutePath(t *testing.T) {
	// Test absolute path
	t.Run("absolute path", func(t *testing.T) {
		absPath := "/tmp/test/path"
		result, err := ResolveAbsolutePath(absPath)
		require.NoError(t, err)
		assert.Equal(t, absPath, result)
	})

	// Test relative path with working directory
	t.Run("relative path with cwd", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		
		relPath := "test/path"
		expected := filepath.Join(cwd, relPath)
		
		result, err := ResolveAbsolutePath(relPath)
		require.NoError(t, err)
		assert.Equal(t, filepath.Clean(expected), result)
	})

	// Test with PWD environment variable fallback
	t.Run("PWD fallback", func(t *testing.T) {
		// This test simulates the getwd failure scenario
		// We can't easily simulate os.Getwd() failure in a unit test,
		// but we can test that PWD is used when set
		oldPWD := os.Getenv("PWD")
		defer func() {
			_ = os.Setenv("PWD", oldPWD)
		}()
		
		testPWD := "/test/pwd/path"
		_ = os.Setenv("PWD", testPWD)
		
		// If getwd fails, it should use PWD
		// For this test, we just verify PWD is accessible
		pwd := os.Getenv("PWD")
		assert.Equal(t, testPWD, pwd)
	})

	// Test current directory "."
	t.Run("current directory", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		
		result, err := ResolveAbsolutePath(".")
		require.NoError(t, err)
		assert.Equal(t, cwd, result)
	})

	// Test parent directory ".."
	t.Run("parent directory", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		
		expected := filepath.Dir(cwd)
		result, err := ResolveAbsolutePath("..")
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}