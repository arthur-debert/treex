// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"testing"

	"github.com/jwaldrip/treex/treex/internal/testutil"
	"github.com/jwaldrip/treex/treex/pathcollection"
	"github.com/jwaldrip/treex/treex/pattern"
)

func TestCombinedFilesOnlyWithDepthAndPatterns(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/complex", map[string]interface{}{
		"root.txt": "root file",
		"root.tmp": "temp file to exclude",
		".hidden":  "hidden file to exclude",
		"level1": map[string]interface{}{
			"file1.go":  "code file",
			"file1.tmp": "temp file to exclude",
			"level2": map[string]interface{}{
				"file2.txt": "text file",
				"level3": map[string]interface{}{
					"file3.py": "should be excluded by depth",
				},
			},
		},
		"node_modules": map[string]interface{}{
			"package.json": "should be excluded by pattern",
		},
	})

	// Combine multiple options: files only + depth limit + pattern exclusions
	filter := pattern.NewFilterBuilder(fs).
		AddUserExcludes([]string{"*.tmp", "node_modules"}).
		AddHiddenFilter(false). // exclude hidden
		Build()

	results, err := pathcollection.NewConfigurator(fs).
		WithRoot("/complex").
		WithFilesOnly().    // Only collect files
		WithMaxDepth(2).    // Max depth 2
		WithFilter(filter). // Apply pattern exclusions
		Collect()

	if err != nil {
		t.Fatalf("Combined options collection failed: %v", err)
	}

	// Extract paths for easier testing
	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
		// Verify all results are files (not directories)
		if result.IsDir {
			t.Errorf("Expected only files, but found directory: %s", result.Path)
		}
	}

	// Should be present: files within depth 2, not excluded by patterns
	expectedPresent := []string{
		"root.txt",        // depth 1, not excluded
		"level1/file1.go", // depth 2, not excluded
	}

	// Should be absent: directories, files beyond depth, excluded files
	expectedAbsent := []string{
		"level1",                        // directory (excluded by FilesOnly)
		"level1/level2",                 // directory (excluded by FilesOnly)
		"level1/level2/level3",          // directory (excluded by FilesOnly)
		"root.tmp",                      // excluded by *.tmp pattern
		"level1/file1.tmp",              // excluded by *.tmp pattern
		".hidden",                       // excluded by hidden filter
		"level1/level2/file2.txt",       // depth 3, excluded by depth limit
		"level1/level2/level3/file3.py", // excluded by depth limit
		"node_modules",                  // excluded by pattern (also directory)
		"node_modules/package.json",     // excluded because parent excluded
	}

	for _, expectedPath := range expectedPresent {
		found := false
		for _, actualPath := range paths {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %q to be present with combined options", expectedPath)
		}
	}

	for _, unexpectedPath := range expectedAbsent {
		for _, actualPath := range paths {
			if actualPath == unexpectedPath {
				t.Errorf("Path %q should be excluded by combined options", unexpectedPath)
			}
		}
	}
}

func TestCombinedDirsOnlyWithPatternsNoDepthLimit(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/project", map[string]interface{}{
		"src": map[string]interface{}{
			"main.go": "code",
			"lib": map[string]interface{}{
				"helper.go": "helper code",
				"deep": map[string]interface{}{
					"nested": map[string]interface{}{
						"util.go": "deep utility",
					},
				},
			},
		},
		"node_modules": map[string]interface{}{
			"package.json": "package",
			"lib": map[string]interface{}{
				"index.js": "library",
			},
		},
		".git": map[string]interface{}{
			"config": "git config",
		},
		"docs": map[string]interface{}{
			"README.md": "documentation",
		},
	})

	// Combine: directories only + pattern exclusions + no depth limit
	filter := pattern.NewFilterBuilder(fs).
		AddUserExcludes([]string{"node_modules", ".git"}).
		Build()

	results, err := pathcollection.NewConfigurator(fs).
		WithRoot("/project").
		WithDirsOnly().     // Only collect directories
		WithFilter(filter). // Apply exclusions, no depth limit
		Collect()

	if err != nil {
		t.Fatalf("Combined dirs only collection failed: %v", err)
	}

	// Verify all results are directories
	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.Path
		if !result.IsDir {
			t.Errorf("Expected only directories, but found file: %s", result.Path)
		}
	}

	// Should include all non-excluded directories at any depth
	expectedPresent := []string{
		".", // root
		"src",
		"src/lib",
		"src/lib/deep",
		"src/lib/deep/nested",
		"docs",
	}

	// Should exclude directories matched by patterns
	expectedAbsent := []string{
		"node_modules",     // excluded by pattern
		"node_modules/lib", // excluded because parent excluded
		".git",             // excluded by pattern
	}

	for _, expectedPath := range expectedPresent {
		found := false
		for _, actualPath := range paths {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected directory %q to be present", expectedPath)
		}
	}

	for _, unexpectedPath := range expectedAbsent {
		for _, actualPath := range paths {
			if actualPath == unexpectedPath {
				t.Errorf("Directory %q should be excluded by patterns", unexpectedPath)
			}
		}
	}
}
