package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

// License-specific errors (using errors package for sentinel errors)
var (
	ErrLicenseExpired      = errors.New("license expired")
	ErrMachineMismatch     = errors.New("machine mismatch")
	ErrLicenseNotActivated = errors.New("license not activated")
	ErrInvalidLicenseKey   = errors.New("invalid license key")
	ErrRateLimited         = errors.New("rate limited")
	ErrNetworkError        = errors.New("network error")
)

// ProblemDetails implements RFC 7807 Problem Details for HTTP APIs
type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	
	// Additional fields for extensibility
	Extensions map[string]interface{} `json:"-"`
}

// Render implements the render.Renderer interface
func (pd *ProblemDetails) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, pd.Status)
	return nil
}

// MarshalJSON custom marshaler to include extensions
func (pd *ProblemDetails) MarshalJSON() ([]byte, error) {
	type Alias ProblemDetails
	data := make(map[string]interface{})
	
	// Add standard fields
	data["type"] = pd.Type
	data["title"] = pd.Title
	data["status"] = pd.Status
	
	if pd.Detail != "" {
		data["detail"] = pd.Detail
	}
	if pd.Instance != "" {
		data["instance"] = pd.Instance
	}
	
	// Add extensions
	for k, v := range pd.Extensions {
		data[k] = v
	}
	
	// Use standard JSON marshaling
	return json.Marshal(data)
}

// NewProblemDetails creates a new RFC 7807 compliant error
func NewProblemDetails(status int, problemType, title, detail, instance string) *ProblemDetails {
	return &ProblemDetails{
		Type:     problemType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Extensions: make(map[string]interface{}),
	}
}

// WithExtension adds an extension field to the problem details
func (pd *ProblemDetails) WithExtension(key string, value interface{}) *ProblemDetails {
	pd.Extensions[key] = value
	return pd
}

// MapLicenseError maps domain errors to HTTP problem details
func MapLicenseError(err error, traceID string) render.Renderer {
	instance := fmt.Sprintf("/api/license#trace-%s", traceID)
	
	// Check if it's an APIError from errors.go
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode == "LICENSE_NOT_FOUND" {
			return NewProblemDetails(
				http.StatusNotFound,
				"/errors/license-not-found",
				"License Not Found",
				"No license file found in the system. Please activate a license.",
				instance,
			).WithExtension("trace_id", traceID).
				WithExtension("error_code", "LICENSE_NOT_FOUND")
		}
	}
	
	switch {
	case errors.Is(err, ErrLicenseExpired):
		return NewProblemDetails(
			http.StatusForbidden,
			"/errors/license-expired",
			"License Expired",
			"Your license has expired. Please renew to continue.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "LICENSE_EXPIRED")
			
	case errors.Is(err, ErrMachineMismatch):
		return NewProblemDetails(
			http.StatusForbidden,
			"/errors/machine-mismatch",
			"Machine Mismatch",
			"This license is registered to a different machine.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "MACHINE_MISMATCH")
			
	case errors.Is(err, ErrLicenseNotActivated):
		return NewProblemDetails(
			http.StatusPreconditionRequired,
			"/errors/license-not-activated",
			"License Not Activated",
			"No license has been activated. Please activate a license to continue.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "LICENSE_NOT_ACTIVATED")
			
	case errors.Is(err, ErrInvalidLicenseKey):
		return NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-license-key",
			"Invalid License Key",
			"The provided license key is invalid or malformed.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "INVALID_LICENSE_KEY")
			
	case errors.Is(err, ErrRateLimited):
		return NewProblemDetails(
			http.StatusTooManyRequests,
			"/errors/rate-limited",
			"Too Many Requests",
			"Too many activation attempts. Please try again later.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "RATE_LIMITED").
			WithExtension("retry_after", 900) // 15 minutes
			
	case errors.Is(err, ErrNetworkError):
		return NewProblemDetails(
			http.StatusServiceUnavailable,
			"/errors/network-error",
			"Network Error",
			"Unable to connect to license server. Please check your connection.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "NETWORK_ERROR")
			
	default:
		// Generic error
		return NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/internal-error",
			"Internal Server Error",
			"An unexpected error occurred while processing your request.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "INTERNAL_ERROR")
	}
}