package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	errors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/infrastructure"
	"github.com/isxcli/isxcli/internal/license"
	"github.com/isxcli/isxcli/internal/operations"
)

// OperationService provides core operation orchestration with clean separation of concerns
type OperationService struct {
	manager            *operations.Manager
	licenseManager     license.ManagerInterface
	logger             *slog.Logger
	paths              *config.Paths
	jobQueue           *operations.JobQueue
	validator          *ValidationService
	configService      *ConfigurationService
	executionService   *ExecutionService
	operationValidator *OperationValidator
}

// NewOperationService creates a new operation service with factory-based stage registration
func NewOperationService(
	ctx context.Context,
	manager *operations.Manager,
	licenseManager license.ManagerInterface,
	logger *slog.Logger,
	paths *config.Paths,
) (*OperationService, error) {
	// Validate required dependencies
	if manager == nil {
		return nil, errors.NewValidationError("Operation manager is required for OperationService")
	}
	if licenseManager == nil {
		return nil, errors.NewValidationError("License manager is required for operations service")
	}
	if paths == nil {
		return nil, errors.NewValidationError("Paths configuration is required for OperationService")
	}

	// Add trace ID if not present
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = infrastructure.GenerateTraceID()
		ctx = infrastructure.WithTraceID(ctx, traceID)
	}

	serviceLogger := logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "NewOperationService"),
	)

	// Log initialization with structured logging
	serviceLogger.InfoContext(ctx, "OperationService initialization with factory-based registration",
		slog.String("executable_dir", paths.ExecutableDir),
		slog.String("data_dir", paths.DataDir),
		slog.String("downloads_dir", paths.DownloadsDir),
		slog.String("reports_dir", paths.ReportsDir),
		slog.Bool("license_enforcement", true),
		slog.String("registration_pattern", "factory_based"))

	// Use enhanced registerStages for robust stage registration
	if err := registerStages(ctx, manager, paths.ExecutableDir, serviceLogger, licenseManager); err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("Failed to register stages: %v", err))
	}

	// Validate operation manager health
	registry := manager.GetRegistry()
	if registry == nil {
		return nil, errors.NewValidationError("Operation manager registry is nil")
	}

	availableStages := registry.List()
	serviceLogger.InfoContext(ctx, "Operation service initialization completed",
		slog.Int("registered_stages", len(availableStages)),
		slog.Int("stage_ids", len(availableStages)),
		slog.String("registration_status", "success"))

	// Validate that we have the minimum required stages
	requiredStages := []string{"scraping", "processing", "indices", "liquidity"}
	missingStages := make([]string, 0)
	for _, required := range requiredStages {
		found := false
		for _, available := range availableStages {
			if available.ID() == required {
				found = true
				break
			}
		}
		if !found {
			missingStages = append(missingStages, required)
		}
	}

	if len(missingStages) > 0 {
		return nil, errors.NewValidationError(fmt.Sprintf("Missing required stages: %v", missingStages))
	}

	service := &OperationService{
		manager:        manager,
		licenseManager: licenseManager,
		logger:         serviceLogger,
		paths:          paths,
	}
	service.validator = NewValidationService(serviceLogger)
	service.configService = NewConfigurationService(serviceLogger)
	service.executionService = NewExecutionService(manager, licenseManager, serviceLogger, paths, nil, service.validator)
	service.operationValidator = NewOperationValidator(licenseManager, serviceLogger, service.ValidateExecutables)

	return service, nil
}

// NOTE: Complex orchestrator functions removed for simplicity and reliability
// The enhanced registerStages function below provides all necessary functionality

// STAGE_ORDER defines the explicit logical order of execution
var STAGE_ORDER = []string{
	"scraping",   // 1. Data Collection - Download ISX reports
	"processing", // 2. Data Processing - Convert Excel to CSV
	"indices",    // 3. Index Extraction - Extract ISX60 data
	"liquidity",  // 4. Liquidity Calculation - Calculate metrics
}

// registerStages creates and registers all pipeline stages with comprehensive error handling
// This is the primary stage registration method using factory pattern
func registerStages(ctx context.Context, manager *operations.Manager, executableDir string, logger *slog.Logger, licenseManager license.ManagerInterface) error {
	traceID := infrastructure.GetTraceID(ctx)
	if traceID == "" {
		traceID = infrastructure.GenerateTraceID()
		ctx = infrastructure.WithTraceID(ctx, traceID)
	}

	stageLogger := logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "registerStages"),
		slog.String("executable_dir", executableDir),
	)

	stageLogger.InfoContext(ctx, "Starting stage registration with factory pattern",
		slog.String("stage_order", strings.Join(STAGE_ORDER, " → ")))

	// Validate inputs
	if manager == nil {
		return errors.NewValidationError("manager cannot be nil")
	}
	if executableDir == "" {
		return errors.NewValidationError("executable directory cannot be empty")
	}
	if logger == nil {
		return errors.NewValidationError("logger cannot be nil")
	}

	// Validate executable directory exists
	if _, err := os.Stat(executableDir); os.IsNotExist(err) {
		return errors.NewValidationError("executable directory does not exist")
	}

	// Create stage options with comprehensive validation
	stageOptions := &operations.StageOptions{
		EnableProgress:   true,
		WebSocketManager: manager.GetHub(),
		LicenseChecker:   licenseManager,
	}

	// Validate stage options components
	if stageOptions.WebSocketManager == nil {
		stageLogger.WarnContext(ctx, "WebSocket hub is nil, progress reporting will be limited")
	}
	if manager.GetBroadcaster() == nil {
		return errors.NewValidationError("status broadcaster cannot be nil")
	}
	if licenseManager == nil {
		stageLogger.WarnContext(ctx, "License manager is nil, license validation will be skipped")
	}

	// Define stage factory configurations with error handling
	type stageFactory struct {
		name        string
		id          string
		createFunc  func(string, *slog.Logger, *operations.StageOptions) operations.Step
		required    bool
		description string
	}

	stages := []stageFactory{
		{
			name: "Data Collection",
			id:   "scraping",
			createFunc: func(dir string, log *slog.Logger, opts *operations.StageOptions) operations.Step {
				return operations.NewScrapingStage(dir, log, opts)
			},
			required:    true,
			description: "Downloads ISX daily reports from the exchange",
		},
		{
			name: "Data Processing",
			id:   "processing",
			createFunc: func(dir string, log *slog.Logger, opts *operations.StageOptions) operations.Step {
				return operations.NewProcessingStage(dir, log, opts)
			},
			required:    true,
			description: "Converts Excel files to CSV format with file-level progress tracking",
		},
		{
			name: "Index Extraction",
			id:   "indices",
			createFunc: func(dir string, log *slog.Logger, opts *operations.StageOptions) operations.Step {
				return operations.NewIndicesStage(dir, log, opts)
			},
			required:    true,
			description: "Extracts ISX60 and other index data from processed files",
		},
		{
			name: "Liquidity Calculation",
			id:   "liquidity",
			createFunc: func(dir string, log *slog.Logger, opts *operations.StageOptions) operations.Step {
				return operations.NewLiquidityStage(dir, log, opts)
			},
			required:    true,
			description: "Calculates liquidity metrics and scoring for all securities",
		},
	}

	// Registry for tracking created stages
	createdStages := make(map[string]operations.Step)
	registrationErrors := make([]error, 0)

	// Create and register each stage with error handling
	for _, stageFactory := range stages {
		stageLogger.InfoContext(ctx, "Creating stage",
			slog.String("stage_id", stageFactory.id),
			slog.String("stage_name", stageFactory.name),
			slog.Bool("required", stageFactory.required))

		// Create stage with panic recovery
		var stage operations.Step
		var createErr error

		func() {
			defer func() {
				if r := recover(); r != nil {
					createErr = fmt.Errorf("panic creating stage %s: %v", stageFactory.id, r)
					stageLogger.ErrorContext(ctx, "Panic recovered during stage creation",
						slog.String("stage_id", stageFactory.id),
						slog.Any("panic", r))
				}
			}()

			stage = stageFactory.createFunc(executableDir, stageLogger, stageOptions)
		}()

		if createErr != nil {
			errMsg := fmt.Sprintf("Failed to create stage %s: %v", stageFactory.id, createErr)
			stageLogger.ErrorContext(ctx, errMsg)

			if stageFactory.required {
				registrationErrors = append(registrationErrors,
					errors.NewValidationError(errMsg))
			} else {
				stageLogger.WarnContext(ctx, "Optional stage creation failed, continuing",
					slog.String("stage_id", stageFactory.id),
					slog.String("error", createErr.Error()))
			}
			continue
		}

		// Validate created stage
		if stage == nil {
			errMsg := fmt.Sprintf("Stage factory returned nil for %s", stageFactory.id)
			stageLogger.ErrorContext(ctx, errMsg)

			if stageFactory.required {
				registrationErrors = append(registrationErrors,
					errors.NewValidationError(errMsg))
			}
			continue
		}

		// Validate stage interface
		if stage.ID() == "" {
			errMsg := fmt.Sprintf("Stage %s has empty ID", stageFactory.id)
			stageLogger.ErrorContext(ctx, errMsg)

			if stageFactory.required {
				registrationErrors = append(registrationErrors,
					errors.NewValidationError(errMsg))
			}
			continue
		}

		if stage.Name() == "" {
			errMsg := fmt.Sprintf("Stage %s has empty name", stageFactory.id)
			stageLogger.ErrorContext(ctx, errMsg)

			if stageFactory.required {
				registrationErrors = append(registrationErrors,
					errors.NewValidationError(errMsg))
			}
			continue
		}

		// Store created stage
		createdStages[stageFactory.id] = stage
		stageLogger.InfoContext(ctx, "Stage created successfully",
			slog.String("stage_id", stage.ID()),
			slog.String("stage_name", stage.Name()))
	}

	// Check if any required stages failed to create
	if len(registrationErrors) > 0 {
		stageLogger.ErrorContext(ctx, "Stage creation failed with errors",
			slog.Int("error_count", len(registrationErrors)))

		// Combine all errors
		combinedErr := registrationErrors[0]
		for i, err := range registrationErrors[1:] {
			combinedErr = fmt.Errorf("%w; %v", combinedErr, err)
			if i == 0 { // Only add detail once
				combinedErr = errors.NewValidationError("Required stage creation failed")
			}
		}
		return combinedErr
	}

	// Register all created stages with error handling
	registry := manager.GetRegistry()
	if registry == nil {
		return errors.NewValidationError("Operation manager registry is nil")
	}

	registrationCount := 0
	for stageID, stage := range createdStages {
		stageLogger.InfoContext(ctx, "Registering stage",
			slog.String("stage_id", stageID),
			slog.String("stage_name", stage.Name()))

		// Register stage with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					stageLogger.ErrorContext(ctx, "Panic recovered during stage registration",
						slog.String("stage_id", stageID),
						slog.Any("panic", r))
					registrationErrors = append(registrationErrors,
						fmt.Errorf("panic registering stage %s: %v", stageID, r))
				}
			}()

			registry.Register(stage)
			registrationCount++
		}()
	}

	// Final validation
	if registrationCount == 0 {
		return errors.NewValidationError("No stages were successfully registered")
	}

	// Log successful registration
	stageLogger.InfoContext(ctx, "Stage registration completed successfully",
		slog.Int("registered_stages", registrationCount),
		slog.Int("expected_stages", len(stages)),
		slog.String("registry_status", "active"))

	// Validate registry state
	availableStages := registry.List()
	stageLogger.InfoContext(ctx, "Registry validation",
		slog.Int("available_stages", len(availableStages)),
		slog.Int("stage_ids_count", len(availableStages)))

	if len(availableStages) < registrationCount {
		stageLogger.WarnContext(ctx, "Registry stage count mismatch",
			slog.Int("registered", registrationCount),
			slog.Int("available", len(availableStages)))
	}

	return nil
}

// ValidateLicenseForOperation performs comprehensive license validation for operations
func (ps *OperationService) ValidateLicenseForOperation(ctx context.Context, operationType string) error {
	if ps.operationValidator == nil {
		return fmt.Errorf("operation validator not initialized")
	}
	return ps.operationValidator.Validate(ctx, operationType)
}

// SetJobQueue sets the job queue for operations
func (ps *OperationService) SetJobQueue(jobQueue *operations.JobQueue) {
	ps.jobQueue = jobQueue
	if ps.executionService != nil {
		ps.executionService.jobQueue = jobQueue
	}
}

// StartOperation starts a new operation execution
//
// DEPRECATED: This method is deprecated as part of queue-driven standardization.
// It now uses the job queue system when available, falling back to synchronous execution only when necessary.
// All new operations should use /api/operations/start endpoint directly.
//
// TODO: Remove this method in v4.0.0 when all clients have migrated to the queue-based API.
func (ps *OperationService) StartOperation(ctx context.Context, params map[string]interface{}) (string, error) {
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "StartOperation"),
	)

	if params == nil {
		params = map[string]interface{}{}
	}

	// Normalize stage identifiers so “start” can be driven by stage_id or step.
	var stageID string
	if v, ok := params["stage_id"].(string); ok && v != "" {
		stageID = v
	} else if v, ok := params["stageId"].(string); ok && v != "" { // alternate casing from UI
		stageID = v
	} else if v, ok := params["step"].(string); ok && v != "" {
		stageID = v
	}

	if stageID != "" {
		params["step"] = stageID

		// If mode is absent or is the placeholder "run_stage", clear it to avoid validation failures.
		if mode, ok := params["mode"].(string); ok {
			if strings.EqualFold(mode, "run_stage") {
				delete(params, "mode")
			}
		}

		// Provide a sensible default scraping mode if none supplied.
		if stageID == operations.StageIDScraping {
			if _, ok := params["mode"]; !ok {
				params["mode"] = "initial"
			}
		}
	} else {
		// If no explicit stage_id/step but we get a scraping mode, map to scraping.
		if mode, ok := params["mode"].(string); ok {
			if mode == "initial" || mode == "incremental" || mode == "backfill" {
				stageID = operations.StageIDScraping
				params["step"] = stageID
				// keep mode as-is (initial/incremental/backfill) for scraper flags
			}
		}
	}

	operationType := "general"
	if stageID != "" {
		operationType = stageID
	} else if mode, ok := params["mode"].(string); ok && mode != "" {
		operationType = mode
	}

	if err := ps.ValidateLicenseForOperation(ctx, operationType); err != nil {
		logger.ErrorContext(ctx, "license validation failed", slog.String("operation_type", operationType), slog.String("error", err.Error()))
		return "", fmt.Errorf("license validation failed for %s operation: %w", operationType, err)
	}

	request := operations.OperationRequest{
		ID:         fmt.Sprintf("operation-%d", time.Now().UnixNano()),
		Mode:       operationType,
		Parameters: params,
	}

	processed, err := ps.ProcessSimplifiedOperationRequest(ctx, request)
	if err != nil {
		return "", err
	}

	resp, err := ps.executionService.ExecuteOperation(ctx, &processed)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// StartScraping starts the scraping step
//
// DEPRECATED: This method is deprecated as part of Phase 3 queue-driven standardization.
// It bypasses the job queue system and creates race conditions with StatusBroadcaster.
// Use /api/operations/start with step="scraping" instead.
//
// TODO: Remove this method in v4.0.0 when all clients have migrated to the queue-based API.
func (ps *OperationService) StartScraping(ctx context.Context, params map[string]interface{}) (string, error) {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "StartScraping"),
	)

	// Perform license validation for scraping operations
	if err := ps.ValidateLicenseForOperation(ctx, "scraping"); err != nil {
		logger.ErrorContext(ctx, "License validation failed for scraping operation",
			slog.String("error", err.Error()))
		return "", fmt.Errorf("license validation failed for scraping operation: %w", err)
	}

	logger.InfoContext(ctx, "License validation passed for scraping operation")

	// Log received parameters for debugging
	if ps.logger != nil {
		logger.Info("StartScraping received params",
			slog.Any("params", params))
	}

	// Extract args from the params structure
	args, ok := params["args"].(map[string]interface{})
	if !ok {
		// Fallback to direct params if no args wrapper
		args = params
		if ps.logger != nil {
			ps.logger.Warn("No 'args' wrapper found, using params directly")
		}
	}

	// Build operation parameters with correct field names
	fromValue := getValue(args, "from", "")
	toValue := getValue(args, "to", "")

	scrapingParams := map[string]interface{}{
		"mode":      getValue(args, "mode", "initial"),
		"from_date": fromValue, // Map 'from' to 'from_date'
		"to_date":   toValue,   // Map 'to' to 'to_date'
		"headless":  getValue(args, "headless", true),
		"step":      "scraping",
	}

	// Log transformed parameters with detailed mapping
	if ps.logger != nil {
		ps.logger.Info("Parameter transformation for scraping",
			slog.String("from_input", fmt.Sprintf("%v", args["from"])),
			slog.String("from_mapped", fromValue.(string)),
			slog.String("to_input", fmt.Sprintf("%v", args["to"])),
			slog.String("to_mapped", toValue.(string)),
			slog.Any("final_params", scrapingParams))
	}

	return ps.StartOperation(ctx, scrapingParams)
}

// ExecuteOperation executes an operation with the given request
func (ps *OperationService) ExecuteOperation(ctx context.Context, request *operations.OperationRequest) (*operations.OperationResponse, error) {
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "ExecuteOperation"),
		slog.String("operation_id", request.ID),
	)

	processedRequest, err := ps.ProcessSimplifiedOperationRequest(ctx, *request)
	if err != nil {
		logger.ErrorContext(ctx, "simplified operation request processing failed",
			slog.String("operation_id", request.ID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("operation request validation failed: %w", err)
	}

	operationType := processedRequest.Mode
	if operationType == "" {
		if step, ok := processedRequest.Parameters["step"].(string); ok && step != "" {
			operationType = step
		} else {
			operationType = "general"
		}
	}

	if err := ps.ValidateLicenseForOperation(ctx, operationType); err != nil {
		logger.ErrorContext(ctx, "license validation failed",
			slog.String("operation_type", operationType),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("license validation failed for operation %s: %w", processedRequest.ID, err)
	}

	return ps.executionService.ExecuteOperation(ctx, &processedRequest)
}

// ProcessSimplifiedOperationRequest converts simplified operation types to the format expected by the manager
func (ps *OperationService) ProcessSimplifiedOperationRequest(ctx context.Context, request operations.OperationRequest) (operations.OperationRequest, error) {
	logger := ps.logger.With(
		slog.String("component", "operations_service"),
		slog.String("method", "processSimplifiedOperationRequest"),
		slog.String("request_id", request.ID),
	)

	// Create a copy to avoid modifying the original
	processedRequest := request

	// Ensure parameters map exists
	if processedRequest.Parameters == nil {
		processedRequest.Parameters = make(map[string]interface{})
	}

	// Normalize legacy scraping modes to the scraping stage
	if processedRequest.Mode == "initial" || processedRequest.Mode == "incremental" || processedRequest.Mode == "backfill" {
		processedRequest.Parameters["mode"] = processedRequest.Mode
		processedRequest.Parameters["step"] = operations.StageIDScraping
		processedRequest.Mode = operations.StageIDScraping

		logger.InfoContext(ctx, "normalizing scraping mode to stage",
			slog.String("normalized_mode", processedRequest.Mode),
			slog.String("scraping_mode", processedRequest.Parameters["mode"].(string)))
	}

	// Handle the new simplified operation types
	switch request.Mode {
	case "run_stage":
		// Validate required stage_id parameter
		stageID, ok := request.Parameters["stage_id"].(string)
		if !ok || stageID == "" {
			logger.ErrorContext(ctx, "run_stage_request_missing_stage_id",
				slog.String("error", "stage_id parameter is required for run_stage operation"),
				slog.String("operation_type", request.Mode))

			return operations.OperationRequest{}, fmt.Errorf("stage_id parameter is required for run_stage operation")
		}

		// Validate stage_id against available stages
		registry := ps.manager.GetRegistry()
		if registry == nil {
			logger.ErrorContext(ctx, "stage_registry_not_available",
				slog.String("error", "operation registry is not initialized"),
				slog.String("stage_id", stageID))

			return operations.OperationRequest{}, fmt.Errorf("operation registry is not available for stage validation")
		}

		// Check if stage exists in registry
		if _, err := registry.Get(stageID); err != nil {
			logger.ErrorContext(ctx, "stage_not_found_in_registry",
				slog.String("stage_id", stageID),
				slog.String("error", err.Error()),
				slog.String("operation_type", request.Mode))

			// Get available stages for better error message
			availableStages := registry.List()
			stageIDs := make([]string, len(availableStages))
			for i, stage := range availableStages {
				stageIDs[i] = stage.ID()
			}

			return operations.OperationRequest{}, fmt.Errorf("invalid stage_id '%s'. Available stages: %v", stageID, stageIDs)
		}

		// Validate stage-specific parameters
		if err := ps.validator.ValidateStageParameters(ctx, stageID, request.Parameters); err != nil {
			logger.ErrorContext(ctx, "stage_parameter_validation_failed",
				slog.String("stage_id", stageID),
				slog.String("error", err.Error()))

			return operations.OperationRequest{}, fmt.Errorf("stage parameter validation failed for '%s': %w", stageID, err)
		}

		// Set stage as step and mode for compatibility
		processedRequest.Parameters["step"] = stageID
		processedRequest.Mode = stageID

		logger.InfoContext(ctx, "converting_run_stage_request",
			slog.String("stage_id", stageID),
			slog.String("operation_mode", processedRequest.Mode),
			slog.String("validation_status", "passed"))

	default:
		// Legacy operation types - validate but leave as-is for backward compatibility
		logger.DebugContext(ctx, "legacy_operation_type_detected",
			slog.String("operation_mode", request.Mode),
			slog.String("action", "preserving_original_request"))
		if request.Mode != "" {
			if err := ps.validator.ValidateStageParameters(ctx, request.Mode, request.Parameters); err != nil {
				logger.WarnContext(ctx, "legacy_stage_parameter_validation_failed",
					slog.String("operation_mode", request.Mode),
					slog.String("error", err.Error()))
			}
		}
	}

	// Process date parameters for backward compatibility
	if from, ok := request.Parameters["from"].(string); ok {
		processedRequest.FromDate = from
	}
	if to, ok := request.Parameters["to"].(string); ok {
		processedRequest.ToDate = to
	}

	// Process mode parameter for backward compatibility
	if mode, ok := request.Parameters["mode"].(string); ok {
		// Only override when the request did not already carry a concrete stage mode.
		// For explicit stage execution (e.g., /api/stages/scraping/execute), keep the
		// stage ID in Mode so the registry resolves correctly.
		if processedRequest.Mode == "" || processedRequest.Mode == "run_stage" {
			processedRequest.Mode = mode
		}
	}

	if err := ps.validator.ValidateOperationRequest(ctx, &processedRequest); err != nil {
		return operations.OperationRequest{}, fmt.Errorf("operation request validation failed: %w", err)
	}

	logger.DebugContext(ctx, "request_processing_completed",
		slog.String("final_mode", processedRequest.Mode),
		slog.String("from_date", processedRequest.FromDate),
		slog.String("to_date", processedRequest.ToDate),
		slog.Int("parameter_count", len(processedRequest.Parameters)),
		slog.String("validation_status", "passed"))

	return processedRequest, nil
}

// GetOperationStatus returns the status of a specific operation
func (ps *OperationService) GetOperationStatus(ctx context.Context, operationID string) (*operations.OperationState, error) {
	state, err := ps.GetStatus(ctx, operationID)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// CancelOperation cancels a running operation
func (ps *OperationService) CancelOperation(ctx context.Context, operationID string) error {
	return ps.StopOperation(ctx, operationID)
}

// StartProcessing starts the processing step
//
// DEPRECATED: This method is deprecated as part of Phase 3 queue-driven standardization.
// It bypasses the job queue system and creates race conditions with StatusBroadcaster.
// Use /api/operations/start with step="processing" instead.
//
// TODO: Remove this method in v4.0.0 when all clients have migrated to the queue-based API.
func (ps *OperationService) StartProcessing(ctx context.Context, params map[string]interface{}) (string, error) {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "StartProcessing"),
	)

	// Perform license validation for processing operations
	if err := ps.ValidateLicenseForOperation(ctx, "processing"); err != nil {
		logger.ErrorContext(ctx, "License validation failed for processing operation",
			slog.String("error", err.Error()))
		return "", fmt.Errorf("license validation failed for processing operation: %w", err)
	}

	logger.InfoContext(ctx, "License validation passed for processing operation")

	// Processing uses default directories, no input_dir needed
	processingParams := map[string]interface{}{
		"step": "processing",
		"mode": getValue(params, "mode", "full"),
	}

	return ps.StartOperation(ctx, processingParams)
}

// StartIndexExtraction starts the index extraction step
//
// DEPRECATED: This method is deprecated as part of Phase 3 queue-driven standardization.
// It bypasses the job queue system and creates race conditions with StatusBroadcaster.
// Use /api/operations/start with step="indices" instead.
//
// TODO: Remove this method in v4.0.0 when all clients have migrated to the queue-based API.
func (ps *OperationService) StartIndexExtraction(ctx context.Context, params map[string]interface{}) (string, error) {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "StartIndexExtraction"),
	)

	// Perform license validation for index extraction operations
	if err := ps.ValidateLicenseForOperation(ctx, "indices"); err != nil {
		logger.ErrorContext(ctx, "License validation failed for index extraction operation",
			slog.String("error", err.Error()))
		return "", fmt.Errorf("license validation failed for index extraction operation: %w", err)
	}

	logger.InfoContext(ctx, "License validation passed for index extraction operation")

	indexParams := map[string]interface{}{
		"step": "indices",
		"mode": "full",
	}

	return ps.StartOperation(ctx, indexParams)
}

// StopOperation stops a running operation
func (ps *OperationService) StopOperation(ctx context.Context, pipelineID string) error {
	if err := ps.manager.CancelOperation(pipelineID); err != nil {
		return fmt.Errorf("failed to stop operation: %w", err)
	}

	if ps.logger != nil {
		ps.logger.Info("operation stopped",
			slog.String("id", pipelineID))
	}
	return nil
}

// GetStatus returns operation status
func (ps *OperationService) GetStatus(ctx context.Context, pipelineID string) (*operations.OperationState, error) {
	if pipelineID == "" {
		return nil, fmt.Errorf("operation ID is required")
	}

	state, err := ps.manager.GetOperation(pipelineID)
	if err != nil {
		return nil, fmt.Errorf("operation not found: %w", err)
	}

	return state, nil
}

// ListOperations returns all operations
func (ps *OperationService) ListOperations(ctx context.Context) ([]*operations.OperationState, error) {
	states := ps.manager.ListOperations()
	return states, nil
}

// ListOperationsByStatus returns operations filtered by status
func (ps *OperationService) ListOperationsByStatus(ctx context.Context, status operations.OperationStatusValue) ([]*operations.OperationState, error) {
	states := ps.manager.ListOperations()
	var result []*operations.OperationState
	for _, state := range states {
		if state.Status == status {
			result = append(result, state)
		}
	}
	return result, nil
}

// GetOperationTypes returns the simplified operation types: run_stage and single_stage_placeholder
func (ps *OperationService) GetOperationTypes(ctx context.Context) ([]operations.OperationType, error) {
	return ps.configService.GetOperationTypes(ctx)
}

// CancelAll stops all running operations
func (ps *OperationService) CancelAll(ctx context.Context) error {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "CancelAll"),
	)

	// Perform basic license validation before allowing operation cancellation
	// This is a management operation, so general license validation is sufficient
	if err := ps.ValidateLicenseForOperation(ctx, "general"); err != nil {
		logger.ErrorContext(ctx, "License validation failed for cancel all operations",
			slog.String("error", err.Error()))
		return fmt.Errorf("license validation failed for cancel all operations: %w", err)
	}

	logger.InfoContext(ctx, "License validation passed - cancelling all operations")

	ops := ps.manager.ListOperations()
	cancelledCount := 0
	failedCount := 0

	for _, p := range ops {
		if p.Status == operations.OperationStatusRunning {
			if err := ps.manager.CancelOperation(p.ID); err != nil {
				if ps.logger != nil {
					logger.Error("Failed to cancel operation",
						slog.String("id", p.ID),
						slog.String("error", err.Error()))
				}
				failedCount++
			} else {
				cancelledCount++
			}
		}
	}

	logger.InfoContext(ctx, "Cancel all operations completed",
		slog.Int("total_operations", len(ops)),
		slog.Int("cancelled", cancelledCount),
		slog.Int("failed", failedCount))

	// Only return error if any cancellations failed
	if failedCount > 0 {
		return fmt.Errorf("failed to cancel %d operations", failedCount)
	}

	return nil
}

// ExecuteStage executes a specific step
func (ps *OperationService) ExecuteStage(stageID string, ctx context.Context) error {
	if ps.logger != nil {
		ps.logger.Info("Executing individual stage",
			slog.String("stage_id", stageID))
	}

	// Get the stage from registry
	registry := ps.manager.GetRegistry()
	stage, err := registry.Get(stageID)
	if err != nil {
		return fmt.Errorf("stage not found: %s - %w", stageID, err)
	}

	// Create a new operation for individual stage execution
	operationID := fmt.Sprintf("stage-%s-%d", stageID, time.Now().Unix())

	// Create operation state
	state := operations.NewOperationState(operationID)

	// Initialize stage state
	stepState := operations.NewStepState(stage.ID(), stage.Name())
	state.SetStage(stage.ID(), stepState)
	state.Start()

	// Set up broadcaster from manager
	broadcaster := ps.manager.GetBroadcaster()
	if broadcaster != nil {
		state.SetBroadcaster(broadcaster)
	}

	if ps.logger != nil {
		ps.logger.Info("Created operation state for individual stage execution",
			slog.String("operation_id", operationID),
			slog.String("stage_id", stageID),
			slog.String("stage_name", stage.Name()))
	}

	// Execute the stage
	if err := stage.Execute(ctx, state); err != nil {
		state.Fail(err)
		return fmt.Errorf("stage execution failed: %w", err)
	}

	// Mark as completed
	state.Complete()

	if ps.logger != nil {
		ps.logger.Info("Individual stage execution completed successfully",
			slog.String("operation_id", operationID),
			slog.String("stage_id", stageID))
	}

	return nil
}

// ValidateExecutables checks if required executables exist
func (ps *OperationService) ValidateExecutables(ctx context.Context) error {
	executables := []string{
		"scraper.exe",
		"process.exe",
		"indexcsv.exe",
	}

	for _, exe := range executables {
		path := filepath.Join(ps.paths.ExecutableDir, exe)
		if ps.logger != nil {
			ps.logger.Debug("Checking for executable",
				slog.String("exe", exe),
				slog.String("path", path))
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			if ps.logger != nil {
				ps.logger.Error("Required executable not found",
					slog.String("exe", exe),
					slog.String("path", path))
			}
			return fmt.Errorf("required executable not found: %s at %s", exe, path)
		}

		if ps.logger != nil {
			ps.logger.Info("Found required executable",
				slog.String("exe", exe),
				slog.String("path", path))
		}
	}

	return nil
}

// GetManager returns the underlying operation manager
func (ps *OperationService) GetManager() *operations.Manager {
	return ps.manager
}

// GetLicenseStatus returns the current license status for operations
func (ps *OperationService) GetLicenseStatus(ctx context.Context) (map[string]interface{}, error) {
	// Ensure context has trace ID for audit logging
	ctx = infrastructure.EnsureTraceID(ctx)
	traceID := infrastructure.GetTraceID(ctx)

	logger := ps.logger.With(
		slog.String("trace_id", traceID),
		slog.String("component", "operations_service"),
		slog.String("method", "GetLicenseStatus"),
	)

	// Get license status from license manager
	info, status, err := ps.licenseManager.GetLicenseStatus()
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get license status",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get license status: %w", err)
	}

	// Build response with license information relevant to operations
	response := map[string]interface{}{
		"license_status":       status,
		"valid_for_operations": ps.isLicenseValidForOperations(status),
		"trace_id":             traceID,
		"timestamp":            time.Now().Format(time.RFC3339),
	}

	// Add additional license information if available
	if info != nil {
		response["expiry_date"] = info.ExpiryDate.Format("2006-01-02")
		response["days_remaining"] = int(info.ExpiryDate.Sub(time.Now()).Hours() / 24)
		response["duration"] = info.Duration
		response["user_email"] = info.UserEmail
		response["activation_id"] = info.ActivationID

		// Add operation-specific restrictions based on license status
		response["operation_restrictions"] = ps.getOperationRestrictions(status, info)
	}

	logger.InfoContext(ctx, "License status retrieved successfully",
		slog.String("license_status", status),
		slog.Bool("valid_for_operations", response["valid_for_operations"].(bool)))

	return response, nil
}

// isLicenseValidForOperations checks if license status is valid for operations
func (ps *OperationService) isLicenseValidForOperations(status string) bool {
	switch status {
	case "Active", "Activated", "Valid", "Warning", "Critical":
		return true
	case "Not Activated", "Expired":
		return false
	default:
		// Unknown status - be conservative and deny operations
		return false
	}
}

// getOperationRestrictions returns operation restrictions based on license status
func (ps *OperationService) getOperationRestrictions(status string, info *license.LicenseInfo) map[string]interface{} {
	restrictions := map[string]interface{}{
		"allow_scraping":                 false,
		"allow_processing":               false,
		"allow_indices":                  false,
		"allow_liquidity":                false,
		"allow_indicators":               false,
		"allow_upload":                   false,
		"allow_single_stage_placeholder": false,
		"allow_management":               false, // cancel, list operations, etc.
	}

	switch status {
	case "Active", "Activated", "Valid":
		// Full access for active licenses
		for key := range restrictions {
			restrictions[key] = true
		}
	case "Warning":
		// Warning status - still full access but with warnings
		for key := range restrictions {
			restrictions[key] = true
		}
		restrictions["warning_message"] = "License expires soon. Consider renewal to avoid interruption."
	case "Critical":
		// Critical status - full access but with strong warnings
		for key := range restrictions {
			restrictions[key] = true
		}
		restrictions["critical_warning"] = "License expires very soon. Immediate renewal recommended."
	case "Not Activated":
		// Only allow license management operations
		restrictions["allow_management"] = true
		restrictions["required_action"] = "activate_license"
	case "Expired":
		// Read-only access for expired licenses
		restrictions["allow_management"] = true // to check status
		restrictions["required_action"] = "renew_license"
	}

	return restrictions
}

// GetStageSummary returns summary information about available steps
func (ps *OperationService) GetStageSummary() map[string]interface{} {
	return ps.configService.GetStageSummary()
}

// getValue safely extracts a value from a map with a default
func getValue(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := m[key]; ok && val != nil {
		return val
	}
	return defaultValue
}

// GetOperationMetrics returns metrics about operations
func (ps *OperationService) GetOperationMetrics(ctx context.Context) (map[string]interface{}, error) {
	// Get basic metrics - simplified implementation for Phase 1
	operations, err := ps.ListOperations(ctx)
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

	if ps.logger != nil {
		ps.logger.DebugContext(ctx, "Retrieved operation metrics",
			slog.Int("total", len(operations)),
			slog.Int("active", activeCount))
	}

	return metrics, nil
}

// NEW: GetStageInfo returns detailed stage information from registry
func (s *OperationService) GetStageInfo(ctx context.Context, stageID string) (*operations.StageDescriptor, error) {
	s.logger.InfoContext(ctx, "getting_stage_info",
		slog.String("stage_id", stageID),
		slog.String("component", "operations_service"),
		slog.String("method", "GetStageInfo"))

	// Get stage from registry
	stage, err := s.manager.GetRegistry().Get(stageID)
	if err != nil {
		return nil, fmt.Errorf("stage not found: %w", err)
	}

	// Create basic stage descriptor from registry stage
	descriptor := &operations.StageDescriptor{
		StageID:      stage.ID(),
		StageName:    stage.Name(),
		Description:  fmt.Sprintf("Stage %s for data processing", stage.Name()),
		Dependencies: []string{},
		Parameters:   make(map[string]interface{}),
		Required:     true,
		Timeout:      30 * time.Minute,
	}

	s.logger.InfoContext(ctx, "stage_info_created_from_registry",
		slog.String("stage_id", stageID))

	return descriptor, nil
}

// NEW: HealthCheck performs comprehensive health check of the service
func (s *OperationService) HealthCheck(ctx context.Context) error {
	s.logger.InfoContext(ctx, "service_health_check",
		slog.String("component", "operations_service"),
		slog.String("method", "HealthCheck"))

	// Check manager health
	if s.manager == nil {
		return fmt.Errorf("operation manager is not initialized")
	}

	// NOTE: Orchestrator system removed for simplicity - no health check needed

	// Check license manager
	if s.licenseManager == nil {
		return fmt.Errorf("license manager is not initialized")
	}

	// Validate license
	if err := s.licenseManager.CheckLicense(); err != nil {
		s.logger.WarnContext(ctx, "license_check_failed_during_health_check",
			slog.String("error", err.Error()))
		// Don't fail health check for license issues, just log them
	}

	// Validate paths
	if s.paths == nil {
		return fmt.Errorf("paths configuration is not initialized")
	}

	// Check critical directories exist
	if _, err := os.Stat(s.paths.ExecutableDir); os.IsNotExist(err) {
		return fmt.Errorf("executable directory does not exist: %s", s.paths.ExecutableDir)
	}

	if _, err := os.Stat(s.paths.DataDir); os.IsNotExist(err) {
		return fmt.Errorf("data directory does not exist: %s", s.paths.DataDir)
	}

	s.logger.InfoContext(ctx, "service_health_check_passed",
		slog.String("executable_dir", s.paths.ExecutableDir),
		slog.String("data_dir", s.paths.DataDir))

	return nil
}
