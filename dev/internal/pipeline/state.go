package pipeline

import (
	"sync"
	"time"
)

// PipelineStatus represents the overall pipeline status
type PipelineStatus string

const (
	PipelineStatusPending   PipelineStatus = "pending"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusCompleted PipelineStatus = "completed"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusCancelled PipelineStatus = "cancelled"
)

// PipelineState represents the complete state of a pipeline execution
type PipelineState struct {
	mu sync.RWMutex

	// Basic pipeline information
	ID        string         `json:"id"`
	Status    PipelineStatus `json:"status"`
	StartTime time.Time      `json:"start_time"`
	EndTime   *time.Time     `json:"end_time,omitempty"`

	// Stage states
	Stages map[string]*StageState `json:"stages"`

	// Pipeline context for passing data between stages
	Context map[string]interface{} `json:"context"`

	// Configuration passed from the request
	Config map[string]interface{} `json:"config"`

	// Error if pipeline failed
	Error error `json:"error,omitempty"`
}

// NewPipelineState creates a new pipeline state
func NewPipelineState(id string) *PipelineState {
	return &PipelineState{
		ID:        id,
		Status:    PipelineStatusPending,
		StartTime: time.Now(),
		Stages:    make(map[string]*StageState),
		Context:   make(map[string]interface{}),
		Config:    make(map[string]interface{}),
	}
}

// Start marks the pipeline as running
func (p *PipelineState) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = PipelineStatusRunning
	p.StartTime = time.Now()
}

// Complete marks the pipeline as completed
func (p *PipelineState) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	p.EndTime = &now
	p.Status = PipelineStatusCompleted
}

// Fail marks the pipeline as failed
func (p *PipelineState) Fail(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	p.EndTime = &now
	p.Status = PipelineStatusFailed
	p.Error = err
}

// Cancel marks the pipeline as cancelled
func (p *PipelineState) Cancel() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	p.EndTime = &now
	p.Status = PipelineStatusCancelled
}

// GetStage returns the state of a specific stage
func (p *PipelineState) GetStage(stageID string) *StageState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Stages[stageID]
}

// SetStage updates the state of a specific stage
func (p *PipelineState) SetStage(stageID string, state *StageState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Stages[stageID] = state
}

// GetContext retrieves a value from the pipeline context
func (p *PipelineState) GetContext(key string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.Context[key]
	return val, ok
}

// SetContext sets a value in the pipeline context
func (p *PipelineState) SetContext(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Context[key] = value
}

// GetConfig retrieves a configuration value
func (p *PipelineState) GetConfig(key string) (interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	val, ok := p.Config[key]
	return val, ok
}

// SetConfig sets a configuration value
func (p *PipelineState) SetConfig(key string, value interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Config[key] = value
}

// Duration returns the duration of the pipeline execution
func (p *PipelineState) Duration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.EndTime != nil {
		return p.EndTime.Sub(p.StartTime)
	}
	return time.Since(p.StartTime)
}

// GetActiveStages returns all currently active stages
func (p *PipelineState) GetActiveStages() []*StageState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var active []*StageState
	for _, stage := range p.Stages {
		if stage.Status == StageStatusActive {
			active = append(active, stage)
		}
	}
	return active
}

// GetCompletedStages returns all completed stages
func (p *PipelineState) GetCompletedStages() []*StageState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var completed []*StageState
	for _, stage := range p.Stages {
		if stage.Status == StageStatusCompleted {
			completed = append(completed, stage)
		}
	}
	return completed
}

// GetFailedStages returns all failed stages
func (p *PipelineState) GetFailedStages() []*StageState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var failed []*StageState
	for _, stage := range p.Stages {
		if stage.Status == StageStatusFailed {
			failed = append(failed, stage)
		}
	}
	return failed
}

// IsComplete returns true if all stages are completed or skipped
func (p *PipelineState) IsComplete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, stage := range p.Stages {
		if stage.Status == StageStatusPending || stage.Status == StageStatusActive {
			return false
		}
	}
	return true
}

// HasFailures returns true if any stage has failed
func (p *PipelineState) HasFailures() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, stage := range p.Stages {
		if stage.Status == StageStatusFailed {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the pipeline state
func (p *PipelineState) Clone() *PipelineState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	clone := &PipelineState{
		ID:        p.ID,
		Status:    p.Status,
		StartTime: p.StartTime,
		Stages:    make(map[string]*StageState),
		Context:   make(map[string]interface{}),
		Config:    make(map[string]interface{}),
		Error:     p.Error,
	}

	if p.EndTime != nil {
		endTime := *p.EndTime
		clone.EndTime = &endTime
	}

	// Clone stages
	for k, v := range p.Stages {
		stageCopy := *v
		clone.Stages[k] = &stageCopy
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