package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"go.opentelemetry.io/otel/trace"

	"github.com/isxcli/isxcli/internal/infrastructure"
)

// RequestIDKey is the context key for request ID
const RequestIDKey = "request-id"

// RequestID middleware generates a unique request ID for each request.
// Uses UUID v4 for truly unique IDs across distributed systems.
// This should be the FIRST middleware in the chain.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID from header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate new UUID
			requestID = uuid.New().String()
		}

		// Set request ID in response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context - this becomes the trace_id
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = infrastructure.WithTraceID(ctx, requestID)
		
		// If there's an active span, use its trace ID instead
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			ctx = infrastructure.WithTraceID(ctx, traceID)
		}

		// Continue with request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetReqID retrieves the request ID from the context
func GetReqID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// StructuredLogger provides Chi-compatible structured logging middleware using slog.
// This should come AFTER RequestID and RealIP middlewares.
func StructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Get request ID from context
			ctx := r.Context()
			requestID, _ := ctx.Value(RequestIDKey).(string)
			
			// Get OpenTelemetry trace ID if available
			traceID := infrastructure.GetTraceID(ctx)
			if traceID == "" && requestID != "" {
				traceID = requestID
			}
			
			// Create logger with request context
			reqLogger := logger
			if traceID != "" {
				reqLogger = logger.With("trace_id", traceID)
			}

			// Wrap response writer to capture status and size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Log request start
			reqLogger.InfoContext(ctx, "request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Process request
			next.ServeHTTP(ww, r)

			// Log request completion
			reqLogger.InfoContext(ctx, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
			)
		})
	}
}

// Recoverer recovers from panics and logs them with slog.
// Uses infrastructure package for proper trace_id handling.
func Recoverer(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					ctx := r.Context()
					
					// Log the panic with stack trace
					logger.ErrorContext(ctx, "panic recovered",
						"panic", rvr,
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
					)

					// Return error response
					w.Header().Set("Content-Type", "application/problem+json")
					w.WriteHeader(http.StatusInternalServerError)
					
					// Get trace ID for error response
					traceID := infrastructure.GetTraceID(ctx)
					if traceID == "" {
						if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
							traceID = reqID
						}
					}

					// RFC 7807 error response
					response := `{"type":"/errors/internal-server-error","title":"Internal Server Error","status":500,"detail":"An unexpected error occurred","trace_id":"` + traceID + `"}`
					w.Write([]byte(response))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter provides rate limiting functionality with logging
type RateLimiter struct {
	limiter *rate.Limiter
	logger  *slog.Logger
}

// NewRateLimiter creates a new rate limiter with logging
func NewRateLimiter(rps float64, burst int, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
		logger:  logger,
	}
}

// Handler implements rate limiting middleware
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		if !rl.limiter.Allow() {
			// Log rate limit violation
			rl.logger.WarnContext(ctx, "rate limit exceeded",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			// Return RFC 7807 error
			w.Header().Set("Content-Type", "application/problem+json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			
			traceID := infrastructure.GetTraceID(ctx)
			response := `{"type":"/errors/rate-limit-exceeded","title":"Too Many Requests","status":429,"detail":"Rate limit exceeded. Please retry after 60 seconds","trace_id":"` + traceID + `"}`
			w.Write([]byte(response))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Timeout middleware with context and logging
func Timeout(timeout time.Duration, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create channel to track if handler completes
			done := make(chan struct{})
			
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Timeout occurred
				logger.ErrorContext(r.Context(), "request timeout",
					"method", r.Method,
					"path", r.URL.Path,
					"timeout", timeout.String(),
				)
				
				// Return timeout error
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusGatewayTimeout)
				
				traceID := infrastructure.GetTraceID(r.Context())
				response := `{"type":"/errors/request-timeout","title":"Request Timeout","status":504,"detail":"The request took too long to process","trace_id":"` + traceID + `"}`
				w.Write([]byte(response))
			}
		})
	}
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
	Logger           *slog.Logger
}

// CORS middleware with logging
func CORS(config CORSConfig) func(next http.Handler) http.Handler {
	// Default values
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 300
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			if len(config.AllowedOrigins) == 0 {
				allowed = true
			} else {
				for _, allowedOrigin := range config.AllowedOrigins {
					if allowedOrigin == "*" || strings.EqualFold(allowedOrigin, origin) {
						allowed = true
						break
					}
				}
			}

			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				if config.Logger != nil {
					config.Logger.DebugContext(r.Context(), "CORS preflight request",
						"origin", origin,
						"allowed", allowed,
					)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


// SecurityHeaders adds security-related headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers per OWASP
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: blob:; font-src 'self' data:")
		
		// HSTS for HTTPS connections
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	// Fallback to trace ID
	return infrastructure.GetTraceID(ctx)
}

// Compress provides response compression middleware using Chi's implementation
func Compress(level int) func(next http.Handler) http.Handler {
	return middleware.Compress(level)
}

// RealIP extracts the real client IP using Chi's implementation
func RealIP(next http.Handler) http.Handler {
	return middleware.RealIP(next)
}

// StripSlashes removes trailing slashes from requests
func StripSlashes(next http.Handler) http.Handler {
	return middleware.StripSlashes(next)
}

// CorsOptions holds CORS configuration (deprecated - use CORSConfig)
// Kept for backward compatibility
type CorsOptions = CORSConfig

// Cors returns CORS middleware (deprecated - use CORS)
// Kept for backward compatibility
func Cors(options CorsOptions) func(http.Handler) http.Handler {
	return CORS(options)
}