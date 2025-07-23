# Complete WebSocket Message Specification

## Overview

This document provides a comprehensive specification of ALL WebSocket messages in the ISX Daily Reports Scrapper system, including both standardized and legacy formats that need migration.

## Current Message Types

### 1. Standardized Messages (Using Hub)

#### Output Messages
```json
{
    "type": "output",
    "data": {
        "message": "Processing completed successfully",
        "level": "info|success|warning|error"
    }
}
```

#### Data Update Messages
```json
{
    "type": "data_update",
    "subtype": "ticker_summary|combined_data|daily_report|index_data|ticker_history",
    "action": "created|updated|deleted",
    "data": {
        "filename": "ticker_summary.json",
        "path": "ticker_summary.json"
    },
    "timestamp": "2025-07-18T10:30:00Z"
}
```

#### Pipeline Status Messages
```json
{
    "type": "pipeline_status",
    "data": {
        "stage": "scraping|processing|indices|analysis",
        "status": "inactive|active|processing|completed|error",
        "message": "Starting ISX data download..."
    }
}
```

#### Pipeline Control Messages
```json
{
    "type": "pipeline_reset"
}

{
    "type": "pipeline_complete",
    "data": {
        "message": "Complete data pipeline finished!"
    }
}
```

### 2. Legacy/Non-Standardized Messages

#### Welcome Message (Direct WebSocket)
**Current Format:**
```json
{
    "Type": "info",
    "Message": "Connected to ISX CLI Web Interface (Licensed - 30 days remaining)"
}
```

**Should Be:**
```json
{
    "type": "connection",
    "data": {
        "status": "connected",
        "message": "Connected to ISX CLI Web Interface",
        "license": {
            "valid": true,
            "days_remaining": 30
        }
    }
}
```

#### Refresh Message
**Current Format:**
```json
{
    "type": "refresh",
    "message": "data_updated",
    "command": "scrape"
}
```

**Should Be:**
```json
{
    "type": "data_update",
    "subtype": "all",
    "action": "refresh",
    "data": {
        "source": "scrape",
        "components": ["files", "tickers", "charts"]
    }
}
```

### 3. Missing Standardized Messages

#### Progress Messages (Defined but Unused)
```json
{
    "type": "progress",
    "data": {
        "stage": "scraping",
        "current": 15,
        "total": 50,
        "percentage": 30,
        "message": "Downloaded 15 of 50 files",
        "eta": "2 minutes remaining"
    }
}
```

#### Status Messages (Defined but Unused)
```json
{
    "type": "status",
    "data": {
        "status": "idle|running|completed|error",
        "message": "System is idle"
    }
}
```

## Message Flow Patterns

### 1. Data Collection Pipeline
```
1. pipeline_reset
2. pipeline_status (scraping, active)
3. output (info) - multiple progress updates
4. pipeline_status (scraping, completed)
5. pipeline_status (processing, active)
6. output (info) - processing updates
7. pipeline_status (processing, completed)
8. pipeline_status (indices, active)
9. output (info) - index updates
10. pipeline_status (indices, completed)
11. pipeline_status (analysis, active)
12. output (info) - analysis updates
13. pipeline_status (analysis, completed)
14. pipeline_complete
15. refresh (data_updated) - triggers UI refresh
```

### 2. File System Changes
```
1. File modified → data_update (ticker_summary, updated)
2. Frontend receives → Updates specific component
3. No manual refresh needed
```

## Non-Standardized Patterns to Fix

### 1. Console Output Parsing
**Problem**: Frontend parses text to determine pipeline state
```javascript
if (msg.includes('Starting ISX data download')) {
    activatePipelineStage('stage-scraping');
}
```

**Solution**: Backend should send explicit pipeline_status messages

### 2. Mixed Message Sources
**Problem**: Messages come from multiple sources:
- broadcastMessage() function
- Direct console output from executables
- File watcher events
- Direct WebSocket writes

**Solution**: All messages should go through the WebSocket hub

### 3. Inconsistent Error Handling
**Problem**: Errors reported differently:
- Some via output messages with level "error"
- Some via direct error text
- Some not reported to UI at all

**Solution**: Standardized error message format:
```json
{
    "type": "error",
    "data": {
        "code": "SCRAPE_FAILED",
        "message": "Failed to download data from ISX",
        "details": "Connection timeout after 30 seconds",
        "stage": "scraping",
        "recoverable": true
    }
}
```

## Recommended Message Type Constants

```go
// Message types
const (
    MessageTypeOutput         = "output"
    MessageTypeDataUpdate     = "data_update"
    MessageTypePipelineStatus = "pipeline_status"
    MessageTypePipelineReset  = "pipeline_reset"
    MessageTypePipelineComplete = "pipeline_complete"
    MessageTypeProgress       = "progress"
    MessageTypeStatus         = "status"
    MessageTypeError          = "error"
    MessageTypeConnection     = "connection"
)

// Message levels
const (
    LevelInfo    = "info"
    LevelSuccess = "success"
    LevelWarning = "warning"
    LevelError   = "error"
)

// Pipeline stages
const (
    StageScraping   = "scraping"
    StageProcessing = "processing"
    StageIndices    = "indices"
    StageAnalysis   = "analysis"
)

// Pipeline status
const (
    StatusInactive   = "inactive"
    StatusActive     = "active"
    StatusProcessing = "processing"
    StatusCompleted  = "completed"
    StatusError      = "error"
)
```

## Migration Plan

### Phase 1: Add Missing Standardized Messages
1. Implement progress tracking in all executables
2. Send structured progress messages during operations
3. Add ETA calculations

### Phase 2: Migrate Legacy Messages
1. Replace welcome message with connection message
2. Replace refresh message with proper data_update
3. Remove direct WebSocket writes

### Phase 3: Remove Text Parsing
1. Update executables to send structured messages
2. Remove string matching in frontend
3. Use only structured message handling

### Phase 4: Add New Features
1. Implement error codes and recovery hints
2. Add system status monitoring
3. Add performance metrics messages

## Benefits of Full Standardization

1. **Reliability**: No more regex/string matching
2. **Maintainability**: Clear message contracts
3. **Extensibility**: Easy to add new message types
4. **Debugging**: Structured logs and clear message flow
5. **Internationalization**: Messages separate from display text
6. **Error Handling**: Consistent error reporting
7. **Progress Tracking**: Real-time progress with ETA
8. **Type Safety**: Can generate TypeScript types from Go constants