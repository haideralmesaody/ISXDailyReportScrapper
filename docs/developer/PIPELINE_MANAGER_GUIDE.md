# Pipeline Manager Developer Guide

## Overview

The Pipeline Manager is a centralized orchestration system for managing the data processing pipeline in the ISX Daily Reports Scrapper. It provides a clean, extensible framework for executing pipeline stages with automatic dependency management, progress tracking, and error handling.

## Current Status (v0.3.0-alpha)

✅ **PRODUCTION READY**: The Pipeline Manager is now the active pipeline execution system, having replaced the previous conditional logic approach. It successfully fixed the critical bug where the pipeline would stop after the scraping stage.

### Key Achievements
- **Simplified Code**: Reduced handleScrape from 200+ lines to ~50 lines
- **Fixed Critical Bug**: Pipeline now correctly executes all stages
- **Enhanced Progress**: Integrated with historical metrics for better ETA predictions
- **Improved Reliability**: Eliminated race conditions and state propagation issues

## Architecture

### Core Components

1. **Pipeline Manager** (`internal/pipeline/manager.go`)
   - Central orchestrator for all pipeline executions
   - Manages multiple concurrent pipelines
   - Handles stage registration and dependency resolution
   - Provides progress monitoring and state management

2. **Pipeline State** (`internal/pipeline/types.go`)
   - Tracks the current state of a pipeline execution
   - Maintains context data between stages
   - Records timing and progress information

3. **Stage Interface** (`internal/pipeline/stage.go`)
   - Defines the contract for pipeline stages
   - Enables modular, reusable components
   - Supports dependency declaration

4. **WebSocket Integration** (`cmd/web-licensed/websocket_adapter.go`)
   - Bridges pipeline events to the WebSocket hub
   - Provides real-time updates to the frontend
   - Maintains backward compatibility

5. **Progress Tracking Integration** (`internal/progress`)
   - **EnhancedCalculator**: Provides historical metrics for accurate ETAs
   - **MetricsManager**: Persists timing data in `data/metrics/` directory
   - **Dynamic Adjustment**: Updates progress based on actual findings
   - **Immediate Estimates**: Shows ETAs even before processing starts

## Quick Start

### Using the Pipeline Manager

```go
// 1. Create a pipeline manager
wsAdapter := NewWebSocketAdapter(wsHub)
logger := NewPipelineLogger("pipeline")
pipelineManager := pipeline.NewManager(wsAdapter, logger)

// 2. Register stages
pipelineManager.RegisterStage(stages.NewScrapingStage(execDir, logger))
pipelineManager.RegisterStage(stages.NewProcessingStage(execDir, logger))
pipelineManager.RegisterStage(stages.NewIndicesStage(execDir, logger))
pipelineManager.RegisterStage(stages.NewAnalysisStage(execDir, logger))

// 3. Create a pipeline request
req := pipeline.PipelineRequest{
    ID:       "pipeline-123",
    Mode:     "initial",
    FromDate: "2025-01-01",
    ToDate:   "2025-01-20",
    Parameters: map[string]interface{}{
        pipeline.ContextKeyDownloadDir: "data/downloads",
        pipeline.ContextKeyReportDir:   "data/reports",
    },
}

// 4. Execute the pipeline
ctx := context.Background()
resp, err := pipelineManager.Execute(ctx, req)
if err != nil {
    log.Printf("Pipeline failed: %v", err)
    return
}

log.Printf("Pipeline completed: %s", resp.Status)
```

## Creating a New Stage

### Step 1: Implement the Stage Interface

```go
package stages

import (
    "context"
    "isxcli/internal/pipeline"
    "isxcli/internal/common"
)

type MyNewStage struct {
    pipeline.BaseStage
    executableDir string
    logger        *common.Logger
}

func NewMyNewStage(executableDir string, logger *common.Logger) *MyNewStage {
    return &MyNewStage{
        BaseStage: pipeline.NewBaseStage(
            "mynewstage",              // Stage ID
            "My New Stage",            // Display name
            []string{"processing"},    // Dependencies
        ),
        executableDir: executableDir,
        logger:        logger,
    }
}
```

### Step 2: Implement Execute Method

```go
func (s *MyNewStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
    stageState := state.GetStage(s.ID())
    
    // Update progress
    stageState.UpdateProgress(10, "Starting my new stage...")
    
    // Get configuration
    inputDir, _ := state.GetConfig(pipeline.ContextKeyReportDir)
    
    // Do your work here
    err := s.processData(ctx, inputDir, stageState)
    if err != nil {
        return pipeline.NewExecutionError(s.ID(), err, false)
    }
    
    // Update final progress
    stageState.UpdateProgress(100, "My new stage completed")
    
    return nil
}
```

### Step 3: Implement Validate Method

```go
func (s *MyNewStage) Validate(state *pipeline.PipelineState) error {
    // Check prerequisites
    reportDir, _ := state.GetConfig(pipeline.ContextKeyReportDir)
    if reportDir == "" {
        return pipeline.NewValidationError(s.ID(), "report directory not configured")
    }
    
    // Check if required files exist
    requiredFile := filepath.Join(reportDir, "required_input.csv")
    if _, err := os.Stat(requiredFile); err != nil {
        return pipeline.NewValidationError(s.ID(), "required input file not found")
    }
    
    return nil
}
```

### Step 4: Register the Stage

```go
// In web-application.go
pipelineManager.RegisterStage(stages.NewMyNewStage(executableDir, logger))
```

## Progress Tracking (Enhanced in v0.3.0-alpha)

### Production Progress Tracking System

The Pipeline Manager now integrates with the enhanced progress tracking system that provides historical metrics, immediate ETA estimates, and dynamic adjustments.

### Using Enhanced Progress Tracking

For stages that execute external commands with historical metrics:

```go
func (s *MyNewStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
    stageState := state.GetStage(s.ID())
    
    // Create command
    cmd := exec.CommandContext(ctx, "mycommand.exe", args...)
    
    // Attach progress parser
    progressParser := NewProgressParser(stageState)
    cmd.Stdout = progressParser
    cmd.Stderr = progressParser
    
    // Execute
    return cmd.Run()
}
```

### Manual Progress Updates

```go
func (s *MyNewStage) processFiles(stageState *pipeline.StageState, files []string) error {
    total := len(files)
    
    for i, file := range files {
        // Update progress
        progress := float64(i+1) / float64(total) * 100
        stageState.UpdateProgress(progress, fmt.Sprintf("Processing %s", file))
        
        // Process file
        if err := s.processFile(file); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Enhanced Progress with Historical Metrics (New in v0.3.0)

For executables that use the enhanced progress system:

```go
// In your executable (e.g., mynewstage.exe)
import "isxcli/internal/progress"

func main() {
    // Initialize metrics manager
    dataPath := "data" // or filepath.Dir(outputDir)
    metricsManager := progress.NewMetricsManager(dataPath)
    
    // Create enhanced calculator with historical metrics
    totalItems := len(filesToProcess)
    calc := progress.NewEnhancedCalculator("mynewstage", totalItems, metricsManager)
    
    // Process items with progress tracking
    for i, item := range filesToProcess {
        // Update progress
        calc.Update(i)
        
        // Send structured progress message
        details := map[string]interface{}{
            "current_file": item.Name,
            "size_mb":      item.SizeMB,
            "elapsed":      time.Since(calc.StartTime).String(),
        }
        
        jsonData, _ := calc.ToJSON(fmt.Sprintf("Processing %s", item.Name), details)
        fmt.Printf("[WEBSOCKET_PROGRESS] %s\n", jsonData)
        
        // Do the actual work
        if err := processItem(item); err != nil {
            return err
        }
    }
    
    // Complete and save metrics for future runs
    calc.Complete()
    
    // Send completion status
    jsonData, _ := progress.StatusToJSON("mynewstage", "completed", "All items processed successfully")
    fmt.Printf("[WEBSOCKET_STATUS] %s\n", jsonData)
    
    // Important: Flush stdout before exit
    os.Stdout.Sync()
}
```

### Dynamic Progress Adjustment Example

```go
// For stages that discover actual work during execution
func (s *ScrapingStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
    // Initial estimate based on date range
    initialEstimate := calculateBusinessDays(fromDate, toDate)
    
    // Execute with dynamic adjustment
    foundDates := make(map[string]bool)
    processedCount := 0
    
    for _, date := range datesToCheck {
        // Check if file exists for this date
        if fileExists(date) {
            processedCount++
            foundDates[date] = true
        }
        
        // Recalculate remaining work based on findings
        actualRemaining := calculateActualExpected(remainingDates, foundDates)
        if actualRemaining != initialEstimate {
            // Adjust progress calculator
            fmt.Printf("[ADJUST] Updated expected files from %d to %d\n", 
                       initialEstimate, processedCount + actualRemaining)
            // Pipeline manager handles the adjustment automatically
        }
    }
}
```

## Error Handling

### Error Types

1. **Validation Errors** - Stage cannot run due to missing prerequisites
   ```go
   return pipeline.NewValidationError(stageID, "missing required file")
   ```

2. **Execution Errors** - Stage failed during execution
   ```go
   return pipeline.NewExecutionError(stageID, err, isRetryable)
   ```

3. **Cancellation Errors** - Stage was cancelled
   ```go
   if ctx.Err() == context.Canceled {
       return pipeline.NewCancellationError(stageID)
   }
   ```

### Error Recovery

```go
func (s *MyNewStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
    // Attempt with retry
    maxRetries := 3
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err := s.doWork(ctx)
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return pipeline.NewExecutionError(s.ID(), err, false)
        }
        
        if attempt < maxRetries {
            time.Sleep(time.Second * time.Duration(attempt))
            continue
        }
        
        return pipeline.NewExecutionError(s.ID(), err, true)
    }
    
    return nil
}
```

## Configuration and Context

### Setting Context Data

```go
// In one stage
state.SetContext("files_processed", 42)
state.SetContext("total_records", 1000)

// In a later stage
filesProcessed, _ := state.GetContext("files_processed")
totalRecords, _ := state.GetContext("total_records")
```

### Using Configuration

```go
// Get typed configuration
downloadDir, ok := state.GetConfig(pipeline.ContextKeyDownloadDir).(string)
if !ok {
    downloadDir = "data/downloads" // default
}

// Check mode
mode, _ := state.GetConfig(pipeline.ContextKeyMode).(string)
if mode == pipeline.ModeFull {
    // Full processing mode
}
```

## Best Practices

### 1. Always Validate First
- Check all prerequisites in Validate()
- Create directories if needed
- Verify executables exist

### 2. Handle Cancellation
```go
select {
case <-ctx.Done():
    return pipeline.NewCancellationError(s.ID())
default:
    // Continue processing
}
```

### 3. Update Progress Regularly
- For long-running operations, update progress frequently
- Include meaningful messages with progress updates
- Use metadata to provide additional context

### 4. Clean Error Messages
- Provide actionable error messages
- Include relevant context
- Specify if errors are retryable

### 5. Resource Cleanup
```go
func (s *MyNewStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
    // Setup
    resource, err := s.acquireResource()
    if err != nil {
        return err
    }
    defer resource.Close() // Always cleanup
    
    // Use resource
    return s.useResource(resource)
}
```

## Testing Stages

### Unit Testing

```go
func TestMyNewStage_Execute(t *testing.T) {
    // Create test state
    state := &pipeline.PipelineState{
        ID:     "test-pipeline",
        Stages: make(map[string]*pipeline.StageState),
        Context: map[string]interface{}{
            pipeline.ContextKeyReportDir: "testdata/reports",
        },
    }
    
    // Create stage
    stage := NewMyNewStage(".", testLogger)
    
    // Execute
    ctx := context.Background()
    err := stage.Execute(ctx, state)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, 100.0, state.GetStage(stage.ID()).Progress)
}
```

### Integration Testing

```go
func TestPipelineWithMyNewStage(t *testing.T) {
    // Create pipeline manager
    manager := pipeline.NewManager(mockHub, testLogger)
    
    // Register stages
    manager.RegisterStage(NewScrapingStage(".", testLogger))
    manager.RegisterStage(NewProcessingStage(".", testLogger))
    manager.RegisterStage(NewMyNewStage(".", testLogger))
    
    // Execute pipeline
    req := pipeline.PipelineRequest{
        ID: "test",
        Parameters: map[string]interface{}{
            // test parameters
        },
    }
    
    resp, err := manager.Execute(context.Background(), req)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, pipeline.PipelineStatusCompleted, resp.Status)
}
```

## Troubleshooting

### Common Issues

1. **Stage Not Executing**
   - Check dependencies are satisfied
   - Verify stage is registered
   - Check validation passes

2. **Progress Not Updating**
   - Ensure stdout is captured (for external commands)
   - Verify progress parser patterns match output
   - Check WebSocket connection is active

3. **Pipeline Hangs**
   - Add context timeout
   - Check for blocking operations
   - Verify commands exit properly

### Debug Mode

```go
// Enable debug logging
os.Setenv("PIPELINE_DEBUG", "true")

// In your stage
if os.Getenv("PIPELINE_DEBUG") == "true" {
    log.Printf("[DEBUG] Stage %s: Processing file %s", s.ID(), filename)
}
```

## Advanced Topics

### Conditional Execution

```go
func (s *MyNewStage) Validate(state *pipeline.PipelineState) error {
    // Skip stage based on condition
    if skipCondition(state) {
        return pipeline.ErrSkipStage
    }
    
    return s.BaseStage.Validate(state)
}
```

### Parallel Processing Within Stage

```go
func (s *MyNewStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
    files := s.getFilesToProcess()
    
    // Process files in parallel
    var wg sync.WaitGroup
    errors := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            if err := s.processFile(f); err != nil {
                errors <- err
            }
        }(file)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### Custom Stage Metadata

```go
// Set custom metadata
stageState.Metadata["custom_field"] = "custom_value"
stageState.Metadata["processing_stats"] = map[string]interface{}{
    "files_processed": 10,
    "records_created": 1000,
    "errors_encountered": 0,
}

// Access in WebSocket updates
// Frontend receives this in progress updates
```

## Summary

The Pipeline Manager provides a robust framework for orchestrating complex data processing pipelines. By following the patterns and practices in this guide, you can create reliable, maintainable pipeline stages that integrate seamlessly with the existing system.

Key benefits:
- Automatic dependency management
- Consistent error handling
- Real-time progress tracking
- Easy testing and debugging
- Clean, modular architecture

For more details, see:
- [Pipeline Architecture Specification](../specifications/PIPELINE_ARCHITECTURE.md)
- [Stage Interface Specification](../specifications/PIPELINE_STAGE_INTERFACE.md)
- [Architecture Overview](ARCHITECTURE.md)