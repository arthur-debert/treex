package info

import (
	"strings"
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherer_GatherFromMap(t *testing.T) {
	gatherer := NewGatherer()

	infoFiles := InfoFileMap{
		".info": `
a.txt  ann from root
b.txt  ann for b from root
`,
		"sub/.info": `
../a.txt  ann from sub for a
c.txt     ann from sub for c
`,
		"sub/d/.info": `
../../a.txt ann from sub/d for a (deepest)
`,
	}

	// Mock pathExists function
	existingPaths := map[string]bool{
		"a.txt":     true,
		"b.txt":     true,
		"sub/c.txt": true,
	}
	pathExists := func(path string) bool {
		return existingPaths[path]
	}

	annotations, err := gatherer.GatherFromMap(infoFiles, pathExists)
	require.NoError(t, err)

	require.Len(t, annotations, 3)

	// a.txt should have annotation from sub/d/.info because it's deepest
	annA, ok := annotations["a.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub/d for a (deepest)", annA.Annotation)
	assert.Equal(t, "sub/d/.info", annA.InfoFile)

	// b.txt should have annotation from root .info
	annB, ok := annotations["b.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for b from root", annB.Annotation)
	assert.Equal(t, ".info", annB.InfoFile)

	// c.txt should have annotation from sub/.info
	annC, ok := annotations["sub/c.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub for c", annC.Annotation)
	assert.Equal(t, "sub/.info", annC.InfoFile)
}

func TestGatherer_GatherFromFileSystem(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info": `
a.txt  ann from root
b.txt  ann for b from root
`,
		"a.txt": "content a",
		"b.txt": "content b",
		"sub": map[string]interface{}{
			".info": `
../a.txt  ann from sub for a
c.txt     ann from sub for c
`,
			"c.txt": "content c",
			"d": map[string]interface{}{
				".info": `
../../a.txt ann from sub/d for a (deepest)
`,
				"e.txt": "content e",
			},
		},
	})

	infoFS := NewAferoInfoFileSystem(fsys)
	gatherer := NewGatherer()
	annotations, err := gatherer.GatherFromFileSystem(infoFS, ".")
	require.NoError(t, err)

	require.Len(t, annotations, 3)

	// a.txt should have annotation from sub/d/.info because it's deepest
	annA, ok := annotations["a.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub/d for a (deepest)", annA.Annotation)
	assert.Equal(t, "sub/d/.info", annA.InfoFile)

	// b.txt should have annotation from root .info
	annB, ok := annotations["b.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for b from root", annB.Annotation)
	assert.Equal(t, ".info", annB.InfoFile)

	// c.txt should have annotation from sub/.info
	annC, ok := annotations["sub/c.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub for c", annC.Annotation)
	assert.Equal(t, "sub/.info", annC.InfoFile)
}

func TestGatherer_WithLogger(t *testing.T) {
	logger := &TestLogger{}
	gatherer := NewGathererWithLogger(logger)

	infoFiles := InfoFileMap{
		".info": `
a.txt  ann for a
a.txt  duplicate (should warn)
invalid_line
`,
		"sub/.info": `
../..  invalid ancestor annotation
../a.txt  valid annotation
`,
	}

	// Mock pathExists function - only a.txt exists
	existingPaths := map[string]bool{
		"a.txt": true,
	}
	pathExists := func(path string) bool {
		return existingPaths[path]
	}

	annotations, err := gatherer.GatherFromMap(infoFiles, pathExists)
	require.NoError(t, err)

	// Should collect valid annotations
	require.Len(t, annotations, 1)
	ann, ok := annotations["a.txt"]
	require.True(t, ok)
	assert.Equal(t, "valid annotation", ann.Annotation) // sub/.info should win (deeper)
	assert.Equal(t, "sub/.info", ann.InfoFile)

	// Should have logged warnings
	messages := logger.GetMessages()
	require.GreaterOrEqual(t, len(messages), 3)

	// Check for specific warning types
	var hasDuplicateWarning, hasInvalidLineWarning, hasAncestorWarning bool
	for _, msg := range messages {
		if strings.Contains(msg, "duplicate path") {
			hasDuplicateWarning = true
		}
		if strings.Contains(msg, "no annotation found") {
			hasInvalidLineWarning = true
		}
		if strings.Contains(msg, "cannot annotate ancestor") {
			hasAncestorWarning = true
		}
	}

	assert.True(t, hasDuplicateWarning, "Should warn about duplicate paths")
	assert.True(t, hasInvalidLineWarning, "Should warn about invalid lines")
	assert.True(t, hasAncestorWarning, "Should warn about ancestor annotations")
}
