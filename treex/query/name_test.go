package query_test

import (
	"testing"

	"treex/treex/query"
	"treex/treex/types"
)

func TestNameQuery_BasicMatching(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "main.go",
			nodeName: "main.go",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "main.go",
			nodeName: "utils.go",
			expected: false,
		},
		{
			name:     "wildcard extension",
			pattern:  "*.go",
			nodeName: "main.go",
			expected: true,
		},
		{
			name:     "wildcard extension no match",
			pattern:  "*.go",
			nodeName: "README.md",
			expected: false,
		},
		{
			name:     "wildcard prefix",
			pattern:  "test*",
			nodeName: "test_main.go",
			expected: true,
		},
		{
			name:     "wildcard prefix no match",
			pattern:  "test*",
			nodeName: "main_test.go",
			expected: false,
		},
		{
			name:     "wildcard suffix",
			pattern:  "*_test.go",
			nodeName: "main_test.go",
			expected: true,
		},
		{
			name:     "wildcard middle",
			pattern:  "test*.go",
			nodeName: "test_utils.go",
			expected: true,
		},
		{
			name:     "multiple wildcards",
			pattern:  "*test*.go",
			nodeName: "unit_test_main.go",
			expected: true,
		},
		{
			name:     "question mark wildcard",
			pattern:  "test?.go",
			nodeName: "test1.go",
			expected: true,
		},
		{
			name:     "question mark no match",
			pattern:  "test?.go",
			nodeName: "test10.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewNameQuery(tt.pattern)
			node := &types.Node{
				Name:  tt.nodeName,
				Path:  "some/path/" + tt.nodeName,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for pattern %q against name %q",
					tt.expected, result, tt.pattern, tt.nodeName)
			}
		})
	}
}

func TestNameQuery_DirectoryNames(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		expected bool
	}{
		{
			name:     "directory exact match",
			pattern:  "src",
			nodeName: "src",
			expected: true,
		},
		{
			name:     "directory wildcard",
			pattern:  "test*",
			nodeName: "test_data",
			expected: true,
		},
		{
			name:     "hidden directory",
			pattern:  ".*",
			nodeName: ".git",
			expected: true,
		},
		{
			name:     "hidden directory no match",
			pattern:  ".*",
			nodeName: "src",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewNameQuery(tt.pattern)
			node := &types.Node{
				Name:  tt.nodeName,
				Path:  "some/path/" + tt.nodeName,
				IsDir: true,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for pattern %q against directory name %q",
					tt.expected, result, tt.pattern, tt.nodeName)
			}
		})
	}
}

func TestNameQuery_BraceExpansion(t *testing.T) {
	// Test doublestar's support for brace expansion
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		expected bool
	}{
		{
			name:     "brace alternatives match first",
			pattern:  "*.{go,js}",
			nodeName: "main.go",
			expected: true,
		},
		{
			name:     "brace alternatives match second",
			pattern:  "*.{go,js}",
			nodeName: "app.js",
			expected: true,
		},
		{
			name:     "brace alternatives no match",
			pattern:  "*.{go,js}",
			nodeName: "README.md",
			expected: false,
		},
		{
			name:     "multiple brace groups",
			pattern:  "{test,spec}*.{go,js}",
			nodeName: "test_utils.go",
			expected: true,
		},
		{
			name:     "multiple brace groups match spec",
			pattern:  "{test,spec}*.{go,js}",
			nodeName: "spec_helper.js",
			expected: true,
		},
		{
			name:     "multiple brace groups no match",
			pattern:  "{test,spec}*.{go,js}",
			nodeName: "main.py",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewNameQuery(tt.pattern)
			node := &types.Node{
				Name:  tt.nodeName,
				Path:  "path/" + tt.nodeName,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for brace pattern %q against name %q",
					tt.expected, result, tt.pattern, tt.nodeName)
			}
		})
	}
}

func TestNameQuery_CharacterClasses(t *testing.T) {
	// Test character class patterns
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		expected bool
	}{
		{
			name:     "digit character class",
			pattern:  "test[0-9].go",
			nodeName: "test1.go",
			expected: true,
		},
		{
			name:     "digit character class no match",
			pattern:  "test[0-9].go",
			nodeName: "testa.go",
			expected: false,
		},
		{
			name:     "letter character class",
			pattern:  "test[a-z].go",
			nodeName: "testx.go",
			expected: true,
		},
		{
			name:     "mixed character class",
			pattern:  "test[a-z0-9].go",
			nodeName: "test5.go",
			expected: true,
		},
		{
			name:     "negated character class",
			pattern:  "test[!0-9].go",
			nodeName: "testa.go",
			expected: true,
		},
		{
			name:     "negated character class no match",
			pattern:  "test[!0-9].go",
			nodeName: "test1.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewNameQuery(tt.pattern)
			node := &types.Node{
				Name:  tt.nodeName,
				Path:  "path/" + tt.nodeName,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for character class pattern %q against name %q",
					tt.expected, result, tt.pattern, tt.nodeName)
			}
		})
	}
}

func TestNameQuery_InvalidPattern(t *testing.T) {
	// Test with an invalid glob pattern
	query := query.NewNameQuery("[invalid")
	node := &types.Node{
		Name:  "test.go",
		Path:  "path/test.go",
		IsDir: false,
	}

	// Should return false for invalid patterns
	result := query.Match(node)
	if result != false {
		t.Error("Expected false for invalid pattern")
	}
}

func TestNameQuery_NilNode(t *testing.T) {
	query := query.NewNameQuery("*.go")
	result := query.Match(nil)
	if result != false {
		t.Error("Expected false for nil node")
	}
}

func TestNameQuery_Name(t *testing.T) {
	pattern := "test*.go"
	query := query.NewNameQuery(pattern)
	expected := "name:" + pattern

	if query.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, query.Name())
	}
}

func TestNameQuery_CaseSensitive(t *testing.T) {
	// Test that matching is case-sensitive (default doublestar behavior)
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		expected bool
	}{
		{
			name:     "exact case match",
			pattern:  "Main.go",
			nodeName: "Main.go",
			expected: true,
		},
		{
			name:     "case mismatch",
			pattern:  "main.go",
			nodeName: "Main.go",
			expected: false,
		},
		{
			name:     "wildcard case sensitive",
			pattern:  "*.GO",
			nodeName: "main.go",
			expected: false,
		},
		{
			name:     "wildcard correct case",
			pattern:  "*.GO",
			nodeName: "MAIN.GO",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewNameQuery(tt.pattern)
			node := &types.Node{
				Name:  tt.nodeName,
				Path:  "path/" + tt.nodeName,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for case-sensitive pattern %q against name %q",
					tt.expected, result, tt.pattern, tt.nodeName)
			}
		})
	}
}

func TestNameQuery_CommonUseCase(t *testing.T) {
	// Test common use cases for name queries
	tests := []struct {
		name     string
		pattern  string
		nodeName string
		expected bool
	}{
		{
			name:     "find all go files",
			pattern:  "*.go",
			nodeName: "main.go",
			expected: true,
		},
		{
			name:     "find test files",
			pattern:  "*_test.go",
			nodeName: "utils_test.go",
			expected: true,
		},
		{
			name:     "find config files",
			pattern:  "config.*",
			nodeName: "config.yaml",
			expected: true,
		},
		{
			name:     "find hidden files",
			pattern:  ".*",
			nodeName: ".gitignore",
			expected: true,
		},
		{
			name:     "find readme files",
			pattern:  "[Rr][Ee][Aa][Dd][Mm][Ee]*",
			nodeName: "README.md",
			expected: true,
		},
		{
			name:     "find readme files lowercase",
			pattern:  "[Rr][Ee][Aa][Dd][Mm][Ee]*",
			nodeName: "readme.txt",
			expected: true,
		},
		{
			name:     "find temp files",
			pattern:  "*.{tmp,temp,bak}",
			nodeName: "backup.bak",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := query.NewNameQuery(tt.pattern)
			node := &types.Node{
				Name:  tt.nodeName,
				Path:  "path/" + tt.nodeName,
				IsDir: false,
			}

			result := query.Match(node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for use case %q (pattern %q against name %q)",
					tt.expected, result, tt.name, tt.pattern, tt.nodeName)
			}
		})
	}
}