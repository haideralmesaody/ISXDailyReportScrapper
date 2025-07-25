package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

// APIError represents a structured API error response
type APIError struct {
	StatusCode int         `json:"status_code"`
	ErrorCode  string      `json:"error_code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// Render implements the render.Renderer interface for chi/render
func (e *APIError) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// New creates a new APIError with the given parameters
func New(statusCode int, errorCode, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

// NewWithDetails creates a new APIError with additional details
func NewWithDetails(statusCode int, errorCode, message string, details interface{}) *APIError {
	return &APIError{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Message:    message,
		Details:    details,
	}
}

// Predefined error types for common scenarios
var (
	// 400 Bad Request
	ErrInvalidRequest     = New(http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
	ErrValidationFailed   = New(http.StatusBadRequest, "VALIDATION_FAILED", "Request validation failed")
	ErrMissingParameter   = New(http.StatusBadRequest, "MISSING_PARAMETER", "Required parameter is missing")
	ErrInvalidParameter   = New(http.StatusBadRequest, "INVALID_PARAMETER", "Invalid parameter value")

	// 401 Unauthorized
	ErrUnauthorized     = New(http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
	ErrInvalidLicense   = New(http.StatusUnauthorized, "INVALID_LICENSE", "Invalid or expired license")

	// 403 Forbidden
	ErrForbidden = New(http.StatusForbidden, "FORBIDDEN", "Access denied")

	// 404 Not Found
	ErrNotFound         = New(http.StatusNotFound, "NOT_FOUND", "Resource not found")
	ErrLicenseNotFound  = New(http.StatusNotFound, "LICENSE_NOT_FOUND", "License file not found")
	ErrPipelineNotFound = New(http.StatusNotFound, "PIPELINE_NOT_FOUND", "Pipeline not found")

	// 409 Conflict
	ErrConflict = New(http.StatusConflict, "CONFLICT", "Resource conflict")

	// 422 Unprocessable Entity
	ErrUnprocessableEntity = New(http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", "Request could not be processed")

	// 429 Too Many Requests
	ErrRateLimitExceeded = New(http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")

	// 500 Internal Server Error
	ErrInternalServer   = New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error")
	ErrPipelineFailed   = New(http.StatusInternalServerError, "PIPELINE_FAILED", "Pipeline execution failed")
	ErrFileSystem       = New(http.StatusInternalServerError, "FILESYSTEM_ERROR", "File system error")
	ErrWebSocketUpgrade = New(http.StatusInternalServerError, "WEBSOCKET_UPGRADE_FAILED", "WebSocket upgrade failed")

	// 503 Service Unavailable
	ErrServiceUnavailable = New(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service temporarily unavailable")
)

// Helper functions for specific error types

// InvalidRequestWithError creates an invalid request error with details
func InvalidRequestWithError(err error) *APIError {
	return NewWithDetails(http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err.Error())
}

// ErrValidation creates a validation error with field details
func ErrValidation(field, message string) *APIError {
	return NewWithDetails(http.StatusBadRequest, "VALIDATION_FAILED", "Request validation failed", ValidationError{
		Field:   field,
		Message: message,
	})
}

// NotFoundError creates a not found error with details
func NotFoundError(resource string) *APIError {
	return NewWithDetails(http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("%s not found", resource), resource)
}

// LicenseNotFoundError creates a license not found error
func LicenseNotFoundError(err error) *APIError {
	return NewWithDetails(http.StatusNotFound, "LICENSE_NOT_FOUND", "License file not found", err.Error())
}

// ErrLicenseActivation creates a license activation error
func ErrLicenseActivation(err error) *APIError {
	return NewWithDetails(http.StatusBadRequest, "LICENSE_ACTIVATION_FAILED", "Failed to activate license", err.Error())
}

// ErrPipelineExecution creates a pipeline execution error
func ErrPipelineExecution(err error) *APIError {
	return NewWithDetails(http.StatusInternalServerError, "PIPELINE_EXECUTION_FAILED", "Pipeline execution failed", err.Error())
}

// FileSystemError creates a filesystem error
func FileSystemError(operation string, err error) *APIError {
	return NewWithDetails(http.StatusInternalServerError, "FILESYSTEM_ERROR", fmt.Sprintf("File system error during %s", operation), err.Error())
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   *APIError `json:"error"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *APIError) *ErrorResponse {
	return &ErrorResponse{
		Success: false,
		Error:   err,
	}
}

// Render implements the render.Renderer interface
func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return e.Error.Render(w, r)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// NewValidationErrors creates validation errors from multiple fields
func NewValidationErrors(errors []ValidationError) *APIError {
	return NewWithDetails(
		http.StatusBadRequest,
		"VALIDATION_FAILED",
		"Request validation failed",
		ValidationErrors{Errors: errors},
	)
}

// PanicRecovery represents panic recovery information
type PanicRecovery struct {
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

// ErrPanic creates a panic recovery error
func ErrPanic(rec interface{}) *APIError {
	return NewWithDetails(
		http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		"Internal server error",
		PanicRecovery{
			Message: fmt.Sprintf("%v", rec),
		},
	)
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(NewErrorResponse(err))
}

// NewValidationError creates a simple validation error
func NewValidationError(message string) *APIError {
	return New(http.StatusBadRequest, "VALIDATION_FAILED", message)
}

// NewInternalError creates a simple internal server error
func NewInternalError(message string) *APIError {
	return New(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}