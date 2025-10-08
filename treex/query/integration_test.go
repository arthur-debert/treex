package query_test

import (
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
	"treex/treex/query"
	"treex/treex/treeconstruction"
	"treex/treex/types"
)

// TestQueryIntegrationWithPipeline tests the query system integrated with the full pipeline
func TestQueryIntegrationWithPipeline(t *testing.T) {
	fs := testutil.NewTestFS()

	// Create a comprehensive project structure
	fs.MustCreateTree("/project", map[string]interface{}{
		"src": map[string]interface{}{
			"main.go":      "package main",
			"utils.go":     "package utils",
			"test_main.go": "package main // test",
			"helper.py":    "print('helper')",
		},
		"docs": map[string]interface{}{
			"README.md":     "# Project",
			"guide.md":      "# Guide",
			"test_guide.md": "# Test Guide",
		},
		"config": map[string]interface{}{
			"app.yaml":      "config: value",
			"test.yaml":     "test: config",
			"database.json": "{}",
		},
		"scripts": map[string]interface{}{
			"build.sh":  "#!/bin/bash",
			"test.sh":   "#!/bin/bash",
			"deploy.py": "print('deploy')",
		},
		"test": map[string]interface{}{
			"unit": map[string]interface{}{
				"main_test.go":  "package main",
				"utils_test.go": "package utils",
			},
			"integration": map[string]interface{}{
				"api_test.go": "package integration",
				"db_test.go":  "package integration",
			},
		},
	})

	// Phase 1 & 2: Path Collection
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/project",
	})
	pathInfos, err := collector.Collect()
	if err != nil {
		t.Fatalf("Path collection failed: %v", err)
	}

	// Phase 3: Tree Construction
	constructor := treeconstruction.NewConstructor()
	tree := constructor.BuildTree(pathInfos)
	if tree == nil {
		t.Fatal("Tree construction failed")
	}

	// Phase 5: Query Processing - Test different query scenarios
	testCases := []struct {
		name            string
		pathPattern     string
		namePattern     string
		expectedFiles   []string
		minExpectedDirs int
	}{
		{
			name:        "Find all Go files",
			pathPattern: "**/*.go",
			namePattern: "",
			expectedFiles: []string{
				"main.go", "utils.go", "test_main.go",
				"main_test.go", "utils_test.go", "api_test.go", "db_test.go",
			},
			minExpectedDirs: 3, // src, test/unit, test/integration
		},
		{
			name:        "Find test files by name pattern",
			pathPattern: "",
			namePattern: "test*",
			expectedFiles: []string{
				"test_main.go", "test_guide.md", "test.yaml", "test.sh",
			},
			minExpectedDirs: 3, // src, docs, config, scripts
		},
		{
			name:        "Find Go test files (combining patterns)",
			pathPattern: "**/*_test.go",
			namePattern: "*_test.go",
			expectedFiles: []string{
				"main_test.go", "utils_test.go", "api_test.go", "db_test.go",
			},
			minExpectedDirs: 2, // test/unit, test/integration
		},
		{
			name:        "Find files in test directory",
			pathPattern: "**/test/**",
			namePattern: "",
			expectedFiles: []string{
				"main_test.go", "utils_test.go", "api_test.go", "db_test.go",
			},
			minExpectedDirs: 2, // test/unit, test/integration
		},
		{
			name:        "Find markdown files",
			pathPattern: "**/*.md",
			namePattern: "",
			expectedFiles: []string{
				"README.md", "guide.md", "test_guide.md",
			},
			minExpectedDirs: 1, // docs
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build query processor
			builder := query.NewBuilder()
			if tc.pathPattern != "" {
				builder.WithPathPattern(tc.pathPattern)
			}
			if tc.namePattern != "" {
				builder.WithNamePattern(tc.namePattern)
			}
			processor := builder.Build()

			// Apply queries to tree
			filteredTree := processor.Process(tree)
			if filteredTree == nil {
				t.Fatal("Query processing returned nil tree")
			}

			// Extract all files from filtered tree
			actualFiles := extractFileNames(filteredTree)
			actualDirs := countDirectories(filteredTree)

			// Verify expected files are present
			for _, expectedFile := range tc.expectedFiles {
				found := false
				for _, actualFile := range actualFiles {
					if actualFile == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %q not found in results: %v", expectedFile, actualFiles)
				}
			}

			// Verify we don't have unexpected files (at least for specific tests)
			if tc.pathPattern == "**/*.go" && tc.namePattern == "" {
				// For Go files test, ensure no non-Go files
				for _, actualFile := range actualFiles {
					if len(actualFile) < 3 || actualFile[len(actualFile)-3:] != ".go" {
						t.Errorf("Unexpected non-Go file in Go files query: %q", actualFile)
					}
				}
			}

			// Verify minimum number of directories
			if actualDirs < tc.minExpectedDirs {
				t.Errorf("Expected at least %d directories, got %d", tc.minExpectedDirs, actualDirs)
			}

			t.Logf("Query %q found %d files in %d directories", tc.name, len(actualFiles), actualDirs)
		})
	}
}

// TestEmptyQueryWithPipeline tests that empty queries return the complete tree
func TestEmptyQueryWithPipeline(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/simple", map[string]interface{}{
		"file1.txt": "content1",
		"file2.go":  "package main",
		"subdir": map[string]interface{}{
			"file3.md": "# Title",
		},
	})

	// Build tree through pipeline
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/simple",
	})
	pathInfos, err := collector.Collect()
	if err != nil {
		t.Fatalf("Path collection failed: %v", err)
	}

	constructor := treeconstruction.NewConstructor()
	tree := constructor.BuildTree(pathInfos)

	// Process with empty query
	processor := query.NewProcessor()
	result := processor.Process(tree)

	// Should get the same tree back
	if result == nil {
		t.Fatal("Expected non-nil result for empty query")
	}

	originalFiles := extractFileNames(tree)
	resultFiles := extractFileNames(result)

	if len(originalFiles) != len(resultFiles) {
		t.Errorf("Empty query changed file count: original %d, result %d",
			len(originalFiles), len(resultFiles))
	}
}

// TestInvalidPatternsWithPipeline tests handling of invalid patterns
func TestInvalidPatternsWithPipeline(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/test", map[string]interface{}{
		"valid.go": "package main",
	})

	// Build tree
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/test",
	})
	pathInfos, err := collector.Collect()
	if err != nil {
		t.Fatalf("Path collection failed: %v", err)
	}

	constructor := treeconstruction.NewConstructor()
	tree := constructor.BuildTree(pathInfos)

	// Test with invalid pattern
	processor := query.NewBuilder().
		WithPathPattern("[invalid"). // Invalid bracket pattern
		Build()

	result := processor.Process(tree)

	// Should return empty tree (no matches due to invalid pattern)
	if result != nil {
		files := extractFileNames(result)
		if len(files) > 0 {
			t.Errorf("Expected no files for invalid pattern, got: %v", files)
		}
	}
}

// Helper function to extract all file names from a tree
func extractFileNames(node *types.Node) []string {
	if node == nil {
		return nil
	}

	var files []string
	var walk func(*types.Node)
	walk = func(n *types.Node) {
		if n == nil {
			return
		}

		if !n.IsDir {
			files = append(files, n.Name)
		}

		for _, child := range n.Children {
			walk(child)
		}
	}

	walk(node)
	return files
}

// Helper function to count directories in a tree
func countDirectories(node *types.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	var walk func(*types.Node)
	walk = func(n *types.Node) {
		if n == nil {
			return
		}

		if n.IsDir {
			count++
		}

		for _, child := range n.Children {
			walk(child)
		}
	}

	walk(node)
	return count
}
