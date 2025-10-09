// Package git provides a plugin for Git status-based file filtering
package git

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jwaldrip/treex/treex/plugins"
	"github.com/jwaldrip/treex/treex/types"
	"github.com/spf13/afero"
)

// GitPlugin categorizes files based on their Git working tree status
// It finds Git repositories and categorizes files as staged, unstaged, or untracked
type GitPlugin struct{}

// NewGitPlugin creates a new Git plugin instance
func NewGitPlugin() *GitPlugin {
	return &GitPlugin{}
}

// Name returns the plugin identifier
func (p *GitPlugin) Name() string {
	return "git"
}

// FindRoots discovers Git repositories by looking for .git directories
// Returns the parent directories of .git folders as Git repository roots
func (p *GitPlugin) FindRoots(fs afero.Fs, searchRoot string) ([]string, error) {
	var roots []string

	// Walk the filesystem looking for .git directories
	err := afero.Walk(fs, searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors, don't fail the entire search
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			// The root is the parent directory containing the .git folder
			gitRepo := filepath.Dir(path)

			// Convert to relative path from search root
			relativeRoot, err := filepath.Rel(searchRoot, gitRepo)
			if err != nil {
				return nil // Skip this root if we can't make it relative
			}

			// Normalize "." for current directory
			if relativeRoot == "." || relativeRoot == "" {
				relativeRoot = "."
			}

			roots = append(roots, relativeRoot)
		}

		return nil
	})

	return roots, err
}

// ProcessRoot analyzes Git status in a repository root and categorizes files
// Uses go-git library to determine file status: staged, unstaged, untracked
func (p *GitPlugin) ProcessRoot(fs afero.Fs, rootPath string) (*plugins.Result, error) {
	result := &plugins.Result{
		PluginName: p.Name(),
		RootPath:   rootPath,
		Categories: make(map[string][]string),
		Metadata:   make(map[string]interface{}),
		Cache:      make(map[string]interface{}),
	}

	// Open the Git repository using go-git
	repo, err := git.PlainOpenWithOptions(rootPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		// If we can't open as Git repo, return empty result (not an error)
		// This handles cases where .git exists but repo is corrupted/invalid
		result.Metadata["error"] = "failed to open git repository: " + err.Error()
		return result, nil
	}

	// Get the working tree to analyze file status
	worktree, err := repo.Worktree()
	if err != nil {
		result.Metadata["error"] = "failed to get git worktree: " + err.Error()
		return result, nil
	}

	// Get the current status of all files in the repository
	status, err := worktree.Status()
	if err != nil {
		result.Metadata["error"] = "failed to get git status: " + err.Error()
		return result, nil
	}

	// Initialize categories
	result.Categories["staged"] = make([]string, 0)
	result.Categories["unstaged"] = make([]string, 0)
	result.Categories["untracked"] = make([]string, 0)

	// Process each file status entry
	stagedCount := 0
	unstagedCount := 0
	untrackedCount := 0

	// Store git status data for caching - map of normalized path to status info
	gitStatusData := make(map[string]types.GitStatus)

	for filePath, fileStatus := range status {
		// Normalize path separators for consistency
		normalizedPath := filepath.ToSlash(filePath)

		// Categorize based on Git status
		// go-git status codes:
		// Staging: ' ', 'M', 'A', 'D', 'R', 'C'
		// Worktree: ' ', 'M', 'A', 'D', 'R', 'C', '?', '!'

		staging := fileStatus.Staging
		worktree := fileStatus.Worktree

		// Create status object for this file
		gitStatus := types.GitStatus{
			Path:      normalizedPath,
			Staged:    staging != git.Unmodified && staging != git.Untracked,
			Unstaged:  worktree != git.Unmodified && worktree != git.Untracked,
			Untracked: worktree == git.Untracked,
		}

		// Set human-readable status description
		if gitStatus.Untracked {
			gitStatus.Status = "untracked"
		} else if gitStatus.Staged && gitStatus.Unstaged {
			gitStatus.Status = "staged+unstaged"
		} else if gitStatus.Staged {
			gitStatus.Status = "staged"
		} else if gitStatus.Unstaged {
			gitStatus.Status = "unstaged"
		} else {
			gitStatus.Status = "clean"
		}

		// Store in cache for data enrichment
		gitStatusData[normalizedPath] = gitStatus

		// Check staging area status
		if gitStatus.Staged {
			// File has changes in staging area (staged for commit)
			result.Categories["staged"] = append(result.Categories["staged"], normalizedPath)
			stagedCount++
		}

		// Check working tree status
		if gitStatus.Untracked {
			// File is untracked
			result.Categories["untracked"] = append(result.Categories["untracked"], normalizedPath)
			untrackedCount++
		} else if gitStatus.Unstaged {
			// File has modifications in working tree (unstaged changes)
			result.Categories["unstaged"] = append(result.Categories["unstaged"], normalizedPath)
			unstagedCount++
		}
	}

	// Store git status data in cache for efficient data enrichment
	result.Cache["git_status"] = gitStatusData

	// Add repository metadata
	result.Metadata["staged_count"] = stagedCount
	result.Metadata["unstaged_count"] = unstagedCount
	result.Metadata["untracked_count"] = untrackedCount
	result.Metadata["total_files"] = stagedCount + unstagedCount + untrackedCount

	// Add Git repository information
	if head, err := repo.Head(); err == nil {
		result.Metadata["branch"] = head.Name().Short()
		result.Metadata["commit"] = head.Hash().String()[:8] // Short commit hash
	}

	// Add remote information if available
	if remotes, err := repo.Remotes(); err == nil && len(remotes) > 0 {
		// Get origin remote if it exists
		for _, remote := range remotes {
			if remote.Config().Name == "origin" {
				if len(remote.Config().URLs) > 0 {
					result.Metadata["remote_origin"] = remote.Config().URLs[0]
				}
				break
			}
		}
	}

	return result, nil
}

// GetRepositoryInfo extracts additional repository information
// This is a helper method for getting more detailed Git metadata
func (p *GitPlugin) GetRepositoryInfo(repoPath string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return info, err
	}

	// Get HEAD reference
	head, err := repo.Head()
	if err != nil {
		return info, err
	}

	info["branch"] = head.Name().Short()
	info["commit_hash"] = head.Hash().String()

	// Get latest commit information
	commit, err := repo.CommitObject(head.Hash())
	if err == nil {
		info["commit_message"] = strings.Split(commit.Message, "\n")[0] // First line only
		info["commit_author"] = commit.Author.Name
		info["commit_date"] = commit.Author.When.Format("2006-01-02 15:04:05")
	}

	// Count total commits
	commits, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err == nil {
		commitCount := 0
		err = commits.ForEach(func(*object.Commit) error {
			commitCount++
			return nil
		})
		if err == nil {
			info["total_commits"] = commitCount
		}
	}

	return info, nil
}

// GetCategories returns the filter categories provided by the git plugin
// Implements FilterPlugin interface
func (p *GitPlugin) GetCategories() []plugins.FilterPluginCategory {
	return []plugins.FilterPluginCategory{
		{
			Name:        "staged",
			Description: "Files staged for commit in git index",
		},
		{
			Name:        "unstaged",
			Description: "Files with unstaged changes in git working tree",
		},
		{
			Name:        "untracked",
			Description: "Files not tracked by git",
		},
	}
}

// EnrichNode attaches git status data to nodes
// Implements DataPlugin interface
func (p *GitPlugin) EnrichNode(fs afero.Fs, node *types.Node) error {
	// Get the directory containing this file to find the git repository
	nodeDir := filepath.Dir(node.Path)
	if node.IsDir {
		nodeDir = node.Path
	}

	// Find the git repository root by walking up the directory tree
	gitRoot := p.findGitRoot(fs, nodeDir)
	if gitRoot == "" {
		// Try current directory as a fallback
		gitRoot = "."
	}

	// Open the Git repository
	repo, err := git.PlainOpenWithOptions(gitRoot, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		// If we can't open git repo, skip enrichment (not an error)
		return nil
	}

	// Get the working tree to analyze file status
	worktree, err := repo.Worktree()
	if err != nil {
		return nil
	}

	// Get the current status of all files in the repository
	status, err := worktree.Status()
	if err != nil {
		return nil
	}

	// Convert node path to relative path from git root for status lookup
	nodePath := node.Path
	if gitRoot != "." {
		if rel, err := filepath.Rel(gitRoot, node.Path); err == nil {
			nodePath = rel
		}
	}
	normalizedNodePath := filepath.ToSlash(nodePath)

	// Look for git status for this specific file
	if fileStatus, exists := status[normalizedNodePath]; exists {
		// Create and attach git status data
		gitStatus := &types.GitStatus{
			Path:      normalizedNodePath,
			Staged:    fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked,
			Unstaged:  fileStatus.Worktree != git.Unmodified && fileStatus.Worktree != git.Untracked,
			Untracked: fileStatus.Worktree == git.Untracked,
		}

		// Set human-readable status description
		if gitStatus.Untracked {
			gitStatus.Status = "untracked"
		} else if gitStatus.Staged && gitStatus.Unstaged {
			gitStatus.Status = "staged+unstaged"
		} else if gitStatus.Staged {
			gitStatus.Status = "staged"
		} else if gitStatus.Unstaged {
			gitStatus.Status = "unstaged"
		} else {
			gitStatus.Status = "clean"
		}

		node.SetPluginData("git", gitStatus)
	}

	return nil
}

// EnrichNodeWithCache attaches git status data using cached results from filtering phase
// Implements CachedDataPlugin interface for efficient data enrichment
func (p *GitPlugin) EnrichNodeWithCache(fs afero.Fs, node *types.Node, pluginResults []*plugins.Result) error {
	// Look through all plugin results to find cached git status data
	for _, result := range pluginResults {
		if result.PluginName != p.Name() {
			continue
		}

		// Check if we have cached git status data for this result
		cachedGitStatus, exists := result.Cache["git_status"]
		if !exists {
			continue
		}

		// Type assert to get the git status map
		gitStatusMap, ok := cachedGitStatus.(map[string]types.GitStatus)
		if !ok {
			continue
		}

		// Convert node path to match cached data format
		nodePath := node.Path
		if result.RootPath != "." {
			if rel, err := filepath.Rel(result.RootPath, node.Path); err == nil && !strings.HasPrefix(rel, "..") {
				nodePath = rel
			} else {
				// If we can't make it relative, use basename for comparison
				nodePath = filepath.Base(node.Path)
			}
		}
		normalizedNodePath := filepath.ToSlash(nodePath)

		// Look for git status for this specific file
		if gitStatus, exists := gitStatusMap[normalizedNodePath]; exists {
			// Create a copy of the status and attach to node
			nodeGitStatus := &types.GitStatus{
				Path:      gitStatus.Path,
				Staged:    gitStatus.Staged,
				Unstaged:  gitStatus.Unstaged,
				Untracked: gitStatus.Untracked,
				Status:    gitStatus.Status,
			}
			node.SetPluginData("git", nodeGitStatus)
			return nil
		}
	}

	// No cached git status found - this is normal for files outside git repos
	return nil
}

// findGitRoot finds the git repository root for a given path
func (p *GitPlugin) findGitRoot(fs afero.Fs, startPath string) string {
	currentPath := startPath
	for {
		// Check if .git exists in current directory
		gitPath := filepath.Join(currentPath, ".git")
		if exists, err := afero.Exists(fs, gitPath); err == nil && exists {
			return currentPath
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory without finding .git
			break
		}
		currentPath = parentPath
	}
	return ""
}

// init registers the git plugin with the default registry
func init() {
	if err := plugins.RegisterPlugin(NewGitPlugin()); err != nil {
		panic("failed to register git plugin: " + err.Error())
	}
}
