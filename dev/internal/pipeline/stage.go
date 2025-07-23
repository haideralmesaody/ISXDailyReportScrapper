package pipeline

import (
	"context"
	"time"
)

// Stage represents a single stage in the pipeline
type Stage interface {
	// ID returns the unique identifier for this stage
	ID() string

	// Name returns the human-readable name for this stage
	Name() string

	// Execute runs the stage with the given context and pipeline state
	Execute(ctx context.Context, state *PipelineState) error

	// Validate checks if the stage can be executed with the current state
	Validate(state *PipelineState) error

	// GetDependencies returns the IDs of stages that must complete before this stage
	GetDependencies() []string
}

// StageStatus represents the current status of a stage
type StageStatus string

const (
	StageStatusPending   StageStatus = "pending"
	StageStatusActive    StageStatus = "active"
	StageStatusCompleted StageStatus = "completed"
	StageStatusFailed    StageStatus = "failed"
	StageStatusSkipped   StageStatus = "skipped"
)

// StageState represents the runtime state of a stage
type StageState struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      StageStatus            `json:"status"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Progress    float64                `json:"progress"`
	Message     string                 `json:"message"`
	Error       error                  `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewStageState creates a new stage state with default values
func NewStageState(id, name string) *StageState {
	return &StageState{
		ID:       id,
		Name:     name,
		Status:   StageStatusPending,
		Progress: 0,
		Metadata: make(map[string]interface{}),
	}
}

// Start marks the stage as active and sets the start time
func (s *StageState) Start() {
	now := time.Now()
	s.StartTime = &now
	s.Status = StageStatusActive
	s.Progress = 0
}

// Complete marks the stage as completed and sets the end time
func (s *StageState) Complete() {
	now := time.Now()
	s.EndTime = &now
	s.Status = StageStatusCompleted
	s.Progress = 100
}

// Fail marks the stage as failed with the given error
func (s *StageState) Fail(err error) {
	now := time.Now()
	s.EndTime = &now
	s.Status = StageStatusFailed
	s.Error = err
}

// Skip marks the stage as skipped with the given reason
func (s *StageState) Skip(reason string) {
	now := time.Now()
	s.EndTime = &now
	s.Status = StageStatusSkipped
	s.Message = reason
}

// UpdateProgress updates the stage progress and message
func (s *StageState) UpdateProgress(progress float64, message string) {
	s.Progress = progress
	s.Message = message
}

// Duration returns the duration of the stage execution
func (s *StageState) Duration() time.Duration {
	if s.StartTime == nil {
		return 0
	}
	if s.EndTime != nil {
		return s.EndTime.Sub(*s.StartTime)
	}
	return time.Since(*s.StartTime)
}

// BaseStage provides common functionality for stage implementations
type BaseStage struct {
	id           string
	name         string
	dependencies []string
}

// NewBaseStage creates a new base stage
func NewBaseStage(id, name string, dependencies []string) BaseStage {
	if dependencies == nil {
		dependencies = []string{}
	}
	return BaseStage{
		id:           id,
		name:         name,
		dependencies: dependencies,
	}
}

// ID returns the stage ID
func (b BaseStage) ID() string {
	return b.id
}

// Name returns the stage name
func (b BaseStage) Name() string {
	return b.name
}

// GetDependencies returns the stage dependencies
func (b BaseStage) GetDependencies() []string {
	return b.dependencies
}

// Validate provides a default validation that always passes
func (b BaseStage) Validate(state *PipelineState) error {
	return nil
}