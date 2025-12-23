package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/operations"
)

// OperationsMetricsHandler handles operations-specific metrics endpoints
type OperationsMetricsHandler struct {
	manager  *operations.Manager
	jobQueue *operations.JobQueue
	logger   *slog.Logger
	tracer   trace.Tracer
	meter    metric.Meter

	// Metrics collectors
	httpRequestDuration   metric.Float64Histogram
	httpRequestsTotal     metric.Int64Counter
	httpActiveRequests    metric.Int64UpDownCounter
}

// NewOperationsMetricsHandler creates a new operations metrics handler
func NewOperationsMetricsHandler(manager *operations.Manager, jobQueue *operations.JobQueue, logger *slog.Logger) (*OperationsMetricsHandler, error) {
	if manager == nil {
		panic("manager cannot be nil")
	}
	if jobQueue == nil {
		panic("jobQueue cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}

	tracer := otel.Tracer("operations-metrics-handler")
	meter := otel.Meter("operations-metrics-handler")

	// Create metrics
	httpRequestDuration, err := meter.Float64Histogram(
		"operations_handler_request_duration_seconds",
		metric.WithDescription("HTTP request duration for operations endpoints in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	httpRequestsTotal, err := meter.Int64Counter(
		"operations_handler_requests_total",
		metric.WithDescription("Total number of HTTP requests to operations endpoints"),
	)
	if err != nil {
		return nil, err
	}

	httpActiveRequests, err := meter.Int64UpDownCounter(
		"operations_handler_active_requests",
		metric.WithDescription("Number of active HTTP requests to operations endpoints"),
	)
	if err != nil {
		return nil, err
	}

	return &OperationsMetricsHandler{
		manager:             manager,
		jobQueue:            jobQueue,
		logger:              logger.With(slog.String("handler", "operations_metrics")),
		tracer:              tracer,
		meter:               meter,
		httpRequestDuration: httpRequestDuration,
		httpRequestsTotal:   httpRequestsTotal,
		httpActiveRequests:  httpActiveRequests,
	}, nil
}

// Routes returns a chi router for operations metrics endpoints
func (h *OperationsMetricsHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Apply middleware with instrumentation
	r.Use(h.instrumentMiddleware)

	// Metrics endpoints
	r.Get("/summary", h.GetOperationsSummary)
	r.Get("/performance", h.GetPerformanceMetrics)
	r.Get("/health", h.GetOperationsHealth)

	return r
}

// instrumentMiddleware adds OpenTelemetry instrumentation to requests
func (h *OperationsMetricsHandler) instrumentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		route := chi.RouteContext(ctx).RoutePattern()
		
		// Record request start
		h.httpActiveRequests.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("route", route),
			),
		)
		defer h.httpActiveRequests.Add(ctx, -1,
			metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("route", route),
			),
		)

		// Track request duration
		startTime := time.Now()
		
		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// Call next handler
		next.ServeHTTP(ww, r)
		
		duration := time.Since(startTime)
		
		// Record metrics
		h.httpRequestsTotal.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("route", route),
				attribute.Int("status", ww.Status()),
			),
		)
		
		h.httpRequestDuration.Record(ctx, duration.Seconds(),
			metric.WithAttributes(
				attribute.String("method", r.Method),
				attribute.String("route", route),
				attribute.Int("status", ww.Status()),
			),
		)
	})
}

// GetOperationsSummary returns a summary of all operations
func (h *OperationsMetricsHandler) GetOperationsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	// Start span
	ctx, span := h.tracer.Start(ctx, "metrics.get_operations_summary",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/api/operations/metrics/summary"),
			attribute.String("request_id", reqID),
		),
	)
	defer span.End()

	h.logger.DebugContext(ctx, "retrieving operations summary",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))

	// Get all operations directly from manager
	operationsList := h.manager.ListOperations()

	// Calculate summary statistics
	summary := h.calculateSummary(operationsList)

	span.SetAttributes(
		attribute.Int("operations.total", summary["total"].(int)),
		attribute.Int("operations.active", summary["active"].(int)),
		attribute.Int("operations.completed", summary["completed"].(int)),
		attribute.Int("operations.failed", summary["failed"].(int)),
	)

	render.JSON(w, r, summary)
}

// GetPerformanceMetrics returns performance metrics for operations
func (h *OperationsMetricsHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	// Start span
	ctx, span := h.tracer.Start(ctx, "metrics.get_performance_metrics",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/api/operations/metrics/performance"),
			attribute.String("request_id", reqID),
		),
	)
	defer span.End()

	h.logger.DebugContext(ctx, "retrieving performance metrics",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))

	// Get current operations
	operationsList := h.manager.ListOperations()

	// Calculate performance metrics
	metrics := h.calculatePerformanceMetrics(operationsList)
	
	span.SetAttributes(
		attribute.Float64("performance.avg_duration_seconds", metrics["avg_duration_seconds"].(float64)),
		attribute.Float64("performance.success_rate", metrics["success_rate"].(float64)),
	)

	render.JSON(w, r, metrics)
}

// GetOperationsHealth returns health status of the operations system
func (h *OperationsMetricsHandler) GetOperationsHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	// Start span
	ctx, span := h.tracer.Start(ctx, "metrics.get_operations_health",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/api/operations/metrics/health"),
			attribute.String("request_id", reqID),
		),
	)
	defer span.End()

	h.logger.DebugContext(ctx, "checking operations health",
		slog.String("request_id", reqID),
		slog.String("trace_id", infrastructure.TraceIDFromContext(ctx)))

	// Get current operations
	operationsList := h.manager.ListOperations()

	// Check health criteria
	health := h.calculateHealth(operationsList)
	
	span.SetAttributes(
		attribute.String("health.status", health["status"].(string)),
		attribute.Bool("health.is_healthy", health["status"].(string) == "healthy"),
	)

	statusCode := http.StatusOK
	if health["status"] != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	render.Status(r, statusCode)
	render.JSON(w, r, health)
}

// calculateSummary calculates summary statistics for operations
func (h *OperationsMetricsHandler) calculateSummary(operationsList []*operations.OperationState) map[string]interface{} {
	summary := map[string]interface{}{
		"total":      len(operationsList),
		"active":     0,
		"pending":    0,
		"running":    0,
		"completed":  0,
		"failed":     0,
		"cancelled":  0,
		"timestamp":  time.Now().UTC(),
	}

	statusCounts := make(map[string]int)
	for _, op := range operationsList {
		statusCounts[string(op.Status)]++
		
		// Count active operations (pending or running)
		if op.Status == operations.OperationStatusPending || op.Status == operations.OperationStatusRunning {
			summary["active"] = summary["active"].(int) + 1
		}
	}

	// Update individual status counts
	summary["pending"] = statusCounts[string(operations.OperationStatusPending)]
	summary["running"] = statusCounts[string(operations.OperationStatusRunning)]
	summary["completed"] = statusCounts[string(operations.OperationStatusCompleted)]
	summary["failed"] = statusCounts[string(operations.OperationStatusFailed)]
	summary["cancelled"] = statusCounts[string(operations.OperationStatusCancelled)]

	// Add breakdown by operation type if needed
	typeBreakdown := make(map[string]map[string]int)
	for _, op := range operationsList {
		// Extract operation type from metadata if available
		opType := "unknown"
		// Get first step name as operation type
		for _, step := range op.Steps {
			opType = step.Name
			break
		}
		
		if _, exists := typeBreakdown[opType]; !exists {
			typeBreakdown[opType] = make(map[string]int)
		}
		typeBreakdown[opType][string(op.Status)]++
	}
	
	summary["by_type"] = typeBreakdown

	return summary
}

// calculatePerformanceMetrics calculates performance metrics
func (h *OperationsMetricsHandler) calculatePerformanceMetrics(operationsList []*operations.OperationState) map[string]interface{} {
	metrics := map[string]interface{}{
		"total_operations":      len(operationsList),
		"avg_duration_seconds":  0.0,
		"min_duration_seconds":  0.0,
		"max_duration_seconds":  0.0,
		"success_rate":         0.0,
		"failure_rate":         0.0,
		"cancellation_rate":    0.0,
		"timestamp":            time.Now().UTC(),
	}

	if len(operationsList) == 0 {
		return metrics
	}

	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration
	var completedCount, successCount, failedCount, cancelledCount int

	for _, op := range operationsList {
		// Only calculate duration for completed operations
		if op.EndTime != nil {
			duration := op.Duration()
			totalDuration += duration
			completedCount++

			if minDuration == 0 || duration < minDuration {
				minDuration = duration
			}
			if duration > maxDuration {
				maxDuration = duration
			}
		}

		// Count outcomes
		switch op.Status {
		case operations.OperationStatusCompleted:
			successCount++
		case operations.OperationStatusFailed:
			failedCount++
		case operations.OperationStatusCancelled:
			cancelledCount++
		}
	}

	// Calculate averages and rates
	if completedCount > 0 {
		metrics["avg_duration_seconds"] = totalDuration.Seconds() / float64(completedCount)
		metrics["min_duration_seconds"] = minDuration.Seconds()
		metrics["max_duration_seconds"] = maxDuration.Seconds()
	}

	totalFinished := successCount + failedCount + cancelledCount
	if totalFinished > 0 {
		metrics["success_rate"] = float64(successCount) / float64(totalFinished)
		metrics["failure_rate"] = float64(failedCount) / float64(totalFinished)
		metrics["cancellation_rate"] = float64(cancelledCount) / float64(totalFinished)
	}

	// Add percentiles if we have enough data
	if completedCount >= 10 {
		durations := make([]time.Duration, 0, completedCount)
		for _, op := range operationsList {
			if op.EndTime != nil {
				durations = append(durations, op.Duration())
			}
		}
		
		// Calculate percentiles (simplified - in production use proper percentile calculation)
		metrics["p50_duration_seconds"] = durations[len(durations)/2].Seconds()
		metrics["p95_duration_seconds"] = durations[int(float64(len(durations))*0.95)].Seconds()
		metrics["p99_duration_seconds"] = durations[int(float64(len(durations))*0.99)].Seconds()
	}

	return metrics
}

// calculateHealth determines the health status of the operations system
func (h *OperationsMetricsHandler) calculateHealth(operationsList []*operations.OperationState) map[string]interface{} {
	health := map[string]interface{}{
		"status":     "healthy",
		"checks":     make(map[string]interface{}),
		"timestamp":  time.Now().UTC(),
	}

	checks := health["checks"].(map[string]interface{})

	// Check 1: Active operations count
	activeCount := 0
	for _, op := range operationsList {
		if op.Status == operations.OperationStatusRunning {
			activeCount++
		}
	}
	
	activeOpsHealthy := activeCount < 100 // Threshold for too many concurrent operations
	checks["active_operations"] = map[string]interface{}{
		"status": conditionalStatus(activeOpsHealthy),
		"count":  activeCount,
		"threshold": 100,
	}

	// Check 2: Recent failure rate
	recentOps := filterRecentOperations(operationsList, 1*time.Hour)
	failureRate := calculateRecentFailureRate(recentOps)
	
	failureRateHealthy := failureRate < 0.1 // 10% failure rate threshold
	checks["failure_rate"] = map[string]interface{}{
		"status":    conditionalStatus(failureRateHealthy),
		"rate":      failureRate,
		"threshold": 0.1,
		"window":    "1h",
	}

	// Check 3: Stuck operations (running for too long)
	stuckCount := 0
	for _, op := range operationsList {
		if op.Status == operations.OperationStatusRunning && op.StartTime.Before(time.Now().Add(-30*time.Minute)) {
			stuckCount++
		}
	}
	
	noStuckOps := stuckCount == 0
	checks["stuck_operations"] = map[string]interface{}{
		"status":    conditionalStatus(noStuckOps),
		"count":     stuckCount,
		"threshold": "30m",
	}

	// Overall health determination
	if !activeOpsHealthy || !failureRateHealthy || !noStuckOps {
		health["status"] = "unhealthy"
	}

	return health
}

// Helper functions

func conditionalStatus(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}

func filterRecentOperations(operationsList []*operations.OperationState, window time.Duration) []*operations.OperationState {
	cutoff := time.Now().Add(-window)
	recent := make([]*operations.OperationState, 0)
	
	for _, op := range operationsList {
		if op.StartTime.After(cutoff) {
			recent = append(recent, op)
		}
	}
	
	return recent
}

func calculateRecentFailureRate(operationsList []*operations.OperationState) float64 {
	if len(operationsList) == 0 {
		return 0.0
	}
	
	failedCount := 0
	completedCount := 0
	
	for _, op := range operationsList {
		if op.Status == operations.OperationStatusFailed {
			failedCount++
			completedCount++
		} else if op.Status == operations.OperationStatusCompleted {
			completedCount++
		}
	}
	
	if completedCount == 0 {
		return 0.0
	}
	
	return float64(failedCount) / float64(completedCount)
}