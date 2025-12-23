package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.28.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ServiceName    = "isx-daily-reports-scrapper"
	ServiceVersion = "enhanced-v3.0.0"
	MeterName      = "isxcli"
)

// OTelConfig holds OpenTelemetry configuration
type OTelConfig struct {
	ServiceName     string
	ServiceVersion  string
	Environment     string
	TraceExporter   string // "stdout", "otlp", "none"
	MetricExporter  string // "prometheus", "stdout", "none"
	EnableMetrics   bool
	EnableTracing   bool
	SampleRatio     float64
	PrometheusPort  string
}

// OTelProviders holds the OpenTelemetry providers
type OTelProviders struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Tracer         trace.Tracer
	Meter          metric.Meter
	PrometheusHTTP http.Handler
	Logger         *slog.Logger
}

// DefaultOTelConfig returns a default OpenTelemetry configuration
func DefaultOTelConfig() *OTelConfig {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return &OTelConfig{
		ServiceName:     ServiceName,
		ServiceVersion:  ServiceVersion,
		Environment:     env,
		TraceExporter:   "stdout", // Use stdout for development
		MetricExporter:  "prometheus",
		EnableMetrics:   true,
		EnableTracing:   true,
		SampleRatio:     1.0, // Sample all traces in development
		PrometheusPort:  "9090",
	}
}

// InitializeOTel initializes OpenTelemetry with comprehensive observability
func InitializeOTel(cfg *OTelConfig, logger *slog.Logger) (*OTelProviders, error) {
	if cfg == nil {
		cfg = DefaultOTelConfig()
	}

	ctx := context.Background()
	
	logger.InfoContext(ctx, "Initializing OpenTelemetry",
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.ServiceVersion),
		slog.String("environment", cfg.Environment),
		slog.Bool("tracing_enabled", cfg.EnableTracing),
		slog.Bool("metrics_enabled", cfg.EnableMetrics))

	// Create resource
	res, err := createResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	providers := &OTelProviders{
		Logger: logger,
	}

	// Initialize tracing
	if cfg.EnableTracing {
		if err := initializeTracing(ctx, cfg, res, providers); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	// Initialize metrics
	if cfg.EnableMetrics {
		if err := initializeMetrics(ctx, cfg, res, providers); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	// Set up global propagators for trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.InfoContext(ctx, "OpenTelemetry initialization complete",
		slog.Bool("tracing_enabled", cfg.EnableTracing),
		slog.Bool("metrics_enabled", cfg.EnableMetrics))

	return providers, nil
}

// createResource creates the OpenTelemetry resource
func createResource(cfg *OTelConfig) (*resource.Resource, error) {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.DeploymentEnvironmentName(cfg.Environment),
		attribute.String("service.instance.id", generateInstanceID()),
	), nil
}

// initializeTracing sets up OpenTelemetry tracing
func initializeTracing(ctx context.Context, cfg *OTelConfig, res *resource.Resource, providers *OTelProviders) error {
	var exporter sdktrace.SpanExporter
	var err error

	switch cfg.TraceExporter {
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	case "none":
		// No exporter - tracing disabled
		return nil
	default:
		return fmt.Errorf("unsupported trace exporter: %s", cfg.TraceExporter)
	}

	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRatio)),
	)

	providers.TracerProvider = tp
	providers.Tracer = tp.Tracer(MeterName, trace.WithInstrumentationVersion(cfg.ServiceVersion))

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	providers.Logger.InfoContext(ctx, "Tracing initialized",
		slog.String("exporter", cfg.TraceExporter),
		slog.Float64("sample_ratio", cfg.SampleRatio))

	return nil
}

// initializeMetrics sets up OpenTelemetry metrics
func initializeMetrics(ctx context.Context, cfg *OTelConfig, res *resource.Resource, providers *OTelProviders) error {
	switch cfg.MetricExporter {
	case "prometheus":
		// Create Prometheus exporter
		exporter, err := prometheus.New()
		if err != nil {
			return fmt.Errorf("failed to create prometheus exporter: %w", err)
		}
		
		// Create Prometheus HTTP handler
		providers.PrometheusHTTP = promhttp.Handler()
		
		// Create meter provider with Prometheus reader
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(exporter),
		)
		
		providers.MeterProvider = mp
		providers.Meter = mp.Meter(MeterName, metric.WithInstrumentationVersion(cfg.ServiceVersion))

		// Set global meter provider
		otel.SetMeterProvider(mp)
		
	case "none":
		// No exporter - metrics disabled
		return nil
	default:
		return fmt.Errorf("unsupported metric exporter: %s", cfg.MetricExporter)
	}

	providers.Logger.InfoContext(ctx, "Metrics initialized",
		slog.String("exporter", cfg.MetricExporter))

	return nil
}

// CreateBusinessMetrics creates application-specific metrics
func CreateBusinessMetrics(meter metric.Meter) (*BusinessMetrics, error) {
	// HTTP metrics
	httpRequestsTotal, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	httpActiveRequests, err := meter.Int64UpDownCounter(
		"http_active_requests",
		metric.WithDescription("Number of active HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	// Operations metrics
	operationExecutionsTotal, err := meter.Int64Counter(
		"operation_executions_total",
		metric.WithDescription("Total number of operation executions"),
	)
	if err != nil {
		return nil, err
	}

	operationExecutionDuration, err := meter.Float64Histogram(
		"operation_execution_duration_seconds",
		metric.WithDescription("Operation execution duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	operationStepsTotal, err := meter.Int64Counter(
		"operation_steps_total",
		metric.WithDescription("Total number of operation steps executed"),
	)
	if err != nil {
		return nil, err
	}

	operationStepDuration, err := meter.Float64Histogram(
		"operation_step_duration_seconds",
		metric.WithDescription("Operation step execution duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	operationActiveOperations, err := meter.Int64UpDownCounter(
		"operation_active_operations",
		metric.WithDescription("Number of active operations"),
	)
	if err != nil {
		return nil, err
	}

	operationErrors, err := meter.Int64Counter(
		"operation_errors_total",
		metric.WithDescription("Total number of operation errors"),
	)
	if err != nil {
		return nil, err
	}

	operationCancellations, err := meter.Int64Counter(
		"operation_cancellations_total",
		metric.WithDescription("Total number of operation cancellations"),
	)
	if err != nil {
		return nil, err
	}

	operationDataProcessed, err := meter.Int64Counter(
		"operation_data_processed_bytes",
		metric.WithDescription("Total bytes of data processed by operations"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// License metrics
	licenseActivationAttempts, err := meter.Int64Counter(
		"license_activation_attempts_total",
		metric.WithDescription("Total number of license activation attempts"),
	)
	if err != nil {
		return nil, err
	}

	licenseActivationSuccess, err := meter.Int64Counter(
		"license_activation_success_total",
		metric.WithDescription("Total number of successful license activations"),
	)
	if err != nil {
		return nil, err
	}

	licenseValidationChecks, err := meter.Int64Counter(
		"license_validation_checks_total",
		metric.WithDescription("Total number of license validation checks"),
	)
	if err != nil {
		return nil, err
	}

	licenseValidationFailures, err := meter.Int64Counter(
		"license_validation_failures_total",
		metric.WithDescription("Total number of license validation failures"),
	)
	if err != nil {
		return nil, err
	}

	licenseActivationDuration, err := meter.Float64Histogram(
		"license_activation_duration_seconds",
		metric.WithDescription("License activation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	licenseValidationDuration, err := meter.Float64Histogram(
		"license_validation_duration_seconds",
		metric.WithDescription("License validation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	licenseCacheHits, err := meter.Int64Counter(
		"license_cache_hits_total",
		metric.WithDescription("Total number of license cache hits"),
	)
	if err != nil {
		return nil, err
	}

	licenseCacheMisses, err := meter.Int64Counter(
		"license_cache_misses_total",
		metric.WithDescription("Total number of license cache misses"),
	)
	if err != nil {
		return nil, err
	}

	licenseSecurityEvents, err := meter.Int64Counter(
		"license_security_events_total",
		metric.WithDescription("Total number of license security events"),
	)
	if err != nil {
		return nil, err
	}

	// System metrics
	systemErrors, err := meter.Int64Counter(
		"system_errors_total",
		metric.WithDescription("Total number of system errors"),
	)
	if err != nil {
		return nil, err
	}

	systemUptime, err := meter.Float64UpDownCounter(
		"system_uptime_seconds",
		metric.WithDescription("System uptime in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &BusinessMetrics{
		// HTTP metrics
		HTTPRequestsTotal:    httpRequestsTotal,
		HTTPRequestDuration:  httpRequestDuration,
		HTTPActiveRequests:   httpActiveRequests,

		// Operations metrics
		OperationExecutionsTotal:   operationExecutionsTotal,
		OperationExecutionDuration: operationExecutionDuration,
		OperationStepsTotal:        operationStepsTotal,
		OperationStepDuration:      operationStepDuration,
		OperationActiveOperations:  operationActiveOperations,
		OperationErrors:            operationErrors,
		OperationCancellations:     operationCancellations,
		OperationDataProcessed:     operationDataProcessed,
		
		// For backward compatibility, also assign to old field names
		OperationStageExecutions:   operationStepsTotal,

		// License metrics
		LicenseActivationAttempts: licenseActivationAttempts,
		LicenseActivationSuccess:  licenseActivationSuccess,
		LicenseValidationChecks:   licenseValidationChecks,
		LicenseValidationFailures: licenseValidationFailures,
		LicenseActivationDuration: licenseActivationDuration,
		LicenseValidationDuration: licenseValidationDuration,
		LicenseCacheHits:         licenseCacheHits,
		LicenseCacheMisses:       licenseCacheMisses,
		LicenseSecurityEvents:    licenseSecurityEvents,

		// System metrics
		SystemErrors: systemErrors,
		SystemUptime: systemUptime,
	}, nil
}

// BusinessMetrics holds all application-specific metrics
type BusinessMetrics struct {
	// HTTP metrics
	HTTPRequestsTotal    metric.Int64Counter
	HTTPRequestDuration  metric.Float64Histogram
	HTTPActiveRequests   metric.Int64UpDownCounter

	// Operations metrics
	OperationExecutionsTotal   metric.Int64Counter
	OperationExecutionDuration metric.Float64Histogram
	OperationStepsTotal        metric.Int64Counter
	OperationStepDuration      metric.Float64Histogram
	OperationActiveOperations  metric.Int64UpDownCounter
	OperationErrors            metric.Int64Counter
	OperationCancellations     metric.Int64Counter
	OperationDataProcessed     metric.Int64Counter
	
	// Backward compatibility - will be removed later
	OperationStageExecutions   metric.Int64Counter

	// License metrics
	LicenseActivationAttempts metric.Int64Counter
	LicenseActivationSuccess  metric.Int64Counter
	LicenseValidationChecks   metric.Int64Counter
	LicenseValidationFailures metric.Int64Counter
	LicenseActivationDuration metric.Float64Histogram
	LicenseValidationDuration metric.Float64Histogram
	LicenseCacheHits         metric.Int64Counter
	LicenseCacheMisses       metric.Int64Counter
	LicenseSecurityEvents    metric.Int64Counter

	// System metrics
	SystemErrors metric.Int64Counter
	SystemUptime metric.Float64UpDownCounter
}

// Shutdown gracefully shuts down OpenTelemetry providers
func (p *OTelProviders) Shutdown(ctx context.Context) error {
	var errs []error

	if p.TracerProvider != nil {
		if err := p.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}

	if p.MeterProvider != nil {
		if err := p.MeterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider shutdown: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("opentelemetry shutdown errors: %v", errs)
	}

	p.Logger.InfoContext(ctx, "OpenTelemetry shutdown complete")
	return nil
}

// generateInstanceID generates a unique instance identifier
func generateInstanceID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
}

// TraceIDFromContext extracts trace ID from context for logging correlation
func TraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// SpanFromContext returns the current span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span with structured attributes
func AddSpanEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := make([]attribute.KeyValue, 0, len(attributes))
	for k, v := range attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}

	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error, options ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.RecordError(err, options...)
	span.SetStatus(codes.Error, err.Error())
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(ctx context.Context, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	for k, v := range attributes {
		switch val := v.(type) {
		case string:
			span.SetAttributes(attribute.String(k, val))
		case int:
			span.SetAttributes(attribute.Int(k, val))
		case int64:
			span.SetAttributes(attribute.Int64(k, val))
		case float64:
			span.SetAttributes(attribute.Float64(k, val))
		case bool:
			span.SetAttributes(attribute.Bool(k, val))
		default:
			span.SetAttributes(attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
}

// RecordOperationMetrics records metrics for operation execution
func RecordOperationMetrics(ctx context.Context, metrics *BusinessMetrics, operationID string, operationType string, duration time.Duration, success bool, err error) {
	if metrics == nil {
		return
	}

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation.id", operationID),
		attribute.String("operation.type", operationType),
	}

	// Record execution
	metrics.OperationExecutionsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record duration
	statusAttr := attribute.String("status", "success")
	if !success {
		statusAttr = attribute.String("status", "failure")
	}
	durationAttrs := append(attrs, statusAttr)
	metrics.OperationExecutionDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(durationAttrs...))

	// Record errors
	if err != nil {
		errorAttrs := append(attrs, attribute.String("error.type", fmt.Sprintf("%T", err)))
		metrics.OperationErrors.Add(ctx, 1, metric.WithAttributes(errorAttrs...))
	}

	// Add span event
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent("operation.metrics_recorded",
			trace.WithAttributes(
				attribute.String("operation.id", operationID),
				attribute.Bool("success", success),
				attribute.Float64("duration_seconds", duration.Seconds()),
			),
		)
	}
}

// RecordOperationStepMetrics records metrics for operation step execution
func RecordOperationStepMetrics(ctx context.Context, metrics *BusinessMetrics, operationID, stepID, stepType string, duration time.Duration, success bool) {
	if metrics == nil {
		return
	}

	// Common attributes
	attrs := []attribute.KeyValue{
		attribute.String("operation.id", operationID),
		attribute.String("step.id", stepID),
		attribute.String("step.type", stepType),
	}

	// Record step execution
	metrics.OperationStepsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

	// Record duration
	statusAttr := attribute.String("status", "success")
	if !success {
		statusAttr = attribute.String("status", "failure")
	}
	durationAttrs := append(attrs, statusAttr)
	metrics.OperationStepDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(durationAttrs...))
}

// RecordActiveOperationChange records changes in active operation count
func RecordActiveOperationChange(ctx context.Context, metrics *BusinessMetrics, delta int64, operationType string) {
	if metrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation.type", operationType),
	}

	metrics.OperationActiveOperations.Add(ctx, delta, metric.WithAttributes(attrs...))
}

// RecordOperationCancellation records an operation cancellation
func RecordOperationCancellation(ctx context.Context, metrics *BusinessMetrics, operationID, operationType, reason string) {
	if metrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation.id", operationID),
		attribute.String("operation.type", operationType),
		attribute.String("reason", reason),
	}

	metrics.OperationCancellations.Add(ctx, 1, metric.WithAttributes(attrs...))
}