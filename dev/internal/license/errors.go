package license

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse implements the render.Renderer interface for API errors
type ErrResponse struct {
	Err            error  `json:"-"`
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	AppCode        string `json:"code,omitempty"`
	ErrorText      string `json:"error,omitempty"`
}

// Render implements the render.Renderer interface
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// Error codes for license operations
const (
	ErrCodeInvalidKey       = "INVALID_LICENSE_KEY"
	ErrCodeExpired          = "LICENSE_EXPIRED"
	ErrCodeMachineMismatch  = "MACHINE_MISMATCH"
	ErrCodeAlreadyActivated = "ALREADY_ACTIVATED"
	ErrCodeNotFound         = "LICENSE_NOT_FOUND"
	ErrCodeNetworkError     = "NETWORK_ERROR"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeNotActivated     = "NOT_ACTIVATED"
)

// Common error responses
var (
	ErrInvalidLicenseKey = &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid license key",
		AppCode:        ErrCodeInvalidKey,
		ErrorText:      "The provided license key is invalid or malformed",
	}

	ErrLicenseExpired = &ErrResponse{
		HTTPStatusCode: http.StatusForbidden,
		StatusText:     "License expired",
		AppCode:        ErrCodeExpired,
		ErrorText:      "Your license has expired. Please renew to continue",
	}

	ErrMachineMismatch = &ErrResponse{
		HTTPStatusCode: http.StatusForbidden,
		StatusText:     "Machine mismatch",
		AppCode:        ErrCodeMachineMismatch,
		ErrorText:      "This license is registered to a different machine",
	}

	ErrLicenseNotFound = &ErrResponse{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "License not found",
		AppCode:        ErrCodeNotFound,
		ErrorText:      "The specified license key was not found in our system",
	}

	ErrNotActivated = &ErrResponse{
		HTTPStatusCode: http.StatusPreconditionRequired,
		StatusText:     "License not activated",
		AppCode:        ErrCodeNotActivated,
		ErrorText:      "No license has been activated. Please activate a license to continue",
	}

	ErrRateLimited = &ErrResponse{
		HTTPStatusCode: http.StatusTooManyRequests,
		StatusText:     "Too many requests",
		AppCode:        ErrCodeRateLimited,
		ErrorText:      "Too many activation attempts. Please try again later",
	}
)

// NewErrResponse creates a custom error response
func NewErrResponse(status int, code, message string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: status,
		StatusText:     http.StatusText(status),
		AppCode:        code,
		ErrorText:      message,
	}
}

// ErrNetwork creates a network error response
func ErrNetwork(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusServiceUnavailable,
		StatusText:     "Network error",
		AppCode:        ErrCodeNetworkError,
		ErrorText:      "Unable to connect to license server. Please check your internet connection",
	}
}

// ErrInvalidRequest creates a bad request error
func ErrInvalidRequest(message string) *ErrResponse {
	return &ErrResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Invalid request",
		ErrorText:      message,
	}
}

// ErrInternal creates an internal server error
func ErrInternal(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal server error",
		ErrorText:      "An unexpected error occurred. Please try again later",
	}
}