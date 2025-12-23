package operations

import (
	"context"
	"log/slog"
	"time"
)

// OperationContext provides centralized logging context with correlation IDs
type OperationContext struct {
	TraceID       string
	OperationID   string
	Stage         string
	SessionID     string
	CorrelationID string
	Logger        *slog.Logger
	StartTime     time.Time
}

// NewOperationContext creates a new operation context with generated correlation IDs
func NewOperationContext(operationID, stage string, logger *slog.Logger) *OperationContext {
	traceID := generateTraceID()
	sessionID := generateSessionID()
	correlationID := generateCorrelationID(traceID, operationID)

	// Create enriched logger with correlation context
	var enrichedLogger *slog.Logger
	if logger != nil {
		enrichedLogger = logger.With(
			slog.String("trace_id", traceID),
			slog.String("operation_id", operationID),
			slog.String("stage", stage),
			slog.String("session_id", sessionID),
			slog.String("correlation_id", correlationID),
		)
	}

	return &OperationContext{
		TraceID:       traceID,
		OperationID:   operationID,
		Stage:         stage,
		SessionID:     sessionID,
		CorrelationID: correlationID,
		Logger:        enrichedLogger,
		StartTime:     time.Now(),
	}
}

// LogEvent logs an event with full correlation context
func (c *OperationContext) LogEvent(level slog.Level, message string, attrs ...slog.Attr) {
	if c.Logger == nil {
		return
	}

	// Add duration from start time for timing information
	durationAttrs := []slog.Attr{
		slog.Duration("operation_duration", time.Since(c.StartTime)),
	}
	durationAttrs = append(durationAttrs, attrs...)

	// Convert slog.Attr to interface{} for Logger.Log
	allAttrs := make([]interface{}, len(durationAttrs)*2)
	for i, attr := range durationAttrs {
		allAttrs[i*2] = attr.Key
		allAttrs[i*2+1] = attr.Value
	}

	c.Logger.Log(context.Background(), level, message, allAttrs...)
}

// LogInfo logs an info level event
func (c *OperationContext) LogInfo(message string, attrs ...slog.Attr) {
	c.LogEvent(slog.LevelInfo, message, attrs...)
}

// LogWarn logs a warning level event
func (c *OperationContext) LogWarn(message string, attrs ...slog.Attr) {
	c.LogEvent(slog.LevelWarn, message, attrs...)
}

// LogError logs an error level event
func (c *OperationContext) LogError(message string, attrs ...slog.Attr) {
	c.LogEvent(slog.LevelError, message, attrs...)
}

// LogDebug logs a debug level event
func (c *OperationContext) LogDebug(message string, attrs ...slog.Attr) {
	c.LogEvent(slog.LevelDebug, message, attrs...)
}

// WithContext adds additional context to the operation context
func (c *OperationContext) WithContext(additionalAttrs ...slog.Attr) *OperationContext {
	if c.Logger == nil {
		return c
	}

	// Convert slog.Attr to interface{} for Logger.With
	allAttrs := make([]interface{}, len(additionalAttrs)*2)
	for i, attr := range additionalAttrs {
		allAttrs[i*2] = attr.Key
		allAttrs[i*2+1] = attr.Value
	}

	// Create a new logger with additional attributes
	newLogger := c.Logger.With(allAttrs...)

	// Create new context with same IDs but enriched logger
	newContext := &OperationContext{
		TraceID:       c.TraceID,
		OperationID:   c.OperationID,
		Stage:         c.Stage,
		SessionID:     c.SessionID,
		CorrelationID: c.CorrelationID,
		Logger:        newLogger,
		StartTime:     c.StartTime,
	}

	return newContext
}

// WithStage creates a new context with a different stage
func (c *OperationContext) WithStage(stage string) *OperationContext {
	newContext := &OperationContext{
		TraceID:       c.TraceID,
		OperationID:   c.OperationID,
		Stage:         stage,
		SessionID:     c.SessionID,
		CorrelationID: c.CorrelationID,
		StartTime:     c.StartTime,
	}

	if c.Logger != nil {
		newContext.Logger = c.Logger.With(
			slog.String("stage", stage),
		)
	}

	return newContext
}

// GetDuration returns the duration since the operation started
func (c *OperationContext) GetDuration() time.Duration {
	return time.Since(c.StartTime)
}

// GetCorrelationAttrs returns the correlation attributes for external use
func (c *OperationContext) GetCorrelationAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("trace_id", c.TraceID),
		slog.String("operation_id", c.OperationID),
		slog.String("stage", c.Stage),
		slog.String("session_id", c.SessionID),
		slog.String("correlation_id", c.CorrelationID),
		slog.Duration("operation_duration", c.GetDuration()),
	}
}

// propagateContext propagates correlation context through standard Go context
func (c *OperationContext) propagateContext(parent context.Context) context.Context {
	return context.WithValue(parent, "operation_context", c)
}

// ContextFromContext extracts OperationContext from Go context
func ContextFromContext(ctx context.Context) *OperationContext {
	if ctx == nil {
		return nil
	}

	if oc, ok := ctx.Value("operation_context").(*OperationContext); ok {
		return oc
	}

	return nil
}

// generateTraceID generates a unique trace ID
func generateTraceID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return "sess_" + time.Now().Format("20060102") + "_" + randomString(6)
}

// generateCorrelationID generates a correlation ID from trace and operation IDs
func generateCorrelationID(traceID, operationID string) string {
	if operationID == "" {
		return traceID
	}
	return traceID + "-" + operationID
}

// randomString generates a random string for ID generation
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		// Simple pseudo-random for ID generation
		result[i] = charset[int(time.Now().UnixNano())%len(charset)]
	}
	return string(result)
}
