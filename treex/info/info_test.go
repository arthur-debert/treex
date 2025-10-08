package info_test

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/jwaldrip/treex/treex/info"
	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/spf13/afero"
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
			".info": `.  ann for sub dir`,
			"a.txt": "",
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

func TestParseWithLogger_DuplicatePaths(t *testing.T) {
	logger := &TestLogger{}
	content := `
a.txt  First annotation
a.txt  Second annotation (should be ignored)
b.txt  Annotation for b
a.txt  Third annotation (should also be ignored)
`
	reader := strings.NewReader(content)
	annotations, err := info.ParseWithLogger(reader, "/test/.info", logger)
	require.NoError(t, err)

	// Should only have 2 annotations (a.txt and b.txt, duplicates ignored)
	require.Len(t, annotations, 2)
	assert.Equal(t, "a.txt", annotations[0].Path)
	assert.Equal(t, "First annotation", annotations[0].Annotation)
	assert.Equal(t, "b.txt", annotations[1].Path)

	// Should have warnings about duplicate paths
	messages := logger.GetMessages()
	require.Len(t, messages, 2)
	assert.Contains(t, messages[0], "ignoring duplicate path \"a.txt\" at line 3")
	assert.Contains(t, messages[1], "ignoring duplicate path \"a.txt\" at line 5")
}

func TestParseWithLogger_InvalidLines(t *testing.T) {
	logger := &TestLogger{}
	// Build content with different types of invalid lines
	lines := []string{
		"# Valid comment",
		"a.txt  Valid annotation",
		"invalid_line_no_space",
		"b.txt", // This line has no space/annotation
		"c.txt  Valid annotation",
	}
	content := strings.Join(lines, "\n")
	reader := strings.NewReader(content)
	annotations, err := info.ParseWithLogger(reader, "/test/.info", logger)
	require.NoError(t, err)

	// Should only have 2 valid annotations
	require.Len(t, annotations, 2)
	assert.Equal(t, "a.txt", annotations[0].Path)
	assert.Equal(t, "c.txt", annotations[1].Path)

	// Should have warnings about invalid lines
	messages := logger.GetMessages()
	require.Len(t, messages, 2)                        // Expect 2 messages now
	assert.Contains(t, messages[0], "ignoring line 3") // invalid_line_no_space
	assert.Contains(t, messages[0], "no annotation found")
	assert.Contains(t, messages[1], "ignoring line 4") // b.txt line with no space
	assert.Contains(t, messages[1], "no annotation found")
}

func TestCollectorWithLogger(t *testing.T) {
	logger := &TestLogger{}
	fsys := testutil.NewTestFS()
	fsys.MustCreateTree(".", map[string]interface{}{
		".info": `
a.txt  ann for a
a.txt  duplicate (should warn)
invalid_line
`,
		"sub": map[string]interface{}{
			".info": `
../..  invalid ancestor annotation
../a.txt  valid annotation
`,
		},
	})

	collector := info.NewCollectorWithLogger(logger)
	annotations, err := collector.CollectAnnotations(fsys, ".")
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

func TestCollectorFileErrors(t *testing.T) {
	logger := &TestLogger{}

	// Create a custom filesystem that will simulate file open errors
	fsys := &ErrorFS{
		Fs:       testutil.NewTestFS(),
		failPath: "subdir/.info",
	}

	// Create the directory structure first
	err := fsys.MkdirAll("subdir", 0755)
	require.NoError(t, err)
	file, err := fsys.Create("subdir/.info")
	require.NoError(t, err)
	_, err = file.WriteString("a.txt  annotation")
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	collector := info.NewCollectorWithLogger(logger)
	annotations, err := collector.CollectAnnotations(fsys, ".")
	require.NoError(t, err)

	// Should return empty annotations and log the error
	require.Empty(t, annotations)

	// Should have logged file open error
	messages := logger.GetMessages()
	require.GreaterOrEqual(t, len(messages), 1) // At least one error message

	// Check that at least one message contains the expected error
	hasFileError := false
	for _, msg := range messages {
		if strings.Contains(msg, "cannot open .info file") && strings.Contains(msg, "subdir/.info") {
			hasFileError = true
			break
		}
	}
	assert.True(t, hasFileError, "Should have file open error message")
}

// ErrorFS is a wrapper around afero.Fs that simulates file open errors for specific paths
type ErrorFS struct {
	afero.Fs
	failPath string
}

func (efs *ErrorFS) Open(name string) (afero.File, error) {
	if name == efs.failPath {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}
	return efs.Fs.Open(name)
}
