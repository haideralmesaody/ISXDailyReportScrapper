package testutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"isxcli/internal/pipeline"
)

// CreateTestPipelineState creates a pipeline state for testing
func CreateTestPipelineState(id string) *pipeline.PipelineState {
	state := pipeline.NewPipelineState(id)
	state.SetConfig(pipeline.ContextKeyFromDate, "2024-01-01")
	state.SetConfig(pipeline.ContextKeyToDate, "2024-01-31")
	state.SetConfig(pipeline.ContextKeyMode, pipeline.ModeInitial)
	return state
}

// CreateTestStageState creates a stage state for testing
func CreateTestStageState(id, name string) *pipeline.StageState {
	return pipeline.NewStageState(id, name)
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() *pipeline.Config {
	return pipeline.NewConfigBuilder().
		WithExecutionMode(pipeline.ExecutionModeSequential).
		WithRetryConfig(pipeline.RetryConfig{
			MaxAttempts:  2,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		}).
		WithStageTimeout(pipeline.StageIDScraping, 1*time.Second).
		WithStageTimeout(pipeline.StageIDProcessing, 1*time.Second).
		WithStageTimeout(pipeline.StageIDIndices, 1*time.Second).
		WithStageTimeout(pipeline.StageIDAnalysis, 1*time.Second).
		Build()
}

// CreateTestRegistry creates a registry with test stages
func CreateTestRegistry() *pipeline.Registry {
	registry := pipeline.NewRegistry()
	
	// Register test stages
	registry.Register(CreateSuccessfulStage("stage1", "Stage 1"))
	registry.Register(CreateSuccessfulStage("stage2", "Stage 2"))
	registry.Register(CreateSuccessfulStage("stage3", "Stage 3"))
	
	return registry
}

// CreateSuccessfulStage creates a stage that always succeeds
func CreateSuccessfulStage(id, name string, deps ...string) *MockStage {
	return &MockStage{
		IDValue:           id,
		NameValue:         name,
		DependenciesValue: deps,
		ExecuteFunc: func(ctx context.Context, state *pipeline.PipelineState) error {
			// Simulate some work
			stageState := state.GetStage(id)
			if stageState != nil {
				stageState.UpdateProgress(50, "Processing...")
				time.Sleep(10 * time.Millisecond)
				stageState.UpdateProgress(100, "Completed")
			}
			return nil
		},
	}
}

// CreateFailingStage creates a stage that always fails
func CreateFailingStage(id, name string, err error, deps ...string) *MockStage {
	if err == nil {
		err = errors.New("stage failed")
	}
	
	return &MockStage{
		IDValue:           id,
		NameValue:         name,
		DependenciesValue: deps,
		ExecuteFunc: func(ctx context.Context, state *pipeline.PipelineState) error {
			return err
		},
	}
}

// CreateRetryableStage creates a stage that fails then succeeds
func CreateRetryableStage(id, name string, failCount int, deps ...string) *MockStage {
	attempts := 0
	
	return &MockStage{
		IDValue:           id,
		NameValue:         name,
		DependenciesValue: deps,
		ExecuteFunc: func(ctx context.Context, state *pipeline.PipelineState) error {
			attempts++
			if attempts <= failCount {
				return pipeline.NewExecutionError(id, errors.New("temporary failure"), true)
			}
			return nil
		},
	}
}

// CreateSlowStage creates a stage that takes a specific duration
func CreateSlowStage(id, name string, duration time.Duration, deps ...string) *MockStage {
	return &MockStage{
		IDValue:           id,
		NameValue:         name,
		DependenciesValue: deps,
		ExecuteFunc: func(ctx context.Context, state *pipeline.PipelineState) error {
			select {
			case <-time.After(duration):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
}

// CreateValidationFailingStage creates a stage that fails validation
func CreateValidationFailingStage(id, name string, validationErr error, deps ...string) *MockStage {
	if validationErr == nil {
		validationErr = errors.New("validation failed")
	}
	
	return &MockStage{
		IDValue:           id,
		NameValue:         name,
		DependenciesValue: deps,
		ValidateFunc: func(state *pipeline.PipelineState) error {
			return validationErr
		},
	}
}

// CreateContextAwareStage creates a stage that reads/writes context
func CreateContextAwareStage(id, name string, readKey, writeKey string, writeValue interface{}, deps ...string) *MockStage {
	return &MockStage{
		IDValue:           id,
		NameValue:         name,
		DependenciesValue: deps,
		ExecuteFunc: func(ctx context.Context, state *pipeline.PipelineState) error {
			// Read from context if readKey is provided
			if readKey != "" {
				if val, ok := state.GetContext(readKey); ok {
					// Log or use the value
					_ = val
				}
			}
			
			// Write to context if writeKey is provided
			if writeKey != "" {
				state.SetContext(writeKey, writeValue)
			}
			
			return nil
		},
	}
}

// CreateComplexPipelineStages creates stages with complex dependencies
func CreateComplexPipelineStages() []pipeline.Stage {
	// Create a diamond dependency pattern:
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	
	stageA := CreateSuccessfulStage("A", "Stage A")
	stageB := CreateSuccessfulStage("B", "Stage B", "A")
	stageC := CreateSuccessfulStage("C", "Stage C", "A")
	stageD := CreateSuccessfulStage("D", "Stage D", "B", "C")
	
	return []pipeline.Stage{stageA, stageB, stageC, stageD}
}

// CreatePipelineRequest creates a test pipeline request
func CreatePipelineRequest(mode string) pipeline.PipelineRequest {
	return pipeline.PipelineRequest{
		ID:       fmt.Sprintf("test-pipeline-%d", time.Now().UnixNano()),
		Mode:     mode,
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
		Parameters: map[string]interface{}{
			"test": true,
		},
	}
}

// StageBuilder provides a fluent interface for creating test stages
type StageBuilder struct {
	stage *MockStage
}

// NewStageBuilder creates a new stage builder
func NewStageBuilder(id, name string) *StageBuilder {
	return &StageBuilder{
		stage: &MockStage{
			IDValue:   id,
			NameValue: name,
		},
	}
}

// WithDependencies sets the stage dependencies
func (b *StageBuilder) WithDependencies(deps ...string) *StageBuilder {
	b.stage.DependenciesValue = deps
	return b
}

// WithExecute sets the execute function
func (b *StageBuilder) WithExecute(fn func(context.Context, *pipeline.PipelineState) error) *StageBuilder {
	b.stage.ExecuteFunc = fn
	return b
}

// WithValidate sets the validate function
func (b *StageBuilder) WithValidate(fn func(*pipeline.PipelineState) error) *StageBuilder {
	b.stage.ValidateFunc = fn
	return b
}

// Build returns the constructed stage
func (b *StageBuilder) Build() *MockStage {
	return b.stage
}