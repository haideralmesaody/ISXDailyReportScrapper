package operations

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Manager orchestrates operation execution
type Manager struct {
	registry    *Registry
	config      *Config
	hub         WebSocketHub
	broadcaster *StatusBroadcaster
	jobQueue    *JobQueue // NEW: Store jobqueue reference

	// Active operations
	mu         sync.RWMutex
	operations map[string]*OperationState
}

// NewManager creates a new operation manager with dependency injection
func NewManager(hub WebSocketHub, registry *Registry, config *Config) *Manager {
	if registry == nil {
		registry = NewRegistry()
	}
	if config == nil {
		config = NewConfig()
	}

	// Create status broadcaster for centralized status management
	broadcaster := NewStatusBroadcaster(hub, slog.Default())

	return &Manager{
		registry:    registry,
		config:      config,
		hub:         hub,
		broadcaster: broadcaster,
		jobQueue:    nil, // Will be set later with SetJobQueue
		operations:  make(map[string]*OperationState),
	}
}

// SetJobQueue sets the job queue for unified execution
func (m *Manager) SetJobQueue(jobQueue *JobQueue) {
	m.jobQueue = jobQueue
	slog.Info("jobqueue_set_for_manager",
		"component", "manager",
		"jobqueue_available", jobQueue != nil)
}

// GetJobQueue returns the job queue
func (m *Manager) GetJobQueue() *JobQueue {
	return m.jobQueue
}

// RegisterStage registers a Step with the operation
func (m *Manager) RegisterStage(Step Step) error {
	return m.registry.Register(Step)
}

// SetConfig updates the operation configuration
func (m *Manager) SetConfig(config *Config) {
	if config != nil {
		m.config = config
	}
}

// GetRegistry returns the registry for accessing registered stages
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// GetHub exposes the underlying WebSocket hub for dependency injection scenarios.
func (m *Manager) GetHub() WebSocketHub {
	return m.hub
}

// GetBroadcaster returns the status broadcaster for centralized status updates
func (m *Manager) GetBroadcaster() *StatusBroadcaster {
	return m.broadcaster
}

// Execute runs a operation with the given request
func (m *Manager) Execute(ctx context.Context, req OperationRequest) (*OperationResponse, error) {
	if m.jobQueue == nil {
		return nil, fmt.Errorf("jobqueue not configured for manager")
	}

	// Generate operation ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("operation-%d", time.Now().Unix())
	}

	// Validate requested stage if specified
	if stepParam, ok := req.Parameters["step"].(string); ok && stepParam != "" {
		if _, err := m.registry.Get(stepParam); err != nil {
			return nil, fmt.Errorf("requested stage not found: %s", stepParam)
		}
	}

	job, err := m.jobQueue.EnqueueOperation(req)
	if err != nil {
		return nil, err
	}

	operationType := "run_stage"

	resp := &OperationResponse{
		ID:          req.ID,
		OperationID: req.ID,
		Status:      string(OperationStatusPending),
		Data: map[string]interface{}{
			"operation_type": operationType,
			"stage_id":       job.StageID,
			"stage_name":     job.StageName,
		},
	}

	return resp, nil
}

// All WebSocket updates now go through StatusBroadcaster - single source of truth
// Direct hub broadcasts have been removed to simplify data flow

// convertStepsToArray converts the steps map to an array for the response
func convertStepsToArray(steps map[string]*StepState) []interface{} {
	result := make([]interface{}, 0, len(steps))
	for _, step := range steps {
		result = append(result, step)
	}
	return result
}

// createResponse creates a operation response from state
func (m *Manager) createResponse(state *OperationState) *OperationResponse {
	resp := &OperationResponse{
		ID:       state.ID,
		Status:   string(state.Status),
		Duration: state.Duration(),
		Steps:    convertStepsToArray(state.Steps),
		Data:     make(map[string]interface{}),
	}

	if state.Error != nil {
		resp.Error = state.Error.Error()
	}

	// Add stage information to response data for individual stage operations
	for _, step := range state.Steps {
		resp.Data["stage_id"] = step.ID
		resp.Data["stage_name"] = step.Name
		resp.Data["operation_type"] = step.ID
		break
	}

	return resp
}

// GetOperation retrieves the state of a running operation
func (m *Manager) GetOperation(id string) (*OperationState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.operations[id]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", id)
	}

	return state.Clone(), nil
}

// ListOperations returns all active operations
func (m *Manager) ListOperations() []*OperationState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	operations := make([]*OperationState, 0, len(m.operations))
	for _, state := range m.operations {
		operations = append(operations, state.Clone())
	}

	return operations
}

// CancelOperation cancels a running operation
func (m *Manager) CancelOperation(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.operations[id]
	if !exists {
		return fmt.Errorf("operation %s not found", id)
	}

	state.Cancel()
	m.broadcaster.FailOperation(id, fmt.Errorf("operation cancelled by user"))
	return nil
}

// storeOperation stores a operation state
func (m *Manager) storeOperation(state *OperationState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.operations[state.ID] = state
}

// removeOperation removes a operation state
func (m *Manager) removeOperation(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.operations, id)
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// determineOperationType determines the operation type from the request parameters
func determineOperationType(req OperationRequest) string {
	// Always treat operations as single-stage run_stage
	return "run_stage"
}
