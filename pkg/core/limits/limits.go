package limits

import (
	"context"
	"fmt"
	"time"
)

// Default performance limits
const (
	// DefaultMaxFileSize is the maximum file size to process (20MB)
	DefaultMaxFileSize = 20 * 1024 * 1024
	
	// DefaultRuntimeLimit is the maximum execution time (5 seconds)
	DefaultRuntimeLimit = 5 * time.Second
)

// Config holds the performance limit configuration
type Config struct {
	MaxFileSize   int64
	RuntimeLimit  time.Duration
}

// DefaultConfig returns the default performance limits configuration
func DefaultConfig() *Config {
	return &Config{
		MaxFileSize:  DefaultMaxFileSize,
		RuntimeLimit: DefaultRuntimeLimit,
	}
}

// Limiter provides methods to check performance limits
type Limiter struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	
	// Stats for reporting
	filesProcessed int
	filesSkipped   int
	startTime      time.Time
}

// NewLimiter creates a new performance limiter
func NewLimiter(config *Config) *Limiter {
	if config == nil {
		config = DefaultConfig()
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), config.RuntimeLimit)
	
	return &Limiter{
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}
}

// CheckFileSize returns true if the file size is within limits
func (l *Limiter) CheckFileSize(size int64) bool {
	if size > l.config.MaxFileSize {
		l.filesSkipped++
		return false
	}
	return true
}

// CheckTimeout returns true if execution can continue (not timed out)
func (l *Limiter) CheckTimeout() bool {
	select {
	case <-l.ctx.Done():
		return false
	default:
		return true
	}
}

// RecordFileProcessed increments the processed file counter
func (l *Limiter) RecordFileProcessed() {
	l.filesProcessed++
}

// Stats returns the current processing statistics
func (l *Limiter) Stats() ProcessingStats {
	return ProcessingStats{
		FilesProcessed: l.filesProcessed,
		FilesSkipped:   l.filesSkipped,
		ElapsedTime:    time.Since(l.startTime),
		TimedOut:       !l.CheckTimeout(),
	}
}

// Close cleans up the limiter resources
func (l *Limiter) Close() {
	l.cancel()
}

// ProcessingStats contains statistics about the processing
type ProcessingStats struct {
	FilesProcessed int
	FilesSkipped   int
	ElapsedTime    time.Duration
	TimedOut       bool
}

// TruncationWarning returns a warning message if results were truncated
func (s ProcessingStats) TruncationWarning(estimatedTotal int) string {
	if s.TimedOut {
		return fmt.Sprintf("Warning: Results truncated due to runtime limit (%s). Processed %d of ~%d files",
			s.ElapsedTime.Round(time.Second), s.FilesProcessed, estimatedTotal)
	}
	if s.FilesSkipped > 0 {
		return fmt.Sprintf("Warning: %d files skipped due to size limit", s.FilesSkipped)
	}
	return ""
}