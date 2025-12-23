package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/isxcli/isxcli/internal/infrastructure"
)

// Problem represents an RFC 7807 problem details object
type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	Trace  string `json:"trace_id,omitempty"`
}

// Render implements the chi render.Renderer interface
func (p Problem) Render(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	return json.NewEncoder(w).Encode(p)
}

// Common error types
var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrBadRequest        = errors.New("bad request")
	ErrInternalServer    = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrRequestTimeout    = errors.New("request timeout")
)

// ErrorHandler creates an error handling middleware with RFC 7807 responses
func ErrorHandler(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom response writer to capture errors
			ew := &errorWriter{
				ResponseWriter: w,
				logger:         logger,
				request:        r,
			}
			
			next.ServeHTTP(ew, r)
			
			// If an error was set, handle it
			if ew.err != nil {
				handleError(w, r, ew.err, logger)
			}
		})
	}
}

// errorWriter wraps http.ResponseWriter to capture errors
type errorWriter struct {
	http.ResponseWriter
	err     error
	logger  *slog.Logger
	request *http.Request
	written bool
}

// WriteHeader captures status codes that indicate errors
func (ew *errorWriter) WriteHeader(code int) {
	if !ew.written && code >= 400 {
		ew.written = true
	}
	ew.ResponseWriter.WriteHeader(code)
}

// Write prevents writing if an error occurred
func (ew *errorWriter) Write(b []byte) (int, error) {
	if ew.err != nil {
		return 0, nil // Prevent writing if error is set
	}
	ew.written = true
	return ew.ResponseWriter.Write(b)
}

// Error sets an error to be handled
func (ew *errorWriter) Error(err error) {
	ew.err = err
}

// handleError converts errors to RFC 7807 problem responses
func handleError(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	ctx := r.Context()
	traceID := infrastructure.GetTraceID(ctx)
	
	// Log the error with context
	logger.ErrorContext(ctx, "request error",
		"error", err,
		"method", r.Method,
		"path", r.URL.Path,
		"trace_id", traceID,
	)
	
	// Map error to problem details
	problem := mapErrorToProblem(err, traceID)
	
	// Write RFC 7807 response
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	json.NewEncoder(w).Encode(problem)
}

// mapErrorToProblem maps errors to RFC 7807 problem details
func mapErrorToProblem(err error, traceID string) Problem {
	// Check for specific error types
	switch {
	case errors.Is(err, ErrNotFound):
		return Problem{
			Type:   "/errors/not-found",
			Title:  "Resource Not Found",
			Status: http.StatusNotFound,
			Detail: err.Error(),
			Trace:  traceID,
		}
	case errors.Is(err, ErrUnauthorized):
		return Problem{
			Type:   "/errors/unauthorized",
			Title:  "Unauthorized",
			Status: http.StatusUnauthorized,
			Detail: "Authentication required",
			Trace:  traceID,
		}
	case errors.Is(err, ErrForbidden):
		return Problem{
			Type:   "/errors/forbidden",
			Title:  "Forbidden",
			Status: http.StatusForbidden,
			Detail: "Access denied",
			Trace:  traceID,
		}
	case errors.Is(err, ErrBadRequest):
		return Problem{
			Type:   "/errors/bad-request",
			Title:  "Bad Request",
			Status: http.StatusBadRequest,
			Detail: err.Error(),
			Trace:  traceID,
		}
	case errors.Is(err, ErrServiceUnavailable):
		return Problem{
			Type:   "/errors/service-unavailable",
			Title:  "Service Unavailable",
			Status: http.StatusServiceUnavailable,
			Detail: "The service is temporarily unavailable",
			Trace:  traceID,
		}
	case errors.Is(err, ErrRateLimitExceeded):
		return Problem{
			Type:   "/errors/rate-limit-exceeded",
			Title:  "Too Many Requests",
			Status: http.StatusTooManyRequests,
			Detail: "Rate limit exceeded. Please retry later",
			Trace:  traceID,
		}
	case errors.Is(err, ErrRequestTimeout):
		return Problem{
			Type:   "/errors/request-timeout",
			Title:  "Request Timeout",
			Status: http.StatusGatewayTimeout,
			Detail: "The request took too long to process",
			Trace:  traceID,
		}
	}
	
	// Check for validation errors
	if strings.Contains(strings.ToLower(err.Error()), "validation") {
		return Problem{
			Type:   "/errors/validation-failed",
			Title:  "Validation Failed",
			Status: http.StatusBadRequest,
			Detail: err.Error(),
			Trace:  traceID,
		}
	}
	
	// Default to internal server error
	return Problem{
		Type:   "/errors/internal-server-error",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: "An unexpected error occurred",
		Trace:  traceID,
	}
}

// NewErrorResponder creates a function that writes RFC 7807 error responses
func NewErrorResponder(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		handleError(w, r, err, logger)
	}
}

// ProblemFromStatus creates a Problem from an HTTP status code
func ProblemFromStatus(status int, detail string, traceID string) Problem {
	var title, problemType string
	
	switch status {
	case http.StatusBadRequest:
		title = "Bad Request"
		problemType = "/errors/bad-request"
	case http.StatusUnauthorized:
		title = "Unauthorized"
		problemType = "/errors/unauthorized"
	case http.StatusForbidden:
		title = "Forbidden"
		problemType = "/errors/forbidden"
	case http.StatusNotFound:
		title = "Not Found"
		problemType = "/errors/not-found"
	case http.StatusMethodNotAllowed:
		title = "Method Not Allowed"
		problemType = "/errors/method-not-allowed"
	case http.StatusConflict:
		title = "Conflict"
		problemType = "/errors/conflict"
	case http.StatusTooManyRequests:
		title = "Too Many Requests"
		problemType = "/errors/rate-limit-exceeded"
	case http.StatusInternalServerError:
		title = "Internal Server Error"
		problemType = "/errors/internal-server-error"
	case http.StatusServiceUnavailable:
		title = "Service Unavailable"
		problemType = "/errors/service-unavailable"
	case http.StatusGatewayTimeout:
		title = "Gateway Timeout"
		problemType = "/errors/gateway-timeout"
	default:
		title = http.StatusText(status)
		problemType = "/errors/unknown"
	}
	
	return Problem{
		Type:   problemType,
		Title:  title,
		Status: status,
		Detail: detail,
		Trace:  traceID,
	}
}