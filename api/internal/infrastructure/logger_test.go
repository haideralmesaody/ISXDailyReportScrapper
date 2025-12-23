package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/isxcli/isxcli/internal/config"
)

func TestInitializeLogger(t *testing.T) {
	// Reset logger state before test
	ResetLoggerForTesting()
	defer ResetLoggerForTesting() // Cleanup after test
	
	// Create temp directory for logs
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	cfg := config.LoggingConfig{
		Level:    "info",
		Format:   "json",
		Output:   "both",
		FilePath: logFile,
	}

	// Initialize logger
	logger, err := InitializeLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger is nil")
	}

	// Test that log file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Test logging
	logger.Info("test message", "key", "value")

	// Close log file to allow reading on Windows
	CloseLogFile()

	// Read log file to verify output
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify JSON format
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Errorf("Log output is not valid JSON: %v", err)
	}

	// Verify fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", logEntry["msg"])
	}
	if logEntry["key"] != "value" {
		t.Errorf("Expected key='value', got %v", logEntry["key"])
	}
	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level='INFO', got %v", logEntry["level"])
	}
}

func TestTraceIDInjection(t *testing.T) {
	// Reset logger state before test
	ResetLoggerForTesting()
	defer ResetLoggerForTesting() // Cleanup after test
	
	// Create temp directory for logs
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	cfg := config.LoggingConfig{
		Level:    "debug",
		Format:   "json",
		Output:   "both",
		FilePath: logFile,
	}

	// Initialize logger
	_, err := InitializeLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create context with trace ID
	ctx := WithTraceID(context.Background(), "test-trace-123")

	// Get logger with context
	logger := LoggerWithContext(ctx)
	logger.InfoContext(ctx, "test with trace")

	// Close log file to allow reading on Windows
	CloseLogFile()

	// Read log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	lastLine := lines[len(lines)-1]
	
	if err := json.Unmarshal([]byte(lastLine), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify trace_id
	if logEntry["trace_id"] != "test-trace-123" {
		t.Errorf("Expected trace_id='test-trace-123', got %v", logEntry["trace_id"])
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"error", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			// Reset logger state before test
			ResetLoggerForTesting()
			defer ResetLoggerForTesting() // Cleanup after test
			
			// Create temp directory
			tempDir := t.TempDir()
			logFile := filepath.Join(tempDir, "test.log")

			cfg := config.LoggingConfig{
				Level:    tt.level,
				Format:   "json",
				Output:   "both",
				FilePath: logFile,
			}

			// Initialize logger
			logger, err := InitializeLogger(cfg)
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			// Log at the configured level
			switch tt.level {
			case "debug":
				logger.Debug("test debug")
			case "info":
				logger.Info("test info")
			case "warn":
				logger.Warn("test warn")
			case "error":
				logger.Error("test error")
			}

			// Close log file to allow reading on Windows
			CloseLogFile()

			// Read and verify
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			var logEntry map[string]interface{}
			if err := json.Unmarshal(content, &logEntry); err != nil {
				t.Fatalf("Failed to parse log JSON: %v", err)
			}

			if logEntry["level"] != tt.expected {
				t.Errorf("Expected level=%s, got %v", tt.expected, logEntry["level"])
			}
		})
	}
}

func TestContextHelpers(t *testing.T) {
	// Reset logger state before test
	ResetLoggerForTesting()
	defer ResetLoggerForTesting() // Cleanup after test
	
	// Initialize logger
	cfg := config.Default()
	cfg.Logging.FilePath = filepath.Join(t.TempDir(), "test.log")
	
	_, err := InitializeLogger(cfg.Logging)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test context with trace ID
	ctx := ContextWithTraceID(context.Background())
	traceID := GetTraceID(ctx)
	
	if traceID == "" {
		t.Error("Expected trace ID to be generated")
	}

	// Test ensure trace ID (should not change existing)
	ctx2 := EnsureTraceID(ctx)
	if GetTraceID(ctx2) != traceID {
		t.Error("EnsureTraceID changed existing trace ID")
	}

	// Test ensure trace ID (should add if missing)
	ctx3 := EnsureTraceID(context.Background())
	if GetTraceID(ctx3) == "" {
		t.Error("EnsureTraceID did not add trace ID")
	}
}

func TestLoggerHelpers(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	
	// Create custom logger that writes to buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	globalLogger = logger

	// Test WithComponent
	componentLogger := WithComponent(logger, "test-component")
	componentLogger.Info("test message")

	// Verify component field
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logEntry["component"] != "test-component" {
		t.Errorf("Expected component='test-component', got %v", logEntry["component"])
	}

	// Test WithError
	buf.Reset()
	errLogger := WithError(logger, os.ErrNotExist)
	errLogger.Info("error test")

	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if !strings.Contains(logEntry["error"].(string), "file does not exist") {
		t.Errorf("Expected error to contain 'file does not exist', got %v", logEntry["error"])
	}

	// Test WithFields
	buf.Reset()
	fields := map[string]interface{}{
		"user_id": "123",
		"action":  "login",
	}
	fieldsLogger := WithFields(logger, fields)
	fieldsLogger.Info("fields test")

	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logEntry["user_id"] != "123" || logEntry["action"] != "login" {
		t.Error("Expected fields not found in log entry")
	}
}