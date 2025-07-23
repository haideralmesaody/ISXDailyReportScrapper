# Pipeline Architecture Specification

## Overview
This document specifies the architecture of the data processing pipeline in the ISX Daily Reports Scrapper. It defines how stages are executed, how transitions occur, and how to properly extend the pipeline.

## Recent Updates (v0.3.0-alpha)
As of January 2025, the pipeline architecture has been significantly improved with the implementation of PIPE-002:
- **Pipeline Manager**: Centralized orchestration system (`internal/pipeline`)
- **Stage-based Architecture**: Modular, reusable pipeline stages with dependency management
- **Automatic Progression**: Stages chain automatically based on exit codes
- **Fixed Critical Bug**: Pipeline no longer stops after scraping stage
- **Enhanced Progress Tracking**: Historical metrics improve ETA accuracy over time
- **Simplified Implementation**: Reduced handleScrape from 200+ lines to ~50 lines

## Pipeline Stages

### Current Pipeline Stages
1. **Scraping** (`scraper.exe`)
   - Downloads Excel reports from ISX website
   - Inputs: Date range, output directory
   - Outputs: Excel files in `data/downloads/`
   - Success: Exit code 0

2. **Processing** (`process.exe`)
   - Converts Excel files to CSV format
   - Combines data into unified dataset
   - Inputs: Download directory path
   - Outputs: CSV files in `data/reports/`
   - Success: Exit code 0

3. **Indices** (`indexcsv.exe`)
   - Extracts market indices (ISX60, ISX15)
   - Inputs: Excel directory, output path
   - Outputs: `indexes.csv`
   - Success: Exit code 0

4. **Analysis** (internal function)
   - Generates ticker summaries
   - Calculates statistics
   - Inputs: Combined CSV data
   - Outputs: `ticker_summary.json`
   - Success: Function returns nil error

## Pipeline Control Flow

### Execution Sequence
```
handleScrape() 
    ↓
Execute scraper.exe → Wait for exit code
    ↓ (success)
Execute process.exe → Wait for exit code
    ↓ (success)
Execute indexcsv.exe → Wait for exit code
    ↓ (success)
Execute generateTickerSummary()
    ↓ (success)
Send pipeline_complete message
```

### Implementation Pattern
```go
func handleScrape(w http.ResponseWriter, r *http.Request) {
    // 1. Parse and validate user input
    var req CommandRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Execute scraper
    scraperArgs := buildScraperArgs(req)
    scraperCmd := exec.Command("scraper.exe", scraperArgs...)
    
    // Send status update
    broadcastStatus("scraping", "active", "Starting download...")
    
    // Wait for completion
    if err := scraperCmd.Run(); err != nil {
        broadcastStatus("scraping", "error", "Scraping failed")
        respondWithError(w, err)
        return
    }
    
    broadcastStatus("scraping", "completed", "Download complete")
    
    // 3. Execute processor
    broadcastStatus("processing", "active", "Processing files...")
    processCmd := exec.Command("process.exe", "-in=data/downloads")
    
    if err := processCmd.Run(); err != nil {
        broadcastStatus("processing", "error", "Processing failed")
        respondWithError(w, err)
        return
    }
    
    broadcastStatus("processing", "completed", "Processing complete")
    
    // 4. Continue with remaining stages...
}
```

## Stage Implementation Requirements

### Executable Requirements
Each pipeline executable must:

1. **Accept command-line arguments**
2. **Send progress via stdout** (optional but recommended)
3. **Return proper exit codes**:
   - 0 = Success
   - 1 = General failure
   - 2+ = Specific error codes (document these)
4. **Complete all work before exiting**

### Status Communication
```go
// Progress updates during execution (from executable)
func sendProgress(current, total int, message string) {
    calc := progress.NewCalculator("stagename", total)
    calc.Update(current)
    jsonData, _ := calc.ToJSON(message, nil)
    fmt.Printf("[WEBSOCKET_PROGRESS] %s\n", jsonData)
}

// Status updates (from web application)
func broadcastStatus(stage, status, message string) {
    wsHub.BroadcastUpdate("pipeline_status", "", "", map[string]interface{}{
        "stage": stage,
        "status": status,
        "message": message,
    })
}
```

## Adding New Pipeline Stages

### Step 1: Create the Executable
```go
// cmd/newstage/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "isxcli/internal/progress"
)

func main() {
    // Parse arguments
    input := flag.String("in", "", "Input directory")
    output := flag.String("out", "", "Output file")
    flag.Parse()
    
    // Validate
    if *input == "" || *output == "" {
        fmt.Println("Error: Missing required arguments")
        os.Exit(1)
    }
    
    // Initialize progress
    calc := progress.NewCalculator("newstage", 100)
    
    // Do work
    if err := processData(*input, *output, calc); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    
    // Success
    os.Exit(0)
}
```

### Step 2: Add to Pipeline
```go
// In web-application.go - handleScrape function
// After existing stages...

// New stage
broadcastStatus("newstage", "active", "Running new analysis...")
newStageCmd := exec.Command("newstage.exe", "-in=data/reports", "-out=data/reports/newdata.json")

if err := newStageCmd.Run(); err != nil {
    broadcastStatus("newstage", "error", "New stage failed")
    // Decide: Is this fatal or can we continue?
    respondWithError(w, err)
    return
}

broadcastStatus("newstage", "completed", "New analysis complete")
```

### Step 3: Update Frontend Display
```javascript
// In index.html - Add UI element
<div id="stage-newstage" class="pipeline-stage">
    <h4>New Stage</h4>
    <div class="status-indicator"></div>
    <div class="status-message"></div>
</div>

// In pipeline status handler
case 'pipeline_status':
    updatePipelineStage(data.stage, data.status, data.message);
    break;
```

## Error Handling Strategy

### Failure Modes
1. **Fatal Errors**: Stop pipeline immediately
   - Missing required files
   - Invalid license
   - Disk full

2. **Recoverable Errors**: Log and continue
   - Some files failed to download
   - Non-critical analysis failed

3. **Warnings**: Note but don't stop
   - Older data than expected
   - Missing optional features

### Implementation
```go
func runStageWithRecovery(stage string, cmd *exec.Cmd, critical bool) error {
    broadcastStatus(stage, "active", fmt.Sprintf("Starting %s...", stage))
    
    err := cmd.Run()
    
    if err != nil {
        if critical {
            broadcastStatus(stage, "error", fmt.Sprintf("%s failed", stage))
            return err // Stop pipeline
        } else {
            broadcastStatus(stage, "warning", fmt.Sprintf("%s had issues but continuing", stage))
            // Log error but continue
        }
    } else {
        broadcastStatus(stage, "completed", fmt.Sprintf("%s completed", stage))
    }
    
    return nil
}
```

## Pipeline Orchestration Patterns

### Sequential Execution (Current)
```go
// Each stage waits for previous to complete
runScraper() → runProcessor() → runIndices() → runAnalysis()
```

### Parallel Execution (Future)
```go
// Some stages can run in parallel
runScraper() → ├─ runProcessor() → runAnalysis()
                └─ runIndices()
```

### Conditional Execution
```go
// Skip stages based on conditions
if mode == "full" {
    runScraper()
}
runProcessor() // Always run

if hasNewData() {
    runAnalysis()
}
```

## WebSocket Message Flow

### During Pipeline Execution
```
START
  ↓
Backend: broadcastStatus("scraping", "active", "Starting download...")
  ↓
Frontend: Updates UI to show scraping active
  ↓
Scraper: Sends progress messages via stdout
  ↓
Backend: Forwards progress to frontend
  ↓
Scraper: Exits with code 0
  ↓
Backend: broadcastStatus("scraping", "completed", "Download complete")
  ↓
Backend: broadcastStatus("processing", "active", "Processing files...")
  ↓
... (continues for each stage)
  ↓
Backend: broadcastUpdate("pipeline_complete", ...)
  ↓
END
```

## Best Practices

### DO:
- Check exit codes to determine success
- Send clear status messages at stage boundaries
- Handle errors gracefully
- Log detailed information for debugging
- Keep stages independent when possible

### DON'T:
- Use WebSocket messages for control flow
- Wait for frontend confirmation
- Mix stage logic across files
- Assume previous stages succeeded without checking
- Leave stages in ambiguous states

## Testing Pipeline Changes

### Unit Testing (Stage Level)
```go
func TestNewStage(t *testing.T) {
    // Setup test data
    setupTestFiles()
    
    // Run stage
    cmd := exec.Command("newstage.exe", "-in=testdata", "-out=testout.json")
    err := cmd.Run()
    
    // Verify
    assert.NoError(t, err)
    assert.FileExists(t, "testout.json")
}
```

### Integration Testing (Pipeline Level)
```go
func TestFullPipeline(t *testing.T) {
    // Run entire pipeline
    response := simulateHTTPRequest("/api/scrape", testDates)
    
    // Wait for completion
    waitForPipelineComplete()
    
    // Verify all outputs exist
    assert.FileExists(t, "data/reports/isx_combined_data.csv")
    assert.FileExists(t, "data/reports/indexes.csv")
    assert.FileExists(t, "data/reports/ticker_summary.json")
}
```

## Monitoring and Debugging

### Key Log Points
1. Stage start/end with timestamps
2. Exit codes from each executable
3. File counts and sizes
4. Error messages with context
5. Performance metrics

### Debug Mode
```go
if os.Getenv("PIPELINE_DEBUG") == "true" {
    log.Printf("[PIPELINE] Stage %s starting with args: %v", stage, args)
    log.Printf("[PIPELINE] Stage %s completed in %v", stage, duration)
}
```

## Pipeline Manager Implementation (v0.3.0-alpha)

### Overview
The Pipeline Manager (`internal/pipeline`) provides a robust, extensible framework for pipeline orchestration, completely replacing the previous conditional logic approach.

### Current Implementation Status
✅ **IMPLEMENTED and DEPLOYED**: The Pipeline Manager is now the active pipeline execution system in production.

### Core Components

#### 1. Pipeline Manager
```go
type Manager struct {
    registry    *Registry           // Stage registry with dependency management
    config      *Config            // Pipeline configuration
    hub         WebSocketHub       // WebSocket hub for status updates
    logger      Logger            // Logging interface
    mu          sync.RWMutex      // Thread-safe access to pipelines
    pipelines   map[string]*PipelineState // Active pipeline tracking
}
```

#### 2. Pipeline State Management
```go
type PipelineState struct {
    ID          string                    // Unique pipeline identifier
    Status      PipelineStatus           // Overall pipeline status
    Stages      map[string]*StageState   // Individual stage states
    Context     map[string]interface{}   // Pipeline context and parameters
    StartTime   time.Time               // Pipeline start timestamp
    EndTime     time.Time              // Pipeline completion timestamp
    Error       error                  // Pipeline error if failed
}
```

#### 3. Stage Registry and Dependencies
```go
// Registry manages stage registration and dependency resolution
type Registry struct {
    stages map[string]Stage
    deps   map[string][]string
}

// Automatic dependency resolution ensures correct execution order
func (r *Registry) GetDependencyOrder() ([]Stage, error) {
    // Returns stages in proper execution order based on dependencies
}
```

#### 4. Enhanced Progress Integration
The Pipeline Manager integrates with the new `internal/progress` package:
```go
// Each stage gets its own progress calculator with historical metrics
calc := progress.NewEnhancedCalculator(stageName, totalItems, metricsManager)

// Progress automatically sent via WebSocket during execution
func (s *Stage) Execute(ctx context.Context, state *PipelineState) error {
    calc.Update(processedItems)
    // Pipeline manager automatically broadcasts progress updates
}
```

### Current Stage Implementation
All stages are implemented in `cmd/web-licensed/stages/`:

1. **ScrapingStage** (`scraping_stage.go`)
   - Wraps `scraper.exe` execution
   - Handles progress parsing from stdout
   - Uses historical metrics for ETA prediction

2. **ProcessingStage** (`processing_stage.go`)
   - Executes `process.exe` with proper arguments
   - Monitors file processing progress
   - Dependencies: [scraping]

3. **IndicesStage** (`indices_stage.go`)
   - Runs `indexcsv.exe` for index extraction
   - Dependencies: [processing]

4. **AnalysisStage** (`analysis_stage.go`)
   - Generates ticker summaries directly in Go
   - Dependencies: [processing]
   - No external executable required

### WebSocket Adapter Integration
```go
// WebSocket adapter bridges pipeline events to frontend
type WebSocketAdapter struct {
    hub WebSocketHub
}

func (w *WebSocketAdapter) BroadcastUpdate(eventType, path, filename string, data map[string]interface{}) {
    // Converts pipeline events to WebSocket messages
    // Maintains compatibility with existing frontend
}
```

### Current handleScrape Implementation
The production implementation in `web-application.go`:

```go
func handleScrape(w http.ResponseWriter, r *http.Request) {
    // Parse and validate request
    var req CommandRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendCommandResponse(w, CommandResponse{Success: false, Error: err.Error()})
        return
    }

    // Create pipeline request with user parameters
    pipelineReq := pipeline.PipelineRequest{
        ID:       fmt.Sprintf("pipeline-%d", time.Now().Unix()),
        Mode:     req.Args["mode"],
        FromDate: req.Args["from"],
        ToDate:   req.Args["to"],
        Parameters: map[string]interface{}{
            "executableDir": executableDir,
        },
    }

    // Execute pipeline asynchronously
    go func() {
        ctx := context.Background()
        resp, err := pipelineManager.Execute(ctx, pipelineReq)
        if err != nil {
            logger.Error("Pipeline execution failed: %v", err)
            return
        }
        logger.Info("Pipeline %s completed successfully", resp.ID)
    }()

    // Return immediate response to client
    sendCommandResponse(w, CommandResponse{
        Success: true,
        Output:  fmt.Sprintf("Pipeline %s started", pipelineReq.ID),
    })
}
```

### Stage Validation and Error Handling
```go
// Each stage implements validation before execution
func (s *ScrapingStage) Validate(state *PipelineState) error {
    fromDate, exists := state.GetConfig(pipeline.ContextKeyFromDate)
    if !exists || fromDate == "" {
        return fmt.Errorf("from date is required for scraping")
    }
    return nil
}

// Manager handles retries and error recovery
func (m *Manager) executeStage(ctx context.Context, state *PipelineState, stage Stage) error {
    retryConfig := m.config.RetryConfig
    for attempt := 1; attempt <= retryConfig.MaxAttempts; attempt++ {
        err := stage.Execute(ctx, state)
        if err == nil || !IsRetryable(err) {
            return err
        }
        // Wait and retry for transient errors
    }
}
```

### Proven Benefits in Production
1. **Fixed Critical Bug**: Pipeline now correctly executes all stages (was stopping after scraping)
2. **No Variable Propagation Issues**: Centralized state management eliminates race conditions
3. **Automatic Stage Chaining**: Dependencies ensure correct execution order
4. **Clean Error Boundaries**: Each stage failure is properly isolated and handled
5. **Consistent Progress Tracking**: All stages use standardized progress reporting
6. **Easy Extension**: Adding new stages requires only implementing the Stage interface
7. **Maintainable Code**: Reduced handleScrape complexity from 200+ lines to ~50 lines

## Future Enhancements

### Planned Improvements
1. **Parallel Stage Execution**: Run independent stages simultaneously
2. **Stage Retry Logic**: Automatic retry for transient failures
3. **Checkpoint/Resume**: Save progress and resume from failure point
4. **Dynamic Pipeline**: Configure stages via configuration file
5. **Stage Metrics**: Track performance over time

### Extension Points
- Plugin architecture for custom stages
- Webhook notifications for stage events
- External stage execution (remote processing)
- Conditional stage execution based on data

## Summary
The pipeline architecture ensures:
- Clear separation of concerns
- Reliable execution flow
- Easy addition of new stages
- Proper error handling
- Consistent status communication

All pipeline control remains in the backend Go code, with WebSocket used only for status updates to the frontend.