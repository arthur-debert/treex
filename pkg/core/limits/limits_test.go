package limits

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.MaxFileSize != DefaultMaxFileSize {
		t.Errorf("Expected MaxFileSize %d, got %d", DefaultMaxFileSize, config.MaxFileSize)
	}
	
	if config.RuntimeLimit != DefaultRuntimeLimit {
		t.Errorf("Expected RuntimeLimit %v, got %v", DefaultRuntimeLimit, config.RuntimeLimit)
	}
}

func TestLimiter_CheckFileSize(t *testing.T) {
	config := &Config{
		MaxFileSize:  1024 * 1024, // 1MB
		RuntimeLimit: 10 * time.Second,
	}
	
	limiter := NewLimiter(config)
	defer limiter.Close()
	
	tests := []struct {
		name     string
		size     int64
		expected bool
	}{
		{"Small file", 1024, true},
		{"Exact limit", 1024 * 1024, true},
		{"Over limit", 1024*1024 + 1, false},
		{"Large file", 10 * 1024 * 1024, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limiter.CheckFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("CheckFileSize(%d) = %v, want %v", tt.size, result, tt.expected)
			}
		})
	}
	
	// Check that files were counted as skipped
	stats := limiter.Stats()
	if stats.FilesSkipped != 2 {
		t.Errorf("Expected 2 files skipped, got %d", stats.FilesSkipped)
	}
}

func TestLimiter_CheckTimeout(t *testing.T) {
	config := &Config{
		MaxFileSize:  DefaultMaxFileSize,
		RuntimeLimit: 100 * time.Millisecond,
	}
	
	limiter := NewLimiter(config)
	defer limiter.Close()
	
	// Initially should not be timed out
	if !limiter.CheckTimeout() {
		t.Error("Limiter should not be timed out initially")
	}
	
	// Wait for timeout
	time.Sleep(150 * time.Millisecond)
	
	// Now should be timed out
	if limiter.CheckTimeout() {
		t.Error("Limiter should be timed out after waiting")
	}
	
	// Stats should reflect timeout
	stats := limiter.Stats()
	if !stats.TimedOut {
		t.Error("Stats should show timed out")
	}
}

func TestLimiter_RecordFileProcessed(t *testing.T) {
	limiter := NewLimiter(nil) // Use default config
	defer limiter.Close()
	
	// Process some files
	for i := 0; i < 5; i++ {
		limiter.RecordFileProcessed()
	}
	
	stats := limiter.Stats()
	if stats.FilesProcessed != 5 {
		t.Errorf("Expected 5 files processed, got %d", stats.FilesProcessed)
	}
}

func TestProcessingStats_TruncationWarning(t *testing.T) {
	tests := []struct {
		name           string
		stats          ProcessingStats
		estimatedTotal int
		expectedMsg    string
	}{
		{
			name: "No truncation",
			stats: ProcessingStats{
				FilesProcessed: 100,
				FilesSkipped:   0,
				ElapsedTime:    2 * time.Second,
				TimedOut:       false,
			},
			estimatedTotal: 100,
			expectedMsg:    "",
		},
		{
			name: "Timeout truncation",
			stats: ProcessingStats{
				FilesProcessed: 1234,
				FilesSkipped:   0,
				ElapsedTime:    5 * time.Second,
				TimedOut:       true,
			},
			estimatedTotal: 5000,
			expectedMsg:    "Warning: Results truncated due to runtime limit (5s). Processed 1234 of ~5000 files",
		},
		{
			name: "Size limit truncation",
			stats: ProcessingStats{
				FilesProcessed: 100,
				FilesSkipped:   10,
				ElapsedTime:    1 * time.Second,
				TimedOut:       false,
			},
			estimatedTotal: 110,
			expectedMsg:    "Warning: 10 files skipped due to size limit",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.stats.TruncationWarning(tt.estimatedTotal)
			if msg != tt.expectedMsg {
				t.Errorf("TruncationWarning() = %q, want %q", msg, tt.expectedMsg)
			}
		})
	}
}