package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
)

// LicenseValidator provides license validation middleware with enhanced security and caching
type LicenseValidator struct {
	validate        func(ctx context.Context) (bool, error)
	logger          *slog.Logger
	cache           *validationCache
	excludePaths    []string
	excludePrefixes []string
	enabled         bool
	redirectOnFail  bool
	licensePageURL  string
	// OpenTelemetry metrics
	metrics         *MiddlewareMetrics
	// Validation mutex to prevent concurrent validations
	validationMu    sync.Mutex
}

// validationCache stores recent validation results with enhanced metadata
type validationCache struct {
	mu               sync.RWMutex
	valid            bool
	checkedAt        time.Time
	ttl              time.Duration
	lastError        error
	errorCount       int
	lastSuccess      time.Time
	validationID     string
	lastFileModTime  time.Time // Track license file modification time for automatic cache invalidation
}

// MiddlewareMetrics holds OpenTelemetry metrics for license middleware
type MiddlewareMetrics struct {
	RequestsTotal        metric.Int64Counter
	ValidationAttempts   metric.Int64Counter
	ValidationSuccess    metric.Int64Counter
	ValidationFailures   metric.Int64Counter
	ValidationDuration   metric.Float64Histogram
	CacheHits           metric.Int64Counter
	CacheMisses         metric.Int64Counter
	PathExclusions      metric.Int64Counter
	RedirectsTotal      metric.Int64Counter
}

// NewLicenseValidator creates a new license validation middleware with enhanced configuration
func NewLicenseValidator(manager LicenseManagerInterface, logger *slog.Logger) *LicenseValidator {
	return NewLicenseValidatorWithValidation(func(ctx context.Context) (bool, error) {
		return manager.ValidateLicense()
	}, logger)
}

// NewLicenseValidatorWithValidation creates a validator using a context-aware validation function.
func NewLicenseValidatorWithValidation(validateFn func(ctx context.Context) (bool, error), logger *slog.Logger) *LicenseValidator {
	return &LicenseValidator{
		validate:      validateFn,
		logger:         logger.With(slog.String("component", "license_middleware")),
		enabled:        true,
		redirectOnFail: true,
		licensePageURL: "/license",
		cache: &validationCache{
			ttl: 30 * time.Second, // Cache validation results for 30 seconds - improved for faster activation feedback
		},
		excludePaths: []string{
			// License pages - MUST remain accessible without license
			"/license",
			"/license/",
			"/api/license/activate",
			"/api/license/status",
			"/api/license/detailed",
			"/api/license/renewal",
			"/api/license/transfer",
			"/api/license/metrics",
			"/api/license/invalidate-cache",
			"/", // Root should remain accessible for landing page/license prompts

			// Infrastructure endpoints - Required for monitoring/health checks
			"/api/health",
			"/api/health/ready",
			"/api/health/live",
			"/api/version",
			"/metrics",

			// Static assets - Required for license page to render properly
			"/favicon.ico",
			"/robots.txt",
			"/manifest.json",
			"/404",
			"/500",
		},
		excludePrefixes: []string{
			// Static asset prefixes - Required for CSS/JS/images to load
			"/static/",
			"/templates/",
			"/_next/",        // Next.js static assets (required for license page)
			"/assets/",       // Frontend assets
			"/legacy/",       // Legacy routes
		},
	}
}

// Handler returns the middleware handler function with enhanced error handling and redirection
func (lv *LicenseValidator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tracer := otel.Tracer("license-middleware")
		
		// Start OpenTelemetry span for license validation
		ctx, span := tracer.Start(ctx, "license_middleware.validate",
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.Path),
				attribute.String("component", "license_middleware"),
			),
		)
		defer span.End()
		
		// Get request ID and trace ID for comprehensive logging
		reqID := middleware.GetReqID(ctx)
		traceID := infrastructure.TraceIDFromContext(ctx)
		if traceID == "" {
			traceID = reqID
		}
		
		// Record request metric
		if lv.metrics != nil {
			lv.metrics.RequestsTotal.Add(ctx, 1, metric.WithAttributes(
				attribute.String("path", r.URL.Path),
				attribute.String("method", r.Method),
			))
		}
		
		// Check if license validation is enabled
		if !lv.enabled {
			lv.logger.DebugContext(ctx, "license validation disabled",
				slog.String("path", r.URL.Path),
				slog.String("trace_id", traceID))
			next.ServeHTTP(w, r)
			return
		}
		
		// Check if path should be excluded from validation
		if lv.shouldExcludePath(r.URL.Path) {
			span.SetAttributes(
				attribute.String("license.validation", "excluded"),
				attribute.String("exclusion_reason", "path_excluded"),
			)

			// Record path exclusion metric
			if lv.metrics != nil {
				lv.metrics.PathExclusions.Add(ctx, 1, metric.WithAttributes(
					attribute.String("path", r.URL.Path),
					attribute.String("reason", "excluded_path"),
				))
			}

			lv.logger.DebugContext(ctx, "skipping license validation for excluded path",
				slog.String("path", r.URL.Path),
				slog.String("trace_id", traceID))
			next.ServeHTTP(w, r)
			return
		}

			// Check cached validation result with enhanced metadata
			if lv.isCacheValid() {
				span.SetAttributes(
					attribute.String("license.validation", "cached"),
				attribute.Bool("cache.hit", true),
				attribute.Int64("cache.age_seconds", int64(time.Since(lv.cache.checkedAt).Seconds())),
			)
			
			// Record cache hit metric
			if lv.metrics != nil {
				lv.metrics.CacheHits.Add(ctx, 1, metric.WithAttributes(
					attribute.String("component", "license_middleware"),
				))
			}
			
			lv.logger.DebugContext(ctx, "using cached license validation result",
				slog.String("trace_id", traceID),
				slog.String("cache_age", time.Since(lv.cache.checkedAt).String()))
			next.ServeHTTP(w, r)
			return
		}
		
		// Record cache miss metric
		if lv.metrics != nil {
			lv.metrics.CacheMisses.Add(ctx, 1, metric.WithAttributes(
				attribute.String("component", "license_middleware"),
			))
		}

		// Acquire validation lock to prevent concurrent validations
		lv.validationMu.Lock()
		defer lv.validationMu.Unlock()
		
		// Double-check cache after acquiring lock - another goroutine might have validated
		if lv.isCacheValid() {
			span.SetAttributes(
				attribute.String("license.validation", "cached_after_lock"),
				attribute.Bool("cache.hit", true),
			)
			
			// Record cache hit metric for double-check scenario
			if lv.metrics != nil {
				lv.metrics.CacheHits.Add(ctx, 1, metric.WithAttributes(
					attribute.String("component", "license_middleware"),
				))
			}
			
			lv.logger.DebugContext(ctx, "using cached license validation result after lock acquisition",
				slog.String("trace_id", traceID))
			next.ServeHTTP(w, r)
			return
		}

		// Perform license validation with timeout and error handling
		start := time.Now()
		valid, err := lv.validateLicense(ctx)
		validationDuration := time.Since(start)
		
		// Record validation metrics
		if lv.metrics != nil {
			lv.metrics.ValidationAttempts.Add(ctx, 1, metric.WithAttributes(
				attribute.String("component", "license_middleware"),
			))
			lv.metrics.ValidationDuration.Record(ctx, validationDuration.Seconds(), metric.WithAttributes(
				attribute.String("component", "license_middleware"),
			))
			
			if err == nil && valid {
				lv.metrics.ValidationSuccess.Add(ctx, 1)
			} else {
				lv.metrics.ValidationFailures.Add(ctx, 1)
			}
		}
		
		// Add span attributes for validation result
		span.SetAttributes(
			attribute.String("license.validation", "performed"),
			attribute.Bool("license.valid", valid),
			attribute.Bool("license.has_error", err != nil),
			attribute.Float64("license.duration_ms", float64(validationDuration.Milliseconds())),
		)
		
		// Log validation attempt with performance metrics
		lv.logger.InfoContext(ctx, "license validation performed",
			slog.String("trace_id", traceID),
			slog.String("path", r.URL.Path),
			slog.Duration("validation_duration", validationDuration),
			slog.Bool("valid", valid),
			slog.Bool("has_error", err != nil))
		
		if err != nil {
			// Record error in span
			span.RecordError(err)
			span.SetAttributes(attribute.String("error.type", classifyValidationError(err)))
			
			// Enhanced error logging with context
			lv.logger.ErrorContext(ctx, "license validation error",
				slog.String("error", err.Error()),
				slog.String("path", r.URL.Path),
				slog.String("trace_id", traceID),
				slog.Duration("validation_duration", validationDuration))
			
			// Update cache with error state
			lv.updateCacheWithError(err)
			
			// Handle different error types
			lv.handleValidationError(w, r, err, traceID)
			return
		}

		if !valid {
			lv.logger.WarnContext(ctx, "license validation failed",
				slog.String("path", r.URL.Path),
				slog.String("trace_id", traceID),
				slog.Duration("validation_duration", validationDuration))
			
			// Update cache with invalid state
			lv.updateCache(false)
			
			// Handle invalid license (redirect or error response)
			lv.handleInvalidLicense(w, r, traceID)
			return
		}

		// Update cache with successful validation
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

// isCacheValid checks if the cached validation result is still valid with enhanced logic
func (lv *LicenseValidator) isCacheValid() bool {
	lv.cache.mu.RLock()
	defer lv.cache.mu.RUnlock()

	// Only return cached result if it was valid
	// Invalid results should trigger re-validation more frequently
	if !lv.cache.valid {
		// For invalid results, use shorter TTL to enable rapid retry after activation
		shortTTL := time.Minute
		if time.Since(lv.cache.checkedAt) <= shortTTL {
			return true
		}
		return false
	}

	// Check if cache has expired for valid results
	if time.Since(lv.cache.checkedAt) > lv.cache.ttl {
		return false
	}

	return true
}

// updateCache updates the cached validation result with enhanced metadata
func (lv *LicenseValidator) updateCache(valid bool) {
	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()
	
	now := time.Now()
	lv.cache.valid = valid
	lv.cache.checkedAt = now
	lv.cache.lastError = nil
	lv.cache.validationID = fmt.Sprintf("val-%d", now.UnixNano())
	
	if valid {
		lv.cache.lastSuccess = now
		lv.cache.errorCount = 0
	}
}

// updateCacheWithError updates the cache when validation fails with an error
func (lv *LicenseValidator) updateCacheWithError(err error) {
	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()
	
	now := time.Now()
	lv.cache.valid = false
	lv.cache.checkedAt = now
	lv.cache.lastError = err
	lv.cache.errorCount++
	lv.cache.validationID = fmt.Sprintf("err-%d", now.UnixNano())
}

// validateLicense performs the actual license validation
func (lv *LicenseValidator) validateLicense(ctx context.Context) (bool, error) {
	// Add timeout to prevent hanging (aligned with test expectations)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Create a channel for the result
	resultCh := make(chan struct {
		valid bool
		err   error
	}, 1)

	// Run validation in goroutine to respect context
	go func() {
		valid, err := lv.validate(ctx)
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

// InvalidateCache invalidates the cached validation result with enhanced logging
func (lv *LicenseValidator) InvalidateCache() {
	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()

	// Record the state before invalidation for logging
	wasValid := lv.cache.valid
	hadError := lv.cache.lastError != nil
	validationID := lv.cache.validationID

	// Invalidate the cache completely
	lv.cache.checkedAt = time.Time{}
	lv.cache.valid = false
	lv.cache.lastError = nil
	lv.cache.errorCount = 0
	lv.cache.validationID = ""

	// Enhanced logging for debugging cache invalidation
	lv.logger.Info("License validation cache invalidated",
		slog.Bool("was_valid", wasValid),
		slog.Bool("had_error", hadError),
		slog.String("previous_validation_id", validationID),
		slog.Time("invalidation_time", time.Now()),
		slog.String("trigger", "manual_invalidation"),
		slog.Duration("cache_ttl_reset", lv.cache.ttl),
	)
}

// InvalidateCacheIfLicenseFileChanged checks if license file was modified and invalidates cache if needed
func (lv *LicenseValidator) InvalidateCacheIfLicenseFileChanged(licenseFilePath string) {
	// Get current license file info
	fileInfo, err := os.Stat(licenseFilePath)
	if err != nil {
		// If file doesn't exist or can't be accessed, invalidate cache to trigger re-validation
		lv.logger.Info("License file not accessible, invalidating cache",
			slog.String("license_file", licenseFilePath),
			slog.String("error", err.Error()),
		)
		lv.InvalidateCache()
		return
	}

	modTime := fileInfo.ModTime()

	lv.cache.mu.Lock()
	defer lv.cache.mu.Unlock()

	// If we haven't tracked file modification time yet, set it now
	if lv.cache.lastFileModTime.IsZero() {
		lv.cache.lastFileModTime = modTime
		lv.logger.Info("License file modification time tracked",
			slog.String("license_file", licenseFilePath),
			slog.Time("mod_time", modTime),
		)
		return
	}

	// If file was modified since last check, invalidate cache
	if !modTime.Equal(lv.cache.lastFileModTime) {
		oldModTime := lv.cache.lastFileModTime
		lv.cache.lastFileModTime = modTime

		// Perform cache invalidation outside of mutex lock to avoid deadlock
		// Use a goroutine to avoid blocking the file check
		go func() {
			lv.logger.Info("License file modification detected, invalidating cache",
				slog.String("license_file", licenseFilePath),
				slog.Time("previous_mod_time", oldModTime),
				slog.Time("current_mod_time", modTime),
				slog.Duration("time_since_change", time.Since(oldModTime)),
			)

			// Invalidate the cache
			lv.InvalidateCache()
		}()
	}
}

// handleValidationError handles different types of validation errors
func (lv *LicenseValidator) handleValidationError(w http.ResponseWriter, r *http.Request, err error, traceID string) {
	ctx := r.Context()
	
	// Check if this is a network/timeout error that should allow graceful degradation
	if isNetworkError(err) || isTimeoutError(err) {
		// For network errors, allow continued access for a grace period if we had a recent successful validation
		lv.cache.mu.RLock()
		hasRecentSuccess := !lv.cache.lastSuccess.IsZero() && time.Since(lv.cache.lastSuccess) < 24*time.Hour
		lv.cache.mu.RUnlock()
		
		if hasRecentSuccess {
			lv.logger.WarnContext(ctx, "license validation network error, allowing graceful degradation",
				slog.String("error", err.Error()),
				slog.String("trace_id", traceID),
				slog.Duration("time_since_last_success", time.Since(lv.cache.lastSuccess)))
			return // Allow request to continue
		}
	}
	
	// Check if this is an API request (JSON response expected)
	if isAPIRequest(r) {
		// Return RFC 7807 compliant error
		problem := errors.NewProblemDetails(
			http.StatusServiceUnavailable,
			"/errors/license-validation-failed",
			"License Validation Failed",
			"Unable to validate license. Please check your connection and try again.",
			fmt.Sprintf("%s#%s", r.URL.Path, traceID),
		).WithExtension("trace_id", traceID).
			WithExtension("error_type", "validation_error")
		
		render.Render(w, r, problem)
		return
	}
	
	// For HTML requests, redirect to license page
	if lv.redirectOnFail {
		lv.redirectToLicensePage(w, r, "validation_error")
		return
	}
	
	// Fallback: return error page
	http.Error(w, "License validation failed. Please activate your license.", http.StatusServiceUnavailable)
}

// handleInvalidLicense handles invalid license scenarios
func (lv *LicenseValidator) handleInvalidLicense(w http.ResponseWriter, r *http.Request, traceID string) {
	ctx := r.Context()
	
	lv.logger.InfoContext(ctx, "redirecting user due to invalid license",
		slog.String("path", r.URL.Path),
		slog.String("trace_id", traceID),
		slog.String("user_agent", r.UserAgent()))
	
	// Check if this is an API request
	if isAPIRequest(r) {
		// Return RFC 7807 compliant error for API requests
		problem := errors.NewProblemDetails(
			http.StatusPreconditionRequired,
			"/errors/license-not-activated",
			"License Not Activated",
			"No valid license found. Please activate a license to access this resource.",
			fmt.Sprintf("%s#%s", r.URL.Path, traceID),
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "LICENSE_NOT_ACTIVATED").
			WithExtension("redirect_url", lv.licensePageURL)
		
		render.Render(w, r, problem)
		return
	}
	
	// For HTML requests, redirect to license page
	if lv.redirectOnFail {
		lv.redirectToLicensePage(w, r, "not_activated")
		return
	}
	
	// Fallback: return error page
	http.Error(w, "License not activated. Please activate your license to continue.", http.StatusPreconditionRequired)
}

// redirectToLicensePage redirects the user to the license activation page
func (lv *LicenseValidator) redirectToLicensePage(w http.ResponseWriter, r *http.Request, reason string) {
	// Build redirect URL with context
	redirectURL := lv.licensePageURL
	if reason != "" {
		if strings.Contains(redirectURL, "?") {
			redirectURL += fmt.Sprintf("&reason=%s", reason)
		} else {
			redirectURL += fmt.Sprintf("?reason=%s", reason)
		}
	}
	
	// Add return URL for better UX
	if r.URL.Path != "/" && r.URL.Path != lv.licensePageURL {
		returnURL := r.URL.Path
		if r.URL.RawQuery != "" {
			returnURL += "?" + r.URL.RawQuery
		}
		if strings.Contains(redirectURL, "?") {
			redirectURL += fmt.Sprintf("&return=%s", returnURL)
		} else {
			redirectURL += fmt.Sprintf("?return=%s", returnURL)
		}
	}
	
	// Perform redirect
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// isAPIRequest checks if the request expects a JSON response
func isAPIRequest(r *http.Request) bool {
	// Check Accept header
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}
	
	// Check Content-Type header
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return true
	}
	
	// Check path prefix
	return strings.HasPrefix(r.URL.Path, "/api/")
}

// isNetworkError checks if the error is network-related
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "network") || 
		   strings.Contains(errStr, "connection") ||
		   strings.Contains(errStr, "timeout") ||
		   strings.Contains(errStr, "unreachable")
}

// isTimeoutError checks if the error is timeout-related
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return err == context.DeadlineExceeded || strings.Contains(err.Error(), "timeout")
}

// Configuration methods for enhanced middleware customization

// SetEnabled enables or disables license validation
func (lv *LicenseValidator) SetEnabled(enabled bool) {
	lv.enabled = enabled
}

// SetRedirectOnFail sets whether to redirect to license page on validation failure
func (lv *LicenseValidator) SetRedirectOnFail(redirect bool) {
	lv.redirectOnFail = redirect
}

// SetLicensePageURL sets the URL of the license activation page
func (lv *LicenseValidator) SetLicensePageURL(url string) {
	lv.licensePageURL = url
}

// GetCacheStats returns cache statistics for monitoring
func (lv *LicenseValidator) GetCacheStats() map[string]interface{} {
	lv.cache.mu.RLock()
	defer lv.cache.mu.RUnlock()
	
	return map[string]interface{}{
		"valid":              lv.cache.valid,
		"last_checked":       lv.cache.checkedAt,
		"ttl_seconds":        int(lv.cache.ttl.Seconds()),
		"error_count":        lv.cache.errorCount,
		"last_success":       lv.cache.lastSuccess,
		"last_error":         lv.cache.lastError,
		"validation_id":      lv.cache.validationID,
		"cache_age_seconds":  int(time.Since(lv.cache.checkedAt).Seconds()),
	}
}

// SetMetrics sets the OpenTelemetry metrics for the middleware
func (lv *LicenseValidator) SetMetrics(metrics *MiddlewareMetrics) {
	lv.metrics = metrics
}

// classifyValidationError categorizes validation errors for observability
func classifyValidationError(err error) string {
	if err == nil {
		return ""
	}
	
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "timeout"), strings.Contains(errStr, "deadline"):
		return "timeout"
	case strings.Contains(errStr, "network"), strings.Contains(errStr, "connection"):
		return "network_error"
	case strings.Contains(errStr, "invalid"), strings.Contains(errStr, "expired"):
		return "license_invalid"
	case strings.Contains(errStr, "machine"):
		return "machine_mismatch"
	default:
		return "unknown_error"
	}
}
