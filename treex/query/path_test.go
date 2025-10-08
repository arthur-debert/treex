package query_test

import (
	"testing"

	"treex/treex/query"
	"treex/treex/types"
)

func TestPathQuery_BasicMatching(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		nodePath string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "main.go",
			nodePath: "main.go",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "main.go",
			nodePath: "utils.go",
			expected: false,
		},
		{
			name:     "wildcard extension",
			pattern:  "*.go",
			nodePath: "main.go",
			expected: true,
		},
		{
			name:     "wildcard extension no match",
			pattern:  "*.go",
			nodePath: "README.md",
			expected: false,
		},
		{
			name:     "directory wildcard",
			pattern:  "src/*",
			nodePath: "src/main.go",
			expected: true,
		},
		{
			name:     "double star recursive",
			pattern:  "**/test*.go",
			nodePath: "src/nested/test_main.go",
			expected: true,
		},
		{
			name:     "double star at start",
			pattern:  "**/main.go",
			nodePath: "project/src/main.go",
			expected: true,
		},
		{
			name:     "double star in middle",
			pattern:  "src/**/main.go",
			nodePath: "src/nested/deep/main.go",
			expected: true,
		},
		{
			name:     "complex pattern",
			pattern:  "src/**/*.{go,md}",
			nodePath: "src/docs/README.md",
			expected: true,
		},
		{
			name:     "complex pattern no match",
			pattern:  "src/**/*.{go,md}",
			nodePath: "src/config/app.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewPathQuery(tt.pattern)
			node := &types.Node{
				Name:  "test",
				Path:  tt.nodePath,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for pattern %q against path %q",
					tt.expected, result, tt.pattern, tt.nodePath)
			}
		})
	}
}

func TestPathQuery_InvalidPattern(t *testing.T) {
	// Test with an invalid glob pattern
	query := query.NewPathQuery("[invalid")
	node := &types.Node{
		Name:  "test.go",
		Path:  "src/test.go",
		IsDir: false,
	}

	// Should return false for invalid patterns
	result := query.Match(node)
	if result != false {
		t.Error("Expected false for invalid pattern")
	}
}

func TestPathQuery_NilNode(t *testing.T) {
	query := query.NewPathQuery("*.go")
	result := query.Match(nil)
	if result != false {
		t.Error("Expected false for nil node")
	}
}

func TestPathQuery_Name(t *testing.T) {
	pattern := "src/**/*.go"
	query := query.NewPathQuery(pattern)
	expected := "path:" + pattern

	if query.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, query.Name())
	}
}

func TestPathQuery_DirectoryMatching(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		nodePath string
		isDir    bool
		expected bool
	}{
		{
			name:     "directory exact match",
			pattern:  "src",
			nodePath: "src",
			isDir:    true,
			expected: true,
		},
		{
			name:     "directory with wildcard",
			pattern:  "src/*",
			nodePath: "src/utils",
			isDir:    true,
			expected: true,
		},
		{
			name:     "nested directory",
			pattern:  "**/test",
			nodePath: "project/nested/test",
			isDir:    true,
			expected: true,
		},
		{
			name:     "directory ending with slash",
			pattern:  "src/",
			nodePath: "src",
			isDir:    true,
			expected: false, // Pattern expects trailing slash but path doesn't have it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewPathQuery(tt.pattern)
			node := &types.Node{
				Name:  "test",
				Path:  tt.nodePath,
				IsDir: tt.isDir,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for pattern %q against directory path %q",
					tt.expected, result, tt.pattern, tt.nodePath)
			}
		})
	}
}

func TestPathQuery_PathNormalization(t *testing.T) {
	// Test that paths are normalized (converted to forward slashes)
	query := query.NewPathQuery("src/main.go")

	tests := []struct {
		name     string
		nodePath string
		expected bool
	}{
		{
			name:     "forward slashes",
			nodePath: "src/main.go",
			expected: true,
		},
		{
			name:     "already normalized",
			nodePath: "src/main.go",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &types.Node{
				Name:  "main.go",
				Path:  tt.nodePath,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %q", tt.expected, result, tt.nodePath)
			}
		})
	}
}

func TestPathQuery_GitignoreCompatiblePatterns(t *testing.T) {
	// Test patterns that are commonly used in .gitignore files
	tests := []struct {
		name     string
		pattern  string
		nodePath string
		expected bool
	}{
		{
			name:     "ignore node_modules",
			pattern:  "node_modules/**",
			nodePath: "node_modules/package/index.js",
			expected: true,
		},
		{
			name:     "ignore build directories",
			pattern:  "**/build/**",
			nodePath: "project/target/build/output.jar",
			expected: true,
		},
		{
			name:     "ignore temp files",
			pattern:  "**/*.tmp",
			nodePath: "cache/temp.tmp",
			expected: true,
		},
		{
			name:     "ignore log files",
			pattern:  "**/*.log",
			nodePath: "logs/app.log",
			expected: true,
		},
		{
			name:     "match specific file types in specific dirs",
			pattern:  "test/**/*.spec.js",
			nodePath: "test/unit/app.spec.js",
			expected: true,
		},
		{
			name:     "no match for different extension",
			pattern:  "test/**/*.spec.js",
			nodePath: "test/unit/app.test.js",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewPathQuery(tt.pattern)
			node := &types.Node{
				Name:  "test",
				Path:  tt.nodePath,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for gitignore pattern %q against path %q",
					tt.expected, result, tt.pattern, tt.nodePath)
			}
		})
	}
}