package app

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	apperrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/security"

	"log/slog"
)

// maskLicenseKey masks a license key for logging by showing only first and last few characters
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

// ValidateLicenseOnStartup validates the license and returns a rich error on failure.
// The caller decides whether to stop or continue in unlicensed mode (non-fail-fast approach).
func (a *Application) ValidateLicenseOnStartup(ctx context.Context) error {
	// Create a new context with trace ID for license validation
	traceID := infrastructure.GenerateTraceID()
	ctx = infrastructure.WithTraceID(ctx, traceID)

	a.Logger.InfoContext(ctx, "Performing startup license validation",
		slog.String("trace_id", traceID),
		slog.String("version", VERSION),
		slog.String("operation", "startup_validation"))

	// Step 1: Check if license file exists
	paths, err := config.GetPaths()
	if err != nil {
		a.Logger.ErrorContext(ctx, "Failed to get application paths",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))
		return fmt.Errorf("failed to get paths for license validation: %w", err)
	}

	// Step 2: Initialize license manager for validation (always create, even without license file)
	licenseManager, err := license.NewManager()
	if err != nil {
		licenseErr := apperrors.NewActivationFailedError(ctx, "", "failed to initialize license manager", err)
		apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", "")

		a.Logger.ErrorContext(ctx, "Failed to initialize license manager",
			slog.String("license_file", paths.LicenseFile),
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))

		fmt.Println()
		fmt.Println("🚫 LICENSE VALIDATION FAILED")
		fmt.Println("================================")
		fmt.Println("❌ Failed to initialize license manager.")
		fmt.Printf("📁 License file: %s\n", paths.LicenseFile)
		fmt.Printf("🔍 Error: %s\n", err.Error())
		fmt.Println()
		fmt.Println("🔧 To resolve this issue:")
		fmt.Println("1. Check if the license file is corrupted")
		fmt.Println("2. Ensure the license file has proper permissions")
		fmt.Println("3. Try re-activating your license")
		fmt.Println("================================")
		fmt.Println()

		return licenseErr
	}

	// Always set the license manager, even if license validation fails
	a.LicenseManager = licenseManager

	// Step 3: Check if license file exists for status reporting
	if !config.FileExists(paths.LicenseFile) {
		// Log the license not found error with structured logging but don't return
		licenseErr := apperrors.NewLicenseNotActivatedError(ctx)
		apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", "")

		a.Logger.WarnContext(ctx, "License file not found on startup - server will continue for web activation",
			slog.String("license_file", paths.LicenseFile),
			slog.String("trace_id", traceID),
			slog.String("error_code", "LICENSE_NOT_ACTIVATED"),
			slog.String("recovery_action", "web_activation"))

		// Print user-friendly message
		fmt.Println()
		fmt.Println("🚫 LICENSE VALIDATION FAILED")
		fmt.Println("================================")
		fmt.Println("❌ No license file found.")
		fmt.Printf("📁 Expected license file: %s\n", paths.LicenseFile)
		fmt.Println()
		fmt.Println("🔧 To resolve this issue:")
		fmt.Println("1. Start the application and navigate to the license page")
		fmt.Println("2. Enter your license key to activate the license")
		fmt.Println("3. The server is running and ready for activation")
		fmt.Println()
		fmt.Println("🌐 Server will start in unlicensed mode")
		fmt.Println("📍 Navigate to: http://localhost:8080/license")
		fmt.Println("================================")
		fmt.Println()

		// Return the error to be caught by the calling function, but don't prevent startup
		return licenseErr
	}

	// Step 4: Get license status
	licenseInfo, licenseStatus, err := a.LicenseManager.GetLicenseStatus()
	if err != nil {
		// Guard against nil licenseInfo to prevent panic
		licenseKey := "unknown"
		if licenseInfo != nil {
			licenseKey = licenseInfo.LicenseKey
		}
		licenseErr := apperrors.NewLicenseValidationFailed(ctx, licenseKey, "failed to get license status", err)
		apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", licenseKey)

		a.Logger.ErrorContext(ctx, "Failed to get license status",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))

		fmt.Println()
		fmt.Println("🚫 LICENSE VALIDATION FAILED")
		fmt.Println("================================")
		fmt.Println("❌ Failed to retrieve license status.")
		fmt.Printf("🔍 Error: %s\n", err.Error())
		fmt.Println()
		fmt.Println("🔧 To resolve this issue:")
		fmt.Println("1. Check your internet connection")
		fmt.Println("2. Verify your license is still valid")
		fmt.Println("3. Try re-activating your license")
		fmt.Println("================================")
		fmt.Println()

		return licenseErr
	}

	// Step 4: Validate license status
	if licenseInfo == nil {
		licenseErr := apperrors.NewLicenseNotActivatedError(ctx)
		apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", "")

		a.Logger.ErrorContext(ctx, "License information is nil",
			slog.String("license_status", licenseStatus),
			slog.String("trace_id", traceID))

		fmt.Println()
		fmt.Println("🚫 LICENSE VALIDATION FAILED")
		fmt.Println("================================")
		fmt.Println("❌ No valid license information found.")
		fmt.Println()
		fmt.Println("🔧 To resolve this issue:")
		fmt.Println("1. Activate a valid license key")
		fmt.Println("2. Ensure the license has been properly validated")
		fmt.Println("================================")
		fmt.Println()

		return licenseErr
	}

	// Step 5: Check if license is active (case-insensitive comparison)
	if !strings.EqualFold(licenseStatus, "active") && !strings.EqualFold(licenseStatus, "warning") {
		// Determine specific error type based on status
		var licenseErr error
		switch licenseStatus {
		case "expired":
			licenseErr = apperrors.NewLicenseExpiredError(ctx, licenseInfo.ExpiryDate, licenseInfo.LicenseKey)
		case "revoked":
			licenseErr = apperrors.NewActivationFailedError(ctx, licenseInfo.LicenseKey, "license has been revoked", apperrors.ErrLicenseValidationFailed)
		case "suspended":
			licenseErr = apperrors.NewActivationFailedError(ctx, licenseInfo.LicenseKey, "license has been suspended", apperrors.ErrLicenseValidationFailed)
		default:
			licenseErr = apperrors.NewLicenseNotActivatedError(ctx)
		}

		apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", licenseInfo.LicenseKey)

		a.Logger.ErrorContext(ctx, "License validation failed - invalid status",
			slog.String("license_status", licenseStatus),
			slog.String("license_key", maskLicenseKey(licenseInfo.LicenseKey)),
			slog.Time("expiry_date", licenseInfo.ExpiryDate),
			slog.String("trace_id", traceID))

		fmt.Println()
		fmt.Println("🚫 LICENSE VALIDATION FAILED")
		fmt.Println("================================")
		fmt.Printf("❌ License status: %s\n", strings.ToUpper(licenseStatus))

		if licenseStatus == "expired" {
			daysExpired := int(time.Since(licenseInfo.ExpiryDate).Hours() / 24)
			fmt.Printf("📅 Expired on: %s (%d days ago)\n", licenseInfo.ExpiryDate.Format("January 2, 2006"), daysExpired)
			fmt.Println()
			fmt.Println("🔧 To resolve this issue:")
			fmt.Println("1. Renew your license at: https://isxpulse.com/renew")
			fmt.Println("2. Enter the new license key in the application")
		} else {
			fmt.Println()
			fmt.Println("🔧 To resolve this issue:")
			fmt.Println("1. Verify your license key is correct")
			fmt.Println("2. Check if your license has been suspended or revoked")
			fmt.Println("3. Contact support for assistance")
		}
		fmt.Println("================================")
		fmt.Println()

		return licenseErr
	}

	// Step 6: Validate license expiry
	if time.Now().After(licenseInfo.ExpiryDate) {
		licenseErr := apperrors.NewLicenseExpiredError(ctx, licenseInfo.ExpiryDate, licenseInfo.LicenseKey)
		apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", licenseInfo.LicenseKey)

		a.Logger.ErrorContext(ctx, "License validation failed - license expired",
			slog.Time("expiry_date", licenseInfo.ExpiryDate),
			slog.String("license_key", maskLicenseKey(licenseInfo.LicenseKey)),
			slog.Int("days_expired", int(time.Since(licenseInfo.ExpiryDate).Hours()/24)),
			slog.String("trace_id", traceID))

		daysExpired := int(time.Since(licenseInfo.ExpiryDate).Hours() / 24)
		fmt.Println()
		fmt.Println("🚫 LICENSE VALIDATION FAILED")
		fmt.Println("================================")
		fmt.Println("❌ License has expired.")
		fmt.Printf("📅 Expired on: %s (%d days ago)\n", licenseInfo.ExpiryDate.Format("January 2, 2006"), daysExpired)
		fmt.Println()
		fmt.Println("🔧 To resolve this issue:")
		fmt.Println("1. Renew your license at: https://isxpulse.com/renew")
		fmt.Println("2. Enter the new license key in the application")
		fmt.Println("================================")
		fmt.Println()

		return licenseErr
	}

	// Step 7: Validate device fingerprint if present (using enhanced validation)
	if licenseInfo.DeviceFingerprint != "" {
		isValid, currentFingerprint, err := a.validateHardwareFingerprint(ctx, licenseInfo)
		if err != nil {
			a.Logger.WarnContext(ctx, "Failed to validate device fingerprint",
				slog.String("error", err.Error()),
				slog.String("trace_id", traceID))
			// Continue without failing if fingerprint validation has issues
		} else if !isValid {
			currentFpStr := currentFingerprint
			if currentFpStr == "" {
				currentFpStr = "unknown"
			}

			licenseErr := apperrors.NewDeviceMismatchError(ctx, currentFpStr, licenseInfo.DeviceFingerprint, 0.0)
			apperrors.LogLicenseError(ctx, licenseErr, "startup_validation", licenseInfo.LicenseKey)

			a.Logger.ErrorContext(ctx, "License validation failed - device mismatch",
				slog.String("registered_device", licenseInfo.DeviceFingerprint),
				slog.String("current_device", currentFpStr),
				slog.String("trace_id", traceID))

			fmt.Println()
			fmt.Println("🚫 LICENSE VALIDATION FAILED")
			fmt.Println("================================")
			fmt.Println("❌ Device mismatch detected.")
			fmt.Println("🔍 This license is registered to a different device.")
			fmt.Println()
			fmt.Println("🔧 To resolve this issue:")
			fmt.Println("1. Try to reactivate your license on this device")
			fmt.Println("2. If that fails, contact support for assistance")
			fmt.Println("3. Provide your license key and device information")
			fmt.Println("================================")
			fmt.Println()

			return licenseErr
		} else {
			a.Logger.InfoContext(ctx, "Device fingerprint validation successful",
				slog.String("trace_id", traceID))
		}
	}

	// Step 8: Log successful validation
	daysRemaining := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)

	a.Logger.InfoContext(ctx, "License validation successful",
		slog.String("license_status", licenseStatus),
		slog.String("license_key", maskLicenseKey(licenseInfo.LicenseKey)),
		slog.Time("expiry_date", licenseInfo.ExpiryDate),
		slog.Int("days_remaining", daysRemaining),
		slog.String("user_email", licenseInfo.UserEmail),
		slog.String("trace_id", traceID))

	// Show license expiry warning if needed
	if daysRemaining <= 30 {
		a.showLicenseExpiryWarning()
	}

	fmt.Println()
	fmt.Println("✅ LICENSE VALIDATION SUCCESSFUL")
	fmt.Println("================================")
	fmt.Printf("📅 License Status: %s\n", strings.ToUpper(licenseStatus))
	fmt.Printf("📅 Expires: %s (%d days)\n", licenseInfo.ExpiryDate.Format("January 2, 2006"), daysRemaining)
	if licenseInfo.UserEmail != "" {
		fmt.Printf("📧 Licensed to: %s\n", licenseInfo.UserEmail)
	}
	fmt.Println("================================")
	fmt.Println()

	return nil
}

// ValidateLicenseRuntime performs runtime license validation for health checks
// This is less strict than startup validation and returns detailed status
func (a *Application) ValidateLicenseRuntime(ctx context.Context) (*LicenseValidationResult, error) {
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = infrastructure.GenerateTraceID()
		ctx = infrastructure.WithTraceID(ctx, traceID)
	}

	result := &LicenseValidationResult{
		IsValid:     false,
		Status:      "unknown",
		Error:       "", // Empty string for no error
		TraceID:     traceID,
		ValidatedAt: time.Now(),
	}

	// Check if license manager is available
	if a.LicenseManager == nil {
		result.Error = "license manager not initialized"
		result.Status = "manager_unavailable"
		a.Logger.WarnContext(ctx, "License manager not available for runtime validation",
			slog.String("trace_id", traceID))
		return result, fmt.Errorf(result.Error)
	}

	// Get license status
	licenseInfo, licenseStatus, err := a.LicenseManager.GetLicenseStatus()
	if err != nil {
		result.Error = err.Error()
		result.Status = "validation_failed"
		a.Logger.ErrorContext(ctx, "Runtime license validation failed",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))
		return result, err
	}

	result.Status = licenseStatus

	// Check license info
	if licenseInfo == nil {
		result.Error = "no license information available"
		result.Status = "no_license"
		return result, fmt.Errorf(result.Error)
	}

	// Check if license is active
	if licenseStatus != "active" && licenseStatus != "warning" {
		result.IsValid = false
		result.Error = fmt.Sprintf("license status is %s", licenseStatus)
		result.DaysRemaining = int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
		return result, fmt.Errorf(result.Error)
	}

	// Check expiry
	if time.Now().After(licenseInfo.ExpiryDate) {
		result.IsValid = false
		result.Error = fmt.Sprintf("license expired on %s", licenseInfo.ExpiryDate.Format("2006-01-02"))
		result.Status = "expired"
		result.DaysRemaining = -int(time.Since(licenseInfo.ExpiryDate).Hours() / 24)
		return result, fmt.Errorf(result.Error)
	}

	// Validate device fingerprint if present (using enhanced validation)
	if licenseInfo.DeviceFingerprint != "" {
		isValid, _, err := a.validateHardwareFingerprint(ctx, licenseInfo)
		if err != nil {
			a.Logger.WarnContext(ctx, "Failed to validate device fingerprint during runtime check",
				slog.String("error", err.Error()),
				slog.String("trace_id", traceID))
		} else if !isValid {
			result.IsValid = false
			result.Error = "device fingerprint mismatch"
			result.Status = "device_mismatch"
			return result, fmt.Errorf(result.Error)
		}
	}

	// All checks passed
	result.IsValid = true
	result.DaysRemaining = int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)
	result.LicenseKey = maskLicenseKey(licenseInfo.LicenseKey)
	result.ExpiryDate = licenseInfo.ExpiryDate
	result.UserEmail = licenseInfo.UserEmail

	a.Logger.InfoContext(ctx, "Runtime license validation successful",
		slog.String("license_status", licenseStatus),
		slog.String("license_key", result.LicenseKey),
		slog.Int("days_remaining", result.DaysRemaining),
		slog.String("trace_id", traceID))

	return result, nil
}

// validateHardwareFingerprint performs hardware validation using the security package directly.
// It returns whether the current device matches the stored fingerprint, along with the current fingerprint for reporting.
func (a *Application) validateHardwareFingerprint(ctx context.Context, licenseInfo *license.LicenseInfo) (bool, string, error) {
	const confidenceThreshold = 0.7

	fingerprinter := security.NewHardwareFingerprinter()

	// Prefer enhanced validation when component data is available.
	if len(licenseInfo.HardwareComponents) > 0 {
		match, err := fingerprinter.ValidateHardwareFingerprintWithComponents(licenseInfo.DeviceFingerprint, licenseInfo.HardwareComponents)
		if err != nil {
			return false, "", err
		}

		a.Logger.DebugContext(ctx, "Hardware fingerprint validation (components)",
			slog.String("match_type", match.MatchType),
			slog.Float64("confidence", match.Confidence),
		)

		return match.Confidence >= confidenceThreshold, match.CurrentFingerprint, nil
	}

	// Fallback to legacy fingerprint-only validation.
	match, err := fingerprinter.ValidateHardwareFingerprint(licenseInfo.DeviceFingerprint)
	if err != nil {
		return false, "", err
	}

	a.Logger.DebugContext(ctx, "Hardware fingerprint validation (legacy)",
		slog.String("match_type", match.MatchType),
		slog.Float64("confidence", match.Confidence),
	)

	return match.Confidence >= confidenceThreshold, match.CurrentFingerprint, nil
}

// LicenseValidationResult represents the result of license validation
type LicenseValidationResult struct {
	IsValid       bool      `json:"is_valid"`
	Status        string    `json:"status"`
	Error         string    `json:"error,omitempty"`
	TraceID       string    `json:"trace_id"`
	ValidatedAt   time.Time `json:"validated_at"`
	DaysRemaining int       `json:"days_remaining"`
	LicenseKey    string    `json:"license_key,omitempty"`
	ExpiryDate    time.Time `json:"expiry_date,omitempty"`
	UserEmail     string    `json:"user_email,omitempty"`
}

// ShutdownOnLicenseFailure handles graceful shutdown when license validation fails
func (a *Application) ShutdownOnLicenseFailure(ctx context.Context, licenseErr error) {
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = infrastructure.GenerateTraceID()
		ctx = infrastructure.WithTraceID(ctx, traceID)
	}

	a.Logger.ErrorContext(ctx, "Application shutting down due to license validation failure",
		slog.String("error", licenseErr.Error()),
		slog.String("trace_id", traceID),
		slog.String("shutdown_reason", "license_validation_failed"))

	// Log any additional context from the license error
	var licErr *apperrors.LicenseError
	if stderrors.As(licenseErr, &licErr) {
		a.Logger.ErrorContext(ctx, "License error details",
			slog.String("error_code", licErr.ErrorCode),
			slog.String("error_type", licErr.Type),
			slog.String("error_detail", licErr.Detail),
			slog.String("license_key", licErr.LicenseKey))
	}

	fmt.Println()
	fmt.Println("🚫 APPLICATION SHUTDOWN")
	fmt.Println("========================")
	fmt.Println("❌ ISX Pulse cannot start due to license validation failure.")
	fmt.Println()
	fmt.Println("Please resolve the license issue and restart the application.")
	fmt.Println("========================")
	fmt.Println()

	// Since this is called during startup, we don't need to shut down services
	// Just log the failure and exit gracefully
	os.Exit(1)
}

// showLicenseExpiryWarning checks license status and displays color-coded expiry warnings
func (a *Application) showLicenseExpiryWarning() {
	if a.LicenseManager == nil {
		return
	}

	// Get license status
	licenseInfo, status, err := a.LicenseManager.GetLicenseStatus()
	if err != nil || licenseInfo == nil {
		// No license or error - don't show warning
		return
	}

	// Calculate days remaining
	daysRemaining := int(time.Until(licenseInfo.ExpiryDate).Hours() / 24)

	// Only show warnings if license is expiring within 30 days
	if daysRemaining > 30 {
		return
	}

	// Determine warning color and message based on days remaining
	var colorCode, resetCode, icon, urgencyLevel string
	var showWarning bool

	switch {
	case daysRemaining <= 0:
		// Expired - Red
		colorCode = "\033[31m" // Red
		resetCode = "\033[0m"
		icon = "🚨"
		urgencyLevel = "EXPIRED"
		showWarning = true
	case daysRemaining <= 1:
		// 1 day - Bright Red
		colorCode = "\033[91m" // Bright Red
		resetCode = "\033[0m"
		icon = "🚨"
		urgencyLevel = "URGENT"
		showWarning = true
	case daysRemaining <= 3:
		// 2-3 days - Red
		colorCode = "\033[31m" // Red
		resetCode = "\033[0m"
		icon = "⚠️"
		urgencyLevel = "CRITICAL"
		showWarning = true
	case daysRemaining <= 7:
		// 4-7 days - Orange/Red
		colorCode = "\033[31m" // Red
		resetCode = "\033[0m"
		icon = "⚠️"
		urgencyLevel = "WARNING"
		showWarning = true
	case daysRemaining <= 15:
		// 8-15 days - Orange/Yellow
		colorCode = "\033[33m" // Yellow
		resetCode = "\033[0m"
		icon = "⚠️"
		urgencyLevel = "NOTICE"
		showWarning = true
	case daysRemaining <= 30:
		// 16-30 days - Yellow
		colorCode = "\033[33m" // Yellow
		resetCode = "\033[0m"
		icon = "💡"
		urgencyLevel = "INFO"
		showWarning = true
	}

	if !showWarning {
		return
	}

	// Print warning message
	fmt.Println()
	fmt.Printf("%s========================================%s\n", colorCode, resetCode)
	fmt.Printf("%s%s LICENSE %s%s\n", colorCode, icon, urgencyLevel, resetCode)
	fmt.Printf("%s========================================%s\n", colorCode, resetCode)

	if daysRemaining <= 0 {
		fmt.Printf("%sYour ISX Pulse license has EXPIRED!%s\n", colorCode, resetCode)
		expiredDays := int(time.Since(licenseInfo.ExpiryDate).Hours() / 24)
		fmt.Printf("Expired %d days ago on %s\n", expiredDays, licenseInfo.ExpiryDate.Format("January 2, 2006"))
		fmt.Println("🔒 Some features may be restricted.")
	} else {
		fmt.Printf("%sYour ISX Pulse license expires in %d day(s)%s\n", colorCode, daysRemaining, resetCode)
		fmt.Printf("Expiry Date: %s\n", licenseInfo.ExpiryDate.Format("January 2, 2006"))

		// Add specific messages based on time remaining
		switch {
		case daysRemaining == 1:
			fmt.Println("⏰ License expires TOMORROW!")
		case daysRemaining <= 3:
			fmt.Println("⏰ License expires very soon!")
		case daysRemaining <= 7:
			fmt.Println("📅 Please renew your license this week.")
		case daysRemaining <= 15:
			fmt.Println("📅 Consider renewing your license soon.")
		default:
			fmt.Println("📅 Renewal recommended.")
		}
	}

	fmt.Println()
	fmt.Println("For license renewal or support:")
	fmt.Println("• Visit: https://github.com/haideralmesaody/ISXDailyReportScrapper")
	fmt.Println("• Contact: License Support Team")
	fmt.Printf("%s========================================%s\n", colorCode, resetCode)
	fmt.Println()

	// Log the warning for record keeping
	a.Logger.Warn("License expiry warning displayed",
		slog.String("urgency_level", urgencyLevel),
		slog.Int("days_remaining", daysRemaining),
		slog.String("expiry_date", licenseInfo.ExpiryDate.Format("2006-01-02")),
		slog.String("license_status", status))
}
