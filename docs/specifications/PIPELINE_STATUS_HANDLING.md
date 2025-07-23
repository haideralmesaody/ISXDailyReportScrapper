# Pipeline Status Handling Specification

## Overview
This document specifies how pipeline status transitions are handled in the ISX Daily Reports Scrapper application. It ensures that all status updates are properly tracked and displayed in the UI.

## Pipeline Stages

The data collection pipeline consists of four sequential stages:

1. **Scraping** - Downloading Excel files from ISX website
2. **Processing** - Converting Excel files to CSV format
3. **Indices** - Extracting market indices (ISX60, ISX15)
4. **Analysis** - Generating ticker summaries and statistics

## Status Types

Each stage can have one of the following statuses:

- `active` - Stage is currently running
- `processing` - Stage is actively processing data (alternative to active)
- `completed` - Stage has finished successfully
- `error` - Stage encountered an error

## Implementation Details

### WebSocket Message Format

Pipeline status updates are sent via WebSocket using the following format:

```json
{
  "type": "pipeline_status",
  "data": {
    "stage": "scraping|processing|indices|analysis",
    "status": "active|processing|completed|error",
    "message": "Human-readable status message"
  }
}
```

### Backend Implementation

#### 1. Executable Status Messages

Each executable sends structured status messages to stdout:

```go
// Format: [WEBSOCKET_STATUS] {json}
sendStatus("scraping", "completed", "Scraping completed: 10 new files")
```

#### 2. Web Application Handler

The web application captures these messages and broadcasts them:

```go
// In executeCommandWithStreaming
if strings.HasPrefix(line, "[WEBSOCKET_STATUS] ") {
    jsonStr := strings.TrimPrefix(line, "[WEBSOCKET_STATUS] ")
    var msg map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &msg); err == nil {
        if data, ok := msg["data"].(map[string]interface{}); ok {
            stage, _ := data["stage"].(string)
            status, _ := data["status"].(string)
            message, _ := data["message"].(string)
            
            // Send structured pipeline status
            wsHub.BroadcastUpdate("pipeline_status", "", "", map[string]interface{}{
                "stage": stage,
                "status": status,
                "message": message,
            })
        }
    }
}
```

#### 3. Additional Status Triggers

The `sendPipelineStatus` function also monitors broadcast messages for specific patterns:

- **Scraping Stage**:
  - Active: "Starting ISX data download", "Downloading fresh data"
  - Completed: "Fresh data downloaded successfully", "Download phase completed"

- **Processing Stage**:
  - Active: "Starting automatic data processing", "Download phase completed"
  - Completed: "Data processing completed"

- **Indices Stage**:
  - Active: "Extracting market indices"
  - Completed: "Index extraction completed"

- **Analysis Stage**:
  - Active: "Generating ticker summary"
  - Completed: "Complete data pipeline finished", "Complete processing pipeline finished"

### Frontend Implementation

#### 1. HTML Structure

Each stage has a corresponding HTML element:

```html
<div class="pipeline-stage" id="stage-scraping">
<div class="pipeline-stage" id="stage-processing">
<div class="pipeline-stage" id="stage-indices">
<div class="pipeline-stage" id="stage-analysis">
```

#### 2. JavaScript Handler

The frontend listens for pipeline_status messages and updates the UI:

```javascript
if (message.type === 'pipeline_status') {
    const data = message.data;
    updatePipelineStage(`stage-${data.stage}`, data.status);
    
    // Update scrape status message
    if (data.message) {
        const statusType = data.status === 'completed' ? 'success' :
                         data.status === 'error' ? 'error' : 'info';
        updateScrapeStatus(data.message, statusType);
    }
}
```

#### 3. CSS Classes

Pipeline stages use CSS classes to indicate their status:

- `.pipeline-stage` - Default state (inactive/gray)
- `.pipeline-stage.active` - Currently running (green glow)
- `.pipeline-stage.processing` - Processing data (green glow with animation)
- `.pipeline-stage.completed` - Successfully completed (solid green)

## Status Flow

### Complete Pipeline Execution

1. User clicks "Download Fresh Data"
2. Pipeline resets all stages to inactive
3. **Scraping Stage**:
   - Status: active (when scraper.exe starts)
   - Status: completed (when downloads finish)
4. **Processing Stage**:
   - Status: active (when process.exe starts)
   - Status: completed (when CSV generation finishes)
5. **Indices Stage**:
   - Status: active (when indexcsv.exe starts)
   - Status: completed (when indices extraction finishes)
6. **Analysis Stage**:
   - Status: active (when generateTickerSummary starts)
   - Status: completed (when summary generation finishes)
7. Pipeline complete message sent

### Error Handling

If any stage encounters an error:
1. The stage status is set to "error"
2. An error message is displayed
3. Subsequent stages are not executed
4. User can retry the operation

## Testing

To verify pipeline status handling:

1. **Normal Flow**: Run a complete pipeline and verify all stages transition properly
2. **Error Cases**: Simulate errors (missing files, network issues) and verify error status
3. **Partial Execution**: Run individual stages and verify only relevant stages update
4. **WebSocket Connection**: Verify status updates work with WebSocket reconnection

## Common Issues and Solutions

### Issue: Stage status not updating
**Solution**: Ensure the executable is using `executeCommandWithStreaming` instead of `executeCommand`

### Issue: Status stuck on active
**Solution**: Verify the executable sends a completion status message via `sendStatus(stage, "completed", message)`

### Issue: Wrong stage highlighted
**Solution**: Check that stage names match exactly: "scraping", "processing", "indices", "analysis"

## Code References

- Status sending: `dev/scraper.go:394`, `dev/cmd/process/data-processor.go:337`
- WebSocket handling: `dev/cmd/web-licensed/web-application.go:1532-1570`
- Frontend updates: `dev/web/index.html:1955-1965`
- Pipeline UI: `dev/web/index.html:1252-1295`