package license

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel

	// Legacy constants for backward compatibility
	LogLevelDebug = DebugLevel
	LogLevelInfo  = InfoLevel
	LogLevelWarn  = WarnLevel
	LogLevelError = ErrorLevel
	LogLevelFatal = FatalLevel
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Level      LogLevel
	Action     string
	Result     string
	LicenseKey string
	UserEmail  string
	MachineID  string
	IPAddress  string // For security logging
	Duration   time.Duration
	Error      error
	Details    map[string]interface{} // For additional security details
	Metadata   map[string]interface{}
}

// Logger wraps slog for backward compatibility during migration
type Logger struct {
	slog *slog.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) (*Logger, error) {
	// Note: slog level configuration is handled by infrastructure logger
	// This function provides backward compatibility

	// Use the default slog logger with appropriate level
	return &Logger{
		slog: slog.Default().With("component", "license"),
	}, nil
}

// Log logs a structured entry
func (l *Logger) Log(entry LogEntry) {
	if l.slog == nil {
		return
	}

	// Build attributes
	attrs := []slog.Attr{
		slog.String("action", entry.Action),
		slog.String("result", entry.Result),
	}

	if entry.LicenseKey != "" {
		// Mask license key for security
		masked := "****"
		if len(entry.LicenseKey) > 8 {
			masked = entry.LicenseKey[:4] + "****" + entry.LicenseKey[len(entry.LicenseKey)-4:]
		}
		attrs = append(attrs, slog.String("license_key", masked))
	}

	if entry.UserEmail != "" {
		attrs = append(attrs, slog.String("user_email", entry.UserEmail))
	}

	if entry.MachineID != "" {
		attrs = append(attrs, slog.String("machine_id", entry.MachineID))
	}

	if entry.Duration > 0 {
		attrs = append(attrs, slog.Duration("duration", entry.Duration))
	}

	if entry.Error != nil {
		attrs = append(attrs, slog.String("error", entry.Error.Error()))
	}

	// Add metadata
	for k, v := range entry.Metadata {
		attrs = append(attrs, slog.Any(k, v))
	}

	// Log based on level
	switch entry.Level {
	case LogLevelDebug:
		l.slog.LogAttrs(nil, slog.LevelDebug, entry.Result, attrs...)
	case LogLevelInfo:
		l.slog.LogAttrs(nil, slog.LevelInfo, entry.Result, attrs...)
	case LogLevelWarn:
		l.slog.LogAttrs(nil, slog.LevelWarn, entry.Result, attrs...)
	case LogLevelError:
		l.slog.LogAttrs(nil, slog.LevelError, entry.Result, attrs...)
	case LogLevelFatal:
		l.slog.LogAttrs(nil, slog.LevelError, entry.Result, attrs...)
		// Note: Not calling os.Exit as per CLAUDE.md standards
	}
}

// Close closes the logger (no-op for slog)
func (l *Logger) Close() error {
	return nil
}

// SlogLogger wraps slog for direct slog usage
type SlogLogger struct {
	logger *slog.Logger
	level  LogLevel
}

// NewLicenseLoggerWithSlog creates a new license logger with direct slog usage
func NewLicenseLoggerWithSlog(slogLogger *slog.Logger, level LogLevel) (*SlogLogger, error) {
	if slogLogger == nil {
		return nil, errors.New("slog logger cannot be nil")
	}

	return &SlogLogger{
		logger: slogLogger.With(slog.String("component", "license")),
		level:  level,
	}, nil
}

// Info logs an info message with context
func (sl *SlogLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if len(args)%2 != 0 {
		sl.logger.InfoContext(ctx, msg)
		return
	}

	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	sl.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Error logs an error message with context
func (sl *SlogLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if len(args)%2 != 0 {
		sl.logger.ErrorContext(ctx, msg)
		return
	}

	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	sl.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// Debug logs a debug message with context
func (sl *SlogLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	if len(args)%2 != 0 {
		sl.logger.DebugContext(ctx, msg)
		return
	}

	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	sl.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Warn logs a warning message with context
func (sl *SlogLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if len(args)%2 != 0 {
		sl.logger.WarnContext(ctx, msg)
		return
	}

	attrs := make([]slog.Attr, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		if key, ok := args[i].(string); ok {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}
	sl.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// toSlogLevel converts LogLevel to slog.Level
func toSlogLevel(level LogLevel) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	case FatalLevel:
		return slog.LevelError // Map Fatal to Error since slog doesn't have Fatal
	default:
		return slog.LevelInfo
	}
}
