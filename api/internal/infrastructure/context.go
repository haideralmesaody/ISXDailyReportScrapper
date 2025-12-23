package infrastructure

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// GenerateTraceID creates a new unique trace ID using UUID v4
func GenerateTraceID() string {
	return uuid.New().String()
}

// ContextWithTraceID creates a new context with a generated trace ID
func ContextWithTraceID(ctx context.Context) context.Context {
	return WithTraceID(ctx, GenerateTraceID())
}

// EnsureTraceID ensures the context has a trace ID, generating one if needed
func EnsureTraceID(ctx context.Context) context.Context {
	if GetTraceID(ctx) == "" {
		return ContextWithTraceID(ctx)
	}
	return ctx
}

// LoggerWithContext creates a logger that includes the trace ID from context.
// This is the preferred way to get a logger for request handling.
func LoggerWithContext(ctx context.Context) *slog.Logger {
	logger := GetLogger()
	
	// Add trace_id if present
	if traceID := GetTraceID(ctx); traceID != "" {
		logger = logger.With("trace_id", traceID)
	}
	
	return logger
}

// InfoContext logs an info message with context awareness
func InfoContext(ctx context.Context, msg string, args ...any) {
	LoggerWithContext(ctx).InfoContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context awareness
func ErrorContext(ctx context.Context, msg string, args ...any) {
	LoggerWithContext(ctx).ErrorContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context awareness
func WarnContext(ctx context.Context, msg string, args ...any) {
	LoggerWithContext(ctx).WarnContext(ctx, msg, args...)
}

// DebugContext logs a debug message with context awareness
func DebugContext(ctx context.Context, msg string, args ...any) {
	LoggerWithContext(ctx).DebugContext(ctx, msg, args...)
}

// WithComponent creates a logger with a component field
func WithComponent(logger *slog.Logger, component string) *slog.Logger {
	return logger.With("component", component)
}

// WithError creates a logger with an error field
func WithError(logger *slog.Logger, err error) *slog.Logger {
	if err == nil {
		return logger
	}
	return logger.With("error", err.Error())
}

// WithFields creates a logger with multiple fields
func WithFields(logger *slog.Logger, fields map[string]interface{}) *slog.Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return logger.With(args...)
}