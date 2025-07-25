package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"isxcli/internal/license"
)

// LicenseHandler handles license-related HTTP requests
type LicenseHandler struct {
	manager *license.Manager
	logger  *slog.Logger
}

// NewLicenseHandler creates a new license handler
func NewLicenseHandler(manager *license.Manager, logger *slog.Logger) *LicenseHandler {
	return &LicenseHandler{
		manager: manager,
		logger:  logger.With(slog.String("handler", "license")),
	}
}

// LicenseActivationRequest represents the license activation request payload
type LicenseActivationRequest struct {
	LicenseKey string `json:"license_key" validate:"required"`
}

// LicenseActivationResponse represents the license activation response
type LicenseActivationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Activate handles license activation requests following Chi best practices
func (h *LicenseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	var req LicenseActivationRequest
	
	// Decode and validate request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.logger.ErrorContext(ctx, "failed to decode license activation request",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID))
		render.Render(w, r, license.ErrInvalidRequest("Invalid request format"))
		return
	}

	// Validate license key
	if req.LicenseKey == "" {
		render.Render(w, r, license.ErrInvalidRequest("License key is required"))
		return
	}

	// Safe key masking for logging
	maskedKey := maskLicenseKey(req.LicenseKey)
	
	h.logger.InfoContext(ctx, "processing license activation",
		slog.String("license_key", maskedKey),
		slog.String("request_id", reqID))
	
	// Activate license with timeout
	activateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	err := h.activateWithContext(activateCtx, req.LicenseKey)
	if err != nil {
		h.logger.ErrorContext(ctx, "license activation failed",
			slog.String("error", err.Error()),
			slog.String("license_key", maskedKey),
			slog.String("request_id", reqID))
		
		// Map specific errors to appropriate responses
		switch {
		case strings.Contains(err.Error(), "expired"):
			render.Render(w, r, license.ErrLicenseExpired)
		case strings.Contains(err.Error(), "machine"):
			render.Render(w, r, license.ErrMachineMismatch)
		case strings.Contains(err.Error(), "not found"):
			render.Render(w, r, license.ErrLicenseNotFound)
		case strings.Contains(err.Error(), "rate limit"):
			render.Render(w, r, license.ErrRateLimited)
		case strings.Contains(err.Error(), "network"):
			render.Render(w, r, license.ErrNetwork(err))
		default:
			render.Render(w, r, license.ErrInternal(err))
		}
		return
	}

	// Success response
	h.logger.InfoContext(ctx, "license activated successfully",
		slog.String("license_key", maskedKey),
		slog.String("request_id", reqID))
	
	render.JSON(w, r, LicenseActivationResponse{
		Success: true,
		Message: "License activated successfully. You can now use all features.",
	})
}

// LicenseStatusResponse represents the license status response
type LicenseStatusResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	LicenseInfo *license.LicenseInfo `json:"license_info,omitempty"`
	DaysLeft    int    `json:"days_left,omitempty"`
}

// GetStatus returns current license status following Chi patterns
func (h *LicenseHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	h.logger.DebugContext(ctx, "checking license status",
		slog.String("request_id", reqID))
	
	// Get status with timeout
	statusCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	info, status, err := h.getStatusWithContext(statusCtx)
	
	// Handle not activated case
	if status == "Not Activated" || status == "not_activated" {
		h.logger.DebugContext(ctx, "no license activated",
			slog.String("request_id", reqID))
		render.JSON(w, r, LicenseStatusResponse{
			Status:  "not_activated",
			Message: "No license activated. Please activate a license to use this application.",
		})
		return
	}
	
	// Handle errors
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get license status",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID))
		render.Render(w, r, license.ErrInternal(err))
		return
	}
	
	// Calculate days left if we have license info
	var daysLeft int
	var message string
	
	if info != nil {
		daysLeft = int(time.Until(info.ExpiryDate).Hours() / 24)
		
		switch status {
		case "Expired":
			message = "Your license has expired. Please renew to continue using the application."
		case "Critical":
			message = fmt.Sprintf("Your license expires in %d days. Please renew soon.", daysLeft)
		case "Warning":
			message = fmt.Sprintf("Your license expires in %d days.", daysLeft)
		case "Active":
			message = fmt.Sprintf("License is active. %d days remaining.", daysLeft)
		default:
			message = "License status verified."
		}
		
		h.logger.InfoContext(ctx, "license status retrieved",
			slog.String("status", status),
			slog.Int("days_left", daysLeft),
			slog.String("request_id", reqID))
	}
	
	render.JSON(w, r, LicenseStatusResponse{
		Status:      strings.ToLower(status),
		Message:     message,
		LicenseInfo: info,
		DaysLeft:    daysLeft,
	})
}

// Helper functions

// maskLicenseKey masks a license key for safe logging
func maskLicenseKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..."
}

// activateWithContext performs license activation with context support
func (h *LicenseHandler) activateWithContext(ctx context.Context, licenseKey string) error {
	// Create a channel for the result
	resultCh := make(chan error, 1)
	
	// Run activation in goroutine
	go func() {
		resultCh <- h.manager.ActivateLicense(licenseKey)
	}()
	
	// Wait for result or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-resultCh:
		return err
	}
}

// getStatusWithContext gets license status with context support
func (h *LicenseHandler) getStatusWithContext(ctx context.Context) (*license.LicenseInfo, string, error) {
	type result struct {
		info   *license.LicenseInfo
		status string
		err    error
	}
	
	resultCh := make(chan result, 1)
	
	// Run status check in goroutine
	go func() {
		info, status, err := h.manager.GetLicenseStatus()
		resultCh <- result{info, status, err}
	}()
	
	// Wait for result or context cancellation
	select {
	case <-ctx.Done():
		return nil, "error", ctx.Err()
	case res := <-resultCh:
		return res.info, res.status, res.err
	}
}