package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"

	"isxcli/internal/config"
	"isxcli/internal/license"
	"isxcli/internal/pipeline"
	"isxcli/internal/services"
	ws "isxcli/internal/websocket"
)

// simpleLogger implements services.Logger interface using slog
type simpleLogger struct {
	slog *slog.Logger
}

func (l *simpleLogger) Debug(msg string, args ...interface{}) {
	l.slog.Debug(msg, args...)
}

func (l *simpleLogger) Info(msg string, args ...interface{}) {
	l.slog.Info(msg, args...)
}

func (l *simpleLogger) Warn(msg string, args ...interface{}) {
	l.slog.Warn(msg, args...)
}

func (l *simpleLogger) Error(msg string, args ...interface{}) {
	l.slog.Error(msg, args...)
}

func (l *simpleLogger) Fatal(msg string, args ...interface{}) {
	l.slog.Error(msg, args...) // Don't actually exit in tests
}

// wsHubAdapter adapts ws.Hub to pipeline.WebSocketHub
type wsHubAdapter struct {
	hub *ws.Hub
}

func (w *wsHubAdapter) BroadcastUpdate(eventType, stage, status string, metadata interface{}) {
	data := map[string]interface{}{
		"type":     eventType,
		"stage":    stage,
		"status":   status,
		"metadata": metadata,
	}
	w.hub.BroadcastJSON(data)
}

func TestHealthHandler_BasicHealthCheck(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	paths := config.PathsConfig{
		DataDir: tempDir,
	}

	// Create logger adapter
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &simpleLogger{slog: slogLogger}
	
	// Create minimal dependencies
	licenseManager, _ := license.NewManager("")
	webSocketHub := ws.NewHub()
	wsAdapter := &wsHubAdapter{hub: webSocketHub}
	pipelineManager := pipeline.NewManager(wsAdapter, logger)
	
	// Create health service
	healthService := services.NewHealthService(
		"v1.0.0-test",
		"https://github.com/example/repo",
		paths,
		licenseManager,
		pipelineManager,
		webSocketHub,
		logger,
	)
	
	// Create handler
	handler := NewHealthHandler(healthService, logger)

	tests := []struct {
		name           string
		endpoint       string
		handlerFunc    http.HandlerFunc
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:           "health check endpoint",
			endpoint:       "/api/health",
			handlerFunc:    handler.HealthCheck,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "ok", response["status"])
				assert.Contains(t, response, "timestamp")
				assert.Equal(t, "v1.0.0-test", response["version"])
			},
		},
		{
			name:           "readiness check endpoint",
			endpoint:       "/api/health/ready",
			handlerFunc:    handler.ReadinessCheck,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				// May be "ready" or "not_ready" depending on license state
				assert.Contains(t, []string{"ready", "not_ready"}, response["status"])
				assert.Contains(t, response, "services")
			},
		},
		{
			name:           "liveness check endpoint",
			endpoint:       "/api/health/live",
			handlerFunc:    handler.LivenessCheck,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "alive", response["status"])
				assert.Contains(t, response, "runtime")
			},
		},
		{
			name:           "version endpoint",
			endpoint:       "/api/version",
			handlerFunc:    handler.Version,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "v1.0.0-test", response["version"])
				assert.Contains(t, response, "go_version")
				assert.Contains(t, response, "os")
				assert.Contains(t, response, "arch")
			},
		},
		{
			name:           "license status endpoint",
			endpoint:       "/api/license/status",
			handlerFunc:    handler.LicenseStatus,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				// License endpoint may return error or status
				// Just verify it returns valid JSON
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			rec := httptest.NewRecorder()

			// Execute handler
			tt.handlerFunc(rec, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rec.Code, "Expected status %d but got %d", tt.expectedStatus, rec.Code)

			// Check response if provided
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.Bytes())
			}
		})
	}
}

func TestHealthHandler_HandlerMethods(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	paths := config.PathsConfig{
		DataDir: tempDir,
	}

	// Create logger adapter
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &simpleLogger{slog: slogLogger}
	
	// Create minimal dependencies
	licenseManager, _ := license.NewManager("")
	webSocketHub := ws.NewHub()
	wsAdapter := &wsHubAdapter{hub: webSocketHub}
	pipelineManager := pipeline.NewManager(wsAdapter, logger)
	
	// Create health service
	healthService := services.NewHealthService(
		"v1.0.0-test",
		"https://github.com/example/repo",
		paths,
		licenseManager,
		pipelineManager,
		webSocketHub,
		logger,
	)
	
	// Create handler
	handler := NewHealthHandler(healthService, logger)

	// Test that all handler methods exist and don't panic
	t.Run("HealthCheck method exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			handler.HealthCheck(rec, req)
		})
	})

	t.Run("ReadinessCheck method exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health/ready", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			handler.ReadinessCheck(rec, req)
		})
	})

	t.Run("LivenessCheck method exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health/live", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			handler.LivenessCheck(rec, req)
		})
	})

	t.Run("Version method exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/version", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			handler.Version(rec, req)
		})
	})

	t.Run("LicenseStatus method exists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/license/status", nil)
		rec := httptest.NewRecorder()
		assert.NotPanics(t, func() {
			handler.LicenseStatus(rec, req)
		})
	})
}

func TestHealthHandler_TimingAndUptime(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	paths := config.PathsConfig{
		DataDir: tempDir,
	}

	// Create logger adapter
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &simpleLogger{slog: slogLogger}
	
	// Create minimal dependencies
	licenseManager, _ := license.NewManager("")
	webSocketHub := ws.NewHub()
	wsAdapter := &wsHubAdapter{hub: webSocketHub}
	pipelineManager := pipeline.NewManager(wsAdapter, logger)
	
	// Create health service
	healthService := services.NewHealthService(
		"v1.0.0-test",
		"https://github.com/example/repo",
		paths,
		licenseManager,
		pipelineManager,
		webSocketHub,
		logger,
	)
	
	// Wait a bit to ensure uptime > 0
	time.Sleep(100 * time.Millisecond)
	
	// Create handler
	handler := NewHealthHandler(healthService, logger)

	t.Run("uptime is greater than zero", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/health/live", nil)
		rec := httptest.NewRecorder()
		
		handler.LivenessCheck(rec, req)
		
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		runtime, ok := response["runtime"].(map[string]interface{})
		assert.True(t, ok, "runtime should be a map")
		
		uptime, ok := runtime["uptime"].(float64)
		assert.True(t, ok, "uptime should be a float64")
		assert.Greater(t, uptime, 0.0, "uptime should be greater than 0")
	})

	t.Run("version endpoint includes uptime", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/version", nil)
		rec := httptest.NewRecorder()
		
		handler.Version(rec, req)
		
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		uptime, ok := response["uptime"].(float64)
		assert.True(t, ok, "uptime should be a float64")
		assert.Greater(t, uptime, 0.0, "uptime should be greater than 0")
	})
}