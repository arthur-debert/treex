package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerger_MergeAnnotations(t *testing.T) {
	merger := NewMerger()

	annotations := []Annotation{
		{
			Path:       "a.txt",
			Annotation: "ann from root",
			InfoFile:   ".info",
			LineNum:    1,
		},
		{
			Path:       "b.txt",
			Annotation: "ann for b from root",
			InfoFile:   ".info",
			LineNum:    2,
		},
		{
			Path:       "../a.txt",
			Annotation: "ann from sub for a",
			InfoFile:   "sub/.info",
			LineNum:    1,
		},
		{
			Path:       "c.txt",
			Annotation: "ann from sub for c",
			InfoFile:   "sub/.info",
			LineNum:    2,
		},
		{
			Path:       "../../a.txt",
			Annotation: "ann from sub/d for a (deepest)",
			InfoFile:   "sub/d/.info",
			LineNum:    1,
		},
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

	result := merger.MergeAnnotations(annotations, pathExists)

	require.Len(t, result, 3)

	// a.txt should have annotation from sub/d/.info because it's deepest
	annA, ok := result["a.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub/d for a (deepest)", annA.Annotation)
	assert.Equal(t, "sub/d/.info", annA.InfoFile)

	// b.txt should have annotation from root .info
	annB, ok := result["b.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for b from root", annB.Annotation)
	assert.Equal(t, ".info", annB.InfoFile)

	// c.txt should have annotation from sub/.info
	annC, ok := result["sub/c.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann from sub for c", annC.Annotation)
	assert.Equal(t, "sub/.info", annC.InfoFile)
}

func TestMerger_TieBreaking(t *testing.T) {
	// Test lexicographical tie-breaking when depth is the same.
	merger := NewMerger()

	annotations := []Annotation{
		{
			Path:       "../target.txt",
			Annotation: "ann from sub_a",
			InfoFile:   "sub_a/.info",
			LineNum:    1,
		},
		{
			Path:       "../target.txt",
			Annotation: "ann from sub_b",
			InfoFile:   "sub_b/.info",
			LineNum:    1,
		},
	}

	// Mock pathExists function
	pathExists := func(path string) bool {
		return path == "target.txt"
	}

	result := merger.MergeAnnotations(annotations, pathExists)

	require.Len(t, result, 1)
	ann, ok := result["target.txt"]
	require.True(t, ok)
	// sub_a comes before sub_b lexicographically
	assert.Equal(t, "ann from sub_a", ann.Annotation)
	assert.Equal(t, "sub_a/.info", ann.InfoFile)
}

func TestMerger_AnnotationForDot(t *testing.T) {
	merger := NewMerger()

	annotations := []Annotation{
		{
			Path:       ".",
			Annotation: "ann for sub dir",
			InfoFile:   "sub/.info",
			LineNum:    1,
		},
	}

	// Mock pathExists function
	pathExists := func(path string) bool {
		return path == "sub"
	}

	result := merger.MergeAnnotations(annotations, pathExists)

	require.Len(t, result, 1)
	ann, ok := result["sub"]
	require.True(t, ok)
	assert.Equal(t, "ann for sub dir", ann.Annotation)
	assert.Equal(t, "sub/.info", ann.InfoFile)
}

func TestMerger_CannotAnnotateAncestors(t *testing.T) {
	merger := NewMerger()

	annotations := []Annotation{
		{
			Path:       "../..",
			Annotation: "ann for root from sub/d (invalid)",
			InfoFile:   "sub/d/.info",
			LineNum:    1,
		},
		{
			Path:       "../c.txt",
			Annotation: "ann for sibling of parent (valid)",
			InfoFile:   "sub/d/.info",
			LineNum:    2,
		},
	}

	// Mock pathExists function
	pathExists := func(path string) bool {
		return path == "sub/c.txt"
	}

	result := merger.MergeAnnotations(annotations, pathExists)

	require.Len(t, result, 1)
	ann, ok := result["sub/c.txt"]
	require.True(t, ok)
	assert.Equal(t, "ann for sibling of parent (valid)", ann.Annotation)

	// Invalid ancestor annotations are handled gracefully (warnings logged via global logger)
}

func TestPathDepth(t *testing.T) {
	tests := []struct {
		path          string
		expectedDepth int
	}{
		{".", 0},
		{"a", 1},
		{"a/b", 2},
		{"a/b/c", 3},
		{"./a", 1},
		{"./a/b", 2},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			depth := pathDepth(tt.path)
			assert.Equal(t, tt.expectedDepth, depth)
		})
	}
}
