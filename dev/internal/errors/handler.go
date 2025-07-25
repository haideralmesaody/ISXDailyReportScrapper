package errors

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// Common error types following RFC 7807
const (
	TypeValidation      = "/errors/validation"
	TypeNotFound        = "/errors/not-found"
	TypeUnauthorized    = "/errors/unauthorized"
	TypeForbidden       = "/errors/forbidden"
	TypeRateLimit       = "/errors/rate-limit"
	TypeInternal        = "/errors/internal"
	TypeServiceDown     = "/errors/service-unavailable"
	TypeTimeout         = "/errors/timeout"
	TypeConflict        = "/errors/conflict"
	TypePayloadTooLarge = "/errors/payload-too-large"
)

// Domain-specific error types
const (
	TypeLicenseExpired      = "/errors/license/expired"
	TypeLicenseNotFound     = "/errors/license/not-found"
	TypeLicenseMismatch     = "/errors/license/machine-mismatch"
	TypePipelineNotFound    = "/errors/pipeline/not-found"
	TypePipelineRunning     = "/errors/pipeline/already-running"
	TypeDataNotFound        = "/errors/data/not-found"
	TypeDataCorrupted       = "/errors/data/corrupted"
	TypeWebSocketUpgrade    = "/errors/websocket/upgrade-failed"
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger      *slog.Logger
	includeStack bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *slog.Logger, includeStack bool) *ErrorHandler {
	return &ErrorHandler{
		logger:      logger.With(slog.String("component", "error_handler")),
		includeStack: includeStack,
	}
}

// HandleError converts any error to RFC 7807 format and responds
func (h *ErrorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Get request ID for tracing
	reqID := middleware.GetReqID(r.Context())

	// Log the error with full context
	h.logger.ErrorContext(r.Context(), "request failed",
		slog.String("error", err.Error()),
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
	)

	// Convert to problem details
	problem := h.ErrorToProblem(err, r)
	problem.WithExtension("trace_id", reqID)

	// Add stack trace in development
	if h.includeStack {
		problem.WithExtension("stack", getStackTrace())
	}

	// Render the error response
	render.Render(w, r, problem)
}

// ErrorToProblem converts an error to RFC 7807 Problem Details
func (h *ErrorHandler) ErrorToProblem(err error, r *http.Request) *ProblemDetails {
	// Check for context errors first
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return NewProblemDetails(
			http.StatusGatewayTimeout,
			TypeTimeout,
			"Request Timeout",
			"The request took too long to process and was cancelled",
			r.URL.Path,
		)
	}

	// Check for our custom API errors
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return h.apiErrorToProblem(apiErr, r)
	}

	// Check if it's an APIError with validation error code
	if apiErr != nil && apiErr.ErrorCode == "VALIDATION_ERROR" {
		// Extract validation errors from details
		if valErrors, ok := apiErr.Details.([]ValidationError); ok {
			return NewProblemDetails(
				http.StatusBadRequest,
				TypeValidation,
				"Validation Failed",
				apiErr.Message,
				r.URL.Path,
			).WithExtension("errors", valErrors)
		}
	}

	// Domain-specific error handling
	switch {
	case strings.Contains(err.Error(), "not found"):
		return NewProblemDetails(
			http.StatusNotFound,
			TypeNotFound,
			"Resource Not Found",
			err.Error(),
			r.URL.Path,
		)

	case strings.Contains(err.Error(), "license expired"):
		return NewProblemDetails(
			http.StatusForbidden,
			TypeLicenseExpired,
			"License Expired",
			"Your license has expired. Please renew to continue.",
			r.URL.Path,
		)

	case strings.Contains(err.Error(), "machine mismatch"):
		return NewProblemDetails(
			http.StatusForbidden,
			TypeLicenseMismatch,
			"License Machine Mismatch",
			"This license is registered to a different machine.",
			r.URL.Path,
		)

	case strings.Contains(err.Error(), "unauthorized"):
		return NewProblemDetails(
			http.StatusUnauthorized,
			TypeUnauthorized,
			"Unauthorized",
			"Authentication required to access this resource",
			r.URL.Path,
		)

	case strings.Contains(err.Error(), "forbidden"):
		return NewProblemDetails(
			http.StatusForbidden,
			TypeForbidden,
			"Forbidden",
			"You don't have permission to access this resource",
			r.URL.Path,
		)

	case strings.Contains(err.Error(), "rate limit"):
		return NewProblemDetails(
			http.StatusTooManyRequests,
			TypeRateLimit,
			"Rate Limit Exceeded",
			"Too many requests. Please try again later.",
			r.URL.Path,
		).WithExtension("retry_after", 60)

	case strings.Contains(err.Error(), "conflict"):
		return NewProblemDetails(
			http.StatusConflict,
			TypeConflict,
			"Conflict",
			err.Error(),
			r.URL.Path,
		)

	case strings.Contains(err.Error(), "payload too large"):
		return NewProblemDetails(
			http.StatusRequestEntityTooLarge,
			TypePayloadTooLarge,
			"Payload Too Large",
			"The request body exceeds the maximum allowed size",
			r.URL.Path,
		)

	default:
		// Generic internal error
		return NewProblemDetails(
			http.StatusInternalServerError,
			TypeInternal,
			"Internal Server Error",
			"An unexpected error occurred while processing your request",
			r.URL.Path,
		)
	}
}

// apiErrorToProblem converts APIError to ProblemDetails
func (h *ErrorHandler) apiErrorToProblem(apiErr *APIError, r *http.Request) *ProblemDetails {
	// Map error codes to problem types
	problemType := TypeInternal
	switch apiErr.ErrorCode {
	case "VALIDATION_FAILED":
		problemType = TypeValidation
	case "NOT_FOUND", "LICENSE_NOT_FOUND", "PIPELINE_NOT_FOUND":
		problemType = TypeNotFound
	case "UNAUTHORIZED", "INVALID_LICENSE":
		problemType = TypeUnauthorized
	case "FORBIDDEN":
		problemType = TypeForbidden
	case "CONFLICT":
		problemType = TypeConflict
	case "RATE_LIMIT_EXCEEDED":
		problemType = TypeRateLimit
	case "SERVICE_UNAVAILABLE":
		problemType = TypeServiceDown
	}

	problem := NewProblemDetails(
		apiErr.StatusCode,
		problemType,
		http.StatusText(apiErr.StatusCode),
		apiErr.Message,
		r.URL.Path,
	).WithExtension("error_code", apiErr.ErrorCode)

	// Add details if present
	if apiErr.Details != nil {
		problem.WithExtension("details", apiErr.Details)
	}

	return problem
}

// HandlePanic recovers from panics and returns RFC 7807 error
func (h *ErrorHandler) HandlePanic(w http.ResponseWriter, r *http.Request, recovered interface{}) {
	reqID := middleware.GetReqID(r.Context())

	// Log the panic
	h.logger.ErrorContext(r.Context(), "panic recovered",
		slog.Any("panic", recovered),
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("stack", string(debug.Stack())),
	)

	// Create problem details
	problem := NewProblemDetails(
		http.StatusInternalServerError,
		TypeInternal,
		"Internal Server Error",
		"An unexpected error occurred",
		r.URL.Path,
	).WithExtension("trace_id", reqID)

	// Add panic details in development
	if h.includeStack {
		problem.WithExtension("panic", fmt.Sprintf("%v", recovered))
		problem.WithExtension("stack", getStackTrace())
	}

	render.Render(w, r, problem)
}

// NotFound returns a standard 404 error
func (h *ErrorHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	problem := NewProblemDetails(
		http.StatusNotFound,
		TypeNotFound,
		"Not Found",
		"The requested resource was not found",
		r.URL.Path,
	).WithExtension("trace_id", middleware.GetReqID(r.Context()))

	render.Render(w, r, problem)
}

// MethodNotAllowed returns a standard 405 error
func (h *ErrorHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	problem := NewProblemDetails(
		http.StatusMethodNotAllowed,
		TypeInternal,
		"Method Not Allowed",
		fmt.Sprintf("Method %s is not allowed for this endpoint", r.Method),
		r.URL.Path,
	).WithExtension("trace_id", middleware.GetReqID(r.Context()))
	
	render.Render(w, r, problem)
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 1024*8)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Middleware returns an error handling middleware
func (h *ErrorHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer to capture errors
		ww := &errorResponseWriter{
			ResponseWriter: w,
			handler:        h,
			request:        r,
		}

		// Defer panic recovery
		defer func() {
			if err := recover(); err != nil {
				h.HandlePanic(ww, r, err)
			}
		}()

		next.ServeHTTP(ww, r)
	})
}

// errorResponseWriter wraps http.ResponseWriter to capture errors
type errorResponseWriter struct {
	http.ResponseWriter
	handler  *ErrorHandler
	request  *http.Request
	written  bool
	status   int
}

func (w *errorResponseWriter) WriteHeader(status int) {
	if !w.written {
		w.status = status
		w.written = true
		
		// Intercept error status codes
		if status >= 400 && status < 600 {
			// Log error responses
			w.handler.logger.WarnContext(w.request.Context(), "error response",
				slog.Int("status", status),
				slog.String("path", w.request.URL.Path),
				slog.String("method", w.request.Method),
			)
		}
		
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// JSON helper for consistent JSON error responses
func (h *ErrorHandler) JSON(w http.ResponseWriter, r *http.Request, status int, v interface{}) {
	render.Status(r, status)
	render.JSON(w, r, v)
}