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

// TestWebSocketMessageFlow simulates the complete WebSocket message flow
// that the frontend expects during pipeline execution
func TestWebSocketMessageFlow(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create realistic stages
	stages := createRealisticPipelineStages()
	for _, stage := range stages {
		testutil.AssertNoError(t, manager.RegisterStage(stage))
	}
	
	// Execute pipeline
	ctx := context.Background()
	req := pipeline.PipelineRequest{
		ID:       "frontend-test",
		Mode:     pipeline.ModeInitial,
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}
	
	_, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Analyze WebSocket messages
	messages := hub.GetMessages()
	
	// Verify message sequence
	expectedSequence := []string{
		pipeline.EventTypePipelineReset,
		pipeline.EventTypePipelineStatus,  // Pipeline started
		pipeline.EventTypePipelineProgress, // Multiple progress updates
		pipeline.EventTypePipelineComplete,
	}
	
	// Check minimum expected messages
	if len(messages) < len(expectedSequence) {
		t.Errorf("Expected at least %d messages, got %d", len(expectedSequence), len(messages))
	}
	
	// Verify first message is reset
	if messages[0].EventType != pipeline.EventTypePipelineReset {
		t.Errorf("First message should be reset, got %s", messages[0].EventType)
	}
	
	// Verify last message is complete
	lastMsg := messages[len(messages)-1]
	if lastMsg.EventType != pipeline.EventTypePipelineComplete {
		t.Errorf("Last message should be complete, got %s", lastMsg.EventType)
	}
	
	// Verify progress messages have required fields
	progressMessages := hub.GetMessagesByType(pipeline.EventTypePipelineProgress)
	for _, msg := range progressMessages {
		metadata, ok := msg.Metadata.(map[string]interface{})
		if !ok {
			t.Error("Progress message metadata should be a map")
			continue
		}
		
		// Check required fields
		requiredFields := []string{"pipeline_id", "stage", "status", "progress"}
		for _, field := range requiredFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("Progress message missing required field: %s", field)
			}
		}
	}
}

// TestWebSocketProgressUpdates verifies progress messages match frontend expectations
func TestWebSocketProgressUpdates(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create a stage that sends specific progress updates
	progressStage := testutil.NewStageBuilder("test-progress", "Progress Test").
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage("test-progress")
			
			// Simulate progress updates like the real scraper
			updates := []struct {
				progress float64
				message  string
			}{
				{0, "Starting..."},
				{25, "Connecting to ISX website..."},
				{50, "Downloading reports..."},
				{75, "Processing data..."},
				{100, "Completed"},
			}
			
			for _, update := range updates {
				stageState.UpdateProgress(update.progress, update.message)
				time.Sleep(10 * time.Millisecond)
			}
			
			return nil
		}).
		Build()
	
	testutil.AssertNoError(t, manager.RegisterStage(progressStage))
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "progress-test"}
	
	_, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Check progress messages
	progressMessages := hub.GetMessagesByType(pipeline.EventTypePipelineProgress)
	
	// Should have at least 2 progress updates (start and complete)
	if len(progressMessages) < 2 {
		t.Errorf("Expected at least 2 progress messages, got %d", len(progressMessages))
	}
	
	// Verify progress values are increasing
	var lastProgress float64 = -1
	for _, msg := range progressMessages {
		metadata := msg.Metadata.(map[string]interface{})
		progress, ok := metadata["progress"].(float64)
		if !ok {
			t.Error("Progress should be a float64")
			continue
		}
		
		if progress < lastProgress {
			t.Errorf("Progress decreased: %f -> %f", lastProgress, progress)
		}
		lastProgress = progress
	}
}

// TestWebSocketErrorMessages verifies error reporting to frontend
func TestWebSocketErrorMessages(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create a failing stage
	failStage := testutil.CreateFailingStage("fail-stage", "Fail Stage", 
		pipeline.NewExecutionError("fail-stage", errors.New("simulated failure"), false))
	
	testutil.AssertNoError(t, manager.RegisterStage(failStage))
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "error-test"}
	
	_, err := manager.Execute(ctx, req)
	testutil.AssertError(t, err, true)
	
	// Check for error message
	errorMessages := hub.GetMessagesByType(pipeline.EventTypePipelineError)
	if len(errorMessages) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(errorMessages))
	}
	
	if len(errorMessages) > 0 {
		metadata := errorMessages[0].Metadata.(map[string]interface{})
		
		// Should have error details
		if _, ok := metadata["error"]; !ok {
			t.Error("Error message should contain error details")
		}
	}
}

// TestWebSocketStageTransitions verifies stage status transitions
func TestWebSocketStageTransitions(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Track stage transitions
	stageTransitions := make(map[string][]string)
	
	// Create a stage that we can monitor
	monitoredStage := testutil.NewStageBuilder("monitored", "Monitored Stage").
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		}).
		Build()
	
	testutil.AssertNoError(t, manager.RegisterStage(monitoredStage))
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{ID: "transition-test"}
	
	_, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Analyze progress messages for stage transitions
	progressMessages := hub.GetMessagesByType(pipeline.EventTypePipelineProgress)
	
	for _, msg := range progressMessages {
		metadata := msg.Metadata.(map[string]interface{})
		if stage, ok := metadata["stage"].(string); ok {
			if status, ok := metadata["status"].(pipeline.StageStatus); ok {
				stageTransitions[stage] = append(stageTransitions[stage], string(status))
			}
		}
	}
	
	// Verify stage went through expected transitions
	transitions := stageTransitions["monitored"]
	expectedTransitions := []string{
		string(pipeline.StageStatusActive),
		string(pipeline.StageStatusCompleted),
	}
	
	if len(transitions) < len(expectedTransitions) {
		t.Errorf("Expected at least %d transitions, got %d", len(expectedTransitions), len(transitions))
	}
}

// TestFrontendCompatibleMessages ensures messages are compatible with current frontend
func TestFrontendCompatibleMessages(t *testing.T) {
	hub := &testutil.MockWebSocketHub{}
	logger := &testutil.MockLogger{}
	manager := pipeline.NewManager(hub, logger)
	
	config := testutil.CreateTestConfig()
	manager.SetConfig(config)
	
	// Create stages that mimic real pipeline
	scrapingStage := createMockScrapingStage()
	processingStage := createMockProcessingStage()
	
	testutil.AssertNoError(t, manager.RegisterStage(scrapingStage))
	testutil.AssertNoError(t, manager.RegisterStage(processingStage))
	
	// Execute
	ctx := context.Background()
	req := pipeline.PipelineRequest{
		ID:       "frontend-compat-test",
		Mode:     pipeline.ModeInitial,
		FromDate: "2024-01-01",
		ToDate:   "2024-01-31",
	}
	
	_, err := manager.Execute(ctx, req)
	testutil.AssertNoError(t, err)
	
	// Verify specific message formats the frontend expects
	_ = hub.GetMessages()
	
	// Check pipeline_status messages
	statusMessages := hub.GetMessagesByType(pipeline.EventTypePipelineStatus)
	for _, msg := range statusMessages {
		metadata := msg.Metadata.(map[string]interface{})
		
		// Must have pipeline_id
		if _, ok := metadata["pipeline_id"]; !ok {
			t.Error("pipeline_status message missing pipeline_id")
		}
		
		// Must have status
		if _, ok := metadata["status"]; !ok {
			t.Error("pipeline_status message missing status")
		}
	}
	
	// Check pipeline_complete message format
	completeMessages := hub.GetMessagesByType(pipeline.EventTypePipelineComplete)
	if len(completeMessages) > 0 {
		metadata := completeMessages[0].Metadata.(map[string]interface{})
		
		// Frontend expects these fields
		expectedFields := []string{"pipeline_id", "status"}
		for _, field := range expectedFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("pipeline_complete missing expected field: %s", field)
			}
		}
	}
}

// Helper functions to create realistic stages

func createRealisticPipelineStages() []pipeline.Stage {
	return []pipeline.Stage{
		createMockScrapingStage(),
		createMockProcessingStage(),
		createMockIndicesStage(),
		createMockAnalysisStage(),
	}
}

func createMockScrapingStage() *testutil.MockStage {
	return testutil.NewStageBuilder(pipeline.StageIDScraping, pipeline.StageNameScraping).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage(pipeline.StageIDScraping)
			
			// Simulate real scraper progress
			stageState.UpdateProgress(0, "Initializing scraper...")
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(25, "Navigating to ISX website...")
			stageState.Metadata["url"] = "https://www.isx-iq.net"
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(50, "Downloading reports for January 2024...")
			stageState.Metadata["current_file"] = "2024 01 15 ISX Daily Report.xlsx"
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(75, "Saving files to disk...")
			stageState.Metadata["files_downloaded"] = 20
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(100, "Scraping completed successfully")
			state.SetContext(pipeline.ContextKeyScraperSuccess, true)
			
			return nil
		}).
		Build()
}

func createMockProcessingStage() *testutil.MockStage {
	return testutil.NewStageBuilder(pipeline.StageIDProcessing, pipeline.StageNameProcessing).
		WithDependencies(pipeline.StageIDScraping).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage(pipeline.StageIDProcessing)
			
			// Simulate processing multiple files
			totalFiles := 20
			for i := 0; i < totalFiles; i++ {
				progress := float64(i+1) / float64(totalFiles) * 100
				stageState.UpdateProgress(progress, fmt.Sprintf("Processing file %d of %d", i+1, totalFiles))
				stageState.Metadata["current_file"] = fmt.Sprintf("2024 01 %02d ISX Daily Report.xlsx", i+1)
				stageState.Metadata["records_processed"] = (i + 1) * 150
				time.Sleep(5 * time.Millisecond)
			}
			
			return nil
		}).
		Build()
}

func createMockIndicesStage() *testutil.MockStage {
	return testutil.NewStageBuilder(pipeline.StageIDIndices, pipeline.StageNameIndices).
		WithDependencies(pipeline.StageIDProcessing).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage(pipeline.StageIDIndices)
			
			stageState.UpdateProgress(33, "Extracting ISX60 index...")
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(66, "Extracting ISX15 index...")
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(100, "Index extraction completed")
			stageState.Metadata["indices_extracted"] = 2
			
			return nil
		}).
		Build()
}

func createMockAnalysisStage() *testutil.MockStage {
	return testutil.NewStageBuilder(pipeline.StageIDAnalysis, pipeline.StageNameAnalysis).
		WithDependencies(pipeline.StageIDIndices).
		WithExecute(func(ctx context.Context, state *pipeline.PipelineState) error {
			stageState := state.GetStage(pipeline.StageIDAnalysis)
			
			stageState.UpdateProgress(50, "Calculating ticker statistics...")
			time.Sleep(20 * time.Millisecond)
			
			stageState.UpdateProgress(100, "Analysis completed")
			stageState.Metadata["tickers_analyzed"] = 104
			
			return nil
		}).
		Build()
}