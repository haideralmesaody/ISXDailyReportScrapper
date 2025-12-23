package license

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"github.com/isxcli/isxcli/internal/infrastructure"
)

const (
	TracerName = "license-manager"
	MeterName  = "license-manager"
)

// LicenseMetrics holds all license-specific OpenTelemetry metrics
type LicenseMetrics struct {
	// Activation metrics
	ActivationAttempts metric.Int64Counter
	ActivationSuccess  metric.Int64Counter
	ActivationFailures metric.Int64Counter
	ActivationDuration metric.Float64Histogram

	// Validation metrics
	ValidationAttempts    metric.Int64Counter
	ValidationSuccess     metric.Int64Counter
	ValidationFailures    metric.Int64Counter
	ValidationDuration    metric.Float64Histogram
	ValidationCacheHits   metric.Int64Counter
	ValidationCacheMisses metric.Int64Counter

	// Transfer metrics
	TransferAttempts metric.Int64Counter
	TransferSuccess  metric.Int64Counter
	TransferFailures metric.Int64Counter
	TransferDuration metric.Float64Histogram

	// Google Sheets connectivity metrics
	SheetsRequests     metric.Int64Counter
	SheetsSuccess      metric.Int64Counter
	SheetsFailures     metric.Int64Counter
	SheetsDuration     metric.Float64Histogram
	SheetsConnectivity metric.Int64UpDownCounter

	// Security metrics
	SecurityEvents     metric.Int64Counter
	RateLimitHits      metric.Int64Counter
	InvalidKeyAttempts metric.Int64Counter

	// Performance metrics
	LicenseDataSize      metric.Int64Histogram
	NetworkLatency       metric.Float64Histogram
	CacheEvictions       metric.Int64Counter
	ConcurrentOperations metric.Int64UpDownCounter

	// Scratch Card specific metrics
	ScratchCardActivations    metric.Int64Counter
	ScratchCardSuccessRate    metric.Float64Histogram
	ScratchCardFailuresByType metric.Int64Counter
	ScratchCardValidation     metric.Float64Histogram

	// Apps Script specific metrics
	AppsScriptRequests     metric.Int64Counter
	AppsScriptResponseTime metric.Float64Histogram
	AppsScriptErrors       metric.Int64Counter
	AppsScriptConnectivity metric.Int64UpDownCounter
	AppsScriptRateLimits   metric.Int64Counter

	// Device fingerprint metrics
	FingerprintGeneration metric.Float64Histogram
	FingerprintMismatches metric.Int64Counter
	FingerprintValidation metric.Int64Counter

	// Card batch tracking
	BatchActivations    metric.Int64Counter
	BatchProcessingTime metric.Float64Histogram
	BatchFailureRate    metric.Float64Histogram
	ActivePendingCards  metric.Int64UpDownCounter
}

// InitializeLicenseMetrics creates all license-specific metrics
func InitializeLicenseMetrics(meter metric.Meter) (*LicenseMetrics, error) {
	metrics := &LicenseMetrics{}

	var err error

	// Activation metrics
	metrics.ActivationAttempts, err = meter.Int64Counter(
		"license_activation_attempts_total",
		metric.WithDescription("Total number of license activation attempts"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create activation attempts counter: %w", err)
	}

	metrics.ActivationSuccess, err = meter.Int64Counter(
		"license_activation_success_total",
		metric.WithDescription("Total number of successful license activations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create activation success counter: %w", err)
	}

	metrics.ActivationFailures, err = meter.Int64Counter(
		"license_activation_failures_total",
		metric.WithDescription("Total number of failed license activations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create activation failures counter: %w", err)
	}

	metrics.ActivationDuration, err = meter.Float64Histogram(
		"license_activation_duration_seconds",
		metric.WithDescription("License activation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create activation duration histogram: %w", err)
	}

	// Validation metrics
	metrics.ValidationAttempts, err = meter.Int64Counter(
		"license_validation_attempts_total",
		metric.WithDescription("Total number of license validation attempts"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation attempts counter: %w", err)
	}

	metrics.ValidationSuccess, err = meter.Int64Counter(
		"license_validation_success_total",
		metric.WithDescription("Total number of successful license validations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation success counter: %w", err)
	}

	metrics.ValidationFailures, err = meter.Int64Counter(
		"license_validation_failures_total",
		metric.WithDescription("Total number of failed license validations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation failures counter: %w", err)
	}

	metrics.ValidationDuration, err = meter.Float64Histogram(
		"license_validation_duration_seconds",
		metric.WithDescription("License validation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation duration histogram: %w", err)
	}

	metrics.ValidationCacheHits, err = meter.Int64Counter(
		"license_validation_cache_hits_total",
		metric.WithDescription("Total number of license validation cache hits"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation cache hits counter: %w", err)
	}

	metrics.ValidationCacheMisses, err = meter.Int64Counter(
		"license_validation_cache_misses_total",
		metric.WithDescription("Total number of license validation cache misses"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation cache misses counter: %w", err)
	}

	// Transfer metrics
	metrics.TransferAttempts, err = meter.Int64Counter(
		"license_transfer_attempts_total",
		metric.WithDescription("Total number of license transfer attempts"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer attempts counter: %w", err)
	}

	metrics.TransferSuccess, err = meter.Int64Counter(
		"license_transfer_success_total",
		metric.WithDescription("Total number of successful license transfers"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer success counter: %w", err)
	}

	metrics.TransferFailures, err = meter.Int64Counter(
		"license_transfer_failures_total",
		metric.WithDescription("Total number of failed license transfers"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer failures counter: %w", err)
	}

	metrics.TransferDuration, err = meter.Float64Histogram(
		"license_transfer_duration_seconds",
		metric.WithDescription("License transfer duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer duration histogram: %w", err)
	}

	// Google Sheets connectivity metrics
	metrics.SheetsRequests, err = meter.Int64Counter(
		"license_sheets_requests_total",
		metric.WithDescription("Total number of Google Sheets API requests"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets requests counter: %w", err)
	}

	metrics.SheetsSuccess, err = meter.Int64Counter(
		"license_sheets_success_total",
		metric.WithDescription("Total number of successful Google Sheets API requests"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets success counter: %w", err)
	}

	metrics.SheetsFailures, err = meter.Int64Counter(
		"license_sheets_failures_total",
		metric.WithDescription("Total number of failed Google Sheets API requests"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets failures counter: %w", err)
	}

	metrics.SheetsDuration, err = meter.Float64Histogram(
		"license_sheets_duration_seconds",
		metric.WithDescription("Google Sheets API request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets duration histogram: %w", err)
	}

	metrics.SheetsConnectivity, err = meter.Int64UpDownCounter(
		"license_sheets_connectivity",
		metric.WithDescription("Google Sheets connectivity status (1=connected, 0=disconnected)"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets connectivity gauge: %w", err)
	}

	// Security metrics
	metrics.SecurityEvents, err = meter.Int64Counter(
		"license_security_events_total",
		metric.WithDescription("Total number of license security events"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create security events counter: %w", err)
	}

	metrics.RateLimitHits, err = meter.Int64Counter(
		"license_rate_limit_hits_total",
		metric.WithDescription("Total number of rate limit hits"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limit hits counter: %w", err)
	}

	metrics.InvalidKeyAttempts, err = meter.Int64Counter(
		"license_invalid_key_attempts_total",
		metric.WithDescription("Total number of invalid license key attempts"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create invalid key attempts counter: %w", err)
	}

	// Performance metrics
	metrics.LicenseDataSize, err = meter.Int64Histogram(
		"license_data_size_bytes",
		metric.WithDescription("Size of license data in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create license data size histogram: %w", err)
	}

	metrics.NetworkLatency, err = meter.Float64Histogram(
		"license_network_latency_seconds",
		metric.WithDescription("Network latency for license operations in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create network latency histogram: %w", err)
	}

	metrics.CacheEvictions, err = meter.Int64Counter(
		"license_cache_evictions_total",
		metric.WithDescription("Total number of license cache evictions"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache evictions counter: %w", err)
	}

	metrics.ConcurrentOperations, err = meter.Int64UpDownCounter(
		"license_concurrent_operations",
		metric.WithDescription("Number of concurrent license operations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create concurrent operations gauge: %w", err)
	}

	// Scratch Card specific metrics
	metrics.ScratchCardActivations, err = meter.Int64Counter(
		"scratch_card_activations_total",
		metric.WithDescription("Total number of scratch card activation attempts by type"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scratch card activations counter: %w", err)
	}

	metrics.ScratchCardSuccessRate, err = meter.Float64Histogram(
		"scratch_card_success_rate",
		metric.WithDescription("Scratch card activation success rate by card type"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scratch card success rate histogram: %w", err)
	}

	metrics.ScratchCardFailuresByType, err = meter.Int64Counter(
		"scratch_card_failures_by_type_total",
		metric.WithDescription("Total number of scratch card failures by failure type"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scratch card failures by type counter: %w", err)
	}

	metrics.ScratchCardValidation, err = meter.Float64Histogram(
		"scratch_card_validation_duration_seconds",
		metric.WithDescription("Scratch card validation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scratch card validation histogram: %w", err)
	}

	// Apps Script specific metrics
	metrics.AppsScriptRequests, err = meter.Int64Counter(
		"apps_script_requests_total",
		metric.WithDescription("Total number of Apps Script API requests"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Apps Script requests counter: %w", err)
	}

	metrics.AppsScriptResponseTime, err = meter.Float64Histogram(
		"apps_script_response_time_seconds",
		metric.WithDescription("Apps Script API response time histogram"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Apps Script response time histogram: %w", err)
	}

	metrics.AppsScriptErrors, err = meter.Int64Counter(
		"apps_script_errors_total",
		metric.WithDescription("Total number of Apps Script API errors"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Apps Script errors counter: %w", err)
	}

	metrics.AppsScriptConnectivity, err = meter.Int64UpDownCounter(
		"apps_script_connectivity",
		metric.WithDescription("Apps Script connectivity status (1=connected, 0=disconnected)"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Apps Script connectivity gauge: %w", err)
	}

	metrics.AppsScriptRateLimits, err = meter.Int64Counter(
		"apps_script_rate_limits_total",
		metric.WithDescription("Total number of Apps Script rate limit hits"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Apps Script rate limits counter: %w", err)
	}

	// Device fingerprint metrics
	metrics.FingerprintGeneration, err = meter.Float64Histogram(
		"fingerprint_generation_duration_seconds",
		metric.WithDescription("Device fingerprint generation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fingerprint generation histogram: %w", err)
	}

	metrics.FingerprintMismatches, err = meter.Int64Counter(
		"fingerprint_mismatches_total",
		metric.WithDescription("Total number of device fingerprint mismatches"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fingerprint mismatches counter: %w", err)
	}

	metrics.FingerprintValidation, err = meter.Int64Counter(
		"fingerprint_validations_total",
		metric.WithDescription("Total number of device fingerprint validations"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fingerprint validation counter: %w", err)
	}

	// Card batch tracking
	metrics.BatchActivations, err = meter.Int64Counter(
		"batch_activations_total",
		metric.WithDescription("Total number of batch activation attempts"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch activations counter: %w", err)
	}

	metrics.BatchProcessingTime, err = meter.Float64Histogram(
		"batch_processing_duration_seconds",
		metric.WithDescription("Batch processing duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch processing time histogram: %w", err)
	}

	metrics.BatchFailureRate, err = meter.Float64Histogram(
		"batch_failure_rate",
		metric.WithDescription("Batch failure rate percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch failure rate histogram: %w", err)
	}

	metrics.ActivePendingCards, err = meter.Int64UpDownCounter(
		"active_pending_cards",
		metric.WithDescription("Number of active pending scratch cards"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active pending cards gauge: %w", err)
	}

	return metrics, nil
}

// TraceActivation wraps license activation with OpenTelemetry tracing
func (m *Manager) TraceActivation(ctx context.Context, licenseKey string, fn func() error) error {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.activation",
		trace.WithAttributes(
			attribute.String("license.operation", "activation"),
			attribute.String("license.key_prefix", maskLicenseKey(licenseKey)),
			attribute.String("component", "license_manager"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordActivationMetrics(ctx, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("license.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("license.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		// Add error classification
		errorType := classifyLicenseError(err)
		span.SetAttributes(attribute.String("license.error_type", errorType))
	} else {
		span.SetStatus(codes.Ok, "License activated successfully")

		// Record security audit event
		infrastructure.AddSpanEvent(ctx, "license.activation.success", map[string]interface{}{
			"license_key_hash": hashLicenseKey(licenseKey),
			"audit_category":   "license_security",
		})
	}

	return err
}

// TraceValidation wraps license validation with OpenTelemetry tracing
func (m *Manager) TraceValidation(ctx context.Context, fn func() (bool, error)) (bool, error) {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.validation",
		trace.WithAttributes(
			attribute.String("license.operation", "validation"),
			attribute.String("component", "license_manager"),
		),
	)
	defer span.End()

	start := time.Now()
	valid, err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordValidationMetrics(ctx, duration, valid, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("license.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("license.valid", valid),
		attribute.Bool("license.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		errorType := classifyLicenseError(err)
		span.SetAttributes(attribute.String("license.error_type", errorType))
	} else if !valid {
		span.SetStatus(codes.Error, "License validation failed")
		span.SetAttributes(attribute.String("license.error_type", "invalid_license"))
	} else {
		span.SetStatus(codes.Ok, "License validation successful")
	}

	return valid, err
}

// TraceTransfer wraps license transfer with OpenTelemetry tracing
func (m *Manager) TraceTransfer(ctx context.Context, licenseKey string, force bool, fn func() error) error {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.transfer",
		trace.WithAttributes(
			attribute.String("license.operation", "transfer"),
			attribute.String("license.key_prefix", maskLicenseKey(licenseKey)),
			attribute.Bool("license.force_transfer", force),
			attribute.String("component", "license_manager"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordTransferMetrics(ctx, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("license.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("license.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		errorType := classifyLicenseError(err)
		span.SetAttributes(attribute.String("license.error_type", errorType))
	} else {
		span.SetStatus(codes.Ok, "License transferred successfully")

		// Record security audit event
		infrastructure.AddSpanEvent(ctx, "license.transfer.success", map[string]interface{}{
			"license_key_hash": hashLicenseKey(licenseKey),
			"force_transfer":   force,
			"audit_category":   "license_security",
		})
	}

	return err
}

// TraceSheetsOperation wraps Google Sheets operations with tracing
func (m *Manager) TraceSheetsOperation(ctx context.Context, operation string, fn func() error) error {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.sheets."+operation,
		trace.WithAttributes(
			attribute.String("license.operation", "sheets_"+operation),
			attribute.String("sheets.operation", operation),
			attribute.String("component", "license_manager"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordSheetsMetrics(ctx, operation, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("sheets.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("sheets.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		errorType := classifyNetworkError(err)
		span.SetAttributes(attribute.String("sheets.error_type", errorType))
	} else {
		span.SetStatus(codes.Ok, "Sheets operation completed successfully")
	}

	return err
}

// TraceScratchCardActivation wraps scratch card activation with OpenTelemetry tracing
func (m *Manager) TraceScratchCardActivation(ctx context.Context, cardType, batchID string, fn func() error) error {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.scratch_card.activation",
		trace.WithAttributes(
			attribute.String("license.operation", "scratch_card_activation"),
			attribute.String("scratch_card.type", cardType),
			attribute.String("scratch_card.batch_id", batchID),
			attribute.String("component", "license_manager"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordScratchCardMetrics(ctx, cardType, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("scratch_card.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("scratch_card.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		errorType := classifyScratchCardError(err)
		span.SetAttributes(attribute.String("scratch_card.error_type", errorType))
	} else {
		span.SetStatus(codes.Ok, "Scratch card activation successful")

		// Record security audit event
		infrastructure.AddSpanEvent(ctx, "scratch_card.activation.success", map[string]interface{}{
			"card_type":      cardType,
			"batch_id":       batchID,
			"audit_category": "license_security",
		})
	}

	return err
}

// TraceAppsScriptOperation wraps Apps Script API calls with tracing
func (m *Manager) TraceAppsScriptOperation(ctx context.Context, operation string, fn func() error) error {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.apps_script."+operation,
		trace.WithAttributes(
			attribute.String("license.operation", "apps_script_"+operation),
			attribute.String("apps_script.operation", operation),
			attribute.String("component", "apps_script_client"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordAppsScriptMetrics(ctx, operation, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("apps_script.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("apps_script.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		errorType := classifyAppsScriptError(err)
		span.SetAttributes(attribute.String("apps_script.error_type", errorType))
	} else {
		span.SetStatus(codes.Ok, "Apps Script operation completed successfully")
	}

	return err
}

// TraceFingerprintGeneration wraps device fingerprint generation with tracing
func (m *Manager) TraceFingerprintGeneration(ctx context.Context, fn func() (string, error)) (string, error) {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.fingerprint.generation",
		trace.WithAttributes(
			attribute.String("license.operation", "fingerprint_generation"),
			attribute.String("component", "fingerprint_generator"),
		),
	)
	defer span.End()

	start := time.Now()
	fingerprint, err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordFingerprintMetrics(ctx, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("fingerprint.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("fingerprint.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "Fingerprint generation successful")
		span.SetAttributes(
			attribute.String("fingerprint.hash", hashFingerprint(fingerprint)),
		)
	}

	return fingerprint, err
}

// TraceBatchActivation wraps batch activation operations with tracing
func (m *Manager) TraceBatchActivation(ctx context.Context, batchID string, cardCount int, fn func() error) error {
	tracer := otel.Tracer(TracerName)

	ctx, span := tracer.Start(ctx, "license.batch.activation",
		trace.WithAttributes(
			attribute.String("license.operation", "batch_activation"),
			attribute.String("batch.id", batchID),
			attribute.Int("batch.card_count", cardCount),
			attribute.String("component", "batch_processor"),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Record metrics if available
	if m.metrics != nil {
		m.recordBatchMetrics(ctx, batchID, cardCount, duration, err == nil)
	}

	// Update span with results
	span.SetAttributes(
		attribute.Float64("batch.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("batch.success", err == nil),
		attribute.Float64("batch.throughput_cards_per_second", float64(cardCount)/duration.Seconds()),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		errorType := classifyBatchError(err)
		span.SetAttributes(attribute.String("batch.error_type", errorType))
	} else {
		span.SetStatus(codes.Ok, "Batch activation completed successfully")

		// Record audit event
		infrastructure.AddSpanEvent(ctx, "batch.activation.success", map[string]interface{}{
			"batch_id":       batchID,
			"card_count":     cardCount,
			"throughput":     float64(cardCount) / duration.Seconds(),
			"audit_category": "license_security",
		})
	}

	return err
}

// recordActivationMetrics records activation-specific metrics
func (m *Manager) recordActivationMetrics(ctx context.Context, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", "activation"),
		attribute.String("component", "license_manager"),
	)

	m.metrics.ActivationAttempts.Add(ctx, 1, labels)
	m.metrics.ActivationDuration.Record(ctx, duration.Seconds(), labels)

	if success {
		m.metrics.ActivationSuccess.Add(ctx, 1, labels)
	} else {
		m.metrics.ActivationFailures.Add(ctx, 1, labels)
	}
}

// recordValidationMetrics records validation-specific metrics
func (m *Manager) recordValidationMetrics(ctx context.Context, duration time.Duration, valid bool, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", "validation"),
		attribute.String("component", "license_manager"),
	)

	m.metrics.ValidationAttempts.Add(ctx, 1, labels)
	m.metrics.ValidationDuration.Record(ctx, duration.Seconds(), labels)

	if success && valid {
		m.metrics.ValidationSuccess.Add(ctx, 1, labels)
	} else {
		m.metrics.ValidationFailures.Add(ctx, 1, labels)
	}
}

// recordTransferMetrics records transfer-specific metrics
func (m *Manager) recordTransferMetrics(ctx context.Context, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", "transfer"),
		attribute.String("component", "license_manager"),
	)

	m.metrics.TransferAttempts.Add(ctx, 1, labels)
	m.metrics.TransferDuration.Record(ctx, duration.Seconds(), labels)

	if success {
		m.metrics.TransferSuccess.Add(ctx, 1, labels)
	} else {
		m.metrics.TransferFailures.Add(ctx, 1, labels)
	}
}

// recordSheetsMetrics records Google Sheets operation metrics
func (m *Manager) recordSheetsMetrics(ctx context.Context, operation string, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("component", "google_sheets"),
	)

	m.metrics.SheetsRequests.Add(ctx, 1, labels)
	m.metrics.SheetsDuration.Record(ctx, duration.Seconds(), labels)

	if success {
		m.metrics.SheetsSuccess.Add(ctx, 1, labels)
		m.metrics.SheetsConnectivity.Add(ctx, 1, labels)
	} else {
		m.metrics.SheetsFailures.Add(ctx, 1, labels)
		m.metrics.SheetsConnectivity.Add(ctx, -1, labels)
	}
}

// recordScratchCardMetrics records scratch card operation metrics
func (m *Manager) recordScratchCardMetrics(ctx context.Context, cardType string, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("card_type", cardType),
		attribute.String("component", "scratch_card"),
	)

	m.metrics.ScratchCardActivations.Add(ctx, 1, labels)
	m.metrics.ScratchCardValidation.Record(ctx, duration.Seconds(), labels)

	if success {
		m.metrics.ScratchCardSuccessRate.Record(ctx, 1.0, labels) // 100% success
	} else {
		m.metrics.ScratchCardSuccessRate.Record(ctx, 0.0, labels) // 0% success
		m.metrics.ScratchCardFailuresByType.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("card_type", cardType),
				attribute.String("failure_type", "activation_failed"),
				attribute.String("component", "scratch_card"),
			),
		)
	}
}

// recordAppsScriptMetrics records Apps Script operation metrics
func (m *Manager) recordAppsScriptMetrics(ctx context.Context, operation string, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("component", "apps_script"),
	)

	m.metrics.AppsScriptRequests.Add(ctx, 1, labels)
	m.metrics.AppsScriptResponseTime.Record(ctx, duration.Seconds(), labels)

	if success {
		m.metrics.AppsScriptConnectivity.Add(ctx, 1, labels)
	} else {
		m.metrics.AppsScriptErrors.Add(ctx, 1, labels)
		m.metrics.AppsScriptConnectivity.Add(ctx, -1, labels)
	}
}

// recordFingerprintMetrics records device fingerprint metrics
func (m *Manager) recordFingerprintMetrics(ctx context.Context, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("component", "fingerprint"),
	)

	m.metrics.FingerprintGeneration.Record(ctx, duration.Seconds(), labels)
	m.metrics.FingerprintValidation.Add(ctx, 1, labels)

	if !success {
		m.metrics.FingerprintMismatches.Add(ctx, 1, labels)
	}
}

// recordBatchMetrics records batch processing metrics
func (m *Manager) recordBatchMetrics(ctx context.Context, batchID string, cardCount int, duration time.Duration, success bool) {
	if m.metrics == nil {
		return
	}

	labels := metric.WithAttributes(
		attribute.String("batch_id", batchID),
		attribute.String("component", "batch_processor"),
	)

	m.metrics.BatchActivations.Add(ctx, 1, labels)
	m.metrics.BatchProcessingTime.Record(ctx, duration.Seconds(), labels)

	if success {
		m.metrics.BatchFailureRate.Record(ctx, 0.0, labels)              // 0% failure
		m.metrics.ActivePendingCards.Add(ctx, int64(-cardCount), labels) // Remove processed cards
	} else {
		m.metrics.BatchFailureRate.Record(ctx, 100.0, labels) // 100% failure
	}
}

// classifyLicenseError categorizes license errors for better observability
func classifyLicenseError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "expired"):
		return "license_expired"
	case contains(errStr, "machine_mismatch"):
		return "machine_mismatch"
	case contains(errStr, "not found"):
		return "license_not_found"
	case contains(errStr, "invalid"):
		return "invalid_license"
	case contains(errStr, "network"), contains(errStr, "timeout"):
		return "network_error"
	case contains(errStr, "rate limit"):
		return "rate_limited"
	case contains(errStr, "unauthorized"):
		return "unauthorized"
	default:
		return "unknown_error"
	}
}

// classifyNetworkError categorizes network errors
func classifyNetworkError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "connection refused"):
		return "connection_refused"
	case contains(errStr, "no such host"):
		return "dns_error"
	case contains(errStr, "403"):
		return "forbidden"
	case contains(errStr, "401"):
		return "unauthorized"
	case contains(errStr, "500"):
		return "server_error"
	default:
		return "network_error"
	}
}

// classifyScratchCardError categorizes scratch card specific errors
func classifyScratchCardError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "card_already_activated"):
		return "card_already_activated"
	case contains(errStr, "invalid_card_format"):
		return "invalid_card_format"
	case contains(errStr, "card_expired"):
		return "card_expired"
	case contains(errStr, "batch_not_found"):
		return "batch_not_found"
	case contains(errStr, "device_mismatch"):
		return "device_mismatch"
	case contains(errStr, "fingerprint_validation_failed"):
		return "fingerprint_validation_failed"
	case contains(errStr, "apps_script_error"):
		return "apps_script_error"
	case contains(errStr, "rate_limited"):
		return "rate_limited"
	default:
		return "unknown_scratch_card_error"
	}
}

// classifyAppsScriptError categorizes Apps Script specific errors
func classifyAppsScriptError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "quota_exceeded"):
		return "quota_exceeded"
	case contains(errStr, "script_timeout"):
		return "script_timeout"
	case contains(errStr, "invalid_request"):
		return "invalid_request"
	case contains(errStr, "permission_denied"):
		return "permission_denied"
	case contains(errStr, "service_unavailable"):
		return "service_unavailable"
	case contains(errStr, "rate_limit"):
		return "rate_limit"
	default:
		return "apps_script_error"
	}
}

// classifyBatchError categorizes batch processing errors
func classifyBatchError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "batch_size_exceeded"):
		return "batch_size_exceeded"
	case contains(errStr, "concurrent_batch_limit"):
		return "concurrent_batch_limit"
	case contains(errStr, "batch_timeout"):
		return "batch_timeout"
	case contains(errStr, "partial_failure"):
		return "partial_failure"
	case contains(errStr, "validation_failed"):
		return "validation_failed"
	default:
		return "batch_error"
	}
}

// contains is a helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		containsInner(s, substr))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// hashFingerprint creates a secure hash of the fingerprint for observability
func hashFingerprint(fingerprint string) string {
	if fingerprint == "" {
		return "empty"
	}
	// Return first 8 characters for tracking without exposing full fingerprint
	if len(fingerprint) > 8 {
		return fingerprint[:8] + "..."
	}
	return fingerprint
}
