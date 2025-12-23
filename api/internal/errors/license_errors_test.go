package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPredefinedLicenseErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		description string
	}{
		{
			name:        "ErrLicenseExpired",
			err:         ErrLicenseExpired,
			description: "should be license expired sentinel error",
		},
		{
			name:        "ErrLicenseNotActivated",
			err:         ErrLicenseNotActivated,
			description: "should be license not activated sentinel error",
		},
		{
			name:        "ErrInvalidLicenseKey",
			err:         ErrInvalidLicenseKey,
			description: "should be invalid license key sentinel error",
		},
		{
			name:        "ErrInvalidLicenseFormat",
			err:         ErrInvalidLicenseFormat,
			description: "should be invalid license format sentinel error",
		},
		{
			name:        "ErrRateLimited",
			err:         ErrRateLimited,
			description: "should be rate limited sentinel error",
		},
		{
			name:        "ErrNetworkError",
			err:         ErrNetworkError,
			description: "should be network error sentinel error",
		},
		{
			name:        "ErrActivationFailed",
			err:         ErrActivationFailed,
			description: "should be activation failed sentinel error",
		},
		{
			name:        "ErrLicenseValidationFailed",
			err:         ErrLicenseValidationFailed,
			description: "should be license validation failed sentinel error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err, tt.description)
			assert.NotEmpty(t, tt.err.Error(), "error should have a message")
		})
	}
}

func TestProblemDetails_Render(t *testing.T) {
	tests := []struct {
		name       string
		problem    *ProblemDetails
		wantStatus int
	}{
		{
			name: "render 400 problem",
			problem: &ProblemDetails{
				Type:   "/errors/validation",
				Title:  "Validation Error",
				Status: http.StatusBadRequest,
				Detail: "Request validation failed",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "render 404 problem",
			problem: &ProblemDetails{
				Type:   "/errors/not-found",
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: "Resource not found",
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "render 500 problem",
			problem: &ProblemDetails{
				Type:   "/errors/internal",
				Title:  "Internal Server Error",
				Status: http.StatusInternalServerError,
				Detail: "An unexpected error occurred",
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			err := tt.problem.Render(w, r)
			assert.NoError(t, err)
		})
	}
}

func TestProblemDetails_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		problem  *ProblemDetails
		wantKeys []string
	}{
		{
			name: "marshal basic problem details",
			problem: &ProblemDetails{
				Type:       "/errors/validation",
				Title:      "Validation Error",
				Status:     http.StatusBadRequest,
				Detail:     "Request validation failed",
				Instance:   "/api/v1/users",
				Extensions: make(map[string]interface{}),
			},
			wantKeys: []string{"type", "title", "status", "detail", "instance"},
		},
		{
			name: "marshal problem with extensions",
			problem: &ProblemDetails{
				Type:   "/errors/license-expired",
				Title:  "License Expired",
				Status: http.StatusForbidden,
				Detail: "Your license has expired",
				Extensions: map[string]interface{}{
					"trace_id":   "12345",
					"error_code": "LICENSE_EXPIRED",
					"retry_after": 3600,
				},
			},
			wantKeys: []string{"type", "title", "status", "detail", "trace_id", "error_code", "retry_after"},
		},
		{
			name: "marshal problem without optional fields",
			problem: &ProblemDetails{
				Type:       "/errors/internal",
				Title:      "Internal Error",
				Status:     http.StatusInternalServerError,
				Extensions: make(map[string]interface{}),
			},
			wantKeys: []string{"type", "title", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.problem)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Check that all expected keys are present
			for _, key := range tt.wantKeys {
				assert.Contains(t, result, key, "Expected key %s to be present", key)
			}

			// Verify standard fields
			assert.Equal(t, tt.problem.Type, result["type"])
			assert.Equal(t, tt.problem.Title, result["title"])
			assert.Equal(t, float64(tt.problem.Status), result["status"]) // JSON numbers are float64

			// Check optional fields
			if tt.problem.Detail != "" {
				assert.Equal(t, tt.problem.Detail, result["detail"])
			}
			if tt.problem.Instance != "" {
				assert.Equal(t, tt.problem.Instance, result["instance"])
			}

			// Check extensions
			for key, expectedValue := range tt.problem.Extensions {
				// JSON unmarshaling converts all numbers to float64
				if intVal, ok := expectedValue.(int); ok {
					assert.Equal(t, float64(intVal), result[key])
				} else {
					assert.Equal(t, expectedValue, result[key])
				}
			}
		})
	}
}

func TestNewProblemDetails(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		problemType string
		title       string
		detail      string
		instance    string
	}{
		{
			name:        "create validation problem",
			status:      http.StatusBadRequest,
			problemType: "/errors/validation",
			title:       "Validation Failed",
			detail:      "Request validation failed",
			instance:    "/api/v1/users",
		},
		{
			name:        "create license problem",
			status:      http.StatusForbidden,
			problemType: "/errors/license-expired",
			title:       "License Expired",
			detail:      "Your license has expired",
			instance:    "/api/license",
		},
		{
			name:        "create minimal problem",
			status:      http.StatusInternalServerError,
			problemType: "/errors/internal",
			title:       "Internal Error",
			detail:      "",
			instance:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problem := NewProblemDetails(tt.status, tt.problemType, tt.title, tt.detail, tt.instance)

			assert.Equal(t, tt.status, problem.Status)
			assert.Equal(t, tt.problemType, problem.Type)
			assert.Equal(t, tt.title, problem.Title)
			assert.Equal(t, tt.detail, problem.Detail)
			assert.Equal(t, tt.instance, problem.Instance)
			assert.NotNil(t, problem.Extensions)
			assert.Empty(t, problem.Extensions)
		})
	}
}

func TestProblemDetails_WithExtension(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "add string extension",
			key:   "trace_id",
			value: "abc123",
		},
		{
			name:  "add integer extension",
			key:   "retry_after",
			value: 60,
		},
		{
			name:  "add boolean extension",
			key:   "retryable",
			value: true,
		},
		{
			name:  "add complex extension",
			key:   "errors",
			value: []string{"field1 required", "field2 invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problem := NewProblemDetails(
				http.StatusBadRequest,
				"/errors/test",
				"Test Error",
				"Test detail",
				"/test",
			)

			result := problem.WithExtension(tt.key, tt.value)

			// Should return the same instance
			assert.Same(t, problem, result)

			// Should have the extension
			assert.Equal(t, tt.value, result.Extensions[tt.key])
		})
	}
}

func TestProblemDetails_WithExtension_Chaining(t *testing.T) {
	t.Run("chain multiple extensions", func(t *testing.T) {
		problem := NewProblemDetails(
			http.StatusBadRequest,
			"/errors/test",
			"Test Error",
			"Test detail",
			"/test",
		)

		result := problem.
			WithExtension("trace_id", "12345").
			WithExtension("error_code", "TEST_ERROR").
			WithExtension("retry_after", 30)

		// Should be the same instance
		assert.Same(t, problem, result)

		// Should have all extensions
		assert.Equal(t, "12345", result.Extensions["trace_id"])
		assert.Equal(t, "TEST_ERROR", result.Extensions["error_code"])
		assert.Equal(t, 30, result.Extensions["retry_after"])
	})
}

func TestMapLicenseError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		traceID        string
		wantStatus     int
		wantType       string
		wantTitle      string
		wantExtensions map[string]interface{}
	}{
		{
			name:       "map license expired error",
			err:        ErrLicenseExpired,
			traceID:    "trace-123",
			wantStatus: http.StatusForbidden,
			wantType:   "/errors/license-expired",
			wantTitle:  "License Expired",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-123",
				"error_code": "LICENSE_EXPIRED",
			},
		},
		{
			name:       "map license not activated error",
			err:        ErrLicenseNotActivated,
			traceID:    "trace-456",
			wantStatus: http.StatusPreconditionRequired,
			wantType:   "/errors/license-not-activated",
			wantTitle:  "License Not Activated",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-456",
				"error_code": "LICENSE_NOT_ACTIVATED",
			},
		},
		{
			name:       "map invalid license key error",
			err:        ErrInvalidLicenseKey,
			traceID:    "trace-789",
			wantStatus: http.StatusBadRequest,
			wantType:   "/errors/invalid-license-key",
			wantTitle:  "Invalid License Key",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-789",
				"error_code": "INVALID_LICENSE_KEY",
			},
		},
		{
			name:       "map invalid license format error",
			err:        ErrInvalidLicenseFormat,
			traceID:    "trace-abc",
			wantStatus: http.StatusBadRequest,
			wantType:   "/errors/invalid-license-format",
			wantTitle:  "Invalid License Format",
			wantExtensions: map[string]interface{}{
				"trace_id":        "trace-abc",
				"error_code":      "INVALID_LICENSE_FORMAT",
				"expected_format": "ISX1Y-XXXXX-XXXXX-XXXXX-XXXXX",
			},
		},
		{
			name:       "map activation failed error",
			err:        ErrActivationFailed,
			traceID:    "trace-def",
			wantStatus: http.StatusUnprocessableEntity,
			wantType:   "/errors/activation-failed",
			wantTitle:  "License Activation Failed",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-def",
				"error_code": "ACTIVATION_FAILED",
			},
		},
		{
			name:       "map validation failed error",
			err:        ErrLicenseValidationFailed,
			traceID:    "trace-ghi",
			wantStatus: http.StatusInternalServerError,
			wantType:   "/errors/validation-failed",
			wantTitle:  "License Validation Failed",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-ghi",
				"error_code": "VALIDATION_FAILED",
			},
		},
		{
			name:       "map rate limited error",
			err:        ErrRateLimited,
			traceID:    "trace-jkl",
			wantStatus: http.StatusTooManyRequests,
			wantType:   "/errors/rate-limited",
			wantTitle:  "Too Many Requests",
			wantExtensions: map[string]interface{}{
				"trace_id":    "trace-jkl",
				"error_code":  "RATE_LIMITED",
				"retry_after": 900,
			},
		},
		{
			name:       "map network error",
			err:        ErrNetworkError,
			traceID:    "trace-mno",
			wantStatus: http.StatusServiceUnavailable,
			wantType:   "/errors/network-error",
			wantTitle:  "Network Error",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-mno",
				"error_code": "NETWORK_ERROR",
			},
		},
		{
			name:       "map generic error",
			err:        fmt.Errorf("unknown error"),
			traceID:    "trace-xyz",
			wantStatus: http.StatusInternalServerError,
			wantType:   "/errors/internal-error",
			wantTitle:  "Internal Server Error",
			wantExtensions: map[string]interface{}{
				"trace_id":   "trace-xyz",
				"error_code": "INTERNAL_ERROR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := MapLicenseError(tt.err, tt.traceID)

			// Should return a ProblemDetails
			problem, ok := renderer.(*ProblemDetails)
			require.True(t, ok, "Expected ProblemDetails type")

			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, tt.wantType, problem.Type)
			assert.Equal(t, tt.wantTitle, problem.Title)

			// Check extensions
			for key, expectedValue := range tt.wantExtensions {
				assert.Equal(t, expectedValue, problem.Extensions[key], "Extension %s mismatch", key)
			}
		})
	}
}

func TestMapLicenseError_APIError(t *testing.T) {
	tests := []struct {
		name       string
		apiError   *APIError
		traceID    string
		wantStatus int
		wantType   string
	}{
		{
			name: "map LICENSE_NOT_FOUND APIError",
			apiError: &APIError{
				StatusCode: http.StatusNotFound,
				ErrorCode:  "LICENSE_NOT_FOUND",
				Message:    "License file not found",
			},
			traceID:    "trace-api-123",
			wantStatus: http.StatusNotFound,
			wantType:   "/errors/license-not-found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := MapLicenseError(tt.apiError, tt.traceID)

			problem, ok := renderer.(*ProblemDetails)
			require.True(t, ok, "Expected ProblemDetails type")

			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, tt.wantType, problem.Type)
			assert.Equal(t, tt.traceID, problem.Extensions["trace_id"])
			assert.Equal(t, tt.apiError.ErrorCode, problem.Extensions["error_code"])
		})
	}
}

func TestMapLicenseError_WrappedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		traceID    string
		wantStatus int
		wantType   string
	}{
		{
			name:       "wrapped license expired error",
			err:        fmt.Errorf("context: %w", ErrLicenseExpired),
			traceID:    "trace-wrapped-123",
			wantStatus: http.StatusForbidden,
			wantType:   "/errors/license-expired",
		},
		{
			name:       "deeply wrapped error",
			err:        fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrInvalidLicenseKey)),
			traceID:    "trace-deep-456",
			wantStatus: http.StatusBadRequest,
			wantType:   "/errors/invalid-license-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := MapLicenseError(tt.err, tt.traceID)

			problem, ok := renderer.(*ProblemDetails)
			require.True(t, ok, "Expected ProblemDetails type")

			assert.Equal(t, tt.wantStatus, problem.Status)
			assert.Equal(t, tt.wantType, problem.Type)
			assert.Equal(t, tt.traceID, problem.Extensions["trace_id"])
		})
	}
}

func TestProblemDetails_RFC7807Compliance(t *testing.T) {
	t.Run("RFC 7807 compliance test", func(t *testing.T) {
		problem := NewProblemDetails(
			http.StatusBadRequest,
			"https://example.com/probs/validation-error",
			"Your request parameters didn't validate.",
			"The request body must contain a valid JSON object.",
			"/users/create",
		).WithExtension("invalid_params", []map[string]string{
			{"name": "email", "reason": "invalid format"},
			{"name": "age", "reason": "must be positive"},
		})

		// Test JSON serialization
		data, err := json.Marshal(problem)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		// RFC 7807 required fields
		assert.Equal(t, "https://example.com/probs/validation-error", result["type"])
		assert.Equal(t, "Your request parameters didn't validate.", result["title"])
		assert.Equal(t, float64(http.StatusBadRequest), result["status"])
		assert.Equal(t, "The request body must contain a valid JSON object.", result["detail"])
		assert.Equal(t, "/users/create", result["instance"])

		// Extension field
		assert.Contains(t, result, "invalid_params")
	})
}

func TestProblemDetails_RenderIntegration(t *testing.T) {
	t.Run("integration with chi render", func(t *testing.T) {
		problem := NewProblemDetails(
			http.StatusForbidden,
			"/errors/license-expired",
			"License Expired",
			"Your license has expired and needs to be renewed",
			"/api/license",
		).WithExtension("trace_id", "test-123").
			WithExtension("error_code", "LICENSE_EXPIRED")

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/license", nil)

		err := render.Render(w, r, problem)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, w.Code)
		// Chi render package does not include charset in Content-Type header by default
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "/errors/license-expired", response["type"])
		assert.Equal(t, "License Expired", response["title"])
		assert.Equal(t, float64(http.StatusForbidden), response["status"])
		assert.Equal(t, "test-123", response["trace_id"])
		assert.Equal(t, "LICENSE_EXPIRED", response["error_code"])
	})
}

func TestProblemDetails_EmptyExtensions(t *testing.T) {
	t.Run("problem with no extensions", func(t *testing.T) {
		problem := NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/internal",
			"Internal Server Error",
			"An unexpected error occurred",
			"/api/test",
		)

		data, err := json.Marshal(problem)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		// Should only have standard RFC 7807 fields
		expectedKeys := []string{"type", "title", "status", "detail", "instance"}
		assert.Len(t, result, len(expectedKeys))

		for _, key := range expectedKeys {
			assert.Contains(t, result, key)
		}
	})
}

func TestProblemDetails_NilExtensions(t *testing.T) {
	t.Run("problem with nil extensions map", func(t *testing.T) {
		problem := &ProblemDetails{
			Type:       "/errors/test",
			Title:      "Test Error",
			Status:     http.StatusBadRequest,
			Detail:     "Test detail",
			Instance:   "/test",
			Extensions: nil,
		}

		// Should not panic when marshaling
		data, err := json.Marshal(problem)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "/errors/test", result["type"])
		assert.Equal(t, "Test Error", result["title"])
	})
}

func TestMapLicenseError_ErrorsIsAs(t *testing.T) {
	t.Run("errors.Is works correctly", func(t *testing.T) {
		// Test direct error
		renderer := MapLicenseError(ErrLicenseExpired, "trace-123")
		problem := renderer.(*ProblemDetails)
		assert.Equal(t, "LICENSE_EXPIRED", problem.Extensions["error_code"])

		// Test wrapped error
		wrappedErr := fmt.Errorf("license check failed: %w", ErrLicenseExpired)
		renderer2 := MapLicenseError(wrappedErr, "trace-456")
		problem2 := renderer2.(*ProblemDetails)
		assert.Equal(t, "LICENSE_EXPIRED", problem2.Extensions["error_code"])
	})

	t.Run("errors.As works with APIError", func(t *testing.T) {
		apiErr := &APIError{
			StatusCode: http.StatusNotFound,
			ErrorCode:  "LICENSE_NOT_FOUND",
			Message:    "License file not found",
		}
		wrappedErr := fmt.Errorf("failed to load license: %w", apiErr)

		renderer := MapLicenseError(wrappedErr, "trace-789")
		problem := renderer.(*ProblemDetails)
		assert.Equal(t, "LICENSE_NOT_FOUND", problem.Extensions["error_code"])
	})
}