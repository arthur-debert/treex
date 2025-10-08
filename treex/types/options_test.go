package types_test

import (
	"testing"

	"github.com/jwaldrip/treex/treex/types"
)

func TestDefaultOptions(t *testing.T) {
	opts := types.DefaultTreeOptions()

	if opts.Root != "." {
		t.Errorf("Expected default root '.', got '%s'", opts.Root)
	}
	if opts.Tree.MaxDepth != 3 {
		t.Errorf("Expected default maxDepth 3, got %d", opts.Tree.MaxDepth)
	}
	if opts.Tree.ShowHidden {
		t.Error("Expected showHidden to be false by default")
	}
	if opts.Tree.DirsOnly {
		t.Error("Expected dirsOnly to be false by default")
	}
	if opts.Patterns.IgnoreFilePath != ".gitignore" {
		t.Errorf("Expected default ignore file '.gitignore', got '%s'", opts.Patterns.IgnoreFilePath)
	}
}

func TestOptionsBuilder(t *testing.T) {
	opts := types.NewOptionsBuilder().
		WithRoot("/project").
		WithMaxDepth(5).
		WithDirsOnly().
		WithHidden().
		WithExcludes("*.tmp", "*.log").
		WithSearch("main", "test").
		Build()

	if opts.Root != "/project" {
		t.Errorf("Expected root '/project', got '%s'", opts.Root)
	}
	if opts.Tree.MaxDepth != 5 {
		t.Errorf("Expected maxDepth 5, got %d", opts.Tree.MaxDepth)
	}
	if !opts.Tree.DirsOnly {
		t.Error("Expected dirsOnly to be true")
	}
	if !opts.Tree.ShowHidden {
		t.Error("Expected showHidden to be true")
	}
	if len(opts.Patterns.Excludes) != 2 {
		t.Errorf("Expected 2 excludes, got %d", len(opts.Patterns.Excludes))
	}
	if len(opts.Search) != 2 {
		t.Errorf("Expected 2 search terms, got %d", len(opts.Search))
	}
}

func TestOptionsValidation(t *testing.T) {
	// Test empty root gets defaulted
	opts := types.TreeOptions{}
	err := opts.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if opts.Root != "." {
		t.Errorf("Expected root to be defaulted to '.', got '%s'", opts.Root)
	}

	// Test negative maxDepth gets corrected
	opts2 := types.TreeOptions{Tree: types.TreeDisplayOptions{MaxDepth: -1}}
	err = opts2.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if opts2.Tree.MaxDepth != 3 {
		t.Errorf("Expected maxDepth to be defaulted to 3, got %d", opts2.Tree.MaxDepth)
	}
}
