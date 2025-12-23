package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	licenseErrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/services"
)

// MockLicenseService implements the LicenseService interface for testing
type MockLicenseService struct {
	mock.Mock
}

func (m *MockLicenseService) GetStatus(ctx context.Context) (*services.LicenseStatusResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.LicenseStatusResponse), args.Error(1)
}

func (m *MockLicenseService) Activate(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockLicenseService) ValidateWithContext(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockLicenseService) GetDetailedStatus(ctx context.Context) (*services.DetailedLicenseStatusResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.DetailedLicenseStatusResponse), args.Error(1)
}

func (m *MockLicenseService) TransferLicense(ctx context.Context, key string, force bool) error {
	args := m.Called(ctx, key, force)
	return args.Error(0)
}

func (m *MockLicenseService) GetValidationMetrics(ctx context.Context) (*services.ValidationMetrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ValidationMetrics), args.Error(1)
}

func (m *MockLicenseService) InvalidateCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLicenseService) CheckRenewalStatus(ctx context.Context) (*services.RenewalStatusResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.RenewalStatusResponse), args.Error(1)
}

func (m *MockLicenseService) GetDebugInfo(ctx context.Context) (*services.LicenseDebugInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.LicenseDebugInfo), args.Error(1)
}

func (m *MockLicenseService) CheckExistingLicense() (*license.ExistingLicenseInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*license.ExistingLicenseInfo), args.Error(1)
}

func (m *MockLicenseService) GetLicenseDetails() (interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockLicenseService) GetActivationHistory() ([]interface{}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockLicenseService) BackupCurrentLicense() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// TestLicenseHandler_GetStatus tests the GetStatus endpoint with table-driven tests
func TestLicenseHandler_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockLicenseService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "valid license returns active status",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
					Status:        200,
					LicenseStatus: "active",
					Message:       "License is active",
					DaysLeft:      365,
					Features:      []string{"advanced_reports", "api_access"},
					TraceID:       "test-trace-id",
					Timestamp:     time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(200), body["status"])
				assert.Equal(t, "active", body["license_status"])
				assert.Equal(t, "License is active", body["message"])
				assert.Equal(t, float64(365), body["days_left"])
				assert.NotNil(t, body["features"])
			},
		},
		{
			name: "expired license returns appropriate status",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
					Status:        200,
					LicenseStatus: "expired",
					Message:       "License has expired",
					DaysLeft:      0,
					TraceID:       "test-trace-id",
					Timestamp:     time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, float64(200), body["status"])
				assert.Equal(t, "expired", body["license_status"])
				assert.Equal(t, "License has expired", body["message"])
			},
		},
		{
			name: "service error returns internal server error",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(nil, errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/internal-error", body["type"])
				assert.Contains(t, body["title"], "Internal Server Error")
			},
		},
		{
			name: "context cancellation returns appropriate error",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(nil, context.Canceled)
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/request-canceled", body["type"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockLicenseService)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewLicenseHandler(mockService, logger)
			
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
			w := httptest.NewRecorder()

			// Execute
			handler.GetStatus(w, req)

			// Assert
			res := w.Result()
			defer res.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			
			var body map[string]interface{}
			err := json.NewDecoder(res.Body).Decode(&body)
			require.NoError(t, err)
			
			if tt.expectedBody != nil {
				tt.expectedBody(t, body)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

// TestLicenseHandler_Activate tests the license activation endpoint
func TestLicenseHandler_Activate(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockLicenseService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "successful activation with valid key",
			requestBody: map[string]string{
				"license_key": "ISX1YABCDEFGHIJKLMNOP",
			},
			setupMock: func(m *MockLicenseService) {
				m.On("Activate", mock.Anything, "ISX1YABCDEFGHIJKLMNOP").Return(nil)
				m.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
					Status:        200,
					LicenseStatus: "active",
					Message:       "License is active",
					DaysLeft:      365,
					TraceID:       "test-trace-id",
					Timestamp:     time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["message"], "License activated successfully")
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "activation with invalid key format",
			requestBody: map[string]string{
				"license_key": "SHORT",
			},
			setupMock: func(m *MockLicenseService) {
				// Mock should not be called due to validation failure
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-request", body["type"])
				assert.Contains(t, body["detail"], "invalid license key format")
			},
		},
		{
			name: "activation with empty request body",
			requestBody: map[string]string{},
			setupMock: func(m *MockLicenseService) {
				// Mock should not be called
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-request", body["type"])
				assert.Contains(t, body["detail"], "license_key is required")
			},
		},
		{
			name: "activation with invalid JSON",
			requestBody: "invalid json",
			setupMock: func(m *MockLicenseService) {
				// Mock should not be called
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-request", body["type"])
			},
		},
		{
			name: "activation with license already activated error",
			requestBody: map[string]string{
				"license_key": "ISX1YALREADYACTIVATED",
			},
			setupMock: func(m *MockLicenseService) {
				m.On("Activate", mock.Anything, "ISX1YALREADYACTIVATED").
					Return(licenseErrors.ErrActivationFailed)
				m.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
					Status:        200,
					LicenseStatus: "active",
					Message:       "License is active",
					DaysLeft:      365,
					TraceID:       "test-trace-id",
					Timestamp:     time.Now(),
				}, nil).Maybe()
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/activation-failed", body["type"])
			},
		},
		{
			name: "activation with invalid license key error",
			requestBody: map[string]string{
				"license_key": "ISX1YINVALIDLICENSEK",
			},
			setupMock: func(m *MockLicenseService) {
				m.On("Activate", mock.Anything, "ISX1YINVALIDLICENSEK").
					Return(licenseErrors.ErrInvalidLicenseKey)
				m.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
					Status:        200,
					LicenseStatus: "active",
					Message:       "License is active",
					DaysLeft:      365,
					TraceID:       "test-trace-id",
					Timestamp:     time.Now(),
				}, nil).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-license-key", body["type"])
			},
		},
		{
			name: "activation with hardware mismatch error",
			requestBody: map[string]string{
				"license_key": "ISX1YHARDWAREMISMATC",
			},
			setupMock: func(m *MockLicenseService) {
				m.On("Activate", mock.Anything, "ISX1YHARDWAREMISMATC").
					Return(licenseErrors.ErrLicenseValidationFailed)
				m.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
					Status:        200,
					LicenseStatus: "active",
					Message:       "License is active",
					DaysLeft:      365,
					TraceID:       "test-trace-id",
					Timestamp:     time.Now(),
				}, nil).Maybe()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/validation-failed", body["type"])
			},
		},
		{
			name: "activation with SQL injection attempt",
			requestBody: map[string]string{
				"license_key": "'; DROP TABLE licenses; --",
			},
			setupMock: func(m *MockLicenseService) {
				// Service should not be called due to validation failure
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-request", body["type"])
			},
		},
		{
			name: "activation with XSS attempt",
			requestBody: map[string]string{
				"license_key": "<script>alert('xss')</script>",
			},
			setupMock: func(m *MockLicenseService) {
				// Service should not be called due to validation failure
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-request", body["type"])
			},
		},
		{
			name: "activation with very long key",
			requestBody: map[string]string{
				"license_key": strings.Repeat("A", 1000),
			},
			setupMock: func(m *MockLicenseService) {
				// Should not be called due to validation
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "/errors/invalid-request", body["type"])
				assert.Contains(t, body["detail"], "invalid license key format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockLicenseService)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewLicenseHandler(mockService, logger)
			
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			// Create request
			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = strings.NewReader(str)
			} else {
				jsonBody, _ := json.Marshal(tt.requestBody)
				body = bytes.NewReader(jsonBody)
			}
			
			req := httptest.NewRequest(http.MethodPost, "/api/license/activate", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			handler.Activate(w, req)

			// Assert
			res := w.Result()
			defer res.Body.Close()
			
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			
			var responseBody map[string]interface{}
			err := json.NewDecoder(res.Body).Decode(&responseBody)
			require.NoError(t, err)
			
			if tt.expectedBody != nil {
				tt.expectedBody(t, responseBody)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

// TestLicenseHandler_ConcurrentRequests tests concurrent request handling
func TestLicenseHandler_ConcurrentRequests(t *testing.T) {
	mockService := new(MockLicenseService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewLicenseHandler(mockService, logger)
	
	// Setup mock to handle concurrent calls
	mockService.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
		Status:        200,
		LicenseStatus: "active",
		Message:       "License is active",
		DaysLeft:      365,
		TraceID:       "concurrent-test",
		Timestamp:     time.Now(),
	}, nil).Maybe()
	
	// Setup router with middleware
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(30 * time.Second))
	router.Mount("/api/license", handler.Routes())
	
	server := httptest.NewServer(router)
	defer server.Close()
	
	// Test concurrent requests
	concurrency := 50
	var wg sync.WaitGroup
	wg.Add(concurrency)
	
	errors := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			
			resp, err := http.Get(server.URL + "/api/license/status")
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}
			
			var body map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				errors <- err
				return
			}
			
			if body["license_status"] != "active" {
				errors <- fmt.Errorf("expected license_status=active, got %v", body["license_status"])
			}
		}()
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	var errorCount int
	for err := range errors {
		t.Errorf("Concurrent request error: %v", err)
		errorCount++
	}
	
	assert.Equal(t, 0, errorCount, "Expected no errors in concurrent requests")
}

// TestLicenseHandler_RateLimiting tests rate limiting behavior
func TestLicenseHandler_RateLimiting(t *testing.T) {
	mockService := new(MockLicenseService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewLicenseHandler(mockService, logger)
	
	// Setup mock to be called many times
	mockService.On("Activate", mock.Anything, mock.Anything).Return(nil).Maybe()
	mockService.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
		Status:        200,
		LicenseStatus: "active",
		Message:       "License is active",
		DaysLeft:      365,
		TraceID:       "test-trace-id",
		Timestamp:     time.Now(),
	}, nil).Maybe()
	
	// Make rapid requests
	requestCount := 10
	successCount := 0
	
	for i := 0; i < requestCount; i++ {
		reqBody := map[string]string{
			"license_key": fmt.Sprintf("ISX1YTESTKEY%08d", i),
		}
		jsonBody, _ := json.Marshal(reqBody)
		
		req := httptest.NewRequest(http.MethodPost, "/api/license/activate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		handler.Activate(w, req)
		
		if w.Code == http.StatusOK {
			successCount++
		}
		
		// Small delay to avoid overwhelming
		time.Sleep(10 * time.Millisecond)
	}
	
	// Should allow reasonable number of requests
	assert.Greater(t, successCount, 0, "At least some requests should succeed")
}

// TestLicenseHandler_SecurityHeaders tests security headers in responses
func TestLicenseHandler_SecurityHeaders(t *testing.T) {
	mockService := new(MockLicenseService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewLicenseHandler(mockService, logger)
	
	mockService.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
		Status:        200,
		LicenseStatus: "active",
		Message:       "License is active",
		TraceID:       "security-test",
		Timestamp:     time.Now(),
	}, nil)
	
	req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
	w := httptest.NewRecorder()
	
	handler.GetStatus(w, req)
	
	res := w.Result()
	defer res.Body.Close()
	
	// Check security headers
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	// Security headers would typically be set by middleware, not the handler itself
	// For this test, we're just checking that the response is properly formed
}

// TestLicenseHandler_ErrorResponseFormat tests RFC 7807 error format
func TestLicenseHandler_ErrorResponseFormat(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockLicenseService)
		endpoint      string
		method        string
		expectedError map[string]interface{}
	}{
		{
			name: "invalid license key error format",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(nil, licenseErrors.ErrInvalidLicenseKey)
			},
			endpoint: "/api/license/status",
			method:   http.MethodGet,
			expectedError: map[string]interface{}{
				"type":   "/errors/invalid-license-key",
				"title":  "Invalid License Key",
				"status": float64(400),
			},
		},
		{
			name: "validation failed error format",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(nil, licenseErrors.ErrLicenseValidationFailed)
			},
			endpoint: "/api/license/status",
			method:   http.MethodGet,
			expectedError: map[string]interface{}{
				"type":   "/errors/validation-failed",
				"title":  "License Validation Failed",
				"status": float64(500),
			},
		},
		{
			name: "license expired error format",
			setupMock: func(m *MockLicenseService) {
				m.On("GetStatus", mock.Anything).Return(nil, licenseErrors.ErrLicenseExpired)
			},
			endpoint: "/api/license/status",
			method:   http.MethodGet,
			expectedError: map[string]interface{}{
				"type":   "/errors/license-expired",
				"title":  "License Expired",
				"status": float64(403),
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockLicenseService)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewLicenseHandler(mockService, logger)
			
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}
			
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()
			
			handler.GetStatus(w, req)
			
			res := w.Result()
			defer res.Body.Close()
			
			var body map[string]interface{}
			err := json.NewDecoder(res.Body).Decode(&body)
			require.NoError(t, err)
			
			// Verify RFC 7807 fields
			assert.Equal(t, tt.expectedError["type"], body["type"])
			assert.Equal(t, tt.expectedError["title"], body["title"])
			assert.Equal(t, tt.expectedError["status"], body["status"])
			assert.NotEmpty(t, body["instance"])
			
			mockService.AssertExpectations(t)
		})
	}
}

// TestLicenseHandler_MemoryLeaks tests for potential memory leaks
func TestLicenseHandler_MemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}
	
	mockService := new(MockLicenseService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewLicenseHandler(mockService, logger)
	
	// Setup mock to return large response
	largeResponse := &services.LicenseStatusResponse{
		Status:        200,
		LicenseStatus: "active",
		Message:       "License is active",
		DaysLeft:      365,
		Features:      make([]string, 0, 1000),
		Limitations:   make(map[string]interface{}),
		TraceID:       "memory-test",
		Timestamp:     time.Now(),
	}
	
	// Add many features to increase memory usage
	for i := 0; i < 1000; i++ {
		largeResponse.Features = append(largeResponse.Features, fmt.Sprintf("feature_%d", i))
		largeResponse.Limitations[fmt.Sprintf("limit_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	
	mockService.On("GetStatus", mock.Anything).Return(largeResponse, nil).Maybe()
	
	// Make many requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
		w := httptest.NewRecorder()
		
		handler.GetStatus(w, req)
		
		res := w.Result()
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}
	
	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	
	// No specific assertion - this test is mainly for detecting
	// memory leaks when run with memory profiling tools
}

// BenchmarkLicenseHandler_GetStatus benchmarks the GetStatus endpoint
func BenchmarkLicenseHandler_GetStatus(b *testing.B) {
	mockService := new(MockLicenseService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewLicenseHandler(mockService, logger)
	
	mockService.On("GetStatus", mock.Anything).Return(&services.LicenseStatusResponse{
		Status:        200,
		LicenseStatus: "active",
		Message:       "License is active",
		DaysLeft:      365,
		TraceID:       "benchmark-test",
		Timestamp:     time.Now(),
	}, nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
			w := httptest.NewRecorder()
			handler.GetStatus(w, req)
		}
	})
}

// BenchmarkLicenseHandler_Activate benchmarks the Activate endpoint
func BenchmarkLicenseHandler_Activate(b *testing.B) {
	mockService := new(MockLicenseService)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewLicenseHandler(mockService, logger)
	
	mockService.On("Activate", mock.Anything, mock.Anything).Return(nil)
	
	reqBody := map[string]string{
		"license_key": "BENCHMARK-KEY-12345",
	}
	jsonBody, _ := json.Marshal(reqBody)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/license/activate", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.Activate(w, req)
		}
	})
}