package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
)

// testLogger implements services.Logger interface for testing
type testLogger struct {
	slog *slog.Logger
}

func (l *testLogger) Debug(msg string, args ...interface{}) {
	l.slog.Debug(msg, args...)
}

func (l *testLogger) Info(msg string, args ...interface{}) {
	l.slog.Info(msg, args...)
}

func (l *testLogger) Warn(msg string, args ...interface{}) {
	l.slog.Warn(msg, args...)
}

func (l *testLogger) Error(msg string, args ...interface{}) {
	l.slog.Error(msg, args...)
}

func (l *testLogger) Fatal(msg string, args ...interface{}) {
	l.slog.Error(msg, args...) // Don't actually exit in tests
}

func TestClientLogHandler_Handle(t *testing.T) {
	// Create logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &testLogger{slog: slogLogger}
	
	// Create handler
	handler := NewClientLogHandler(logger)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "valid log entry",
			body: map[string]interface{}{
				"level":   "info",
				"message": "Test log message",
				"data": map[string]interface{}{
					"component": "test",
					"action":    "testing",
				},
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "log entry with error level",
			body: map[string]interface{}{
				"level":   "error",
				"message": "Test error message",
				"error":   "Something went wrong",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "log entry with debug level",
			body: map[string]interface{}{
				"level":   "debug",
				"message": "Debug message",
				"details": "Detailed debug information",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "log entry with warn level",
			body: map[string]interface{}{
				"level":   "warn",
				"message": "Warning message",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name:           "empty body",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid request format",
		},
		{
			name:           "invalid JSON",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Invalid request format",
		},
		{
			name: "missing level",
			body: map[string]interface{}{
				"message": "Test message without level",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
		{
			name: "missing message",
			body: map[string]interface{}{
				"level": "info",
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, err = json.Marshal(tt.body)
					assert.NoError(t, err)
				}
			}

			// Create request
			req := httptest.NewRequest("POST", "/api/log/client", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute handler
			handler.Handle(rec, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rec.Code, "Expected status %d but got %d", tt.expectedStatus, rec.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response["success"].(bool))
			} else {
				// Error responses have nested structure
				assert.False(t, response["success"].(bool))
				if errorData, ok := response["error"].(map[string]interface{}); ok {
					assert.Equal(t, tt.expectedMsg, errorData["message"])
				}
			}
		})
	}
}

func TestClientLogHandler_LogLevels(t *testing.T) {
	// Create logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &testLogger{slog: slogLogger}
	
	// Create handler
	handler := NewClientLogHandler(logger)

	levels := []string{"debug", "info", "warn", "error", "fatal"}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			body := map[string]interface{}{
				"level":   level,
				"message": "Test message for " + level,
				"timestamp": "2024-01-01T00:00:00Z",
				"userAgent": "TestClient/1.0",
			}

			bodyBytes, err := json.Marshal(body)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/api/log/client", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute handler
			handler.Handle(rec, req)

			// Assert success
			assert.Equal(t, http.StatusOK, rec.Code)
			
			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response["success"].(bool))
		})
	}
}

func TestClientLogHandler_LargePayload(t *testing.T) {
	// Create logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &testLogger{slog: slogLogger}
	
	// Create handler
	handler := NewClientLogHandler(logger)

	// Create a large payload
	largeData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		key := string(rune('a'+i%26)) + fmt.Sprintf("%d", i)
		largeData[key] = "value" + fmt.Sprintf("%d", i)
	}

	body := map[string]interface{}{
		"level":   "info",
		"message": "Large payload test",
		"data":    largeData,
	}

	bodyBytes, err := json.Marshal(body)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest("POST", "/api/log/client", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute handler
	handler.Handle(rec, req)

	// Assert success
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestClientLogHandler_SpecialCharacters(t *testing.T) {
	// Create logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &testLogger{slog: slogLogger}
	
	// Create handler
	handler := NewClientLogHandler(logger)

	tests := []struct {
		name    string
		message string
	}{
		{"unicode", "Test with unicode: 你好世界 🌍"},
		{"quotes", "Test with \"quotes\" and 'apostrophes'"},
		{"newlines", "Test with\nnewlines\nand\ttabs"},
		{"html", "Test with <html>tags</html>"},
		{"special", "Test with special chars: !@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]interface{}{
				"level":   "info",
				"message": tt.message,
			}

			bodyBytes, err := json.Marshal(body)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/api/log/client", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute handler
			handler.Handle(rec, req)

			// Assert success
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}