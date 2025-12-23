package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/isxcli/isxcli/internal/infrastructure"
)

// LicenseError represents a structured license error with RFC 7807 compliance
type LicenseError struct {
	// RFC 7807 standard fields
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`

	// License-specific fields
	ErrorCode        string                    `json:"error_code"`
	LicenseKey       string                    `json:"license_key,omitempty"`
	TraceID          string                    `json:"trace_id"`
	Timestamp        time.Time                 `json:"timestamp"`
	Service          string                    `json:"service"`
	Operation        string                    `json:"operation"`

	// Additional context
	Extensions       map[string]interface{}    `json:"-"`

	// Original error for wrapping
	Err              error                     `json:"-"`
}

// Error implements the error interface
func (le *LicenseError) Error() string {
	if le.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", le.ErrorCode, le.Title, le.Detail)
	}
	return fmt.Sprintf("[%s] %s", le.ErrorCode, le.Title)
}

// Unwrap returns the underlying error for error wrapping
func (le *LicenseError) Unwrap() error {
	return le.Err
}

// WithContext adds context information to the license error
func (le *LicenseError) WithContext(ctx context.Context) *LicenseError {
	if traceID := GetTraceID(ctx); traceID != "" {
		le.TraceID = traceID
	}
	return le
}

// WithLicenseKey adds license key context
func (le *LicenseError) WithLicenseKey(key string) *LicenseError {
	le.LicenseKey = maskLicenseKey(key)
	return le
}

// WithExtension adds an extension field
func (le *LicenseError) WithExtension(key string, value interface{}) *LicenseError {
	if le.Extensions == nil {
		le.Extensions = make(map[string]interface{})
	}
	le.Extensions[key] = value
	return le
}

// MarshalJSON custom marshaler to include extensions
func (le *LicenseError) MarshalJSON() ([]byte, error) {
	type Alias LicenseError
	data := make(map[string]interface{})

	// Add standard fields
	data["type"] = le.Type
	data["title"] = le.Title
	data["status"] = le.Status
	data["error_code"] = le.ErrorCode
	data["trace_id"] = le.TraceID
	data["timestamp"] = le.Timestamp.UTC().Format(time.RFC3339)
	data["service"] = le.Service
	data["operation"] = le.Operation

	if le.Detail != "" {
		data["detail"] = le.Detail
	}
	if le.Instance != "" {
		data["instance"] = le.Instance
	}
	if le.LicenseKey != "" {
		data["license_key"] = le.LicenseKey
	}

	// Add extensions
	for k, v := range le.Extensions {
		data[k] = v
	}

	return json.Marshal(data)
}

// ToProblemDetails converts LicenseError to ProblemDetails for HTTP responses
func (le *LicenseError) ToProblemDetails() *ProblemDetails {
	pd := &ProblemDetails{
		Type:       le.Type,
		Title:      le.Title,
		Status:     le.Status,
		Detail:     le.Detail,
		Instance:   le.Instance,
		Extensions: make(map[string]interface{}),
	}

	// Add license-specific fields as extensions
	pd.Extensions["error_code"] = le.ErrorCode
	pd.Extensions["trace_id"] = le.TraceID
	pd.Extensions["timestamp"] = le.Timestamp.UTC().Format(time.RFC3339)
	pd.Extensions["service"] = le.Service
	pd.Extensions["operation"] = le.Operation

	if le.LicenseKey != "" {
		pd.Extensions["license_key"] = le.LicenseKey
	}

	// Copy extensions
	for k, v := range le.Extensions {
		pd.Extensions[k] = v
	}

	return pd
}

// License-specific errors (using errors package for sentinel errors)
var (
	ErrLicenseExpired          = errors.New("license expired")
	ErrLicenseNotActivated     = errors.New("license not activated")
	ErrInvalidLicenseKey       = errors.New("invalid license key")
	ErrInvalidLicenseFormat    = errors.New("invalid license key format")
	ErrRateLimited            = errors.New("rate limited")
	ErrNetworkError           = errors.New("network error")
	ErrActivationFailed       = errors.New("activation failed")
	ErrLicenseValidationFailed = errors.New("license validation failed")
	ErrLicenseAlreadyActivated = errors.New("license already activated")
	
	// Reactivation-specific errors
	ErrLicenseReactivated           = errors.New("license reactivated")
	ErrReactivationLimitExceeded    = errors.New("reactivation limit exceeded")
	ErrAlreadyActivatedOnDevice     = errors.New("already activated on different device")
)

// Factory functions for common license errors

// NewLicenseExpiredError creates a license expired error with context
func NewLicenseExpiredError(ctx context.Context, expiryDate time.Time, licenseKey string) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/license-expired",
		Title:     "License Expired",
		Status:    http.StatusForbidden,
		Detail:    fmt.Sprintf("License expired on %s. Please renew to continue.", expiryDate.Format("2006-01-02")),
		ErrorCode: "LICENSE_EXPIRED",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_license",
		Extensions: map[string]interface{}{
			"expiry_date":       expiryDate.Format(time.RFC3339),
			"days_expired":      int(time.Since(expiryDate).Hours() / 24),
			"can_renew":         true,
			"renewal_url":       "https://isxpulse.com/renew",
		},
		Err: ErrLicenseExpired,
	}
	return le.WithContext(ctx).WithLicenseKey(licenseKey)
}

// NewLicenseNotActivatedError creates a license not activated error
func NewLicenseNotActivatedError(ctx context.Context) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/license-not-activated",
		Title:     "License Not Activated",
		Status:    http.StatusPreconditionRequired,
		Detail:    "No license has been activated. Please activate a license to continue.",
		ErrorCode: "LICENSE_NOT_ACTIVATED",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_license",
		Extensions: map[string]interface{}{
			"activation_required": true,
			"activation_url":      "/api/license/activate",
		},
		Err: ErrLicenseNotActivated,
	}
	return le.WithContext(ctx)
}

// NewInvalidLicenseKeyError creates an invalid license key error
func NewInvalidLicenseKeyError(ctx context.Context, licenseKey string, reason string) *LicenseError {
	detail := "The provided license key is invalid."
	if reason != "" {
		detail = fmt.Sprintf("The provided license key is invalid: %s", reason)
	}

	le := &LicenseError{
		Type:      "/errors/invalid-license-key",
		Title:     "Invalid License Key",
		Status:    http.StatusBadRequest,
		Detail:    detail,
		ErrorCode: "INVALID_LICENSE_KEY",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_license",
		Extensions: map[string]interface{}{
			"validation_reason": reason,
			"expected_format":   "ISX-XXXX-XXXX-XXXX-XXXX",
		},
		Err: ErrInvalidLicenseKey,
	}
	return le.WithContext(ctx).WithLicenseKey(licenseKey)
}

// NewInvalidLicenseFormatError creates an invalid license format error
func NewInvalidLicenseFormatError(ctx context.Context, licenseKey string) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/invalid-license-format",
		Title:     "Invalid License Format",
		Status:    http.StatusBadRequest,
		Detail:    "License key must be in format: ISX-XXXX-XXXX-XXXX-XXXX",
		ErrorCode: "INVALID_LICENSE_FORMAT",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_format",
		Extensions: map[string]interface{}{
			"expected_format":  "ISX-XXXX-XXXX-XXXX-XXXX",
			"provided_length":  len(licenseKey),
			"valid_pattern":    `^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`,
		},
		Err: ErrInvalidLicenseFormat,
	}
	return le.WithContext(ctx).WithLicenseKey(licenseKey)
}

// NewActivationFailedError creates an activation failed error
func NewActivationFailedError(ctx context.Context, licenseKey string, reason string, err error) *LicenseError {
	detail := "Unable to activate the license. Please verify the key and try again."
	if reason != "" {
		detail = fmt.Sprintf("License activation failed: %s", reason)
	}

	le := &LicenseError{
		Type:      "/errors/activation-failed",
		Title:     "License Activation Failed",
		Status:    http.StatusUnprocessableEntity,
		Detail:    detail,
		ErrorCode: "ACTIVATION_FAILED",
		Timestamp: time.Now(),
		Service:   "license-activator",
		Operation: "activate_license",
		Extensions: map[string]interface{}{
			"failure_reason": reason,
			"retry_allowed":  true,
		},
		Err: err,
	}
	return le.WithContext(ctx).WithLicenseKey(licenseKey)
}

// NewDeviceMismatchError creates a device mismatch error
func NewDeviceMismatchError(ctx context.Context, currentDevice, registeredDevice string, similarity float64) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/device-mismatch",
		Title:     "Device Mismatch",
		Status:    http.StatusConflict,
		Detail:    fmt.Sprintf("This license is registered to a different device. Device similarity: %.1f%%", similarity*100),
		ErrorCode: "DEVICE_MISMATCH",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_device",
		Extensions: map[string]interface{}{
			"similarity_score":     similarity,
			"registered_device":    registeredDevice,
			"current_device":       currentDevice,
			"can_reactivate":       similarity > 0.7, // 70% threshold
			"reactivation_url":     "/api/license/reactivate",
		},
		Err: ErrAlreadyActivatedOnDevice,
	}
	return le.WithContext(ctx)
}

// NewRateLimitedError creates a rate limited error
func NewRateLimitedError(ctx context.Context, retryAfter int) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/rate-limited",
		Title:     "Too Many Requests",
		Status:    http.StatusTooManyRequests,
		Detail:    "Too many activation attempts. Please try again later.",
		ErrorCode: "RATE_LIMITED",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "rate_limit_check",
		Extensions: map[string]interface{}{
			"retry_after_seconds": retryAfter,
			"retry_after_time":    time.Now().Add(time.Duration(retryAfter) * time.Second).Format(time.RFC3339),
			"limit_window":        "15 minutes",
		},
		Err: ErrRateLimited,
	}
	return le.WithContext(ctx)
}

// NewLicenseValidationFailedError creates a license validation failed error
func NewLicenseValidationFailed(ctx context.Context, licenseKey, reason string, err error) *LicenseError {
	detail := "License validation failed."
	if reason != "" {
		detail = fmt.Sprintf("License validation failed: %s", reason)
	}

	le := &LicenseError{
		Type:      "/errors/license-validation-failed",
		Title:     "License Validation Failed",
		Status:    http.StatusInternalServerError,
		Detail:    detail,
		ErrorCode: "VALIDATION_FAILED",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_license",
		Extensions: map[string]interface{}{
			"validation_reason": reason,
			"can_retry":         true,
		},
		Err: err,
	}
	return le.WithContext(ctx).WithLicenseKey(licenseKey)
}

// NewLicenseNetworkError creates a network error for license operations
func NewLicenseNetworkError(ctx context.Context, err error, endpoint string) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/network-error",
		Title:     "Network Error",
		Status:    http.StatusServiceUnavailable,
		Detail:    "Unable to connect to license server. Please check your connection.",
		ErrorCode: "NETWORK_ERROR",
		Timestamp: time.Now(),
		Service:   "license-client",
		Operation: "connect_license_server",
		Extensions: map[string]interface{}{
			"endpoint":         endpoint,
			"retry_allowed":    true,
			"max_retries":      3,
			"retry_backoff":    "exponential",
		},
		Err: err,
	}
	return le.WithContext(ctx)
}

// Structured logging utilities for license errors

// LogLicenseError logs a license error with structured information
func LogLicenseError(ctx context.Context, err error, operation string, licenseKey string) {
	logger := slog.With(
		"service", "license-validator",
		"operation", operation,
		"trace_id", GetTraceID(ctx),
	)

	if licenseKey != "" {
		logger = logger.With("license_key", maskLicenseKey(licenseKey))
	}

	// Check if it's a LicenseError
	var licenseErr *LicenseError
	if errors.As(err, &licenseErr) {
		logger.ErrorContext(ctx, "License error occurred",
			"error_code", licenseErr.ErrorCode,
			"error_type", licenseErr.Type,
			"error_title", licenseErr.Title,
			"error_detail", licenseErr.Detail,
			"http_status", licenseErr.Status,
			"timestamp", licenseErr.Timestamp,
		)

		// Log extensions if any
		if len(licenseErr.Extensions) > 0 {
			logger.ErrorContext(ctx, "License error context",
				"extensions", licenseErr.Extensions,
			)
		}

		// Log original error if wrapped
		if licenseErr.Err != nil {
			logger.ErrorContext(ctx, "Original license error",
				"original_error", licenseErr.Err.Error(),
			)
		}
	} else {
		// Log generic error
		logger.ErrorContext(ctx, "License operation failed",
			"error", err.Error(),
		)
	}
}

// LogLicenseOperation logs a license operation with success/failure
func LogLicenseOperation(ctx context.Context, operation string, licenseKey string, success bool, metadata map[string]interface{}) {
	logger := slog.With(
		"service", "license-validator",
		"operation", operation,
		"trace_id", GetTraceID(ctx),
		"success", success,
		"timestamp", time.Now(),
	)

	if licenseKey != "" {
		logger = logger.With("license_key", maskLicenseKey(licenseKey))
	}

	// Add metadata
	for k, v := range metadata {
		logger = logger.With(k, v)
	}

	if success {
		logger.InfoContext(ctx, "License operation completed successfully")
	} else {
		logger.WarnContext(ctx, "License operation failed")
	}
}

// LogLicenseValidation logs detailed license validation information
func LogLicenseValidation(ctx context.Context, licenseKey string, validationResults map[string]bool) {
	logger := slog.With(
		"service", "license-validator",
		"operation", "validate_license",
		"trace_id", GetTraceID(ctx),
		"timestamp", time.Now(),
	)

	if licenseKey != "" {
		logger = logger.With("license_key", maskLicenseKey(licenseKey))
	}

	// Count passed/failed validations
	passed := 0
	failed := 0
	for check, result := range validationResults {
		if result {
			passed++
		} else {
			failed++
			logger.WarnContext(ctx, "License validation check failed",
				"check", check,
				"result", result,
			)
		}
	}

	logger.InfoContext(ctx, "License validation completed",
		"total_checks", len(validationResults),
		"passed_checks", passed,
		"failed_checks", failed,
		"validation_passed", failed == 0,
	)
}

// Utility functions

// maskLicenseKey masks a license key for logging
func maskLicenseKey(key string) string {
	if len(key) < 10 {
		return "****"
	}

	// Show first 4 and last 4 characters
	if len(key) <= 20 {
		return key[:4] + "****" + key[len(key)-4:]
	}

	// For longer keys, show first 8 and last 8
	return key[:8] + "****" + key[len(key)-8:]
}

// GetTraceID extracts trace ID from context using infrastructure package
func GetTraceID(ctx context.Context) string {
	return infrastructure.GetTraceID(ctx)
}

// LicenseActivationDetails provides additional context for license errors
type LicenseActivationDetails struct {
	ActivationDate   *time.Time `json:"activation_date,omitempty"`
	ExpiryDate       *time.Time `json:"expiry_date,omitempty"`
	DeviceInfo       string     `json:"device_info,omitempty"`
	CurrentStatus    string     `json:"current_status,omitempty"`
	DaysRemaining    int        `json:"days_remaining,omitempty"`
	SupportEmail     string     `json:"support_email,omitempty"`
	CanRecover       bool       `json:"can_recover,omitempty"`
	RecoveryDeadline *time.Time `json:"recovery_deadline,omitempty"`
	
	// Reactivation-specific fields
	ReactivationCount      int     `json:"reactivation_count,omitempty"`
	MaxReactivations       int     `json:"max_reactivations,omitempty"`
	SimilarityScore        float64 `json:"similarity_score,omitempty"`
	PreviousDeviceInfo     string  `json:"previous_device_info,omitempty"`
	ReactivationTimestamp  *time.Time `json:"reactivation_timestamp,omitempty"`
}

// ProblemDetails implements RFC 7807 Problem Details for HTTP APIs
type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	
	// Additional fields for extensibility
	Extensions map[string]interface{} `json:"-"`
}

// Render implements the render.Renderer interface
func (pd *ProblemDetails) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, pd.Status)
	return nil
}

// MarshalJSON custom marshaler to include extensions
func (pd *ProblemDetails) MarshalJSON() ([]byte, error) {
	type Alias ProblemDetails
	data := make(map[string]interface{})
	
	// Add standard fields
	data["type"] = pd.Type
	data["title"] = pd.Title
	data["status"] = pd.Status
	
	if pd.Detail != "" {
		data["detail"] = pd.Detail
	}
	if pd.Instance != "" {
		data["instance"] = pd.Instance
	}
	
	// Add extensions
	for k, v := range pd.Extensions {
		data[k] = v
	}
	
	// Use standard JSON marshaling
	return json.Marshal(data)
}

// UnmarshalJSON implements custom JSON unmarshaling for ProblemDetails.
// It extracts standard RFC 7807 fields and puts all other fields into Extensions.
func (pd *ProblemDetails) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to get all fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract standard fields
	if v, ok := raw["type"].(string); ok {
		pd.Type = v
	}
	if v, ok := raw["title"].(string); ok {
		pd.Title = v
	}
	if v, ok := raw["status"].(float64); ok {
		pd.Status = int(v)
	}
	if v, ok := raw["detail"].(string); ok {
		pd.Detail = v
	}
	if v, ok := raw["instance"].(string); ok {
		pd.Instance = v
	}

	// Initialize Extensions map if nil
	if pd.Extensions == nil {
		pd.Extensions = make(map[string]interface{})
	}

	// Add all non-standard fields to Extensions
	standardFields := map[string]bool{
		"type":     true,
		"title":    true,
		"status":   true,
		"detail":   true,
		"instance": true,
	}

	for k, v := range raw {
		if !standardFields[k] {
			pd.Extensions[k] = v
		}
	}

	return nil
}

// NewProblemDetails creates a new RFC 7807 compliant error
func NewProblemDetails(status int, problemType, title, detail, instance string) *ProblemDetails {
	return &ProblemDetails{
		Type:     problemType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Extensions: make(map[string]interface{}),
	}
}

// WithExtension adds an extension field to the problem details
func (pd *ProblemDetails) WithExtension(key string, value interface{}) *ProblemDetails {
	pd.Extensions[key] = value
	return pd
}

// NewLicenseAlreadyActivatedError creates an enhanced error for already activated licenses
func NewLicenseAlreadyActivatedError(details *LicenseActivationDetails, traceID string) *ProblemDetails {
	problem := NewProblemDetails(
		http.StatusConflict,
		"/errors/license-already-activated",
		"License Already Activated",
		"This license has already been activated on another device. To transfer it to this device, please contact support.",
		fmt.Sprintf("/api/license/activate#%s", traceID),
	)
	
	problem.WithExtension("error_type", "already_activated").
		WithExtension("trace_id", traceID).
		WithExtension("support_email", "support@isxpulse.com").
		WithExtension("transfer_info", "Contact support with your license key and proof of purchase to transfer this license.")
	
	if details != nil {
		if details.ActivationDate != nil {
			problem.WithExtension("original_activation_date", details.ActivationDate.Format("2006-01-02T15:04:05Z"))
		}
		if details.ExpiryDate != nil {
			problem.WithExtension("expiry_date", details.ExpiryDate.Format("2006-01-02T15:04:05Z"))
		}
		if details.DeviceInfo != "" {
			problem.WithExtension("registered_device", details.DeviceInfo)
		}
		if details.CurrentStatus != "" {
			problem.WithExtension("current_status", details.CurrentStatus)
		}
		problem.WithExtension("can_recover", details.CanRecover)
		if details.RecoveryDeadline != nil {
			problem.WithExtension("recovery_deadline", details.RecoveryDeadline.Format("2006-01-02T15:04:05Z"))
		}
	}
	
	return problem
}

// NewLicenseReactivatedResponse creates a success response for license reactivation
func NewLicenseReactivatedResponse(details *LicenseActivationDetails, traceID string) *ProblemDetails {
	problem := NewProblemDetails(
		http.StatusOK,
		"/success/license-reactivated",
		"License Successfully Reactivated",
		"Your license has been successfully reactivated on this device.",
		fmt.Sprintf("/api/license/activate#%s", traceID),
	)
	
	problem.WithExtension("success_type", "reactivated").
		WithExtension("trace_id", traceID).
		WithExtension("message", "License has been reactivated for this device")
	
	if details != nil {
		if details.ReactivationCount > 0 {
			problem.WithExtension("reactivation_count", details.ReactivationCount)
		}
		if details.MaxReactivations > 0 {
			problem.WithExtension("max_reactivations", details.MaxReactivations)
			problem.WithExtension("remaining_reactivations", details.MaxReactivations-details.ReactivationCount)
		}
		if details.SimilarityScore > 0 {
			problem.WithExtension("device_similarity_score", details.SimilarityScore)
		}
		if details.ReactivationTimestamp != nil {
			problem.WithExtension("reactivation_timestamp", details.ReactivationTimestamp.Format("2006-01-02T15:04:05Z"))
		}
		if details.ExpiryDate != nil {
			problem.WithExtension("expiry_date", details.ExpiryDate.Format("2006-01-02T15:04:05Z"))
		}
	}
	
	return problem
}

// NewReactivationLimitExceededError creates an error for when reactivation limit is reached
func NewReactivationLimitExceededError(details *LicenseActivationDetails, traceID string) *ProblemDetails {
	problem := NewProblemDetails(
		http.StatusConflict,
		"/errors/reactivation-limit-exceeded",
		"Reactivation Limit Exceeded",
		"This license has reached its maximum number of device reactivations. Please contact support for assistance.",
		fmt.Sprintf("/api/license/activate#%s", traceID),
	)
	
	problem.WithExtension("error_type", "reactivation_limit_exceeded").
		WithExtension("trace_id", traceID).
		WithExtension("support_email", "support@isxpulse.com").
		WithExtension("support_info", "Contact support with your license key to increase reactivation limit or transfer to a new device.")
	
	if details != nil {
		if details.ReactivationCount > 0 {
			problem.WithExtension("current_reactivations", details.ReactivationCount)
		}
		if details.MaxReactivations > 0 {
			problem.WithExtension("max_reactivations", details.MaxReactivations)
		}
		if details.ActivationDate != nil {
			problem.WithExtension("original_activation_date", details.ActivationDate.Format("2006-01-02T15:04:05Z"))
		}
		if details.ExpiryDate != nil {
			problem.WithExtension("expiry_date", details.ExpiryDate.Format("2006-01-02T15:04:05Z"))
		}
	}
	
	return problem
}

// NewAlreadyActivatedOnDeviceError creates an error for when license is already active on a different device
func NewAlreadyActivatedOnDeviceError(details *LicenseActivationDetails, traceID string) *ProblemDetails {
	problem := NewProblemDetails(
		http.StatusConflict,
		"/errors/already-activated-different-device",
		"License Already Active on Different Device",
		"This license is currently active on a different device. You may be able to reactivate it depending on device similarity and reactivation limits.",
		fmt.Sprintf("/api/license/activate#%s", traceID),
	)
	
	problem.WithExtension("error_type", "already_activated_different_device").
		WithExtension("trace_id", traceID).
		WithExtension("support_email", "support@isxpulse.com")
	
	if details != nil {
		if details.PreviousDeviceInfo != "" {
			problem.WithExtension("previous_device", details.PreviousDeviceInfo)
		}
		if details.SimilarityScore > 0 {
			problem.WithExtension("device_similarity_score", details.SimilarityScore)
		}
		if details.ReactivationCount > 0 && details.MaxReactivations > 0 {
			remaining := details.MaxReactivations - details.ReactivationCount
			problem.WithExtension("reactivations_remaining", remaining)
			if remaining > 0 {
				problem.WithExtension("can_reactivate", true)
				problem.WithExtension("reactivation_info", "This license can still be reactivated on this device.")
			} else {
				problem.WithExtension("can_reactivate", false)
				problem.WithExtension("reactivation_info", "Maximum reactivations reached. Contact support for assistance.")
			}
		}
		if details.ActivationDate != nil {
			problem.WithExtension("activation_date", details.ActivationDate.Format("2006-01-02T15:04:05Z"))
		}
		if details.ExpiryDate != nil {
			problem.WithExtension("expiry_date", details.ExpiryDate.Format("2006-01-02T15:04:05Z"))
		}
	}
	
	return problem
}

// Error mapping utilities and context enrichment

// EnrichLicenseErrorWithContext adds additional context to a license error
func EnrichLicenseErrorWithContext(ctx context.Context, err error, operation string, licenseKey string, metadata map[string]interface{}) error {
	var licenseErr *LicenseError
	if errors.As(err, &licenseErr) {
		// Enhance existing LicenseError
		licenseErr.WithContext(ctx)
		if licenseKey != "" {
			licenseErr.WithLicenseKey(licenseKey)
		}
		if operation != "" {
			licenseErr.Operation = operation
		}
		for k, v := range metadata {
			licenseErr.WithExtension(k, v)
		}
		return licenseErr
	}

	// Wrap generic error in LicenseError
	le := &LicenseError{
		Type:      "/errors/license-operation-failed",
		Title:     "License Operation Failed",
		Status:    http.StatusInternalServerError,
		Detail:    err.Error(),
		ErrorCode: "LICENSE_OPERATION_FAILED",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: operation,
		Extensions: make(map[string]interface{}),
		Err: err,
	}

	le.WithContext(ctx).WithLicenseKey(licenseKey)
	for k, v := range metadata {
		le.WithExtension(k, v)
	}

	return le
}

// MapLicenseErrorToHTTP maps license errors to HTTP status codes and problem details
func MapLicenseErrorToHTTP(ctx context.Context, err error, traceID string) render.Renderer {
	// First check if it's already a LicenseError
	var licenseErr *LicenseError
	if errors.As(err, &licenseErr) {
		// Ensure trace ID is set
		if licenseErr.TraceID == "" {
			licenseErr.TraceID = traceID
		}
		return licenseErr.ToProblemDetails()
	}

	// Use existing MapLicenseError function for backward compatibility
	return MapLicenseError(err, traceID)
}

// CreateLicenseErrorFromValidation creates a license error from validation failures
func CreateLicenseErrorFromValidation(ctx context.Context, failures []string, licenseKey string) *LicenseError {
	le := &LicenseError{
		Type:      "/errors/validation-failed",
		Title:     "License Validation Failed",
		Status:    http.StatusBadRequest,
		Detail:    fmt.Sprintf("License validation failed: %s", strings.Join(failures, ", ")),
		ErrorCode: "VALIDATION_FAILED",
		Timestamp: time.Now(),
		Service:   "license-validator",
		Operation: "validate_license",
		Extensions: map[string]interface{}{
			"validation_failures": failures,
			"failure_count":       len(failures),
		},
	}

	return le.WithContext(ctx).WithLicenseKey(licenseKey)
}

// IsLicenseError checks if an error is a license error
func IsLicenseError(err error) bool {
	var licenseErr *LicenseError
	return errors.As(err, &licenseErr)
}

// GetLicenseErrorCode extracts the error code from a license error
func GetLicenseErrorCode(err error) string {
	var licenseErr *LicenseError
	if errors.As(err, &licenseErr) {
		return licenseErr.ErrorCode
	}

	// Check sentinel errors
	switch {
	case errors.Is(err, ErrLicenseExpired):
		return "LICENSE_EXPIRED"
	case errors.Is(err, ErrLicenseNotActivated):
		return "LICENSE_NOT_ACTIVATED"
	case errors.Is(err, ErrInvalidLicenseKey):
		return "INVALID_LICENSE_KEY"
	case errors.Is(err, ErrInvalidLicenseFormat):
		return "INVALID_LICENSE_FORMAT"
	case errors.Is(err, ErrActivationFailed):
		return "ACTIVATION_FAILED"
	case errors.Is(err, ErrRateLimited):
		return "RATE_LIMITED"
	case errors.Is(err, ErrNetworkError):
		return "NETWORK_ERROR"
	case errors.Is(err, ErrLicenseValidationFailed):
		return "VALIDATION_FAILED"
	case errors.Is(err, ErrLicenseAlreadyActivated):
		return "LICENSE_ALREADY_ACTIVATED"
	case errors.Is(err, ErrReactivationLimitExceeded):
		return "REACTIVATION_LIMIT_EXCEEDED"
	case errors.Is(err, ErrAlreadyActivatedOnDevice):
		return "DEVICE_MISMATCH"
	default:
		return "UNKNOWN_LICENSE_ERROR"
	}
}

// IsRecoverableLicenseError checks if a license error is recoverable (user can take action)
func IsRecoverableLicenseError(err error) bool {
	errorCode := GetLicenseErrorCode(err)

	recoverableErrors := map[string]bool{
		"LICENSE_NOT_ACTIVATED":       true,
		"INVALID_LICENSE_KEY":         true,
		"INVALID_LICENSE_FORMAT":      true,
		"RATE_LIMITED":                true,
		"NETWORK_ERROR":               true,
		"VALIDATION_FAILED":           true,
		"DEVICE_MISMATCH":             true, // Can reactivate
	}

	return recoverableErrors[errorCode]
}

// GetLicenseErrorRecoveryInfo provides recovery information for license errors
func GetLicenseErrorRecoveryInfo(err error) map[string]interface{} {
	errorCode := GetLicenseErrorCode(err)

	recoveryInfo := map[string]interface{}{
		"recoverable":      IsRecoverableLicenseError(err),
		"error_code":       errorCode,
	}

	switch errorCode {
	case "LICENSE_NOT_ACTIVATED":
		recoveryInfo["action"] = "activate_license"
		recoveryInfo["message"] = "Please activate a license to continue using the application"
		recoveryInfo["url"] = "/api/license/activate"

	case "LICENSE_EXPIRED":
		recoveryInfo["action"] = "renew_license"
		recoveryInfo["message"] = "Your license has expired. Please renew it to continue"
		recoveryInfo["url"] = "https://isxpulse.com/renew"

	case "INVALID_LICENSE_KEY", "INVALID_LICENSE_FORMAT":
		recoveryInfo["action"] = "check_license_key"
		recoveryInfo["message"] = "Please verify your license key and try again"
		recoveryInfo["expected_format"] = "ISX-XXXX-XXXX-XXXX-XXXX"

	case "RATE_LIMITED":
		recoveryInfo["action"] = "wait_and_retry"
		recoveryInfo["message"] = "Too many attempts. Please wait before trying again"
		recoveryInfo["wait_time"] = "15 minutes"

	case "NETWORK_ERROR":
		recoveryInfo["action"] = "check_connection"
		recoveryInfo["message"] = "Please check your internet connection and try again"

	case "DEVICE_MISMATCH":
		recoveryInfo["action"] = "reactivate_or_contact_support"
		recoveryInfo["message"] = "You may be able to reactivate this license or contact support"

	case "REACTIVATION_LIMIT_EXCEEDED":
		recoveryInfo["action"] = "contact_support"
		recoveryInfo["message"] = "Please contact support to transfer your license"
		recoveryInfo["support_email"] = "support@isxpulse.com"

	default:
		recoveryInfo["action"] = "contact_support"
		recoveryInfo["message"] = "An unexpected error occurred. Please contact support"
		recoveryInfo["support_email"] = "support@isxpulse.com"
	}

	return recoveryInfo
}

// MapLicenseError maps domain errors to HTTP problem details
func MapLicenseError(err error, traceID string) render.Renderer {
	instance := fmt.Sprintf("/api/license#trace-%s", traceID)
	
	// Check if it's an APIError from errors.go
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode == "LICENSE_NOT_FOUND" {
			return NewProblemDetails(
				http.StatusNotFound,
				"/errors/license-not-found",
				"License Not Found",
				"No license file found in the system. Please activate a license.",
				instance,
			).WithExtension("trace_id", traceID).
				WithExtension("error_code", "LICENSE_NOT_FOUND")
		}
		if apiErr.ErrorCode == "VALIDATION_FAILED" {
			return NewProblemDetails(
				http.StatusBadRequest,
				"/errors/validation-failed",
				"License Validation Failed",
				apiErr.Message,
				instance,
			).WithExtension("trace_id", traceID).
				WithExtension("error_code", "VALIDATION_FAILED")
		}
	}
	
	switch {
	case errors.Is(err, ErrLicenseReactivated):
		return NewLicenseReactivatedResponse(nil, traceID)
	case errors.Is(err, ErrReactivationLimitExceeded):
		return NewReactivationLimitExceededError(nil, traceID)
	case errors.Is(err, ErrAlreadyActivatedOnDevice):
		return NewAlreadyActivatedOnDeviceError(nil, traceID)
	case errors.Is(err, ErrLicenseAlreadyActivated):
		return NewLicenseAlreadyActivatedError(nil, traceID)
	case errors.Is(err, ErrLicenseExpired):
		return NewProblemDetails(
			http.StatusForbidden,
			"/errors/license-expired",
			"License Expired",
			"Your license has expired. Please renew to continue.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "LICENSE_EXPIRED")
			
			
	case errors.Is(err, ErrLicenseNotActivated):
		return NewProblemDetails(
			http.StatusPreconditionRequired,
			"/errors/license-not-activated",
			"License Not Activated",
			"No license has been activated. Please activate a license to continue.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "LICENSE_NOT_ACTIVATED")
			
	case errors.Is(err, ErrInvalidLicenseKey):
		return NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-license-key",
			"Invalid License Key",
			"The provided license key is invalid or malformed.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "INVALID_LICENSE_KEY")
			
	case errors.Is(err, ErrInvalidLicenseFormat):
		return NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-license-format",
			"Invalid License Format",
			"License key must be in format: ISX-XXXX-XXXX-XXXX-XXXX",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "INVALID_LICENSE_FORMAT").
			WithExtension("expected_format", "ISX-XXXX-XXXX-XXXX-XXXX")
			
	case errors.Is(err, ErrActivationFailed):
		return NewProblemDetails(
			http.StatusUnprocessableEntity,
			"/errors/activation-failed",
			"License Activation Failed",
			"Unable to activate the license. Please verify the key and try again.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "ACTIVATION_FAILED")
			
	case errors.Is(err, ErrValidationFailed):
		return NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/validation-failed",
			"License Validation Failed",
			"Unable to validate license status. Please try again later.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "VALIDATION_FAILED")

	case errors.Is(err, ErrLicenseValidationFailed):
		return NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/validation-failed",
			"License Validation Failed",
			"Unable to validate license status. Please try again later.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "VALIDATION_FAILED")

	case errors.Is(err, ErrRateLimited):
		return NewProblemDetails(
			http.StatusTooManyRequests,
			"/errors/rate-limited",
			"Too Many Requests",
			"Too many activation attempts. Please try again later.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "RATE_LIMITED").
			WithExtension("retry_after", 900) // 15 minutes
			
	case errors.Is(err, ErrNetworkError):
		return NewProblemDetails(
			http.StatusServiceUnavailable,
			"/errors/network-error",
			"Network Error",
			"Unable to connect to license server. Please check your connection.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "NETWORK_ERROR")
			
	default:
		// Generic error
		return NewProblemDetails(
			http.StatusInternalServerError,
			"/errors/internal-error",
			"Internal Server Error",
			"An unexpected error occurred while processing your request.",
			instance,
		).WithExtension("trace_id", traceID).
			WithExtension("error_code", "INTERNAL_ERROR")
	}
}