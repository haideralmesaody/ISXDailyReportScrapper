package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"isxcli/internal/services"
	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Handler implements the http.Handler interface for rate limiting
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded","retry_after":"60s"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CorsOptions holds CORS configuration
type CorsOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// Cors returns CORS middleware
func Cors(options CorsOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			if len(options.AllowedOrigins) == 0 {
				allowed = true
			} else {
				for _, allowedOrigin := range options.AllowedOrigins {
					if allowedOrigin == "*" || strings.EqualFold(allowedOrigin, origin) {
						allowed = true
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(options.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(options.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(options.ExposedHeaders, ", "))
			
			if options.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(options.MaxAge))

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


// StructuredLogger provides structured logging middleware
type StructuredLogger struct {
	logger services.Logger
	next   http.Handler
}

// NewStructuredLogger creates a new structured logger middleware
func NewStructuredLogger(logger services.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &StructuredLogger{logger: logger, next: next}
	}
}

// ServeHTTP implements the http.Handler interface
func (sl *StructuredLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Create a response writer wrapper to capture status code
	wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	
	sl.next.ServeHTTP(wrapped, r)
	
	duration := time.Since(start)
	
	sl.logger.Info("HTTP request",
		"method", r.Method,
		"path", r.URL.Path,
		"status", wrapped.statusCode,
		"duration", duration.String(),
		"user_agent", r.UserAgent(),
		"remote_addr", r.RemoteAddr,
		"request_id", r.Header.Get("X-Request-ID"),
	)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Recoverer recovers from panics and logs them
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"Internal server error","request_id":"%s"}`, r.Header.Get("X-Request-ID"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Timeout adds request timeout middleware
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS returns a CORS middleware
func CORS(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
			
			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID adds request ID middleware
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
			w.Header().Set("X-Request-ID", requestID)
		}
		r.Header.Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

// RealIP extracts the real client IP address
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}
		
		// Clean up IP address
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = strings.TrimSpace(ip[:idx])
		}
		
		r.Header.Set("X-Real-IP", ip)
		next.ServeHTTP(w, r)
	})
}

// Compress provides response compression middleware
func Compress(level int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encoding := r.Header.Get("Accept-Encoding")
			if strings.Contains(encoding, "gzip") {
				w.Header().Set("Content-Encoding", "gzip")
			}
			next.ServeHTTP(w, r)
		})
	}
}