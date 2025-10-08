package info_test

import (
	"strings"
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/jwaldrip/treex/treex/info"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	content := `
# This is a comment
a.txt   Annotation for a
b/c.txt Annotation for b/c

  d.txt		Annotation for d with tabs and spaces
.       Annotation for current dir
path\ with\ spaces.txt  An annotation for a path with spaces
a.txt   This should be ignored because a.txt is already present
no_annotation
`
	reader := strings.NewReader(content)
	annotations, err := info.Parse(reader, "/path/.info")
	require.NoError(t, err)

	require.Len(t, annotations, 5)
	assert.Equal(t, "a.txt", annotations[0].Path)
	assert.Equal(t, "Annotation for a", annotations[0].Annotation)
	assert.Equal(t, "/path/.info", annotations[0].InfoFile)
	assert.Equal(t, 3, annotations[0].LineNum)

	assert.Equal(t, "b/c.txt", annotations[1].Path)
	assert.Equal(t, "Annotation for b/c", annotations[1].Annotation)

	assert.Equal(t, "d.txt", annotations[2].Path)
	assert.Equal(t, "Annotation for d with tabs and spaces", annotations[2].Annotation)

	assert.Equal(t, ".", annotations[3].Path)
	assert.Equal(t, "Annotation for current dir", annotations[3].Annotation)

	assert.Equal(t, "path with spaces.txt", annotations[4].Path)
	assert.Equal(t, "An annotation for a path with spaces", annotations[4].Annotation)
}

func TestCollectAndMerge(t *testing.T) {
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

	collector := info.NewCollector()
	annotations, err := collector.CollectAnnotations(fsys, ".")
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

func TestMergeTieBreaking(t *testing.T) {
	// Test lexicographical tie-breaking when depth is the same.
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		"sub_a": map[string]interface{}{
			".info": `../target.txt  ann from sub_a`,
		},
		"sub_b": map[string]interface{}{
			".info": `../target.txt  ann from sub_b`,
		},
		"target.txt": "",
	})

	collector := info.NewCollector()
	annotations, err := collector.CollectAnnotations(fsys, ".")
	require.NoError(t, err)

	require.Len(t, annotations, 1)
	ann, ok := annotations["target.txt"]
	require.True(t, ok)
	// sub_a comes before sub_b lexicographically
	assert.Equal(t, "ann from sub_a", ann.Annotation)
	assert.Equal(t, "sub_a/.info", ann.InfoFile)
}

func TestAnnotationForDot(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		"sub": map[string]interface{}{
			".info":     `.  ann for sub dir`,
			"a.txt":     "",
		},
	})

	collector := info.NewCollector()
	annotations, err := collector.CollectAnnotations(fsys, ".")
	require.NoError(t, err)

	require.Len(t, annotations, 1)
	ann, ok := annotations["sub"]
	require.True(t, ok)
	assert.Equal(t, "ann for sub dir", ann.Annotation)
	assert.Equal(t, "sub/.info", ann.InfoFile)
}

func TestCannotAnnotateAncestors(t *testing.T) {
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		"sub": map[string]interface{}{
			"d": map[string]interface{}{
				".info": `
../..  ann for root from sub/d (invalid)
../c.txt ann for sibling of parent (valid)
`,
			},
			"c.txt": "",
		},
	})

	collector := info.NewCollector()
	annotations, err := collector.CollectAnnotations(fsys, ".")
	require.NoError(t, err)

	require.Len(t, annotations, 1)
	ann, ok := annotations["sub/c.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for sibling of parent (valid)", ann.Annotation)
}
