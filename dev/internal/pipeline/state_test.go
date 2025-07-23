package pipeline_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"isxcli/internal/pipeline"
	"isxcli/internal/pipeline/testutil"
)

func TestNewPipelineState(t *testing.T) {
	id := "test-pipeline"
	state := pipeline.NewPipelineState(id)
	
	// Verify initial values
	testutil.AssertEqual(t, state.ID, id)
	testutil.AssertPipelineStatus(t, state, pipeline.PipelineStatusPending)
	testutil.AssertNotNil(t, state.Stages)
	testutil.AssertNotNil(t, state.Context)
	testutil.AssertNotNil(t, state.Config)
	
	if state.EndTime != nil {
		t.Error("EndTime should be nil initially")
	}
	if state.Error != nil {
		t.Error("Error should be nil initially")
	}
	
	// Verify start time is set
	if state.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}

func TestPipelineStateTransitions(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*pipeline.PipelineState)
		wantStatus pipeline.PipelineStatus
		checkEnd   bool
		checkError bool
	}{
		{
			name: "Start",
			transition: func(p *pipeline.PipelineState) {
				p.Start()
			},
			wantStatus: pipeline.PipelineStatusRunning,
			checkEnd:   false,
		},
		{
			name: "Complete",
			transition: func(p *pipeline.PipelineState) {
				p.Complete()
			},
			wantStatus: pipeline.PipelineStatusCompleted,
			checkEnd:   true,
		},
		{
			name: "Fail",
			transition: func(p *pipeline.PipelineState) {
				p.Fail(errors.New("test error"))
			},
			wantStatus: pipeline.PipelineStatusFailed,
			checkEnd:   true,
			checkError: true,
		},
		{
			name: "Cancel",
			transition: func(p *pipeline.PipelineState) {
				p.Cancel()
			},
			wantStatus: pipeline.PipelineStatusCancelled,
			checkEnd:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := pipeline.NewPipelineState("test")
			
			tt.transition(state)
			
			testutil.AssertPipelineStatus(t, state, tt.wantStatus)
			
			if tt.checkEnd && state.EndTime == nil {
				t.Error("EndTime should be set")
			}
			if !tt.checkEnd && state.EndTime != nil {
				t.Error("EndTime should not be set")
			}
			if tt.checkError && state.Error == nil {
				t.Error("Error should be set")
			}
		})
	}
}

func TestPipelineStateStageManagement(t *testing.T) {
	state := pipeline.NewPipelineState("test")
	
	// Add stages
	stage1 := pipeline.NewStageState("stage1", "Stage 1")
	stage2 := pipeline.NewStageState("stage2", "Stage 2")
	stage3 := pipeline.NewStageState("stage3", "Stage 3")
	
	state.SetStage("stage1", stage1)
	state.SetStage("stage2", stage2)
	state.SetStage("stage3", stage3)
	
	// Retrieve stages
	got1 := state.GetStage("stage1")
	got2 := state.GetStage("stage2")
	got3 := state.GetStage("stage3")
	gotNil := state.GetStage("nonexistent")
	
	if got1 != stage1 {
		t.Error("Stage 1 not retrieved correctly")
	}
	if got2 != stage2 {
		t.Error("Stage 2 not retrieved correctly")
	}
	if got3 != stage3 {
		t.Error("Stage 3 not retrieved correctly")
	}
	if gotNil != nil {
		t.Error("Nonexistent stage should return nil")
	}
}

func TestPipelineStateContext(t *testing.T) {
	state := pipeline.NewPipelineState("test")
	
	// Test setting and getting context values
	state.SetContext("key1", "value1")
	state.SetContext("key2", 42)
	state.SetContext("key3", true)
	
	// Get values
	val1, ok1 := state.GetContext("key1")
	val2, ok2 := state.GetContext("key2")
	val3, ok3 := state.GetContext("key3")
	_, ok4 := state.GetContext("nonexistent")
	
	if !ok1 || val1 != "value1" {
		t.Error("Context key1 not retrieved correctly")
	}
	if !ok2 || val2 != 42 {
		t.Error("Context key2 not retrieved correctly")
	}
	if !ok3 || val3 != true {
		t.Error("Context key3 not retrieved correctly")
	}
	if ok4 {
		t.Error("Nonexistent key should return false")
	}
}

func TestPipelineStateConfig(t *testing.T) {
	state := pipeline.NewPipelineState("test")
	
	// Test setting and getting config values
	state.SetConfig("mode", "initial")
	state.SetConfig("timeout", 30)
	state.SetConfig("retry", true)
	
	// Get values
	val1, ok1 := state.GetConfig("mode")
	val2, ok2 := state.GetConfig("timeout")
	val3, ok3 := state.GetConfig("retry")
	_, ok4 := state.GetConfig("nonexistent")
	
	if !ok1 || val1 != "initial" {
		t.Error("Config mode not retrieved correctly")
	}
	if !ok2 || val2 != 30 {
		t.Error("Config timeout not retrieved correctly")
	}
	if !ok3 || val3 != true {
		t.Error("Config retry not retrieved correctly")
	}
	if ok4 {
		t.Error("Nonexistent key should return false")
	}
}

func TestPipelineStateDuration(t *testing.T) {
	state := pipeline.NewPipelineState("test")
	
	// Start the pipeline
	state.Start()
	time.Sleep(50 * time.Millisecond)
	
	// Check duration while running
	duration := state.Duration()
	if duration <= 0 {
		t.Error("Duration should be > 0 while running")
	}
	
	// Complete the pipeline
	state.Complete()
	finalDuration := state.Duration()
	
	// Duration should be fixed after completion
	time.Sleep(10 * time.Millisecond)
	if state.Duration() != finalDuration {
		t.Error("Duration should not change after completion")
	}
	
	// Verify duration is reasonable
	testutil.AssertDuration(t, finalDuration, 50*time.Millisecond, 20*time.Millisecond)
}

func TestPipelineStateStageQueries(t *testing.T) {
	state := pipeline.NewPipelineState("test")
	
	// Add stages with different statuses
	active1 := pipeline.NewStageState("active1", "Active 1")
	active1.Status = pipeline.StageStatusActive
	
	active2 := pipeline.NewStageState("active2", "Active 2")
	active2.Status = pipeline.StageStatusActive
	
	completed := pipeline.NewStageState("completed", "Completed")
	completed.Status = pipeline.StageStatusCompleted
	
	failed := pipeline.NewStageState("failed", "Failed")
	failed.Status = pipeline.StageStatusFailed
	
	pending := pipeline.NewStageState("pending", "Pending")
	pending.Status = pipeline.StageStatusPending
	
	state.SetStage("active1", active1)
	state.SetStage("active2", active2)
	state.SetStage("completed", completed)
	state.SetStage("failed", failed)
	state.SetStage("pending", pending)
	
	// Test GetActiveStages
	activeStages := state.GetActiveStages()
	if len(activeStages) != 2 {
		t.Errorf("Active stages count = %d, want 2", len(activeStages))
	}
	
	// Test GetCompletedStages
	completedStages := state.GetCompletedStages()
	if len(completedStages) != 1 {
		t.Errorf("Completed stages count = %d, want 1", len(completedStages))
	}
	
	// Test GetFailedStages
	failedStages := state.GetFailedStages()
	if len(failedStages) != 1 {
		t.Errorf("Failed stages count = %d, want 1", len(failedStages))
	}
}

func TestPipelineStateIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		stages   map[string]pipeline.StageStatus
		want     bool
	}{
		{
			name: "All completed",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusCompleted,
				"s3": pipeline.StageStatusCompleted,
			},
			want: true,
		},
		{
			name: "Some skipped",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusSkipped,
				"s3": pipeline.StageStatusCompleted,
			},
			want: true,
		},
		{
			name: "Has pending",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusPending,
				"s3": pipeline.StageStatusCompleted,
			},
			want: false,
		},
		{
			name: "Has active",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusActive,
				"s3": pipeline.StageStatusCompleted,
			},
			want: false,
		},
		{
			name: "Has failed",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusFailed,
				"s3": pipeline.StageStatusCompleted,
			},
			want: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := pipeline.NewPipelineState("test")
			
			for id, status := range tt.stages {
				stage := pipeline.NewStageState(id, id)
				stage.Status = status
				state.SetStage(id, stage)
			}
			
			got := state.IsComplete()
			if got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineStateHasFailures(t *testing.T) {
	tests := []struct {
		name   string
		stages map[string]pipeline.StageStatus
		want   bool
	}{
		{
			name: "No failures",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusCompleted,
			},
			want: false,
		},
		{
			name: "Has failure",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusCompleted,
				"s2": pipeline.StageStatusFailed,
			},
			want: true,
		},
		{
			name: "Multiple failures",
			stages: map[string]pipeline.StageStatus{
				"s1": pipeline.StageStatusFailed,
				"s2": pipeline.StageStatusFailed,
			},
			want: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := pipeline.NewPipelineState("test")
			
			for id, status := range tt.stages {
				stage := pipeline.NewStageState(id, id)
				stage.Status = status
				state.SetStage(id, stage)
			}
			
			got := state.HasFailures()
			if got != tt.want {
				t.Errorf("HasFailures() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPipelineStateClone(t *testing.T) {
	// Create original state
	original := pipeline.NewPipelineState("test")
	original.Status = pipeline.PipelineStatusRunning
	original.SetContext("key1", "value1")
	original.SetConfig("config1", "configValue1")
	
	// Add stages
	stage1 := pipeline.NewStageState("stage1", "Stage 1")
	stage1.Status = pipeline.StageStatusCompleted
	original.SetStage("stage1", stage1)
	
	// Clone
	clone := original.Clone()
	
	// Verify clone has same values
	testutil.AssertEqual(t, clone.ID, original.ID)
	testutil.AssertPipelineStatus(t, clone, original.Status)
	
	// Verify context was cloned
	val, ok := clone.GetContext("key1")
	if !ok || val != "value1" {
		t.Error("Context not cloned correctly")
	}
	
	// Verify config was cloned
	val, ok = clone.GetConfig("config1")
	if !ok || val != "configValue1" {
		t.Error("Config not cloned correctly")
	}
	
	// Verify stages were cloned
	clonedStage := clone.GetStage("stage1")
	if clonedStage == nil || clonedStage.Status != pipeline.StageStatusCompleted {
		t.Error("Stages not cloned correctly")
	}
	
	// Verify modifications to clone don't affect original
	clone.SetContext("key2", "value2")
	_, ok = original.GetContext("key2")
	if ok {
		t.Error("Clone modifications affected original")
	}
}

func TestPipelineStateConcurrency(t *testing.T) {
	state := pipeline.NewPipelineState("test")
	
	// Run concurrent operations
	var wg sync.WaitGroup
	operations := 100
	
	// Concurrent context writes
	wg.Add(operations)
	for i := 0; i < operations; i++ {
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			state.SetContext(key, n)
		}(i)
	}
	
	// Concurrent config writes
	wg.Add(operations)
	for i := 0; i < operations; i++ {
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("config%d", n)
			state.SetConfig(key, n)
		}(i)
	}
	
	// Concurrent stage writes
	wg.Add(operations)
	for i := 0; i < operations; i++ {
		go func(n int) {
			defer wg.Done()
			id := fmt.Sprintf("stage%d", n)
			stage := pipeline.NewStageState(id, id)
			state.SetStage(id, stage)
		}(i)
	}
	
	// Concurrent reads
	wg.Add(operations)
	for i := 0; i < operations; i++ {
		go func(n int) {
			defer wg.Done()
			state.GetActiveStages()
			state.GetCompletedStages()
			state.GetFailedStages()
			state.IsComplete()
			state.HasFailures()
			state.Duration()
		}(i)
	}
	
	wg.Wait()
	
	// Verify all writes succeeded
	for i := 0; i < operations; i++ {
		key := fmt.Sprintf("key%d", i)
		val, ok := state.GetContext(key)
		if !ok || val != i {
			t.Errorf("Context %s not set correctly", key)
		}
		
		key = fmt.Sprintf("config%d", i)
		val, ok = state.GetConfig(key)
		if !ok || val != i {
			t.Errorf("Config %s not set correctly", key)
		}
		
		id := fmt.Sprintf("stage%d", i)
		stage := state.GetStage(id)
		if stage == nil {
			t.Errorf("Stage %s not set correctly", id)
		}
	}
}