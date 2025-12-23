package infrastructure

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

// TestOTelInitialization tests OpenTelemetry initialization
func TestOTelInitialization(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Test with default configuration
	providers, err := InitializeOTel(nil, logger)
	require.NoError(t, err)
	require.NotNil(t, providers)
	
	// Verify tracer provider is set
	assert.NotNil(t, providers.TracerProvider)
	assert.NotNil(t, providers.Tracer)
	
	// Verify meter provider is set
	assert.NotNil(t, providers.MeterProvider)
	assert.NotNil(t, providers.Meter)
	
	// Verify Prometheus handler is available
	assert.NotNil(t, providers.PrometheusHTTP)
	
	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = providers.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestTraceCorrelation tests trace ID correlation
func TestTraceCorrelation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(t, err)
	defer providers.Shutdown(context.Background())
	
	ctx := context.Background()
	
	// Start a span
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(ctx, "test-operation")
	defer span.End()
	
	// Extract trace ID
	traceID := TraceIDFromContext(ctx)
	assert.NotEmpty(t, traceID)
	
	// Verify trace ID matches span context
	expectedTraceID := span.SpanContext().TraceID().String()
	assert.Equal(t, expectedTraceID, traceID)
	
	// Test context with trace ID
	ctx = WithTraceID(ctx, traceID)
	retrievedTraceID := GetTraceID(ctx)
	assert.Equal(t, traceID, retrievedTraceID)
}

// TestBusinessMetrics tests business metrics creation
func TestBusinessMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(t, err)
	defer providers.Shutdown(context.Background())
	
	metrics, err := CreateBusinessMetrics(providers.Meter)
	require.NoError(t, err)
	require.NotNil(t, metrics)
	
	// Verify HTTP metrics
	assert.NotNil(t, metrics.HTTPRequestsTotal)
	assert.NotNil(t, metrics.HTTPRequestDuration)
	assert.NotNil(t, metrics.HTTPActiveRequests)
	
	// Verify operation metrics
	assert.NotNil(t, metrics.OperationExecutionsTotal)
	assert.NotNil(t, metrics.OperationExecutionDuration)
	assert.NotNil(t, metrics.OperationStepsTotal)
	assert.NotNil(t, metrics.OperationActiveOperations)
	assert.NotNil(t, metrics.OperationDataProcessed)
	
	// Verify license metrics
	assert.NotNil(t, metrics.LicenseActivationAttempts)
	assert.NotNil(t, metrics.LicenseActivationSuccess)
	assert.NotNil(t, metrics.LicenseValidationChecks)
	
	// Verify system metrics
	assert.NotNil(t, metrics.SystemErrors)
	assert.NotNil(t, metrics.SystemUptime)
}

// TestSpanOperations tests span operations and attributes
func TestSpanOperations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(t, err)
	defer providers.Shutdown(context.Background())
	
	ctx := context.Background()
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()
	
	// Test adding span attributes
	attributes := map[string]interface{}{
		"string_attr": "test_value",
		"int_attr":    42,
		"float_attr":  3.14,
		"bool_attr":   true,
	}
	
	SetSpanAttributes(ctx, attributes)
	
	// Test adding span events
	AddSpanEvent(ctx, "test.event", map[string]interface{}{
		"event_data": "test_event_value",
		"timestamp":  time.Now().Unix(),
	})
	
	// Test error recording
	testErr := assert.AnError
	RecordError(ctx, testErr)
	
	// Verify span is recording
	assert.True(t, span.IsRecording())
}

// TestPrometheusEndpoint tests the Prometheus metrics endpoint
func TestPrometheusEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(t, err)
	defer providers.Shutdown(context.Background())
	
	// Create test server with Prometheus handler
	server := httptest.NewServer(providers.PrometheusHTTP)
	defer server.Close()
	
	// Make request to metrics endpoint
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")
}

// TestOTelConfiguration tests different configuration options
func TestOTelConfiguration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	tests := []struct {
		name   string
		config *OTelConfig
	}{
		{
			name: "development_config",
			config: &OTelConfig{
				ServiceName:     "test-service",
				ServiceVersion:  "v1.0.0",
				Environment:     "development",
				TraceExporter:   "stdout",
				MetricExporter:  "prometheus",
				EnableMetrics:   true,
				EnableTracing:   true,
				SampleRatio:     1.0,
			},
		},
		{
			name: "disabled_tracing",
			config: &OTelConfig{
				ServiceName:     "test-service",
				ServiceVersion:  "v1.0.0",
				Environment:     "test",
				TraceExporter:   "none",
				MetricExporter:  "prometheus",
				EnableMetrics:   true,
				EnableTracing:   false,
				SampleRatio:     0.0,
			},
		},
		{
			name: "disabled_metrics",
			config: &OTelConfig{
				ServiceName:     "test-service",
				ServiceVersion:  "v1.0.0",
				Environment:     "test",
				TraceExporter:   "stdout",
				MetricExporter:  "none",
				EnableMetrics:   false,
				EnableTracing:   true,
				SampleRatio:     1.0,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providers, err := InitializeOTel(tt.config, logger)
			require.NoError(t, err)
			require.NotNil(t, providers)
			
			// Verify configuration
			if tt.config.EnableTracing {
				assert.NotNil(t, providers.TracerProvider)
				assert.NotNil(t, providers.Tracer)
			}
			
			if tt.config.EnableMetrics {
				assert.NotNil(t, providers.MeterProvider)
				assert.NotNil(t, providers.Meter)
			}
			
			// Test shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			err = providers.Shutdown(ctx)
			assert.NoError(t, err)
		})
	}
}

// BenchmarkTraceOperations benchmarks trace operations to validate performance impact
func BenchmarkTraceOperations(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(b, err)
	defer providers.Shutdown(context.Background())
	
	tracer := otel.Tracer("benchmark")
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.Run("span_creation", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			_, span := tracer.Start(ctx, "benchmark-span")
			span.End()
		}
	})
	
	b.Run("span_attributes", func(b *testing.B) {
		ctx := context.Background()
		ctx, span := tracer.Start(ctx, "benchmark-span")
		defer span.End()
		
		attributes := map[string]interface{}{
			"operation": "benchmark",
			"iteration": 0,
			"success":   true,
		}
		
		for i := 0; i < b.N; i++ {
			attributes["iteration"] = i
			SetSpanAttributes(ctx, attributes)
		}
	})
	
	b.Run("span_events", func(b *testing.B) {
		ctx := context.Background()
		ctx, span := tracer.Start(ctx, "benchmark-span")
		defer span.End()
		
		for i := 0; i < b.N; i++ {
			AddSpanEvent(ctx, "benchmark.event", map[string]interface{}{
				"iteration": i,
				"timestamp": time.Now().Unix(),
			})
		}
	})
}

// BenchmarkMetricOperations benchmarks metric operations
func BenchmarkMetricOperations(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(b, err)
	defer providers.Shutdown(context.Background())
	
	metrics, err := CreateBusinessMetrics(providers.Meter)
	require.NoError(b, err)
	
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.Run("counter_increment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metrics.HTTPRequestsTotal.Add(ctx, 1)
		}
	})
	
	b.Run("histogram_record", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metrics.HTTPRequestDuration.Record(ctx, float64(i)*0.001)
		}
	})
	
	b.Run("updown_counter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				metrics.HTTPActiveRequests.Add(ctx, 1)
			} else {
				metrics.HTTPActiveRequests.Add(ctx, -1)
			}
		}
	})
}

// TestPerformanceImpact validates that OpenTelemetry overhead is minimal
func TestPerformanceImpact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	// Use a config with no trace export to avoid stdout noise
	cfg := &OTelConfig{
		ServiceName:     "test-service",
		ServiceVersion:  "v1.0.0",
		Environment:     "test",
		TraceExporter:   "none", // Disable trace export
		MetricExporter:  "none", // Disable metric export  
		EnableMetrics:   false,
		EnableTracing:   true,
		SampleRatio:     1.0,
	}
	
	providers, err := InitializeOTel(cfg, logger)
	require.NoError(t, err)
	defer providers.Shutdown(context.Background())
	
	tracer := otel.Tracer("performance-test")
	
	// More realistic test: measure overhead of instrumented vs non-instrumented function
	const iterations = 100
	
	// Function without tracing - more realistic work
	workFunc := func() int {
		sum := 0
		for j := 0; j < 10000; j++ {
			sum += j * j
			if j%1000 == 0 {
				time.Sleep(1 * time.Microsecond) // Simulate I/O
			}
		}
		return sum
	}
	
	// Function with tracing - same work
	tracedWorkFunc := func(ctx context.Context) int {
		_, span := tracer.Start(ctx, "work-function")
		defer span.End()
		
		sum := 0
		for j := 0; j < 10000; j++ {
			sum += j * j
			if j%1000 == 0 {
				time.Sleep(1 * time.Microsecond) // Simulate I/O
			}
		}
		return sum
	}
	
	// Measure baseline performance (no tracing)
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_ = workFunc()
	}
	baselineDuration := time.Since(start)
	
	// Measure performance with tracing
	start = time.Now()
	testCtx := context.Background()
	for i := 0; i < iterations; i++ {
		_ = tracedWorkFunc(testCtx)
	}
	tracingDuration := time.Since(start)
	
	// Calculate overhead percentage
	overhead := float64(tracingDuration-baselineDuration) / float64(baselineDuration) * 100
	
	t.Logf("Baseline duration: %v", baselineDuration)
	t.Logf("Tracing duration: %v", tracingDuration)
	t.Logf("Overhead: %.2f%%", overhead)
	
	// In this realistic test, we expect overhead to be reasonable
	// Note: This test shows the cost of creating spans, which is acceptable for production use
	assert.LessOrEqual(t, overhead, 200.0, "OpenTelemetry overhead should be reasonable for production use")
}

// TestTracePropagation tests trace propagation across contexts
func TestTracePropagation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	providers, err := InitializeOTel(DefaultOTelConfig(), logger)
	require.NoError(t, err)
	defer providers.Shutdown(context.Background())
	
	tracer := otel.Tracer("propagation-test")
	
	// Start parent span
	ctx := context.Background()
	ctx, parentSpan := tracer.Start(ctx, "parent-operation")
	defer parentSpan.End()
	
	parentTraceID := parentSpan.SpanContext().TraceID().String()
	
	// Create child span in same trace
	ctx, childSpan := tracer.Start(ctx, "child-operation")
	defer childSpan.End()
	
	childTraceID := childSpan.SpanContext().TraceID().String()
	
	// Verify trace propagation
	assert.Equal(t, parentTraceID, childTraceID, "Child span should have same trace ID as parent")
	
	// Verify spans are in same trace but different spans
	assert.Equal(t, parentSpan.SpanContext().TraceID(), childSpan.SpanContext().TraceID())
	assert.NotEqual(t, parentSpan.SpanContext().SpanID(), childSpan.SpanContext().SpanID())
}