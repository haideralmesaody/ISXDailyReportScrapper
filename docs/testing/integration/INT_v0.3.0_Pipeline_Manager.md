# Integration Test Plan: Pipeline Manager

**Version**: v0.3.0-alpha  
**Component**: Pipeline Manager (PIPE-002)  
**Date**: January 2025  
**Document Type**: Integration Test Plan (INT)

## Executive Summary

This document outlines the integration test plan for the new Pipeline Manager implementation. The Pipeline Manager fixes a critical bug where the pipeline would stop after the scraping stage and provides a robust framework for pipeline orchestration.

## Test Objectives

1. Verify Pipeline Manager correctly orchestrates all stages
2. Confirm the fix for pipeline stopping after scraping
3. Validate stage dependency management
4. Test error handling and recovery mechanisms
5. Verify WebSocket integration for real-time updates

## Test Environment

- **OS**: Windows 10/11
- **Go Version**: 1.23+
- **Test Data**: Sample Excel files in `testdata/`
- **Test Mode**: Integration (multiple components)

## Test Suite 1: Stage Registration and Validation

### Test 1.1: Register All Pipeline Stages

**Purpose**: Verify all stages can be registered successfully

**Setup**:
```go
manager := pipeline.NewManager(mockHub, testLogger)
```

**Test Steps**:
1. Register ScrapingStage
2. Register ProcessingStage  
3. Register IndicesStage
4. Register AnalysisStage

**Expected Results**:
- All stages registered without error
- Manager.GetStage(id) returns correct stage
- No duplicate registration allowed

**Test Code**:
```go
func TestStageRegistration(t *testing.T) {
    manager := pipeline.NewManager(mockHub, testLogger)
    
    // Register stages
    stages := []pipeline.Stage{
        stages.NewScrapingStage(".", logger),
        stages.NewProcessingStage(".", logger),
        stages.NewIndicesStage(".", logger),
        stages.NewAnalysisStage(".", logger),
    }
    
    for _, stage := range stages {
        err := manager.RegisterStage(stage)
        assert.NoError(t, err)
    }
    
    // Verify registration
    assert.Equal(t, 4, manager.StageCount())
    
    // Test duplicate registration
    err := manager.RegisterStage(stages[0])
    assert.Error(t, err)
}
```

### Test 1.2: Validate Stage Dependencies

**Purpose**: Ensure dependency validation works correctly

**Test Cases**:
1. Valid dependencies (processing depends on scraping)
2. Missing dependency (stage depends on non-existent stage)
3. Circular dependency detection

**Expected Results**:
- Valid dependencies pass validation
- Missing dependencies return error
- Circular dependencies detected and rejected

### Test 1.3: Stage Validation Checks

**Purpose**: Verify stage pre-execution validation

**Test Steps**:
1. Create pipeline state with missing configuration
2. Call Validate() on each stage
3. Fix configuration and revalidate

**Expected Results**:
- Validation fails with clear error messages
- After fixing config, validation passes

## Test Suite 2: Pipeline Execution

### Test 2.1: Successful Full Pipeline Execution

**Purpose**: Verify complete pipeline runs successfully

**Test Setup**:
- Place test Excel files in downloads directory
- Clear reports directory
- Configure pipeline request

**Test Steps**:
1. Create PipelineRequest with date range
2. Execute pipeline
3. Monitor progress updates
4. Verify all stages complete

**Expected Results**:
- All 4 stages execute in order
- Progress updates sent via WebSocket
- Output files created correctly
- Pipeline status = completed

**Test Code**:
```go
func TestFullPipelineExecution(t *testing.T) {
    // Setup
    setupTestFiles(t)
    defer cleanupTestFiles(t)
    
    manager := createTestManager(t)
    
    // Create request
    req := pipeline.PipelineRequest{
        ID:       "test-pipeline-1",
        FromDate: "2025-01-01",
        ToDate:   "2025-01-20",
        Parameters: map[string]interface{}{
            pipeline.ContextKeyDownloadDir: "testdata/downloads",
            pipeline.ContextKeyReportDir:   "testdata/reports",
        },
    }
    
    // Execute
    ctx := context.Background()
    resp, err := manager.Execute(ctx, req)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, pipeline.PipelineStatusCompleted, resp.Status)
    assert.Equal(t, 4, len(resp.Stages))
    
    // Check output files
    assert.FileExists(t, "testdata/reports/isx_combined_data.csv")
    assert.FileExists(t, "testdata/reports/indexes.csv")
    assert.FileExists(t, "testdata/reports/ticker_summary.json")
}
```

### Test 2.2: Pipeline With Stage Failure

**Purpose**: Test pipeline behavior when a stage fails

**Test Scenarios**:
1. Scraping fails (no files downloaded)
2. Processing fails (corrupted Excel)
3. Indices fails (missing data)
4. Analysis fails (invalid CSV)

**Expected Results**:
- Pipeline stops at failed stage
- Later stages remain pending
- Error details available in response
- WebSocket sends error notification

### Test 2.3: Pipeline Cancellation

**Purpose**: Verify pipeline can be cancelled mid-execution

**Test Steps**:
1. Start pipeline execution
2. Cancel context during processing stage
3. Check pipeline state

**Expected Results**:
- Current stage marked as cancelled
- Pending stages not executed
- Cleanup performed correctly

### Test 2.4: Pipeline Skip Conditions

**Purpose**: Test conditional stage execution

**Test Cases**:
1. Skip scraping if files already exist
2. Skip analysis if no new data

**Expected Results**:
- Skipped stages marked appropriately
- Dependent stages still execute
- Skip reasons logged

## Test Suite 3: Progress Tracking

### Test 3.1: Real-time Progress Updates

**Purpose**: Verify progress updates are sent correctly

**Test Setup**:
- Mock WebSocket hub to capture messages
- Enable progress tracking

**Test Verification**:
- Progress messages sent regularly
- Progress values between 0-100
- Stage metadata included
- ETA calculations present

**Test Code**:
```go
func TestProgressTracking(t *testing.T) {
    mockHub := &MockWebSocketHub{
        messages: []WebSocketMessage{},
    }
    
    manager := pipeline.NewManager(mockHub, testLogger)
    // ... register stages ...
    
    // Execute pipeline
    req := createTestRequest()
    _, err := manager.Execute(context.Background(), req)
    
    // Verify progress messages
    assert.NoError(t, err)
    assert.Greater(t, len(mockHub.messages), 10) // Multiple progress updates
    
    // Check message format
    for _, msg := range mockHub.messages {
        if msg.Type == "pipeline_progress" {
            assert.Contains(t, msg.Data, "stage")
            assert.Contains(t, msg.Data, "progress")
            assert.Contains(t, msg.Data, "message")
        }
    }
}
```

### Test 3.2: Progress Persistence

**Purpose**: Verify progress can be recovered

**Test Steps**:
1. Start pipeline execution
2. Simulate crash mid-execution
3. Query pipeline state
4. Verify progress data intact

### Test 3.3: Concurrent Progress Updates

**Purpose**: Test thread-safe progress updates

**Test Steps**:
1. Execute multiple stages concurrently
2. Send progress updates from each
3. Verify no race conditions

## Test Suite 4: Error Recovery

### Test 4.1: Retry Mechanism

**Purpose**: Test automatic retry for transient failures

**Test Setup**:
- Configure stage to fail first 2 attempts
- Set retry count to 3

**Expected Results**:
- Stage retried automatically
- Succeeds on 3rd attempt
- Pipeline continues normally

### Test 4.2: Partial Pipeline Recovery

**Purpose**: Test resuming failed pipeline

**Test Steps**:
1. Execute pipeline that fails at processing
2. Fix issue
3. Resume pipeline from processing stage

**Expected Results**:
- Completed stages not re-executed
- Pipeline resumes from failure point
- Final state correct

### Test 4.3: Error Propagation

**Purpose**: Verify errors propagate correctly

**Test Cases**:
1. Validation errors
2. Execution errors
3. Cancellation errors
4. System errors

**Expected Results**:
- Error type preserved
- Error context included
- Stack trace available (debug mode)

## Test Suite 5: WebSocket Integration

### Test 5.1: Message Delivery

**Purpose**: Verify all WebSocket messages delivered

**Test Verification**:
- Pipeline start notification
- Stage status updates
- Progress messages
- Completion notification
- Error notifications

### Test 5.2: Message Ordering

**Purpose**: Ensure messages arrive in correct order

**Test Steps**:
1. Execute pipeline
2. Capture all WebSocket messages
3. Verify chronological order

### Test 5.3: Multi-client Updates

**Purpose**: Test broadcasting to multiple clients

**Test Setup**:
- Connect 3 WebSocket clients
- Execute pipeline
- Verify all receive updates

## Performance Tests

### Test P1: Large Dataset Processing

**Purpose**: Test pipeline with 1 year of data

**Metrics to Measure**:
- Total execution time
- Memory usage per stage
- Progress update frequency
- Resource cleanup

**Acceptance Criteria**:
- Completes within 30 minutes
- Memory usage < 2GB
- No memory leaks

### Test P2: Concurrent Pipeline Execution

**Purpose**: Test multiple pipelines running simultaneously

**Test Setup**:
- Execute 3 pipelines with different date ranges
- Monitor resource usage
- Verify isolation

**Expected Results**:
- All pipelines complete successfully
- No interference between pipelines
- Resource usage scales linearly

## Integration Points

### Test I1: Web Application Integration

**Purpose**: Verify handleScrape uses Pipeline Manager correctly

**Test Steps**:
1. Send HTTP request to /api/scrape
2. Verify pipeline starts
3. Check response format
4. Monitor WebSocket updates

### Test I2: File System Integration

**Purpose**: Test file operations during pipeline

**Verification**:
- Files created in correct locations
- Permissions set correctly
- Cleanup on failure
- No file locks left

### Test I3: Process Execution

**Purpose**: Verify external executables called correctly

**Test Cases**:
- scraper.exe with correct arguments
- process.exe receives input directory
- indexcsv.exe creates output
- Exit codes handled properly

## Test Data Requirements

### Input Data
- 10 sample Excel files (various formats)
- 1 corrupted Excel file (for error testing)
- Empty Excel file
- Large Excel file (10MB+)

### Expected Output
- isx_combined_data.csv with all records
- indexes.csv with ISX60/ISX15 data
- ticker_summary.json with statistics
- Consistent data across all outputs

## Test Execution Plan

### Phase 1: Unit Integration (2 hours)
1. Stage registration tests
2. Validation tests
3. Basic execution tests

### Phase 2: Full Integration (4 hours)
1. Complete pipeline tests
2. Error handling tests
3. Progress tracking tests

### Phase 3: Performance (2 hours)
1. Large dataset tests
2. Concurrent execution tests
3. Resource usage monitoring

### Phase 4: System Integration (2 hours)
1. Web application integration
2. WebSocket communication
3. End-to-end validation

## Success Criteria

1. **Bug Fix Verified**: Pipeline completes all 4 stages automatically
2. **No Regressions**: Existing functionality still works
3. **Performance**: No significant slowdown vs previous version
4. **Reliability**: 100% success rate in normal conditions
5. **Error Handling**: Graceful failure with clear messages

## Risk Mitigation

1. **Test Data Corruption**: Keep backup of test data
2. **Environment Issues**: Document exact test environment
3. **Timing Issues**: Use proper synchronization in tests
4. **Resource Leaks**: Monitor with profiler during tests

## Appendix A: Test Utilities

```go
// MockWebSocketHub for capturing messages
type MockWebSocketHub struct {
    messages []WebSocketMessage
    mu       sync.Mutex
}

func (m *MockWebSocketHub) BroadcastUpdate(eventType, stage, status string, metadata interface{}) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.messages = append(m.messages, WebSocketMessage{
        Type:     eventType,
        Stage:    stage,
        Status:   status,
        Metadata: metadata,
    })
}

// Test data setup
func setupTestFiles(t *testing.T) {
    // Create test directories
    os.MkdirAll("testdata/downloads", 0755)
    os.MkdirAll("testdata/reports", 0755)
    
    // Copy sample Excel files
    copyTestFile(t, "samples/2025-01-01.xlsx", "testdata/downloads/")
    copyTestFile(t, "samples/2025-01-02.xlsx", "testdata/downloads/")
}
```

## Appendix B: Common Assertions

```go
// Pipeline completion assertion
func assertPipelineCompleted(t *testing.T, resp *pipeline.PipelineResponse) {
    assert.Equal(t, pipeline.PipelineStatusCompleted, resp.Status)
    for _, stage := range resp.Stages {
        assert.Equal(t, pipeline.StageStatusCompleted, stage.Status)
        assert.Equal(t, 100.0, stage.Progress)
    }
}

// File output assertion
func assertOutputFilesExist(t *testing.T, reportDir string) {
    files := []string{
        filepath.Join(reportDir, "isx_combined_data.csv"),
        filepath.Join(reportDir, "indexes.csv"),
        filepath.Join(reportDir, "ticker_summary.json"),
    }
    
    for _, file := range files {
        assert.FileExists(t, file)
        stat, err := os.Stat(file)
        assert.NoError(t, err)
        assert.Greater(t, stat.Size(), int64(0))
    }
}
```

---

*This integration test plan ensures comprehensive validation of the Pipeline Manager implementation and verifies the fix for the critical pipeline orchestration bug.*