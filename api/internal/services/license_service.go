package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	licenseErrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
)

// LicenseService provides business logic for license operations with enhanced capabilities
type LicenseService interface {
	// Core operations
	GetStatus(ctx context.Context) (*LicenseStatusResponse, error)
	Activate(ctx context.Context, key string) error
	ValidateWithContext(ctx context.Context) (bool, error)
	
	// License stacking and management
	CheckExistingLicense() (*license.ExistingLicenseInfo, error)
	GetLicenseDetails() (interface{}, error)
	GetActivationHistory() ([]interface{}, error)
	BackupCurrentLicense() (string, error)
	
	// Enhanced operations
	GetDetailedStatus(ctx context.Context) (*DetailedLicenseStatusResponse, error)
	CheckRenewalStatus(ctx context.Context) (*RenewalStatusResponse, error)
	TransferLicense(ctx context.Context, key string, force bool) error
	GetValidationMetrics(ctx context.Context) (*ValidationMetrics, error)
	InvalidateCache(ctx context.Context) error
	
	// Debug operations
	GetDebugInfo(ctx context.Context) (*LicenseDebugInfo, error)
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
	
	// Iraqi Investor specific fields
	UserInfo      *UserInfo             `json:"user_info,omitempty"`
	Features      []string              `json:"features,omitempty"`
	Limitations   map[string]interface{} `json:"limitations,omitempty"`
	RenewalInfo   *RenewalInfo          `json:"renewal_info,omitempty"`
	BrandingInfo  *BrandingInfo         `json:"branding_info,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`
}

// DetailedLicenseStatusResponse provides comprehensive license information
type DetailedLicenseStatusResponse struct {
	LicenseStatusResponse
	
	// Additional detailed fields
	ActivationDate     *time.Time             `json:"activation_date,omitempty"`
	LastValidation     *time.Time             `json:"last_validation,omitempty"`
	ValidationCount    int64                  `json:"validation_count"`
	NetworkStatus      string                 `json:"network_status"`
	PerformanceMetrics *ValidationMetrics     `json:"performance_metrics,omitempty"`
	Recommendations    []string               `json:"recommendations,omitempty"`
}

// RenewalStatusResponse provides license renewal information
type RenewalStatusResponse struct {
	NeedsRenewal     bool      `json:"needs_renewal"`
	IsExpired        bool      `json:"is_expired"`
	DaysUntilExpiry  int       `json:"days_until_expiry"`
	ExpiryDate       time.Time `json:"expiry_date"`
	RenewalUrgency   string    `json:"renewal_urgency"` // low|medium|high|critical
	RenewalMessage   string    `json:"renewal_message"`
	ContactInfo      string    `json:"contact_info,omitempty"`
	TraceID          string    `json:"trace_id"`
}

// ValidationMetrics provides performance and reliability metrics
type ValidationMetrics struct {
	TotalValidations    int64         `json:"total_validations"`
	SuccessfulValidations int64       `json:"successful_validations"`
	FailedValidations   int64         `json:"failed_validations"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastValidationTime  time.Time     `json:"last_validation_time"`
	CacheHitRate        float64       `json:"cache_hit_rate"`
	NetworkErrors       int64         `json:"network_errors"`
	Uptime              time.Duration `json:"uptime"`
}

// UserInfo represents license user information
type UserInfo struct {
	Email         string    `json:"email,omitempty"`
	Company       string    `json:"company,omitempty"`
	Tier          string    `json:"tier,omitempty"` // basic|professional|enterprise
	ActivatedAt   time.Time `json:"activated_at,omitempty"`
	LastSeen      time.Time `json:"last_seen,omitempty"`
}

// RenewalInfo represents license renewal information
type RenewalInfo struct {
	NeedsRenewal    bool      `json:"needs_renewal"`
	IsExpired       bool      `json:"is_expired"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	ExpiryDate      time.Time `json:"expiry_date"`
	RenewalUrgency  string    `json:"renewal_urgency"` // low|medium|high|critical
	RenewalMessage  string    `json:"renewal_message"`
	ContactInfo     string    `json:"contact_info,omitempty"`
	RenewalURL      string    `json:"renewal_url,omitempty"`
}

// BrandingInfo represents Iraqi Investor branding information
type BrandingInfo struct {
	ApplicationName string `json:"application_name"`
	Version         string `json:"version"`
	BrandName       string `json:"brand_name"`
	LogoURL         string `json:"logo_url,omitempty"`
	WebsiteURL      string `json:"website_url,omitempty"`
	SupportEmail    string `json:"support_email,omitempty"`
}

// LicenseDebugInfo represents diagnostic information for license troubleshooting
type LicenseDebugInfo struct {
	FilePath        string                 `json:"file_path"`
	FileExists      bool                   `json:"file_exists"`
	IsReadable      bool                   `json:"is_readable"`
	FileSize        int64                  `json:"file_size,omitempty"`
	FilePermissions string                 `json:"file_permissions,omitempty"`
	FileModTime     *time.Time             `json:"file_mod_time,omitempty"`
	WorkingDir      string                 `json:"working_dir"`
	ExecPath        string                 `json:"exec_path"`
	ConfigPath      string                 `json:"config_path"`
	LicenseStatus   string                 `json:"license_status"`
	LastError       string                 `json:"last_error,omitempty"`
	Environment     map[string]string      `json:"environment,omitempty"`
	TraceID         string                 `json:"trace_id"`
	Timestamp       time.Time              `json:"timestamp"`
}

// licenseService implements LicenseService with enhanced capabilities
type licenseService struct {
	activator license.ActivationService
	validator license.ValidationService
	logger    *slog.Logger
	
	// Enhanced tracking
	startTime          time.Time
	validationCount    int64
	successCount       int64
	errorCount         int64
	lastValidation     time.Time
	totalResponseTime  time.Duration
}

// NewLicenseService creates a new license service with enhanced tracking
func NewLicenseService(manager license.ManagerInterface, logger *slog.Logger) LicenseService {
	if logger == nil {
		logger = slog.Default()
	}
	return &licenseService{
		activator: license.NewActivationService(manager),
		validator: license.NewValidationService(manager),
		logger:    logger.With(slog.String("service", "license")),
		startTime: time.Now(),
	}
}

// GetStatus returns the current license status with comprehensive observability
func (s *licenseService) GetStatus(ctx context.Context) (*LicenseStatusResponse, error) {
	start := time.Now()
	traceID := middleware.GetReqID(ctx)
	if traceID == "" {
		traceID = infrastructure.TraceIDFromContext(ctx)
	}
	
	// Log the operation start with detailed context
	s.logger.InfoContext(ctx, "license status check started",
		slog.String("trace_id", traceID),
		slog.String("operation", "get_status"),
		slog.String("component", "license_service"),
		slog.String("method", "GetStatus"),
	)
	
	// Track operation in performance metrics
	defer func() {
		s.validationCount++
		s.totalResponseTime += time.Since(start)
		s.lastValidation = time.Now()
	}()
	
	// Get license info from manager with detailed logging
	s.logger.DebugContext(ctx, "calling license validation status",
		slog.String("trace_id", traceID),
		slog.String("validator_type", fmt.Sprintf("%T", s.validator)),
	)
	
	info, status, err := s.validator.Status(ctx)
	operationLatency := time.Since(start)
	
	// Log the manager call result
	s.logger.InfoContext(ctx, "license manager call completed",
		slog.String("trace_id", traceID),
		slog.String("operation", "get_status"),
		slog.Duration("manager_latency", operationLatency),
		slog.String("manager_status", status),
		slog.Bool("has_error", err != nil),
		slog.Bool("has_info", info != nil),
	)
	
	// Update success/error counters
	if err != nil {
		s.errorCount++
	} else {
		s.successCount++
	}
	
	// Handle not activated case
	if status == "Not Activated" {
		return &LicenseStatusResponse{
			Type:          "/license/not-activated",
			Title:         "License Not Activated",
			Status:        200, // Not an error, just a state
			Detail:        "No license has been activated on this system",
			Instance:      fmt.Sprintf("/api/license/status#%s", traceID),
			LicenseStatus: "not_activated",
			Message:       "No license activated. Please activate a license to access Iraqi Investor features.",
			TraceID:       traceID,
			Features:      s.buildFeaturesList("not_activated"),
			Limitations:   s.buildLimitations("not_activated"),
			BrandingInfo:  s.buildBrandingInfo(),
			Timestamp:     time.Now(),
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
		// Use proper date calculation to avoid floating point precision issues
		now := time.Now()
		// Truncate both times to day precision for accurate day calculation
		nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		expiryDay := time.Date(info.ExpiryDate.Year(), info.ExpiryDate.Month(), info.ExpiryDate.Day(), 0, 0, 0, 0, info.ExpiryDate.Location())
		daysLeft = int(expiryDay.Sub(nowDay).Hours() / 24)
		
		// VERBOSE LOGGING: Pre-status determination
		s.logger.InfoContext(ctx, "[VERBOSE] Pre-status determination state",
			slog.String("trace_id", traceID),
			slog.String("manager_provided_status", status),
			slog.Int("calculated_days_left", daysLeft),
			slog.Time("current_time", now),
			slog.Time("expiry_date", info.ExpiryDate),
			slog.String("license_duration", info.Duration),
			slog.String("license_key_prefix", maskLicenseKey(info.LicenseKey)),
		)
		
		// Determine license status and message
		licenseStatus := s.determineLicenseStatus(status, daysLeft)
		message = s.generateStatusMessage(licenseStatus, daysLeft)
		
		// VERBOSE LOGGING: Post-status determination
		s.logger.InfoContext(ctx, "[VERBOSE] Post-status determination result",
			slog.String("trace_id", traceID),
			slog.String("input_status", status),
			slog.String("determined_status", licenseStatus),
			slog.String("generated_message", message),
			slog.Int("days_left", daysLeft),
			slog.Time("expiry_date", info.ExpiryDate),
		)
		
		
		// Create Iraqi Investor specific response for other statuses
		response := &LicenseStatusResponse{
			Status:        200,
			LicenseStatus: licenseStatus,
			Message:       message,
			DaysLeft:      daysLeft,
			LicenseInfo:   info,
			TraceID:       traceID,
			UserInfo:      s.buildUserInfo(info),
			Features:      s.buildFeaturesList(licenseStatus),
			Limitations:   s.buildLimitations(licenseStatus),
			RenewalInfo:   s.buildRenewalInfo(daysLeft, info.ExpiryDate),
			BrandingInfo:  s.buildBrandingInfo(),
			Timestamp:     time.Now(),
		}
		
		return response, nil
	}
	
	// No license info available
	return &LicenseStatusResponse{
		Type:          "/license/not-found",
		Title:         "License Not Found",
		Status:        200,
		Detail:        "No license information available",
		Instance:      fmt.Sprintf("/api/license/status#%s", traceID),
		LicenseStatus: "not_activated",
		Message:       "No license found. Please activate a license for Iraqi Investor access.",
		TraceID:       traceID,
		Features:      s.buildFeaturesList("not_activated"),
		Limitations:   s.buildLimitations("not_activated"),
		BrandingInfo:  s.buildBrandingInfo(),
		Timestamp:     time.Now(),
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
	err := s.activator.Activate(ctx, key)
	
	// Handle reactivation success scenario
	if err != nil && errors.Is(err, licenseErrors.ErrLicenseReactivated) {
		s.logger.InfoContext(ctx, "license reactivation succeeded",
			slog.String("trace_id", traceID),
			slog.String("operation", "reactivate"),
			slog.String("license_key", maskedKey),
			slog.Duration("latency", time.Since(start)),
			slog.String("result", "reactivated"),
		)
		
		// Invalidate cache to ensure fresh status
		if cacheErr := s.InvalidateCache(ctx); cacheErr != nil {
			s.logger.WarnContext(ctx, "failed to invalidate cache after reactivation",
				slog.String("trace_id", traceID),
				slog.String("cache_error", cacheErr.Error()),
			)
		}
		
		// Reactivation is a success case - return nil
		return nil
	}
	
	// Log other errors
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

// ValidateWithContext validates the current license with enhanced tracking
func (s *licenseService) ValidateWithContext(ctx context.Context) (bool, error) {
	start := time.Now()
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = middleware.GetReqID(ctx)
	}
	
	s.logger.DebugContext(ctx, "license validation started",
		slog.String("trace_id", traceID),
		slog.String("operation", "validate"))
	
	// Update metrics
	s.validationCount++
	
	// Create a channel for the result
	type result struct {
		valid bool
		err   error
	}
	
	resultCh := make(chan result, 1)
	
	// Run validation in goroutine to respect context
	go func() {
		valid, err := s.validator.Validate(ctx)
		resultCh <- result{valid, err}
	}()
	
	// Wait for result or context cancellation
	select {
	case <-ctx.Done():
		s.errorCount++
		s.logger.WarnContext(ctx, "license validation cancelled",
			slog.String("trace_id", traceID),
			slog.Duration("latency", time.Since(start)))
		return false, ctx.Err()
		
	case res := <-resultCh:
		// Update metrics
		duration := time.Since(start)
		s.totalResponseTime += duration
		s.lastValidation = time.Now()
		
		if res.err != nil {
			s.errorCount++
		} else if res.valid {
			s.successCount++
		}
		
		s.logger.DebugContext(ctx, "license validation completed",
			slog.String("trace_id", traceID),
			slog.String("operation", "validate"),
			slog.Duration("latency", duration),
			slog.Bool("valid", res.valid),
			slog.Bool("has_error", res.err != nil))
		
		return res.valid, res.err
	}
}

// Helper functions

// determineLicenseStatus determines the license status based on the manager status and days left
func (s *licenseService) determineLicenseStatus(managerStatus string, daysLeft int) string {
	// VERBOSE LOGGING: Track status determination flow
	ctx := context.Background()
	s.logger.InfoContext(ctx, "[VERBOSE] License status determination flow started",
		slog.String("input_manager_status", managerStatus),
		slog.Int("input_days_left", daysLeft),
		slog.String("component", "license_service.determineLicenseStatus"),
	)
	
	var finalStatus string
	var reason string
	
	// Normalize the status
	switch managerStatus {
	case "Expired":
		finalStatus = "expired"
		reason = "Manager reported status as 'Expired'"
	case "Critical":
		finalStatus = "critical"
		reason = "Manager reported status as 'Critical'"
	case "Warning":
		finalStatus = "warning"
		reason = "Manager reported status as 'Warning'"
	case "Active", "Activated", "Valid":
		// Further categorize based on days left
		if daysLeft <= 0 {
			finalStatus = "expired"
			reason = fmt.Sprintf("Manager status is '%s' but days_left=%d (<=0)", managerStatus, daysLeft)
		} else if daysLeft <= 7 {
			finalStatus = "critical"
			reason = fmt.Sprintf("Manager status is '%s' but days_left=%d (<=7)", managerStatus, daysLeft)
		} else if daysLeft <= 30 {
			finalStatus = "warning"
			reason = fmt.Sprintf("Manager status is '%s' but days_left=%d (<=30)", managerStatus, daysLeft)
		} else {
			finalStatus = "active"
			reason = fmt.Sprintf("Manager status is '%s' and days_left=%d (>30)", managerStatus, daysLeft)
		}
	default:
		finalStatus = "not_activated"
		reason = fmt.Sprintf("Unknown manager status: '%s'", managerStatus)
	}
	
	// VERBOSE LOGGING: Log the final determination
	s.logger.InfoContext(ctx, "[VERBOSE] License status determination completed",
		slog.String("input_manager_status", managerStatus),
		slog.Int("input_days_left", daysLeft),
		slog.String("final_status", finalStatus),
		slog.String("determination_reason", reason),
		slog.String("component", "license_service.determineLicenseStatus"),
	)
	
	return finalStatus
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

// GetDetailedStatus returns comprehensive license status information
func (s *licenseService) GetDetailedStatus(ctx context.Context) (*DetailedLicenseStatusResponse, error) {
	start := time.Now()
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = middleware.GetReqID(ctx)
	}
	
	s.logger.InfoContext(ctx, "detailed license status check started",
		slog.String("trace_id", traceID),
		slog.String("operation", "get_detailed_status"))
	
	// Get basic status first
	basicStatus, err := s.GetStatus(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get basic license status",
			slog.String("trace_id", traceID),
			slog.String("error", err.Error()))
		return nil, err
	}
	
	// Get additional details
	info, _, err := s.validator.Status(ctx)
	
	detailed := &DetailedLicenseStatusResponse{
		LicenseStatusResponse: *basicStatus,
		NetworkStatus:         s.determineNetworkStatus(err),
		ValidationCount:       s.validationCount,
	}
	
	if info != nil {
		if !info.IssuedDate.IsZero() {
			detailed.ActivationDate = &info.IssuedDate
		}
		if !info.LastChecked.IsZero() {
			detailed.LastValidation = &info.LastChecked
		}
	}
	
	// Add performance metrics
	if metrics, err := s.GetValidationMetrics(ctx); err == nil {
		detailed.PerformanceMetrics = metrics
	}
	
	// Add recommendations
	detailed.Recommendations = s.generateRecommendations(basicStatus.LicenseStatus, err)
	
	s.logger.InfoContext(ctx, "detailed license status check completed",
		slog.String("trace_id", traceID),
		slog.Duration("latency", time.Since(start)),
		slog.String("status", detailed.LicenseStatus))
	
	return detailed, nil
}

// CheckRenewalStatus provides license renewal information
func (s *licenseService) CheckRenewalStatus(ctx context.Context) (*RenewalStatusResponse, error) {
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = middleware.GetReqID(ctx)
	}
	
	info, _, err := s.validator.Status(ctx)
	if err != nil {
		return &RenewalStatusResponse{
			NeedsRenewal:   true,
			IsExpired:      true,
			RenewalUrgency: "critical",
			RenewalMessage: "Unable to check license status. Please contact support.",
			ContactInfo:    "Please contact your license provider for assistance.",
			TraceID:        traceID,
		}, nil
	}
	
	if info == nil {
		return &RenewalStatusResponse{
			NeedsRenewal:   true,
			IsExpired:      true,
			RenewalUrgency: "critical", 
			RenewalMessage: "No license found. Please activate a license.",
			ContactInfo:    "Contact your license provider to obtain a license.",
			TraceID:        traceID,
		}, nil
	}
	
	now := time.Now()
	daysLeft := int(info.ExpiryDate.Sub(now).Hours() / 24)
	
	response := &RenewalStatusResponse{
		DaysUntilExpiry: daysLeft,
		ExpiryDate:      info.ExpiryDate,
		TraceID:         traceID,
	}
	
	if daysLeft <= 0 {
		response.NeedsRenewal = true
		response.IsExpired = true
		response.RenewalUrgency = "critical"
		response.RenewalMessage = fmt.Sprintf("License expired %d days ago. Immediate renewal required.", -daysLeft)
	} else if daysLeft <= 7 {
		response.NeedsRenewal = true
		response.IsExpired = false
		response.RenewalUrgency = "critical"
		response.RenewalMessage = fmt.Sprintf("License expires in %d days! Urgent renewal needed.", daysLeft)
	} else if daysLeft <= 30 {
		response.NeedsRenewal = true
		response.IsExpired = false
		response.RenewalUrgency = "high"
		response.RenewalMessage = fmt.Sprintf("License expires in %d days. Please renew soon.", daysLeft)
	} else if daysLeft <= 90 {
		response.NeedsRenewal = false
		response.IsExpired = false
		response.RenewalUrgency = "medium"
		response.RenewalMessage = fmt.Sprintf("License expires in %d days. Consider renewal planning.", daysLeft)
	} else {
		response.NeedsRenewal = false
		response.IsExpired = false
		response.RenewalUrgency = "low"
		response.RenewalMessage = fmt.Sprintf("License is active with %d days remaining.", daysLeft)
	}
	
	response.ContactInfo = "Contact your license provider for renewal options."
	
	return response, nil
}

// TransferLicense transfers a license to current machine
func (s *licenseService) TransferLicense(ctx context.Context, key string, force bool) error {
	start := time.Now()
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = middleware.GetReqID(ctx)
	}
	
	maskedKey := maskLicenseKey(key)
	
	s.logger.InfoContext(ctx, "license transfer started",
		slog.String("trace_id", traceID),
		slog.String("operation", "transfer"),
		slog.String("license_key", maskedKey),
		slog.Bool("force", force))
	
	// Check if manager supports transfer operation
	if v, ok := s.validator.(*license.DefaultValidationService); ok {
		if transferManager, ok := v.Manager.(*license.Manager); ok {
			err := transferManager.TransferLicense(key, force)
			
			// Update metrics
			s.validationCount++
			s.totalResponseTime += time.Since(start)
			s.lastValidation = time.Now()
			
			if err != nil {
				s.errorCount++
				s.logger.ErrorContext(ctx, "license transfer failed",
					slog.String("trace_id", traceID),
					slog.String("license_key", maskedKey),
					slog.Duration("latency", time.Since(start)),
					slog.String("error", err.Error()))
				
				return s.mapTransferError(err)
			}
			
			s.successCount++
			s.logger.InfoContext(ctx, "license transfer succeeded",
				slog.String("trace_id", traceID),
				slog.String("license_key", maskedKey),
				slog.Duration("latency", time.Since(start)))
			
			return nil
		}
	}
	
	return licenseErrors.NewLicenseError("transfer not supported by current manager", nil)
}

// GetValidationMetrics returns performance and reliability metrics
func (s *licenseService) GetValidationMetrics(ctx context.Context) (*ValidationMetrics, error) {
	uptime := time.Since(s.startTime)
	
	var avgResponseTime time.Duration
	if s.validationCount > 0 {
		avgResponseTime = time.Duration(int64(s.totalResponseTime) / s.validationCount)
	}
	
	var cacheHitRate float64
	if s.validationCount > 0 {
		// Estimate cache hit rate based on response times
		// This is a simplified calculation
		cacheHitRate = float64(s.successCount) / float64(s.validationCount)
	}
	
	return &ValidationMetrics{
		TotalValidations:      s.validationCount,
		SuccessfulValidations: s.successCount,
		FailedValidations:     s.errorCount,
		AverageResponseTime:   avgResponseTime,
		LastValidationTime:    s.lastValidation,
		CacheHitRate:          cacheHitRate,
		NetworkErrors:         s.errorCount, // Simplified - in real implementation, distinguish network errors
		Uptime:                uptime,
	}, nil
}

// InvalidateCache invalidates any cached license validation results
func (s *licenseService) InvalidateCache(ctx context.Context) error {
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = middleware.GetReqID(ctx)
	}
	
	s.logger.InfoContext(ctx, "invalidating license cache",
		slog.String("trace_id", traceID),
		slog.String("operation", "invalidate_cache"))
	
	// If underlying manager supports cache invalidation, invoke it
	// Manager exposes cache invalidation via the LicenseCacheInvalidator interface internally; no-op here until a public hook exists.
	
	return nil
}

// Helper methods

// determineNetworkStatus checks network connectivity status
func (s *licenseService) determineNetworkStatus(err error) string {
	if err == nil {
		return "connected"
	}
	
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return "disconnected"
	}
	if strings.Contains(errStr, "timeout") {
		return "timeout"
	}
	
	return "unknown"
}

// maskMachineID masks machine ID for safe logging
func (s *licenseService) maskMachineID(machineID string) string {
	if len(machineID) <= 8 {
		return machineID
	}
	return machineID[:8] + "..."
}

// generateRecommendations provides actionable recommendations based on license status
func (s *licenseService) generateRecommendations(status string, err error) []string {
	var recommendations []string
	
	switch status {
	case "expired":
		recommendations = append(recommendations, 
			"Contact your license provider immediately to renew your license",
			"Backup your current configuration before license renewal",
			"Check if you qualify for a grace period extension")
			
	case "critical":
		recommendations = append(recommendations,
			"Plan for license renewal within the next week",
			"Contact your license provider to discuss renewal options",
			"Ensure uninterrupted access by renewing before expiration")
			
	case "warning":
		recommendations = append(recommendations,
			"Begin license renewal process to avoid interruption",
			"Review license terms for any changes in the new period",
			"Contact your license provider for renewal procedures")
			
	case "not_activated":
		recommendations = append(recommendations,
			"Activate your license to access all application features",
			"Ensure you have a valid license key from your provider",
			"Check your internet connection for license activation")
	}
	
	if err != nil {
		if s.determineNetworkStatus(err) == "disconnected" {
			recommendations = append(recommendations,
				"Check your internet connection",
				"Verify firewall settings allow license server access",
				"Contact your IT administrator if connectivity issues persist")
		}
	}
	
	return recommendations
}

// mapTransferError maps transfer errors to appropriate error types
func (s *licenseService) mapTransferError(err error) error {
	if err == nil {
		return nil
	}
	
	errStr := strings.ToLower(err.Error())
	
	if strings.Contains(errStr, "already activated") {
		return licenseErrors.ErrActivationFailed
	}
	if strings.Contains(errStr, "expired") {
		return licenseErrors.ErrLicenseExpired
	}
	if strings.Contains(errStr, "not found") {
		return licenseErrors.ErrInvalidLicenseKey
	}
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return licenseErrors.ErrNetworkError
	}
	
	return err
}

// Iraqi Investor specific helper methods

// buildUserInfo creates user information from license info
func (s *licenseService) buildUserInfo(info *license.LicenseInfo) *UserInfo {
	if info == nil {
		return nil
	}
	
	return &UserInfo{
		Email:       info.UserEmail,
		Company:     "Iraqi Investor",
		Tier:        s.determineLicenseTier(info),
		ActivatedAt: info.IssuedDate,
		LastSeen:    info.LastChecked,
	}
}

// buildFeaturesList returns available features based on license status
func (s *licenseService) buildFeaturesList(licenseStatus string) []string {
	baseFeatures := []string{
		"Daily Reports Access",
		"Market Data Viewing",
		"Basic Charts",
	}
	
	if licenseStatus == "active" || licenseStatus == "warning" {
		return append(baseFeatures, []string{
			"Advanced Analytics",
			"Historical Data Export",
			"Custom Report Generation",
			"Real-time WebSocket Updates",
			"Iraqi Stock Exchange Integration",
			"Market Movers Analysis",
			"Ticker Performance Tracking",
		}...)
	}
	
	return baseFeatures
}

// buildLimitations returns current license limitations
func (s *licenseService) buildLimitations(licenseStatus string) map[string]interface{} {
	limitations := make(map[string]interface{})
	
	switch licenseStatus {
	case "expired":
		limitations["data_access"] = "Read-only mode"
		limitations["export_limit"] = 0
		limitations["real_time_updates"] = false
		limitations["message"] = "License expired - functionality limited"
		
	case "critical":
		limitations["export_limit"] = 5
		limitations["warning"] = "License expires soon - renew to maintain full access"
		
	case "warning":
		limitations["reminder"] = "License renewal recommended"
		
	case "not_activated":
		limitations["data_access"] = "Demo mode only"
		limitations["export_limit"] = 0
		limitations["real_time_updates"] = false
		limitations["message"] = "Please activate license for full functionality"
		
	default:
		// Active license - no limitations
		limitations["export_limit"] = -1 // unlimited
		limitations["real_time_updates"] = true
	}
	
	return limitations
}

// buildRenewalInfo creates renewal information
func (s *licenseService) buildRenewalInfo(daysLeft int, expiryDate time.Time) *RenewalInfo {
	renewal := &RenewalInfo{
		DaysUntilExpiry: daysLeft,
		ExpiryDate:     expiryDate,
		ContactInfo:    "Contact your Iraqi Investor license provider",
		RenewalURL:     "https://iraqiinvestor.gov.iq/license/renew",
	}
	
	if daysLeft <= 0 {
		renewal.NeedsRenewal = true
		renewal.IsExpired = true
		renewal.RenewalUrgency = "critical"
		renewal.RenewalMessage = "License has expired. Immediate renewal required to restore full functionality."
	} else if daysLeft <= 7 {
		renewal.NeedsRenewal = true
		renewal.IsExpired = false
		renewal.RenewalUrgency = "critical"
		renewal.RenewalMessage = fmt.Sprintf("License expires in %d days! Urgent renewal needed.", daysLeft)
	} else if daysLeft <= 30 {
		renewal.NeedsRenewal = true
		renewal.IsExpired = false
		renewal.RenewalUrgency = "high"
		renewal.RenewalMessage = fmt.Sprintf("License expires in %d days. Please renew soon.", daysLeft)
	} else if daysLeft <= 90 {
		renewal.NeedsRenewal = false
		renewal.IsExpired = false
		renewal.RenewalUrgency = "medium"
		renewal.RenewalMessage = fmt.Sprintf("License expires in %d days. Consider renewal planning.", daysLeft)
	} else {
		renewal.NeedsRenewal = false
		renewal.IsExpired = false
		renewal.RenewalUrgency = "low"
		renewal.RenewalMessage = fmt.Sprintf("License is active with %d days remaining.", daysLeft)
	}
	
	return renewal
}

// buildBrandingInfo creates Iraqi Investor branding information
func (s *licenseService) buildBrandingInfo() *BrandingInfo {
	return &BrandingInfo{
		ApplicationName: "ISX Daily Reports Scrapper",
		Version:         "2.0",
		BrandName:       "Iraqi Investor",
		LogoURL:         "/static/images/iraqi-investor-logo.svg",
		WebsiteURL:      "https://iraqiinvestor.gov.iq",
		SupportEmail:    "support@iraqiinvestor.gov.iq",
	}
}

// determineLicenseTier determines license tier from license info
func (s *licenseService) determineLicenseTier(info *license.LicenseInfo) string {
	// This could be enhanced to read from license info or external source
	// For now, default to professional
	if info.Duration == "Lifetime" {
		return "enterprise"
	} else if info.Duration == "Yearly" {
		return "professional"
	}
	return "basic"
}

// GetDebugInfo retrieves diagnostic information for license troubleshooting
func (s *licenseService) GetDebugInfo(ctx context.Context) (*LicenseDebugInfo, error) {
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = middleware.GetReqID(ctx)
	}
	
	s.logger.InfoContext(ctx, "retrieving license debug information",
		slog.String("trace_id", traceID),
		slog.String("operation", "get_debug_info"))
	
	// Get basic paths and environment info
	debugInfo := &LicenseDebugInfo{
		TraceID:   traceID,
		Timestamp: time.Now(),
		Environment: make(map[string]string),
	}
	
	// Get file paths from license manager
	if s.validator != nil {
		mgr := s.validator.(*license.DefaultValidationService).Manager
		licensePath := mgr.GetLicensePath()
		debugInfo.FilePath = licensePath
		
		// Check file existence and permissions
		fileInfo, err := os.Stat(licensePath)
		if err == nil {
			debugInfo.FileExists = true
			debugInfo.FileSize = fileInfo.Size()
			debugInfo.FilePermissions = fileInfo.Mode().String()
			modTime := fileInfo.ModTime()
			debugInfo.FileModTime = &modTime
			
			// Check if file is readable
			if file, err := os.Open(licensePath); err == nil {
				debugInfo.IsReadable = true
				file.Close()
			} else {
				debugInfo.IsReadable = false
				debugInfo.LastError = fmt.Sprintf("Cannot read file: %v", err)
			}
		} else {
			debugInfo.FileExists = false
			if !os.IsNotExist(err) {
				debugInfo.LastError = fmt.Sprintf("File stat error: %v", err)
			}
		}
	}
	
	// Get working directory
	if wd, err := os.Getwd(); err == nil {
		debugInfo.WorkingDir = wd
	} else {
		debugInfo.WorkingDir = fmt.Sprintf("error: %v", err)
	}
	
	// Get executable path
	if execPath, err := os.Executable(); err == nil {
		debugInfo.ExecPath = execPath
	} else {
		debugInfo.ExecPath = fmt.Sprintf("error: %v", err)
	}
	
	// Get config path if available
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		debugInfo.ConfigPath = configPath
	} else {
		debugInfo.ConfigPath = "not set"
	}
	
	// Get current license status
	if statusResp, err := s.GetStatus(ctx); err == nil {
		debugInfo.LicenseStatus = statusResp.LicenseStatus
	} else {
		debugInfo.LicenseStatus = "error"
		debugInfo.LastError = fmt.Sprintf("Status check error: %v", err)
	}
	
	// Add relevant environment variables
	debugInfo.Environment["ISX_LICENSE_PATH"] = os.Getenv("ISX_LICENSE_PATH")
	debugInfo.Environment["LICENSE_PATH"] = os.Getenv("LICENSE_PATH")
	debugInfo.Environment["CONFIG_PATH"] = os.Getenv("CONFIG_PATH")
	debugInfo.Environment["WORKING_DIR"] = debugInfo.WorkingDir
	debugInfo.Environment["EXEC_PATH"] = debugInfo.ExecPath
	
	s.logger.InfoContext(ctx, "license debug info retrieved",
		slog.String("trace_id", traceID),
		slog.String("file_path", debugInfo.FilePath),
		slog.Bool("file_exists", debugInfo.FileExists),
		slog.Bool("is_readable", debugInfo.IsReadable),
		slog.String("license_status", debugInfo.LicenseStatus))
	
	return debugInfo, nil
}

// CheckExistingLicense checks if there's an existing license and returns its details
func (s *licenseService) CheckExistingLicense() (*license.ExistingLicenseInfo, error) {
	// Delegate to underlying manager via validator facade for now
	return s.validator.(*license.DefaultValidationService).Manager.CheckExistingLicense()
}

// GetLicenseDetails returns comprehensive license information including stacking history
func (s *licenseService) GetLicenseDetails() (interface{}, error) {
	ctx := context.Background()
	
	// Get basic license info
	licenseInfo, err := s.validator.(*license.DefaultValidationService).Manager.GetLicenseInfo()
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get license info",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("no license found")
	}
	
	// Parse stacking history from license key (if stacked)
	var stackedKeys []string
	if strings.Contains(licenseInfo.LicenseKey, "+") {
		stackedKeys = strings.Split(licenseInfo.LicenseKey, "+")
	}
	
	// Parse activation history from activation ID (if stacked)
	var activationHistory []string
	if strings.Contains(licenseInfo.ActivationID, "+") {
		activationHistory = strings.Split(licenseInfo.ActivationID, "+")
	}
	
	// Calculate total days from stacking
	daysRemaining := 0
	if time.Now().Before(licenseInfo.ExpiryDate) {
		daysRemaining = int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
	}
	
	// Build detailed response
	details := map[string]interface{}{
		"license_key":        MaskLicenseKey(licenseInfo.LicenseKey),
		"user_email":         licenseInfo.UserEmail,
		"expiry_date":        licenseInfo.ExpiryDate,
		"days_remaining":     daysRemaining,
		"status":             licenseInfo.Status,
		"duration":           licenseInfo.Duration,
		"issued_date":        licenseInfo.IssuedDate,
		"last_checked":       licenseInfo.LastChecked,
		"activation_id":      licenseInfo.ActivationID,
		"is_stacked":         len(stackedKeys) > 1,
		"stacked_count":      len(stackedKeys),
		"stacked_keys":       stackedKeys,
		"activation_history": activationHistory,
		"device_fingerprint": licenseInfo.DeviceFingerprint[:min(16, len(licenseInfo.DeviceFingerprint))],
	}
	
	return details, nil
}

// GetActivationHistory returns activation history from audit logs
func (s *licenseService) GetActivationHistory() ([]interface{}, error) {
	ctx := context.Background()
	
	// Read audit log file
	auditFile := filepath.Join("logs", "license_audit.json")
	
	// Check if audit file exists
	if _, err := os.Stat(auditFile); os.IsNotExist(err) {
		s.logger.InfoContext(ctx, "no audit history found",
			slog.String("file", auditFile),
		)
		return []interface{}{}, nil
	}
	
	// Read audit entries
	data, err := os.ReadFile(auditFile)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to read audit file",
			slog.String("error", err.Error()),
			slog.String("file", auditFile),
		)
		return nil, fmt.Errorf("failed to read audit history: %w", err)
	}
	
	// Parse JSON lines
	var history []interface{}
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			s.logger.WarnContext(ctx, "failed to parse audit entry",
				slog.String("error", err.Error()),
				slog.String("line", line),
			)
			continue
		}
		
		history = append(history, entry)
	}
	
	// Sort by timestamp (newest first)
	// Note: In production, you might want to use a proper sorting algorithm
	
	s.logger.InfoContext(ctx, "retrieved activation history",
		slog.Int("count", len(history)),
	)
	
	return history, nil
}

// BackupCurrentLicense creates a backup of the current license
func (s *licenseService) BackupCurrentLicense() (string, error) {
	ctx := context.Background()
	
	// Get current license
	licenseInfo, err := s.validator.(*license.DefaultValidationService).Manager.GetLicenseInfo()
	if err != nil {
		s.logger.ErrorContext(ctx, "no license to backup",
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("no license to backup")
	}
	
	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("license_backup_%s.json", timestamp)
	backupPath := filepath.Join("logs", backupFile)
	
	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Marshal license data
	data, err := json.MarshalIndent(licenseInfo, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal license data: %w", err)
	}
	
	// Write backup file
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}
	
	s.logger.InfoContext(ctx, "license backup created",
		slog.String("backup_path", backupPath),
		slog.String("license_key", MaskLicenseKey(licenseInfo.LicenseKey)),
		slog.Int("size_bytes", len(data)),
	)
	
	return backupPath, nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaskLicenseKey is a wrapper to call the license package's MaskLicenseKey function
func MaskLicenseKey(key string) string {
	return license.MaskLicenseKey(key)
}
