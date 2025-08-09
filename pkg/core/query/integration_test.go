package query

import (
	"testing"
	"time"
	
	"github.com/adebert/treex/pkg/core/limits"
	"github.com/adebert/treex/pkg/core/types"
)

func TestQueryWithDifferentLimits(t *testing.T) {
	// Initialize the query system
	if err := InitializeQuerySystem(); err != nil {
		t.Fatal(err)
	}
	
	t.Run("file size limit tracking", func(t *testing.T) {
		// The current implementation tracks file processing but doesn't filter based on size
		// in FilterTreeWithLimits. File size filtering happens in readFileContent for text search.
		// This test verifies the limiter correctly tracks file sizes.
		
		config := &limits.Config{
			MaxFileSize:  10 * 1024 * 1024, // 10MB
			RuntimeLimit: 5 * time.Second,
		}
		limiter := limits.NewLimiter(config)
		defer limiter.Close()
		
		// Test file size checking directly
		if !limiter.CheckFileSize(1024) { // 1KB
			t.Error("Should allow small file")
		}
		
		if !limiter.CheckFileSize(10 * 1024 * 1024) { // 10MB (at limit)
			t.Error("Should allow file at limit")
		}
		
		if limiter.CheckFileSize(11 * 1024 * 1024) { // 11MB (over limit)
			t.Error("Should reject file over limit")
		}
		
		stats := limiter.Stats()
		if stats.FilesSkipped != 1 {
			t.Errorf("Expected 1 file skipped, got %d", stats.FilesSkipped)
		}
	})
	
	t.Run("timeout during processing", func(t *testing.T) {
		// Create many nodes to process
		bigRoot := &types.Node{
			Name:         "root",
			Path:         "/root",
			RelativePath: "",
			IsDir:        true,
			Children:     []*types.Node{},
		}
		
		// Add 100 files
		for i := 0; i < 100; i++ {
			bigRoot.Children = append(bigRoot.Children, &types.Node{
				Name:         "file.txt",
				Path:         "/root/file.txt",
				RelativePath: "file.txt",
				IsDir:        false,
			})
		}
		
		// Create a limiter with instant timeout
		config := &limits.Config{
			MaxFileSize:  limits.DefaultMaxFileSize,
			RuntimeLimit: 1 * time.Nanosecond, // Instant timeout
		}
		limiter := limits.NewLimiter(config)
		defer limiter.Close()
		
		// Wait to ensure timeout
		time.Sleep(1 * time.Millisecond)
		
		// Create a query
		matcher := NewMatcher(GetGlobalRegistry(), &Query{})
		
		// Should timeout immediately
		filtered, err := FilterTreeWithLimits(bigRoot, matcher, limiter)
		if err != nil {
			t.Fatal(err)
		}
		
		if filtered != nil {
			t.Error("Expected nil due to immediate timeout")
		}
		
		stats := limiter.Stats()
		if !stats.TimedOut {
			t.Error("Expected timeout")
		}
		
		// Generate warning message
		warning := stats.TruncationWarning(100)
		if warning == "" {
			t.Error("Expected truncation warning")
		}
		if !contains(warning, "runtime limit") {
			t.Errorf("Warning should mention runtime limit: %s", warning)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}