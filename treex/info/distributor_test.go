package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistributor_AddAnnotation(t *testing.T) {
	distributor := NewDistributor()

	tests := []struct {
		name            string
		targetPath      string
		annotation      string
		infoFile        string
		existingContent string
		expectedContent string
	}{
		{
			name:            "add to empty file",
			targetPath:      "target.txt",
			annotation:      "Test annotation",
			infoFile:        ".info",
			existingContent: "",
			expectedContent: "target.txt  Test annotation\n",
		},
		{
			name:            "add to existing file",
			targetPath:      "target.txt",
			annotation:      "Test annotation",
			infoFile:        ".info",
			existingContent: "existing.txt  Existing annotation",
			expectedContent: "existing.txt  Existing annotation\ntarget.txt  Test annotation\n",
		},
		{
			name:            "add with relative path",
			targetPath:      "sub/target.txt",
			annotation:      "Sub annotation",
			infoFile:        "sub/.info",
			existingContent: "",
			expectedContent: "target.txt  Sub annotation\n",
		},
		{
			name:            "add with spaces in path",
			targetPath:      "path with spaces.txt",
			annotation:      "Spaced annotation",
			infoFile:        ".info",
			existingContent: "",
			expectedContent: "path\\ with\\ spaces.txt  Spaced annotation\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := distributor.AddAnnotation(tt.targetPath, tt.annotation, tt.infoFile, tt.existingContent)

			assert.Equal(t, OpUpdate, operation.Type)
			assert.Equal(t, tt.infoFile, operation.FilePath)
			assert.Equal(t, tt.expectedContent, operation.Content)
		})
	}
}

func TestDistributor_RemoveAnnotation(t *testing.T) {
	distributor := NewDistributor()

	tests := []struct {
		name            string
		annotation      Annotation
		existingContent string
		expectedType    FileOperationType
		expectedContent string
	}{
		{
			name: "remove from multi-line file",
			annotation: Annotation{
				Path:     "target.txt",
				InfoFile: ".info",
				LineNum:  2,
			},
			existingContent: "first.txt  First annotation\ntarget.txt  Target annotation\nthird.txt  Third annotation",
			expectedType:    OpUpdate,
			expectedContent: "first.txt  First annotation\nthird.txt  Third annotation",
		},
		{
			name: "remove only line",
			annotation: Annotation{
				Path:     "target.txt",
				InfoFile: ".info",
				LineNum:  1,
			},
			existingContent: "target.txt  Only annotation",
			expectedType:    OpDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation := distributor.RemoveAnnotation(tt.annotation, tt.existingContent)

			assert.Equal(t, tt.expectedType, operation.Type)
			assert.Equal(t, tt.annotation.InfoFile, operation.FilePath)
			if tt.expectedType == OpUpdate {
				assert.Equal(t, tt.expectedContent, operation.Content)
			}
		})
	}
}

func TestDistributor_UpdateAnnotation(t *testing.T) {
	distributor := NewDistributor()

	annotation := Annotation{
		Path:       "target.txt",
		Annotation: "Old annotation",
		InfoFile:   ".info",
		LineNum:    2,
	}

	existingContent := "first.txt  First annotation\ntarget.txt  Old annotation\nthird.txt  Third annotation"
	newAnnotation := "Updated annotation"

	operation := distributor.UpdateAnnotation(annotation, newAnnotation, existingContent)

	assert.Equal(t, OpUpdate, operation.Type)
	assert.Equal(t, ".info", operation.FilePath)

	expectedContent := "first.txt  First annotation\ntarget.txt  Updated annotation\nthird.txt  Third annotation"
	assert.Equal(t, expectedContent, operation.Content)
}

func TestDistributor_DistributeAnnotations(t *testing.T) {
	distributor := NewDistributor()

	annotations := map[string]Annotation{
		"path1": {
			Path:       "a.txt",
			Annotation: "Annotation for a",
			InfoFile:   ".info",
			LineNum:    1,
		},
		"path2": {
			Path:       "b.txt",
			Annotation: "Annotation for b",
			InfoFile:   ".info",
			LineNum:    2,
		},
		"path3": {
			Path:       "c.txt",
			Annotation: "Annotation for c",
			InfoFile:   "sub/.info",
			LineNum:    1,
		},
	}

	operations := distributor.DistributeAnnotations(annotations)

	require.Len(t, operations, 2) // Should have operations for .info and sub/.info

	// Check that all operations are updates
	for _, op := range operations {
		assert.Equal(t, OpUpdate, op.Type)
		assert.NotEmpty(t, op.Content)
	}

	// Check that the root .info file contains a.txt and b.txt
	var rootOp *FileOperation
	for _, op := range operations {
		if op.FilePath == ".info" {
			rootOp = &op
			break
		}
	}
	require.NotNil(t, rootOp)
	assert.Contains(t, rootOp.Content, "a.txt  Annotation for a")
	assert.Contains(t, rootOp.Content, "b.txt  Annotation for b")
}

func TestDistributor_generateInfoFileContent(t *testing.T) {
	distributor := NewDistributor()

	annotations := []Annotation{
		{
			Path:       "z.txt",
			Annotation: "Last annotation",
			InfoFile:   ".info",
			LineNum:    1,
		},
		{
			Path:       "a.txt",
			Annotation: "First annotation",
			InfoFile:   ".info",
			LineNum:    2,
		},
		{
			Path:       "m.txt",
			Annotation: "Middle annotation",
			InfoFile:   ".info",
			LineNum:    3,
		},
	}

	content := distributor.editor.GenerateContent(annotations, ".info")

	// Should be sorted by path
	expectedContent := "a.txt  First annotation\nm.txt  Middle annotation\nz.txt  Last annotation\n"
	assert.Equal(t, expectedContent, content)
}

func TestDistributor_determineInfoFile(t *testing.T) {
	distributor := NewDistributor()

	tests := []struct {
		name       string
		targetPath string
		strategy   DistributionStrategy
		expected   string
	}{
		{
			name:       "root file with depth strategy",
			targetPath: "target.txt",
			strategy:   DistributeByDepth,
			expected:   ".info",
		},
		{
			name:       "subdirectory file with depth strategy",
			targetPath: "sub/target.txt",
			strategy:   DistributeByDepth,
			expected:   "sub/.info",
		},
		{
			name:       "any file with consolidate strategy",
			targetPath: "sub/deep/target.txt",
			strategy:   DistributeConsolidate,
			expected:   ".info",
		},
		{
			name:       "subdirectory file with proximity strategy",
			targetPath: "sub/target.txt",
			strategy:   DistributeByProximity,
			expected:   "sub/.info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distributor.determineInfoFile(tt.targetPath, tt.strategy)
			assert.Equal(t, tt.expected, result)
		})
	}
}
