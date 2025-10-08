package info

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditor_AddAnnotation(t *testing.T) {
	editor := NewEditor()

	tests := []struct {
		name            string
		content         string
		targetPath      string
		annotation      string
		infoFilePath    string
		expectedContent string
	}{
		{
			name:            "add to empty content",
			content:         "",
			targetPath:      "target.txt",
			annotation:      "Test annotation",
			infoFilePath:    ".info",
			expectedContent: "target.txt  Test annotation\n",
		},
		{
			name:            "add to existing content",
			content:         "existing.txt  Existing annotation",
			targetPath:      "target.txt",
			annotation:      "Test annotation",
			infoFilePath:    ".info",
			expectedContent: "existing.txt  Existing annotation\ntarget.txt  Test annotation\n",
		},
		{
			name:            "add to existing content with newline",
			content:         "existing.txt  Existing annotation\n",
			targetPath:      "target.txt",
			annotation:      "Test annotation",
			infoFilePath:    ".info",
			expectedContent: "existing.txt  Existing annotation\ntarget.txt  Test annotation\n",
		},
		{
			name:            "add with relative path conversion",
			targetPath:      "sub/target.txt",
			annotation:      "Sub annotation",
			infoFilePath:    "sub/.info",
			content:         "",
			expectedContent: "target.txt  Sub annotation\n",
		},
		{
			name:            "add with spaces in path",
			targetPath:      "path with spaces.txt",
			annotation:      "Spaced annotation",
			infoFilePath:    ".info",
			content:         "",
			expectedContent: "path\\ with\\ spaces.txt  Spaced annotation\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.AddAnnotation(tt.content, tt.targetPath, tt.annotation, tt.infoFilePath)
			assert.Equal(t, tt.expectedContent, result)
		})
	}
}

func TestEditor_RemoveAnnotation(t *testing.T) {
	editor := NewEditor()

	tests := []struct {
		name            string
		content         string
		lineNum         int
		expectedContent string
	}{
		{
			name:            "remove from multi-line content",
			content:         "first.txt  First annotation\ntarget.txt  Target annotation\nthird.txt  Third annotation",
			lineNum:         2,
			expectedContent: "first.txt  First annotation\nthird.txt  Third annotation",
		},
		{
			name:            "remove only line",
			content:         "target.txt  Only annotation",
			lineNum:         1,
			expectedContent: "",
		},
		{
			name:            "remove first line",
			content:         "first.txt  First annotation\nsecond.txt  Second annotation",
			lineNum:         1,
			expectedContent: "second.txt  Second annotation",
		},
		{
			name:            "remove last line",
			content:         "first.txt  First annotation\nsecond.txt  Second annotation",
			lineNum:         2,
			expectedContent: "first.txt  First annotation",
		},
		{
			name:            "remove out of bounds",
			content:         "first.txt  First annotation",
			lineNum:         5,
			expectedContent: "first.txt  First annotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.RemoveAnnotation(tt.content, tt.lineNum)
			assert.Equal(t, tt.expectedContent, result)
		})
	}
}

func TestEditor_UpdateAnnotation(t *testing.T) {
	editor := NewEditor()

	tests := []struct {
		name            string
		content         string
		lineNum         int
		targetPath      string
		newAnnotation   string
		infoFilePath    string
		expectedContent string
	}{
		{
			name:            "update middle line",
			content:         "first.txt  First annotation\ntarget.txt  Old annotation\nthird.txt  Third annotation",
			lineNum:         2,
			targetPath:      "target.txt",
			newAnnotation:   "Updated annotation",
			infoFilePath:    ".info",
			expectedContent: "first.txt  First annotation\ntarget.txt  Updated annotation\nthird.txt  Third annotation",
		},
		{
			name:            "update with relative path",
			content:         "target.txt  Old annotation",
			lineNum:         1,
			targetPath:      "sub/target.txt",
			newAnnotation:   "Updated annotation",
			infoFilePath:    "sub/.info",
			expectedContent: "target.txt  Updated annotation",
		},
		{
			name:            "update with spaces in path",
			content:         "path\\ with\\ spaces.txt  Old annotation",
			lineNum:         1,
			targetPath:      "path with spaces.txt",
			newAnnotation:   "Updated annotation",
			infoFilePath:    ".info",
			expectedContent: "path\\ with\\ spaces.txt  Updated annotation",
		},
		{
			name:            "update out of bounds",
			content:         "first.txt  First annotation",
			lineNum:         5,
			targetPath:      "target.txt",
			newAnnotation:   "Updated annotation",
			infoFilePath:    ".info",
			expectedContent: "first.txt  First annotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.UpdateAnnotation(tt.content, tt.lineNum, tt.targetPath, tt.newAnnotation, tt.infoFilePath)
			assert.Equal(t, tt.expectedContent, result)
		})
	}
}

func TestEditor_GenerateContent(t *testing.T) {
	editor := NewEditor()

	tests := []struct {
		name            string
		annotations     []Annotation
		infoFilePath    string
		expectedContent string
	}{
		{
			name:            "empty annotations",
			annotations:     []Annotation{},
			infoFilePath:    ".info",
			expectedContent: "",
		},
		{
			name: "single annotation",
			annotations: []Annotation{
				{
					Path:       "target.txt",
					Annotation: "Test annotation",
					InfoFile:   ".info",
					LineNum:    1,
				},
			},
			infoFilePath:    ".info",
			expectedContent: "target.txt  Test annotation\n",
		},
		{
			name: "multiple annotations sorted by path",
			annotations: []Annotation{
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
			},
			infoFilePath:    ".info",
			expectedContent: "a.txt  First annotation\nm.txt  Middle annotation\nz.txt  Last annotation\n",
		},
		{
			name: "annotation with spaces in path",
			annotations: []Annotation{
				{
					Path:       "path with spaces.txt",
					Annotation: "Spaced annotation",
					InfoFile:   ".info",
					LineNum:    1,
				},
			},
			infoFilePath:    ".info",
			expectedContent: "path\\ with\\ spaces.txt  Spaced annotation\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.GenerateContent(tt.annotations, tt.infoFilePath)
			assert.Equal(t, tt.expectedContent, result)
		})
	}
}

func TestEditor_makeRelativePath(t *testing.T) {
	editor := NewEditor()

	tests := []struct {
		name         string
		targetPath   string
		infoFilePath string
		expected     string
	}{
		{
			name:         "same directory",
			targetPath:   "target.txt",
			infoFilePath: ".info",
			expected:     "target.txt",
		},
		{
			name:         "subdirectory target, root info",
			targetPath:   "sub/target.txt",
			infoFilePath: ".info",
			expected:     "sub/target.txt",
		},
		{
			name:         "same directory as info file",
			targetPath:   "sub/target.txt",
			infoFilePath: "sub/.info",
			expected:     "target.txt",
		},
		{
			name:         "parent directory reference",
			targetPath:   "../target.txt",
			infoFilePath: "sub/.info",
			expected:     "../../target.txt",
		},
		{
			name:         "current directory reference",
			targetPath:   "./target.txt",
			infoFilePath: ".info",
			expected:     "target.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.makeRelativePath(tt.targetPath, tt.infoFilePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEditor_escapePath(t *testing.T) {
	editor := NewEditor()

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "no spaces",
			path:     "target.txt",
			expected: "target.txt",
		},
		{
			name:     "single space",
			path:     "path with space.txt",
			expected: "path\\ with\\ space.txt",
		},
		{
			name:     "multiple spaces",
			path:     "path with many spaces.txt",
			expected: "path\\ with\\ many\\ spaces.txt",
		},
		{
			name:     "leading and trailing spaces",
			path:     " path with spaces ",
			expected: "\\ path\\ with\\ spaces\\ ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editor.escapePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEditor_AddAnnotationToFile(t *testing.T) {
	editor := NewEditor()

	result := editor.AddAnnotationToFile("target.txt", "Test annotation", ".info", "existing.txt  Existing")
	expected := "existing.txt  Existing\ntarget.txt  Test annotation\n"
	assert.Equal(t, expected, result)
}

func TestEditor_RemoveAnnotationFromContent(t *testing.T) {
	editor := NewEditor()

	annotations := map[string]Annotation{
		"target.txt": {
			Path:     "target.txt",
			InfoFile: ".info",
			LineNum:  2,
		},
	}

	content := "first.txt  First\ntarget.txt  Target\nthird.txt  Third"
	newContent, found := editor.RemoveAnnotationFromContent("target.txt", content, annotations)

	assert.True(t, found)
	assert.Equal(t, "first.txt  First\nthird.txt  Third", newContent)

	// Test not found
	_, found = editor.RemoveAnnotationFromContent("missing.txt", content, annotations)
	assert.False(t, found)
}

func TestEditor_UpdateAnnotationInContent(t *testing.T) {
	editor := NewEditor()

	annotations := map[string]Annotation{
		"target.txt": {
			Path:     "target.txt",
			InfoFile: ".info",
			LineNum:  2,
		},
	}

	content := "first.txt  First\ntarget.txt  Old\nthird.txt  Third"
	newContent, found := editor.UpdateAnnotationInContent("target.txt", "New annotation", content, annotations)

	assert.True(t, found)
	assert.Equal(t, "first.txt  First\ntarget.txt  New annotation\nthird.txt  Third", newContent)

	// Test not found
	_, found = editor.UpdateAnnotationInContent("missing.txt", "New", content, annotations)
	assert.False(t, found)
}
