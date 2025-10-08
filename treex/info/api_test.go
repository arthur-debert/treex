package info

import (
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
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
	fsys.MustCreateTree(".", map[string]interface{}{
		"target.txt": "content",
	})

	api := NewInfoAPI(fsys)

	// Add annotation
	err := api.Add("target.txt", "Test annotation")
	require.NoError(t, err)

	// Verify annotation was added
	annotations, err := api.Gather(".")
	require.NoError(t, err)

	ann, ok := annotations["target.txt"]
	require.True(t, ok)
	assert.Equal(t, "Test annotation", ann.Annotation)
}

func TestInfoAPI_AddToSubdirectory(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		"sub": map[string]interface{}{
			"target.txt": "content",
		},
	})

	api := NewInfoAPI(fsys)

	// Add annotation to subdirectory file
	err := api.Add("sub/target.txt", "Sub annotation")
	require.NoError(t, err)

	// Verify annotation was added to sub/.info
	annotations, err := api.Gather(".")
	require.NoError(t, err)

	ann, ok := annotations["sub/target.txt"]
	require.True(t, ok, "Should find annotation for sub/target.txt")
	assert.Equal(t, "Sub annotation", ann.Annotation)
	assert.Equal(t, "sub/.info", ann.InfoFile)
}

func TestInfoAPI_Remove(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info":      "target.txt  Test annotation",
		"target.txt": "content",
	})

	api := NewInfoAPI(fsys)

	// Verify annotation exists first
	annotations, err := api.Gather(".")
	require.NoError(t, err)
	_, ok := annotations["target.txt"]
	require.True(t, ok)

	// Remove annotation
	err = api.Remove("target.txt")
	require.NoError(t, err)

	// Verify annotation was removed
	annotations, err = api.Gather(".")
	require.NoError(t, err)
	_, ok = annotations["target.txt"]
	assert.False(t, ok)
}

func TestInfoAPI_Update(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info":      "target.txt  Original annotation",
		"target.txt": "content",
	})

	api := NewInfoAPI(fsys)

	// Update annotation
	err := api.Update("target.txt", "Updated annotation")
	require.NoError(t, err)

	// Verify annotation was updated
	annotations, err := api.Gather(".")
	require.NoError(t, err)

	ann, ok := annotations["target.txt"]
	require.True(t, ok)
	assert.Equal(t, "Updated annotation", ann.Annotation)
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
