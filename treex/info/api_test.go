package info

import (
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoAPI_Gather(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info": `
a.txt  ann from root
b.txt  ann for b from root
sub/d.txt  ann for d from root
`,
		"a.txt": "content a",
		"b.txt": "content b",
		"sub": map[string]interface{}{
			".info": `
c.txt  ann from sub for c
d.txt  ann for d from sub
`,
			"c.txt": "content c",
			"d.txt": "content d",
		},
	})

	api := NewInfoAPI(fsys)
	annotations, err := api.Gather(".")
	require.NoError(t, err)

	require.Len(t, annotations, 4)

	// a.txt should have annotation from root .info (only one annotation)
	annA, ok := annotations["a.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from root", annA.Annotation)
	assert.Equal(t, ".info", annA.InfoFile)

	// b.txt should have annotation from root .info (only one annotation)
	annB, ok := annotations["b.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for b from root", annB.Annotation)
	assert.Equal(t, ".info", annB.InfoFile)

	// c.txt should have annotation from sub/.info (only one annotation)
	annC, ok := annotations["sub/c.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub for c", annC.Annotation)
	assert.Equal(t, "sub/.info", annC.InfoFile)

	// d.txt has conflicting annotations - sub/.info should win because it's closer
	annD, ok := annotations["sub/d.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for d from sub", annD.Annotation)
	assert.Equal(t, "sub/.info", annD.InfoFile)
}

func TestInfoAPI_Validate(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info": `
valid.txt    Valid annotation
invalid_line
missing.txt  Missing file annotation
`,
		"valid.txt": "content",
	})

	api := NewInfoAPI(fsys)
	result, err := api.Validate(".")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have issues for invalid line and missing file
	assert.Greater(t, len(result.Issues), 0)

	// Check for expected issue types
	var hasInvalidFormat, hasPathNotExists bool
	for _, issue := range result.Issues {
		if issue.Type == IssueInvalidFormat {
			hasInvalidFormat = true
		}
		if issue.Type == IssuePathNotExists {
			hasPathNotExists = true
		}
	}

	assert.True(t, hasInvalidFormat, "Should have invalid format issue")
	assert.True(t, hasPathNotExists, "Should have path not exists issue")
}

func TestInfoAPI_Add(t *testing.T) {
	fsys := testutil.NewTestFS()
	api := NewInfoAPI(fsys)

	// Add annotation to specific .info file
	err := api.Add(".info", "target.txt", "Test annotation")
	require.NoError(t, err)

	// Test simple validation by checking the file was written
	exists, err := afero.Exists(fsys, ".info")
	require.NoError(t, err)
	assert.True(t, exists, ".info file should exist after Add")

	// Read and verify content format
	content, err := afero.ReadFile(fsys, ".info")
	require.NoError(t, err)
	assert.Equal(t, "target.txt Test annotation", string(content))
}

func TestInfoAPI_Add_Validation(t *testing.T) {
	fsys := testutil.NewTestFS()
	api := NewInfoAPI(fsys)

	// Test empty path validation
	err := api.Add(".info", "", "Test annotation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target path cannot be empty")

	// Test empty annotation validation
	err = api.Add(".info", "target.txt", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "annotation cannot be empty")

	// Test whitespace-only path
	err = api.Add(".info", "   ", "Test annotation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target path cannot be empty")

	// Test whitespace-only annotation
	err = api.Add(".info", "target.txt", "   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "annotation cannot be empty")
}

func TestInfoAPI_Remove(t *testing.T) {
	fsys := testutil.NewTestFS()
	api := NewInfoAPI(fsys)

	// First add an annotation
	err := api.Add(".info", "target.txt", "Test annotation")
	require.NoError(t, err)

	// Remove the annotation
	err = api.Remove(".info", "target.txt")
	require.NoError(t, err)

	// Verify annotation was removed
	annotations, err := api.Gather(".")
	require.NoError(t, err)
	assert.Len(t, annotations, 0)
}

func TestInfoAPI_Update(t *testing.T) {
	fsys := testutil.NewTestFS()
	api := NewInfoAPI(fsys)

	// First add an annotation
	err := api.Add(".info", "target.txt", "Original annotation")
	require.NoError(t, err)

	// Update the annotation
	err = api.Update(".info", "target.txt", "Updated annotation")
	require.NoError(t, err)

	// Verify the file content was updated correctly
	content, err := afero.ReadFile(fsys, ".info")
	require.NoError(t, err)
	assert.Equal(t, "target.txt Updated annotation", string(content))
}

func TestInfoAPI_List(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info": `
a.txt  Annotation for a
b.txt  Annotation for b
`,
		"a.txt": "content a",
		"b.txt": "content b",
		"sub": map[string]interface{}{
			".info": "c.txt  Annotation for c",
			"c.txt": "content c",
		},
	})

	api := NewInfoAPI(fsys)
	annotations, err := api.List(".")
	require.NoError(t, err)

	// Should have 3 annotations
	assert.Len(t, annotations, 3)

	// Check that all expected annotations are present
	paths := make(map[string]bool)
	for _, ann := range annotations {
		switch ann.InfoFile {
		case ".info", "sub/.info":
			paths[ann.Path] = true
		}
	}

	assert.True(t, paths["a.txt"] || paths["b.txt"], "Should have annotations from root .info")
}

func TestInfoAPI_GetAnnotation(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info":      "target.txt  Test annotation",
		"target.txt": "content",
	})

	api := NewInfoAPI(fsys)

	// Get existing annotation
	ann, err := api.GetAnnotation("target.txt")
	require.NoError(t, err)
	require.NotNil(t, ann)
	assert.Equal(t, "Test annotation", ann.Annotation)

	// Try to get non-existent annotation
	_, err = api.GetAnnotation("nonexistent.txt")
	assert.Error(t, err)
}

func TestInfoAPI_Clean(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info": `
valid.txt     Valid annotation
missing.txt   Annotation for missing file
duplicate.txt First annotation
duplicate.txt Duplicate annotation
invalid_line
`,
		"valid.txt": "content",
	})

	api := NewInfoAPI(fsys)

	// Clean invalid annotations
	result, err := api.Clean(".")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have removed some annotations
	assert.Greater(t, len(result.RemovedAnnotations), 0)
	assert.Greater(t, result.Summary.FilesModified, 0)

	// Verify that only valid annotations remain
	annotations, err := api.Gather(".")
	require.NoError(t, err)

	// Should only have the valid annotation
	assert.Len(t, annotations, 1)
	ann, ok := annotations["valid.txt"]
	require.True(t, ok)
	assert.Equal(t, "Valid annotation", ann.Annotation)
}
