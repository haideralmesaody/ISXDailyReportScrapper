package contracts

import (
	"fmt"
	"time"
)

// BroadcastType defines the type of broadcast message
type BroadcastType string

const (
	// Operation lifecycle events
	BroadcastTypeOperationCreated  BroadcastType = "operation:created"
	BroadcastTypeOperationProgress BroadcastType = "operation:progress"
	BroadcastTypeOperationComplete BroadcastType = "operation:complete"
	BroadcastTypeOperationFailed    BroadcastType = "operation:failed"
	BroadcastTypeOperationCancelled BroadcastType = "operation:cancelled"

	// Stage lifecycle events
	BroadcastTypeStageCreated   BroadcastType = "stage:created"
	BroadcastTypeStageProgress  BroadcastType = "stage:progress"
	BroadcastTypeStageComplete  BroadcastType = "stage:complete"
	BroadcastTypeStageFailed    BroadcastType = "stage:failed"
	BroadcastTypeStageCancelled BroadcastType = "stage:cancelled"

	// Progress updates
	BroadcastTypeProgress      BroadcastType = "progress"
	BroadcastTypeStepProgress  BroadcastType = "step:progress"
	BroadcastTypeFileProgress BroadcastType = "file:progress"

	// System events
	BroadcastTypeSystemUpdate   BroadcastType = "system:update"
	BroadcastTypeSystemAlert    BroadcastType = "system:alert"
	BroadcastTypeError         BroadcastType = "error"
	BroadcastTypeWarning       BroadcastType = "warning"

	// Data updates
	BroadcastTypeDataUpdate    BroadcastType = "data:update"
	BroadcastTypeDataRefresh    BroadcastType = "data:refresh"
)

// Request/Response types for broadcasting operations

// CreateOperationRequest creates a new operation
type CreateOperationRequest struct {
	OperationID   string                 `json:"operation_id"`
	StepNames     []string               `json:"step_names"`
	OperationType string                 `json:"operation_type"`
	StartTime      time.Time               `json:"start_time,omitempty"`
}

// CompleteOperationRequest marks an operation as complete
type CompleteOperationRequest struct {
	OperationID string                 `json:"operation_id"`
	Message     string                 `json:"message"`
	EndTime     time.Time               `json:"end_time,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// FailOperationRequest marks an operation as failed
type FailOperationRequest struct {
	OperationID string                 `json:"operation_id"`
	Error       error                  `json:"error"`
	EndTime     time.Time               `json:"end_time,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// CancelOperationRequest cancels an operation
type CancelOperationRequest struct {
	OperationID string                 `json:"operation_id"`
	Reason      string                 `json:"reason"`
	EndTime     time.Time               `json:"end_time,omitempty"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
}

// ProgressUpdateRequest updates operation or stage progress
type ProgressUpdateRequest struct {
	OperationID string                 `json:"operation_id"`
	StepID      string                 `json:"step_id,omitempty"`
	Progress    int                     `json:"progress"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
	Timestamp   time.Time               `json:"timestamp,omitempty"`
}

// StepUpdateRequest updates a specific step
type StepUpdateRequest struct {
	OperationID string                 `json:"operation_id"`
	StepID      string                 `json:"step_id"`
	Progress    int                     `json:"progress"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{}   `json:"metadata,omitempty"`
	Timestamp   time.Time               `json:"timestamp,omitempty"`
}

// Snapshot represents operation state
type Snapshot struct {
	OperationID    string                 `json:"operation_id"`
	Status         string                 `json:"status"`
	Progress       int                     `json:"progress"`
	CurrentStep    string                 `json:"current_step,omitempty"`
	StartTime      time.Time               `json:"start_time,omitempty"`
	EndTime        time.Time               `json:"end_time,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Metadata       map[string]interface{}   `json:"metadata,omitempty"`
	UpdatedAt      time.Time               `json:"updated_at,omitempty"`
	Steps          []StepSnapshot          `json:"steps,omitempty"`
}

// StepSnapshot represents individual step state
type StepSnapshot struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Progress  int                     `json:"progress"`
	Status    string                 `json:"status,omitempty"`
	Message   string                 `json:"message,omitempty"`
	StartTime time.Time               `json:"start_time,omitempty"`
	EndTime   time.Time               `json:"end_time,omitempty"`
	Metadata  map[string]interface{}   `json:"metadata,omitempty"`
	UpdatedAt time.Time               `json:"updated_at,omitempty"`
}

// CleanupRequest requests cleanup of old operations
type CleanupRequest struct {
	MaxAge time.Duration `json:"max_age"`
}

// Metrics represents broadcasting performance metrics
type Metrics struct {
	TotalBroadcasts    int64     `json:"total_broadcasts"`
	SuccessfulBroadcasts int64     `json:"successful_broadcasts"`
	FailedBroadcasts     int64     `json:"failed_broadcasts"`
	LastBroadcastTime    time.Time `json:"last_broadcast_time,omitempty"`
	AverageResponseTime   string    `json:"average_response_time,omitempty"`
}

// CustomError types for broadcasting
var (
	ErrInvalidOperationID     = fmt.Errorf("invalid operation ID")
	ErrTemporaryOperationID   = fmt.Errorf("temporary operation ID not allowed")
	ErrBroadcasterDisabled    = fmt.Errorf("broadcaster is disabled")
	ErrWebSocketHubNotConfigured = fmt.Errorf("WebSocket hub not configured")
)