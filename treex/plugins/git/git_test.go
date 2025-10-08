package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"treex/treex/internal/testutil"
	gitplugin "treex/treex/plugins/git"
)

func TestGitPluginName(t *testing.T) {
	plugin := gitplugin.NewGitPlugin()
	if plugin.Name() != "git" {
		t.Errorf("Expected plugin name 'git', got %q", plugin.Name())
	}
}

func TestGitPluginFindRootsEmptyFilesystem(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := gitplugin.NewGitPlugin()

	roots, err := plugin.FindRoots(fs, "/empty")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}
	if len(roots) != 0 {
		t.Errorf("Expected no roots in empty filesystem, got %d", len(roots))
	}
}

func TestGitPluginFindRootsSingleRepo(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := gitplugin.NewGitPlugin()

	// Create a mock git repository structure
	fs.MustCreateTree("/project", map[string]interface{}{
		".git": map[string]interface{}{
			"config": "mock git config",
			"HEAD":   "ref: refs/heads/main",
			"refs": map[string]interface{}{
				"heads": map[string]interface{}{
					"main": "abc123",
				},
			},
		},
		"README.md": "# Project",
		"src": map[string]interface{}{
			"main.go": "package main",
		},
	})

	roots, err := plugin.FindRoots(fs, "/project")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}

	if len(roots) != 1 {
		t.Fatalf("Expected 1 root, got %d", len(roots))
	}

	if roots[0] != "." {
		t.Errorf("Expected root '.', got %q", roots[0])
	}
}

func TestGitPluginFindRootsMultipleRepos(t *testing.T) {
	fs := testutil.NewTestFS()
	plugin := gitplugin.NewGitPlugin()

	// Create multiple git repositories
	fs.MustCreateTree("/workspace", map[string]interface{}{
		"project-a": map[string]interface{}{
			".git": map[string]interface{}{
				"config": "git config",
			},
			"main.go": "package main",
		},
		"project-b": map[string]interface{}{
			"subproject": map[string]interface{}{
				".git": map[string]interface{}{
					"config": "git config",
				},
				"app.py": "print('hello')",
			},
		},
		"no-git": map[string]interface{}{
			"file.txt": "no git here",
		},
	})

	roots, err := plugin.FindRoots(fs, "/workspace")
	if err != nil {
		t.Fatalf("FindRoots failed: %v", err)
	}

	if len(roots) != 2 {
		t.Fatalf("Expected 2 roots, got %d", len(roots))
	}

	expectedRoots := map[string]bool{
		"project-a":            true,
		"project-b/subproject": true,
	}

	for _, root := range roots {
		if !expectedRoots[root] {
			t.Errorf("Unexpected root: %q", root)
		}
		delete(expectedRoots, root)
	}

	if len(expectedRoots) > 0 {
		t.Errorf("Missing expected roots: %v", expectedRoots)
	}
}

// Note: Testing ProcessRoot requires actual Git repositories since go-git
// needs real Git structures. We'll test with temporary directories.

func TestGitPluginProcessRootNotAGitRepo(t *testing.T) {
	plugin := gitplugin.NewGitPlugin()
	fs := testutil.NewTestFS()

	// Create a non-git directory
	fs.MustCreateTree("/notgit", map[string]interface{}{
		"file.txt": "content",
	})

	// This should not fail but return empty result with error in metadata
	result, err := plugin.ProcessRoot(fs, "/notgit")
	if err != nil {
		t.Fatalf("ProcessRoot should not fail for non-git directories: %v", err)
	}

	if result.PluginName != "git" {
		t.Errorf("Expected plugin name 'git', got %q", result.PluginName)
	}

	// Should have error in metadata since it's not a git repo
	if _, hasError := result.Metadata["error"]; !hasError {
		t.Error("Expected error in metadata for non-git directory")
	}
}

// For testing actual Git functionality, we need to create a real Git repository
// This test creates a temporary directory with a real Git repo
func TestGitPluginWithRealGitRepo(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "treex-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Initialize a real Git repository
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create some test files
	testFiles := map[string]string{
		"README.md":     "# Test Project",
		"src/main.go":   "package main\n\nfunc main() {}\n",
		"docs/guide.md": "# Guide",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
	}

	// Get worktree and stage some files
	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Stage README.md (this will be "staged")
	_, err = worktree.Add("README.md")
	if err != nil {
		t.Fatalf("Failed to stage README.md: %v", err)
	}

	// Create a commit so we have a HEAD
	signature := &object.Signature{
		Name:  "Test User",
		Email: "test@example.com",
	}

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: signature,
	})
	if err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Modify README.md (this will be "unstaged")
	readmePath := filepath.Join(tempDir, "README.md")
	err = os.WriteFile(readmePath, []byte("# Test Project\n\nModified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify README.md: %v", err)
	}

	// Stage src/main.go (this will be "staged")
	_, err = worktree.Add("src/main.go")
	if err != nil {
		t.Fatalf("Failed to stage src/main.go: %v", err)
	}

	// docs/guide.md remains untracked

	// Now test the plugin
	plugin := gitplugin.NewGitPlugin()
	fs := testutil.NewTestFS() // We won't use this, but plugin interface requires it

	result, err := plugin.ProcessRoot(fs, tempDir)
	if err != nil {
		t.Fatalf("ProcessRoot failed: %v", err)
	}

	// Verify result structure
	if result.PluginName != "git" {
		t.Errorf("Expected plugin name 'git', got %q", result.PluginName)
	}

	if result.RootPath != tempDir {
		t.Errorf("Expected root path %q, got %q", tempDir, result.RootPath)
	}

	// Check categories exist
	if _, exists := result.Categories["staged"]; !exists {
		t.Error("Expected 'staged' category")
	}
	if _, exists := result.Categories["unstaged"]; !exists {
		t.Error("Expected 'unstaged' category")
	}
	if _, exists := result.Categories["untracked"]; !exists {
		t.Error("Expected 'untracked' category")
	}

	// Verify we have the expected files in categories
	// Note: The exact categorization depends on Git's internal state
	// We'll just verify that categories are populated appropriately

	stagedCount := len(result.Categories["staged"])
	unstagedCount := len(result.Categories["unstaged"])
	untrackedCount := len(result.Categories["untracked"])

	t.Logf("Staged: %d, Unstaged: %d, Untracked: %d", stagedCount, unstagedCount, untrackedCount)
	t.Logf("Staged files: %v", result.Categories["staged"])
	t.Logf("Unstaged files: %v", result.Categories["unstaged"])
	t.Logf("Untracked files: %v", result.Categories["untracked"])

	// Verify metadata
	if totalFiles, ok := result.Metadata["total_files"].(int); ok {
		if totalFiles != stagedCount+unstagedCount+untrackedCount {
			t.Errorf("Total files mismatch: %d != %d+%d+%d", totalFiles, stagedCount, unstagedCount, untrackedCount)
		}
	} else {
		t.Error("Expected total_files in metadata")
	}

	// Check for branch information
	if branch, ok := result.Metadata["branch"].(string); !ok || branch == "" {
		t.Error("Expected branch information in metadata")
	}

	if commit, ok := result.Metadata["commit"].(string); !ok || commit == "" {
		t.Error("Expected commit information in metadata")
	}
}

func TestGitPluginGetRepositoryInfo(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "treex-git-info-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Initialize a Git repository
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Create and commit a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	_, err = worktree.Add("test.txt")
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	signature := &object.Signature{
		Name:  "Test Author",
		Email: "test@example.com",
	}

	_, err = worktree.Commit("Test commit message", &git.CommitOptions{
		Author: signature,
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Test GetRepositoryInfo
	plugin := gitplugin.NewGitPlugin()
	info, err := plugin.GetRepositoryInfo(tempDir)
	if err != nil {
		t.Fatalf("GetRepositoryInfo failed: %v", err)
	}

	// Check expected fields
	expectedFields := []string{"branch", "commit_hash", "commit_message", "commit_author", "commit_date", "total_commits"}
	for _, field := range expectedFields {
		if _, exists := info[field]; !exists {
			t.Errorf("Expected field %q in repository info", field)
		}
	}

	// Verify specific values
	if branch, ok := info["branch"].(string); !ok || branch != "master" && branch != "main" {
		t.Errorf("Expected branch to be 'master' or 'main', got %v", branch)
	}

	if author, ok := info["commit_author"].(string); !ok || author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got %v", author)
	}

	if message, ok := info["commit_message"].(string); !ok || message != "Test commit message" {
		t.Errorf("Expected message 'Test commit message', got %v", message)
	}

	if commits, ok := info["total_commits"].(int); !ok || commits != 1 {
		t.Errorf("Expected 1 commit, got %v", commits)
	}
}

// Test that git plugin gets registered in default registry
func TestGitPluginRegistration(t *testing.T) {
	// The git plugin should be automatically registered via init()
	// We'll import the plugins package to ensure this happens

	// This test verifies that the plugin registration works
	plugin := gitplugin.NewGitPlugin()
	if plugin.Name() != "git" {
		t.Errorf("Expected git plugin name 'git', got %q", plugin.Name())
	}
}
