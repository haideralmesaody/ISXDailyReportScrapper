# WebSocket Status Protocol Specification

## Overview

This document defines the standardized WebSocket message protocol for status updates between backend processes and the frontend UI, particularly for the processing pipeline status indicators.

## Message Format

All WebSocket messages follow this structure:

```json
{
    "type": "string",      // Message type: "status", "progress", "output", "data_update"
    "subtype": "string",   // Optional: More specific categorization
    "data": {              // Message payload
        // Type-specific fields
    },
    "timestamp": "string"  // ISO 8601 timestamp
}
```

## Pipeline Status Messages

### 1. Pipeline Stage Status

Used to update the visual pipeline stages (Scraping → Processing → Indices → Analysis).

```json
{
    "type": "pipeline_status",
    "data": {
        "stage": "scraping|processing|indices|analysis",
        "status": "inactive|active|processing|completed|error",
        "message": "Human-readable status message"
    }
}
```

**Stage Values:**
- `scraping` - Downloading Excel files from ISX
- `processing` - Converting Excel to CSV
- `indices` - Extracting market indices
- `analysis` - Generating summaries and analytics

**Status Values:**
- `inactive` - Stage not started (default)
- `active` - Stage is starting
- `processing` - Stage is actively working
- `completed` - Stage finished successfully
- `error` - Stage encountered an error

### 2. Progress Updates

Detailed progress within a stage.

```json
{
    "type": "progress",
    "data": {
        "stage": "scraping|processing|indices|analysis",
        "current": 15,
        "total": 50,
        "percentage": 30,
        "message": "Processing file 15 of 50",
        "eta": "3 minutes remaining",
        "details": {
            // Stage-specific details
        }
    }
}
```

### 3. Output Messages

Console output that should be displayed to the user.

```json
{
    "type": "output",
    "data": {
        "message": "Downloaded ISX report for 2024-01-15",
        "level": "info|success|warning|error",
        "stage": "scraping"  // Optional: Associate with pipeline stage
    }
}
```

## Stage-Specific Messages

### Scraping Stage

```json
// Starting
{
    "type": "pipeline_status",
    "data": {
        "stage": "scraping",
        "status": "active",
        "message": "Starting ISX data download..."
    }
}

// Progress
{
    "type": "progress",
    "data": {
        "stage": "scraping",
        "current": 15,
        "total": 45,
        "percentage": 33,
        "message": "Downloading Excel reports",
        "eta": "2 minutes remaining",
        "details": {
            "downloaded": 15,
            "existing": 30,
            "page": 3
        }
    }
}

// Completed
{
    "type": "pipeline_status",
    "data": {
        "stage": "scraping",
        "status": "completed",
        "message": "Downloaded 15 new reports"
    }
}
```

### Processing Stage

```json
// Starting
{
    "type": "pipeline_status",
    "data": {
        "stage": "processing",
        "status": "active",
        "message": "Starting data processing..."
    }
}

// Progress
{
    "type": "progress",
    "data": {
        "stage": "processing",
        "current": 10,
        "total": 15,
        "percentage": 67,
        "message": "Processing Excel files",
        "details": {
            "currentFile": "2024-01-15.xlsx",
            "recordsProcessed": 1250,
            "filesGenerated": 82
        }
    }
}

// Completed
{
    "type": "pipeline_status",
    "data": {
        "stage": "processing",
        "status": "completed",
        "message": "Processed 15 files, generated 82 ticker reports"
    }
}
```

### Indices Stage

```json
// Starting
{
    "type": "pipeline_status",
    "data": {
        "stage": "indices",
        "status": "active",
        "message": "Extracting market indices..."
    }
}

// Progress
{
    "type": "progress",
    "data": {
        "stage": "indices",
        "current": 5,
        "total": 15,
        "percentage": 33,
        "message": "Extracting ISX60 and ISX15 values",
        "details": {
            "filesProcessed": 5,
            "indicesFound": 10
        }
    }
}

// Completed
{
    "type": "pipeline_status",
    "data": {
        "stage": "indices",
        "status": "completed",
        "message": "Extracted indices for 15 trading days"
    }
}
```

### Analysis Stage

```json
// Starting
{
    "type": "pipeline_status",
    "data": {
        "stage": "analysis",
        "status": "active",
        "message": "Generating analytics..."
    }
}

// Progress
{
    "type": "progress",
    "data": {
        "stage": "analysis",
        "current": 50,
        "total": 82,
        "percentage": 61,
        "message": "Generating ticker summaries",
        "details": {
            "tickersProcessed": 50,
            "summariesGenerated": 50
        }
    }
}

// Completed
{
    "type": "pipeline_status",
    "data": {
        "stage": "analysis",
        "status": "completed",
        "message": "Generated summaries for 82 tickers"
    }
}
```

## Complete Pipeline Example

Here's the sequence of messages for a complete pipeline run:

```json
// 1. Reset pipeline
{ "type": "pipeline_reset" }

// 2. Start scraping
{ "type": "pipeline_status", "data": { "stage": "scraping", "status": "active", "message": "Starting download..." }}

// 3. Scraping progress
{ "type": "progress", "data": { "stage": "scraping", "current": 5, "total": 10, "percentage": 50 }}

// 4. Scraping complete
{ "type": "pipeline_status", "data": { "stage": "scraping", "status": "completed", "message": "Downloaded 10 files" }}

// 5. Start processing
{ "type": "pipeline_status", "data": { "stage": "processing", "status": "active", "message": "Processing files..." }}

// ... continue for all stages

// Final. Pipeline complete
{ "type": "pipeline_complete", "data": { "duration": "5 minutes", "summary": "Processed 10 files, 82 tickers" }}
```

## Frontend Implementation

### Handling Pipeline Status

```javascript
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'pipeline_status':
            updatePipelineStage(message.data.stage, message.data.status);
            updateStatusMessage(message.data.message);
            break;
            
        case 'progress':
            updateProgressBar(message.data);
            break;
            
        case 'pipeline_reset':
            resetAllStages();
            break;
            
        case 'pipeline_complete':
            markAllStagesComplete();
            showCompletionSummary(message.data);
            break;
    }
};

function updatePipelineStage(stage, status) {
    const stageElement = document.getElementById(`stage-${stage}`);
    if (stageElement) {
        stageElement.className = `pipeline-stage ${status}`;
    }
}
```

## Backend Implementation

### Sending Status Updates

```go
// In WebSocket hub
func (h *Hub) SendPipelineStatus(stage, status, message string) {
    msg := map[string]interface{}{
        "type": "pipeline_status",
        "data": map[string]interface{}{
            "stage":   stage,
            "status":  status,
            "message": message,
        },
        "timestamp": time.Now().Format(time.RFC3339),
    }
    h.Broadcast(msg)
}

// In scraper
hub.SendPipelineStatus("scraping", "active", "Starting ISX data download...")

// Progress update
hub.SendProgress("scraping", downloaded, total, "Downloading Excel reports")

// Completion
hub.SendPipelineStatus("scraping", "completed", fmt.Sprintf("Downloaded %d new reports", count))
```

## Migration from Text-Based Parsing

### Current (Text Parsing)
```javascript
if (msg.includes('Starting ISX data download')) {
    activatePipelineStage('stage-scraping');
}
```

### New (Structured Messages)
```javascript
if (message.type === 'pipeline_status' && 
    message.data.stage === 'scraping' && 
    message.data.status === 'active') {
    activatePipelineStage('stage-scraping');
}
```

## Benefits

1. **Reliability**: No more regex parsing of text messages
2. **Consistency**: Same format across all backend components
3. **Extensibility**: Easy to add new fields without breaking existing code
4. **Debugging**: Clear, structured messages in dev tools
5. **Internationalization**: UI text separate from protocol

## Backward Compatibility

During migration, the frontend should support both:
1. New structured messages (primary)
2. Legacy text parsing (fallback)

This allows gradual migration of backend components.