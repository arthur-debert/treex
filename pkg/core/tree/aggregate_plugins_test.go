package tree

import (
	"testing"

	"github.com/adebert/treex/pkg/core/types"
	"github.com/stretchr/testify/assert"
)

func TestAggregatePluginData(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *types.Node
		expected map[string]interface{}
	}{
		{
			name: "aggregates size from visible files",
			setup: func() *types.Node {
				root := &types.Node{
					Name:     "root",
					Path:     "/root",
					IsDir:    true,
					Metadata: make(map[string]interface{}),
					Children: []*types.Node{
						{
							Name:     "file1.txt",
							Path:     "/root/file1.txt",
							IsDir:    false,
							Metadata: map[string]interface{}{"size_bytes": int64(100)},
						},
						{
							Name:     "file2.txt",
							Path:     "/root/file2.txt",
							IsDir:    false,
							Metadata: map[string]interface{}{"size_bytes": int64(200)},
						},
					},
				}
				return root
			},
			expected: map[string]interface{}{
				"size_bytes":        int64(300),
				"size_is_aggregate": true,
			},
		},
		{
			name: "aggregates line counts from visible files",
			setup: func() *types.Node {
				root := &types.Node{
					Name:     "root",
					Path:     "/root",
					IsDir:    true,
					Metadata: make(map[string]interface{}),
					Children: []*types.Node{
						{
							Name:     "file1.go",
							Path:     "/root/file1.go",
							IsDir:    false,
							Metadata: map[string]interface{}{"lc_lines": int64(50)},
						},
						{
							Name:     "file2.go",
							Path:     "/root/file2.go",
							IsDir:    false,
							Metadata: map[string]interface{}{"lc_lines": int64(75)},
						},
					},
				}
				return root
			},
			expected: map[string]interface{}{
				"lc_lines":        int64(125),
				"lc_is_aggregate": true,
			},
		},
		{
			name: "aggregates from nested directories",
			setup: func() *types.Node {
				subdir := &types.Node{
					Name:     "subdir",
					Path:     "/root/subdir",
					IsDir:    true,
					Metadata: make(map[string]interface{}),
					Children: []*types.Node{
						{
							Name:     "nested.txt",
							Path:     "/root/subdir/nested.txt",
							IsDir:    false,
							Metadata: map[string]interface{}{"size_bytes": int64(50)},
						},
					},
				}
				
				root := &types.Node{
					Name:     "root",
					Path:     "/root",
					IsDir:    true,
					Metadata: make(map[string]interface{}),
					Children: []*types.Node{
						{
							Name:     "file.txt",
							Path:     "/root/file.txt",
							IsDir:    false,
							Metadata: map[string]interface{}{"size_bytes": int64(100)},
						},
						subdir,
					},
				}
				
				// Set parent relationships
				for _, child := range root.Children {
					child.Parent = root
				}
				for _, child := range subdir.Children {
					child.Parent = subdir
				}
				
				return root
			},
			expected: map[string]interface{}{
				"size_bytes":        int64(150), // 100 + 50
				"size_is_aggregate": true,
			},
		},
		{
			name: "skips more items indicators",
			setup: func() *types.Node {
				root := &types.Node{
					Name:     "root",
					Path:     "/root",
					IsDir:    true,
					Metadata: make(map[string]interface{}),
					Children: []*types.Node{
						{
							Name:     "file.txt",
							IsDir:    false,
							Path:     "/path/to/file.txt",
							Metadata: map[string]interface{}{"size_bytes": int64(100)},
						},
						{
							Name:     "... 5 more items",
							IsDir:    false,
							Path:     "", // Empty path indicates "more items" node
							Metadata: map[string]interface{}{"size_bytes": int64(999)}, // Should be ignored
						},
					},
				}
				return root
			},
			expected: map[string]interface{}{
				"size_bytes":        int64(100), // Only real file counted
				"size_is_aggregate": true,
			},
		},
		{
			name: "handles mixed metadata",
			setup: func() *types.Node {
				root := &types.Node{
					Name:     "root",
					Path:     "/root",
					IsDir:    true,
					Metadata: make(map[string]interface{}),
					Children: []*types.Node{
						{
							Name:     "code.go",
							Path:     "/root/code.go",
							IsDir:    false,
							Metadata: map[string]interface{}{
								"size_bytes": int64(1000),
								"lc_lines":   int64(50),
							},
						},
						{
							Name:     "data.json",
							Path:     "/root/data.json",
							IsDir:    false,
							Metadata: map[string]interface{}{
								"size_bytes": int64(500),
								// No line count for JSON
							},
						},
						{
							Name:     "script.py",
							Path:     "/root/script.py",
							IsDir:    false,
							Metadata: map[string]interface{}{
								"lc_lines": int64(30),
								// No size for some reason
							},
						},
					},
				}
				return root
			},
			expected: map[string]interface{}{
				"size_bytes":        int64(1500), // 1000 + 500
				"size_is_aggregate": true,
				"lc_lines":          int64(80), // 50 + 30
				"lc_is_aggregate":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.setup()
			t.Logf("Before aggregation: %+v", root.Metadata)
			for i, child := range root.Children {
				t.Logf("Child %d (%s): %+v", i, child.Name, child.Metadata)
			}
			AggregatePluginData(root)
			t.Logf("After aggregation: %+v", root.Metadata)
			
			for key, expectedValue := range tt.expected {
				actualValue, exists := root.Metadata[key]
				assert.True(t, exists, "Expected key %s to exist in metadata", key)
				assert.Equal(t, expectedValue, actualValue, "Key %s has wrong value", key)
			}
		})
	}
}

func TestAggregatePluginDataEmptyDir(t *testing.T) {
	root := &types.Node{
		Name:     "empty",
		IsDir:    true,
		Metadata: make(map[string]interface{}),
		Children: []*types.Node{},
	}
	
	AggregatePluginData(root)
	
	// Should not add any aggregate values for empty directory
	_, hasSize := root.Metadata["size_bytes"]
	assert.False(t, hasSize, "Empty directory should not have size_bytes")
	
	_, hasLines := root.Metadata["lc_lines"]
	assert.False(t, hasLines, "Empty directory should not have lc_lines")
}

func TestAggregatePluginDataNonDirectory(t *testing.T) {
	file := &types.Node{
		Name:     "file.txt",
		IsDir:    false,
		Metadata: map[string]interface{}{"size_bytes": int64(100)},
	}
	
	originalMetadata := make(map[string]interface{})
	for k, v := range file.Metadata {
		originalMetadata[k] = v
	}
	
	AggregatePluginData(file)
	
	// Should not modify file metadata
	assert.Equal(t, originalMetadata, file.Metadata, "File metadata should not be modified")
}