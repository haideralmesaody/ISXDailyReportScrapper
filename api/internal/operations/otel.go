package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/isxcli/isxcli/internal/infrastructure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerName = "isxcli.operation"
)

// OperationTracer provides OpenTelemetry instrumentation for operation operations
type OperationTracer struct {
	tracer          trace.Tracer
	meter           metric.Meter
	businessMetrics *infrastructure.BusinessMetrics
}

// NewOperationTracer creates a new operation tracer
func NewOperationTracer(providers *infrastructure.OTelProviders) (*OperationTracer, error) {
	businessMetrics, err := infrastructure.CreateBusinessMetrics(providers.Meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create business metrics: %w", err)
	}

	return &OperationTracer{
		tracer:          otel.Tracer(TracerName),
		meter:           providers.Meter,
		businessMetrics: businessMetrics,
	}, nil
}

// TraceOperationExecution creates a span for the entire operation execution
func (pt *OperationTracer) TraceOperationExecution(ctx context.Context, operationID string, req OperationRequest) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("operation.execute.%s", req.Mode)
	ctx, span := pt.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("operation.id", operationID),
			attribute.String("operation.mode", req.Mode),
			attribute.String("operation.from_date", req.FromDate),
			attribute.String("operation.to_date", req.ToDate),
			attribute.String("operation.operation", "execute"),
		),
	)

	// Record operation start metric
	pt.businessMetrics.OperationExecutionsTotal.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("operation_mode", req.Mode),
			attribute.String("operation", "start"),
		),
	)

	pt.businessMetrics.OperationActiveOperations.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("operation_mode", req.Mode),
		),
	)

	return ctx, span
}

// TraceStageExecution creates a span for individual Step execution
func (pt *OperationTracer) TraceStageExecution(ctx context.Context, operationID, stageID string, stageType string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("operation.Step.%s", stageType)
	ctx, span := pt.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("operation.id", operationID),
			attribute.String("Step.id", stageID),
			attribute.String("Step.type", stageType),
			attribute.String("Step.operation", "execute"),
		),
	)

	// Record Step start metric
	pt.businessMetrics.OperationStageExecutions.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("stage_type", stageType),
			attribute.String("operation", "start"),
		),
	)

	return ctx, span
}

// RecordOperationCompletion records operation completion with metrics and span events
func (pt *OperationTracer) RecordOperationCompletion(ctx context.Context, span trace.Span, operationID string, duration time.Duration, status string, dataProcessed int64) {
	// Update span attributes
	span.SetAttributes(
		attribute.String("operation.status", status),
		attribute.Float64("operation.duration_seconds", duration.Seconds()),
		attribute.Int64("operation.data_processed_bytes", dataProcessed),
	)

	// Record completion metrics
	attrs := []attribute.KeyValue{
		attribute.String("status", status),
	}

	pt.businessMetrics.OperationExecutionDuration.Record(ctx, duration.Seconds(),
		metric.WithAttributes(attrs...),
	)

	pt.businessMetrics.OperationActiveOperations.Add(ctx, -1)

	if dataProcessed > 0 {
		pt.businessMetrics.OperationDataProcessed.Add(ctx, dataProcessed,
			metric.WithAttributes(
				attribute.String("operation_id", operationID),
			),
		)
	}

	// Add span event
	infrastructure.AddSpanEvent(ctx, "operation.completed", map[string]interface{}{
		"operation_id":   operationID,
		"status":         status,
		"duration":       duration.Seconds(),
		"data_processed": dataProcessed,
	})

	// Set span status
	if status == "success" {
		span.SetStatus(codes.Ok, "operation completed successfully")
	} else {
		span.SetStatus(codes.Error, fmt.Sprintf("operation failed with status: %s", status))
	}
}

// RecordStageCompletion records Step completion with metrics and span events
func (pt *OperationTracer) RecordStageCompletion(ctx context.Context, span trace.Span, operationID, stageID string, duration time.Duration, success bool, itemsProcessed int64) {
	status := "success"
	if !success {
		status = "failure"
	}

	// Update span attributes
	span.SetAttributes(
		attribute.String("Step.status", status),
		attribute.Float64("Step.duration_seconds", duration.Seconds()),
		attribute.Int64("Step.items_processed", itemsProcessed),
	)

	// Record Step metrics
	_ = []attribute.KeyValue{
		attribute.String("stage_id", stageID),
		attribute.String("status", status),
	}

	if itemsProcessed > 0 {
		span.SetAttributes(attribute.Int64("Step.items_processed", itemsProcessed))
	}

	// Add span event
	infrastructure.AddSpanEvent(ctx, "Step.completed", map[string]interface{}{
		"stage_id":        stageID,
		"status":          status,
		"duration":        duration.Seconds(),
		"items_processed": itemsProcessed,
	})

	// Set span status
	if success {
		span.SetStatus(codes.Ok, "Step completed successfully")
	} else {
		span.SetStatus(codes.Error, "Step execution failed")
	}
}

// RecordStageProgress records Step progress as span events
func (pt *OperationTracer) RecordStageProgress(ctx context.Context, operationID, stageID string, progress int, message string) {
	// Add progress event to span
	infrastructure.AddSpanEvent(ctx, "Step.progress", map[string]interface{}{
		"stage_id": stageID,
		"progress": progress,
		"message":  message,
	})

	// Update span attributes with current progress
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int("Step.progress_percent", progress),
			attribute.String("Step.progress_message", message),
		)
	}
}

// RecordStageError records Step errors with proper error tracking
func (pt *OperationTracer) RecordStageError(ctx context.Context, operationID, stageID string, err error) {
	infrastructure.RecordError(ctx, err,
		trace.WithAttributes(
			attribute.String("stage_id", stageID),
			attribute.String("error.type", "stage_execution_error"),
		),
	)

	// Record error metrics
	pt.businessMetrics.OperationStageExecutions.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("stage_id", stageID),
			attribute.String("status", "error"),
		),
	)
}

// RecordOperationError records operation errors with proper error tracking
func (pt *OperationTracer) RecordOperationError(ctx context.Context, operationID string, err error) {
	infrastructure.RecordError(ctx, err,
		trace.WithAttributes(
			attribute.String("operation_id", operationID),
			attribute.String("error.type", "operation_execution_error"),
		),
	)

	// Decrement active operation count on error
	pt.businessMetrics.OperationActiveOperations.Add(ctx, -1)
}

// TraceDataProcessing creates a span for data processing operations
func (pt *OperationTracer) TraceDataProcessing(ctx context.Context, operationType string, itemCount int) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("operation.data.%s", operationType)
	ctx, span := pt.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("data.operation", operationType),
			attribute.Int("data.item_count", itemCount),
		),
	)

	return ctx, span
}

// RecordDataProcessingCompletion records completion of data processing operations
func (pt *OperationTracer) RecordDataProcessingCompletion(ctx context.Context, span trace.Span, operationType string, itemsProcessed, bytesProcessed int64, duration time.Duration) {
	span.SetAttributes(
		attribute.Int64("data.items_processed", itemsProcessed),
		attribute.Int64("data.bytes_processed", bytesProcessed),
		attribute.Float64("data.duration_seconds", duration.Seconds()),
		attribute.Float64("data.throughput_items_per_second", float64(itemsProcessed)/duration.Seconds()),
	)

	if bytesProcessed > 0 {
		pt.businessMetrics.OperationDataProcessed.Add(ctx, bytesProcessed,
			metric.WithAttributes(
				attribute.String("operation", operationType),
			),
		)
	}

	infrastructure.AddSpanEvent(ctx, "data.processing.completed", map[string]interface{}{
		"operation":       operationType,
		"items_processed": itemsProcessed,
		"bytes_processed": bytesProcessed,
		"duration":        duration.Seconds(),
		"throughput_ips":  float64(itemsProcessed) / duration.Seconds(),
	})

	span.SetStatus(codes.Ok, fmt.Sprintf("Processed %d items in %v", itemsProcessed, duration))
}

// TraceWebSocketNotification creates a span for WebSocket notifications
func (pt *OperationTracer) TraceWebSocketNotification(ctx context.Context, messageType string, clientCount int) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("operation.websocket.%s", messageType)
	ctx, span := pt.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("websocket.message_type", messageType),
			attribute.Int("websocket.client_count", clientCount),
		),
	)

	return ctx, span
}

// RecordWebSocketNotificationCompletion records WebSocket notification completion
func (pt *OperationTracer) RecordWebSocketNotificationCompletion(ctx context.Context, span trace.Span, messageType string, successCount, failureCount int) {
	span.SetAttributes(
		attribute.Int("websocket.success_count", successCount),
		attribute.Int("websocket.failure_count", failureCount),
		attribute.Float64("websocket.success_rate", float64(successCount)/float64(successCount+failureCount)),
	)

	infrastructure.AddSpanEvent(ctx, "websocket.notification.completed", map[string]interface{}{
		"message_type":  messageType,
		"success_count": successCount,
		"failure_count": failureCount,
	})

	if failureCount > 0 {
		span.SetStatus(codes.Error, fmt.Sprintf("WebSocket notification had %d failures", failureCount))
	} else {
		span.SetStatus(codes.Ok, "WebSocket notification completed successfully")
	}
}

// TraceChromeOperation creates a span for Chrome/CDP operations
func (pt *OperationTracer) TraceChromeOperation(ctx context.Context, operation string) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("operation.chrome.%s", operation)
	ctx, span := pt.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("chrome.operation", operation),
			attribute.String("browser.name", "chromium"),
		),
	)

	return ctx, span
}

// RecordChromeOperationCompletion records Chrome operation completion
func (pt *OperationTracer) RecordChromeOperationCompletion(ctx context.Context, span trace.Span, operation string, success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "failure"
	}

	span.SetAttributes(
		attribute.String("chrome.status", status),
		attribute.Float64("chrome.duration_seconds", duration.Seconds()),
	)

	if success {
		span.SetStatus(codes.Ok, fmt.Sprintf("Chrome %s completed successfully", operation))
	} else {
		span.SetStatus(codes.Error, fmt.Sprintf("Chrome %s failed", operation))
	}
}

// GetGlobalOperationTracer returns a global operation tracer instance
var globalOperationTracer *OperationTracer

// InitGlobalOperationTracer initializes the global operation tracer
func InitGlobalOperationTracer(providers *infrastructure.OTelProviders) error {
	tracer, err := NewOperationTracer(providers)
	if err != nil {
		return err
	}
	globalOperationTracer = tracer
	return nil
}

// GetOperationTracer returns the global operation tracer
func GetOperationTracer() *OperationTracer {
	return globalOperationTracer
}
