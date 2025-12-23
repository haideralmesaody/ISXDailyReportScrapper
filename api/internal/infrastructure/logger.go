package infrastructure

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/isxcli/isxcli/internal/config"
)

var (
	// globalLogger holds the application-wide logger instance
	globalLogger     *slog.Logger
	globalLoggerOnce sync.Once
	// globalLogFile holds the open log file for cleanup
	globalLogFile *os.File
	// mu protects globalLogFile
	logFileMu sync.Mutex
)

// contextKey is a type for context keys
type contextKey string

const (
	// TraceIDContextKey is the key for storing trace ID in context
	TraceIDContextKey contextKey = "trace_id"
	// RequestIDContextKey is an alias for TraceIDContextKey
	RequestIDContextKey = TraceIDContextKey
)

// InitializeLogger creates and configures the global slog logger instance.
// This should be called once during application startup.
// Per CLAUDE.md: Always use JSON format, always dual output (stdout + file).
func InitializeLogger(cfg config.LoggingConfig) (*slog.Logger, error) {
	var err error
	globalLoggerOnce.Do(func() {
		globalLogger, err = createLogger(cfg)
		if globalLogger != nil {
			slog.SetDefault(globalLogger)
		}
	})
	return globalLogger, err
}

// GetLogger returns the global logger instance.
// If not initialized, returns the default slog logger.
func GetLogger() *slog.Logger {
	if globalLogger == nil {
		return slog.Default()
	}
	return globalLogger
}

// createLogger creates a new slog logger based on configuration
func createLogger(cfg config.LoggingConfig) (*slog.Logger, error) {
	// Parse log level
	level := parseLogLevel(cfg.Level)

	// Create handler options
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}

	var output io.Writer

	// Debug: Print the configuration
	slog.Info("DEBUG: Logging config - Output: %s, FilePath: %s\n", cfg.Output, cfg.FilePath)
	
	// Handle different output modes
	switch strings.ToLower(cfg.Output) {
	case "file":
		file, err := openLogFile(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		globalLogFile = file
		output = file
	case "both":
		file, err := openLogFile(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		globalLogFile = file
		output = io.MultiWriter(os.Stdout, file)
	default:
		output = os.Stdout
	}

	// Per CLAUDE.md: Always use JSON format
	handler := slog.NewJSONHandler(output, opts)

	// Wrap handler to inject trace_id from context
	traceHandlerInstance := &traceHandler{Handler: handler}

	return slog.New(traceHandlerInstance), nil
}

// traceHandler wraps a slog.Handler to automatically inject trace_id from context
type traceHandler struct {
	slog.Handler
}

// Handle adds trace_id to the record if present in context
func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract trace_id from context if present
	if traceID := GetTraceID(ctx); traceID != "" {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new Handler with additional attributes
func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new Handler with the given group name
func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithGroup(name)}
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDContextKey, traceID)
}

// GetTraceID retrieves the trace ID from context
func GetTraceID(ctx context.Context) string {
	// Handle nil context gracefully
	if ctx == nil {
		return ""
	}

	if traceID, ok := ctx.Value(TraceIDContextKey).(string); ok {
		return traceID
	}
	// Also check for the common "X-Request-ID" pattern
	if traceID, ok := ctx.Value("request-id").(string); ok {
		return traceID
	}
	return ""
}

// LoggerFromContext extracts or creates a logger from context.
// This is a helper for components that need context-aware logging.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger := GetLogger()
	
	// If there's a trace ID in context, create a logger with it as an attribute
	if traceID := GetTraceID(ctx); traceID != "" {
		return logger.With("trace_id", traceID)
	}
	
	return logger
}

// MustInitializeLogger is like InitializeLogger but panics on error.
// Use this in main() where errors are fatal.
func MustInitializeLogger(cfg config.LoggingConfig) *slog.Logger {
	logger, err := InitializeLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logger
}

// DefaultConfig returns a default logging configuration that follows CLAUDE.md
func DefaultConfig() config.LoggingConfig {
	return config.LoggingConfig{
		Level:    "info",
		Format:   "json",       // Always JSON per CLAUDE.md
		Output:   "both",       // Both stdout and file
		FilePath: "logs/app.log",
	}
}

// CloseLogFile closes the global log file if open.
// This should be called during graceful shutdown or in tests.
func CloseLogFile() error {
	logFileMu.Lock()
	defer logFileMu.Unlock()
	
	if globalLogFile != nil {
		err := globalLogFile.Close()
		globalLogFile = nil
		return err
	}
	return nil
}

// ResetLoggerForTesting resets the global logger state.
// This should only be called in tests.
func ResetLoggerForTesting() {
	CloseLogFile()
	globalLogger = nil
	globalLoggerOnce = sync.Once{}
}

// openLogFile opens or creates a log file with proper permissions
func openLogFile(filePath string) (*os.File, error) {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}
	
	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}
	
	return file, nil
}