package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/infrastructure"
)


// AuthMiddleware provides authentication middleware with logging
func AuthMiddleware(logger *slog.Logger, authService AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.WarnContext(ctx, "missing authorization header",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				
				problem := ProblemFromStatus(
					http.StatusUnauthorized,
					"Missing authorization header",
					infrastructure.GetTraceID(ctx),
				)
				problem.Render(w, r)
				return
			}
			
			// Parse Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				logger.WarnContext(ctx, "invalid authorization format",
					"method", r.Method,
					"path", r.URL.Path,
					"format", parts[0],
				)
				
				problem := ProblemFromStatus(
					http.StatusUnauthorized,
					"Invalid authorization format. Use: Bearer <token>",
					infrastructure.GetTraceID(ctx),
				)
				problem.Render(w, r)
				return
			}
			
			token := parts[1]
			
			// Validate token
			userInfo, err := authService.ValidateToken(ctx, token)
			if err != nil {
				logger.WarnContext(ctx, "authentication failed",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				
				problem := ProblemFromStatus(
					http.StatusUnauthorized,
					"Invalid or expired token",
					infrastructure.GetTraceID(ctx),
				)
				problem.Render(w, r)
				return
			}
			
			// Add user info to context
			ctx = context.WithValue(ctx, "user", userInfo)
			
			// Log successful authentication
			logger.DebugContext(ctx, "authentication successful",
				"user_id", userInfo.ID,
				"user_name", userInfo.Name,
				"method", r.Method,
				"path", r.URL.Path,
			)
			
			// Continue with authenticated request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthService interface for authentication
type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Roles []string `json:"roles"`
}

// APIKeyAuth provides API key authentication middleware
func APIKeyAuth(logger *slog.Logger, validKeys map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			// Check X-API-Key header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Try query parameter as fallback
				apiKey = r.URL.Query().Get("api_key")
			}
			
			if apiKey == "" {
				logger.WarnContext(ctx, "missing API key",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				
				problem := ProblemFromStatus(
					http.StatusUnauthorized,
					"API key required",
					infrastructure.GetTraceID(ctx),
				)
				problem.Render(w, r)
				return
			}
			
			// Validate API key
			clientName, valid := validKeys[apiKey]
			if !valid {
				logger.WarnContext(ctx, "invalid API key",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				
				problem := ProblemFromStatus(
					http.StatusUnauthorized,
					"Invalid API key",
					infrastructure.GetTraceID(ctx),
				)
				problem.Render(w, r)
				return
			}
			
			// Add client info to context
			ctx = context.WithValue(ctx, "api_client", clientName)
			
			logger.DebugContext(ctx, "API key authentication successful",
				"client", clientName,
				"method", r.Method,
				"path", r.URL.Path,
			)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}


// SecureHeaders provides configurable security headers
type SecureHeaders struct {
	// HSTS settings
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool
	
	// CSP settings
	ContentSecurityPolicy string
	
	// Frame options
	XFrameOptions string
	
	// Other security headers
	XContentTypeOptions string
	XSSProtection       string
	ReferrerPolicy      string
	PermissionsPolicy   string
	
	// Development mode (relaxes some policies)
	DevMode bool
}

// DefaultSecureHeaders returns secure headers with default settings
func DefaultSecureHeaders() *SecureHeaders {
	return &SecureHeaders{
		HSTSMaxAge:            63072000, // 2 years
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
		XSSProtection:         "1; mode=block",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// Handler returns the middleware handler
func (sh *SecureHeaders) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip security headers for WebSocket upgrades
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}
		
		// HSTS
		if sh.HSTSMaxAge > 0 && (r.TLS != nil || sh.DevMode) {
			hsts := fmt.Sprintf("max-age=%d", sh.HSTSMaxAge)
			if sh.HSTSIncludeSubdomains {
				hsts += "; includeSubDomains"
			}
			if sh.HSTSPreload {
				hsts += "; preload"
			}
			w.Header().Set("Strict-Transport-Security", hsts)
		}
		
		// CSP
		if sh.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", sh.ContentSecurityPolicy)
		} else if !sh.DevMode {
			// Default CSP for production
			w.Header().Set("Content-Security-Policy", sh.defaultCSP())
		}
		
		// X-Frame-Options
		if sh.XFrameOptions != "" {
			w.Header().Set("X-Frame-Options", sh.XFrameOptions)
		}
		
		// X-Content-Type-Options
		if sh.XContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", sh.XContentTypeOptions)
		}
		
		// X-XSS-Protection
		if sh.XSSProtection != "" {
			w.Header().Set("X-XSS-Protection", sh.XSSProtection)
		}
		
		// Referrer-Policy
		if sh.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", sh.ReferrerPolicy)
		}
		
		// Permissions-Policy
		if sh.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", sh.PermissionsPolicy)
		} else if !sh.DevMode {
			w.Header().Set("Permissions-Policy", sh.defaultPermissionsPolicy())
		}
		
		next.ServeHTTP(w, r)
	})
}

// defaultCSP returns the default Content Security Policy
func (sh *SecureHeaders) defaultCSP() string {
	policies := []string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://code.highcharts.com",
		"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://code.highcharts.com",
		"img-src 'self' data: https: blob:",
		"font-src 'self' https://cdnjs.cloudflare.com",
		"connect-src 'self' ws: wss:",
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"upgrade-insecure-requests",
	}
	
	if sh.DevMode {
		// Relax policies for development
		policies = []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' *",
			"style-src 'self' 'unsafe-inline' *",
			"img-src * data: blob:",
			"font-src *",
			"connect-src *",
		}
	}
	
	return strings.Join(policies, "; ")
}

// defaultPermissionsPolicy returns the default Permissions Policy
func (sh *SecureHeaders) defaultPermissionsPolicy() string {
	policies := []string{
		"accelerometer=()",
		"camera=()",
		"geolocation=()",
		"gyroscope=()",
		"magnetometer=()",
		"microphone=()",
		"payment=()",
		"usb=()",
		"interest-cohort=()", // FLoC opt-out
	}
	return strings.Join(policies, ", ")
}

// AuditLog provides audit logging middleware for sensitive operations
func AuditLog(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			start := time.Now()
			
			// Capture response for audit
			ww := &auditResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			
			// Get user info from context
			var userID, userName string
			if user, ok := ctx.Value("user").(*UserInfo); ok {
				userID = user.ID
				userName = user.Name
			} else if client, ok := ctx.Value("api_client").(string); ok {
				userID = "api"
				userName = client
			}
			
			// Log audit entry
			logger.InfoContext(ctx, "audit log",
				"event_type", "api_access",
				"user_id", userID,
				"user_name", userName,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.Query().Encode(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
			
			// Process request
			next.ServeHTTP(ww, r)
			
			// Log completion with response details
			logger.InfoContext(ctx, "audit log complete",
				"event_type", "api_response",
				"user_id", userID,
				"user_name", userName,
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.statusCode,
				"duration", time.Since(start).String(),
			)
		})
	}
}

// auditResponseWriter captures the response status code
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *auditResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}