package operations

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/isxcli/isxcli/pkg/contracts/events"
)

// StatusBroadcaster provides simple WebSocket broadcasting for operation status.
// It acts as a direct bridge between backend operations and frontend components.
type StatusBroadcaster struct {
	mu      sync.RWMutex
	stopped bool
	logger  *slog.Logger
	hub     WebSocketHub // WebSocket hub for broadcasting
}

// NewStatusBroadcaster creates a real-time broadcaster for operation status updates.
// It provides the Single Source of Truth for WebSocket communication between operations and frontend.
func NewStatusBroadcaster(hub WebSocketHub, logger *slog.Logger) *StatusBroadcaster {
	return &StatusBroadcaster{
		hub:    hub,
		logger: logger,
	}
}

// Stop marks the broadcaster as stopped.
func (sb *StatusBroadcaster) Stop() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.stopped = true
}

// UpdateStepWithMetadata broadcasts real-time progress updates via WebSocket.
// Validates operation ID to prevent "temp" ID broadcasts and ensures SSOT compliance.
func (sb *StatusBroadcaster) UpdateStepWithMetadata(operationID, stepID string, progress int, message string, metadata map[string]interface{}) {
	sb.mu.RLock()
	stopped := sb.stopped
	hub := sb.hub
	sb.mu.RUnlock()

	if stopped {
		return
	}

	// CRITICAL: Validate operation ID to prevent "temp" ID broadcasts
	if operationID == "" || operationID == "temp" {
		if sb.logger != nil {
			sb.logger.Debug("status_broadcaster.UpdateStepWithMetadata skipped - invalid operation ID",
				"operation_id", operationID,
				"step_id", stepID,
				"progress", progress,
				"reason", "invalid_operation_id",
			)
		}
		return
	}

	// Legacy progress events are deprecated; stage broadcasters emit `operation:snapshot`/`operation:delta`.
	_ = hub
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.UpdateStepWithMetadata noop (deprecated progress channel)",
			"operation_id", operationID,
			"step_id", stepID,
			"progress", progress,
			"message", message,
			"metadata_keys", len(metadata),
		)
	}
}

// UpdateStepProgress proxies to UpdateStepWithMetadata.
func (sb *StatusBroadcaster) UpdateStepProgress(operationID, stepID string, progress int, message string) {
	sb.UpdateStepWithMetadata(operationID, stepID, progress, message, nil)
}

// CompleteStep is a compatibility no-op.
func (sb *StatusBroadcaster) CompleteStep(operationID, stepID string, message string) {
	sb.UpdateStepWithMetadata(operationID, stepID, 100, message, map[string]interface{}{"status": "completed"})
}

// FailStep is a compatibility no-op.
func (sb *StatusBroadcaster) FailStep(operationID, stepID string, err error) {
	metadata := map[string]interface{}{
		"status": "failed",
	}
	if err != nil {
		metadata["error"] = err.Error()
	}
	sb.UpdateStepWithMetadata(operationID, stepID, 0, "Failed", metadata)
}

// GetSnapshot returns a placeholder snapshot so callers that expect a value do
// not crash; the UI ignores these since the real snapshots come from the stage
// broadcasters.
func (sb *StatusBroadcaster) GetSnapshot(operationID string) (*events.OperationSnapshot, bool) {
	return &events.OperationSnapshot{
		OperationID: operationID,
		Status:      "unknown",
		Progress:    0,
		StartedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, true
}

// GetAllSnapshots returns no state – snapshots are owned by stage broadcasters.
func (sb *StatusBroadcaster) GetAllSnapshots() []*events.OperationSnapshot {
	return []*events.OperationSnapshot{}
}

// CompleteOperation is a compatibility no-op. Stage broadcasters emit the
// final completion snapshot already, so duplicating it here would create
// conflicting data.
func (sb *StatusBroadcaster) CompleteOperation(operationID string, message string) {
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.CompleteOperation noop",
			"operation_id", operationID,
			"message", message,
		)
	}
}

// FailOperation is a compatibility no-op.
func (sb *StatusBroadcaster) FailOperation(operationID string, err error) {
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.FailOperation noop",
			"operation_id", operationID,
			"error", err,
		)
	}
}

// CancelOperation is a compatibility no-op.
func (sb *StatusBroadcaster) CancelOperation(operationID string) {
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.CancelOperation noop",
			"operation_id", operationID,
		)
	}
}

// SkipStep is a compatibility no-op.
func (sb *StatusBroadcaster) SkipStep(operationID, stepID string, message string) {
	sb.UpdateStepWithMetadata(operationID, stepID, 0, message, map[string]interface{}{"status": "skipped"})
}

// CreateOperation broadcasts operation creation event to establish correct operation type.
func (sb *StatusBroadcaster) CreateOperation(operationID string, stepNames []string, operationType string) {
	sb.mu.RLock()
	stopped := sb.stopped
	hub := sb.hub
	sb.mu.RUnlock()

	if stopped {
		return
	}

	// Validate operation ID to prevent invalid broadcasts
	if operationID == "" || operationID == "temp" || operationID == "undefined" || operationID == "null" {
		if sb.logger != nil {
			sb.logger.Debug("status_broadcaster.CreateOperation skipped - invalid operation ID",
				"operation_id", operationID,
				"reason", "invalid_operation_id",
			)
		}
		return
	}

	// Legacy creation events are deprecated; stages own their own broadcasting.
	_ = hub
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.CreateOperation noop (deprecated)",
			"operation_id", operationID,
			"step_count", len(stepNames),
			"operation_type", operationType,
		)
	}
}

// CreateOperationWithTrigger proxies to CreateOperation.
func (sb *StatusBroadcaster) CreateOperationWithTrigger(operationID string, stepNames []string, operationType, trigger, triggerDetails string) {
	sb.CreateOperation(operationID, stepNames, operationType)
}

// StartOperation is a compatibility no-op.
func (sb *StatusBroadcaster) StartOperation(operationID string) {
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.StartOperation noop",
			"operation_id", operationID,
		)
	}
}

// UpdateStatus is a compatibility no-op.
func (sb *StatusBroadcaster) UpdateStatus(operationID string, updateFunc func(*events.OperationSnapshot)) {
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.UpdateStatus noop",
			"operation_id", operationID,
		)
	}
	if updateFunc != nil {
		updateFunc(&events.OperationSnapshot{})
	}
}

// BroadcastUpdate sends real-time updates via WebSocket for any event type.
// Enables holiday detection messages and other dynamic progress updates.
func (sb *StatusBroadcaster) BroadcastUpdate(eventType, operationID, status string, metadata interface{}) {
	sb.mu.RLock()
	stopped := sb.stopped
	hub := sb.hub
	sb.mu.RUnlock()

	if stopped {
		return
	}

	// CRITICAL: Validate operation ID to prevent "temp" ID broadcasts
	if operationID == "" || operationID == "temp" {
		if sb.logger != nil {
			sb.logger.Debug("status_broadcaster.BroadcastUpdate skipped - invalid operation ID",
				"operation_id", operationID,
				"event_type", eventType,
				"status", status,
				"reason", "invalid_operation_id",
			)
		}
		return
	}

	// Deprecated: Stage broadcasters (`BaseStageBroadcaster`) own all operation messaging.
	_ = hub
	if sb.logger != nil {
		sb.logger.Debug("status_broadcaster.BroadcastUpdate noop (deprecated)",
			"operation_id", operationID,
			"event_type", eventType,
			"status", status,
		)
	}
}

// GetMetrics reports minimal diagnostics so health endpoints remain stable.
func (sb *StatusBroadcaster) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"status":  "minimal",
		"stopped": sb.stopped,
	}
}

// CleanupOldOperations is a compatibility no-op.
func (sb *StatusBroadcaster) CleanupOldOperations(ctx context.Context, maxAge time.Duration) {}
