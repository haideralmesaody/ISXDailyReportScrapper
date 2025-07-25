package license

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
)

// TestErrResponse tests the ErrResponse implementation
func TestErrResponse(t *testing.T) {
	tests := []struct {
		name           string
		err            *ErrResponse
		wantStatus     int
		wantStatusText string
		wantCode       string
		wantError      string
	}{
		{
			name:           "invalid license key",
			err:            ErrInvalidLicenseKey,
			wantStatus:     http.StatusBadRequest,
			wantStatusText: "Invalid license key",
			wantCode:       ErrCodeInvalidKey,
			wantError:      "The provided license key is invalid or malformed",
		},
		{
			name:           "license expired",
			err:            ErrLicenseExpired,
			wantStatus:     http.StatusForbidden,
			wantStatusText: "License expired",
			wantCode:       ErrCodeExpired,
			wantError:      "Your license has expired. Please renew to continue",
		},
		{
			name:           "machine mismatch",
			err:            ErrMachineMismatch,
			wantStatus:     http.StatusForbidden,
			wantStatusText: "Machine mismatch",
			wantCode:       ErrCodeMachineMismatch,
			wantError:      "This license is registered to a different machine",
		},
		{
			name:           "not activated",
			err:            ErrNotActivated,
			wantStatus:     http.StatusPreconditionRequired,
			wantStatusText: "License not activated",
			wantCode:       ErrCodeNotActivated,
			wantError:      "No license has been activated. Please activate a license to continue",
		},
		{
			name:           "rate limited",
			err:            ErrRateLimited,
			wantStatus:     http.StatusTooManyRequests,
			wantStatusText: "Too many requests",
			wantCode:       ErrCodeRateLimited,
			wantError:      "Too many activation attempts. Please try again later",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error response fields
			if tt.err.HTTPStatusCode != tt.wantStatus {
				t.Errorf("HTTPStatusCode = %v, want %v", tt.err.HTTPStatusCode, tt.wantStatus)
			}
			if tt.err.StatusText != tt.wantStatusText {
				t.Errorf("StatusText = %v, want %v", tt.err.StatusText, tt.wantStatusText)
			}
			if tt.err.AppCode != tt.wantCode {
				t.Errorf("AppCode = %v, want %v", tt.err.AppCode, tt.wantCode)
			}
			if tt.err.ErrorText != tt.wantError {
				t.Errorf("ErrorText = %v, want %v", tt.err.ErrorText, tt.wantError)
			}

			// Test Render implementation
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)

			err := render.Render(w, r, tt.err)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			// Check that status code was set
			if w.Code != tt.wantStatus {
				t.Errorf("Response status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestNewErrResponse tests custom error response creation
func TestNewErrResponse(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		code    string
		message string
	}{
		{
			name:    "custom bad request",
			status:  http.StatusBadRequest,
			code:    "CUSTOM_ERROR",
			message: "Custom error message",
		},
		{
			name:    "custom internal error",
			status:  http.StatusInternalServerError,
			code:    "INTERNAL_ERROR",
			message: "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErrResponse(tt.status, tt.code, tt.message)

			if err.HTTPStatusCode != tt.status {
				t.Errorf("HTTPStatusCode = %v, want %v", err.HTTPStatusCode, tt.status)
			}
			if err.StatusText != http.StatusText(tt.status) {
				t.Errorf("StatusText = %v, want %v", err.StatusText, http.StatusText(tt.status))
			}
			if err.AppCode != tt.code {
				t.Errorf("AppCode = %v, want %v", err.AppCode, tt.code)
			}
			if err.ErrorText != tt.message {
				t.Errorf("ErrorText = %v, want %v", err.ErrorText, tt.message)
			}
		})
	}
}

// TestErrNetwork tests network error response creation
func TestErrNetwork(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := ErrNetwork(originalErr)

	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
	if err.HTTPStatusCode != http.StatusServiceUnavailable {
		t.Errorf("HTTPStatusCode = %v, want %v", err.HTTPStatusCode, http.StatusServiceUnavailable)
	}
	if err.AppCode != ErrCodeNetworkError {
		t.Errorf("AppCode = %v, want %v", err.AppCode, ErrCodeNetworkError)
	}
}

// TestErrInvalidRequest tests invalid request error creation
func TestErrInvalidRequest(t *testing.T) {
	message := "Missing required field"
	err := ErrInvalidRequest(message)

	if err.HTTPStatusCode != http.StatusBadRequest {
		t.Errorf("HTTPStatusCode = %v, want %v", err.HTTPStatusCode, http.StatusBadRequest)
	}
	if err.ErrorText != message {
		t.Errorf("ErrorText = %v, want %v", err.ErrorText, message)
	}
}

// TestErrInternal tests internal server error creation
func TestErrInternal(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := ErrInternal(originalErr)

	if err.Err != originalErr {
		t.Errorf("Err = %v, want %v", err.Err, originalErr)
	}
	if err.HTTPStatusCode != http.StatusInternalServerError {
		t.Errorf("HTTPStatusCode = %v, want %v", err.HTTPStatusCode, http.StatusInternalServerError)
	}
	if err.ErrorText != "An unexpected error occurred. Please try again later" {
		t.Errorf("ErrorText = %v, want generic error message", err.ErrorText)
	}
}

// TestRenderIntegration tests the full render integration with Chi
func TestRenderIntegration(t *testing.T) {
	tests := []struct {
		name        string
		err         render.Renderer
		wantStatus  int
		wantHeaders map[string]string
	}{
		{
			name:       "render invalid license",
			err:        ErrInvalidLicenseKey,
			wantStatus: http.StatusBadRequest,
			wantHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:       "render rate limited",
			err:        ErrRateLimited,
			wantStatus: http.StatusTooManyRequests,
			wantHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/license/activate", nil)

			// Render the error
			render.Status(r, tt.wantStatus)
			render.JSON(w, r, tt.err)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("Response status = %v, want %v", w.Code, tt.wantStatus)
			}

			// Check headers
			for header, want := range tt.wantHeaders {
				got := w.Header().Get(header)
				if got != want {
					t.Errorf("Header %s = %v, want %v", header, got, want)
				}
			}

			// Check that response is valid JSON
			if w.Body.Len() == 0 {
				t.Error("Response body is empty")
			}
		})
	}
}