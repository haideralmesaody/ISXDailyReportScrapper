package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"isxcli/internal/pipeline"
)

// IntegrationTestHarness provides utilities for integration testing
type IntegrationTestHarness struct {
	t         *testing.T
	manager   *pipeline.Manager
	hub       *MockWebSocketHub
	logger    *MockLogger
	dataGen   *TestDataGenerator
	baseDir   string
}

// NewIntegrationTestHarness creates a new test harness
func NewIntegrationTestHarness(t *testing.T) *IntegrationTestHarness {
	baseDir := CreateTestDirectory(t, "integration-test")
	hub := &MockWebSocketHub{}
	logger := &MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	return &IntegrationTestHarness{
		t:       t,
		manager: manager,
		hub:     hub,
		logger:  logger,
		dataGen: NewTestDataGenerator(t, baseDir),
		baseDir: baseDir,
	}
}

// SetupStandardPipeline sets up a standard 4-stage pipeline
func (h *IntegrationTestHarness) SetupStandardPipeline() {
	// Create standard stages with test implementations
	scrapingStage := CreateSuccessfulStage(pipeline.StageIDScraping, pipeline.StageNameScraping)
	processingStage := CreateSuccessfulStage(pipeline.StageIDProcessing, pipeline.StageNameProcessing, pipeline.StageIDScraping)
	indicesStage := CreateSuccessfulStage(pipeline.StageIDIndices, pipeline.StageNameIndices, pipeline.StageIDProcessing)
	analysisStage := CreateSuccessfulStage(pipeline.StageIDAnalysis, pipeline.StageNameAnalysis, pipeline.StageIDIndices)
	
	h.manager.RegisterStage(scrapingStage)
	h.manager.RegisterStage(processingStage)
	h.manager.RegisterStage(indicesStage)
	h.manager.RegisterStage(analysisStage)
}

// ExecutePipeline executes a pipeline with standard configuration
func (h *IntegrationTestHarness) ExecutePipeline(fromDate, toDate string) (*pipeline.PipelineResponse, error) {
	req := pipeline.PipelineRequest{
		ID:       "test-pipeline",
		Mode:     pipeline.ModeInitial,
		FromDate: fromDate,
		ToDate:   toDate,
		Parameters: map[string]interface{}{
			pipeline.ContextKeyDownloadDir: h.baseDir + "/downloads",
			pipeline.ContextKeyReportDir:   h.baseDir + "/reports",
		},
	}
	
	ctx := context.Background()
	return h.manager.Execute(ctx, req)
}

// ExecutePipelineWithTimeout executes a pipeline with a timeout
func (h *IntegrationTestHarness) ExecutePipelineWithTimeout(fromDate, toDate string, timeout time.Duration) (*pipeline.PipelineResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	req := pipeline.PipelineRequest{
		ID:       "test-pipeline-timeout",
		Mode:     pipeline.ModeInitial,
		FromDate: fromDate,
		ToDate:   toDate,
		Parameters: map[string]interface{}{
			pipeline.ContextKeyDownloadDir: h.baseDir + "/downloads",
			pipeline.ContextKeyReportDir:   h.baseDir + "/reports",
		},
	}
	
	return h.manager.Execute(ctx, req)
}

// GetManager returns the pipeline manager
func (h *IntegrationTestHarness) GetManager() *pipeline.Manager {
	return h.manager
}

// GetHub returns the mock WebSocket hub
func (h *IntegrationTestHarness) GetHub() *MockWebSocketHub {
	return h.hub
}

// GetLogger returns the mock logger
func (h *IntegrationTestHarness) GetLogger() *MockLogger {
	return h.logger
}

// GetDataGenerator returns the test data generator
func (h *IntegrationTestHarness) GetDataGenerator() *TestDataGenerator {
	return h.dataGen
}

// GetBaseDir returns the base test directory
func (h *IntegrationTestHarness) GetBaseDir() string {
	return h.baseDir
}

// AssertPipelineSuccess verifies a pipeline completed successfully
func (h *IntegrationTestHarness) AssertPipelineSuccess(resp *pipeline.PipelineResponse, err error) {
	h.t.Helper()
	
	AssertNoError(h.t, err)
	AssertPipelineStatus(h.t, &pipeline.PipelineState{Status: resp.Status}, pipeline.PipelineStatusCompleted)
	
	// Verify all stages completed
	for _, stage := range resp.Stages {
		AssertStageStatus(h.t, stage, pipeline.StageStatusCompleted)
		AssertProgress(h.t, stage, 100)
	}
}

// AssertWebSocketMessages verifies expected WebSocket messages were sent
func (h *IntegrationTestHarness) AssertWebSocketMessages() {
	h.t.Helper()
	
	// Check for required message types
	AssertWebSocketMessage(h.t, h.hub, pipeline.EventTypePipelineReset)
	AssertWebSocketMessage(h.t, h.hub, pipeline.EventTypePipelineStatus)
	AssertWebSocketMessage(h.t, h.hub, pipeline.EventTypePipelineProgress)
	AssertWebSocketMessage(h.t, h.hub, pipeline.EventTypePipelineComplete)
}

// ClearMessages clears all captured messages
func (h *IntegrationTestHarness) ClearMessages() {
	h.hub.Clear()
	h.logger.Clear()
}

// WaitForPipelineCompletion waits for a pipeline to complete
func (h *IntegrationTestHarness) WaitForPipelineCompletion(pipelineID string, timeout time.Duration) (*pipeline.PipelineState, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		state, err := h.manager.GetPipeline(pipelineID)
		if err == nil && (state.Status == pipeline.PipelineStatusCompleted || 
			state.Status == pipeline.PipelineStatusFailed ||
			state.Status == pipeline.PipelineStatusCancelled) {
			return state, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	return nil, &TimeoutError{Message: "pipeline did not complete in time"}
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return e.Message
}

// RunConcurrentPipelines runs multiple pipelines concurrently
func (h *IntegrationTestHarness) RunConcurrentPipelines(count int) []error {
	errors := make(chan error, count)
	
	for i := 0; i < count; i++ {
		go func(n int) {
			req := pipeline.PipelineRequest{
				ID:       fmt.Sprintf("concurrent-%d", n),
				Mode:     pipeline.ModeInitial,
				FromDate: "2024-01-01",
				ToDate:   "2024-01-05",
				Parameters: map[string]interface{}{
					pipeline.ContextKeyDownloadDir: h.baseDir + "/downloads",
					pipeline.ContextKeyReportDir:   h.baseDir + "/reports",
				},
			}
			
			_, err := h.manager.Execute(context.Background(), req)
			errors <- err
		}(i)
	}
	
	// Collect results
	var results []error
	for i := 0; i < count; i++ {
		results = append(results, <-errors)
	}
	
	return results
}

// SimulateStageFailure creates a stage that will fail
func (h *IntegrationTestHarness) SimulateStageFailure(stageID string, errorMsg string) {
	failingStage := CreateFailingStage(stageID, "Failing "+stageID, fmt.Errorf(errorMsg))
	
	// Replace existing stage if present
	stages := []pipeline.Stage{
		failingStage,
	}
	
	// Re-register all stages
	for _, stage := range stages {
		h.manager.RegisterStage(stage)
	}
}