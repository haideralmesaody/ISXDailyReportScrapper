package license

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"github.com/isxcli/isxcli/internal/infrastructure"
)

// logOperation logs operation start/end with performance metrics and OpenTelemetry integration
func (m *Manager) logOperation(ctx context.Context, operation string, start time.Time, err error) {
	logger := infrastructure.LoggerWithContext(ctx)
	duration := time.Since(start)

	// Update performance metrics
	m.recordPerformanceMetric(operation, duration, err == nil)

	// Get trace information for correlation
	traceID := infrastructure.TraceIDFromContext(ctx)
	span := trace.SpanFromContext(ctx)

	// Add operation attributes to span if present
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("license.operation", operation),
			attribute.Float64("license.duration_ms", float64(duration.Milliseconds())),
			attribute.Bool("license.success", err == nil),
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "Operation completed successfully")
		}
	}

	// Log operation completion with enhanced attributes
	attrs := []slog.Attr{
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("trace_id", traceID),
		slog.String("component", "license_manager"),
	}

	if err != nil {
		attrs = append(attrs,
			slog.String("error", err.Error()),
			slog.String("error_type", "license_operation_error"),
		)
		logger.LogAttrs(ctx, slog.LevelError, "License operation failed", attrs...)
	} else {
		logger.LogAttrs(ctx, slog.LevelInfo, "License operation completed successfully", attrs...)
	}
}

// logAction logs a specific action with structured data and OpenTelemetry correlation
func (m *Manager) logAction(ctx context.Context, level slog.Level, action, result string, attrs ...slog.Attr) {
	logger := infrastructure.LoggerWithContext(ctx)
	traceID := infrastructure.TraceIDFromContext(ctx)
	span := trace.SpanFromContext(ctx)

	// Add span event for license action
	if span.IsRecording() {
		infrastructure.AddSpanEvent(ctx, "license."+action, map[string]interface{}{
			"action":    action,
			"result":    result,
			"component": "license_manager",
		})
	}

	// Add standard attributes with enhanced observability
	allAttrs := []slog.Attr{
		slog.String("component", "license_manager"),
		slog.String("action", action),
		slog.String("result", result),
		slog.String("trace_id", traceID),
		slog.String("service_name", "isx-daily-reports-scrapper"),
	}
	allAttrs = append(allAttrs, attrs...)

	logger.LogAttrs(ctx, level, result, allAttrs...)
}

// logLicenseAction logs license-specific actions with enhanced security and observability
func (m *Manager) logLicenseAction(ctx context.Context, level slog.Level, action, result string, licenseKey, userEmail string, attrs ...slog.Attr) {
	// Mask license key for security
	maskedKey := maskLicenseKey(licenseKey)
	maskedEmail := maskEmail(userEmail)
	span := trace.SpanFromContext(ctx)

	// Add license-specific span attributes (without sensitive data)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.String("license.action", action),
			attribute.String("license.result", result),
			attribute.String("license.key_prefix", maskedKey),
			attribute.Bool("license.has_email", userEmail != ""),
			attribute.String("license.operation_category", getLicenseOperationCategory(action)),
		)

		// Record security audit event
		infrastructure.AddSpanEvent(ctx, "license.audit", map[string]interface{}{
			"action":           action,
			"result":           result,
			"license_key_hash": hashLicenseKey(licenseKey),
			"security_level":   "license_operation",
		})
	}

	// Add license-specific attributes with enhanced security
	licenseAttrs := []slog.Attr{
		slog.String("license_key_masked", maskedKey),
		slog.String("license_key_hash", hashLicenseKey(licenseKey)),
		slog.String("user_email_masked", maskedEmail),
		slog.String("license_operation_category", getLicenseOperationCategory(action)),
		slog.String("audit_category", "license_security"),
	}
	licenseAttrs = append(licenseAttrs, attrs...)

	m.logAction(ctx, level, action, result, licenseAttrs...)
}

// maskLicenseKey masks the license key for security
func maskLicenseKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// maskEmail masks email address for security while preserving domain for analytics
func maskEmail(email string) string {
	if email == "" {
		return ""
	}

	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return "****"
	}

	username := email[:atIndex]
	domain := email[atIndex:]

	if len(username) <= 2 {
		return "**" + domain
	}

	return username[:1] + "****" + username[len(username)-1:] + domain
}

// hashLicenseKey creates a secure hash of the license key for audit trails
func hashLicenseKey(key string) string {
	if key == "" {
		return ""
	}
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h)[:16] // First 16 chars of hash for audit correlation
}

// getLicenseOperationCategory categorizes license operations for metrics
func getLicenseOperationCategory(action string) string {
	switch {
	case strings.Contains(action, "activation"):
		return "activation"
	case strings.Contains(action, "validation"):
		return "validation"
	case strings.Contains(action, "transfer"):
		return "transfer"
	case strings.Contains(action, "renewal"):
		return "renewal"
	case strings.Contains(action, "cache"):
		return "cache"
	default:
		return "other"
	}
}

// Helper methods for specific log levels
func (m *Manager) logDebug(ctx context.Context, action, result string, attrs ...slog.Attr) {
	m.logAction(ctx, slog.LevelDebug, action, result, attrs...)
}

func (m *Manager) logInfo(ctx context.Context, action, result string, attrs ...slog.Attr) {
	m.logAction(ctx, slog.LevelInfo, action, result, attrs...)
}

func (m *Manager) logWarn(ctx context.Context, action, result string, attrs ...slog.Attr) {
	m.logAction(ctx, slog.LevelWarn, action, result, attrs...)
}

func (m *Manager) logError(ctx context.Context, action, result string, attrs ...slog.Attr) {
	m.logAction(ctx, slog.LevelError, action, result, attrs...)
}
