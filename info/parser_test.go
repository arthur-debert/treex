package info

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogger captures log messages for testing
type TestLogger struct {
	messages []string
}

func (l *TestLogger) Printf(format string, v ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(format, v...))
}

func (l *TestLogger) GetMessages() []string {
	return l.messages
}

func (l *TestLogger) Reset() {
	l.messages = nil
}

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
	infoFile := NewInfoFile("/path/.info", content)
	annotations := infoFile.GetAllAnnotations()

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

func TestParse_DuplicatePaths(t *testing.T) {
	content := `
a.txt  First annotation
a.txt  Second annotation (should be ignored)
b.txt  Annotation for b
a.txt  Third annotation (should also be ignored)
`
	infoFile := NewInfoFile("/test/.info", content)
	annotations := infoFile.GetAllAnnotations()

	// Should only have 2 annotations (a.txt and b.txt, duplicates ignored)
	require.Len(t, annotations, 2)
	assert.Equal(t, "a.txt", annotations[0].Path)
	assert.Equal(t, "First annotation", annotations[0].Annotation)
	assert.Equal(t, "b.txt", annotations[1].Path)

	// Duplicate paths are handled gracefully (warnings logged via global logger)
}

func TestParse_InvalidLines(t *testing.T) {
	// Build content with different types of invalid lines
	lines := []string{
		"# Valid comment",
		"a.txt  Valid annotation",
		"invalid_line_no_space",
		"b.txt", // This line has no space/annotation
		"c.txt  Valid annotation",
	}
	content := strings.Join(lines, "\n")
	infoFile := NewInfoFile("/test/.info", content)
	annotations := infoFile.GetAllAnnotations()

	// Should only have 2 valid annotations
	require.Len(t, annotations, 2)
	assert.Equal(t, "a.txt", annotations[0].Path)
	assert.Equal(t, "c.txt", annotations[1].Path)

	// Invalid lines are handled gracefully (warnings logged via global logger)
}

func TestParser_ParseLine(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name          string
		line          string
		expectedPath  string
		expectedAnnot string
		expectedOk    bool
	}{
		{
			name:          "valid line",
			line:          "path.txt  annotation text",
			expectedPath:  "path.txt",
			expectedAnnot: "annotation text",
			expectedOk:    true,
		},
		{
			name:          "escaped spaces",
			line:          "path\\ with\\ spaces.txt  annotation",
			expectedPath:  "path with spaces.txt",
			expectedAnnot: "annotation",
			expectedOk:    true,
		},
		{
			name:       "no space separator",
			line:       "pathwithoutspace",
			expectedOk: false,
		},
		{
			name:       "empty annotation",
			line:       "path.txt  ",
			expectedOk: false,
		},
		{
			name:          "tabs and spaces",
			line:          "path.txt\t\tAnnotation with tabs",
			expectedPath:  "path.txt",
			expectedAnnot: "Annotation with tabs",
			expectedOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, annotation, ok := parser.ParseLine(tt.line)
			assert.Equal(t, tt.expectedOk, ok)
			if tt.expectedOk {
				assert.Equal(t, tt.expectedPath, path)
				assert.Equal(t, tt.expectedAnnot, annotation)
			}
		})
	}
}
