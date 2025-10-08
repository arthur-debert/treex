package testutil

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/jwaldrip/treex/treex/pathcollection"
	"github.com/jwaldrip/treex/treex/pattern"
	"github.com/jwaldrip/treex/treex/treeconstruction"
	"github.com/jwaldrip/treex/treex/types"
)

// TreeBuilder builds a tree from a test filesystem with options
func BuildFileTree(fs *TestFS, optionsMap map[string]interface{}) (*types.Node, error) {
	opts := buildOptionsFromMap(optionsMap)

	// Phase 1: Build pattern filter from options
	filterBuilder := pattern.NewFilterBuilder(fs)
	if len(opts.Patterns.Excludes) > 0 {
		filterBuilder = filterBuilder.AddUserExcludes(opts.Patterns.Excludes)
	}
	if !opts.Tree.ShowHidden {
		filterBuilder = filterBuilder.AddHiddenFilter(false) // exclude hidden files
	}
	filter := filterBuilder.Build()

	// Phase 2: Collect paths with early pruning
	collectionOpts := pathcollection.CollectionOptions{
		Root:     opts.Root,
		MaxDepth: opts.Tree.MaxDepth,
		Filter:   filter,
		DirsOnly: opts.Tree.DirsOnly,
		// Note: FilesOnly would be !DirsOnly if we had that option
	}

	collector := pathcollection.NewCollector(fs, collectionOpts)
	pathInfos, err := collector.Collect()
	if err != nil {
		return nil, fmt.Errorf("path collection failed: %w", err)
	}

	// Phase 3: Build tree from collected paths
	constructor := treeconstruction.NewConstructor()
	root := constructor.BuildTree(pathInfos)

	return root, nil
}

// MustBuildFileTree builds a tree and panics on error (for tests)
func MustBuildFileTree(fs *TestFS, optionsMap map[string]interface{}) *types.Node {
	tree, err := BuildFileTree(fs, optionsMap)
	if err != nil {
		panic(fmt.Sprintf("failed to build tree: %v", err))
	}
	return tree
}

// AssertTreeMatchesMap asserts that a tree structure matches the expected map
func AssertTreeMatchesMap(t *testing.T, tree *types.Node, expectedMap map[string]interface{}) {
	t.Helper()

	if tree == nil {
		t.Fatal("tree is nil")
	}

	// Convert tree to comparable map structure
	actualMap := treeToMap(tree)

	// Compare structures
	if !mapsEqual(actualMap, expectedMap) {
		t.Errorf("Tree structure mismatch:\nExpected: %s\nActual:   %s",
			formatMap(expectedMap), formatMap(actualMap))
	}
}

// buildOptionsFromMap converts a simple map to TreeOptions
func buildOptionsFromMap(optionsMap map[string]interface{}) types.TreeOptions {
	builder := types.NewOptionsBuilder()

	for key, value := range optionsMap {
		switch key {
		case "root":
			if root, ok := value.(string); ok {
				builder = builder.WithRoot(root)
			}
		case "hidden":
			if hidden, ok := value.(bool); ok && hidden {
				builder = builder.WithHidden()
			}
		case "dirs_only":
			if dirsOnly, ok := value.(bool); ok && dirsOnly {
				builder = builder.WithDirsOnly()
			}
		case "max_depth":
			if depth, ok := value.(int); ok {
				builder = builder.WithMaxDepth(depth)
			}
		case "excludes":
			if excludes, ok := value.([]string); ok {
				builder = builder.WithExcludes(excludes...)
			}
		case "search":
			if search, ok := value.([]string); ok {
				builder = builder.WithSearch(search...)
			}
		}
	}

	return builder.Build()
}

// treeToMap converts a tree structure to a map for comparison
func treeToMap(node *types.Node) map[string]interface{} {
	if node == nil {
		return nil
	}

	result := make(map[string]interface{})

	for _, child := range node.Children {
		if child.IsDir {
			// Directory: recurse
			result[child.Name] = treeToMap(child)
		} else {
			// File: just mark as present
			result[child.Name] = "file"
		}
	}

	return result
}

// mapsEqual compares two map structures recursively
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}

		// Both are maps (directories)
		if mapA, okA := valueA.(map[string]interface{}); okA {
			if mapB, okB := valueB.(map[string]interface{}); okB {
				if !mapsEqual(mapA, mapB) {
					return false
				}
			} else {
				return false
			}
		} else {
			// Both should be files (string values)
			if !reflect.DeepEqual(valueA, valueB) {
				return false
			}
		}
	}

	return true
}

// formatMap formats a map for readable test output
func formatMap(m map[string]interface{}) string {
	return formatMapIndented(m, 0)
}

func formatMapIndented(m map[string]interface{}, indent int) string {
	if m == nil {
		return "nil"
	}

	var lines []string
	prefix := strings.Repeat("  ", indent)

	for key, value := range m {
		if subMap, ok := value.(map[string]interface{}); ok {
			lines = append(lines, fmt.Sprintf("%s%s/:", prefix, key))
			subLines := formatMapIndented(subMap, indent+1)
			if subLines != "" {
				lines = append(lines, subLines)
			}
		} else {
			lines = append(lines, fmt.Sprintf("%s%s", prefix, key))
		}
	}

	return strings.Join(lines, "\n")
}

// Common test helper functions

// ExpectedFileMap creates a map representing files (convenience function)
func ExpectedFileMap(files ...string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, file := range files {
		result[file] = "file"
	}
	return result
}

// ExpectedDirMap creates a map with subdirectories
func ExpectedDirMap(dirs map[string]map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for name, contents := range dirs {
		result[name] = contents
	}
	return result
}

// CombineMaps merges multiple maps into one (for mixed files and dirs)
func CombineMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
