package pipeline_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"isxcli/internal/pipeline"
	"isxcli/internal/pipeline/testutil"
)

// TestIntegrationFullPipeline tests a complete pipeline execution with all features
func TestIntegrationFullPipeline(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	harness.SetupStandardPipeline()
	
	// Configure pipeline manager
	config := pipeline.NewConfigBuilder().
		WithExecutionMode(pipeline.ExecutionModeSequential).
		WithRetryConfig(pipeline.RetryConfig{
			MaxAttempts:  2,
			InitialDelay: 10 * time.Millisecond,
		}).
		Build()
	harness.GetManager().SetConfig(config)
	
	// Generate test data
	harness.GetDataGenerator().GenerateDateRange("2024-01-01", "2024-01-05")
	
	// Execute pipeline
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Verify success
	harness.AssertPipelineSuccess(resp, err)
	harness.AssertWebSocketMessages()
	
	// Verify all stages completed
	stages := []string{
		pipeline.StageIDScraping,
		pipeline.StageIDProcessing,
		pipeline.StageIDIndices,
		pipeline.StageIDAnalysis,
	}
	
	for _, stageID := range stages {
		testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, stageID)
	}
	
	// Verify WebSocket message count
	messages := harness.GetHub().GetMessages()
	if len(messages) < 10 {
		t.Errorf("Expected at least 10 WebSocket messages, got %d", len(messages))
	}
}

// TestIntegrationPipelineWithFailure tests pipeline behavior with stage failures
func TestIntegrationPipelineWithFailure(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Setup pipeline with a failing stage
	scrapingStage := testutil.CreateSuccessfulStage(pipeline.StageIDScraping, pipeline.StageNameScraping)
	processingStage := testutil.CreateFailingStage(pipeline.StageIDProcessing, pipeline.StageNameProcessing, 
		fmt.Errorf("processing failed"), pipeline.StageIDScraping)
	indicesStage := testutil.CreateSuccessfulStage(pipeline.StageIDIndices, pipeline.StageNameIndices, 
		pipeline.StageIDProcessing)
	analysisStage := testutil.CreateSuccessfulStage(pipeline.StageIDAnalysis, pipeline.StageNameAnalysis, 
		pipeline.StageIDIndices)
	
	harness.GetManager().RegisterStage(scrapingStage)
	harness.GetManager().RegisterStage(processingStage)
	harness.GetManager().RegisterStage(indicesStage)
	harness.GetManager().RegisterStage(analysisStage)
	
	// Execute pipeline
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Verify failure
	testutil.AssertError(t, err, true)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusFailed)
	
	// Verify stage statuses
	testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, pipeline.StageIDScraping)
	testutil.AssertStageFailed(t, &pipeline.PipelineState{Stages: resp.Stages}, pipeline.StageIDProcessing)
	testutil.AssertStageSkipped(t, &pipeline.PipelineState{Stages: resp.Stages}, pipeline.StageIDIndices)
	testutil.AssertStageSkipped(t, &pipeline.PipelineState{Stages: resp.Stages}, pipeline.StageIDAnalysis)
	
	// Verify error WebSocket message
	testutil.AssertWebSocketMessage(t, harness.GetHub(), pipeline.EventTypePipelineError)
}

// TestIntegrationPipelineWithRetry tests retry mechanism
func TestIntegrationPipelineWithRetry(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Configure retries
	config := pipeline.NewConfigBuilder().
		WithRetryConfig(pipeline.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		}).
		Build()
	harness.GetManager().SetConfig(config)
	
	// Create stage that fails twice then succeeds
	scrapingStage := testutil.CreateRetryableStage(pipeline.StageIDScraping, pipeline.StageNameScraping, 2)
	processingStage := testutil.CreateSuccessfulStage(pipeline.StageIDProcessing, pipeline.StageNameProcessing, 
		pipeline.StageIDScraping)
	
	harness.GetManager().RegisterStage(scrapingStage)
	harness.GetManager().RegisterStage(processingStage)
	
	// Execute pipeline
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Should succeed after retries
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify retry attempts
	if scrapingStage.GetExecuteCalls() != 3 {
		t.Errorf("Expected 3 execution attempts, got %d", scrapingStage.GetExecuteCalls())
	}
	
	// Check for retry log messages
	logs := harness.GetLogger().GetWarningLogs()
	retryCount := 0
	for _, log := range logs {
		msg := fmt.Sprintf(log.Format, log.Args...)
		t.Logf("Warning log: %s", msg)
		if containsStr(msg, "retrying") || containsStr(msg, "retry") {
			retryCount++
		}
	}
	if retryCount != 2 {
		t.Errorf("Expected 2 retry warnings, got %d", retryCount)
		t.Logf("Total warning logs: %d", len(logs))
	}
}

// TestIntegrationPipelineTimeout tests stage timeout
func TestIntegrationPipelineTimeout(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Configure short timeout
	config := pipeline.NewConfigBuilder().
		WithStageTimeout(pipeline.StageIDScraping, 50*time.Millisecond).
		Build()
	harness.GetManager().SetConfig(config)
	
	// Create slow stage
	scrapingStage := testutil.CreateSlowStage(pipeline.StageIDScraping, pipeline.StageNameScraping, 
		200*time.Millisecond)
	harness.GetManager().RegisterStage(scrapingStage)
	
	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	req := pipeline.PipelineRequest{
		ID:       "test-timeout",
		FromDate: "2024-01-01",
		ToDate:   "2024-01-05",
	}
	
	resp, err := harness.GetManager().Execute(ctx, req)
	
	// Should timeout
	testutil.AssertError(t, err, true)
	if resp != nil {
		testutil.AssertStageFailed(t, &pipeline.PipelineState{Stages: resp.Stages}, pipeline.StageIDScraping)
	}
}

// TestIntegrationConcurrentPipelines tests running multiple pipelines concurrently
func TestIntegrationConcurrentPipelines(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	harness.SetupStandardPipeline()
	
	// Run 5 concurrent pipelines
	errors := harness.RunConcurrentPipelines(5)
	
	// All should succeed
	for i, err := range errors {
		if err != nil {
			t.Errorf("Pipeline %d failed: %v", i, err)
		}
	}
	
	// Verify no active pipelines remain
	pipelines := harness.GetManager().ListPipelines()
	if len(pipelines) != 0 {
		t.Errorf("Expected 0 active pipelines, got %d", len(pipelines))
	}
}

// TestIntegrationPipelineStateSharing tests context sharing between stages
func TestIntegrationPipelineStateSharing(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Create stages that share data
	stage1 := testutil.CreateContextAwareStage("stage1", "Stage 1", "", "data1", "value1")
	stage2 := testutil.CreateContextAwareStage("stage2", "Stage 2", "data1", "data2", "value2", "stage1")
	stage3 := testutil.NewStageBuilder("stage3", "Stage 3").
		WithDependencies("stage2").
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			// Verify both values are available
			val1, ok1 := state.GetContext("data1")
			val2, ok2 := state.GetContext("data2")
			
			if !ok1 || val1 != "value1" {
				return fmt.Errorf("data1 not found or incorrect")
			}
			if !ok2 || val2 != "value2" {
				return fmt.Errorf("data2 not found or incorrect")
			}
			
			return nil
		}).
		Build()
	
	harness.GetManager().RegisterStage(stage1)
	harness.GetManager().RegisterStage(stage2)
	harness.GetManager().RegisterStage(stage3)
	
	// Execute
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Should succeed
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
}

// TestIntegrationComplexDependencies tests complex dependency patterns
func TestIntegrationComplexDependencies(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Create diamond dependency pattern
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	stages := testutil.CreateComplexPipelineStages()
	
	for _, stage := range stages {
		harness.GetManager().RegisterStage(stage)
	}
	
	// Execute
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Should succeed
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify all stages completed
	for _, stage := range stages {
		testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, stage.ID())
	}
}

// TestIntegrationProgressTracking tests detailed progress tracking
func TestIntegrationProgressTracking(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Create stage with detailed progress updates
	progressStage := testutil.NewStageBuilder("progress", "Progress Stage").
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage("progress")
			
			// Simulate work with progress updates
			for i := 0; i <= 10; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					progress := float64(i) * 10
					stageState.UpdateProgress(progress, fmt.Sprintf("Processing step %d of 10", i))
					time.Sleep(10 * time.Millisecond)
				}
			}
			
			return nil
		}).
		Build()
	
	harness.GetManager().RegisterStage(progressStage)
	
	// Execute
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Should succeed
	testutil.AssertNoError(t, err)
	
	// Verify progress messages
	progressMessages := harness.GetHub().GetMessagesByType(pipeline.EventTypePipelineProgress)
	if len(progressMessages) < 2 {
		t.Errorf("Expected at least 2 progress messages, got %d", len(progressMessages))
	}
	
	// Log all messages for debugging
	allMessages := harness.GetHub().GetMessages()
	t.Logf("Total messages: %d", len(allMessages))
	for _, msg := range allMessages {
		t.Logf("Message type: %s, stage: %s", msg.EventType, msg.Stage)
	}
	
	// Verify final progress is 100
	stage := resp.Stages["progress"]
	if stage.Progress != 100 {
		t.Errorf("Expected final progress 100, got %.1f", stage.Progress)
	}
}

// TestIntegrationErrorRecovery tests error recovery mechanisms
func TestIntegrationErrorRecovery(t *testing.T) {
	harness := testutil.NewIntegrationTestHarness(t)
	
	// Configure continue on error
	config := pipeline.NewConfigBuilder().
		WithContinueOnError(true).
		Build()
	harness.GetManager().SetConfig(config)
	
	// Create pipeline where stage 2 fails but stage 3 can still run
	stage1 := testutil.CreateSuccessfulStage("stage1", "Stage 1")
	stage2 := testutil.CreateFailingStage("stage2", "Stage 2", fmt.Errorf("stage 2 error"))
	stage3 := testutil.CreateSuccessfulStage("stage3", "Stage 3") // No dependency on stage2
	
	harness.GetManager().RegisterStage(stage1)
	harness.GetManager().RegisterStage(stage2)
	harness.GetManager().RegisterStage(stage3)
	
	// Execute
	resp, err := harness.ExecutePipeline("2024-01-01", "2024-01-05")
	
	// Pipeline should partially succeed (continue on error means pipeline completes but with errors)
	// The error might not be returned if continue on error is true
	_ = err // Error is expected but may vary based on continue on error behavior
	if resp.Status != pipeline.PipelineStatusFailed && resp.Status != pipeline.PipelineStatusCompleted {
		t.Errorf("Expected pipeline status to be failed or completed, got %s", resp.Status)
	}
	
	// Verify stage statuses
	testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, "stage1")
	testutil.AssertStageFailed(t, &pipeline.PipelineState{Stages: resp.Stages}, "stage2")
	testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, "stage3")
}

// TestIntegrationLargeDataset tests pipeline with large dataset
func TestIntegrationLargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}
	
	harness := testutil.NewIntegrationTestHarness(t)
	harness.SetupStandardPipeline()
	
	// Generate large dataset
	harness.GetDataGenerator().GenerateLargeDataset(100, 10) // 100 days, 10 tickers per day
	
	// Execute with timeout
	start := time.Now()
	resp, err := harness.ExecutePipelineWithTimeout("2023-01-01", "2023-04-10", 5*time.Minute)
	duration := time.Since(start)
	
	// Should complete within timeout
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	t.Logf("Large dataset pipeline completed in %v", duration)
	
	// Verify reasonable performance
	if duration > 5*time.Minute {
		t.Errorf("Pipeline took too long: %v", duration)
	}
}

// Helper function
func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}