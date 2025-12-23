package license

import (
	"context"
	"math"
	"time"

	"log/slog"
)

// daysLeftRounded returns whole days until expiry using ceiling for positive durations
// and floor for negative durations to count full days expired.
func daysLeftRounded(expiryDate time.Time) int {
	if expiryDate.IsZero() {
		return 0
	}

	days := time.Until(expiryDate).Hours() / 24
	if days >= 0 {
		return int(math.Ceil(days))
	}

	return int(math.Floor(days))
}

// calculateExpireStatus calculates the expire status based on days remaining.
func (m *Manager) calculateExpireStatus(expiryDate time.Time) string {
	if expiryDate.IsZero() {
		return "Available"
	}

	daysLeft := daysLeftRounded(expiryDate)

	switch {
	case daysLeft <= 0:
		return "Expired"
	case daysLeft <= 7:
		return "Critical"
	case daysLeft <= 30:
		return "Warning"
	default:
		return "Active"
	}
}

// GetLicenseStatus returns detailed license status information with comprehensive observability.
func (m *Manager) GetLicenseStatus() (*LicenseInfo, string, error) {
	ctx := context.Background()
	start := time.Now()

	m.logInfo(ctx, "[VERBOSE] GetLicenseStatus entry", "Starting license status check flow",
		slog.String("operation", "get_license_status"),
		slog.String("license_file_path", m.licenseFile),
		slog.String("working_dir", m.getWorkingDir()),
		slog.Time("check_time", time.Now()),
	)

	m.logDebug(ctx, "loading_local_license", "Loading license from local file",
		slog.String("license_file", m.licenseFile),
	)

	license, err := m.loadLicenseLocal()
	loadLatency := time.Since(start)

	if err != nil {
		m.logDebug(ctx, "no_local_license", "No local license file found",
			slog.Duration("load_latency", loadLatency),
			slog.String("error", err.Error()),
			slog.String("status_result", "not_activated"),
		)
		return nil, "Not Activated", nil
	}

	m.logDebug(ctx, "local_license_loaded", "Successfully loaded local license",
		slog.Duration("load_latency", loadLatency),
		slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
		slog.String("duration", license.Duration),
		slog.Time("expiry_date", license.ExpiryDate),
		slog.String("current_status", license.Status),
	)

	now := time.Now()
	status, daysLeft := calculateLicenseStatus(license, now)

	m.logInfo(ctx, "[VERBOSE] License status calculation", "Starting detailed status calculation",
		slog.Time("current_time", now),
		slog.Time("expiry_date", license.ExpiryDate),
		slog.Int("days_left", daysLeft),
		slog.Bool("is_expired", now.After(license.ExpiryDate)),
		slog.String("time_until_expiry", time.Until(license.ExpiryDate).String()),
		slog.String("stored_status", license.Status),
	)

	switch status {
	case "Expired":
		m.logWarn(ctx, "license_expired", "License has expired",
			slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
			slog.Time("expiry_date", license.ExpiryDate),
			slog.Int("days_expired", -daysLeft),
		)
	case "Critical":
		m.logWarn(ctx, "license_critical", "License expires within 7 days",
			slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
			slog.Int("days_left", daysLeft),
			slog.Time("expiry_date", license.ExpiryDate),
		)
	case "Warning":
		m.logInfo(ctx, "license_warning", "License expires within 30 days",
			slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
			slog.Int("days_left", daysLeft),
			slog.Time("expiry_date", license.ExpiryDate),
		)
	default:
		m.logDebug(ctx, "license_active", "License is active",
			slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
			slog.Int("days_left", daysLeft),
			slog.Time("expiry_date", license.ExpiryDate),
		)
	}

	totalLatency := time.Since(start)
	m.logInfo(ctx, "[VERBOSE] GetLicenseStatus complete", "License status check completed with full details",
		slog.Duration("total_latency", totalLatency),
		slog.String("final_status", status),
		slog.Int("days_left", daysLeft),
		slog.String("license_key_prefix", license.LicenseKey[:min(8, len(license.LicenseKey))]),
		slog.Time("license_issued", license.IssuedDate),
		slog.Time("license_expiry", license.ExpiryDate),
		slog.String("license_duration", license.Duration),
		slog.String("stored_status", license.Status),
		slog.Time("last_checked", license.LastChecked),
		slog.Bool("returning_info", true),
		slog.Bool("returning_error", false),
	)

	return &license, status, nil
}
