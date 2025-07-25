package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// mockLicenseManager is a mock implementation of license.Manager for testing
type mockLicenseManager struct {
	validateFunc func() (bool, error)
}

func (m *mockLicenseManager) ValidateLicense() (bool, error) {
	if m.validateFunc != nil {
		return m.validateFunc()
	}
	return true, nil
}

// Other methods would be implemented as needed for the interface

// TestLicenseValidator tests the license validation middleware
func TestLicenseValidator(t *testing.T) {
	// Create a test logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		path           string
		validateFunc   func() (bool, error)
		wantStatusCode int
		wantNextCalled bool
	}{
		{
			name: "excluded path - root",
			path: "/",
			validateFunc: func() (bool, error) {
				t.Error("ValidateLicense should not be called for excluded paths")
				return false, nil
			},
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name: "excluded path - license page",
			path: "/license",
			validateFunc: func() (bool, error) {
				t.Error("ValidateLicense should not be called for excluded paths")
				return false, nil
			},
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name: "excluded path - static files",
			path: "/static/css/style.css",
			validateFunc: func() (bool, error) {
				t.Error("ValidateLicense should not be called for excluded paths")
				return false, nil
			},
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name: "excluded path - health check",
			path: "/api/health",
			validateFunc: func() (bool, error) {
				t.Error("ValidateLicense should not be called for excluded paths")
				return false, nil
			},
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name: "valid license",
			path: "/api/data",
			validateFunc: func() (bool, error) {
				return true, nil
			},
			wantStatusCode: http.StatusOK,
			wantNextCalled: true,
		},
		{
			name: "invalid license",
			path: "/api/data",
			validateFunc: func() (bool, error) {
				return false, nil
			},
			wantStatusCode: http.StatusPreconditionRequired,
			wantNextCalled: false,
		},
		{
			name: "license validation error",
			path: "/api/data",
			validateFunc: func() (bool, error) {
				return false, errors.New("network error")
			},
			wantStatusCode: http.StatusServiceUnavailable,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock manager
			mockManager := &mockLicenseManager{
				validateFunc: tt.validateFunc,
			}

			// Create validator
			validator := NewLicenseValidator(mockManager, logger)

			// Create test handler
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			// Execute middleware
			handler := validator.Handler(nextHandler)
			handler.ServeHTTP(rec, req)

			// Check results
			if rec.Code != tt.wantStatusCode {
				t.Errorf("Response code = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if nextCalled != tt.wantNextCalled {
				t.Errorf("Next handler called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
		})
	}
}

// TestLicenseValidatorCache tests the caching functionality
func TestLicenseValidatorCache(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validateCallCount := 0

	mockManager := &mockLicenseManager{
		validateFunc: func() (bool, error) {
			validateCallCount++
			return true, nil
		},
	}

	validator := NewLicenseValidator(mockManager, logger)
	validator.SetCacheTTL(100 * time.Millisecond) // Short TTL for testing

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := validator.Handler(nextHandler)

	// First request - should call validate
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if validateCallCount != 1 {
		t.Errorf("First request: validateCallCount = %v, want 1", validateCallCount)
	}

	// Second request immediately - should use cache
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if validateCallCount != 1 {
		t.Errorf("Second request: validateCallCount = %v, want 1 (cached)", validateCallCount)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third request - should call validate again
	req3 := httptest.NewRequest("GET", "/api/data", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if validateCallCount != 2 {
		t.Errorf("Third request: validateCallCount = %v, want 2 (cache expired)", validateCallCount)
	}
}

// TestLicenseValidatorInvalidateCache tests cache invalidation
func TestLicenseValidatorInvalidateCache(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	validateCallCount := 0

	mockManager := &mockLicenseManager{
		validateFunc: func() (bool, error) {
			validateCallCount++
			return true, nil
		},
	}

	validator := NewLicenseValidator(mockManager, logger)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := validator.Handler(nextHandler)

	// First request
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if validateCallCount != 1 {
		t.Errorf("First request: validateCallCount = %v, want 1", validateCallCount)
	}

	// Invalidate cache
	validator.InvalidateCache()

	// Second request - should call validate again despite cache
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if validateCallCount != 2 {
		t.Errorf("Second request after invalidation: validateCallCount = %v, want 2", validateCallCount)
	}
}

// TestLicenseValidatorCustomExcludes tests custom path exclusions
func TestLicenseValidatorCustomExcludes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	mockManager := &mockLicenseManager{
		validateFunc: func() (bool, error) {
			return false, nil // Would fail if called
		},
	}

	validator := NewLicenseValidator(mockManager, logger)
	
	// Add custom exclusions
	validator.AddExcludePath("/custom/path")
	validator.AddExcludePrefix("/api/public/")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := validator.Handler(nextHandler)

	tests := []struct {
		path       string
		shouldPass bool
	}{
		{"/custom/path", true},
		{"/api/public/endpoint", true},
		{"/api/private/endpoint", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tt.shouldPass && rec.Code != http.StatusOK {
				t.Errorf("Path %s: expected to pass, got status %v", tt.path, rec.Code)
			}
			if !tt.shouldPass && rec.Code == http.StatusOK {
				t.Errorf("Path %s: expected to fail, but passed", tt.path)
			}
		})
	}
}

// TestLicenseValidatorWithRouter tests the middleware with a real Chi router
func TestLicenseValidatorWithRouter(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	mockManager := &mockLicenseManager{
		validateFunc: func() (bool, error) {
			return true, nil
		},
	}

	validator := NewLicenseValidator(mockManager, logger)

	// Create Chi router
	r := chi.NewRouter()
	r.Use(validator.Handler)

	// Define routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("home"))
	})

	r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	})

	r.Get("/license", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("license"))
	})

	// Test requests
	tests := []struct {
		path         string
		wantStatus   int
		wantBody     string
	}{
		{"/", http.StatusOK, "home"},
		{"/api/data", http.StatusOK, "data"},
		{"/license", http.StatusOK, "license"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %v, want %v", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("Body = %v, want %v", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

// TestLicenseValidatorTimeout tests timeout handling
func TestLicenseValidatorTimeout(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	mockManager := &mockLicenseManager{
		validateFunc: func() (bool, error) {
			// Simulate a slow validation
			time.Sleep(10 * time.Second)
			return true, nil
		},
	}

	validator := NewLicenseValidator(mockManager, logger)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := validator.Handler(nextHandler)

	req := httptest.NewRequest("GET", "/api/data", nil)
	rec := httptest.NewRecorder()

	// This should timeout after 5 seconds (as defined in validateLicense method)
	start := time.Now()
	handler.ServeHTTP(rec, req)
	duration := time.Since(start)

	// Should timeout within 6 seconds (5s timeout + some overhead)
	if duration > 6*time.Second {
		t.Errorf("Request took too long: %v", duration)
	}

	// Should return service unavailable due to timeout
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %v, want %v", rec.Code, http.StatusServiceUnavailable)
	}
}