// see docs/dev/architecture.txt - Phase 2: Path Collection
package pathcollection_test

import (
	"strings"
	"testing"

	"treex/treex/internal/testutil"
	"treex/treex/pathcollection"
)

// TestLogger captures log messages for testing
type TestLogger struct {
	messages []string
}

func (tl *TestLogger) Printf(format string, v ...interface{}) {
	// Store formatted message for verification
	tl.messages = append(tl.messages, strings.TrimSpace(sprintf(format, v...)))
}

func (tl *TestLogger) GetMessages() []string {
	return tl.messages
}

// Simple sprintf implementation for testing
func sprintf(format string, v ...interface{}) string {
	// For testing purposes, just return format with basic substitution
	result := format
	for i, arg := range v {
		if i < 2 { // Handle up to 2 args for our test cases
			switch a := arg.(type) {
			case string:
				result = strings.Replace(result, "%q", `"`+a+`"`, 1)
				result = strings.Replace(result, "%v", a, 1)
				result = strings.Replace(result, "%s", a, 1)
			default:
				result = strings.Replace(result, "%v", "error", 1)
			}
		}
	}
	return result
}

func TestLoggingWithCustomLogger(t *testing.T) {
	fs := testutil.NewTestFS()
	logger := &TestLogger{}

	// Create a simple structure that we can collect successfully
	fs.MustCreateTree("/test", map[string]interface{}{
		"normal.txt": "content",
	})

	// Test with custom logger
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root:   "/test",
		Logger: logger,
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to collect some paths")
	}

	// Since we didn't create any permission errors, there should be no log messages
	messages := logger.GetMessages()
	if len(messages) > 0 {
		t.Errorf("Expected no log messages for successful collection, got: %v", messages)
	}
}

func TestLoggingWithDefaultLogger(t *testing.T) {
	fs := testutil.NewTestFS()

	fs.MustCreateTree("/test", map[string]interface{}{
		"file.txt": "content",
	})

	// Test that default logger doesn't crash (uses log.Printf)
	collector := pathcollection.NewCollector(fs, pathcollection.CollectionOptions{
		Root: "/test",
		// No logger specified, should use default log.Printf
	})

	results, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collection with default logger failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to collect some paths")
	}
}

func TestOptionsConfiguratorWithLogger(t *testing.T) {
	fs := testutil.NewTestFS()
	logger := &TestLogger{}

	fs.MustCreateTree("/test", map[string]interface{}{
		"file.txt": "content",
	})

	// Test fluent interface with logger
	results, err := pathcollection.NewConfigurator(fs).
		WithRoot("/test").
		WithLogger(logger).
		Collect()

	if err != nil {
		t.Fatalf("Collection with logger configurator failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected to collect some paths")
	}

	// Verify logger was set (no messages expected for successful collection)
	messages := logger.GetMessages()
	if len(messages) > 0 {
		t.Errorf("Expected no log messages for successful collection, got: %v", messages)
	}
}
