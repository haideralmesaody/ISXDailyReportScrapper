# WebSocket Message Specifications

## Architectural Context
**IMPORTANT**: WebSocket communication is strictly for one-way status updates from backend to frontend. It must NEVER be used for control flow, decision making, or pipeline progression. See [Architecture Principles](../design/ARCHITECTURE_PRINCIPLES.md) for details.

## Overview
This document specifies all WebSocket message types used in the ISX Daily Reports Scrapper application for real-time status communication from the backend to the frontend.

## Message Structure

All WebSocket messages follow a consistent JSON structure:

```json
{
  "type": "message_type",
  "data": {
    // Type-specific data
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Stage Naming Conventions

### Standard Stage Names
All executables and the web application MUST use these standardized stage names for consistency:

| Executable | Stage Name | Description |
|------------|------------|-------------|
| scraper.exe | `scraping` | Downloading Excel reports from ISX |
| process.exe | `processing` | Processing Excel files into CSV |
| indexcsv.exe | `indices` | Extracting market indices |
| web-licensed.exe | `analysis` | Generating ticker summaries |

### Important Implementation Notes:
1. **Executable Stage Names**: Each executable must use its designated stage name when creating progress calculators:
   ```go
   // In scraper.go
   calc := progress.NewEnhancedCalculator("scraping", totalFiles, metricsManager)
   
   // In data-processor.go
   calc := progress.NewEnhancedCalculator("processing", totalFiles, metricsManager)
   
   // In index-extractor.go
   calc := progress.NewEnhancedCalculator("indices", len(files), metricsManager)
   ```

2. **Web Application Consistency**: The web application must pass the correct stage name when executing commands:
   ```go
   // In web-application.go
   scraperResponse := executeCommandWithStreaming(scraperPath, args, "scraping")
   processResponse := executeCommandWithStreaming(processPath, args, "processing")
   indexResponse := executeCommandWithStreaming(indexcsvPath, args, "indices")
   ```

3. **Process Exit Codes**: Pipeline progression is determined by process exit codes (0 = success), NOT by WebSocket messages. WebSocket messages are only for UI updates.

## Message Types

### 1. Connection Messages

#### connection
Sent when a client connects or reconnects to the WebSocket.

```json
{
  "type": "connection",
  "message": "connected",
  "data": {
    "valid": true,
    "days_remaining": 45,
    "expiry_date": "2024-03-01T00:00:00Z"
  }
}
```

### 2. Pipeline Status Messages

#### pipeline_status
Updates the status of a pipeline stage.

```json
{
  "type": "pipeline_status",
  "data": {
    "stage": "scraping|processing|indices|analysis",
    "status": "active|processing|completed|error",
    "message": "Downloading Excel reports (5 new, 10 existing)"
  }
}
```

#### pipeline_reset
Resets all pipeline stages to inactive state.

```json
{
  "type": "pipeline_reset",
  "data": {}
}
```

#### pipeline_complete
Sent when the entire pipeline finishes successfully.

```json
{
  "type": "pipeline_complete",
  "data": {
    "message": "Complete data pipeline finished! All data updated."
  }
}
```

### 3. Progress Messages

#### progress
Real-time progress updates during long operations.

```json
{
  "type": "progress",
  "data": {
    "stage": "scraping",
    "current": 15,
    "total": 50,
    "percentage": 30.0,
    "message": "Downloading Excel reports",
    "eta": "2 minutes remaining",
    "details": {
      "downloaded": 10,
      "existing": 5,
      "elapsed": "1m 30s"
    }
  }
}
```

### 4. Output Messages

#### output
General output messages from executables.

```json
{
  "type": "output",
  "message": "[DOWNLOAD] File 5/20: 2024 01 15 ISX Daily Report.xlsx",
  "level": "info|success|error|warning"
}
```

### 5. Data Update Messages

#### file_update
Sent when files are added, modified, or deleted.

```json
{
  "type": "file_update",
  "data": {
    "action": "created|modified|deleted",
    "filename": "ticker_summary.json",
    "path": "data/reports/ticker_summary.json",
    "component": "tickers"
  }
}
```

#### refresh
Instructs frontend to refresh specific components.

```json
{
  "type": "refresh",
  "data": {
    "source": "scrape|process|watcher",
    "components": ["files", "tickers", "charts"]
  }
}
```

### 6. Error Messages

#### error
Structured error information.

```json
{
  "type": "error",
  "data": {
    "code": "DOWNLOAD_ERROR",
    "message": "Failed to download file",
    "details": "Network timeout after 30 seconds",
    "stage": "scraping",
    "recoverable": true,
    "hint": "Check network connection or try again"
  }
}
```

## Frontend Message Handling

### Message Router
The frontend uses a central message router to handle different message types:

```javascript
switch (message.type) {
    case 'connection':
        handleConnectionMessage(message);
        break;
    case 'pipeline_status':
        handlePipelineStatus(message.data);
        break;
    case 'progress':
        handleProgressUpdate(message.data);
        break;
    case 'output':
        handleOutputMessage(message);
        break;
    case 'file_update':
        handleFileUpdate(message.data);
        break;
    case 'refresh':
        handleRefreshRequest(message.data);
        break;
    case 'error':
        handleErrorMessage(message.data);
        break;
}
```

### Component Updates
Different components subscribe to relevant message types:

- **Pipeline UI**: Listens for pipeline_status, pipeline_reset, pipeline_complete
- **Progress Bar**: Listens for progress messages
- **Console Output**: Listens for output and error messages
- **Data Tables**: Listens for file_update and refresh messages

## Backend Message Broadcasting

### Hub Methods

```go
// Send general output message
wsHub.BroadcastOutput(message, level)

// Send pipeline status update
wsHub.BroadcastUpdate("pipeline_status", "", "", data)

// Send progress update
wsHub.BroadcastProgressWithDetails(stage, current, total, percentage, message, eta, details)

// Send error message
wsHub.BroadcastError(code, message, details, stage, recoverable)

// Send refresh request
wsHub.BroadcastRefresh(source, components)

// Send file update notification
wsHub.BroadcastFileUpdate(action, filename, path, component)
```

## Progress Calculation Utilities

### Overview (v0.3.0-alpha)
The `internal/progress` package provides standardized utilities for creating consistent WebSocket messages and tracking progress with historical accuracy. This system is now in production and significantly improves user experience with accurate ETAs.

#### Components:
1. **Calculator**: Basic progress tracking with real-time ETA calculation
2. **EnhancedCalculator**: Extended calculator using historical metrics for improved ETA accuracy
3. **MetricsManager**: Persists timing data for future predictions and immediate ETA estimates
4. **Message Helpers**: Functions for creating properly formatted status, progress, and error messages
5. **Dynamic Adjustment**: Updates progress based on actual findings (non-trading days, existing files)

### Usage Examples

#### Basic Progress Tracking
```go
import "isxcli/internal/progress"

// Create a basic calculator
calc := progress.NewCalculator("scraping", totalFiles)

// Update progress
calc.Update(processedCount)

// Get formatted message
jsonData, _ := calc.ToJSON("Downloading files...", details)
fmt.Printf("[WEBSOCKET_PROGRESS] %s\n", jsonData)
```

#### Enhanced Progress with Historical Data (Production)
```go
// Initialize metrics manager (now stores data in data/metrics/)
dataPath := filepath.Dir(outDir) // Parent of downloads directory
metricsManager := progress.NewMetricsManager(dataPath)

// Create enhanced calculator with historical metrics support
calc := progress.NewEnhancedCalculator("scraping", expectedFiles, metricsManager)

// Update progress during execution
calc.Update(processedCount)

// Send progress message with detailed metadata
sendProgress(calc, "Downloading Excel reports", map[string]interface{}{
    "downloaded":     newDownloads,
    "existing":       existingFiles,
    "current_page":   pageNumber,
    "elapsed":        time.Since(calc.StartTime).String(),
    "expected_remaining": calc.TotalItems - calc.ProcessedItems,
})

// Dynamic adjustment based on findings
if nonTradingDayDiscovered {
    // Reduce expected total when market was closed
    calc.TotalItems = calc.ProcessedItems + recalculatedRemaining
}

// Save metrics on completion for future ETA predictions
calc.Complete()
```

#### Dynamic Progress Adjustment (New Feature)
```go
// Real-time adjustment example from scraper.go
func adjustProgressForActualFindings(calc *progress.EnhancedCalculator, foundDates map[string]bool, fromDate, toDate time.Time) {
    // Recalculate expected files based on actual findings
    actualExpected := calculateActualExpectedFiles(fromDate, toDate, foundDates)
    
    if actualExpected != calc.TotalItems && actualExpected > 0 {
        fmt.Printf("[ADJUST] Updated expected remaining files from %d to %d for better ETA\n", 
                   calc.TotalItems - calc.ProcessedItems, actualExpected)
        calc.TotalItems = calc.ProcessedItems + actualExpected
    }
}
```

#### Standardized Message Creation
```go
// Status message
jsonData, _ := progress.StatusToJSON("scraping", "completed", "All files downloaded")
fmt.Printf("[WEBSOCKET_STATUS] %s\n", jsonData)

// Error message
jsonData, _ := progress.ErrorToJSON(
    "DOWNLOAD_ERROR",                    // code
    "Failed to download file",           // message
    err.Error(),                         // details
    "scraping",                          // stage
    true,                                // recoverable
    "Check network connection",          // hint
)
fmt.Printf("[WEBSOCKET_ERROR] %s\n", jsonData)
```

### Metrics Storage (Production Implementation)
Historical metrics are now actively persisted in JSON format within the `data/metrics/` directory:

```
data/metrics/
├── scraping_metrics.json      # Download timing history
├── processing_metrics.json    # Processing timing history
├── indices_metrics.json       # Index extraction timing history
└── analysis_metrics.json      # Analysis timing history (added v0.3.0)
```

Each metrics file contains detailed timing data:
```json
{
  "stage": "scraping",
  "history": [
    {
      "timestamp": "2025-01-20T10:30:00Z",
      "total_items": 15,
      "total_duration_seconds": 225.3,
      "avg_per_item_seconds": 15.02,
      "metadata": {
        "mode": "initial",
        "date_range": "2025-01-01 to 2025-01-15"
      }
    },
    {
      "timestamp": "2025-01-21T14:15:00Z", 
      "total_items": 5,
      "total_duration_seconds": 72.1,
      "avg_per_item_seconds": 14.42,
      "metadata": {
        "mode": "accumulative",
        "date_range": "2025-01-20 to 2025-01-21"
      }
    }
  ],
  "average_times": {
    "per_item": 14.72,
    "recent_trend": "improving"
  },
  "last_updated": "2025-01-21T14:15:00Z"
}
```

### Proven Benefits in Production
1. **Immediate ETA**: Shows estimated time even before first file is downloaded
2. **Learning System**: ETA accuracy improves significantly after 2-3 runs
3. **Consistency**: All four pipeline stages use identical message format
4. **Fallback Safety**: Works perfectly without historical data
5. **Dynamic Adjustment**: Accounts for non-trading days and existing files
6. **User Experience**: Users see realistic time estimates instead of "calculating..."

### Executable Communication

Executables communicate with the web application using special stdout prefixes:

```go
// Status update
fmt.Printf("[WEBSOCKET_STATUS] %s\n", jsonString)

// Progress update  
fmt.Printf("[WEBSOCKET_PROGRESS] %s\n", jsonString)

// Error message
fmt.Printf("[WEBSOCKET_ERROR] %s\n", jsonString)

// Important: Flush stdout before exiting to ensure all messages are sent
os.Stdout.Sync()
```

**Note**: All executables MUST call `os.Stdout.Sync()` before exiting to ensure all WebSocket messages are properly flushed and received by the web application. This prevents message loss when executables complete their work.

### Status Update Best Practices

1. **Send Status Updates for UI Display**: Send progress and status messages to keep the user informed:
   ```go
   // Progress updates during processing
   sendProgress(calc, "Processing files...", details)
   
   // Status updates at stage boundaries
   sendStatus("scraping", "completed", "Download finished")
   
   // Always flush before exit
   os.Stdout.Sync()
   ```

2. **Pipeline Progression**: The web application determines stage completion by:
   - **Process Exit Code**: 0 = success, continue to next stage
   - **Process Exit Code**: non-0 = failure, stop pipeline
   - WebSocket messages are ONLY for updating the UI display

3. **Dynamic Progress Tracking**: Executables can adjust progress dynamically for better UX:
   - Adjust expected counts based on actual findings
   - Provide accurate ETAs using historical data
   - This is purely for display and does not affect pipeline control

## Message Flow Examples

### 1. Complete Pipeline Execution

```
1. User clicks "Download Fresh Data"
   → pipeline_reset
   
2. Scraper starts
   → pipeline_status {stage: "scraping", status: "active"}
   → progress {stage: "scraping", current: 0, total: 20}
   → output "[DOWNLOAD] File 1/20..."
   → progress {stage: "scraping", current: 1, total: 20}
   ...
   → pipeline_status {stage: "scraping", status: "completed"}
   
3. Processor starts
   → pipeline_status {stage: "processing", status: "active"}
   → output "Processing Excel files..."
   → pipeline_status {stage: "processing", status: "completed"}
   
4. Index extractor starts
   → pipeline_status {stage: "indices", status: "active"}
   → output "Extracting market indices..."
   → pipeline_status {stage: "indices", status: "completed"}
   
5. Analysis starts
   → pipeline_status {stage: "analysis", status: "active"}
   → output "Generating ticker summary..."
   → pipeline_status {stage: "analysis", status: "completed"}
   
6. Pipeline complete
   → pipeline_complete
   → refresh {components: ["files", "tickers", "charts"]}
```

### 2. File Watcher Update

```
1. File modified detected
   → file_update {action: "modified", filename: "ticker_summary.json"}
   
2. Components refresh
   → refresh {source: "watcher", components: ["tickers"]}
```

### 3. Error Handling

```
1. Download fails
   → error {
       code: "DOWNLOAD_ERROR",
       stage: "scraping",
       recoverable: true
     }
   → pipeline_status {stage: "scraping", status: "error"}
```

## Best Practices

### DO:
1. **Always include timestamps** for debugging and sequencing
2. **Use structured data** instead of plain strings
3. **Include context** (stage, source, component) for targeted updates
4. **Make errors actionable** with hints and recovery options
5. **Send status updates** at stage boundaries for UI clarity
6. **Test reconnection** scenarios to ensure state consistency

### DON'T:
1. **Use WebSocket for control flow** - Pipeline decisions must be based on exit codes
2. **Wait for WebSocket confirmation** - Backend proceeds based on process completion
3. **Parse WebSocket messages for logic** - They are display-only
4. **Block on WebSocket delivery** - Continue pipeline regardless of UI updates
5. **Mix concerns** - Keep pipeline logic separate from status broadcasting

### Common Anti-Patterns to Avoid:
```go
// WRONG: Using WebSocket to control pipeline
if websocketMessage.Status == "completed" {
    startNextStage() // Never do this!
}

// WRONG: Waiting for WebSocket confirmation
sendWebSocketMessage(status)
waitForWebSocketAck() // This creates race conditions

// RIGHT: Use exit codes for control
if cmd.Wait() == nil { // Exit code 0
    broadcastStatus("completed") // For UI only
    startNextStage() // Direct control flow
}
```

## Testing WebSocket Messages

### Manual Testing
1. Open browser developer console
2. Monitor Network → WS tab
3. Verify message format and sequence
4. Test reconnection by refreshing page

### Automated Testing
```javascript
// Mock WebSocket for testing
const mockWS = {
    send: jest.fn(),
    close: jest.fn(),
    readyState: WebSocket.OPEN
};

// Test message handling
const testMessage = {
    type: 'pipeline_status',
    data: {
        stage: 'scraping',
        status: 'completed',
        message: 'Test completed'
    }
};

handleWebSocketMessage(testMessage);
expect(document.getElementById('stage-scraping').classList.contains('completed')).toBe(true);
```

## Code References

- WebSocket hub: `dev/internal/websocket/hub.go`
- Message types: `dev/internal/websocket/types.go`
- Frontend handler: `dev/web/index.html:1850-2000`
- Backend broadcasting: `dev/cmd/web-licensed/web-application.go:900-970`