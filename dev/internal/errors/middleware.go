package errors

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// ErrorMiddleware provides centralized error handling and logging
type ErrorMiddleware struct {
	handler *ErrorHandler
	logger  *slog.Logger
}

// NewErrorMiddleware creates a new error handling middleware
func NewErrorMiddleware(handler *ErrorHandler, logger *slog.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		handler: handler,
		logger:  logger.With(slog.String("component", "error_middleware")),
	}
}

// Handler returns the middleware handler function
func (m *ErrorMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create wrapped response writer to capture status
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// Capture request body for error logging (if needed)
		var requestBody []byte
		if r.Body != nil && r.ContentLength > 0 && r.ContentLength < 1024*1024 { // Max 1MB
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(requestBody))
		}

		// Track request timing
		start := time.Now()

		// Defer panic recovery
		defer func() {
			if err := recover(); err != nil {
				m.handler.HandlePanic(ww, r, err)
			}
		}()

		// Serve the request
		next.ServeHTTP(ww, r)

		// Log request details
		duration := time.Since(start)
		status := ww.Status()

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if status >= 400 && status < 500 {
			logLevel = slog.LevelWarn
		} else if status >= 500 {
			logLevel = slog.LevelError
		}

		// Build log attributes
		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Duration("duration", duration),
			slog.Int("bytes", ww.BytesWritten()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		}

		// Add query parameters if present
		if r.URL.RawQuery != "" {
			attrs = append(attrs, slog.String("query", r.URL.RawQuery))
		}

		// Add request body for errors (sanitized)
		if status >= 400 && len(requestBody) > 0 {
			bodyStr := string(requestBody)
			// Sanitize sensitive data
			bodyStr = sanitizeRequestBody(bodyStr)
			if len(bodyStr) > 500 {
				bodyStr = bodyStr[:500] + "..."
			}
			attrs = append(attrs, slog.String("request_body", bodyStr))
		}

		// Log the request
		m.logger.LogAttrs(r.Context(), logLevel, "http request", attrs...)
	})
}

// sanitizeRequestBody removes sensitive data from request body for logging
func sanitizeRequestBody(body string) string {
	// Parse as JSON if possible
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err == nil {
		// Remove sensitive fields
		sensitiveFields := []string{
			"password", "token", "secret", "api_key", "apiKey",
			"license_key", "licenseKey", "credit_card", "ssn",
		}
		
		for _, field := range sensitiveFields {
			if _, exists := data[field]; exists {
				data[field] = "[REDACTED]"
			}
		}
		
		// Convert back to JSON
		sanitized, _ := json.Marshal(data)
		return string(sanitized)
	}
	
	// If not JSON, return as-is (could implement other sanitization)
	return body
}

// RecoveryMiddleware provides panic recovery with proper error responses
func RecoveryMiddleware(handler *ErrorHandler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					handler.HandlePanic(w, r, err)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}