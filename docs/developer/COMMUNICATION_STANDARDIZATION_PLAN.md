# Communication Standardization Plan

## Overview

This document outlines the plan to standardize all communication between frontend and backend in the ISX Daily Reports Scrapper application.

## Current State Analysis

### 1. WebSocket Messages (Partially Standardized)
- ✅ Output messages (via hub)
- ✅ Data update messages (file changes)
- ✅ Pipeline status messages (basic)
- ❌ Welcome messages (direct write)
- ❌ Refresh messages (non-standard)
- ❌ Progress tracking (unused)
- ❌ Error handling (inconsistent)

### 2. HTTP/REST APIs (Not Standardized)
- Multiple endpoints returning different JSON structures
- No consistent error response format
- No API versioning
- No request/response logging

### 3. Command Output (Not Standardized)
- Plain text console output
- No structured progress reporting
- No machine-readable status codes

## Standardization Goals

1. **All WebSocket messages** follow the hub specification
2. **All API responses** use consistent format
3. **All errors** include error codes and recovery hints
4. **All long operations** report progress with ETA
5. **All status changes** are explicitly communicated

## Implementation Plan

### Phase 1: Standardize Existing WebSocket Messages (1-2 days)

#### 1.1 Fix Welcome Messages
```go
// Replace direct WebSocket write with:
wsHub.BroadcastConnection("connected", licenseInfo)
```

#### 1.2 Fix Refresh Messages
```go
// Replace refresh message with:
wsHub.BroadcastDataUpdate("all", "refresh", map[string]interface{}{
    "source": command,
    "components": []string{"files", "tickers", "charts"},
})
```

#### 1.3 Add Message Type Constants
```go
// internal/websocket/types.go
package websocket

const (
    // Message types
    TypeOutput           = "output"
    TypeDataUpdate       = "data_update"
    TypePipelineStatus   = "pipeline_status"
    TypeProgress         = "progress"
    TypeError            = "error"
    TypeConnection       = "connection"
    
    // Levels
    LevelInfo    = "info"
    LevelSuccess = "success"
    LevelWarning = "warning"
    LevelError   = "error"
)
```

### Phase 2: Add Progress Tracking (2-3 days)

#### 2.1 Update Executables
Modify scraper.exe, process.exe, and indexcsv.exe to send structured progress:

```go
// In scraper
hub.SendProgress("scraping", downloaded, total, "Downloading Excel files")
```

#### 2.2 Implement Progress Calculation
```go
type ProgressTracker struct {
    Stage     string
    Total     int
    Current   int
    StartTime time.Time
}

func (p *ProgressTracker) GetETA() string {
    elapsed := time.Since(p.StartTime)
    rate := float64(p.Current) / elapsed.Seconds()
    remaining := float64(p.Total - p.Current) / rate
    return fmt.Sprintf("%.0f seconds remaining", remaining)
}
```

### Phase 3: Standardize API Responses (2-3 days)

#### 3.1 Define Standard API Response
```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
    Meta    *APIMeta    `json:"meta,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type APIMeta struct {
    Timestamp   time.Time `json:"timestamp"`
    Version     string    `json:"version"`
    RequestID   string    `json:"request_id"`
}
```

#### 3.2 Update All API Endpoints
Example transformation:
```go
// Before
json.NewEncoder(w).Encode(data)

// After
response := APIResponse{
    Success: true,
    Data:    data,
    Meta:    &APIMeta{
        Timestamp: time.Now(),
        Version:   "2.0.0",
    },
}
json.NewEncoder(w).Encode(response)
```

### Phase 4: Implement Error Codes (1-2 days)

#### 4.1 Define Error Code System
```go
const (
    // Scraping errors (1xxx)
    ErrScrapingTimeout     = "ERR_1001"
    ErrScrapingNoData      = "ERR_1002"
    ErrScrapingAuthFailed  = "ERR_1003"
    
    // Processing errors (2xxx)
    ErrProcessingInvalidFile = "ERR_2001"
    ErrProcessingNoColumns   = "ERR_2002"
    
    // System errors (9xxx)
    ErrSystemOutOfMemory = "ERR_9001"
    ErrSystemDiskFull    = "ERR_9002"
)
```

#### 4.2 Error Recovery Hints
```go
var ErrorRecoveryHints = map[string]string{
    ErrScrapingTimeout:     "Check internet connection and try again",
    ErrScrapingNoData:      "No data available for the specified date range",
    ErrProcessingInvalidFile: "File may be corrupted, try re-downloading",
}
```

### Phase 5: Add System Monitoring (2-3 days)

#### 5.1 System Status Messages
```go
// Send periodic system status
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        wsHub.BroadcastSystemStatus(getSystemStatus())
    }
}()

func getSystemStatus() SystemStatus {
    return SystemStatus{
        CPU:        getCPUUsage(),
        Memory:     getMemoryUsage(),
        Disk:       getDiskUsage(),
        Processing: getProcessingStatus(),
    }
}
```

#### 5.2 Performance Metrics
```go
type PerformanceMetrics struct {
    OperationType string        `json:"operation_type"`
    Duration      time.Duration `json:"duration"`
    RecordsProcessed int        `json:"records_processed"`
    ErrorCount    int           `json:"error_count"`
}
```

## Testing Strategy

### 1. Unit Tests
- Test each message type creation
- Test error code generation
- Test progress calculations

### 2. Integration Tests
- Test full pipeline with progress tracking
- Test error scenarios and recovery
- Test API response consistency

### 3. Frontend Tests
- Verify all message types handled correctly
- Test progress bar updates
- Test error display and recovery hints

## Migration Strategy

1. **Backward Compatibility**: Keep legacy handlers during migration
2. **Feature Flags**: Use flags to enable new features gradually
3. **Monitoring**: Log both old and new message formats
4. **Gradual Rollout**: Migrate one component at a time

## Success Metrics

- 100% of messages follow standard format
- All long operations show progress with ETA
- All errors include recovery hints
- API response time < 100ms for all endpoints
- Zero message parsing errors in frontend

## Timeline

- Week 1: Phase 1 & 2 (WebSocket standardization & progress)
- Week 2: Phase 3 & 4 (API standardization & error codes)
- Week 3: Phase 5 & testing (System monitoring & validation)

## Benefits

1. **Developer Experience**: Clear contracts, easier debugging
2. **User Experience**: Better progress feedback, helpful errors
3. **Maintainability**: Consistent patterns, less code duplication
4. **Extensibility**: Easy to add new message types
5. **Reliability**: No more string parsing, type-safe messages