package http

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	licenseErrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/services"
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// LicenseHandler handles license-related HTTP requests with clean architecture
type LicenseHandler struct {
	service services.LicenseService
	logger  *slog.Logger
}

// NewLicenseHandler creates a new license handler
func NewLicenseHandler(service services.LicenseService, logger *slog.Logger) *LicenseHandler {
	return &LicenseHandler{
		service: service,
		logger:  logger.With(slog.String("handler", "license")),
	}
}

// LicenseActivationRequest is an alias to the canonical contract type
type LicenseActivationRequest = domain.LicenseActivationRequest

// LicenseTransferRequest represents the license transfer request payload
type LicenseTransferRequest struct {
	LicenseKey string `json:"license_key" validate:"required"`
	Force      bool   `json:"force,omitempty"`
}

// CacheInvalidationRequest represents cache invalidation request
type CacheInvalidationRequest struct {
	Reason string `json:"reason,omitempty"`
}

// BindLicenseActivationRequest validates license activation requests
func BindLicenseActivationRequest(r *http.Request, req *LicenseActivationRequest) error {
	if req.LicenseKey == "" {
		return errors.New("license_key is required")
	}
	// Use the correct validation function that handles both formats (with and without dashes)
	if !isValidLicenseKeyFormat(req.LicenseKey) {
		return errors.New("invalid license key format. Expected: ISX-XXXX-XXXX-XXXX-XXXX or ISX1MXXXXX, ISX3MXXXXX, ISX6MXXXXX, ISX1YXXXXX")
	}
	if req.Email != "" && !isValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	return nil
}

// Bind implements the render.Binder interface for transfer request validation
func (l *LicenseTransferRequest) Bind(r *http.Request) error {
	if l.LicenseKey == "" {
		return errors.New("license_key is required")
	}
	if len(l.LicenseKey) < 8 {
		return errors.New("license_key is too short")
	}
	return nil
}

// Bind implements the render.Binder interface for cache invalidation
func (c *CacheInvalidationRequest) Bind(r *http.Request) error {
	// No required fields for cache invalidation
	return nil
}

// LicenseActivationResponse represents the license activation response
type LicenseActivationResponse struct {
	Success      bool                  `json:"success"`
	Message      string                `json:"message"`
	LicenseInfo  *services.LicenseStatusResponse `json:"license_info,omitempty"`
	TraceID      string                `json:"trace_id"`
	Timestamp    time.Time             `json:"timestamp"`
	ActivatedAt  *time.Time            `json:"activated_at,omitempty"`
}

// Routes returns a chi router for license endpoints with comprehensive API
func (h *LicenseHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	// Apply timeout middleware to all license routes
	r.Use(middleware.Timeout(30 * time.Second))
	
	// Basic license operations
	r.Get("/status", h.GetStatus)
	r.Get("/detailed", h.GetDetailedStatus)
	r.Post("/activate", h.Activate)
	
	// License stacking and management
	r.Get("/check-existing", h.CheckExistingLicense)
	r.Get("/details", h.GetLicenseDetails)
	r.Get("/history", h.GetActivationHistory)
	r.Post("/backup", h.BackupCurrentLicense)
	
	// Advanced license operations
	r.Get("/renewal", h.GetRenewalStatus)
	r.Post("/transfer", h.TransferLicense)
	r.Get("/metrics", h.GetMetrics)
	r.Post("/invalidate-cache", h.InvalidateCache)
	
	// Debug endpoints
	r.Get("/debug", h.GetDebugInfo)
	
	return r
}

// GetStatus handles GET /api/license/status with comprehensive observability
func (h *LicenseHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	tracer := otel.Tracer("license-handler")
	start := time.Now()
	
	// Start OpenTelemetry span for license status check
	ctx, span := tracer.Start(ctx, "license_handler.get_status",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/api/license/status"),
			attribute.String("request_id", reqID),
			attribute.String("component", "license_handler"),
			attribute.String("operation", "get_status"),
		),
	)
	defer span.End()
	
	// Log request start with comprehensive context
	h.logger.InfoContext(ctx, "license status request started",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
		slog.String("operation", "get_status"),
		slog.String("user_agent", r.UserAgent()),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("x_forwarded_for", r.Header.Get("X-Forwarded-For")),
	)
	
	// Get status with timeout and detailed logging
	statusCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	h.logger.DebugContext(ctx, "calling license service get_status",
		slog.String("request_id", reqID),
		slog.String("timeout", "5s"),
	)
	
	response, err := h.service.GetStatus(statusCtx)
	latency := time.Since(start)
	
	// Add span attributes for the result
	span.SetAttributes(
		attribute.Int64("request.latency_ms", latency.Milliseconds()),
		attribute.Bool("request.success", err == nil),
	)
	
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("error.type", "service_error"),
			attribute.String("error.message", err.Error()),
		)
		
		h.logger.ErrorContext(ctx, "license status request failed",
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
			slog.Duration("latency", latency),
			slog.String("error", err.Error()),
		)
		
		h.handleError(w, r, err)
		return
	}
	
	// Log successful response with status details
	span.SetAttributes(
		attribute.String("license.status", response.LicenseStatus),
		attribute.Int("license.days_left", response.DaysLeft),
		attribute.Bool("license.has_info", response.LicenseInfo != nil),
	)
	
	h.logger.InfoContext(ctx, "license status request completed successfully",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
		slog.Duration("latency", latency),
		slog.String("license_status", response.LicenseStatus),
		slog.Int("days_left", response.DaysLeft),
		slog.Bool("has_license_info", response.LicenseInfo != nil),
		slog.String("message", response.Message),
	)
	
	// Validate response before sending
	if response == nil {
		h.logger.ErrorContext(ctx, "nil response from license service",
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/invalid-response",
			"Invalid Response",
			"Received invalid response from license service",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", infrastructure.TraceIDFromContext(ctx))
		
		render.Render(w, r, problem)
		return
	}
	
	// Validate license status is present
	if response.LicenseStatus == "" {
		h.logger.WarnContext(ctx, "empty license status in response",
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))
		
		// Set a default status
		response.LicenseStatus = "not_activated"
		response.Message = "License status unavailable"
	}
	
	// Add span event for successful response
	infrastructure.AddSpanEvent(ctx, "license.status.success", map[string]interface{}{
		"license_status": response.LicenseStatus,
		"days_left":      response.DaysLeft,
		"has_info":       response.LicenseInfo != nil,
		"component":      "license_handler",
		"operation":      "get_status",
	})
	
	// Return the standardized response
	render.JSON(w, r, response)
}

// Activate handles POST /api/license/activate
func (h *LicenseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	tracer := otel.Tracer("license-handler")
	
	// Start OpenTelemetry span for license activation
	ctx, span := tracer.Start(ctx, "license_handler.activate",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/api/license/activate"),
			attribute.String("request_id", reqID),
			attribute.String("component", "license_handler"),
		),
	)
	defer span.End()
	
	// Decode and validate request
	data := &LicenseActivationRequest{}
	if err := render.Decode(r, data); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.type", "request_decode"))
		
		h.logger.ErrorContext(ctx, "failed to decode license activation request",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-request",
			"Invalid Request",
			err.Error(),
			"/api/license/activate#"+reqID,
		).WithExtension("trace_id", infrastructure.TraceIDFromContext(ctx))
		
		render.Render(w, r, problem)
		return
	}
	
	// Validate the request
	if err := BindLicenseActivationRequest(r, data); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.type", "request_validation"))
		
		h.logger.ErrorContext(ctx, "failed to bind license activation request",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-request",
			"Invalid Request",
			err.Error(),
			"/api/license/activate#"+reqID,
		).WithExtension("trace_id", infrastructure.TraceIDFromContext(ctx))
		
		render.Render(w, r, problem)
		return
	}
	
	// Additional validation for license key
	if len(data.LicenseKey) == 0 {
		span.SetAttributes(attribute.String("error.type", "empty_license_key"))
		
		h.logger.WarnContext(ctx, "empty license key provided",
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/empty-license-key",
			"Empty License Key",
			"License key cannot be empty",
			"/api/license/activate#"+reqID,
		).WithExtension("trace_id", infrastructure.TraceIDFromContext(ctx)).
			WithExtension("validation_field", "license_key")
		
		render.Render(w, r, problem)
		return
	}
	
	// Validate license key format
	if !isValidLicenseKeyFormat(data.LicenseKey) {
		span.SetAttributes(attribute.String("error.type", "invalid_license_format"))
		
		h.logger.WarnContext(ctx, "invalid license key format",
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
			slog.String("key_length", fmt.Sprintf("%d", len(data.LicenseKey))))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-license-format",
			"Invalid License Format",
			"License key must be in format: ISX1Y-XXXXX-XXXXX-XXXXX-XXXXX",
			"/api/license/activate#"+reqID,
		).WithExtension("trace_id", infrastructure.TraceIDFromContext(ctx)).
			WithExtension("expected_format", "ISX1Y-XXXXX-XXXXX-XXXXX-XXXXX").
			WithExtension("validation_regex", "^ISX(1M|3M|6M|1Y)[A-Z0-9]{10,}$")
		
		render.Render(w, r, problem)
		return
	}
	
	// Add license key attributes to span (masked for security)
	maskedKey := maskLicenseKeyForLogging(data.LicenseKey)
	span.SetAttributes(
		attribute.String("license.key_prefix", maskedKey),
		attribute.String("license.operation", "activation"),
	)
	
	// Activate license with timeout
	activateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	err := h.service.Activate(activateCtx, data.LicenseKey)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("license.result", "failure"),
			attribute.String("error.type", classifyLicenseError(err)),
		)
		h.handleError(w, r, err)
		return
	}
	
	// Record successful activation
	span.SetAttributes(
		attribute.String("license.result", "success"),
		attribute.Bool("license.activated", true),
	)
	infrastructure.AddSpanEvent(ctx, "license.activation.success", map[string]interface{}{
		"license_key_hash": hashLicenseKeyForAudit(data.LicenseKey),
		"component": "license_handler",
		"operation": "activation",
	})
	
	// Get updated license status after activation
	statusCtx, statusCancel := context.WithTimeout(ctx, 5*time.Second)
	defer statusCancel()
	
	licenseStatus, statusErr := h.service.GetStatus(statusCtx)
	if statusErr != nil {
		h.logger.WarnContext(ctx, "failed to get license status after activation",
			slog.String("error", statusErr.Error()),
			slog.String("request_id", reqID))
	}
	
	// Success response with license information
	now := time.Now()
	
	// Determine if this was a reactivation based on license status
	message := "License activated successfully. You can now access all Iraqi Investor features."
	if licenseStatus != nil && strings.Contains(strings.ToLower(licenseStatus.Message), "reactivat") {
		message = "License reactivated successfully on this device. You can now access all Iraqi Investor features."
	}
	
	response := LicenseActivationResponse{
		Success:     true,
		Message:     message,
		LicenseInfo: licenseStatus,
		TraceID:     infrastructure.TraceIDFromContext(ctx),
		Timestamp:   now,
		ActivatedAt: &now,
	}
	
	render.JSON(w, r, response)
}

// GetDetailedStatus handles GET /api/license/detailed
func (h *LicenseHandler) GetDetailedStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get detailed status with timeout
	statusCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	response, err := h.service.GetDetailedStatus(statusCtx)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	render.JSON(w, r, response)
}

// GetRenewalStatus handles GET /api/license/renewal
func (h *LicenseHandler) GetRenewalStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	statusCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	response, err := h.service.CheckRenewalStatus(statusCtx)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	render.JSON(w, r, response)
}

// TransferLicense handles POST /api/license/transfer
func (h *LicenseHandler) TransferLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	// Decode and validate request
	data := &LicenseTransferRequest{}
	if err := render.Bind(r, data); err != nil {
		h.logger.ErrorContext(ctx, "failed to bind license transfer request",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-request",
			"Invalid Request",
			err.Error(),
			"/api/license/transfer#"+reqID,
		).WithExtension("trace_id", reqID)
		
		render.Render(w, r, problem)
		return
	}
	
	// Transfer license with timeout
	transferCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	
	err := h.service.TransferLicense(transferCtx, data.LicenseKey, data.Force)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	// Get updated license status after transfer
	statusCtx, statusCancel := context.WithTimeout(ctx, 5*time.Second)
	defer statusCancel()
	
	licenseStatus, statusErr := h.service.GetStatus(statusCtx)
	if statusErr != nil {
		h.logger.WarnContext(ctx, "failed to get license status after transfer",
			slog.String("error", statusErr.Error()),
			slog.String("request_id", reqID))
	}
	
	// Success response with license information
	now := time.Now()
	response := LicenseActivationResponse{
		Success:     true,
		Message:     "License transferred successfully to this machine.",
		LicenseInfo: licenseStatus,
		TraceID:     reqID,
		Timestamp:   now,
		ActivatedAt: &now,
	}
	
	render.JSON(w, r, response)
}

// GetMetrics handles GET /api/license/metrics
func (h *LicenseHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	metricsCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	metrics, err := h.service.GetValidationMetrics(metricsCtx)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	// Wrap metrics in standard response format
	response := struct {
		Success bool                         `json:"success"`
		Data    *services.ValidationMetrics `json:"data"`
		TraceID string                       `json:"trace_id"`
	}{
		Success: true,
		Data:    metrics,
		TraceID: middleware.GetReqID(ctx),
	}
	
	render.JSON(w, r, response)
}

// InvalidateCache handles POST /api/license/invalidate-cache
func (h *LicenseHandler) InvalidateCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	// Decode request (optional reason)
	data := &CacheInvalidationRequest{}
	render.Bind(r, data) // Ignore errors since reason is optional
	
	h.logger.InfoContext(ctx, "cache invalidation requested",
		slog.String("request_id", reqID),
		slog.String("reason", data.Reason))
	
	// Invalidate cache
	err := h.service.InvalidateCache(ctx)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	// Success response
	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		TraceID string `json:"trace_id"`
	}{
		Success: true,
		Message: "License validation cache invalidated successfully.",
		TraceID: reqID,
	}
	
	render.JSON(w, r, response)
}

// GetDebugInfo handles GET /api/license/debug - returns diagnostic information
func (h *LicenseHandler) GetDebugInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	tracer := otel.Tracer("license-handler")
	
	// Start OpenTelemetry span for debug info request
	ctx, span := tracer.Start(ctx, "license_handler.get_debug_info",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/api/license/debug"),
			attribute.String("request_id", reqID),
			attribute.String("component", "license_handler"),
			attribute.String("operation", "debug_info"),
		),
	)
	defer span.End()
	
	h.logger.InfoContext(ctx, "license debug info requested",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
		slog.String("operation", "debug_info"),
	)
	
	// Get debug info with timeout
	debugCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	debugInfo, err := h.service.GetDebugInfo(debugCtx)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(
			attribute.String("error.type", "service_error"),
			attribute.String("error.message", err.Error()),
		)
		
		h.logger.ErrorContext(ctx, "failed to get license debug info",
			slog.String("request_id", reqID),
			slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
			slog.String("error", err.Error()),
		)
		
		h.handleError(w, r, err)
		return
	}
	
	// Log successful response
	span.SetAttributes(
		attribute.Bool("license.file_exists", debugInfo.FileExists),
		attribute.String("license.file_path", debugInfo.FilePath),
		attribute.Bool("license.is_readable", debugInfo.IsReadable),
	)
	
	h.logger.InfoContext(ctx, "license debug info retrieved successfully",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)),
		slog.String("license_path", debugInfo.FilePath),
		slog.Bool("file_exists", debugInfo.FileExists),
		slog.Bool("is_readable", debugInfo.IsReadable),
	)
	
	// Return the debug information
	render.JSON(w, r, debugInfo)
}

// Helper functions for validation

// isValidISXLicenseFormat validates the ISX license key format
func isValidISXLicenseFormat(licenseKey string) bool {
	// Strip dashes for consistent processing - accept both ISX1M-02L-YE1-F9Q and ISX1M02LYE1F9Q formats
	licenseKey = strings.ReplaceAll(licenseKey, "-", "")
	
	// Expected format: ISX{duration}{base64string}
	// Examples: ISX1M02LYE1F9QJHR9D7Z, ISX3MABC123DEF456, ISX6MXYZ789, ISX1YABC123DEF456
	
	if len(licenseKey) < 9 || len(licenseKey) > 50 {
		return false
	}
	
	// Validate prefix - must be one of the valid ISX prefixes
	validPrefixes := []string{"ISX1M", "ISX3M", "ISX6M", "ISX1Y"}
	var prefix string
	var suffix string
	
	// Find which prefix matches
	for _, vp := range validPrefixes {
		if strings.HasPrefix(licenseKey, vp) {
			prefix = vp
			suffix = licenseKey[len(vp):]
			break
		}
	}
	
	// If no valid prefix found
	if prefix == "" {
		return false
	}
	
	// Validate suffix - should be base64-like string (alphanumeric)
	if len(suffix) < 5 {
		return false
	}
	
	for _, char := range suffix {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || 
		     (char >= '0' && char <= '9')) {
			return false
		}
	}
	
	return true
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	
	// Basic email validation - contains @ and at least one dot after @
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}
	
	domain := email[atIndex+1:]
	if !strings.Contains(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	
	return true
}

// handleError centralizes error handling for the handler with comprehensive error mapping
func (h *LicenseHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = reqID
	}
	
	// Log error with comprehensive context
	h.logger.ErrorContext(ctx, "request failed",
		slog.String("error", err.Error()),
		slog.String("error_type", classifyLicenseError(err)),
		slog.String("request_id", reqID),
		slog.String("trace_id", traceID),
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.String("user_agent", r.UserAgent()),
		slog.String("remote_addr", r.RemoteAddr))
	
	// Handle specific error types with detailed responses
	var problem render.Renderer
	
	switch {
	// Context errors
	case errors.Is(err, context.DeadlineExceeded):
		problem = licenseErrors.NewProblemDetails(
			http.StatusGatewayTimeout,
			"/errors/timeout",
			"Request Timeout",
			"The request timed out while processing. Please try again.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("timeout_type", "deadline_exceeded")
		
	case errors.Is(err, context.Canceled):
		problem = licenseErrors.NewProblemDetails(
			http.StatusRequestTimeout,
			"/errors/request-canceled",
			"Request Canceled",
			"The request was canceled before completion.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("cancellation_reason", "client_disconnect")
	
	// File system errors
	case errors.Is(err, os.ErrNotExist):
		problem = licenseErrors.NewProblemDetails(
			http.StatusNotFound,
			"/errors/license-file-not-found",
			"License File Not Found",
			"No license file found. Please activate a license to continue.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("help_url", "/license")
	
	case errors.Is(err, os.ErrPermission):
		problem = licenseErrors.NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/permission-denied",
			"Permission Denied",
			"Unable to access license file due to permission issues.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("support_action", "contact_administrator")
	
	// Validation errors
	case errors.Is(err, licenseErrors.ErrInvalidLicenseFormat):
		problem = licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-license-format",
			"Invalid License Format",
			"License key must be in format: ISX1Y-XXXXX-XXXXX-XXXXX-XXXXX",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("expected_format", "ISX1Y-XXXXX-XXXXX-XXXXX-XXXXX").
			WithExtension("validation_regex", "^ISX(1M|3M|6M|1Y)[A-Z0-9]{10,}$")
	
	// Rate limiting
	case errors.Is(err, licenseErrors.ErrRateLimited):
		retryAfter := 900 // 15 minutes default
		w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
		problem = licenseErrors.NewProblemDetails(
			http.StatusTooManyRequests,
			"/errors/rate-limited",
			"Too Many Requests",
			"Too many license operations. Please wait before trying again.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("retry_after", retryAfter).
			WithExtension("limit_type", "license_operations")
	
	// Check for "already activated" error
	case strings.Contains(strings.ToLower(err.Error()), "already been activated") || 
	     strings.Contains(strings.ToLower(err.Error()), "already activated"):
		problem = licenseErrors.NewProblemDetails(
			http.StatusConflict,
			"/errors/license-already-activated",
			"License Already Activated",
			"This license has already been activated on another device. Please contact support to transfer the license.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("error_type", "already_activated").
			WithExtension("support_email", "support@isxpulse.com").
			WithExtension("transfer_info", "Contact support with your license key and proof of purchase to transfer this license.")
	
	// Check for rate limiting errors from Google Sheets
	case strings.Contains(strings.ToLower(err.Error()), "too many attempts"):
		problem = licenseErrors.NewProblemDetails(
			http.StatusTooManyRequests,
			"/errors/rate-limited",
			"Too Many Attempts",
			"The license server has temporarily blocked activation attempts. Please wait 5 minutes before trying again.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("error_type", "rate_limited").
			WithExtension("retry_after", 300) // 5 minutes
	
	// Check for access denied errors from Google Sheets
	case strings.Contains(strings.ToLower(err.Error()), "access denied"):
		problem = licenseErrors.NewProblemDetails(
			http.StatusForbidden,
			"/errors/access-denied",
			"Access Denied",
			"Your access has been temporarily blocked. Please wait 15 minutes or contact support.",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", traceID).
			WithExtension("error_type", "blacklisted").
			WithExtension("support_email", "support@isxpulse.com")
	
	// Reactivation-specific errors
	case errors.Is(err, licenseErrors.ErrLicenseReactivated):
		// This should be treated as a success case and handled in the activation flow
		// If it reaches here, something went wrong in the flow
		problem = licenseErrors.NewLicenseReactivatedResponse(nil, traceID)
		
	case errors.Is(err, licenseErrors.ErrReactivationLimitExceeded):
		problem = licenseErrors.NewReactivationLimitExceededError(nil, traceID)
		
	case errors.Is(err, licenseErrors.ErrAlreadyActivatedOnDevice):
		problem = licenseErrors.NewAlreadyActivatedOnDeviceError(nil, traceID)
	
	default:
		// Use the centralized error mapper for other errors
		problem = licenseErrors.MapLicenseError(err, traceID)
	}
	
	// Add common extensions to all problems
	if pd, ok := problem.(*licenseErrors.ProblemDetails); ok {
		pd.WithExtension("timestamp", time.Now().UTC()).
			WithExtension("request_id", reqID).
			WithExtension("path", r.URL.Path).
			WithExtension("method", r.Method)
	}
	
	// Render the error response
	render.Render(w, r, problem)
}

// Helper functions for observability

// maskLicenseKeyForLogging masks license key for secure logging
func maskLicenseKeyForLogging(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// hashLicenseKeyForAudit creates a secure hash for audit trails
func hashLicenseKeyForAudit(key string) string {
	if key == "" {
		return ""
	}
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)[:16] // First 16 chars for audit correlation
}

// classifyLicenseError categorizes license errors for observability
func classifyLicenseError(err error) string {
	if err == nil {
		return ""
	}
	
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "expired"):
		return "license_expired"
	case strings.Contains(errStr, "machine"):
		return "machine_mismatch"
	case strings.Contains(errStr, "not found"):
		return "license_not_found"
	case strings.Contains(errStr, "invalid"):
		return "invalid_license"
	case strings.Contains(errStr, "network"), strings.Contains(errStr, "timeout"):
		return "network_error"
	case strings.Contains(errStr, "rate limit"):
		return "rate_limited"
	case strings.Contains(errStr, "unauthorized"):
		return "unauthorized"
	case strings.Contains(errStr, "reactivation limit exceeded"):
		return "reactivation_limit_exceeded"
	case strings.Contains(errStr, "already activated on different device"):
		return "already_activated_different_device"
	case strings.Contains(errStr, "license reactivated"):
		return "license_reactivated"
	default:
		return "unknown_error"
	}
}

// isValidLicenseKeyFormat validates license key format
func isValidLicenseKeyFormat(key string) bool {
	upperKey := strings.ToUpper(strings.TrimSpace(key))
	
	// Check scratch card format first: ISX-XXXX-XXXX-XXXX-XXXX
	if strings.Contains(upperKey, "-") {
		// Validate exact scratch card pattern
		parts := strings.Split(upperKey, "-")
		if len(parts) == 5 && parts[0] == "ISX" {
			// Each segment after ISX should be 4 alphanumeric characters
			for i := 1; i < 5; i++ {
				if len(parts[i]) != 4 {
					return false
				}
				for _, ch := range parts[i] {
					if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
						return false
					}
				}
			}
			return true
		}
	}
	
	// Check standard format (no dashes): ISX1MXXXXX, ISX3MXXXXX, ISX6MXXXXX, ISX1YXXXXX
	key = strings.ReplaceAll(upperKey, "-", "")
	
	if len(key) < 9 {
		return false
	}
	
	if !strings.HasPrefix(key, "ISX") {
		return false
	}
	
	// Check if it's a standard format with duration code
	durationCodes := []string{"1M", "3M", "6M", "1Y"}
	for _, dur := range durationCodes {
		if strings.HasPrefix(key[3:], dur) {
			// Validate the rest is alphanumeric
			for i := 5; i < len(key); i++ {
				ch := key[i]
				if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
					return false
				}
			}
			return true
		}
	}
	
	// If no duration code found but starts with ISX and has 16+ chars, might be scratch card without dashes
	if len(key) == 19 { // ISX + 16 chars = 19 total
		for i := 3; i < len(key); i++ {
			ch := key[i]
			if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
				return false
			}
		}
		return true
	}
	
	return false
}

// CheckExistingLicense handles GET /api/license/check-existing
// Returns information about any existing license before activation
func (h *LicenseHandler) CheckExistingLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	tracer := otel.Tracer("license-handler")
	
	ctx, span := tracer.Start(ctx, "CheckExistingLicense",
		trace.WithAttributes(
			attribute.String("request.id", reqID),
		),
	)
	defer span.End()
	
	h.logger.InfoContext(ctx, "checking existing license",
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	
	// Check for existing license
	existingInfo, err := h.service.CheckExistingLicense()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to check existing license",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		span.RecordError(err)
		span.SetStatus(1, err.Error())
		
		h.handleError(w, r, err)
		return
	}
	
	// Return existing license info
	response := map[string]interface{}{
		"has_license":    existingInfo.HasLicense,
		"days_remaining": existingInfo.DaysRemaining,
		"expiry_date":    existingInfo.ExpiryDate,
		"license_key":    existingInfo.LicenseKey,
		"status":         existingInfo.Status,
		"is_expired":     existingInfo.IsExpired,
		"trace_id":       reqID,
		"timestamp":      time.Now(),
	}
	
	render.JSON(w, r, response)
}

// GetLicenseDetails handles GET /api/license/details
// Returns comprehensive license information including stacking history
func (h *LicenseHandler) GetLicenseDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	h.logger.InfoContext(ctx, "getting license details",
		slog.String("request_id", reqID),
	)
	
	// Get license details from service
	details, err := h.service.GetLicenseDetails()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get license details",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		h.handleError(w, r, err)
		return
	}
	
	render.JSON(w, r, details)
}

// GetActivationHistory handles GET /api/license/history
// Returns activation history from audit logs
func (h *LicenseHandler) GetActivationHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	h.logger.InfoContext(ctx, "getting activation history",
		slog.String("request_id", reqID),
	)
	
	// Get activation history from service
	history, err := h.service.GetActivationHistory()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get activation history",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		h.handleError(w, r, err)
		return
	}
	
	response := map[string]interface{}{
		"history":   history,
		"count":     len(history),
		"trace_id":  reqID,
		"timestamp": time.Now(),
	}
	
	render.JSON(w, r, response)
}

// BackupCurrentLicense handles POST /api/license/backup
// Creates a backup of the current license before changes
func (h *LicenseHandler) BackupCurrentLicense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	h.logger.InfoContext(ctx, "creating license backup",
		slog.String("request_id", reqID),
	)
	
	// Create backup
	backupPath, err := h.service.BackupCurrentLicense()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to create license backup",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		h.handleError(w, r, err)
		return
	}
	
	response := map[string]interface{}{
		"success":     true,
		"backup_path": backupPath,
		"message":     "License backup created successfully",
		"trace_id":    reqID,
		"timestamp":   time.Now(),
	}
	
	render.JSON(w, r, response)
}