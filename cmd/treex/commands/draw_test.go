package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVirtualTree(t *testing.T) {
	// Test simple annotation
	annotations := map[string]*types.Annotation{
		"file1.txt": {Path: "file1.txt", Notes: "First file"},
		"file2.txt": {Path: "file2.txt", Notes: "Second file"},
	}
	
	tree, err := BuildVirtualTree(annotations)
	require.NoError(t, err)
	
	assert.Equal(t, "root", tree.Name)
	assert.True(t, tree.IsDir)
	assert.Len(t, tree.Children, 2)
	
	// Test with empty annotations
	emptyTree, err := BuildVirtualTree(map[string]*types.Annotation{})
	assert.Error(t, err)
	assert.Nil(t, emptyTree)
}

func TestDrawCommandBasic(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.info")
	
	testContent := `Dad Chill, dad
Mom Listen to your mother`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)
	
	// Test that the command can be executed (basic smoke test)
	cmd := drawCmd
	cmd.SetArgs([]string{"--info-file", testFile})
	
	// Just check that the command doesn't crash
	err = cmd.Execute()
	require.NoError(t, err)
}