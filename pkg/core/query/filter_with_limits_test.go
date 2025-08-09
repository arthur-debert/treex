package query

import (
	"testing"
	"time"
	
	"github.com/adebert/treex/pkg/core/limits"
	"github.com/adebert/treex/pkg/core/types"
)

func TestFilterTreeWithLimits_Timeout(t *testing.T) {
	// Create a test tree - just a few nodes is enough
	root := &types.Node{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*types.Node{
			{Name: "file1.txt", Path: "/root/file1.txt", IsDir: false},
			{Name: "file2.txt", Path: "/root/file2.txt", IsDir: false},
		},
	}
	
	// Create a limiter with already expired timeout
	config := &limits.Config{
		MaxFileSize:  limits.DefaultMaxFileSize,
		RuntimeLimit: 1 * time.Nanosecond, // Immediately timeout
	}
	limiter := limits.NewLimiter(config)
	defer limiter.Close()
	
	// Wait a tiny bit to ensure timeout
	time.Sleep(1 * time.Millisecond)
	
	// Create a matcher that matches everything
	registry := GetGlobalRegistry()
	matcher := NewMatcher(registry, &Query{})
	
	// Filter should stop immediately due to timeout
	filtered, err := FilterTreeWithLimits(root, matcher, limiter)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Should return nil due to timeout
	if filtered != nil {
		t.Error("Expected nil result due to timeout")
	}
	
	// Check stats
	stats := limiter.Stats()
	if !stats.TimedOut {
		t.Error("Expected timeout in stats")
	}
	
	// Should have processed 0 files since it timed out immediately
	if stats.FilesProcessed > 0 {
		t.Errorf("Expected 0 files processed due to immediate timeout, got %d", stats.FilesProcessed)
	}
}

func TestFilterTreeWithLimits_Normal(t *testing.T) {
	// Create a small test tree
	root := &types.Node{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*types.Node{
			{
				Name:  "file1.txt",
				Path:  "/root/file1.txt",
				IsDir: false,
			},
			{
				Name:  "file2.txt",
				Path:  "/root/file2.txt",
				IsDir: false,
			},
		},
	}
	
	// Create a limiter with normal timeout
	limiter := limits.NewLimiter(nil) // Use defaults
	defer limiter.Close()
	
	// Create a matcher that matches everything
	registry := GetGlobalRegistry()
	matcher := NewMatcher(registry, &Query{})
	
	// Filter should complete normally
	filtered, err := FilterTreeWithLimits(root, matcher, limiter)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if filtered == nil {
		t.Fatal("Expected non-nil result")
	}
	
	// Check stats
	stats := limiter.Stats()
	if stats.TimedOut {
		t.Error("Should not timeout with small tree")
	}
	if stats.FilesProcessed != 2 {
		t.Errorf("Expected 2 files processed, got %d", stats.FilesProcessed)
	}
}

func TestCountTotalFiles(t *testing.T) {
	tests := []struct {
		name     string
		node     *types.Node
		expected int
	}{
		{
			name:     "nil node",
			node:     nil,
			expected: 0,
		},
		{
			name: "single file",
			node: &types.Node{
				Name:  "file.txt",
				IsDir: false,
			},
			expected: 1,
		},
		{
			name: "empty directory",
			node: &types.Node{
				Name:  "dir",
				IsDir: true,
			},
			expected: 0,
		},
		{
			name: "directory with files",
			node: &types.Node{
				Name:  "dir",
				IsDir: true,
				Children: []*types.Node{
					{Name: "file1.txt", IsDir: false},
					{Name: "file2.txt", IsDir: false},
					{Name: "subdir", IsDir: true, Children: []*types.Node{
						{Name: "file3.txt", IsDir: false},
					}},
				},
			},
			expected: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountTotalFiles(tt.node)
			if count != tt.expected {
				t.Errorf("CountTotalFiles() = %d, want %d", count, tt.expected)
			}
		})
	}
}