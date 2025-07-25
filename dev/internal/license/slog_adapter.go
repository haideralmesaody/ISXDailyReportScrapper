package license

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// SlogAdapter adapts the existing Logger interface to use slog
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new slog adapter for license logging
func NewSlogAdapter(level slog.Level) (*SlogAdapter, error) {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &SlogAdapter{
		logger: logger.With(
			slog.String("component", "license"),
		),
	}, nil
}

// Log logs a structured entry using slog
func (s *SlogAdapter) Log(entry LogEntry) {
	ctx := context.Background()
	
	// Convert log level
	var level slog.Level
	switch entry.Level {
	case LogLevelDebug:
		level = slog.LevelDebug
	case LogLevelInfo:
		level = slog.LevelInfo
	case LogLevelWarn:
		level = slog.LevelWarn
	case LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Build attributes
	attrs := []slog.Attr{
		slog.String("action", entry.Action),
		slog.String("result", entry.Result),
	}

	if entry.LicenseKey != "" {
		attrs = append(attrs, slog.String("license_key", entry.LicenseKey))
	}

	if entry.MachineID != "" {
		attrs = append(attrs, slog.String("machine_id", entry.MachineID))
	}

	if entry.Error != "" {
		attrs = append(attrs, slog.String("error", entry.Error))
	}

	if entry.Duration > 0 {
		attrs = append(attrs, slog.Duration("duration", time.Duration(entry.Duration)*time.Millisecond))
	}

	if entry.Details != nil {
		// Type assert to map[string]interface{}
		if details, ok := entry.Details.(map[string]interface{}); ok {
			for k, v := range details {
				attrs = append(attrs, slog.Any(k, v))
			}
		}
	}

	// Log with appropriate level
	s.logger.LogAttrs(ctx, level, entry.Result, attrs...)
}

// Close implements the Logger interface (no-op for slog)
func (s *SlogAdapter) Close() error {
	return nil
}

// Convenience methods for structured logging

// Debug logs a debug message
func (s *SlogAdapter) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

// Info logs an info message
func (s *SlogAdapter) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

// Warn logs a warning message
func (s *SlogAdapter) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}

// Error logs an error message
func (s *SlogAdapter) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

// WithContext returns a logger with context
func (s *SlogAdapter) WithContext(ctx context.Context) *SlogAdapter {
	return &SlogAdapter{
		logger: s.logger,
	}
}

// With returns a logger with additional attributes
func (s *SlogAdapter) With(args ...any) *SlogAdapter {
	return &SlogAdapter{
		logger: s.logger.With(args...),
	}
}