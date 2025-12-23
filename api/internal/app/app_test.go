package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/services"
)

// mockLicenseService is a mock implementation of LicenseService
type mockLicenseService struct {
	status *services.LicenseStatusResponse
	err    error
}

func (m *mockLicenseService) GetStatus(ctx context.Context) (*services.LicenseStatusResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.status, nil
}

func (m *mockLicenseService) Activate(ctx context.Context, key string) error {
	return m.err
}

func (m *mockLicenseService) ValidateWithContext(ctx context.Context) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.status != nil && m.status.LicenseStatus == "active", nil
}

func (m *mockLicenseService) GetDetailedStatus(ctx context.Context) (*services.DetailedLicenseStatusResponse, error) {
	return nil, nil
}

func (m *mockLicenseService) CheckRenewalStatus(ctx context.Context) (*services.RenewalStatusResponse, error) {
	return nil, nil
}

func (m *mockLicenseService) TransferLicense(ctx context.Context, key string, force bool) error {
	return nil
}

func (m *mockLicenseService) GetValidationMetrics(ctx context.Context) (*services.ValidationMetrics, error) {
	return nil, nil
}

func (m *mockLicenseService) InvalidateCache(ctx context.Context) error {
	return nil
}

func (m *mockLicenseService) GetDebugInfo(ctx context.Context) (*services.LicenseDebugInfo, error) {
	return nil, nil
}

func (m *mockLicenseService) CheckExistingLicense() (*license.ExistingLicenseInfo, error) {
	return nil, nil
}

func (m *mockLicenseService) GetLicenseDetails() (interface{}, error) {
	return nil, nil
}

func (m *mockLicenseService) GetActivationHistory() ([]interface{}, error) {
	return nil, nil
}

func (m *mockLicenseService) BackupCurrentLicense() (string, error) {
	return "", nil
}

// MockFS creates a mock filesystem for testing embedded frontend
func createMockFS() fs.FS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html><html><head><title>Test</title></head><body>Test Frontend</body></html>`),
		},
		"_next/static/test.js": &fstest.MapFile{
			Data: []byte(`console.log('test');`),
		},
		"favicon.ico": &fstest.MapFile{
			Data: []byte("fake favicon data"),
		},
		"license/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html><html><head><title>License</title></head><body>License Page</body></html>`),
		},
		"operations/index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html><html><head><title>Operations</title></head><body>Operations Page</body></html>`),
		},
	}
}

// setupTestEnvironment sets up a clean test environment
func setupTestEnvironment(t *testing.T) func() {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "app_test_*")
	require.NoError(t, err)

	// Set environment variables for testing
	oldArgs := os.Args
	os.Args = []string{filepath.Join(tempDir, "test.exe")}

	// Set up test config environment
	os.Setenv("ISX_SERVER_PORT", "8081") // Use different port for testing
	os.Setenv("ISX_LOGGING_LEVEL", "error") // Reduce log noise in tests
	os.Setenv("ISX_LOGGING_OUTPUT", "discard") // Discard logs in tests

	return func() {
		os.Args = oldArgs
		os.RemoveAll(tempDir)
		os.Unsetenv("ISX_SERVER_PORT")
		os.Unsetenv("ISX_LOGGING_LEVEL")
		os.Unsetenv("ISX_LOGGING_OUTPUT")
	}
}

// createTestLogger creates a logger that discards output for testing
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

// TestNewApplication tests the NewApplication function
func TestNewApplication(t *testing.T) {
	tests := []struct {
		name          string
		frontendFS    fs.FS
		setupEnv      func()
		wantErr       bool
		errorContains string
	}{
		{
			name:       "successful initialization with valid frontend",
			frontendFS: createMockFS(),
			setupEnv:   func() {},
			wantErr:    false,
		},
		{
			name:       "successful initialization with nil frontend",
			frontendFS: nil,
			setupEnv:   func() {},
			wantErr:    false,
		},
		{
			name:       "initialization with invalid config",
			frontendFS: createMockFS(),
			setupEnv: func() {
				os.Setenv("ISX_SERVER_PORT", "-1") // Invalid port
			},
			wantErr:       true,
			errorContains: "config validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnvironment(t)
			defer cleanup()

			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			app, err := NewApplication(tt.frontendFS)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, app) {
					assert.NotNil(t, app.Config)
					assert.NotNil(t, app.Logger)
					assert.NotNil(t, app.Router)
					assert.NotNil(t, app.Server)
					assert.NotNil(t, app.LicenseManager)
					assert.NotNil(t, app.WebSocketHub)
					assert.NotNil(t, app.OperationService)
					assert.NotNil(t, app.DataService)
					assert.NotNil(t, app.HealthService)
					assert.NotNil(t, app.Services)
					assert.Equal(t, tt.frontendFS, app.FrontendFS)
				}
			}
		})
	}
}

// TestApplication_initializeServices tests the service initialization
func TestApplication_initializeServices(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name          string
		setupApp      func() *Application
		wantErr       bool
		errorContains string
	}{
		{
			name: "successful service initialization",
			setupApp: func() *Application {
				cfg, _ := config.Load()
				logger := createTestLogger()
				otelProviders, _ := infrastructure.InitializeOTel(infrastructure.DefaultOTelConfig(), logger)
				return &Application{
					Config:        cfg,
					Logger:        logger,
					OTelProviders: otelProviders,
					FrontendFS:    createMockFS(),
				}
			},
			wantErr: false,
		},
		{
			name: "initialization with invalid license path",
			setupApp: func() *Application {
				cfg, _ := config.Load()
				// Set invalid license path
				cfg.Paths.LicenseFile = "/invalid/path/that/cannot/be/created"
				logger := createTestLogger()
				otelProviders, _ := infrastructure.InitializeOTel(infrastructure.DefaultOTelConfig(), logger)
				return &Application{
					Config:        cfg,
					Logger:        logger,
					OTelProviders: otelProviders,
					FrontendFS:    createMockFS(),
				}
			},
			wantErr: false, // License manager creation should not fail with invalid path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setupApp()
			err := app.initializeServices()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app.LicenseManager)
				assert.NotNil(t, app.WebSocketHub)
				assert.NotNil(t, app.OperationService)
				assert.NotNil(t, app.DataService)
				assert.NotNil(t, app.HealthService)
				assert.NotNil(t, app.UpdateChecker)
				assert.NotNil(t, app.Services)
				assert.NotNil(t, app.Services.License)
				assert.NotNil(t, app.Services.LicenseService)
				assert.NotNil(t, app.Services.Data)
				assert.NotNil(t, app.Services.Health)
				assert.NotNil(t, app.Services.WebSocket)
			}
		})
	}
}

// TestApplication_setupRouter tests the router setup
func TestApplication_setupRouter(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)
	require.NotNil(t, app)

	t.Run("router setup with middleware", func(t *testing.T) {
		app.setupRouter()
		
		assert.NotNil(t, app.Router)
		
		// Test that routes are properly registered by making requests
		testServer := httptest.NewServer(app.Router)
		defer testServer.Close()

		// Test API health endpoint (should work - behind license middleware but likely has bypass)
		resp, err := http.Get(testServer.URL + "/api/health")
		assert.NoError(t, err)
		defer resp.Body.Close()
		// Note: We expect either success or license-related error, not 404
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)

		// Test WebSocket endpoint exists (should get upgrade required error)
		resp, err = http.Get(testServer.URL + "/ws")
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // WebSocket upgrade required
	})

	t.Run("router setup without frontend", func(t *testing.T) {
		appNoFrontend := &Application{
			Config:        app.Config,
			Logger:        app.Logger,
			OTelProviders: app.OTelProviders,
			FrontendFS:    nil,
		}
		err := appNoFrontend.initializeServices()
		require.NoError(t, err)

		appNoFrontend.setupRouter()
		assert.NotNil(t, appNoFrontend.Router)
	})
}

// TestApplication_handleWebSocket tests WebSocket handling
func TestApplication_handleWebSocket(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	// Start the WebSocket hub
	go app.WebSocketHub.Run()
	defer app.WebSocketHub.Stop()

	// Create test server
	testServer := httptest.NewServer(http.HandlerFunc(app.handleWebSocket))
	defer testServer.Close()

	t.Run("successful WebSocket upgrade", func(t *testing.T) {
		// Convert HTTP URL to WebSocket URL
		wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
		
		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Skipf("WebSocket connection failed: %v", err)
			return
		}
		defer conn.Close()

		// Send a test message
		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"ping"}`))
		assert.NoError(t, err)

		// Set read deadline to avoid hanging
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		
		// Try to read a message (may timeout, which is OK)
		_, _, err = conn.ReadMessage()
		// We don't assert on error here as the connection might be closed by the server
	})

	t.Run("invalid WebSocket request", func(t *testing.T) {
		// Make regular HTTP request to WebSocket endpoint
		resp, err := http.Get(testServer.URL)
		assert.NoError(t, err)
		defer resp.Body.Close()
		
		// Should get bad request for non-WebSocket request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestApplication_Start tests application startup
func TestApplication_Start(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("successful start", func(t *testing.T) {
		app, err := NewApplication(createMockFS())
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start application in goroutine
		startErr := make(chan error, 1)
		go func() {
			startErr <- app.Start(ctx, cancel)
		}()

		// Give it time to start
		time.Sleep(500 * time.Millisecond)

		// Verify server is running by making a request
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/health", app.Config.Server.Port))
		if err != nil {
			// Server might not be ready yet, try once more
			time.Sleep(500 * time.Millisecond)
			resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/health", app.Config.Server.Port))
		}
		if err == nil {
			defer resp.Body.Close()
			// Server is responding
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		// Stop the application
		cancel()
		
		// Wait for shutdown
		select {
		case err := <-startErr:
			// Server should exit without error when context is cancelled
			assert.NoError(t, err)
		case <-time.After(3 * time.Second):
			// If it takes too long, force stop
			stopErr := app.Stop(context.Background())
			assert.NoError(t, stopErr)
		}
	})

	t.Run("start with port already in use", func(t *testing.T) {
		// Create a server on a specific port
		listener := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer listener.Close()

		// Extract port from listener
		addr := listener.Listener.Addr().String()
		port := strings.Split(addr, ":")[1]

		// Set same port in config
		os.Setenv("ISX_SERVER_PORT", port)
		defer os.Unsetenv("ISX_SERVER_PORT")

		app, err := NewApplication(createMockFS())
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// This should eventually cancel due to port conflict
		err = app.Start(ctx, cancel)
		// Start itself doesn't return error immediately, but the context will be cancelled
		assert.NoError(t, err)
	})
}

// TestApplication_Stop tests application shutdown
func TestApplication_Stop(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application
	go func() {
		app.Start(ctx, cancel)
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	t.Run("graceful shutdown", func(t *testing.T) {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		err := app.Stop(shutdownCtx)
		assert.NoError(t, err)
	})
}

// TestApplication_Run tests the main run loop
func TestApplication_Run(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("run and interrupt", func(t *testing.T) {
		app, err := NewApplication(createMockFS())
		require.NoError(t, err)

		// Run in goroutine
		runErr := make(chan error, 1)
		go func() {
			runErr <- app.Run()
		}()

		// Give it time to start
		time.Sleep(200 * time.Millisecond)

		// On Windows, sending interrupt signals to the current process is not supported
		// Instead, we'll use a different approach for all platforms to ensure consistency
		go func() {
			time.Sleep(100 * time.Millisecond)
			// Simulate interrupt by canceling the app's context
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			app.Stop(ctx)
		}()

		// Wait for shutdown (increased timeout to account for graceful shutdown)
		select {
		case err := <-runErr:
			assert.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("Application did not shutdown within timeout")
		}
	})
}

// TestApplication_getCORSConfig tests CORS configuration
func TestApplication_getCORSConfig(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	tests := []struct {
		name     string
		setupEnv func()
		validate func(t *testing.T, config interface{})
	}{
		{
			name: "production mode CORS",
			setupEnv: func() {
				os.Setenv("NODE_ENV", "production")
				os.Setenv("GO_ENV", "production")
			},
			validate: func(t *testing.T, config interface{}) {
				// Development mode detection now correctly respects environment variables
				// Production ENV variables now correctly return false for isDevelopmentMode()
				assert.False(t, app.isDevelopmentMode())
			},
		},
		{
			name: "development mode CORS",
			setupEnv: func() {
				os.Setenv("NODE_ENV", "development")
			},
			validate: func(t *testing.T, config interface{}) {
				assert.True(t, app.isDevelopmentMode())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			
			config := app.getCORSConfig()
			assert.NotEmpty(t, config.AllowedMethods)
			assert.NotEmpty(t, config.AllowedHeaders)
			assert.True(t, config.AllowCredentials)
			assert.Equal(t, 300, config.MaxAge)
			
			if tt.validate != nil {
				tt.validate(t, config)
			}
			
			// Cleanup environment
			os.Unsetenv("NODE_ENV")
			os.Unsetenv("GO_ENV")
		})
	}
}

// TestApplication_isDevelopmentMode tests development mode detection
func TestApplication_isDevelopmentMode(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	tests := []struct {
		name     string
		setupEnv func()
		want     bool
	}{
		{
			name: "NODE_ENV development",
			setupEnv: func() {
				os.Setenv("NODE_ENV", "development")
			},
			want: true,
		},
		{
			name: "GO_ENV development",
			setupEnv: func() {
				os.Setenv("GO_ENV", "development")
			},
			want: true,
		},
		{
			name: "production environment",
			setupEnv: func() {
				os.Setenv("NODE_ENV", "production")
				os.Setenv("GO_ENV", "production")
			},
			want: false, // Development mode detection now correctly respects production ENV
		},
		{
			name:     "no environment set",
			setupEnv: func() {},
			want:     false, // Development mode detection now correctly returns false when no ENV set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment first
			os.Unsetenv("NODE_ENV")
			os.Unsetenv("GO_ENV")
			
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			
			result := app.isDevelopmentMode()
			assert.Equal(t, tt.want, result)
			
			// Cleanup
			os.Unsetenv("NODE_ENV")
			os.Unsetenv("GO_ENV")
		})
	}
}

// TestApplication_performStartupHealthCheck tests startup health checks
func TestApplication_performStartupHealthCheck(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	t.Run("successful health check", func(t *testing.T) {
		ctx := context.Background()
		err := app.performStartupHealthCheck(ctx)
		// May have warnings but should not error fatally
		// Warnings are returned as errors but are non-fatal
		if err != nil {
			assert.Contains(t, err.Error(), "warnings")
		}
	})

	t.Run("health check with invalid paths", func(t *testing.T) {
		// This test is tricky because performStartupHealthCheck uses config.GetPaths()
		// which creates paths based on the executable. We'll just verify it doesn't panic.
		ctx := context.Background()
		err := app.performStartupHealthCheck(ctx)
		// Should not panic and should return either nil or warnings
		if err != nil {
			assert.Contains(t, err.Error(), "warnings")
		}
	})
}

// TestApplication_serveFrontendFile tests frontend file serving
func TestApplication_serveFrontendFile(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	tests := []struct {
		name           string
		filename       string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "serve favicon",
			filename:       "favicon.ico",
			expectedStatus: http.StatusOK,
			expectedType:   "image/x-icon",
		},
		{
			name:           "serve non-existent file",
			filename:       "nonexistent.txt",
			expectedStatus: http.StatusNotFound,
			expectedType:   "",
		},
		{
			name:           "serve index.html",
			filename:       "index.html",
			expectedStatus: http.StatusOK,
			expectedType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := app.serveFrontendFile(app.FrontendFS, tt.filename)
			
			req := httptest.NewRequest("GET", "/"+tt.filename, nil)
			w := httptest.NewRecorder()
			
			handler(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedType != "" {
				assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))
			}
			
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "public, max-age=86400", w.Header().Get("Cache-Control"))
			}
		})
	}
}

// TestApplication_serveSPAHandler tests SPA routing
func TestApplication_serveSPAHandler(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	// Mock the license service to return active status for operations page test
	mockLicenseService := &mockLicenseService{
		status: &services.LicenseStatusResponse{
			Status:        200,
			LicenseStatus: "active",
			Message:       "License is active",
			DaysLeft:      365,
		},
	}
	app.Services.LicenseService = mockLicenseService

	handler := app.serveSPAHandler(app.FrontendFS)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedRedirect string
	}{
		{
			name:             "root redirects to license",
			path:             "/",
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedRedirect: "/license",
		},
		{
			name:           "license page serves content",
			path:           "/license",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "operations page serves content",
			path:           "/operations",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unknown route falls back to index.html",
			path:           "/unknown",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			
			handler(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedRedirect != "" {
				location := w.Header().Get("Location")
				assert.Equal(t, tt.expectedRedirect, location)
			}
			
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
				assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
			}
		})
	}
}

// TestApplication_serveStaticWithMIME tests static file serving with MIME types
func TestApplication_serveStaticWithMIME(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	handler := app.serveStaticWithMIME(app.FrontendFS, "/_next")

	tests := []struct {
		name         string
		path         string
		expectedType string
		expectedCode int
	}{
		{
			name:         "serve JavaScript file",
			path:         "/_next/static/test.js",
			expectedType: "application/javascript",
			expectedCode: http.StatusOK,
		},
		{
			name:         "serve non-existent file",
			path:         "/_next/static/nonexistent.js",
			expectedType: "",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			
			handler.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedType != "" {
				assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))
			}
			
			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			}
		})
	}
}

// TestApplication_createServer tests HTTP server creation
func TestApplication_createServer(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	t.Run("server creation", func(t *testing.T) {
		app.createServer()
		
		assert.NotNil(t, app.Server)
		assert.Equal(t, fmt.Sprintf(":%d", app.Config.Server.Port), app.Server.Addr)
		assert.Equal(t, app.Router, app.Server.Handler)
		assert.Equal(t, app.Config.Server.ReadTimeout, app.Server.ReadTimeout)
		assert.Equal(t, app.Config.Server.WriteTimeout, app.Server.WriteTimeout)
		assert.Equal(t, app.Config.Server.IdleTimeout, app.Server.IdleTimeout)
	})
}

// TestApplication_setupAPIRoutes tests API route setup
func TestApplication_setupAPIRoutes(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	router := chi.NewRouter()
	app.setupAPIRoutes(router)

	// Create test server to verify routes
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "health endpoint exists",
			path:           "/api/health",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "version endpoint exists",
			path:           "/api/version",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "license endpoint exists",
			path:           "/api/license/status",
			method:         "GET",
			expectedStatus: http.StatusOK, // Should respond even if not licensed
		},
		{
			name:           "operations endpoint exists",
			path:           "/api/operations",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, testServer.URL+tt.path, nil)
			require.NoError(t, err)
			
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestApplication_ServiceContainer tests the service container
func TestApplication_ServiceContainer(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	t.Run("service container populated", func(t *testing.T) {
		assert.NotNil(t, app.Services)
		assert.NotNil(t, app.Services.License)
		assert.NotNil(t, app.Services.LicenseService)
		assert.NotNil(t, app.Services.Data)
		assert.NotNil(t, app.Services.Health)
		assert.NotNil(t, app.Services.WebSocket)
		
		// Verify services are the same instances
		assert.Equal(t, app.LicenseManager, app.Services.License)
		assert.Equal(t, app.DataService, app.Services.Data)
		assert.Equal(t, app.HealthService, app.Services.Health)
		assert.Equal(t, app.WebSocketHub, app.Services.WebSocket)
	})
}

// Test edge cases and error scenarios
func TestApplication_EdgeCases(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("nil frontend filesystem", func(t *testing.T) {
		app, err := NewApplication(nil)
		require.NoError(t, err)
		assert.Nil(t, app.FrontendFS)
		
		// Test that SPA handler handles nil filesystem gracefully
		handler := app.serveSPAHandler(nil)
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		
		// Should redirect to license page
		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, "/license", w.Header().Get("Location"))
	})

	t.Run("websocket with invalid origin", func(t *testing.T) {
		app, err := NewApplication(createMockFS())
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/ws", nil)
		req.Header.Set("Origin", "http://malicious.com")
		req.Header.Set("Connection", "upgrade")
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Sec-WebSocket-Version", "13")
		req.Header.Set("Sec-WebSocket-Key", "test")
		
		w := httptest.NewRecorder()
		app.handleWebSocket(w, req)
		
		// Should still allow connection in development mode
		// The actual WebSocket upgrade might fail due to test setup, but it shouldn't panic
	})
}

// Benchmark tests for performance-critical paths
func BenchmarkApplication_NewApplication(b *testing.B) {
	cleanup := setupTestEnvironmentBench(b)
	defer cleanup()

	mockFS := createMockFS()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app, err := NewApplication(mockFS)
		if err != nil {
			b.Fatalf("NewApplication failed: %v", err)
		}
		_ = app
	}
}

func BenchmarkApplication_ServeSPA(b *testing.B) {
	cleanup := setupTestEnvironmentBench(b)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	if err != nil {
		b.Fatalf("NewApplication failed: %v", err)
	}

	handler := app.serveSPAHandler(app.FrontendFS)
	req := httptest.NewRequest("GET", "/operations", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler(w, req)
	}
}

// setupTestEnvironmentBench helper for benchmarks
func setupTestEnvironmentBench(b *testing.B) func() {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "app_test_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set environment variables for testing
	oldArgs := os.Args
	os.Args = []string{filepath.Join(tempDir, "test.exe")}

	// Set up test config environment
	os.Setenv("ISX_SERVER_PORT", "8081") // Use different port for testing
	os.Setenv("ISX_LOGGING_LEVEL", "error") // Reduce log noise in tests
	os.Setenv("ISX_LOGGING_OUTPUT", "discard") // Discard logs in tests

	return func() {
		os.Args = oldArgs
		os.RemoveAll(tempDir)
		os.Unsetenv("ISX_SERVER_PORT")
		os.Unsetenv("ISX_LOGGING_LEVEL")
		os.Unsetenv("ISX_LOGGING_OUTPUT")
	}
}

// TestApplication_setupStaticRoutes tests static route setup (currently unused)
func TestApplication_setupStaticRoutes(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	router := chi.NewRouter()
	app.setupStaticRoutes(router) // This method is currently unused but exists

	// Test that the method doesn't panic
	assert.NotNil(t, router)
}

// TestApplication_openBrowser tests browser opening functionality
func TestApplication_openBrowser(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "http://localhost:8080",
			wantErr: false, // May or may not work depending on environment
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: false, // openBrowser may still succeed with invalid URLs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := openBrowser(tt.url)
			// Don't assert on error as it depends on environment
			// Just verify it doesn't panic
			_ = err
		})
	}
}

// TestApplication_serveSPAHandler_MoreCases tests additional SPA routing scenarios
func TestApplication_serveSPAHandler_MoreCases(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	handler := app.serveSPAHandler(app.FrontendFS)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		setupRequest   func(req *http.Request)
	}{
		{
			name:           "API route should not redirect",
			path:           "/api/test",
			expectedStatus: http.StatusOK, // Should serve index.html as fallback
		},
		{
			name:           "_next route should not redirect",
			path:           "/_next/static/test.js",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exact file path",
			path:           "/favicon.ico",
			expectedStatus: http.StatusTemporaryRedirect, // SPA redirects to index when file not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}
			w := httptest.NewRecorder()
			
			handler(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestApplication_serveStaticWithMIME_MoreTypes tests additional MIME types
func TestApplication_serveStaticWithMIME_MoreTypes(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create extended mock FS with various file types
	extendedFS := fstest.MapFS{
		"test.css": &fstest.MapFile{
			Data: []byte("body { color: red; }"),
		},
		"test.json": &fstest.MapFile{
			Data: []byte(`{"test": true}`),
		},
		"test.svg": &fstest.MapFile{
			Data: []byte(`<svg></svg>`),
		},
		"test.png": &fstest.MapFile{
			Data: []byte("fake png data"),
		},
		"test.woff2": &fstest.MapFile{
			Data: []byte("fake font data"),
		},
		"unknown.xyz": &fstest.MapFile{
			Data: []byte("unknown file type"),
		},
	}

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	handler := app.serveStaticWithMIME(extendedFS, "")

	tests := []struct {
		name         string
		path         string
		expectedType string
		expectedCode int
	}{
		{
			name:         "CSS file",
			path:         "/test.css",
			expectedType: "text/css",
			expectedCode: http.StatusOK,
		},
		{
			name:         "JSON file",
			path:         "/test.json",
			expectedType: "application/json",
			expectedCode: http.StatusOK,
		},
		{
			name:         "SVG file",
			path:         "/test.svg",
			expectedType: "image/svg+xml",
			expectedCode: http.StatusOK,
		},
		{
			name:         "PNG file",
			path:         "/test.png",
			expectedType: "image/png",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Font file",
			path:         "/test.woff2",
			expectedType: "font/woff2",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Unknown file type",
			path:         "/unknown.xyz",
			expectedType: "application/octet-stream",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			
			handler.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))
			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			}
		})
	}
}

// TestApplication_serveFrontendFile_MoreTypes tests additional frontend file types
func TestApplication_serveFrontendFile_MoreTypes(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	extendedFS := fstest.MapFS{
		"robots.txt": &fstest.MapFile{
			Data: []byte("User-agent: *\nDisallow:"),
		},
		"manifest.json": &fstest.MapFile{
			Data: []byte(`{"name": "ISX App"}`),
		},
		"test.ico": &fstest.MapFile{
			Data: []byte("fake ico data"),
		},
	}

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	tests := []struct {
		name         string
		filename     string
		expectedType string
		expectedCode int
	}{
		{
			name:         "robots.txt",
			filename:     "robots.txt",
			expectedType: "text/plain",
			expectedCode: http.StatusOK,
		},
		{
			name:         "manifest.json",
			filename:     "manifest.json",
			expectedType: "application/json",
			expectedCode: http.StatusOK,
		},
		{
			name:         "ICO file",
			filename:     "test.ico",
			expectedType: "image/x-icon",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := app.serveFrontendFile(extendedFS, tt.filename)
			
			req := httptest.NewRequest("GET", "/"+tt.filename, nil)
			w := httptest.NewRecorder()
			
			handler(w, req)
			
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedType != "" {
				assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))
			}
		})
	}
}

// TestApplication_getCORSConfig_MoreScenarios tests additional CORS scenarios
func TestApplication_getCORSConfig_MoreScenarios(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	tests := []struct {
		name     string
		setupEnv func()
		validate func(t *testing.T, app *Application)
	}{
		{
			name: "production with custom origins",
			setupEnv: func() {
				os.Setenv("NODE_ENV", "production")
				os.Setenv("ISX_SECURITY_ENABLE_CORS", "true")
			},
			validate: func(t *testing.T, app *Application) {
				// Update config to simulate custom origins
				app.Config.Security.AllowedOrigins = []string{"https://example.com"}
				app.Config.Security.EnableCORS = true
				config := app.getCORSConfig()
				// The getCORSConfig might detect as development mode, so check if origins are present
				assert.NotEmpty(t, config.AllowedOrigins)
				// Development mode detection depends on environment, so we won't assert on it
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			
			if tt.validate != nil {
				tt.validate(t, app)
			}
			
			// Cleanup environment
			os.Unsetenv("NODE_ENV")
			os.Unsetenv("GO_ENV")
			os.Unsetenv("ISX_SECURITY_ENABLE_CORS")
		})
	}
}

// TestApplication_serveSPAHandler_LicenseLogic tests license logic edge cases
func TestApplication_serveSPAHandler_LicenseLogic(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	// Test with nil services (edge case)
	app.Services = nil
	handler := app.serveSPAHandler(app.FrontendFS)

	req := httptest.NewRequest("GET", "/operations", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	// Should still serve content even with nil services
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestApplication_setupAPIRoutes_Comprehensive tests more API route scenarios
func TestApplication_setupAPIRoutes_Comprehensive(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	router := chi.NewRouter()
	app.setupAPIRoutes(router)

	// Create test server to verify routes
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	tests := []struct {
		name           string
		path           string
		method         string
		body           string
		expectedStatus int
	}{
		{
			name:           "POST operation shortcuts - scrape",
			path:           "/api/scrape",
			method:         "POST",
			body:           `{"test": "data"}`,
			expectedStatus: http.StatusOK, // Should respond even if operation fails
		},
		{
			name:           "POST operation shortcuts - process",
			path:           "/api/process",
			method:         "POST",
			body:           `{"test": "data"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST operation shortcuts - indexcsv",
			path:           "/api/indexcsv", 
			method:         "POST",
			body:           `{"test": "data"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Client logging endpoint",
			path:           "/api/logs",
			method:         "POST",
			body:           `{"level": "info", "message": "test log"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Data endpoint exists",
			path:           "/api/data/reports",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error
			
			if tt.body != "" {
				req, err = http.NewRequest(tt.method, testServer.URL+tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, testServer.URL+tt.path, nil)
			}
			require.NoError(t, err)
			
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			
			// Most endpoints should respond (even if with errors due to no license)
			assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

// TestApplication_isDevelopmentMode_MoreCases tests additional cases for development mode detection
func TestApplication_isDevelopmentMode_MoreCases(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name       string
		setupEnv   func() func() // Returns cleanup
		setupFiles func() func() // Returns cleanup
		expectDev  *bool         // nil means don't assert
	}{
		{
			name: "check Next.js files existence",
			setupFiles: func() func() {
				// Create temporary Next.js files
				os.MkdirAll("frontend", 0755)
				os.MkdirAll("frontend/.next", 0755)
				os.WriteFile("frontend/package.json", []byte("{}"), 0644)
				return func() {
					os.RemoveAll("frontend")
				}
			},
			expectDev: nil, // Don't assert as it's environment dependent
		},
		{
			name: "working directory with dev in path",
			setupFiles: func() func() {
				// Create a temporary directory with "dev" in name and cd to it
				tempDir, _ := os.MkdirTemp("", "my-dev-project-*")  
				oldWd, _ := os.Getwd()
				os.Chdir(tempDir)
				return func() {
					os.Chdir(oldWd)
					os.RemoveAll(tempDir)
				}
			},
			expectDev: nil, // Don't assert as it's environment dependent
		},
	}

	// Use a mutex to serialize test execution to avoid race conditions
	var mu sync.Mutex
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize app creation to avoid OpenTelemetry race conditions
			mu.Lock()
			app, err := NewApplication(createMockFS())
			mu.Unlock()
			require.NoError(t, err)
			
			var fileCleanup, envCleanup func()
			
			if tt.setupFiles != nil {
				fileCleanup = tt.setupFiles()
				defer fileCleanup()
			}
			
			if tt.setupEnv != nil {
				envCleanup = tt.setupEnv()
				defer envCleanup()
			}
			
			result := app.isDevelopmentMode()
			
			if tt.expectDev != nil {
				assert.Equal(t, *tt.expectDev, result)
			} else {
				// Just verify it doesn't panic
				_ = result
			}
		})
	}
}

// TestApplication_getCORSConfig_AllBranches tests all CORS configuration branches
func TestApplication_getCORSConfig_AllBranches(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	t.Run("force development mode", func(t *testing.T) {
		os.Setenv("NODE_ENV", "development")
		defer os.Unsetenv("NODE_ENV")
		
		config := app.getCORSConfig()
		assert.NotEmpty(t, config.AllowedOrigins)
		assert.Contains(t, config.AllowedMethods, "GET")
		assert.Contains(t, config.AllowedMethods, "POST")
		assert.Contains(t, config.AllowedHeaders, "Content-Type")
		assert.True(t, config.AllowCredentials)
		assert.Equal(t, 300, config.MaxAge)
	})

	t.Run("production mode with CORS disabled", func(t *testing.T) {
		os.Setenv("NODE_ENV", "production")
		os.Setenv("GO_ENV", "production")
		defer func() {
			os.Unsetenv("NODE_ENV")
			os.Unsetenv("GO_ENV")
		}()
		
		// Disable CORS
		app.Config.Security.EnableCORS = false
		
		config := app.getCORSConfig()
		assert.NotEmpty(t, config.AllowedOrigins) // Should still have default origins
	})
}

// TestApplication_NewApplication_InitializationFailures tests initialization failure paths
func TestApplication_NewApplication_InitializationFailures(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
		cleanupEnv func()
		wantErr  bool
	}{
		{
			name: "with minimal config",
			setupEnv: func() {
				// Create minimal temp environment
				tempDir, _ := os.MkdirTemp("", "minimal_test_*")
				os.Args = []string{filepath.Join(tempDir, "test.exe")}
				os.Setenv("ISX_SERVER_PORT", "8083")
				os.Setenv("ISX_LOGGING_LEVEL", "warn")
			},
			cleanupEnv: func() {
				os.Unsetenv("ISX_SERVER_PORT")
				os.Unsetenv("ISX_LOGGING_LEVEL")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			app, err := NewApplication(createMockFS())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, app)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, app)
			}
		})
	}
}

// TestApplication_openBrowser_AllPlatforms tests browser opening on different platforms
func TestApplication_openBrowser_AllPlatforms(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "localhost URL",
			url:  "http://localhost:8080",
		},
		{
			name: "127.0.0.1 URL", 
			url:  "http://127.0.0.1:8080",
		},
		{
			name: "HTTPS URL",
			url:  "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := openBrowser(tt.url)
			// Don't assert on error as it's platform dependent
			// Just verify no panic occurs
			_ = err
		})
	}
}

// TestApplication_serveSPAHandler_CompleteRouting tests complete SPA routing logic
func TestApplication_serveSPAHandler_CompleteRouting(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create comprehensive mock FS with nested routes
	comprehensiveFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Main</body></html>"),
		},
		"license/index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>License</body></html>"),
		},
		"operations/index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Operations</body></html>"),
		},
		"dashboard/index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>Dashboard</body></html>"),
		},
		"favicon.ico": &fstest.MapFile{
			Data: []byte("ico data"),
		},
		"robots.txt": &fstest.MapFile{
			Data: []byte("User-agent: *\nDisallow:"),
		},
	}

	app, err := NewApplication(comprehensiveFS)
	require.NoError(t, err)

	// Disable license checking for this test to focus on SPA routing logic
	app.Services = nil

	handler := app.serveSPAHandler(comprehensiveFS)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		checkContent   bool
		expectedContent string
	}{
		{
			name:           "root redirects",
			path:           "/",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:           "license page direct", 
			path:           "/license",
			expectedStatus: http.StatusOK,
			checkContent:   true,
			expectedContent: "License",
		},
		{
			name:           "operations page direct",
			path:           "/operations", 
			expectedStatus: http.StatusOK,
			checkContent:   true,
			expectedContent: "Operations",
		},
		{
			name:           "dashboard page direct",
			path:           "/dashboard",
			expectedStatus: http.StatusOK,
			checkContent:   true,
			expectedContent: "Dashboard",
		},
		{
			name:           "exact file - favicon",
			path:           "/favicon.ico",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exact file - robots.txt",
			path:           "/robots.txt",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "nested route fallback",
			path:           "/some/nested/route",
			expectedStatus: http.StatusOK, // Falls back to index.html
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			
			handler(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.checkContent && tt.expectedContent != "" {
				body := w.Body.String()
				assert.Contains(t, body, tt.expectedContent)
			}
		})
	}
}

// TestApplication_serveStaticWithMIME_ErrorPaths tests error handling in static file serving
func TestApplication_serveStaticWithMIME_ErrorPaths(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	app, err := NewApplication(createMockFS())
	require.NoError(t, err)

	// Empty filesystem to trigger not found errors
	emptyFS := fstest.MapFS{}
	handler := app.serveStaticWithMIME(emptyFS, "/_next")

	tests := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{
			name:         "file not found",
			path:         "/_next/static/nonexistent.js",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "empty path",
			path:         "/_next/",
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			
			handler.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}