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

func TestNewManager(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	
	manager := pipeline.NewManager(hub, logger)
	
	testutil.AssertNotNil(t, manager)
	testutil.AssertNotNil(t, manager.GetRegistry())
	testutil.AssertNotNil(t, manager.GetConfig())
}

func TestManagerRegisterStage(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Create and register a stage
	stage := testutil.CreateSuccessfulStage("test", "Test Stage")
	
	testutil.AssertNoError(t, manager.RegisterStage(stage))
	
	// Verify stage is in registry
	registry := manager.GetRegistry()
	if !registry.Has("test") {
		t.Error("Stage not found in registry after registration")
	}
}

func TestManagerSetConfig(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Create custom config
	config := pipeline.NewConfigBuilder().
		WithExecutionMode(pipeline.ExecutionModeParallel).
		WithMaxConcurrency(4).
		Build()
	
	manager.SetConfig(config)
	
	// Verify config was set
	gotConfig := manager.GetConfig()
	testutil.AssertEqual(t, gotConfig.ExecutionMode, pipeline.ExecutionModeParallel)
	testutil.AssertEqual(t, gotConfig.MaxConcurrency, 4)
}

func TestManagerExecuteSequential(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Configure for sequential execution
	config := pipeline.NewConfigBuilder().
		WithExecutionMode(pipeline.ExecutionModeSequential).
		Build()
	manager.SetConfig(config)
	
	// Create stages
	stage1 := testutil.CreateSuccessfulStage("stage1", "Stage 1")
	stage2 := testutil.CreateSuccessfulStage("stage2", "Stage 2", "stage1")
	stage3 := testutil.CreateSuccessfulStage("stage3", "Stage 3", "stage2")
	
	manager.RegisterStage(stage1)
	manager.RegisterStage(stage2)
	manager.RegisterStage(stage3)
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-sequential"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify execution order
	testutil.AssertStageOrder(t, []*testutil.MockStage{stage1, stage2, stage3}, 
		[]string{"stage1", "stage2", "stage3"})
}

func TestManagerExecuteWithRetries(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Configure with retries
	config := pipeline.NewConfigBuilder().
		WithRetryConfig(pipeline.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
		}).
		Build()
	manager.SetConfig(config)
	
	// Create stage that fails twice then succeeds
	retryStage := testutil.CreateRetryableStage("retry", "Retry Stage", 2)
	manager.RegisterStage(retryStage)
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-retry"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify stage was called 3 times (2 failures + 1 success)
	testutil.AssertEqual(t, retryStage.GetExecuteCalls(), 3)
	
	// Check logs for retry messages
	warningLogs := logger.GetWarningLogs()
	retryCount := 0
	for _, log := range warningLogs {
		if contains(log.Format, "retrying") {
			retryCount++
		}
	}
	testutil.AssertEqual(t, retryCount, 2)
}

func TestManagerExecuteWithTimeout(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Configure with short timeout
	config := pipeline.NewConfigBuilder().
		WithStageTimeout("slow", 50*time.Millisecond).
		Build()
	manager.SetConfig(config)
	
	// Create slow stage
	slowStage := testutil.CreateSlowStage("slow", "Slow Stage", 200*time.Millisecond)
	manager.RegisterStage(slowStage)
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-timeout"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertError(t, err, true)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusFailed)
	
	// Verify timeout error
	testutil.AssertStageFailed(t, &pipeline.PipelineState{Stages: resp.Stages}, "slow")
}

func TestManagerExecuteWithCancellation(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create stages
	fastStage := testutil.CreateSuccessfulStage("fast", "Fast Stage")
	slowStage := testutil.CreateSlowStage("slow", "Slow Stage", 500*time.Millisecond, "fast")
	
	manager.RegisterStage(fastStage)
	manager.RegisterStage(slowStage)
	
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start pipeline in goroutine
	done := make(chan struct{})
	var resp *pipeline.PipelineResponse
	var err error
	
	go func() {
		req := pipeline.PipelineRequest{ID: "test-cancel"}
		resp, err = manager.Execute(ctx, req)
		close(done)
	}()
	
	// Cancel after short delay
	time.Sleep(100 * time.Millisecond)
	cancel()
	
	// Wait for completion
	<-done
	
	// Should have error
	testutil.AssertError(t, err, true)
	
	// Pipeline should be failed or cancelled
	if resp != nil && resp.Status != pipeline.PipelineStatusFailed && resp.Status != pipeline.PipelineStatusCancelled {
		t.Errorf("Pipeline status = %v, want failed or cancelled", resp.Status)
	}
}

func TestManagerExecuteWithDependencies(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create stages with dependencies
	stage1 := testutil.CreateSuccessfulStage("s1", "Stage 1")
	stage2 := testutil.CreateSuccessfulStage("s2", "Stage 2", "s1")
	stage3 := testutil.CreateSuccessfulStage("s3", "Stage 3", "s1")
	stage4 := testutil.CreateSuccessfulStage("s4", "Stage 4", "s2", "s3")
	
	manager.RegisterStage(stage1)
	manager.RegisterStage(stage2)
	manager.RegisterStage(stage3)
	manager.RegisterStage(stage4)
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-deps"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify all stages completed
	for _, stageID := range []string{"s1", "s2", "s3", "s4"} {
		testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, stageID)
	}
	
	// Verify s1 ran before s2 and s3
	s1Time := stage1.ExecuteArgs[0].Time
	s2Time := stage2.ExecuteArgs[0].Time
	s3Time := stage3.ExecuteArgs[0].Time
	
	if s2Time.Before(s1Time) || s3Time.Before(s1Time) {
		t.Error("Dependent stages ran before their dependency")
	}
	
	// Verify s4 ran after s2 and s3
	s4Time := stage4.ExecuteArgs[0].Time
	if s4Time.Before(s2Time) || s4Time.Before(s3Time) {
		t.Error("Stage 4 ran before its dependencies")
	}
}

func TestManagerExecuteWithFailure(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	config.ContinueOnError = false
	manager.SetConfig(config)
	
	// Create stages where second fails
	stage1 := testutil.CreateSuccessfulStage("s1", "Stage 1")
	stage2 := testutil.CreateFailingStage("s2", "Stage 2", errors.New("stage 2 failed"), "s1")
	stage3 := testutil.CreateSuccessfulStage("s3", "Stage 3", "s2")
	
	manager.RegisterStage(stage1)
	manager.RegisterStage(stage2)
	manager.RegisterStage(stage3)
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-failure"}
	
	resp, err := manager.Execute(ctx, req)
	testutil.AssertError(t, err, true)
	testutil.AssertPipelineStatus(t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusFailed)
	
	// Verify stage statuses
	testutil.AssertStageCompleted(t, &pipeline.PipelineState{Stages: resp.Stages}, "s1")
	testutil.AssertStageFailed(t, &pipeline.PipelineState{Stages: resp.Stages}, "s2")
	testutil.AssertStageSkipped(t, &pipeline.PipelineState{Stages: resp.Stages}, "s3")
	
	// Verify stage 3 was not executed
	testutil.AssertEqual(t, stage3.GetExecuteCalls(), 0)
}

func TestManagerGetPipeline(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Should return error for non-existent pipeline
	_, err1 := manager.GetPipeline("nonexistent")
	testutil.AssertError(t, err1, true)
	
	// Create and execute a pipeline
	stage := testutil.CreateSuccessfulStage("test", "Test")
	manager.RegisterStage(stage)
	
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-get"}
	
	// Start execution in background
	done := make(chan struct{})
	go func() {
		manager.Execute(ctx, req)
		close(done)
	}()
	
	// Wait for pipeline to be registered (with timeout)
	var state *pipeline.PipelineState
	var err error
	for i := 0; i < 10; i++ {
		time.Sleep(10 * time.Millisecond)
		state, err = manager.GetPipeline("test-get")
		if err == nil {
			break
		}
	}
	
	// Should be able to get the pipeline
	testutil.AssertNoError(t, err)
	testutil.AssertEqual(t, state.ID, "test-get")
	
	// Wait for completion
	<-done
}

func TestManagerListPipelines(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Initially should be empty
	pipelines := manager.ListPipelines()
	testutil.AssertEqual(t, len(pipelines), 0)
	
	// Create stage for testing
	stage := testutil.CreateSlowStage("test", "Test", 100*time.Millisecond)
	manager.RegisterStage(stage)
	
	// Start multiple pipelines
	ctx := context.Background()
	count := 3
	done := make(chan struct{}, count)
	
	for i := 0; i < count; i++ {
		go func(n int) {
			req := pipeline.PipelineRequest{ID: fmt.Sprintf("pipeline-%d", n)}
			manager.Execute(ctx, req)
			done <- struct{}{}
		}(i)
	}
	
	// Give them time to start
	time.Sleep(20 * time.Millisecond)
	
	// Should have active pipelines
	pipelines = manager.ListPipelines()
	if len(pipelines) != count {
		t.Errorf("Active pipelines = %d, want %d", len(pipelines), count)
	}
	
	// Wait for completion
	for i := 0; i < count; i++ {
		<-done
	}
}

func TestManagerCancelPipeline(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	// Should error on non-existent pipeline
	err := manager.CancelPipeline("nonexistent")
	testutil.AssertError(t, err, true)
	
	// Create slow stage
	stage := testutil.CreateSlowStage("test", "Test", 500*time.Millisecond)
	manager.RegisterStage(stage)
	
	// Start pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-cancel-mgr"}
	
	done := make(chan struct{})
	go func() {
		manager.Execute(ctx, req)
		close(done)
	}()
	
	// Give it time to start
	time.Sleep(50 * time.Millisecond)
	
	// Cancel the pipeline
	err = manager.CancelPipeline("test-cancel-mgr")
	testutil.AssertNoError(t, err)
	
	// Wait for completion
	<-done
	
	// Check for cancellation status message
	messages := hub.GetMessagesByType(pipeline.EventTypePipelineStatus)
	found := false
	for _, msg := range messages {
		if metadata, ok := msg.Metadata.(map[string]interface{}); ok {
			if status, ok := metadata["status"].(pipeline.PipelineStatus); ok {
				if status == pipeline.PipelineStatusCancelled {
					found = true
					break
				}
			}
		}
	}
	
	if !found {
		t.Error("Expected cancellation status message")
	}
}

func TestManagerWebSocketUpdates(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create stage
	stage := testutil.CreateSuccessfulStage("test", "Test Stage")
	manager.RegisterStage(stage)
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "test-ws"}
	
	_, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Verify WebSocket messages
	messages := hub.GetMessages()
	
	// Should have specific message types
	messageTypes := make(map[string]int)
	for _, msg := range messages {
		messageTypes[msg.EventType]++
	}
	
	// Verify required message types
	requiredTypes := []string{
		pipeline.EventTypePipelineReset,
		pipeline.EventTypePipelineStatus,
		pipeline.EventTypePipelineProgress,
		pipeline.EventTypePipelineComplete,
	}
	
	for _, msgType := range requiredTypes {
		if count := messageTypes[msgType]; count == 0 {
			t.Errorf("Missing WebSocket message type: %s", msgType)
		}
	}
}

func TestManagerConcurrentExecutions(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create stage
	stage := testutil.CreateSuccessfulStage("test", "Test Stage")
	manager.RegisterStage(stage)
	
	// Execute multiple pipelines concurrently
	ctx := context.Background()
	count := 5
	errors := make(chan error, count)
	
	for i := 0; i < count; i++ {
		go func(n int) {
			req := pipeline.PipelineRequest{ID: fmt.Sprintf("concurrent-%d", n)}
			_, err := manager.Execute(ctx, req)
			errors <- err
		}(i)
	}
	
	// Collect results
	for i := 0; i < count; i++ {
		err := <-errors
		testutil.AssertNoError(t, err)
	}
	
	// Verify all pipelines completed
	// Note: They should have been removed from active pipelines after completion
	pipelines := manager.ListPipelines()
	testutil.AssertEqual(t, len(pipelines), 0)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}