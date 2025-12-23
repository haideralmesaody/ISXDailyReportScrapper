package license

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/security"
)

// LicenseCacheInvalidator interface for cache invalidation functionality
// This allows us to avoid tight coupling with the middleware package
type LicenseCacheInvalidator interface {
	InvalidateCache()
}

// LicenseInfo represents license data
type LicenseInfo struct {
	LicenseKey        string    `json:"license_key"`
	UserEmail         string    `json:"user_email"`
	ExpiryDate        time.Time `json:"expiry_date"`
	Duration          string    `json:"duration"`
	IssuedDate        time.Time `json:"issued_date"`
	Status            string    `json:"status"`
	LastChecked       time.Time `json:"last_checked"`
	ActivationID      string    `json:"activation_id"`      // Unique activation identifier from Apps Script
	DeviceFingerprint string    `json:"device_fingerprint"` // Device fingerprint for validation (legacy)

	// Enhanced hardware fingerprinting fields for weighted confidence validation
	HardwareComponents     map[string]string `json:"hardware_components,omitempty"`      // Individual hardware components
	LastHardwareConfidence float64           `json:"last_hardware_confidence,omitempty"` // Last calculated confidence score
}

// ReactivationDetails holds information about license reactivation
type ReactivationDetails struct {
	ReactivationCount     int       `json:"reactivation_count"`
	MaxReactivations      int       `json:"max_reactivations"`
	SimilarityScore       float64   `json:"similarity_score"`
	PreviousDeviceInfo    string    `json:"previous_device_info"`
	ReactivationTimestamp time.Time `json:"reactivation_timestamp"`
}

// GoogleSheetsConfig represents Google Sheets configuration
type GoogleSheetsConfig struct {
	SheetID           string `json:"sheet_id"`
	APIKey            string `json:"api_key"`
	SheetName         string `json:"sheet_name"`
	UseServiceAccount bool   `json:"use_service_account"`
	// Legacy credential fields removed - only embedded encrypted credentials are supported
}

// PerformanceMetrics tracks operation performance
type PerformanceMetrics struct {
	Count        int64         `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AverageTime  time.Duration `json:"average_time"`
	MaxTime      time.Duration `json:"max_time"`
	MinTime      time.Duration `json:"min_time"`
	ErrorCount   int64         `json:"error_count"`
	SuccessCount int64         `json:"success_count"`
	LastUpdated  time.Time     `json:"last_updated"`
}

// Manager handles license operations with enhanced logging, caching, and security
type Manager struct {
	config             GoogleSheetsConfig
	licenseFile        string
	sheetsService      *sheets.Service
	cache              *LicenseCache
	security           *SecurityManager
	performanceTracker *PerformanceTracker
	validationCache    *ValidationCache
	// OpenTelemetry metrics
	metrics *LicenseMetrics
	// Secure credentials management
	credentialsManager *security.SecureCredentialsManager
	secureMode         bool
	// Device fingerprinting for scratch card system
	fingerprintManager *security.FingerprintManager
	// Hardware fingerprint manager for hardware-only system
	hardwareFingerprinter *security.HardwareFingerprinter
	// Hybrid fingerprint manager for migration support
	hybridFingerprintManager *security.HybridFingerprintManager
	// License validator for middleware cache coordination
	licenseValidator LicenseCacheInvalidator
	// timeNow allows tests to override the current time
	timeNow func() time.Time
}

// ValidationResult holds cached validation results
type ValidationResult struct {
	IsValid     bool
	Error       error
	ErrorType   string // "expired", "network_error", etc.
	CachedUntil time.Time
	RetryAfter  time.Duration
}

// RenewalInfo contains information about license renewal requirements
type RenewalInfo struct {
	DaysLeft     int    `json:"days_left"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	NeedsRenewal bool   `json:"needs_renewal"`
	IsExpired    bool   `json:"is_expired"`
}

// ManagerInterface defines the interface for license managers to enable proper testing and mocking
type ManagerInterface interface {
	// Core license operations
	GetLicenseStatus() (*LicenseInfo, string, error)
	ActivateLicense(key string) error
	ValidateLicense() (bool, error)
	CheckRenewalStatus() (*RenewalInfo, error)

	// License stacking operations
	CheckExistingLicense() (*ExistingLicenseInfo, error)
	GetLicenseInfo() (*LicenseInfo, error)

	// Path operations
	GetLicensePath() string

	// Operations interface compatibility
	CheckLicense() error
	RequiresLicense() bool
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidateScratchCardFormat validates the scratch card license key format
func ValidateScratchCardFormat(licenseKey string) error {
	// Remove dashes and spaces but preserve original casing for validation rules
	cleanKey := strings.ReplaceAll(strings.ReplaceAll(licenseKey, "-", ""), " ", "")

	// Check if it starts with ISX
	if !strings.HasPrefix(cleanKey, "ISX") {
		return fmt.Errorf("license key must start with 'ISX'")
	}

	// Check length - should be 15 characters (ISX + 12 chars) without dashes
	if len(cleanKey) != 15 {
		return fmt.Errorf("license key must be 15 characters long (ISX + 12 characters)")
	}

	// Check that all characters after ISX are alphanumeric
	suffix := cleanKey[3:]
	for _, char := range suffix {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return fmt.Errorf("license key must contain only uppercase letters and numbers")
		}
	}

	return nil
}

// NormalizeScratchCardKey normalizes scratch card key to standard format
func NormalizeScratchCardKey(licenseKey string) string {
	// Remove all dashes and spaces
	cleanKey := strings.ReplaceAll(strings.ReplaceAll(licenseKey, "-", ""), " ", "")
	cleanKey = strings.ToUpper(cleanKey)

	// If it's in the format ISXXX12CHARS, return as-is
	if len(cleanKey) == 15 && strings.HasPrefix(cleanKey, "ISX") {
		return cleanKey
	}

	return cleanKey
}

// FormatScratchCardKeyWithDashes formats key with dashes for display
func FormatScratchCardKeyWithDashes(licenseKey string) string {
	cleanKey := NormalizeScratchCardKey(licenseKey)

	if len(cleanKey) != 15 {
		return cleanKey // Return as-is if invalid length
	}

	// Format as ISX-XXXX-XXXX-XXXX
	return fmt.Sprintf("%s-%s-%s-%s",
		cleanKey[:3],    // ISX
		cleanKey[3:7],   // Next 4 chars
		cleanKey[7:11],  // Next 4 chars
		cleanKey[11:15], // Last 4 chars
	)
}

type managerInitOptions struct {
	licenseFile      string
	pathSource       string
	cacheInvalidator LicenseCacheInvalidator
}

func newManagerWithOptions(opts managerInitOptions) (*Manager, error) {
	ctx := context.Background()
	logger := slog.Default()

	var err error
	opts, err = resolveLicensePath(opts, logger)
	if err != nil {
		return nil, err
	}

	logInitDiagnostics(logger, opts)

	if !config.FileExists(opts.licenseFile) {
		if logger != nil {
			logger.Warn("License file not found - new installation expected",
				slog.String("license_path", opts.licenseFile),
				slog.String("path_source", opts.pathSource),
				slog.String("suggestion", "Use activate-license tool to activate a license"),
			)
		}
	} else if logger != nil {
		logger.Info("License file found and accessible",
			slog.String("license_path", opts.licenseFile),
			slog.String("path_source", opts.pathSource),
		)
	}

	manager, err := initManagerComponents(opts)
	if err != nil {
		return nil, err
	}

	initMessage := "License manager initialized successfully with enhanced diagnostics"
	initAttrs := []slog.Attr{
		slog.String("license_path", opts.licenseFile),
		slog.String("path_source", opts.pathSource),
		slog.Bool("license_exists", config.FileExists(opts.licenseFile)),
		slog.String("cache_ttl", "5m"),
		slog.Int("cache_max_size", 1000),
		slog.Int("security_max_attempts", 5),
		slog.String("security_block_duration", "15m"),
		slog.Bool("secure_mode", true),
	}
	if opts.cacheInvalidator != nil {
		initMessage = "License manager with cache invalidator initialized successfully with enhanced diagnostics"
		initAttrs = append(initAttrs, slog.Bool("has_cache_invalidator", true))
	}

	manager.logInfo(ctx, "manager_initialization", initMessage, initAttrs...)
	manager.logInfo(ctx, "security_initialization", "Security manager initialized with encrypted credentials")

	if err := manager.initializeSheetsService(ctx); err != nil {
		return nil, err
	}

	return manager, nil
}

func (m *Manager) initializeSheetsService(ctx context.Context) error {
	if !m.config.UseServiceAccount || m.credentialsManager == nil {
		return nil
	}

	credentialsJSON, err := m.credentialsManager.GetCredentials(ctx)
	if err != nil {
		m.logError(ctx, "sheets_initialization", "Failed to get embedded credentials",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to get embedded credentials: %v", err)
	}

	if len(credentialsJSON) == 0 {
		m.logError(ctx, "sheets_initialization", "Embedded credentials are empty")
		return fmt.Errorf("embedded credentials are empty - ensure build includes encrypted credentials")
	}

	credentialsOption := option.WithCredentialsJSON(credentialsJSON)
	sheetsService, err := sheets.NewService(ctx, credentialsOption)
	if err != nil {
		m.logError(ctx, "sheets_initialization", "Failed to create Google Sheets service",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create sheets service with embedded credentials: %v", err)
	}

	m.sheetsService = sheetsService
	m.logInfo(ctx, "sheets_initialization", "Google Sheets service initialized with embedded encrypted credentials only")
	return nil
}

// newManagerInternal is the shared internal constructor with enhanced initialization logic
func newManagerInternal(licenseFile string, pathSource string) (*Manager, error) {
	return newManagerWithOptions(managerInitOptions{
		licenseFile: licenseFile,
		pathSource:  pathSource,
	})
}

// newManagerInternalWithCacheInvalidator creates a new license manager with cache invalidation support
func newManagerInternalWithCacheInvalidator(licenseFile string, licenseCacheInvalidator LicenseCacheInvalidator, pathSource string) (*Manager, error) {
	return newManagerWithOptions(managerInitOptions{
		licenseFile:      licenseFile,
		pathSource:       pathSource,
		cacheInvalidator: licenseCacheInvalidator,
	})
}

// NewManagerWithPath creates a new license manager with a custom license file path
// This constructor is ideal for testing and scenarios requiring custom license file locations
func NewManagerWithPath(licenseFile string) (*Manager, error) {
	return newManagerInternal(licenseFile, "custom_path")
}

// NewManagerWithCacheInvalidator creates a new license manager with cache invalidation support
// This is the recommended constructor when license validation middleware is used
func NewManagerWithCacheInvalidator(licenseCacheInvalidator LicenseCacheInvalidator) (*Manager, error) {
	// Use centralized path management system - always resolve relative to executable
	licensePath, err := config.GetLicensePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get license path: %v", err)
	}

	return newManagerInternalWithCacheInvalidator(licensePath, licenseCacheInvalidator, "config_system")
}

// NewManager creates a new license manager using the centralized path management system
// This is the recommended constructor for production use
func NewManager() (*Manager, error) {
	// Use centralized path management system - always resolve relative to executable
	licensePath, err := config.GetLicensePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get license path: %v", err)
	}

	return newManagerWithOptions(managerInitOptions{
		licenseFile: licensePath,
		pathSource:  "config_system",
	})
}

// SetMetrics sets the OpenTelemetry metrics for the manager
func (m *Manager) SetMetrics(metrics *LicenseMetrics) {
	m.metrics = metrics
}

// GenerateLicense creates a new license key
func (m *Manager) GenerateLicense(userEmail string, duration string) (string, error) {
	// Generate random license key
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	licenseKey := base64.URLEncoding.EncodeToString(bytes)
	licenseKey = strings.ReplaceAll(licenseKey, "=", "")

	// Add prefix based on duration
	prefix := "ISX"
	switch duration {
	case "1m":
		prefix = "ISX1M"
	case "3m":
		prefix = "ISX3M"
	case "6m":
		prefix = "ISX6M"
	case "1y":
		prefix = "ISX1Y"
	}

	licenseKey = fmt.Sprintf("%s%s", prefix, licenseKey)

	// Calculate expiry date - expires at 12am next day after standard period
	var standardExpiry time.Time
	switch duration {
	case "1m":
		standardExpiry = time.Now().AddDate(0, 1, 0)
	case "3m":
		standardExpiry = time.Now().AddDate(0, 3, 0)
	case "6m":
		standardExpiry = time.Now().AddDate(0, 6, 0)
	case "1y":
		standardExpiry = time.Now().AddDate(1, 0, 0)
	default:
		standardExpiry = time.Now().AddDate(0, 1, 0)
	}
	// Set expiry to 12:00 AM next day after standard expiry
	expiryDate := time.Date(standardExpiry.Year(), standardExpiry.Month(), standardExpiry.Day()+1, 0, 0, 0, 0, standardExpiry.Location())

	// Create license info
	license := LicenseInfo{
		LicenseKey:  licenseKey,
		UserEmail:   userEmail,
		ExpiryDate:  expiryDate,
		Duration:    duration,
		IssuedDate:  time.Now(),
		Status:      "issued",
		LastChecked: time.Now(),
	}

	// Save to Google Sheets
	if err := m.saveLicenseToSheets(license); err != nil {
		return "", fmt.Errorf("failed to save license: %v", err)
	}

	return licenseKey, nil
}

// ActivateLicense activates a license using background context
func (m *Manager) ActivateLicense(licenseKey string) error {
	return m.ActivateLicenseWithContext(context.Background(), licenseKey)
}

// ActivateLicenseWithContext activates a license with tracing and metrics
func (m *Manager) ActivateLicenseWithContext(ctx context.Context, licenseKey string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	return m.TraceActivation(ctx, licenseKey, func() error {
		return m.TrackOperation("license_activation", func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return m.performActivation(ctx, licenseKey)
		})
	})
}

// ValidateLicense checks if current license is valid with enhanced tracking
func (m *Manager) ValidateLicense() (bool, error) {
	return m.ValidateLicenseWithContext(context.Background())
}

// ValidateLicenseWithContext checks if current license is valid with context and enhanced observability
func (m *Manager) ValidateLicenseWithContext(ctx context.Context) (bool, error) {
	if m.validationCache == nil {
		m.validationCache = NewValidationCache()
	}

	if result, ok := m.validationCache.GetValidResult(); ok {
		if m.metrics != nil {
			m.metrics.ValidationCacheHits.Add(ctx, 1, metric.WithAttributes(
				attribute.String("component", "license_manager"),
				attribute.String("cache_result", "hit"),
			))
		}
		return result.IsValid, sanitizeValidationError(result.Error)
	}

	if m.metrics != nil {
		m.metrics.ValidationCacheMisses.Add(ctx, 1, metric.WithAttributes(
			attribute.String("component", "license_manager"),
			attribute.String("cache_result", "miss"),
		))
	}

	return m.TraceValidation(ctx, func() (bool, error) {
		var valid bool
		var err error

		trackErr := m.TrackOperation("license_validation_complete", func() error {
			valid, err = m.performValidation()
			m.cacheValidationResult(valid, err)

			if !valid {
				return err
			}
			return nil
		})

		if trackErr != nil {
			return false, sanitizeValidationError(trackErr)
		}

		return valid, sanitizeValidationError(err)
	})
}

// GetValidationState returns the current validation state for better user feedback
func (m *Manager) GetValidationState() (*ValidationResult, error) {
	if m.validationCache == nil {
		return nil, fmt.Errorf("no validation performed yet")
	}

	result, ok := m.validationCache.Snapshot()
	if !ok {
		return nil, fmt.Errorf("no validation performed yet")
	}

	return result, nil
}

// GetLicenseInfo returns current license information
func (m *Manager) GetLicenseInfo() (*LicenseInfo, error) {
	license, err := m.loadLicenseLocal()
	if err != nil {
		return nil, err
	}
	return &license, nil
}

// ExistingLicenseInfo provides information about an existing license for pre-activation checks
type ExistingLicenseInfo struct {
	HasLicense    bool      `json:"has_license"`
	DaysRemaining int       `json:"days_remaining"`
	ExpiryDate    time.Time `json:"expiry_date,omitempty"`
	LicenseKey    string    `json:"license_key_masked"` // Masked format: ISX-XXXX-****-****
	Status        string    `json:"status"`
	IsExpired     bool      `json:"is_expired"`
}

// CheckExistingLicense checks if there's an existing license and returns its details
func (m *Manager) CheckExistingLicense() (*ExistingLicenseInfo, error) {
	ctx := context.Background()

	// Try to load existing license
	license, err := m.loadLicenseLocal()
	if err != nil {
		// No existing license
		m.logDebug(ctx, "check_existing_license", "No existing license found",
			slog.String("error", err.Error()),
		)
		return &ExistingLicenseInfo{
			HasLicense: false,
			Status:     "not_activated",
		}, nil
	}

	// Calculate days remaining
	now := time.Now()
	daysRemaining := 0
	isExpired := false

	if now.Before(license.ExpiryDate) {
		daysRemaining = int(time.Until(license.ExpiryDate).Hours() / 24)
	} else {
		isExpired = true
	}

	// Mask the license key for security (show only first segment)
	maskedKey := MaskLicenseKey(license.LicenseKey)

	m.logInfo(ctx, "check_existing_license", "Existing license found",
		slog.String("license_key_masked", maskedKey),
		slog.Int("days_remaining", daysRemaining),
		slog.Bool("is_expired", isExpired),
		slog.String("status", license.Status),
	)

	return &ExistingLicenseInfo{
		HasLicense:    true,
		DaysRemaining: daysRemaining,
		ExpiryDate:    license.ExpiryDate,
		LicenseKey:    maskedKey,
		Status:        license.Status,
		IsExpired:     isExpired,
	}, nil
}

// parseLicenseDuration converts license duration string to time.Duration
func parseLicenseDuration(duration string) time.Duration {
	// Handle various duration formats: "30 days", "1 month", "3 months", "6 months", "1 year"
	duration = strings.ToLower(strings.TrimSpace(duration))

	// Extract number and unit
	var value int
	var unit string

	// Try to parse "X days", "X months", "X month", "X year", etc.
	if _, err := fmt.Sscanf(duration, "%d %s", &value, &unit); err != nil {
		// Try without space: "30days", "1month", etc.
		if _, err := fmt.Sscanf(duration, "%d%s", &value, &unit); err != nil {
			// Default to 30 days if can't parse
			return 30 * 24 * time.Hour
		}
	}

	// Normalize unit
	unit = strings.TrimSuffix(unit, "s") // Remove plural 's'

	switch unit {
	case "day":
		return time.Duration(value) * 24 * time.Hour
	case "month":
		// Approximate: 30 days per month
		return time.Duration(value) * 30 * 24 * time.Hour
	case "year":
		// Approximate: 365 days per year
		return time.Duration(value) * 365 * 24 * time.Hour
	default:
		// Default to days if unit not recognized
		return time.Duration(value) * 24 * time.Hour
	}
}

// LicenseAudit represents a license change audit entry
type LicenseAudit struct {
	Timestamp      time.Time `json:"timestamp"`
	Action         string    `json:"action"` // "activated", "stacked", "replaced_expired", "new_activation"
	PreviousKey    string    `json:"previous_key,omitempty"`
	NewKey         string    `json:"new_key"`
	PreviousExpiry time.Time `json:"previous_expiry,omitempty"`
	NewExpiry      time.Time `json:"new_expiry"`
	DeviceID       string    `json:"device_id"`
	TraceID        string    `json:"trace_id"`
	UserEmail      string    `json:"user_email,omitempty"`
}

// auditLicenseChange logs license changes to audit file
func (m *Manager) auditLicenseChange(ctx context.Context, action string, previousLicense, newLicense LicenseInfo, deviceID string) {
	audit := LicenseAudit{
		Timestamp: time.Now(),
		Action:    action,
		NewKey:    MaskLicenseKey(newLicense.LicenseKey),
		NewExpiry: newLicense.ExpiryDate,
		DeviceID:  deviceID[:min(16, len(deviceID))], // Truncate device ID for privacy
		TraceID:   newLicense.ActivationID,
		UserEmail: newLicense.UserEmail,
	}

	// Add previous license info if applicable
	if previousLicense.LicenseKey != "" {
		audit.PreviousKey = MaskLicenseKey(previousLicense.LicenseKey)
		audit.PreviousExpiry = previousLicense.ExpiryDate
	}

	// Log using structured logging per CLAUDE.md
	m.logInfo(ctx, "license_audit", "License change audited",
		slog.String("action", action),
		slog.String("new_key", audit.NewKey),
		slog.String("previous_key", audit.PreviousKey),
		slog.Time("new_expiry", audit.NewExpiry),
		slog.Time("previous_expiry", audit.PreviousExpiry),
		slog.String("device_id", audit.DeviceID),
		slog.String("trace_id", audit.TraceID),
	)

	// Also write to dedicated audit file
	auditFile := filepath.Join("logs", "license_audit.json")
	if err := m.writeAuditToFile(audit, auditFile); err != nil {
		m.logError(ctx, "audit_write", "Failed to write audit to file",
			slog.String("error", err.Error()),
			slog.String("file", auditFile),
		)
	}
}

// now returns the current time, allowing tests to override for determinism.
func (m *Manager) now() time.Time {
	if m != nil && m.timeNow != nil {
		return m.timeNow()
	}
	return time.Now()
}

// GetLicensePath returns the path to the license file
func (m *Manager) GetLicensePath() string {
	// Prefer the configured path if it exists (even if non-standard name)
	if info, err := os.Stat(m.licenseFile); err == nil && !info.IsDir() {
		return m.licenseFile
	}

	if strings.Contains(m.licenseFile, "license.dat") {
		return m.licenseFile
	}

	// Normalize to standard license filename for interface expectations
	dir := filepath.Dir(m.licenseFile)
	return filepath.Join(dir, "license.dat")
}

// Close properly shuts down the manager and its components
func (m *Manager) Close() error {
	// Stop cache cleanup goroutine
	if m.cache != nil {
		m.cache.Stop()
	}

	// Stop security manager cleanup goroutine
	if m.security != nil {
		m.security.Stop()
	}

	// Close secure credentials manager if in secure mode
	if m.secureMode && m.credentialsManager != nil {
		if err := m.credentialsManager.Close(); err != nil {
			m.logError(context.Background(), "credentials_manager_close", "Failed to close credentials manager",
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("failed to close credentials manager: %v", err)
		}
	}

	// Log manager shutdown
	m.logInfo(context.Background(), "manager_shutdown", "License manager closed successfully",
		slog.Bool("secure_mode", m.secureMode),
	)

	return nil
}

// validateLicenseFromAppsScriptWithCache validates license via Apps Script with caching support
func (m *Manager) validateLicenseFromAppsScriptWithCache(licenseKey string) (LicenseInfo, error) {
	// Check cache first
	if m.cache != nil {
		if cachedInfo, found := m.cache.Get(licenseKey); found {
			m.logDebug(context.Background(), "cache_hit", "License found in cache",
				slog.String("license_key_prefix", licenseKey[:min(8, len(licenseKey))]),
			)
			return *cachedInfo, nil
		}
	}

	// Cache miss - fetch from Apps Script
	licenseInfo, err := m.validateLicenseFromAppsScript(licenseKey)
	if err != nil {
		return licenseInfo, err
	}

	// Store in cache
	if m.cache != nil {
		m.cache.Set(licenseKey, licenseInfo)
		m.logDebug(context.Background(), "cache_store", "License stored in cache",
			slog.String("license_key_prefix", licenseKey[:min(8, len(licenseKey))]),
		)
	}

	return licenseInfo, nil
}

// CheckLicense implements operations.LicenseChecker interface
func (m *Manager) CheckLicense() error {
	_, err := m.ValidateLicense()
	return err
}

// RequiresLicense implements operations.LicenseChecker interface
func (m *Manager) RequiresLicense() bool {
	info, _, err := m.GetLicenseStatus()
	if err != nil {
		return true // Assume license required if validation fails
	}
	return info != nil
}
