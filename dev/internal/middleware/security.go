package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict Transport Security
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		
		// Content Security Policy
		csp := []string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://code.highcharts.com",
			"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com https://code.highcharts.com",
			"img-src 'self' data: https: blob:",
			"font-src 'self' https://cdnjs.cloudflare.com",
			"connect-src 'self' ws: wss:",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))
		
		// X-Frame-Options (legacy support)
		w.Header().Set("X-Frame-Options", "DENY")
		
		// X-Content-Type-Options
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// X-XSS-Protection (legacy support)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Referrer-Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions-Policy
		permissions := []string{
			"accelerometer=()",
			"camera=()",
			"geolocation=()",
			"gyroscope=()",
			"magnetometer=()",
			"microphone=()",
			"payment=()",
			"usb=()",
		}
		w.Header().Set("Permissions-Policy", strings.Join(permissions, ", "))
		
		next.ServeHTTP(w, r)
	})
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