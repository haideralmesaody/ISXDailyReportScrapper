package testutil

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// LogRecord represents a captured log record for testing
type LogRecord struct {
	Time    time.Time
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

// BufferedSlogHandler captures log records for testing
type BufferedSlogHandler struct {
	mu      sync.Mutex
	records []LogRecord
	t       *testing.T
}

// NewBufferedSlogHandler creates a new buffered handler for testing
func NewBufferedSlogHandler(t *testing.T) *BufferedSlogHandler {
	return &BufferedSlogHandler{
		records: make([]LogRecord, 0),
		t:       t,
	}
}

// Handle implements slog.Handler
func (h *BufferedSlogHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.records = append(h.records, LogRecord{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
		Attrs:   attrs,
	})

	// Also log to test output for debugging
	if h.t != nil {
		h.t.Logf("[%s] %s %v", r.Level, r.Message, attrs)
	}

	return nil
}

// Enabled implements slog.Handler
func (h *BufferedSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true // Capture all levels in tests
}

// WithAttrs implements slog.Handler
func (h *BufferedSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For testing, we can return the same handler
	// Real implementation would create a new handler with attrs
	return h
}

// WithGroup implements slog.Handler
func (h *BufferedSlogHandler) WithGroup(name string) slog.Handler {
	// For testing, we can return the same handler
	// Real implementation would create a new handler with group
	return h
}

// GetRecords returns all captured log records
func (h *BufferedSlogHandler) GetRecords() []LogRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Return a copy to avoid race conditions
	records := make([]LogRecord, len(h.records))
	copy(records, h.records)
	return records
}

// GetRecordsByLevel returns log records filtered by level
func (h *BufferedSlogHandler) GetRecordsByLevel(level slog.Level) []LogRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	var filtered []LogRecord
	for _, r := range h.records {
		if r.Level == level {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// ContainsMessage checks if any log record contains the given message
func (h *BufferedSlogHandler) ContainsMessage(message string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, r := range h.records {
		if strings.Contains(r.Message, message) {
			return true
		}
	}
	return false
}

// ContainsAttr checks if any log record contains the given attribute
func (h *BufferedSlogHandler) ContainsAttr(key string, value any) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, r := range h.records {
		if val, ok := r.Attrs[key]; ok && val == value {
			return true
		}
	}
	return false
}

// Clear removes all captured records
func (h *BufferedSlogHandler) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = h.records[:0]
}

// Count returns the number of captured records
func (h *BufferedSlogHandler) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.records)
}

// NewTestLogger creates a logger with a buffered handler for testing
func NewTestLogger(t *testing.T) (*slog.Logger, *BufferedSlogHandler) {
	handler := NewBufferedSlogHandler(t)
	logger := slog.New(handler)
	return logger, handler
}

// AssertLogContains checks if the handler contains a log with the given message
func AssertLogContains(t *testing.T, handler *BufferedSlogHandler, level slog.Level, message string) {
	t.Helper()
	
	records := handler.GetRecordsByLevel(level)
	for _, r := range records {
		if strings.Contains(r.Message, message) {
			return
		}
	}
	
	t.Errorf("Expected log message not found at level %s: %q", level, message)
	t.Logf("Captured logs at level %s:", level)
	for _, r := range records {
		t.Logf("  - %s", r.Message)
	}
}

// AssertLogAttr checks if the handler contains a log with the given attribute
func AssertLogAttr(t *testing.T, handler *BufferedSlogHandler, key string, expectedValue any) {
	t.Helper()
	
	if !handler.ContainsAttr(key, expectedValue) {
		t.Errorf("Expected log attribute not found: %s=%v", key, expectedValue)
		t.Logf("Captured logs:")
		for _, r := range handler.GetRecords() {
			t.Logf("  - %s: %v", r.Message, r.Attrs)
		}
	}
}

// AssertNoErrors checks that no error-level logs were recorded
func AssertNoErrors(t *testing.T, handler *BufferedSlogHandler) {
	t.Helper()
	
	errors := handler.GetRecordsByLevel(slog.LevelError)
	if len(errors) > 0 {
		t.Errorf("Unexpected error logs found:")
		for _, r := range errors {
			t.Errorf("  - %s: %v", r.Message, r.Attrs)
		}
	}
}