package operations

import (
	"sync"
	"time"
)

// OperationStatusValue represents the overall operation status enum
type OperationStatusValue string

// OperationStatus is an alias for OperationStatusValue for backward compatibility
type OperationStatus = OperationStatusValue

const (
	OperationStatusPending   OperationStatusValue = "pending"
	OperationStatusRunning   OperationStatusValue = "running"
	OperationStatusCompleted OperationStatusValue = "completed"
	OperationStatusFailed    OperationStatusValue = "failed"
	OperationStatusCancelled OperationStatusValue = "cancelled"
)

// OperationState represents the complete state of a operation execution
type OperationState struct {
	mu sync.RWMutex

	// Basic operation information
	ID        string               `json:"id"`
	Status    OperationStatusValue `json:"status"`
	StartTime time.Time            `json:"start_time"`
	EndTime   *time.Time           `json:"end_time,omitempty"`

	// Step states
	Steps map[string]*StepState `json:"steps"`

	// operation context for passing data between steps
	Context map[string]interface{} `json:"context"`

	// Configuration passed from the request
	Config map[string]interface{} `json:"config"`

	// Error if operation failed
	Error error `json:"error,omitempty"`

	// Broadcaster for real-time progress updates (not serialized)
	broadcaster *StatusBroadcaster `json:"-"`
}

// NewOperationState creates a new operation state
func NewOperationState(id string) *OperationState {
	return &OperationState{
		ID:        id,
		Status:    OperationStatusPending,
		StartTime: time.Now(),
		Steps:     make(map[string]*StepState),
		Context:   make(map[string]interface{}),
		Config:    make(map[string]interface{}),
	}
}

// Start marks the operation as running
func (p *OperationState) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = OperationStatusRunning
	p.StartTime = time.Now()
}

// Complete marks the operation as completed
func (p *OperationState) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	p.EndTime = &now
	p.Status = OperationStatusCompleted
}

// Fail marks the operation as failed
func (p *OperationState) Fail(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	p.EndTime = &now
	p.Status = OperationStatusFailed
	p.Error = err
}

// Cancel marks the operation as cancelled
func (p *OperationState) Cancel() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	p.EndTime = &now
	p.Status = OperationStatusCancelled
}

// GetStage returns the state of a specific Step and auto-creates it if missing.
func (p *OperationState) GetStage(stageID string) *StepState {
	if stageID == "" {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// DEBUG: Log stepState access for debugging
	if len(p.Steps) == 0 {
		// This could indicate a problem with Steps initialization
		// Create a new Steps map if it's nil or empty
		if p.Steps == nil {
			p.Steps = make(map[string]*StepState)
		}
	}

	if step, exists := p.Steps[stageID]; exists && step != nil {
		return step
	}

	// Create new stepState since it doesn't exist
	step := NewStepState(stageID, stageNameForID(stageID))
	if p.broadcaster != nil {
		step.broadcaster = p.broadcaster
		step.operationID = p.ID
	}

	p.Steps[stageID] = step
	return step
}

func stageNameForID(stageID string) string {
	switch stageID {
	case StageIDScraping:
		return StageNameScraping
	case StageIDProcessing:
		return StageNameProcessing
	case StageIDIndices:
		return StageNameIndices
	case StageIDLiquidity:
		return StageNameLiquidity
	default:
		return stageID
	}
}

// SetStage updates the state of a specific Step
func (p *OperationState) SetStage(stageID string, state *StepState) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Set broadcaster and operation ID for real-time updates
	if p.broadcaster != nil {
		state.broadcaster = p.broadcaster
		state.operationID = p.ID
	}

	p.Steps[stageID] = state
}

// SetBroadcaster sets the broadcaster for real-time progress updates
func (p *OperationState) SetBroadcaster(broadcaster *StatusBroadcaster) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.broadcaster = broadcaster

	// Update all existing step states
	for _, stepState := range p.Steps {
		stepState.broadcaster = broadcaster
		stepState.operationID = p.ID
	}
}

// GetContext retrieves a value from the operation context
func (p *OperationState) GetContext(key string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.Context[key]
	return val, ok
}

// SetContext sets a value in the operation context
func (p *OperationState) SetContext(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Context[key] = value
}

// GetConfig retrieves a configuration value
func (p *OperationState) GetConfig(key string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.Config[key]
	return val, ok
}

// SetConfig sets a configuration value
func (p *OperationState) SetConfig(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Config[key] = value
}

// Duration returns the duration of the operation execution
func (p *OperationState) Duration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.EndTime != nil {
		return p.EndTime.Sub(p.StartTime)
	}
	return time.Since(p.StartTime)
}

// GetActiveStages returns all currently active steps
func (p *OperationState) GetActiveStages() []*StepState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var active []*StepState
	for _, Step := range p.Steps {
		if Step.Status == StepStatusActive {
			active = append(active, Step)
		}
	}
	return active
}

// GetCompletedStages returns all completed steps
func (p *OperationState) GetCompletedStages() []*StepState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var completed []*StepState
	for _, Step := range p.Steps {
		if Step.Status == StepStatusCompleted {
			completed = append(completed, Step)
		}
	}
	return completed
}

// GetFailedStages returns all failed steps
func (p *OperationState) GetFailedStages() []*StepState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var failed []*StepState
	for _, Step := range p.Steps {
		if Step.Status == StepStatusFailed {
			failed = append(failed, Step)
		}
	}
	return failed
}

// IsComplete returns true if all steps are completed or skipped
func (p *OperationState) IsComplete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, Step := range p.Steps {
		if Step.Status == StepStatusPending || Step.Status == StepStatusActive {
			return false
		}
	}
	return true
}

// HasFailures returns true if any Step has failed
func (p *OperationState) HasFailures() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, Step := range p.Steps {
		if Step.Status == StepStatusFailed {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the operation state
func (p *OperationState) Clone() *OperationState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	clone := &OperationState{
		ID:        p.ID,
		Status:    p.Status,
		StartTime: p.StartTime,
		Steps:     make(map[string]*StepState),
		Context:   make(map[string]interface{}),
		Config:    make(map[string]interface{}),
		Error:     p.Error,
	}

	if p.EndTime != nil {
		endTime := *p.EndTime
		clone.EndTime = &endTime
	}

	// Clone steps
	for k, v := range p.Steps {
		v.mu.RLock()
		stageCopy := &StepState{
			ID:        v.ID,
			Name:      v.Name,
			Status:    v.Status,
			StartTime: v.StartTime,
			EndTime:   v.EndTime,
			Progress:  v.Progress,
			Message:   v.Message,
			Error:     v.Error,
			Metadata:  make(map[string]interface{}),
		}
		// Copy metadata
		for mk, mv := range v.Metadata {
			stageCopy.Metadata[mk] = mv
		}
		v.mu.RUnlock()
		clone.Steps[k] = stageCopy
	}

	// Clone context
	for k, v := range p.Context {
		clone.Context[k] = v
	}

	// Clone config
	for k, v := range p.Config {
		clone.Config[k] = v
	}

	return clone
}
