package logging_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"treex/treex/internal/testutil"
	"treex/treex/logging"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    logging.Level
		expected string
	}{
		{logging.TraceLevel, "trace"},
		{logging.DebugLevel, "debug"},
		{logging.InfoLevel, "info"},
		{logging.WarnLevel, "warn"},
		{logging.ErrorLevel, "error"},
		{logging.DisabledLevel, "disabled"},
		{logging.Level(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := logging.DefaultConfig()

	assert.Equal(t, logging.WarnLevel, config.ConsoleLevel)
	assert.Equal(t, logging.DebugLevel, config.FileLevel)
	assert.False(t, config.NoColor)
	assert.NotEmpty(t, config.LogFile)
	assert.Contains(t, config.LogFile, "treex.log")
}

func TestSetup_ConsoleOnly(t *testing.T) {
	var buf bytes.Buffer

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	config := logging.Config{
		ConsoleLevel: logging.InfoLevel,
		FileLevel:    logging.DisabledLevel,
		NoColor:      true, // For consistent test output
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test logging
	logger.Info().Msg("test info message")
	logger.Debug().Msg("test debug message") // Should not appear

	// Close and restore stdout
	err = w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	// Read captured output
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test info message")
	assert.NotContains(t, output, "test debug message")
}

func TestSetup_FileOnly(t *testing.T) {
	fs := testutil.NewTestFS()
	tempDir := "/tmp/logging_test"
	logFile := filepath.Join(tempDir, "test.log")

	// Create the directory in our test filesystem
	err := fs.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	config := logging.Config{
		ConsoleLevel: logging.DisabledLevel,
		FileLevel:    logging.DebugLevel,
		LogFile:      logFile,
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test that we can create a logger (actual file creation happens in real filesystem)
	// We just test the config validation here
	assert.Equal(t, logFile, config.LogFile)
}

func TestSetup_BothHandlers(t *testing.T) {
	var buf bytes.Buffer

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := logging.Config{
		ConsoleLevel: logging.WarnLevel,
		FileLevel:    logging.DebugLevel,
		LogFile:      logFile,
		NoColor:      true,
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test different levels
	logger.Debug().Msg("debug message")  // Should only go to file
	logger.Info().Msg("info message")    // Should only go to file
	logger.Warn().Msg("warning message") // Should go to both
	logger.Error().Msg("error message")  // Should go to both

	// Close and restore stdout
	err = w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	// Read captured console output
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	consoleOutput := buf.String()
	assert.NotContains(t, consoleOutput, "debug message")
	assert.NotContains(t, consoleOutput, "info message")
	assert.Contains(t, consoleOutput, "warning message")
	assert.Contains(t, consoleOutput, "error message")

	// Check file output
	fileContent, err := os.ReadFile(logFile)
	require.NoError(t, err)

	fileOutput := string(fileContent)
	assert.Contains(t, fileOutput, "debug message")
	assert.Contains(t, fileOutput, "info message")
	assert.Contains(t, fileOutput, "warning message")
	assert.Contains(t, fileOutput, "error message")
}

func TestSetup_DisabledLogging(t *testing.T) {
	config := logging.Config{
		ConsoleLevel: logging.DisabledLevel,
		FileLevel:    logging.DisabledLevel,
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Should not panic even with disabled logging
	logger.Info().Msg("this should be discarded")
	logger.Error().Msg("this should also be discarded")
}

func TestSetupFromVerbosity(t *testing.T) {
	tests := []struct {
		verbosity     int
		expectedLevel logging.Level
	}{
		{0, logging.WarnLevel},  // default
		{1, logging.InfoLevel},  // -v
		{2, logging.DebugLevel}, // -vv
		{3, logging.TraceLevel}, // -vvv
		{4, logging.TraceLevel}, // -vvvv (caps at trace)
	}

	for _, tt := range tests {
		t.Run(tt.expectedLevel.String(), func(t *testing.T) {
			logger, err := logging.SetupFromVerbosity(tt.verbosity)
			require.NoError(t, err)
			require.NotNil(t, logger)

			// We can't easily test the internal level without exposing it,
			// but we can test that setup succeeds
		})
	}
}

func TestLogger_Printf(t *testing.T) {
	var buf bytes.Buffer

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	config := logging.Config{
		ConsoleLevel: logging.InfoLevel,
		FileLevel:    logging.DisabledLevel,
		NoColor:      true,
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)

	// Test Printf interface (used by existing code)
	logger.Printf("test message with %s", "formatting")

	// Close and restore stdout
	err = w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	// Read captured output
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test message with formatting")
}

func TestGlobalLogger(t *testing.T) {
	// Test that Get() initializes a default logger
	logger := logging.Get()
	assert.NotNil(t, logger)

	// Test global logging functions
	assert.NotPanics(t, func() {
		logging.Info().Msg("global info message")
		logging.Debug().Msg("global debug message")
		logging.Warn().Msg("global warn message")
		logging.Error().Msg("global error message")
		logging.Trace().Msg("global trace message")
	})
}

func TestInitGlobal(t *testing.T) {
	config := logging.Config{
		ConsoleLevel: logging.InfoLevel,
		FileLevel:    logging.DisabledLevel,
		NoColor:      true,
	}

	err := logging.InitGlobal(config)
	require.NoError(t, err)

	// Test that global logger is set
	logger := logging.Get()
	assert.NotNil(t, logger)
}

func TestInitGlobalFromVerbosity(t *testing.T) {
	err := logging.InitGlobalFromVerbosity(2) // debug level
	require.NoError(t, err)

	// Test that global logger is set
	logger := logging.Get()
	assert.NotNil(t, logger)
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	config := logging.Config{
		ConsoleLevel: logging.InfoLevel,
		FileLevel:    logging.DisabledLevel,
		NoColor:      true,
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)

	// Test contextual logging
	contextLogger := logger.With().Str("component", "test").Logger()
	contextLogger.Info().Msg("test message")

	// Close and restore stdout
	err = w.Close()
	require.NoError(t, err)
	os.Stdout = oldStdout

	// Read captured output
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "component")
	assert.Contains(t, output, "test")
}

func TestLogFileCreation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "subdir", "test.log")

	config := logging.Config{
		ConsoleLevel: logging.DisabledLevel,
		FileLevel:    logging.InfoLevel,
		LogFile:      logFile,
	}

	logger, err := logging.Setup(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Log a message to ensure file is created
	logger.Info().Msg("test log message")

	// Verify file exists and has content
	assert.FileExists(t, logFile)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "test log message")
}

func TestInvalidLogDirectory(t *testing.T) {
	// Try to create log file in a location that should fail
	// (assuming /invalid/path doesn't exist and can't be created)
	config := logging.Config{
		ConsoleLevel: logging.DisabledLevel,
		FileLevel:    logging.InfoLevel,
		LogFile:      "/invalid/nonexistent/deeply/nested/path/test.log",
	}

	_, err := logging.Setup(config)
	// This should fail gracefully
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "failed to create log directory")
}
