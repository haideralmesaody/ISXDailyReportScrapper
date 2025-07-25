package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"isxcli/internal/license"
)

// LicenseValidator provides license validation middleware
type LicenseValidator struct {
	manager       LicenseManagerInterface
	logger        *slog.Logger
	cache         *validationCache
	excludePaths  []string
	excludePrefixes []string
}

// validationCache stores recent validation results
type validationCache struct {
	mu      sync.RWMutex
	valid   bool
	checkedAt time.Time
	ttl     time.Duration
}

// NewLicenseValidator creates a new license validation middleware
func NewLicenseValidator(manager LicenseManagerInterface, logger *slog.Logger) *LicenseValidator {
	return &LicenseValidator{
		manager: manager,
		logger:  logger.With(slog.String("component", "license_middleware")),
		cache: &validationCache{
			ttl: 5 * time.Minute, // Cache validation results for 5 minutes
		},
		excludePaths: []string{
			"/",
			"/license",
			"/api/license/activate",
			"/api/license/status",
			"/api/health",
			"/api/health/ready",
			"/api/health/live",
			"/static/",
			"/ws",
		},
		excludePrefixes: []string{
			"/static/",
			"/templates/",
		},
	}
}

// Handler returns the middleware handler function
func (lv *LicenseValidator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID for logging
		reqID := middleware.GetReqID(r.Context())
		
		// Check if path should be excluded
		if lv.shouldExcludePath(r.URL.Path) {
			lv.logger.Debug("skipping license validation for excluded path",
				slog.String("path", r.URL.Path),
				slog.String("request_id", reqID))
			next.ServeHTTP(w, r)
			return
		}

		// Check cached validation result
		if lv.isCacheValid() {
			lv.logger.Debug("using cached license validation",
				slog.String("request_id", reqID))
			next.ServeHTTP(w, r)
			return
		}

		// Perform license validation
		valid, err := lv.validateLicense(r.Context())
		
		if err != nil {
			lv.logger.Error("license validation error",
				slog.String("error", err.Error()),
				slog.String("path", r.URL.Path),
				slog.String("request_id", reqID))
			
			// Return network error
			render.Render(w, r, license.ErrNetwork(err))
			return
		}

		if !valid {
			lv.logger.Warn("license validation failed",
				slog.String("path", r.URL.Path),
				slog.String("request_id", reqID))
			
			// Return not activated error
			render.Render(w, r, license.ErrNotActivated)
			return
		}

		// Update cache
		lv.updateCache(true)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// shouldExcludePath checks if a path should be excluded from validation
func (lv *LicenseValidator) shouldExcludePath(path string) bool {
	// Check exact matches
	for _, excluded := range lv.excludePaths {
		if path == excluded {
			return true
		}
	}
	
	// Check prefix matches
	for _, prefix := range lv.excludePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	
	return false
}

// isCacheValid checks if the cached validation result is still valid
func (lv *LicenseValidator) isCacheValid() bool {
	lv.cache.mu.RLock()
	defer lv.cache.mu.RUnlock()
	
	if time.Since(lv.cache.checkedAt) > lv.cache.ttl {
		return false
	}
	
	return lv.cache.valid
}

// updateCache updates the cached validation result
func (lv *LicenseValidator) updateCache(valid bool) {
	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()
	
	lv.cache.valid = valid
	lv.cache.checkedAt = time.Now()
}

// validateLicense performs the actual license validation
func (lv *LicenseValidator) validateLicense(ctx context.Context) (bool, error) {
	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Create a channel for the result
	resultCh := make(chan struct {
		valid bool
		err   error
	}, 1)
	
	// Run validation in goroutine to respect context
	go func() {
		valid, err := lv.manager.ValidateLicense()
		resultCh <- struct {
			valid bool
			err   error
		}{valid, err}
	}()
	
	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case result := <-resultCh:
		return result.valid, result.err
	}
}

// AddExcludePath adds a path to be excluded from license validation
func (lv *LicenseValidator) AddExcludePath(path string) {
	lv.excludePaths = append(lv.excludePaths, path)
}

// AddExcludePrefix adds a path prefix to be excluded from license validation
func (lv *LicenseValidator) AddExcludePrefix(prefix string) {
	lv.excludePrefixes = append(lv.excludePrefixes, prefix)
}

// SetCacheTTL sets the cache time-to-live duration
func (lv *LicenseValidator) SetCacheTTL(ttl time.Duration) {
	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()
	lv.cache.ttl = ttl
}

// InvalidateCache invalidates the cached validation result
func (lv *LicenseValidator) InvalidateCache() {
	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()
	lv.cache.checkedAt = time.Time{}
}