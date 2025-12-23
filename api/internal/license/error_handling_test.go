package license

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Error Handling and API Response Tests
// =============================================================================

func TestErrorHandling(t *testing.T) {
	t.Run("ErrNetwork", func(t *testing.T) {
		testErr := errors.New("test network error")
		errResp := ErrNetwork(testErr)
		assert.NotNil(t, errResp)
		assert.Equal(t, http.StatusServiceUnavailable, errResp.HTTPStatusCode)
		assert.Equal(t, ErrCodeNetworkError, errResp.AppCode)
		assert.Contains(t, errResp.ErrorText, "Unable to connect to license server")
	})

	t.Run("ErrInvalidRequest", func(t *testing.T) {
		errResp := ErrInvalidRequest("invalid license key format")
		assert.NotNil(t, errResp)
		assert.Equal(t, http.StatusBadRequest, errResp.HTTPStatusCode)
		assert.Equal(t, "invalid license key format", errResp.ErrorText)
	})

	t.Run("ErrInternal", func(t *testing.T) {
		testErr := errors.New("database connection failed")
		errResp := ErrInternal(testErr)
		assert.NotNil(t, errResp)
		assert.Equal(t, http.StatusInternalServerError, errResp.HTTPStatusCode)
		assert.Equal(t, testErr, errResp.Err)
		assert.Contains(t, errResp.ErrorText, "An unexpected error occurred")
	})

	t.Run("NewErrResponse", func(t *testing.T) {
		errResp := NewErrResponse(http.StatusBadRequest, "TEST_ERROR", "Test error message")
		assert.NotNil(t, errResp)
		assert.Equal(t, http.StatusBadRequest, errResp.HTTPStatusCode)
		assert.Equal(t, "TEST_ERROR", errResp.AppCode)
		assert.Equal(t, "Test error message", errResp.ErrorText)
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("ErrInvalidLicenseKey", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidLicenseKey)
		assert.Equal(t, http.StatusBadRequest, ErrInvalidLicenseKey.HTTPStatusCode)
		assert.Equal(t, ErrCodeInvalidKey, ErrInvalidLicenseKey.AppCode)
	})

	t.Run("ErrLicenseExpired", func(t *testing.T) {
		assert.NotNil(t, ErrLicenseExpired)
		assert.Equal(t, http.StatusForbidden, ErrLicenseExpired.HTTPStatusCode)
		assert.Equal(t, ErrCodeExpired, ErrLicenseExpired.AppCode)
	})

	t.Run("ErrNotActivated", func(t *testing.T) {
		assert.NotNil(t, ErrNotActivated)
		assert.Equal(t, http.StatusPreconditionRequired, ErrNotActivated.HTTPStatusCode) // 428 - license activation required
		assert.Equal(t, ErrCodeNotActivated, ErrNotActivated.AppCode)
	})

	t.Run("ErrRateLimited", func(t *testing.T) {
		assert.NotNil(t, ErrRateLimited)
		assert.Equal(t, http.StatusTooManyRequests, ErrRateLimited.HTTPStatusCode)
		assert.Equal(t, ErrCodeRateLimited, ErrRateLimited.AppCode)
	})
}

func TestErrorResponseRendering(t *testing.T) {
	t.Run("Render error response", func(t *testing.T) {
		errResp := NewErrResponse(http.StatusBadRequest, "VALIDATION_FAILED", "License validation failed")

		// Create a mock ResponseWriter to test rendering
		recorder := &mockResponseWriter{
			headers: make(http.Header),
		}

		// Create a minimal request to prevent nil pointer dereference
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)

		err = errResp.Render(recorder, req)
		assert.NoError(t, err)
		// Note: The actual status is set via render.Status, which we can't easily test here
		// But the function should not return an error
	})
}

// Mock ResponseWriter for testing
type mockResponseWriter struct {
	headers    http.Header
	body       string
	statusCode int
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body += string(data)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}
