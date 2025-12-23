package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/operations"
)

// ExecutionService handles operation execution and lifecycle management
type ExecutionService struct {
	manager        *operations.Manager
	licenseManager license.ManagerInterface
	logger         *slog.Logger
	paths          *config.Paths
	jobQueue       *operations.JobQueue
	validator      *ValidationService
}

// NewExecutionService creates a new execution service
func NewExecutionService(
	manager *operations.Manager,
	licenseManager license.ManagerInterface,
	logger *slog.Logger,
	paths *config.Paths,
	jobQueue *operations.JobQueue,
	validator *ValidationService,
) *ExecutionService {
	return &ExecutionService{
		manager:        manager,
		licenseManager: licenseManager,
		logger:         logger,
		paths:          paths,
		jobQueue:       jobQueue,
		validator:      validator,
	}
}

// ExecuteOperation executes an operation with the given request
func (es *ExecutionService) ExecuteOperation(ctx context.Context, request *operations.OperationRequest) (*operations.OperationResponse, error) {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := es.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "execution_service"),
		slog.String("method", "ExecuteOperation"),
		slog.String("operation_id", request.ID),
	)

	// Validate operation request
	if err := es.validator.ValidateOperationRequest(ctx, request); err != nil {
		logger.ErrorContext(ctx, "operation_request_validation_failed",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("operation request validation failed: %w", err)
	}

	// Perform license validation before executing operation
	operationType := request.Mode
	if operationType == "" {
		if step, ok := request.Parameters["step"].(string); ok {
			operationType = step
		} else {
			operationType = "general"
		}
	}

	if err := es.ValidateLicenseForOperation(ctx, operationType); err != nil {
		logger.ErrorContext(ctx, "license_validation_failed",
			slog.String("operation_type", operationType),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("license validation failed for %s operation: %w", operationType, err)
	}

	logger.InfoContext(ctx, "license_validation_passed",
		slog.String("operation_type", operationType))

	// Prefer job queue when available (queue-driven standardization)
	if es.jobQueue != nil {
		logger.InfoContext(ctx, "using_job_queue_for_operation",
			slog.String("operation_type", operationType))

		// Create job for queue execution
		job := &operations.Job{
			ID:          fmt.Sprintf("job-%d", time.Now().Unix()),
			OperationID: request.ID,
			StageID:     operationType,
			StageName:   operationType,
			Status:      operations.JobStatusPending,
			CreatedAt:   time.Now(),
			Request:     request,
		}

		// Submit to job queue
		if err := es.jobQueue.Enqueue(job); err != nil {
			logger.ErrorContext(ctx, "failed_to_submit_job_to_queue",
				slog.String("job_id", job.ID),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to submit job to queue: %w", err)
		}

		logger.InfoContext(ctx, "job_submitted_to_queue_successfully",
			slog.String("job_id", job.ID),
			slog.String("operation_id", request.ID))

		// Return response indicating job queued
		return &operations.OperationResponse{
			ID:      request.ID,
			Status:  string(operations.OperationStatusPending),
			Message: "Operation queued for execution",
		}, nil
	}

	// Fallback to synchronous execution (legacy path)
	logger.WarnContext(ctx, "job_queue_not_available_using_synchronous_execution",
		slog.String("operation_type", operationType))

	resp, err := es.manager.Execute(ctx, *request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute operation: %w", err)
	}

	logger.InfoContext(ctx, "operation_executed_successfully",
		slog.String("id", resp.ID),
		slog.String("status", string(resp.Status)))

	return resp, nil
}

// ExecuteStage executes a specific stage
func (es *ExecutionService) ExecuteStage(ctx context.Context, stageID string, parameters map[string]interface{}) (*operations.OperationResponse, error) {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := es.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "execution_service"),
		slog.String("method", "ExecuteStage"),
		slog.String("stage_id", stageID),
	)

	// Get stage from registry to validate it exists
	registry := es.manager.GetRegistry()
	_, err := registry.Get(stageID)
	if err != nil {
		logger.ErrorContext(ctx, "stage_not_found_in_registry",
			slog.String("stage_id", stageID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("stage not found: %s - %w", stageID, err)
	}

	// Validate stage parameters
	if err := es.validator.ValidateStageParameters(ctx, stageID, parameters); err != nil {
		logger.ErrorContext(ctx, "stage_parameter_validation_failed",
			slog.String("stage_id", stageID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("stage parameter validation failed for '%s': %w", stageID, err)
	}

	// Perform license validation for stage
	if err := es.ValidateLicenseForOperation(ctx, stageID); err != nil {
		logger.ErrorContext(ctx, "license_validation_failed_for_stage",
			slog.String("stage_id", stageID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("license validation failed for stage %s: %w", stageID, err)
	}

	logger.InfoContext(ctx, "license_validation_passed_for_stage",
		slog.String("stage_id", stageID))

	// Create operation request for stage execution
	request := &operations.OperationRequest{
		ID:         fmt.Sprintf("stage-%s-%d", stageID, time.Now().Unix()),
		Mode:       stageID,
		Parameters: parameters,
	}

	// Execute the stage
	resp, err := es.manager.Execute(ctx, *request)
	if err != nil {
		return nil, fmt.Errorf("stage execution failed: %w", err)
	}

	logger.InfoContext(ctx, "stage_executed_successfully",
		slog.String("operation_id", resp.ID),
		slog.String("stage_id", stageID))

	return resp, nil
}

// ExecuteFullPipeline is not supported in single-stage mode
func (es *ExecutionService) ExecuteFullPipeline(ctx context.Context, parameters map[string]interface{}) (*operations.OperationResponse, error) {
	return nil, fmt.Errorf("full pipeline not supported")
}

func (es *ExecutionService) GetOperationStatus(ctx context.Context, operationID string) (*operations.OperationState, error) {
	logger := es.logger.With(
		slog.String("component", "execution_service"),
		slog.String("method", "GetOperationStatus"),
		slog.String("operation_id", operationID),
	)

	state, err := es.manager.GetOperation(operationID)
	if err != nil {
		logger.ErrorContext(ctx, "operation_not_found",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("operation not found: %w", err)
	}

	logger.DebugContext(ctx, "operation_status_retrieved",
		slog.String("operation_id", operationID),
		slog.String("status", string(state.Status)))

	return state, nil
}

// ListOperations returns all operations
func (es *ExecutionService) ListOperations(ctx context.Context) ([]*operations.OperationState, error) {
	logger := es.logger.With(
		slog.String("component", "execution_service"),
		slog.String("method", "ListOperations"),
	)

	states := es.manager.ListOperations()

	logger.DebugContext(ctx, "operations_listed",
		slog.Int("total_operations", len(states)))

	return states, nil
}

// CancelOperation cancels a running operation
func (es *ExecutionService) CancelOperation(ctx context.Context, operationID string) error {
	logger := es.logger.With(
		slog.String("component", "execution_service"),
		slog.String("method", "CancelOperation"),
		slog.String("operation_id", operationID),
	)

	if err := es.manager.CancelOperation(operationID); err != nil {
		logger.ErrorContext(ctx, "failed_to_cancel_operation",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to cancel operation: %w", err)
	}

	logger.InfoContext(ctx, "operation_cancelled_successfully",
		slog.String("operation_id", operationID))

	return nil
}

// ValidateLicenseForOperation performs comprehensive license validation for operations
func (es *ExecutionService) ValidateLicenseForOperation(ctx context.Context, operationType string) error {
	if es.licenseManager == nil {
		return fmt.Errorf("license manager not available")
	}

	// Get license status
	_, status, err := es.licenseManager.GetLicenseStatus()
	if err != nil {
		return fmt.Errorf("failed to get license status: %w", err)
	}

	// Check if license is valid for operations
	switch status {
	case "Active", "Activated", "Valid", "Warning", "Critical":
		return nil // License is valid for operations
	case "Not Activated", "Expired":
		return fmt.Errorf("license is %s - operations not permitted", status)
	default:
		// Unknown status - be conservative and deny operations
		return fmt.Errorf("unknown license status: %s - operations not permitted", status)
	}
}

// GetOperationMetrics returns metrics about operations
func (es *ExecutionService) GetOperationMetrics(ctx context.Context) (map[string]interface{}, error) {
	logger := es.logger.With(
		slog.String("component", "execution_service"),
		slog.String("method", "GetOperationMetrics"),
	)

	operations, err := es.ListOperations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list operations: %w", err)
	}

	activeCount := 0
	completedCount := 0
	failedCount := 0

	for _, op := range operations {
		switch op.Status {
		case "running", "pending":
			activeCount++
		case "completed":
			completedCount++
		case "failed", "cancelled":
			failedCount++
		}
	}

	metrics := map[string]interface{}{
		"total_operations":     len(operations),
		"active_operations":    activeCount,
		"completed_operations": completedCount,
		"failed_operations":    failedCount,
		"timestamp":            time.Now().Unix(),
	}

	logger.DebugContext(ctx, "operation_metrics_retrieved",
		slog.Int("total", len(operations)),
		slog.Int("active", activeCount))

	return metrics, nil
}
