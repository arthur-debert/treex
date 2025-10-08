// Package logging provides centralized logging infrastructure for treex.
// It supports multiple handlers (console and file) with configurable levels.
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Level represents logging levels
type Level int

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	DisabledLevel
)

// String returns the string representation of the level
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DisabledLevel:
		return "disabled"
	default:
		return "unknown"
	}
}

// toZerolog converts our Level to zerolog.Level
func (l Level) toZerolog() zerolog.Level {
	switch l {
	case TraceLevel:
		return zerolog.TraceLevel
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case DisabledLevel:
		return zerolog.Disabled
	default:
		return zerolog.WarnLevel
	}
}

// Config holds the logging configuration
type Config struct {
	ConsoleLevel Level
	FileLevel    Level
	LogFile      string
	NoColor      bool
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() Config {
	return Config{
		ConsoleLevel: WarnLevel,
		FileLevel:    DebugLevel,
		LogFile:      getDefaultLogFile(),
		NoColor:      false,
	}
}

// Logger wraps zerolog.Logger and provides our interface
type Logger struct {
	logger zerolog.Logger
}

// Printf implements the interface expected by existing code
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Trace logs at trace level
func (l *Logger) Trace() *zerolog.Event {
	return l.logger.Trace()
}

// Debug logs at debug level
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info logs at info level
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn logs at warn level
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error logs at error level
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// With creates a new logger with additional context
func (l *Logger) With() zerolog.Context {
	return l.logger.With()
}

// getDefaultLogFile returns the default log file path using XDG cache directory
func getDefaultLogFile() string {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.TempDir(), "treex", "treex.log")
		}
		cacheDir = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(cacheDir, "treex", "treex.log")
}

// LevelWriter wraps a writer to only accept messages at or above a specified level
type LevelWriter struct {
	Writer io.Writer
	Level  Level
}

func (lw LevelWriter) Write(p []byte) (n int, err error) {
	return lw.Writer.Write(p)
}

func (lw LevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	// Convert zerolog level to our level (note: zerolog levels are reversed)
	var ourLevel Level
	switch level {
	case zerolog.TraceLevel:
		ourLevel = TraceLevel
	case zerolog.DebugLevel:
		ourLevel = DebugLevel
	case zerolog.InfoLevel:
		ourLevel = InfoLevel
	case zerolog.WarnLevel:
		ourLevel = WarnLevel
	case zerolog.ErrorLevel:
		ourLevel = ErrorLevel
	default:
		ourLevel = WarnLevel
	}

	// Only write if the message level is high enough (lower values are more severe)
	if ourLevel >= lw.Level {
		return lw.Writer.Write(p)
	}
	return len(p), nil // Pretend we wrote it to satisfy the interface
}

// Setup initializes the logging infrastructure with the given configuration
func Setup(config Config) (*Logger, error) {
	var writers []zerolog.LevelWriter

	// Console writer (stdout)
	if config.ConsoleLevel != DisabledLevel {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    config.NoColor,
		}
		leveledConsole := LevelWriter{
			Writer: consoleWriter,
			Level:  config.ConsoleLevel,
		}
		writers = append(writers, leveledConsole)
	}

	// File writer
	if config.FileLevel != DisabledLevel && config.LogFile != "" {
		// Ensure log directory exists
		logDir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}

		// Open or create log file
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", config.LogFile, err)
		}

		leveledFile := LevelWriter{
			Writer: file,
			Level:  config.FileLevel,
		}
		writers = append(writers, leveledFile)
	}

	// Create multi-level writer
	var writer zerolog.LevelWriter
	if len(writers) == 0 {
		// If no writers configured, use discard
		writer = LevelWriter{Writer: io.Discard, Level: DisabledLevel}
	} else if len(writers) == 1 {
		writer = writers[0]
	} else {
		// Convert to io.Writer slice for MultiLevelWriter
		ioWriters := make([]io.Writer, len(writers))
		for i, w := range writers {
			ioWriters[i] = w
		}
		writer = zerolog.MultiLevelWriter(ioWriters...)
	}

	// Create logger
	logger := zerolog.New(writer).With().Timestamp().Logger()

	// Set global level to the most verbose level to ensure events reach handlers
	minLevel := config.ConsoleLevel
	if config.FileLevel < minLevel {
		minLevel = config.FileLevel
	}
	zerolog.SetGlobalLevel(minLevel.toZerolog())

	return &Logger{logger: logger}, nil
}

// SetupFromVerbosity configures logging based on verbosity level
// 0 = default (warn console, debug file)
// 1 = info console
// 2 = debug console
// 3 = trace console
func SetupFromVerbosity(verbosity int) (*Logger, error) {
	config := DefaultConfig()

	switch verbosity {
	case 1:
		config.ConsoleLevel = InfoLevel
	case 2:
		config.ConsoleLevel = DebugLevel
	case 3:
		config.ConsoleLevel = TraceLevel
	}

	return Setup(config)
}

// Global logger instance
var globalLogger *Logger

// InitGlobal initializes the global logger
func InitGlobal(config Config) error {
	logger, err := Setup(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// InitGlobalFromVerbosity initializes the global logger from verbosity level
func InitGlobalFromVerbosity(verbosity int) error {
	logger, err := SetupFromVerbosity(verbosity)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Get returns the global logger, initializing it with defaults if needed
func Get() *Logger {
	if globalLogger == nil {
		// Initialize with default config if not set
		logger, err := Setup(DefaultConfig())
		if err != nil {
			// Fallback to basic zerolog
			logger = &Logger{logger: log.Logger}
		}
		globalLogger = logger
	}
	return globalLogger
}

// Trace logs a trace message using the global logger
func Trace() *zerolog.Event {
	return Get().Trace()
}

// Debug logs a debug message using the global logger
func Debug() *zerolog.Event {
	return Get().Debug()
}

// Info logs an info message using the global logger
func Info() *zerolog.Event {
	return Get().Info()
}

// Warn logs a warning message using the global logger
func Warn() *zerolog.Event {
	return Get().Warn()
}

// Error logs an error message using the global logger
func Error() *zerolog.Event {
	return Get().Error()
}
