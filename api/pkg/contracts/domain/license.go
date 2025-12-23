// Package domain contains the core domain models for the ISX Daily Reports Scrapper.
// These types serve as the Single Source of Truth (SSOT) for all layers of the application.
package domain

import (
	"time"
)

// LicenseInfo represents the complete information about a license
type LicenseInfo struct {
	LicenseKey     string    `json:"license_key" db:"license_key" validate:"required,min=10"`
	UserEmail      string    `json:"user_email" db:"user_email" validate:"email"`
	ExpiryDate     time.Time `json:"expiry_date" db:"expiry_date" validate:"required"`
	Status         LicenseStatus `json:"status" db:"status" validate:"required"`
	ActivationDate time.Time `json:"activation_date" db:"activation_date"`
	LastCheckDate  time.Time `json:"last_check_date" db:"last_check_date"`
	Features       []string  `json:"features" db:"features"`
	MaxActivations int       `json:"max_activations" db:"max_activations" validate:"min=1"`
	CurrentActivations int   `json:"current_activations" db:"current_activations" validate:"min=0"`
	Duration       LicenseDuration `json:"duration" db:"duration"`
	Tier           string    `json:"tier" db:"tier"` // basic, professional, enterprise
	Metadata       map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// LicenseStatus represents the status of a license
type LicenseStatus string

const (
	LicenseStatusActive    LicenseStatus = "active"
	LicenseStatusSuspended LicenseStatus = "suspended"
	LicenseStatusExpired   LicenseStatus = "expired"
	LicenseStatusRevoked   LicenseStatus = "revoked"
)

// LicenseDuration represents the duration type of a license
type LicenseDuration string

const (
	LicenseDurationMonthly   LicenseDuration = "monthly"
	LicenseDurationQuarterly LicenseDuration = "quarterly"
	LicenseDurationYearly    LicenseDuration = "yearly"
	LicenseDurationLifetime  LicenseDuration = "lifetime"
)

// GoogleSheetsConfig represents configuration for Google Sheets integration
type GoogleSheetsConfig struct {
	SpreadsheetID   string `json:"spreadsheet_id" validate:"required"`
	CredentialsPath string `json:"credentials_path" validate:"required,file"`
	SheetName       string `json:"sheet_name" validate:"required"`
	RangeStart      string `json:"range_start" validate:"required"`
	RangeEnd        string `json:"range_end" validate:"required"`
	UpdateInterval  int    `json:"update_interval" validate:"min=60"` // seconds
}

// ValidationResult represents the result of license validation
type ValidationResult struct {
	Valid           bool      `json:"valid"`
	Reason          string    `json:"reason,omitempty"`
	ExpiresIn       int       `json:"expires_in,omitempty"` // days
	Warnings        []string  `json:"warnings,omitempty"`
	Features        []string  `json:"features,omitempty"`
	Limitations     map[string]interface{} `json:"limitations,omitempty"`
	CheckedAt       time.Time `json:"checked_at"`
	NextCheckAt     time.Time `json:"next_check_at"`
}

// RenewalInfo represents license renewal information
type RenewalInfo struct {
	Eligible       bool      `json:"eligible"`
	RenewalDate    time.Time `json:"renewal_date"`
	GracePeriodEnd time.Time `json:"grace_period_end"`
	DiscountPercent float64  `json:"discount_percent,omitempty"`
	RenewalURL     string    `json:"renewal_url,omitempty"`
}

// StateFile represents the license state file structure
type StateFile struct {
	Version        string    `json:"version"`
	LicenseKey     string    `json:"license_key"`
	EncryptedData  string    `json:"encrypted_data"`
	Checksum       string    `json:"checksum"`
	LastModified   time.Time `json:"last_modified"`
}

// PerformanceMetrics represents license-related performance metrics
type PerformanceMetrics struct {
	ValidationTime   float64 `json:"validation_time_ms"`
	APICallDuration  float64 `json:"api_call_duration_ms"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
	ValidationCount  int64   `json:"validation_count"`
	FailureCount     int64   `json:"failure_count"`
	LastFailure      time.Time `json:"last_failure,omitempty"`
}

// LicenseActivationRequest represents a license activation request
type LicenseActivationRequest struct {
	LicenseKey string `json:"license_key" validate:"required,min=10"`
	Email      string `json:"email" validate:"required,email"`
}

// LicenseActivationResponse represents a license activation response
type LicenseActivationResponse struct {
	Success      bool         `json:"success"`
	LicenseInfo  *LicenseInfo `json:"license_info,omitempty"`
	Message      string       `json:"message"`
	ErrorCode    string       `json:"error_code,omitempty"`
	ActivatedAt  time.Time    `json:"activated_at"`
}

// License error codes
const (
	ErrCodeInvalidLicense     = "INVALID_LICENSE"
	ErrCodeExpiredLicense     = "LICENSE_EXPIRED"
	ErrCodeNetworkError       = "NETWORK_ERROR"
	ErrCodeQuotaExceeded      = "QUOTA_EXCEEDED"
	ErrCodeInvalidFormat      = "INVALID_FORMAT"
	ErrCodeActivationFailed   = "ACTIVATION_FAILED"
	ErrCodeValidationFailed   = "VALIDATION_FAILED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)