package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Manager orchestrates pipeline execution
type Manager struct {
	registry *Registry
	config   *Config
	hub      WebSocketHub
	logger   Logger

	// Active pipelines
	mu        sync.RWMutex
	pipelines map[string]*PipelineState
}

// NewManager creates a new pipeline manager
func NewManager(hub WebSocketHub, logger Logger) *Manager {
	return &Manager{
		registry:  NewRegistry(),
		config:    DefaultConfig(),
		hub:       hub,
		logger:    logger,
		pipelines: make(map[string]*PipelineState),
	}
}

// RegisterStage registers a stage with the pipeline
func (m *Manager) RegisterStage(stage Stage) error {
	return m.registry.Register(stage)
}

// SetConfig updates the pipeline configuration
func (m *Manager) SetConfig(config *Config) {
	if config != nil {
		m.config = config
	}
}

// Execute runs a pipeline with the given request
func (m *Manager) Execute(ctx context.Context, req PipelineRequest) (*PipelineResponse, error) {
	// Generate pipeline ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("pipeline-%d", time.Now().Unix())
	}

	// Create pipeline state
	state := NewPipelineState(req.ID)
	
	// Set configuration from request
	if req.FromDate != "" {
		state.SetConfig(ContextKeyFromDate, req.FromDate)
	}
	if req.ToDate != "" {
		state.SetConfig(ContextKeyToDate, req.ToDate)
	}
	if req.Mode != "" {
		state.SetConfig(ContextKeyMode, req.Mode)
	}
	
	// Copy additional parameters
	for k, v := range req.Parameters {
		state.SetConfig(k, v)
	}

	// Store pipeline state
	m.storePipeline(state)
	defer m.removePipeline(req.ID)

	// Send pipeline reset event
	m.sendWebSocketUpdate(EventTypePipelineReset, state)

	// Get stages in dependency order
	stages, err := m.registry.GetDependencyOrder()
	if err != nil {
		m.logger.Error("Failed to get dependency order: %v", err)
		state.Fail(err)
		return m.createResponse(state), err
	}

	// Initialize stage states
	for _, stage := range stages {
		stageState := NewStageState(stage.ID(), stage.Name())
		state.SetStage(stage.ID(), stageState)
	}

	// Start pipeline execution
	state.Start()
	m.sendWebSocketUpdate(EventTypePipelineStatus, state)

	// Execute stages based on execution mode
	if m.config.ExecutionMode == ExecutionModeSequential {
		err = m.executeSequential(ctx, state, stages)
	} else {
		err = m.executeParallel(ctx, state, stages)
	}

	// Update final pipeline state
	if err != nil {
		state.Fail(err)
		m.sendWebSocketUpdate(EventTypePipelineError, state)
	} else {
		state.Complete()
		m.sendWebSocketUpdate(EventTypePipelineComplete, state)
	}

	return m.createResponse(state), err
}

// executeSequential executes stages one by one
func (m *Manager) executeSequential(ctx context.Context, state *PipelineState, stages []Stage) error {
	m.logger.Info("Starting sequential execution of %d stages", len(stages))
	for i, stage := range stages {
		select {
		case <-ctx.Done():
			m.logger.Warn("Pipeline cancelled at stage %s", stage.ID())
			return NewCancellationError(stage.ID())
		default:
			// Check if stage should be skipped due to failed dependencies
			stageState := state.GetStage(stage.ID())
			if stageState != nil && stageState.Status == StageStatusSkipped {
				m.logger.Info("Skipping stage %s (%d/%d) - marked as skipped", stage.ID(), i+1, len(stages))
				continue
			}
			
			m.logger.Info("Executing stage %s (%d/%d)", stage.ID(), i+1, len(stages))
			if err := m.executeStage(ctx, state, stage); err != nil {
				m.logger.Error("Stage %s failed: %v", stage.ID(), err)
				if !m.config.ContinueOnError {
					// Skip all dependent stages
					m.skipDependentStages(state, stages, stage.ID())
					return err
				}
				m.logger.Warn("Stage %s failed but continuing: %v", stage.ID(), err)
			} else {
				m.logger.Info("Stage %s completed successfully", stage.ID())
			}
		}
	}
	m.logger.Info("All stages completed")
	return nil
}

// executeParallel executes independent stages in parallel
func (m *Manager) executeParallel(ctx context.Context, state *PipelineState, stages []Stage) error {
	// TODO: Implement parallel execution
	// For now, fall back to sequential
	return m.executeSequential(ctx, state, stages)
}

// executeStage executes a single stage with retry logic
func (m *Manager) executeStage(ctx context.Context, pipelineState *PipelineState, stage Stage) error {
	m.logger.Info("Starting execution of stage %s", stage.ID())
	stageState := pipelineState.GetStage(stage.ID())
	if stageState == nil {
		m.logger.Error("Stage state not found for %s", stage.ID())
		return NewFatalError("Stage state not found", nil)
	}

	// Check dependencies
	m.logger.Debug("Checking dependencies for stage %s", stage.ID())
	if err := m.checkDependencies(pipelineState, stage); err != nil {
		m.logger.Warn("Dependencies not met for stage %s: %v", stage.ID(), err)
		stageState.Skip(fmt.Sprintf("Dependencies not met: %v", err))
		m.sendStageUpdate(pipelineState, stageState)
		return err
	}

	// Validate stage
	m.logger.Debug("Validating stage %s", stage.ID())
	if err := stage.Validate(pipelineState); err != nil {
		m.logger.Warn("Validation failed for stage %s: %v", stage.ID(), err)
		stageState.Skip(fmt.Sprintf("Validation failed: %v", err))
		m.sendStageUpdate(pipelineState, stageState)
		return NewValidationError(stage.ID(), err.Error())
	}

	// Get stage timeout
	timeout := m.config.GetStageTimeout(stage.ID())
	stageCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute with retries
	retryConfig := m.config.RetryConfig
	var lastErr error
	
	for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
		// Start stage
		stageState.Start()
		m.sendStageUpdate(pipelineState, stageState)

		// Execute stage
		m.logger.Info("Calling Execute for stage %s (attempt %d)", stage.ID(), attempt)
		startTime := time.Now()
		err := stage.Execute(stageCtx, pipelineState)
		duration := time.Since(startTime)
		
		if err == nil {
			// Success
			m.logger.Info("Stage %s executed successfully in %v", stage.ID(), duration)
			stageState.Complete()
			m.sendStageUpdate(pipelineState, stageState)
			return nil
		}
		
		m.logger.Error("Stage %s execution failed after %v: %v", stage.ID(), duration, err)
		
		// Log stage metadata for debugging
		if stageState.Metadata != nil {
			if metaJSON, err := json.Marshal(stageState.Metadata); err == nil {
				m.logger.Error("Stage %s metadata: %s", stage.ID(), string(metaJSON))
			}
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) || attempt >= retryConfig.MaxAttempts {
			stageState.Fail(err)
			m.sendStageUpdate(pipelineState, stageState)
			return WrapError(err, stage.ID(), "Stage execution failed")
		}

		// Calculate retry delay
		delay := m.calculateRetryDelay(attempt, retryConfig)
		m.logger.Warn("Stage %s failed (attempt %d/%d), retrying in %v: %v", 
			stage.ID(), attempt, retryConfig.MaxAttempts, delay, err)

		// Wait before retry
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-stageCtx.Done():
			stageState.Fail(NewTimeoutError(stage.ID(), timeout.String()))
			m.sendStageUpdate(pipelineState, stageState)
			return NewTimeoutError(stage.ID(), timeout.String())
		}
	}

	// All retries exhausted
	stageState.Fail(lastErr)
	m.sendStageUpdate(pipelineState, stageState)
	return WrapError(lastErr, stage.ID(), "Stage execution failed after retries")
}

// skipDependentStages marks all stages that depend on the failed stage as skipped
func (m *Manager) skipDependentStages(state *PipelineState, stages []Stage, failedStageID string) {
	for _, stage := range stages {
		deps := stage.GetDependencies()
		for _, dep := range deps {
			if dep == failedStageID {
				stageState := state.GetStage(stage.ID())
				if stageState != nil && stageState.Status == StageStatusPending {
					stageState.Skip(fmt.Sprintf("Dependency %s failed", failedStageID))
					m.sendStageUpdate(state, stageState)
					// Recursively skip stages that depend on this one
					m.skipDependentStages(state, stages, stage.ID())
				}
				break
			}
		}
	}
}

// checkDependencies verifies that all dependencies are satisfied
func (m *Manager) checkDependencies(state *PipelineState, stage Stage) error {
	deps := stage.GetDependencies()
	for _, dep := range deps {
		depState := state.GetStage(dep)
		if depState == nil {
			return fmt.Errorf("dependency %s not found", dep)
		}
		if depState.Status != StageStatusCompleted {
			return fmt.Errorf("dependency %s not completed (status: %s)", dep, depState.Status)
		}
	}
	return nil
}

// calculateRetryDelay calculates the delay before next retry
func (m *Manager) calculateRetryDelay(attempt int, config RetryConfig) time.Duration {
	delay := config.InitialDelay * time.Duration(float64(attempt-1)*config.Multiplier)
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	return delay
}

// sendStageUpdate sends a WebSocket update for a stage
func (m *Manager) sendStageUpdate(pipelineState *PipelineState, stageState *StageState) {
	update := map[string]interface{}{
		"pipeline_id": pipelineState.ID,
		"stage":       stageState.ID,
		"status":      stageState.Status,
		"progress":    stageState.Progress,
		"message":     stageState.Message,
	}

	if stageState.Metadata != nil {
		update["metadata"] = stageState.Metadata
	}

	// Send pipeline_progress for progress updates
	m.hub.BroadcastUpdate(EventTypePipelineProgress, "", "", update)
	
	// Also send pipeline_status when stage completes or fails
	if stageState.Status == StageStatusCompleted || stageState.Status == StageStatusFailed {
		m.hub.BroadcastUpdate(EventTypePipelineStatus, "", "", map[string]interface{}{
			"pipeline_id": pipelineState.ID,
			"stage":       stageState.ID,
			"status":      string(stageState.Status),
			"message":     stageState.Message,
		})
	}
}

// sendWebSocketUpdate sends a pipeline-level WebSocket update
func (m *Manager) sendWebSocketUpdate(eventType string, state *PipelineState) {
	update := map[string]interface{}{
		"pipeline_id": state.ID,
		"status":      state.Status,
		"stages":      state.Stages,
	}

	if eventType == EventTypePipelineError && state.Error != nil {
		update["error"] = state.Error.Error()
	}

	m.hub.BroadcastUpdate(eventType, "", "", update)
}

// createResponse creates a pipeline response from state
func (m *Manager) createResponse(state *PipelineState) *PipelineResponse {
	resp := &PipelineResponse{
		ID:       state.ID,
		Status:   state.Status,
		Duration: state.Duration(),
		Stages:   state.Stages,
	}

	if state.Error != nil {
		resp.Error = state.Error.Error()
	}

	return resp
}

// GetPipeline retrieves the state of a running pipeline
func (m *Manager) GetPipeline(id string) (*PipelineState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("pipeline %s not found", id)
	}

	return state.Clone(), nil
}

// ListPipelines returns all active pipelines
func (m *Manager) ListPipelines() []*PipelineState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pipelines := make([]*PipelineState, 0, len(m.pipelines))
	for _, state := range m.pipelines {
		pipelines = append(pipelines, state.Clone())
	}

	return pipelines
}

// CancelPipeline cancels a running pipeline
func (m *Manager) CancelPipeline(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.pipelines[id]
	if !exists {
		return fmt.Errorf("pipeline %s not found", id)
	}

	state.Cancel()
	m.sendWebSocketUpdate(EventTypePipelineStatus, state)
	return nil
}

// storePipeline stores a pipeline state
func (m *Manager) storePipeline(state *PipelineState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pipelines[state.ID] = state
}

// removePipeline removes a pipeline state
func (m *Manager) removePipeline(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pipelines, id)
}

// GetRegistry returns the stage registry
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}