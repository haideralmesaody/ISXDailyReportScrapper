package pipeline_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"isxcli/internal/pipeline"
	"isxcli/internal/pipeline/testutil"
)

// TestFullPipelineExecution tests a complete pipeline flow
func TestFullPipelineExecution(t *testing.T) {
	// Create test infrastructure
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Configure pipeline
	config := pipeline.NewConfigBuilder().
		WithExecutionMode(pipeline.ExecutionModeSequential).
		WithStageTimeout(pipeline.StageIDScraping, 1*time.Second).
		WithStageTimeout(pipeline.StageIDProcessing, 1*time.Second).
		WithRetryConfig(pipeline.RetryConfig{
			MaxAttempts:  2,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
		}).
		Build()
	manager.SetConfig(config)
	
	// Create and register stages that simulate the real pipeline
	scrapingStage := testutil.NewStageBuilder(pipeline.StageIDScraping, pipeline.StageNameScraping).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			// Simulate scraping
			stageState := state.GetStage(pipeline.StageIDScraping)
			stageState.UpdateProgress(25, "Connecting to ISX website...")
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(50, "Downloading reports...")
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(75, "Saving files...")
			time.Sleep(20 * time.Millisecond)
			
			// Set context for next stage
			state.SetContext(pipeline.ContextKeyFilesFound, 10)
			state.SetContext(pipeline.ContextKeyScraperSuccess, true)
			
			stageState.UpdateProgress(100, "Scraping completed")
			return nil
		}).
		Build()
	
	processingStage := testutil.NewStageBuilder(pipeline.StageIDProcessing, pipeline.StageNameProcessing).
		WithDependencies(pipeline.StageIDScraping).
		WithValidate(func(state *pipeline.PipelineState) error {
			// Check if scraper succeeded
			if success, ok := state.GetContext(pipeline.ContextKeyScraperSuccess); !ok || !success.(bool) {
				return errors.New("scraper did not succeed")
			}
			return nil
		}).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			// Simulate processing
			filesFound, _ := state.GetContext(pipeline.ContextKeyFilesFound)
			totalFiles := filesFound.(int)
			
			stageState := state.GetStage(pipeline.StageIDProcessing)
			for i := 0; i < totalFiles; i++ {
				progress := float64(i+1) / float64(totalFiles) * 100
				stageState.UpdateProgress(progress, fmt.Sprintf("Processing file %d of %d", i+1, totalFiles))
				time.Sleep(10 * time.Millisecond)
			}
			
			state.SetContext(pipeline.ContextKeyFilesProcessed, totalFiles)
			return nil
		}).
		Build()
	
	indicesStage := testutil.NewStageBuilder(pipeline.StageIDIndices, pipeline.StageNameIndices).
		WithDependencies(pipeline.StageIDProcessing).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage(pipeline.StageIDIndices)
			stageState.UpdateProgress(50, "Extracting ISX60...")
			time.Sleep(20 * time.Millisecond)
			stageState.UpdateProgress(100, "Indices extracted")
			return nil
		}).
		Build()
	
	analysisStage := testutil.NewStageBuilder(pipeline.StageIDAnalysis, pipeline.StageNameAnalysis).
		WithDependencies(pipeline.StageIDIndices).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage(pipeline.StageIDAnalysis)
			stageState.UpdateProgress(100, "Analysis completed")
			return nil
		}).
		Build()
	
	// Register stages
	testutil.AssertNoError(t, manager.RegisterStage(scrapingStage))
	testutil.AssertNoError(t, manager.RegisterStage(processingStage))
	testutil.AssertNoError(t, manager.RegisterStage(indicesStage))
	testutil.AssertNoError(t, manager.RegisterStage(analysisStage))
	
	// Execute pipeline
	ctx := context.Background()
	req := testutil.CreatePipelineRequest(pipeline.ModeInitial)
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Verify response
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	testutil.AssertEqual(t, len(resp.Stages), 4)
	
	// Verify all stages completed
	for _, stage := range resp.Stages {
		testutil.AssertStageStatus(t, stage, pipeline.StageStatusCompleted)
	}
	
	// Verify WebSocket messages
	_ = hub.GetMessages()
	
	// Should have pipeline reset at start
	testutil.AssertWebSocketMessage(t, hub, pipeline.EventTypePipelineReset)
	
	// Should have multiple progress updates
	progressMessages := hub.GetMessagesByType(pipeline.EventTypePipelineProgress)
	if len(progressMessages) < 4 {
		t.Errorf("Expected at least 4 progress messages, got %d", len(progressMessages))
	}
	
	// Should have pipeline complete at end
	testutil.AssertWebSocketMessage(t, hub, pipeline.EventTypePipelineComplete)
	
	// Verify stage execution order
	testutil.AssertStageOrder(t, []*testutil.MockStage{
		scrapingStage, processingStage, indicesStage, analysisStage,
	}, []string{
		pipeline.StageIDScraping,
		pipeline.StageIDProcessing,
		pipeline.StageIDIndices,
		pipeline.StageIDAnalysis,
	})
}

// TestPipelineWithFailureAndRetry tests stage failure with retry
func TestPipelineWithFailureAndRetry(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Configure with retry
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create a stage that fails once then succeeds
	retryStage := testutil.CreateRetryableStage("retry-stage", "Retry Stage", 1)
	
	// Register stage
	testutil.AssertNoError(t, manager.RegisterStage(retryStage))
	
	// Execute pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-retry"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Verify success after retry
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify stage was called twice
	testutil.AssertEqual(t, retryStage.GetExecuteCalls(), 2)
	
	// Check warning logs for retry
	testutil.AssertLogContains(t, logger, "warning", "retrying")
}

// TestPipelineWithDependencyFailure tests dependency handling
func TestPipelineWithDependencyFailure(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	config.ContinueOnError = false // Stop on error
	manager.SetConfig(config)
	
	// Create stages where first fails
	stage1 := testutil.CreateFailingStage("stage1", "Stage 1", errors.New("stage1 failed"))
	stage2 := testutil.CreateSuccessfulStage("stage2", "Stage 2", "stage1")
	stage3 := testutil.CreateSuccessfulStage("stage3", "Stage 3", "stage2")
	
	// Register stages
	testutil.AssertNoError(t, manager.RegisterStage(stage1))
	testutil.AssertNoError(t, manager.RegisterStage(stage2))
	testutil.AssertNoError(t, manager.RegisterStage(stage3))
	
	// Execute pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-deps"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertError(t, err, true)
	
	// Verify pipeline failed
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusFailed)
	
	// Verify stage statuses
	testutil.AssertStageFailed(t, &pipeline.PipelineState{Stages: resp.Stages}, "stage1")
	
	// Stages 2 and 3 should be skipped due to dependency failure
	stage2State := resp.Stages["stage2"]
	stage3State := resp.Stages["stage3"]
	
	if stage2State.Status != pipeline.StageStatusSkipped {
		t.Errorf("Stage2 status = %v, want skipped", stage2State.Status)
	}
	if stage3State.Status != pipeline.StageStatusSkipped {
		t.Errorf("Stage3 status = %v, want skipped", stage3State.Status)
	}
}

// TestPipelineTimeout tests stage timeout handling
func TestPipelineTimeout(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Configure with short timeout
	config := pipeline.NewConfigBuilder().
		WithStageTimeout("slow-stage", 50*time.Millisecond).
		Build()
	manager.SetConfig(config)
	
	// Create a slow stage
	slowStage := testutil.CreateSlowStage("slow-stage", "Slow Stage", 200*time.Millisecond)
	
	// Register stage
	testutil.AssertNoError(t, manager.RegisterStage(slowStage))
	
	// Execute pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-timeout"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertError(t, err, true)
	
	// Verify pipeline failed
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusFailed)
	
	// Verify timeout error
	stageState := resp.Stages["slow-stage"]
	if stageState.Status != pipeline.StageStatusFailed {
		t.Errorf("Stage status = %v, want failed", stageState.Status)
	}
}

// TestPipelineCancellation tests context cancellation
func TestPipelineCancellation(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create a slow stage
	slowStage := testutil.CreateSlowStage("slow-stage", "Slow Stage", 200*time.Millisecond)
	
	// Register stage
	testutil.AssertNoError(t, manager.RegisterStage(slowStage))
	
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	
	// Execute pipeline
	req := pipeline.PipelineRequest{ID: "test-cancel"}
	_, err := manager.Execute(ctx, req)
	
	// Should have error
	testutil.AssertError(t, err, true)
	
	// Check for cancellation
	if ctx.Err() != context.Canceled {
		t.Error("Context should be cancelled")
	}
}

// TestComplexPipelineDependencies tests diamond dependency pattern
func TestComplexPipelineDependencies(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create diamond pattern stages
	stages := testutil.CreateComplexPipelineStages()
	
	// Register all stages
	for _, stage := range stages {
		testutil.AssertNoError(t, manager.RegisterStage(stage))
	}
	
	// Execute pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-diamond"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Verify all stages completed
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify execution order
	// A must run first
	// B and C can run in any order after A
	// D must run last
	mockStages := make([]*testutil.MockStage, len(stages))
	for i, s := range stages {
		mockStages[i] = s.(*testutil.MockStage)
	}
	
	// Check A was first
	aTime := mockStages[0].ExecuteArgs[0].Time
	for i := 1; i < len(mockStages); i++ {
		if mockStages[i].ExecuteArgs[0].Time.Before(aTime) {
			t.Error("Stage A should execute first")
		}
	}
	
	// Check D was last
	dTime := mockStages[3].ExecuteArgs[0].Time
	for i := 0; i < 3; i++ {
		if mockStages[i].ExecuteArgs[0].Time.After(dTime) {
			t.Error("Stage D should execute last")
		}
	}
}

// TestPipelineStateSharing tests context sharing between stages
func TestPipelineStateSharing(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create stages that share data
	stage1 := testutil.CreateContextAwareStage("stage1", "Stage 1", "", "shared_data", "Hello from stage1")
	stage2 := testutil.CreateContextAwareStage("stage2", "Stage 2", "shared_data", "stage2_data", "Modified", "stage1")
	
	// Stage 2 should read from stage 1 and write its own data
	stage2.ExecuteFunc = func(ctx context.Context, state *pipeline.PipelineState) error {
		// Read shared data
		data, ok := state.GetContext("shared_data")
		if !ok {
			return errors.New("shared_data not found")
		}
		
		// Modify and write back
		modified := data.(string) + " - Modified by stage2"
		state.SetContext("stage2_data", modified)
		return nil
	}
	
	// Register stages
	testutil.AssertNoError(t, manager.RegisterStage(stage1))
	testutil.AssertNoError(t, manager.RegisterStage(stage2))
	
	// Execute pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-sharing"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Verify pipeline completed
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Note: We can't verify context values from response as they're not included
	// In a real implementation, we might want to add a way to retrieve final state
}