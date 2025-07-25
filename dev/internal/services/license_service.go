package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"isxcli/internal/license"
)

// LicenseService provides business logic for license operations
type LicenseService interface {
	GetStatus(ctx context.Context) (*LicenseStatusResponse, error)
	Activate(ctx context.Context, key string) error
	ValidateWithContext(ctx context.Context) (bool, error)
}

// LicenseStatusResponse represents the standardized license status response
type LicenseStatusResponse struct {
	// RFC 7807 Problem Details
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`

	// Application-specific fields
	LicenseStatus string                `json:"license_status"` // active|expired|not_activated|critical|warning
	Message       string                `json:"message"`
	DaysLeft      int                   `json:"days_left,omitempty"`
	LicenseInfo   *license.LicenseInfo  `json:"license_info,omitempty"`
	TraceID       string                `json:"trace_id"`
}

// licenseService implements LicenseService
type licenseService struct {
	manager *license.Manager
	logger  *slog.Logger
}

// NewLicenseService creates a new license service
func NewLicenseService(manager *license.Manager, logger *slog.Logger) LicenseService {
	return &licenseService{
		manager: manager,
		logger:  logger.With(slog.String("service", "license")),
	}
}

// GetStatus returns the current license status with proper error handling
func (s *licenseService) GetStatus(ctx context.Context) (*LicenseStatusResponse, error) {
	start := time.Now()
	traceID := middleware.GetReqID(ctx)
	
	// Log the operation start
	s.logger.InfoContext(ctx, "license status check started",
		slog.String("trace_id", traceID),
		slog.String("operation", "get_status"),
	)
	
	// Get license info from manager
	info, status, err := s.manager.GetLicenseStatus()
	
	// Log the result
	s.logger.InfoContext(ctx, "license status check completed",
		slog.String("trace_id", traceID),
		slog.String("operation", "get_status"),
		slog.Duration("latency", time.Since(start)),
		slog.String("status", status),
		slog.Bool("has_error", err != nil),
	)
	
	// Handle not activated case
	if status == "Not Activated" {
		return &LicenseStatusResponse{
			Type:          "/license/not-activated",
			Title:         "License Not Activated",
			Status:        200, // Not an error, just a state
			Detail:        "No license has been activated on this system",
			Instance:      fmt.Sprintf("/api/license/status#%s", traceID),
			LicenseStatus: "not_activated",
			Message:       "No license activated. Please activate a license to use this application.",
			TraceID:       traceID,
		}, nil
	}
	
	// Handle error cases
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get license status",
			slog.String("trace_id", traceID),
			slog.String("error", err.Error()),
		)
		
		return &LicenseStatusResponse{
			Type:          "/errors/license-check-failed",
			Title:         "License Check Failed",
			Status:        500,
			Detail:        "Unable to verify license status",
			Instance:      fmt.Sprintf("/api/license/status#%s", traceID),
			LicenseStatus: "error",
			Message:       "Unable to retrieve license information. Please contact support.",
			TraceID:       traceID,
		}, nil
	}
	
	// Calculate days left
	var daysLeft int
	var message string
	
	if info != nil {
		daysLeft = int(time.Until(info.ExpiryDate).Hours() / 24)
		
		// Determine license status and message
		licenseStatus := s.determineLicenseStatus(status, daysLeft)
		message = s.generateStatusMessage(licenseStatus, daysLeft)
		
		// Log license details
		s.logger.InfoContext(ctx, "license details",
			slog.String("trace_id", traceID),
			slog.String("license_status", licenseStatus),
			slog.Int("days_left", daysLeft),
			slog.Time("expiry_date", info.ExpiryDate),
		)
		
		return &LicenseStatusResponse{
			Status:        200,
			LicenseStatus: licenseStatus,
			Message:       message,
			DaysLeft:      daysLeft,
			LicenseInfo:   info,
			TraceID:       traceID,
		}, nil
	}
	
	// No license info available
	return &LicenseStatusResponse{
		Type:          "/license/not-found",
		Title:         "License Not Found",
		Status:        200,
		Detail:        "No license information available",
		Instance:      fmt.Sprintf("/api/license/status#%s", traceID),
		LicenseStatus: "not_activated",
		Message:       "No license found. Please activate a license.",
		TraceID:       traceID,
	}, nil
}

// Activate activates a license with the given key
func (s *licenseService) Activate(ctx context.Context, key string) error {
	start := time.Now()
	traceID := middleware.GetReqID(ctx)
	
	// Mask the key for logging
	maskedKey := maskLicenseKey(key)
	
	s.logger.InfoContext(ctx, "license activation started",
		slog.String("trace_id", traceID),
		slog.String("operation", "activate"),
		slog.String("license_key", maskedKey),
	)
	
	// Perform activation
	err := s.manager.ActivateLicense(key)
	
	// Log the result
	if err != nil {
		s.logger.ErrorContext(ctx, "license activation failed",
			slog.String("trace_id", traceID),
			slog.String("operation", "activate"),
			slog.String("license_key", maskedKey),
			slog.Duration("latency", time.Since(start)),
			slog.String("error", err.Error()),
		)
		
		// Wrap error with context
		return fmt.Errorf("activation failed: %w", err)
	}
	
	s.logger.InfoContext(ctx, "license activation succeeded",
		slog.String("trace_id", traceID),
		slog.String("operation", "activate"),
		slog.String("license_key", maskedKey),
		slog.Duration("latency", time.Since(start)),
	)
	
	return nil
}

// ValidateWithContext validates the current license
func (s *licenseService) ValidateWithContext(ctx context.Context) (bool, error) {
	start := time.Now()
	traceID := middleware.GetReqID(ctx)
	
	s.logger.DebugContext(ctx, "license validation started",
		slog.String("trace_id", traceID),
		slog.String("operation", "validate"),
	)
	
	// Create a channel for the result
	type result struct {
		valid bool
		err   error
	}
	
	resultCh := make(chan result, 1)
	
	// Run validation in goroutine to respect context
	go func() {
		valid, err := s.manager.ValidateLicense()
		resultCh <- result{valid, err}
	}()
	
	// Wait for result or context cancellation
	select {
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "license validation cancelled",
			slog.String("trace_id", traceID),
			slog.Duration("latency", time.Since(start)),
		)
		return false, ctx.Err()
		
	case res := <-resultCh:
		s.logger.DebugContext(ctx, "license validation completed",
			slog.String("trace_id", traceID),
			slog.String("operation", "validate"),
			slog.Duration("latency", time.Since(start)),
			slog.Bool("valid", res.valid),
			slog.Bool("has_error", res.err != nil),
		)
		return res.valid, res.err
	}
}

// Helper functions

// determineLicenseStatus determines the license status based on the manager status and days left
func (s *licenseService) determineLicenseStatus(managerStatus string, daysLeft int) string {
	// Normalize the status
	switch managerStatus {
	case "Expired":
		return "expired"
	case "Critical":
		return "critical"
	case "Warning":
		return "warning"
	case "Active", "Activated", "Valid":
		// Further categorize based on days left
		if daysLeft <= 0 {
			return "expired"
		} else if daysLeft <= 7 {
			return "critical"
		} else if daysLeft <= 30 {
			return "warning"
		}
		return "active"
	default:
		return "not_activated"
	}
}

// generateStatusMessage generates a user-friendly message based on the license status
func (s *licenseService) generateStatusMessage(status string, daysLeft int) string {
	switch status {
	case "expired":
		return "Your license has expired. Please renew to continue using the application."
	case "critical":
		return fmt.Sprintf("Your license expires in %d days. Please renew soon to avoid interruption.", daysLeft)
	case "warning":
		return fmt.Sprintf("Your license expires in %d days. Consider renewing to ensure continued access.", daysLeft)
	case "active":
		return fmt.Sprintf("License is active. %d days remaining until expiration.", daysLeft)
	default:
		return "License status unknown. Please contact support."
	}
}

// maskLicenseKey masks a license key for safe logging
func maskLicenseKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..."
}