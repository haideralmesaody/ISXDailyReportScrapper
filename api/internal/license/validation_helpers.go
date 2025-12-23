package license

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// performValidation contains the actual validation logic - simplified to only check expiry and periodic remote checks.
func (m *Manager) performValidation() (bool, error) {
	license, err := m.loadLicenseLocal()
	if err != nil {
		m.logDebug(context.Background(), "license_validation", "No local license found during validation",
			slog.String("error", err.Error()),
		)
		return false, fmt.Errorf("no local license found: %w", err)
	}

	if time.Now().After(license.ExpiryDate) {
		license.Status = "expired"
		m.saveLicenseLocal(license)

		m.logLicenseAction(context.Background(), slog.LevelWarn, "license_validation", "License expired",
			license.LicenseKey, license.UserEmail,
			slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
			slog.String("expiry_date", license.ExpiryDate.Format("2006-01-02")),
		)

		return false, fmt.Errorf("license expired on %s", license.ExpiryDate.Format("2006-01-02"))
	}

	if license.DeviceFingerprint != "" {
		currentFingerprint, err := m.hardwareFingerprinter.GenerateHardwareFingerprint()
		if err != nil {
			m.logWarn(context.Background(), "license_validation", "Failed to generate current hardware fingerprint",
				slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
				slog.String("error", err.Error()),
			)
		} else if currentFingerprint.Fingerprint != license.DeviceFingerprint {
			m.logWarn(context.Background(), "license_validation", "Hardware fingerprint mismatch detected",
				slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
				slog.String("stored_fingerprint", license.DeviceFingerprint[:min(16, len(license.DeviceFingerprint))]),
				slog.String("current_fingerprint", currentFingerprint.Fingerprint[:min(16, len(currentFingerprint.Fingerprint))]),
				slog.Int("current_components", currentFingerprint.Components),
			)
		}
	}

	if time.Since(license.LastChecked) > 6*time.Hour {
		if err := m.validateWithAppsScript(&license); err != nil {
			offlineDuration := time.Since(license.LastChecked)
			level := slog.LevelWarn
			message := "Remote validation failed, using local cache"
			if offlineDuration > 48*time.Hour {
				level = slog.LevelError
				message = "Remote validation failed (beyond grace period) - continuing with cached license"
			}

			m.logLicenseAction(context.Background(), level, "license_validation", message,
				license.LicenseKey, license.UserEmail,
				slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
				slog.String("error", err.Error()),
				slog.Duration("offline_duration", offlineDuration),
			)

			// Refresh last checked to avoid immediate repeated remote attempts and allow offline operation.
			license.LastChecked = time.Now()
			_ = m.saveLicenseLocal(license)
		}
	}

	return true, nil
}

// cacheValidationResult caches validation results with appropriate durations.
func (m *Manager) cacheValidationResult(isValid bool, err error) {
	if m.validationCache == nil {
		m.validationCache = NewValidationCache()
	}

	m.validationCache.StoreResult(isValid, sanitizeValidationError(err))
}

// sanitizeValidationError converts missing/corrupted license errors to nil for caller friendliness.
func sanitizeValidationError(err error) error {
	if err == nil {
		return nil
	}
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.Is(err, os.ErrNotExist) || errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
		return nil
	}
	if strings.Contains(err.Error(), "corrupted license file: missing license key") {
		return nil
	}
	return err
}

// invalidateValidationCache clears the validation cache so future checks hit disk again.
func (m *Manager) invalidateValidationCache(reason string) {
	if m.validationCache == nil {
		return
	}

	previous, ok := m.validationCache.Invalidate()
	if !ok {
		return
	}

	m.logInfo(context.Background(), "validation_cache_invalidated", "Validation cache cleared",
		slog.String("reason", reason),
		slog.Bool("had_previous", previous != nil),
	)
}

// GetCachedValidationResult returns the cached validation result if available.
func (m *Manager) GetCachedValidationResult() (*ValidationResult, error) {
	if m.validationCache == nil {
		return nil, fmt.Errorf("validation cache not initialized")
	}

	result, ok := m.validationCache.Snapshot()
	if !ok {
		return nil, fmt.Errorf("no validation performed yet")
	}

	return result, nil
}
